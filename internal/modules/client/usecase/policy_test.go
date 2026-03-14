package usecase

import (
	"testing"

	"e-plan-ai/internal/modules/client/domain"
)

func uint64Ptr(v uint64) *uint64 {
	return &v
}

func TestIsValidActorRole(t *testing.T) {
	roles := []string{RoleAdmin, RoleOperator, RolePerencana, RoleVerifikator, RolePimpinan}
	for _, role := range roles {
		if !isValidActorRole(role) {
			t.Fatalf("expected role %s to be valid", role)
		}
	}

	if isValidActorRole("UNKNOWN") {
		t.Fatalf("expected UNKNOWN role to be invalid")
	}
}

func TestCanRoleWriteClient(t *testing.T) {
	if !canRoleWriteClient(RoleAdmin) {
		t.Fatalf("expected admin to be able to write client")
	}
	if !canRoleWriteClient(RoleOperator) {
		t.Fatalf("expected operator to be able to write client")
	}
	if !canRoleWriteClient(RolePerencana) {
		t.Fatalf("expected perencana to be able to write client")
	}
	if canRoleWriteClient(RoleVerifikator) {
		t.Fatalf("expected verifikator to not be able to write client")
	}
	if canRoleWriteClient(RolePimpinan) {
		t.Fatalf("expected pimpinan to not be able to write client")
	}
}

func TestCanRoleExecuteAction(t *testing.T) {
	tests := []struct {
		name   string
		role   string
		action string
		want   bool
	}{
		{name: "admin can approve", role: RoleAdmin, action: "approve", want: true},
		{name: "admin can reject", role: RoleAdmin, action: "reject", want: true},
		{name: "operator can submit", role: RoleOperator, action: "submit", want: true},
		{name: "operator cannot reject", role: RoleOperator, action: "reject", want: false},
		{name: "perencana can re-evaluate", role: RolePerencana, action: "re-evaluate", want: true},
		{name: "verifikator can reject", role: RoleVerifikator, action: "reject", want: true},
		{name: "verifikator cannot approve", role: RoleVerifikator, action: "approve", want: false},
		{name: "pimpinan can approve", role: RolePimpinan, action: "approve", want: true},
		{name: "pimpinan cannot unsubmit", role: RolePimpinan, action: "unsubmit", want: false},
		{name: "unknown action denied", role: RoleAdmin, action: "unknown", want: false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := canRoleExecuteAction(tc.role, tc.action)
			if got != tc.want {
				t.Fatalf("canRoleExecuteAction(%s, %s) = %v, want %v", tc.role, tc.action, got, tc.want)
			}
		})
	}
}

func TestCanActorMutateClient(t *testing.T) {
	ownerID := uint64(10)
	otherID := uint64(20)

	draftOwned := domain.Client{Status: domain.StatusDraft, CreatedBy: uint64Ptr(ownerID)}
	draftOther := domain.Client{Status: domain.StatusDraft, CreatedBy: uint64Ptr(otherID)}
	nonDraftOwned := domain.Client{Status: domain.StatusDiajukan, CreatedBy: uint64Ptr(ownerID)}

	if !canActorMutateClient(RoleAdmin, ownerID, nonDraftOwned) {
		t.Fatalf("admin should be able to mutate non-draft client")
	}
	if !canActorMutateClient(RoleOperator, ownerID, draftOwned) {
		t.Fatalf("operator owner should be able to mutate draft client")
	}
	if canActorMutateClient(RoleOperator, ownerID, draftOther) {
		t.Fatalf("operator non-owner should not be able to mutate draft client")
	}
	if canActorMutateClient(RolePerencana, ownerID, nonDraftOwned) {
		t.Fatalf("perencana should not be able to mutate non-draft client")
	}
	if canActorMutateClient(RoleVerifikator, ownerID, draftOwned) {
		t.Fatalf("verifikator should not be able to mutate client")
	}
}

func TestCanActorTouchClient(t *testing.T) {
	ownerID := uint64(10)
	otherID := uint64(20)

	draftOwned := domain.Client{Status: domain.StatusDraft, CreatedBy: uint64Ptr(ownerID)}
	ditolakOwned := domain.Client{Status: domain.StatusDitolak, CreatedBy: uint64Ptr(ownerID)}
	draftOther := domain.Client{Status: domain.StatusDraft, CreatedBy: uint64Ptr(otherID)}

	if !canActorTouchClient(RoleOperator, ownerID, draftOwned, "submit") {
		t.Fatalf("operator owner should be able to submit own draft client")
	}
	if !canActorTouchClient(RolePerencana, ownerID, ditolakOwned, "re-evaluate") {
		t.Fatalf("perencana owner should be able to re-evaluate own rejected client")
	}
	if canActorTouchClient(RolePerencana, ownerID, draftOwned, "re-evaluate") {
		t.Fatalf("perencana should not be able to re-evaluate non-rejected client")
	}
	if canActorTouchClient(RoleOperator, ownerID, draftOther, "submit") {
		t.Fatalf("operator should not be able to submit other user's client")
	}
	if !canActorTouchClient(RoleVerifikator, ownerID, draftOwned, "reject") {
		t.Fatalf("verifikator should be allowed to touch client for reject flow")
	}
}
