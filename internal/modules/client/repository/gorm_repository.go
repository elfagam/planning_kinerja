package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"e-plan-ai/internal/modules/client/domain"
	"e-plan-ai/internal/modules/client/usecase"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type GormRepository struct {
	db *gorm.DB
}

type gormTxContextKey struct{}

type GormTxManager struct {
	db *gorm.DB
}

const clientAuditResourceType = "CLIENT"

type clientRow struct {
	ID             uint64     `gorm:"column:id"`
	Kode           string     `gorm:"column:kode"`
	Nama           string     `gorm:"column:nama"`
	Status         string     `gorm:"column:status"`
	UnitPengusulID *uint64    `gorm:"column:unit_pengusul_id"`
	CreatedBy      *uint64    `gorm:"column:created_by"`
	UpdatedBy      *uint64    `gorm:"column:updated_by"`
	ApprovedBy     *uint64    `gorm:"column:approved_by"`
	ApprovedAt     *time.Time `gorm:"column:approved_at"`
	RejectedBy     *uint64    `gorm:"column:rejected_by"`
	RejectedAt     *time.Time `gorm:"column:rejected_at"`
	RejectedReason *string    `gorm:"column:rejected_reason"`
	CreatedAt      time.Time  `gorm:"column:created_at"`
	UpdatedAt      time.Time  `gorm:"column:updated_at"`
	DeletedAt      *time.Time `gorm:"column:deleted_at"`
}

type historyRow struct {
	ID         uint64    `gorm:"column:id"`
	ClientID   uint64    `gorm:"column:client_id"`
	FromStatus *string   `gorm:"column:from_status"`
	ToStatus   string    `gorm:"column:to_status"`
	Action     string    `gorm:"column:action"`
	Reason     *string   `gorm:"column:reason"`
	Note       *string   `gorm:"column:note"`
	ActorID    *uint64   `gorm:"column:actor_id"`
	ActorName  *string   `gorm:"column:actor_name"`
	CreatedAt  time.Time `gorm:"column:created_at"`
}

type auditLogRow struct {
	ID             uint64    `gorm:"column:id"`
	UserID         *uint64   `gorm:"column:user_id"`
	UserName       *string   `gorm:"column:user_name;->"`
	Action         string    `gorm:"column:action"`
	ResourceType   string    `gorm:"column:resource_type"`
	ResourceID     *uint64   `gorm:"column:resource_id"`
	RequestPayload *string   `gorm:"column:request_payload"`
	IPAddress      *string   `gorm:"column:ip_address"`
	UserAgent      *string   `gorm:"column:user_agent"`
	CreatedAt      time.Time `gorm:"column:created_at"`
}

func NewGormRepository(db *gorm.DB) *GormRepository {
	return &GormRepository{db: db}
}

func NewGormTxManager(db *gorm.DB) *GormTxManager {
	return &GormTxManager{db: db}
}

func (m *GormTxManager) WithinTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	if m == nil || m.db == nil {
		return fmt.Errorf("nil gorm tx manager")
	}

	return m.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		txCtx := context.WithValue(ctx, gormTxContextKey{}, tx)
		return fn(txCtx)
	})
}

