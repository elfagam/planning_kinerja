package usecase

import (
	"context"

	"e-plan-ai/internal/modules/renja/domain"
)

type Repository interface {
	GetByID(ctx context.Context, id int64) (domain.Renja, error)
	Save(ctx context.Context, renja domain.Renja) error
	GetRencanaKerjaCSVData(ctx context.Context, subKegiatanID uint) ([]domain.ExportIndikatorCSVFlatDTO, error)
	AppendAudit(ctx context.Context, actorID int64, action string, resourceID int64, notes string) error
}

type TxManager interface {
	WithinTransaction(ctx context.Context, fn func(ctx context.Context) error) error
}

type Service struct {
	tx   TxManager
	repo Repository
}

func NewService(tx TxManager, repo Repository) *Service {
	return &Service{tx: tx, repo: repo}
}
