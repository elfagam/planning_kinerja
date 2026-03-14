package usecase

import "e-plan-ai/internal/modules/client/domain"

var validActorRoles = map[string]struct{}{
	RoleAdmin:       {},
	RoleOperator:    {},
	RolePerencana:   {},
	RoleVerifikator: {},
	RolePimpinan:    {},
}

var writerRoles = map[string]struct{}{
	RoleAdmin:     {},
	RoleOperator:  {},
	RolePerencana: {},
}

var auditReaderRoles = map[string]struct{}{
	RoleAdmin: {},
}

var actionRolePolicy = map[string]map[string]struct{}{
	"submit": {
		RoleAdmin:     {},
		RoleOperator:  {},
		RolePerencana: {},
	},
	"unsubmit": {
		RoleAdmin:     {},
		RoleOperator:  {},
		RolePerencana: {},
	},
	"re-evaluate": {
		RoleAdmin:     {},
		RoleOperator:  {},
		RolePerencana: {},
	},
	"reject": {
		RoleAdmin:       {},
		RoleVerifikator: {},
	},
	"approve": {
		RoleAdmin:    {},
		RolePimpinan: {},
	},
}

func isValidActorRole(role string) bool {
	_, ok := validActorRoles[role]
	return ok
}

func canRoleWriteClient(role string) bool {
	_, ok := writerRoles[role]
	return ok
}

func canRoleReadClientAudit(role string) bool {
	_, ok := auditReaderRoles[role]
	return ok
}

func canRoleExecuteAction(role string, action string) bool {
	allowedRoles, ok := actionRolePolicy[action]
	if !ok {
		return false
	}
	_, allowed := allowedRoles[role]
	return allowed
}

func canActorMutateClient(role string, actorID uint64, client domain.Client) bool {
	if role == RoleAdmin {
		return true
	}

	if role != RoleOperator && role != RolePerencana {
		return false
	}

	if client.Status != domain.StatusDraft {
		return false
	}

	if client.CreatedBy == nil {
		return false
	}

	return *client.CreatedBy == actorID
}

func canActorTouchClient(role string, actorID uint64, client domain.Client, action string) bool {
	if role == RoleAdmin {
		return true
	}

	if role == RoleOperator || role == RolePerencana {
		if client.CreatedBy == nil {
			return false
		}
		if *client.CreatedBy != actorID {
			return false
		}
		if action == "submit" || action == "unsubmit" {
			return true
		}
		if action == "re-evaluate" {
			return client.Status == domain.StatusDitolak
		}
		return false
	}

	return true
}
