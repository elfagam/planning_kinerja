package usecase

import (
	"context"

	"e-plan-ai/internal/modules/client/domain"
)

type ListFilter struct {
	Q              string
	Status         string
	UnitPengusulID *uint64
	Page           int
	Limit          int
}

type Repository interface {
	List(ctx context.Context, filter ListFilter) ([]domain.Client, int64, error)
	ListAuditLogs(ctx context.Context, filter AuditListFilter) ([]AuditLog, int64, error)
	GetByID(ctx context.Context, id uint64) (domain.Client, error)
	GetByIDForUpdate(ctx context.Context, id uint64) (domain.Client, error)
	Create(ctx context.Context, client domain.Client) (domain.Client, error)
	Update(ctx context.Context, client domain.Client) (domain.Client, error)
	SoftDelete(ctx context.Context, id uint64) error
	CreateHistory(ctx context.Context, history domain.StatusHistory) error
	ListHistory(ctx context.Context, clientID uint64) ([]domain.StatusHistory, error)
	AppendAudit(ctx context.Context, entry AuditLogEntry) error
}

type TxManager interface {
	WithinTransaction(ctx context.Context, fn func(ctx context.Context) error) error
}

type Actor struct {
	ID        uint64
	Role      string
	Name      string
	IPAddress string
	UserAgent string
}

type AuditLogEntry struct {
	UserID         *uint64
	Action         string
	ResourceType   string
	ResourceID     *uint64
	RequestPayload *string
	IPAddress      *string
	UserAgent      *string
}

type AuditListFilter struct {
	Action     string
	UserID     *uint64
	ResourceID *uint64
	Page       int
	Limit      int
}

type AuditLog struct {
	ID             uint64
	UserID         *uint64
	UserName       *string
	Action         string
	ResourceType   string
	ResourceID     *uint64
	RequestPayload *string
	IPAddress      *string
	UserAgent      *string
	CreatedAt      string
}

const (
	RoleAdmin       = "ADMIN"
	RoleOperator    = "OPERATOR"
	RolePerencana   = "PERENCANA"
	RoleVerifikator = "VERIFIKATOR"
	RolePimpinan    = "PIMPINAN"
)

type Service struct {
	tx   TxManager
	repo Repository
}

func NewService(tx TxManager, repo Repository) *Service {
	return &Service{tx: tx, repo: repo}
}
