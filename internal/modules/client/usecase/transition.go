package usecase

import (
	"context"
	"strings"

	"e-plan-ai/internal/modules/client/domain"
)

func (s *Service) Submit(ctx context.Context, clientID uint64, actor Actor, note string) error {
	return s.transition(ctx, "submit", clientID, actor, strings.TrimSpace(note), "")
}

func (s *Service) Unsubmit(ctx context.Context, clientID uint64, actor Actor, reason string) error {
	return s.transition(ctx, "unsubmit", clientID, actor, "", strings.TrimSpace(reason))
}

func (s *Service) Reject(ctx context.Context, clientID uint64, actor Actor, reason string) error {
	reason = strings.TrimSpace(reason)
	if reason == "" {
		return domain.ErrReasonRequired
	}
	return s.transition(ctx, "reject", clientID, actor, "", reason)
}

func (s *Service) ReEvaluate(ctx context.Context, clientID uint64, actor Actor, reason string) error {
	return s.transition(ctx, "re-evaluate", clientID, actor, "", strings.TrimSpace(reason))
}

func (s *Service) Approve(ctx context.Context, clientID uint64, actor Actor, note string) error {
	return s.transition(ctx, "approve", clientID, actor, strings.TrimSpace(note), "")
}

func (s *Service) transition(ctx context.Context, action string, clientID uint64, actor Actor, note string, reason string) error {
	if actor.ID == 0 {
		return domain.ErrForbiddenOperation
	}

	role := normalizeActorRole(actor.Role)
	if !isValidActorRole(role) {
		return domain.ErrForbiddenOperation
	}

	toStatus, ok := domain.ActionTargetStatus(action)
	if !ok {
		return domain.ErrInvalidTransition
	}

	if !canRoleExecuteAction(role, action) {
		return domain.ErrForbiddenOperation
	}

	return s.tx.WithinTransaction(ctx, func(txCtx context.Context) error {
		client, err := s.repo.GetByIDForUpdate(txCtx, clientID)
		if err != nil {
			return err
		}
		fromStatus := client.Status

		if !domain.CanTransition(fromStatus, toStatus) {
			return domain.ErrInvalidTransition
		}

		if !canActorTouchClient(role, actor.ID, client, action) {
			return domain.ErrForbiddenOperation
		}

		client.Status = toStatus
		client.UpdatedBy = &actor.ID
		cleanActorName := strings.TrimSpace(actor.Name)

		switch toStatus {
		case domain.StatusDisetujui:
			client.ApprovedBy = &actor.ID
		case domain.StatusDitolak:
			client.RejectedBy = &actor.ID
			if reason != "" {
				client.RejectedReason = &reason
			}
		case domain.StatusDraft:
			client.ApprovedBy = nil
		}

		if _, err := s.repo.Update(txCtx, client); err != nil {
			return err
		}

		history := domain.StatusHistory{
			ClientID:   clientID,
			FromStatus: &fromStatus,
			ToStatus:   toStatus,
			Action:     action,
			ActorID:    &actor.ID,
		}
		if note != "" {
			history.Note = &note
		}
		if reason != "" {
			history.Reason = &reason
		}
		if cleanActorName != "" {
			history.ActorName = &cleanActorName
		}
		if err := s.repo.CreateHistory(txCtx, history); err != nil {
			return err
		}

		auditPayload := map[string]any{
			"from_status": fromStatus,
			"to_status":   toStatus,
			"action":      action,
		}
		if note != "" {
			auditPayload["note"] = note
		}
		if reason != "" {
			auditPayload["reason"] = reason
		}

		return s.repo.AppendAudit(txCtx, buildAuditEntry(actor, "CLIENT_"+strings.ToUpper(action), clientID, auditPayload))
	})
}

func normalizeActorRole(raw string) string {
	role := strings.ToUpper(strings.TrimSpace(raw))
	switch role {
	case "SUPER_ADMIN":
		return RoleAdmin
	case "PLANNER":
		return RolePerencana
	case "VERIFIER", "REVIEWER":
		return RoleVerifikator
	case "APPROVER":
		return RolePimpinan
	default:
		return role
	}
}