func (r *GormRepository) List(ctx context.Context, filter usecase.ListFilter) ([]domain.Client, int64, error) {
	if r == nil || r.db == nil {
		return nil, 0, fmt.Errorf("nil client repository db")
	}

	query := r.dbFromContext(ctx).Table("clients").Where("deleted_at IS NULL")
	if filter.Q != "" {
		like := "%" + strings.TrimSpace(filter.Q) + "%"
		query = query.Where("kode LIKE ? OR nama LIKE ?", like, like)
	}
	if filter.Status != "" {
		query = query.Where("status = ?", filter.Status)
	}
	if filter.UnitPengusulID != nil {
		query = query.Where("unit_pengusul_id = ?", *filter.UnitPengusulID)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if filter.Page <= 0 {
		filter.Page = 1
	}
	if filter.Limit <= 0 {
		filter.Limit = 10
	}
	offset := (filter.Page - 1) * filter.Limit

	var rows []clientRow
	if err := query.Order("id DESC").Offset(offset).Limit(filter.Limit).Find(&rows).Error; err != nil {
		return nil, 0, err
	}

	items := make([]domain.Client, 0, len(rows))
	for _, row := range rows {
		items = append(items, row.toDomain())
	}
	return items, total, nil
}

func (r *GormRepository) ListAuditLogs(ctx context.Context, filter usecase.AuditListFilter) ([]usecase.AuditLog, int64, error) {
	if r == nil || r.db == nil {
		return nil, 0, fmt.Errorf("nil client repository db")
	}

	query := r.dbFromContext(ctx).
		Table("audit_logs al").
		Select("al.id, al.user_id, u.nama_lengkap AS user_name, al.action, al.resource_type, al.resource_id, al.request_payload, al.ip_address, al.user_agent, al.created_at").
		Joins("LEFT JOIN users u ON u.id = al.user_id").
		Where("al.resource_type = ?", clientAuditResourceType)

	if filter.Action != "" {
		query = query.Where("al.action = ?", filter.Action)
	}
	if filter.UserID != nil {
		query = query.Where("al.user_id = ?", *filter.UserID)
	}
	if filter.ResourceID != nil {
		query = query.Where("al.resource_id = ?", *filter.ResourceID)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if filter.Page <= 0 {
		filter.Page = 1
	}
	if filter.Limit <= 0 {
		filter.Limit = 10
	}
	offset := (filter.Page - 1) * filter.Limit

	var rows []auditLogRow
	if err := query.Order("al.created_at DESC, al.id DESC").Offset(offset).Limit(filter.Limit).Find(&rows).Error; err != nil {
		return nil, 0, err
	}

	items := make([]usecase.AuditLog, 0, len(rows))
	for _, row := range rows {
		items = append(items, row.toAuditLog())
	}

	return items, total, nil
}

func (r *GormRepository) GetByID(ctx context.Context, id uint64) (domain.Client, error) {
	return r.getByID(ctx, id, false)
}

func (r *GormRepository) GetByIDForUpdate(ctx context.Context, id uint64) (domain.Client, error) {
	return r.getByID(ctx, id, true)
}

func (r *GormRepository) getByID(ctx context.Context, id uint64, forUpdate bool) (domain.Client, error) {
	if r == nil || r.db == nil {
		return domain.Client{}, fmt.Errorf("nil client repository db")
	}
	if id == 0 {
		return domain.Client{}, domain.ErrClientNotFound
	}

	query := r.dbFromContext(ctx).Table("clients").Where("id = ? AND deleted_at IS NULL", id)
	if forUpdate {
		query = query.Clauses(clause.Locking{Strength: "UPDATE"})
	}

	var row clientRow
	if err := query.Take(&row).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domain.Client{}, domain.ErrClientNotFound
		}
		return domain.Client{}, err
	}
	return row.toDomain(), nil
}

func (r *GormRepository) Create(ctx context.Context, client domain.Client) (domain.Client, error) {
	if r == nil || r.db == nil {
		return domain.Client{}, fmt.Errorf("nil client repository db")
	}

	now := time.Now()
	row := clientRow{
		Kode:           client.Kode,
		Nama:           client.Nama,
		Status:         string(client.Status),
		UnitPengusulID: client.UnitPengusulID,
		CreatedBy:      client.CreatedBy,
		UpdatedBy:      client.UpdatedBy,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	if err := r.dbFromContext(ctx).Table("clients").Create(&row).Error; err != nil {
		return domain.Client{}, err
	}
	return row.toDomain(), nil
}

func (r *GormRepository) Update(ctx context.Context, client domain.Client) (domain.Client, error) {
	if r == nil || r.db == nil {
		return domain.Client{}, fmt.Errorf("nil client repository db")
	}
	if client.ID == 0 {
		return domain.Client{}, domain.ErrClientNotFound
	}

	updates := map[string]any{
		"kode":             client.Kode,
		"nama":             client.Nama,
		"status":           string(client.Status),
		"unit_pengusul_id": client.UnitPengusulID,
		"updated_by":       client.UpdatedBy,
		"approved_by":      client.ApprovedBy,
		"approved_at":      client.ApprovedAt,
		"rejected_by":      client.RejectedBy,
		"rejected_at":      client.RejectedAt,
		"rejected_reason":  client.RejectedReason,
		"updated_at":       time.Now(),
	}

	res := r.dbFromContext(ctx).Table("clients").Where("id = ? AND deleted_at IS NULL", client.ID).Updates(updates)
	if res.Error != nil {
		return domain.Client{}, res.Error
	}
	if res.RowsAffected == 0 {
		return domain.Client{}, domain.ErrClientNotFound
	}
	return r.GetByID(ctx, client.ID)
}

func (r *GormRepository) SoftDelete(ctx context.Context, id uint64) error {
	if r == nil || r.db == nil {
		return fmt.Errorf("nil client repository db")
	}
	if id == 0 {
		return domain.ErrClientNotFound
	}

	res := r.dbFromContext(ctx).Table("clients").Where("id = ? AND deleted_at IS NULL", id).Updates(map[string]any{
		"deleted_at": time.Now(),
	})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return domain.ErrClientNotFound
	}
	return nil
}

