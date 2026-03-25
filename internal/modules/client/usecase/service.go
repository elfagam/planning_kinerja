package usecase

import (
	"context"
	"fmt"
	"strings"

	"e-plan-ai/internal/modules/client/domain"
)

func (s *Service) List(ctx context.Context, actor Actor, filter ListFilter) ([]domain.Client, int64, error) {
	role := normalizeActorRole(actor.Role)
	if actor.ID == 0 || !isValidActorRole(role) {
		return nil, 0, domain.ErrForbiddenOperation
	}
	if !canRoleReadClient(role) {
		return nil, 0, domain.ErrForbiddenOperation
	}

	if filter.Page <= 0 {
		filter.Page = 1
	}
	if filter.Limit <= 0 {
		filter.Limit = 10
	}
	if filter.Limit > 100 {
		filter.Limit = 100
	}
	filter.Q = strings.TrimSpace(filter.Q)
	filter.Status = strings.TrimSpace(strings.ToUpper(filter.Status))
	if filter.Status != "" {
		if _, err := domain.ParseStatus(filter.Status); err != nil {
			return nil, 0, err
		}
	}

	return s.repo.List(ctx, filter)
}

func (s *Service) Get(ctx context.Context, actor Actor, id uint64) (domain.Client, error) {
	role := normalizeActorRole(actor.Role)
	if actor.ID == 0 || !isValidActorRole(role) {
		return domain.Client{}, domain.ErrForbiddenOperation
	}
	if !canRoleReadClient(role) {
		return domain.Client{}, domain.ErrForbiddenOperation
	}
	return s.repo.GetByID(ctx, id)
}

func (s *Service) Create(ctx context.Context, actor Actor, client domain.Client) (domain.Client, error) {
	role := normalizeActorRole(actor.Role)
	if actor.ID == 0 || !isValidActorRole(role) {
		return domain.Client{}, domain.ErrForbiddenOperation
	}
	if !canRoleWriteClient(role) {
		return domain.Client{}, domain.ErrForbiddenOperation
	}

	client.Kode = strings.TrimSpace(client.Kode)
	client.Nama = strings.TrimSpace(client.Nama)
	if client.Kode == "" || client.Nama == "" {
		return domain.Client{}, fmt.Errorf("%w: kode and nama are required", domain.ErrValidation)
	}
	if client.Status == "" {
		client.Status = domain.StatusDraft
	}
	if _, err := domain.ParseStatus(string(client.Status)); err != nil {
		return domain.Client{}, err
	}
	client.CreatedBy = &actor.ID
	client.UpdatedBy = &actor.ID

	var created domain.Client
	err := s.tx.WithinTransaction(ctx, func(txCtx context.Context) error {
		result, err := s.repo.Create(txCtx, client)
		if err != nil {
			return err
		}
		created = result

		auditPayload := map[string]any{
			"kode":   created.Kode,
			"nama":   created.Nama,
			"status": created.Status,
		}
		if created.UnitPengusulID != nil {
			auditPayload["unit_pengusul_id"] = *created.UnitPengusulID
		}

		return s.repo.AppendAudit(txCtx, buildAuditEntry(actor, "CLIENT_CREATE", created.ID, auditPayload))
	})
	if err != nil {
		return domain.Client{}, err
	}

	return created, nil
}

func (s *Service) Update(ctx context.Context, id uint64, actor Actor, payload domain.Client) (domain.Client, error) {
	role := normalizeActorRole(actor.Role)
	if actor.ID == 0 || !isValidActorRole(role) {
		return domain.Client{}, domain.ErrForbiddenOperation
	}
	if !canRoleWriteClient(role) {
		return domain.Client{}, domain.ErrForbiddenOperation
	}

	current, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return domain.Client{}, err
	}
	if !canActorMutateClient(role, actor.ID, current) {
		return domain.Client{}, domain.ErrForbiddenOperation
	}

	payload.Kode = strings.TrimSpace(payload.Kode)
	payload.Nama = strings.TrimSpace(payload.Nama)
	if payload.Kode != "" {
		current.Kode = payload.Kode
	}
	if payload.Nama != "" {
		current.Nama = payload.Nama
	}
	current.UnitPengusulID = payload.UnitPengusulID
	current.UpdatedBy = &actor.ID

	var updated domain.Client
	err = s.tx.WithinTransaction(ctx, func(txCtx context.Context) error {
		result, err := s.repo.Update(txCtx, current)
		if err != nil {
			return err
		}
		updated = result

		auditPayload := map[string]any{
			"kode":   updated.Kode,
			"nama":   updated.Nama,
			"status": updated.Status,
		}
		if updated.UnitPengusulID != nil {
			auditPayload["unit_pengusul_id"] = *updated.UnitPengusulID
		}

		return s.repo.AppendAudit(txCtx, buildAuditEntry(actor, "CLIENT_UPDATE", updated.ID, auditPayload))
	})
	if err != nil {
		return domain.Client{}, err
	}

	return updated, nil
}

func (s *Service) Delete(ctx context.Context, id uint64, actor Actor) error {
	role := normalizeActorRole(actor.Role)
	if actor.ID == 0 || !isValidActorRole(role) {
		return domain.ErrForbiddenOperation
	}
	if !canRoleWriteClient(role) {
		return domain.ErrForbiddenOperation
	}

	current, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if !canActorMutateClient(role, actor.ID, current) {
		return domain.ErrForbiddenOperation
	}

	return s.tx.WithinTransaction(ctx, func(txCtx context.Context) error {
		if err := s.repo.SoftDelete(txCtx, id); err != nil {
			return err
		}

		auditPayload := map[string]any{
			"kode":   current.Kode,
			"nama":   current.Nama,
			"status": current.Status,
		}
		if current.UnitPengusulID != nil {
			auditPayload["unit_pengusul_id"] = *current.UnitPengusulID
		}

		return s.repo.AppendAudit(txCtx, buildAuditEntry(actor, "CLIENT_DELETE", id, auditPayload))
	})
}

func (s *Service) StatusHistory(ctx context.Context, clientID uint64) ([]domain.StatusHistory, error) {
	if _, err := s.repo.GetByID(ctx, clientID); err != nil {
		return nil, err
	}
	return s.repo.ListHistory(ctx, clientID)
}
