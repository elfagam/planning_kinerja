package domain

import "testing"

func TestRencanaKerja_AjukanFromDraft(t *testing.T) {
	r := RencanaKerja{Status: StatusDraft}
	if err := r.Ajukan(11); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if r.Status != StatusSubmitted {
		t.Fatalf("expected status %q, got %q", StatusSubmitted, r.Status)
	}
	if r.UpdatedBy != 11 {
		t.Fatalf("expected updated_by=11, got %d", r.UpdatedBy)
	}
}

func TestRencanaKerja_SetujuiRequiresSubmitted(t *testing.T) {
	r := RencanaKerja{Status: StatusDraft}
	if err := r.Setujui(12); err != ErrInvalidTransition {
		t.Fatalf("expected ErrInvalidTransition, got %v", err)
	}
}

func TestRencanaKerja_TolakRequiresReason(t *testing.T) {
	r := RencanaKerja{Status: StatusSubmitted}
	if err := r.Tolak(13, ""); err != ErrRejectionReasonMissing {
		t.Fatalf("expected ErrRejectionReasonMissing, got %v", err)
	}
}

func TestRencanaKerja_BackwardCompatibleMethods(t *testing.T) {
	r := RencanaKerja{Status: StatusDraft}
	if err := r.Submit(21); err != nil {
		t.Fatalf("submit should succeed: %v", err)
	}
	if err := r.Approve(22); err != nil {
		t.Fatalf("approve should succeed: %v", err)
	}
	if r.Status != StatusApproved {
		t.Fatalf("expected status %q, got %q", StatusApproved, r.Status)
	}
}
