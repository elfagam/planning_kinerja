package usecase

import (
	"context"
	"fmt"
	"strings"

	"e-plan-ai/internal/modules/client/domain"
)

func (s *Service) ListAuditLogs(ctx context.Context, actor Actor, filter AuditListFilter) ([]AuditLog, int64, error) {
	role := normalizeActorRole(actor.Role)
	if actor.ID == 0 || !isValidActorRole(role) {
		return nil, 0, domain.ErrForbiddenOperation
	}
	if !canRoleReadClientAudit(role) {
		return nil, 0, domain.ErrForbiddenOperation
	}

	filter.Action = strings.ToUpper(strings.TrimSpace(filter.Action))
	if filter.Page <= 0 {
		filter.Page = 1
	}
	if filter.Limit <= 0 {
		filter.Limit = 10
	}
	if filter.Limit > 100 {
		filter.Limit = 100
	}

	if filter.Action != "" && !strings.HasPrefix(filter.Action, "CLIENT_") {
		return nil, 0, fmt.Errorf("%w: action must start with CLIENT_", domain.ErrValidation)
	}

	return s.repo.ListAuditLogs(ctx, filter)
}