func (r *GormRepository) CreateHistory(ctx context.Context, history domain.StatusHistory) error {
	if r == nil || r.db == nil {
		return fmt.Errorf("nil client repository db")
	}

	var fromStatus *string
	if history.FromStatus != nil {
		v := string(*history.FromStatus)
		fromStatus = &v
	}

	row := historyRow{
		ClientID:   history.ClientID,
		FromStatus: fromStatus,
		ToStatus:   string(history.ToStatus),
		Action:     history.Action,
		Reason:     history.Reason,
		Note:       history.Note,
		ActorID:    history.ActorID,
		ActorName:  history.ActorName,
		CreatedAt:  time.Now(),
	}
	return r.dbFromContext(ctx).Table("client_status_histories").Create(&row).Error
}

func (r *GormRepository) ListHistory(ctx context.Context, clientID uint64) ([]domain.StatusHistory, error) {
	if r == nil || r.db == nil {
		return nil, fmt.Errorf("nil client repository db")
	}

	var rows []historyRow
	if err := r.dbFromContext(ctx).Table("client_status_histories").Where("client_id = ?", clientID).Order("created_at DESC").Find(&rows).Error; err != nil {
		return nil, err
	}

	items := make([]domain.StatusHistory, 0, len(rows))
	for _, row := range rows {
		items = append(items, row.toDomain())
	}
	return items, nil
}

func (r *GormRepository) AppendAudit(ctx context.Context, entry usecase.AuditLogEntry) error {
	if r == nil || r.db == nil {
		return fmt.Errorf("nil client repository db")
	}

	row := auditLogRow{
		UserID:         entry.UserID,
		Action:         entry.Action,
		ResourceType:   entry.ResourceType,
		ResourceID:     entry.ResourceID,
		RequestPayload: entry.RequestPayload,
		IPAddress:      entry.IPAddress,
		UserAgent:      entry.UserAgent,
		CreatedAt:      time.Now(),
	}

	return r.dbFromContext(ctx).Table("audit_logs").Create(&row).Error
}

func (r *GormRepository) dbFromContext(ctx context.Context) *gorm.DB {
	if r.db == nil {
		return nil
	}
	if tx := GormTxFromContext(ctx); tx != nil {
		return tx
	}
	return r.db.WithContext(ctx)
}

func GormTxFromContext(ctx context.Context) *gorm.DB {
	if ctx == nil {
		return nil
	}
	tx, ok := ctx.Value(gormTxContextKey{}).(*gorm.DB)
	if !ok {
		return nil
	}
	return tx
}

func (r clientRow) toDomain() domain.Client {
	status, _ := domain.ParseStatus(r.Status)
	return domain.Client{
		ID:             r.ID,
		Kode:           r.Kode,
		Nama:           r.Nama,
		Status:         status,
		UnitPengusulID: r.UnitPengusulID,
		CreatedBy:      r.CreatedBy,
		UpdatedBy:      r.UpdatedBy,
		ApprovedBy:     r.ApprovedBy,
		ApprovedAt:     r.ApprovedAt,
		RejectedBy:     r.RejectedBy,
		RejectedAt:     r.RejectedAt,
		RejectedReason: r.RejectedReason,
		CreatedAt:      r.CreatedAt,
		UpdatedAt:      r.UpdatedAt,
		DeletedAt:      r.DeletedAt,
	}
}

func (r historyRow) toDomain() domain.StatusHistory {
	var fromStatus *domain.Status
	if r.FromStatus != nil {
		if parsed, err := domain.ParseStatus(*r.FromStatus); err == nil {
			fromStatus = &parsed
		}
	}
	toStatus, _ := domain.ParseStatus(r.ToStatus)
	return domain.StatusHistory{
		ID:         r.ID,
		ClientID:   r.ClientID,
		FromStatus: fromStatus,
		ToStatus:   toStatus,
		Action:     r.Action,
		Reason:     r.Reason,
		Note:       r.Note,
		ActorID:    r.ActorID,
		ActorName:  r.ActorName,
		CreatedAt:  r.CreatedAt,
	}
}

func (r auditLogRow) toAuditLog() usecase.AuditLog {
	return usecase.AuditLog{
		ID:             r.ID,
		UserID:         r.UserID,
		UserName:       r.UserName,
		Action:         r.Action,
		ResourceType:   r.ResourceType,
		ResourceID:     r.ResourceID,
		RequestPayload: r.RequestPayload,
		IPAddress:      r.IPAddress,
		UserAgent:      r.UserAgent,
		CreatedAt:      r.CreatedAt.Format(time.RFC3339),
	}
}
