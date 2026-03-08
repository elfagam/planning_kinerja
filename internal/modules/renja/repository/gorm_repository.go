package repository

import (
	"context"
	"fmt"
	"sync"
	"time"

	"e-plan-ai/internal/modules/renja/domain"
	"e-plan-ai/internal/shared/database"

	"gorm.io/gorm"
)

type RenjaGormRepository struct {
	db            *gorm.DB
	auditTableMux sync.Once
}

func NewRenjaGormRepository(db *gorm.DB) *RenjaGormRepository {
	return &RenjaGormRepository{db: db}
}

func (r *RenjaGormRepository) GetByID(ctx context.Context, id int64) (domain.Renja, error) {
	if r == nil || r.db == nil {
		return domain.Renja{}, fmt.Errorf("nil renja repository db")
	}

	if id <= 0 {
		return domain.Renja{}, fmt.Errorf("invalid renja id: %d", id)
	}

	var model database.RencanaKerja
	err := r.dbFromContext(ctx).Where("id = ?", id).First(&model).Error
	if err != nil {
		return domain.Renja{}, err
	}

	return ToDomainRencanaKerja(model), nil
}

func (r *RenjaGormRepository) Save(ctx context.Context, renja domain.Renja) error {
	if r == nil || r.db == nil {
		return fmt.Errorf("nil renja repository db")
	}

	if renja.ID <= 0 {
		return fmt.Errorf("invalid renja id: %d", renja.ID)
	}

	model := ToGormRencanaKerja(renja)
	updates := map[string]any{
		"status":              model.Status,
		"catatan":             model.Catatan,
		"disetujui_oleh":      model.DisetujuiOleh,
		"tanggal_persetujuan": model.TanggalPersetujuan,
		"updated_at":          time.Now(),
	}

	res := r.dbFromContext(ctx).
		Model(&database.RencanaKerja{}).
		Where("id = ?", model.ID).
		Updates(updates)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}

func (r *RenjaGormRepository) AppendAudit(ctx context.Context, actorID int64, action string, resourceID int64, notes string) error {
	if r == nil || r.db == nil {
		return fmt.Errorf("nil renja repository db")
	}

	if err := r.ensureAuditTable(ctx); err != nil {
		return err
	}

	const insertAudit = `
INSERT INTO renja_audit_logs (actor_id, action, resource_id, notes, created_at)
VALUES (?, ?, ?, ?, ?)`

	return r.dbFromContext(ctx).
		Exec(insertAudit, actorID, action, resourceID, notes, time.Now()).
		Error
}

func (r *RenjaGormRepository) ensureAuditTable(ctx context.Context) error {
	var ensureErr error
	r.auditTableMux.Do(func() {
		const ddl = `
CREATE TABLE IF NOT EXISTS renja_audit_logs (
    id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
    actor_id BIGINT NOT NULL,
    action VARCHAR(64) NOT NULL,
    resource_id BIGINT NOT NULL,
    notes TEXT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_renja_audit_actor (actor_id),
    INDEX idx_renja_audit_resource (resource_id),
    INDEX idx_renja_audit_action (action)
)`
		ensureErr = r.dbFromContext(ctx).Exec(ddl).Error
	})
	return ensureErr
}

func (r *RenjaGormRepository) dbFromContext(ctx context.Context) *gorm.DB {
	if r.db == nil {
		return nil
	}
	if tx := GormTxFromContext(ctx); tx != nil {
		return tx
	}
	return r.db.WithContext(ctx)
}

var _ interface {
	GetByID(context.Context, int64) (domain.Renja, error)
	Save(context.Context, domain.Renja) error
	AppendAudit(context.Context, int64, string, int64, string) error
} = (*RenjaGormRepository)(nil)
