package usecase

import (
	"context"
	"fmt"
)

func (s *Service) Submit(ctx context.Context, renjaID int64, actorID int64) error {
	return s.tx.WithinTransaction(ctx, func(txCtx context.Context) error {
		renja, err := s.repo.GetByID(txCtx, renjaID)
		if err != nil {
			return err
		}
		if err := renja.Submit(actorID); err != nil {
			return err
		}
		if err := s.repo.Save(txCtx, renja); err != nil {
			return err
		}
		return s.repo.AppendAudit(txCtx, actorID, "RENJA_SUBMIT", renjaID, "renja submitted")
	})
}

func (s *Service) Approve(ctx context.Context, renjaID int64, actorID int64) error {
	return s.tx.WithinTransaction(ctx, func(txCtx context.Context) error {
		renja, err := s.repo.GetByID(txCtx, renjaID)
		if err != nil {
			return err
		}
		if err := renja.Approve(actorID); err != nil {
			return err
		}
		if err := s.repo.Save(txCtx, renja); err != nil {
			return err
		}
		note := fmt.Sprintf("renja approved by actor %d", actorID)
		return s.repo.AppendAudit(txCtx, actorID, "RENJA_APPROVE", renjaID, note)
	})
}

func (s *Service) Reject(ctx context.Context, renjaID int64, actorID int64, reason string) error {
	return s.tx.WithinTransaction(ctx, func(txCtx context.Context) error {
		renja, err := s.repo.GetByID(txCtx, renjaID)
		if err != nil {
			return err
		}
		if err := renja.Reject(actorID, reason); err != nil {
			return err
		}
		if err := s.repo.Save(txCtx, renja); err != nil {
			return err
		}
		note := fmt.Sprintf("renja rejected: %s", reason)
		return s.repo.AppendAudit(txCtx, actorID, "RENJA_REJECT", renjaID, note)
	})
}
