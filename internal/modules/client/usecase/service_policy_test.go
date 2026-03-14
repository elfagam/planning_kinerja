package usecase

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"e-plan-ai/internal/modules/client/domain"
)

func TestUpdate_OperatorNonDraft_Forbidden(t *testing.T) {
	creatorID := uint64(10)
	repo := &fakeRepo{client: domain.Client{ID: 1, Status: domain.StatusDiajukan, CreatedBy: &creatorID, CreatedAt: time.Now()}}
	svc := NewService(fakeTxManager{}, repo)

	_, err := svc.Update(context.Background(), 1, Actor{ID: 10, Role: RoleOperator}, domain.Client{Nama: "Revisi"})
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if err != domain.ErrForbiddenOperation {
		t.Fatalf("expected ErrForbiddenOperation, got %v", err)
	}
	if repo.updated {
		t.Fatalf("expected no update when client is not draft")
	}
}

func TestDelete_OperatorDraftOwner_Allowed(t *testing.T) {
	creatorID := uint64(10)
	repo := &fakeRepo{client: domain.Client{ID: 1, Status: domain.StatusDraft, CreatedBy: &creatorID, CreatedAt: time.Now()}}
	svc := NewService(fakeTxManager{}, repo)

	err := svc.Delete(context.Background(), 1, Actor{ID: 10, Role: RoleOperator, IPAddress: "127.0.0.1", UserAgent: "go-test"})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if len(repo.audits) != 1 {
		t.Fatalf("expected 1 audit entry, got %d", len(repo.audits))
	}

	audit := repo.audits[0]
	if audit.Action != "CLIENT_DELETE" {
		t.Fatalf("expected action CLIENT_DELETE, got %s", audit.Action)
	}
	if audit.ResourceType != clientResourceType {
		t.Fatalf("expected resource type %s, got %s", clientResourceType, audit.ResourceType)
	}
	if audit.ResourceID == nil || *audit.ResourceID != 1 {
		t.Fatalf("expected resource id 1, got %v", audit.ResourceID)
	}
	if audit.IPAddress == nil || *audit.IPAddress != "127.0.0.1" {
		t.Fatalf("expected ip 127.0.0.1, got %v", audit.IPAddress)
	}
	if audit.UserAgent == nil || *audit.UserAgent != "go-test" {
		t.Fatalf("expected user agent go-test, got %v", audit.UserAgent)
	}

	if audit.RequestPayload == nil {
		t.Fatalf("expected request payload, got nil")
	}
	var payload map[string]any
	if err := json.Unmarshal([]byte(*audit.RequestPayload), &payload); err != nil {
		t.Fatalf("failed to unmarshal payload: %v", err)
	}
	if payload["kode"] != "" {
		t.Fatalf("expected empty kode in payload, got %v", payload["kode"])
	}
	if payload["nama"] != "" {
		t.Fatalf("expected empty nama in payload, got %v", payload["nama"])
	}
	if payload["status"] != string(domain.StatusDraft) {
		t.Fatalf("expected status %s, got %v", domain.StatusDraft, payload["status"])
	}
}

func TestDelete_OperatorNonOwner_Forbidden(t *testing.T) {
	creatorID := uint64(20)
	repo := &fakeRepo{client: domain.Client{ID: 1, Status: domain.StatusDraft, CreatedBy: &creatorID, CreatedAt: time.Now()}}
	svc := NewService(fakeTxManager{}, repo)

	err := svc.Delete(context.Background(), 1, Actor{ID: 10, Role: RoleOperator})
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if err != domain.ErrForbiddenOperation {
		t.Fatalf("expected ErrForbiddenOperation, got %v", err)
	}
}

func TestListAuditLogs_AdminAllowed(t *testing.T) {
	repo := &fakeRepo{
		auditLogs: []AuditLog{{ID: 11, Action: "CLIENT_CREATE", ResourceType: "CLIENT", CreatedAt: time.Now().Format(time.RFC3339)}},
	}
	svc := NewService(fakeTxManager{}, repo)

	items, total, err := svc.ListAuditLogs(context.Background(), Actor{ID: 1, Role: RoleAdmin}, AuditListFilter{Page: 1, Limit: 10})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if total != 1 {
		t.Fatalf("expected total 1, got %d", total)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}
	if items[0].Action != "CLIENT_CREATE" {
		t.Fatalf("expected action CLIENT_CREATE, got %s", items[0].Action)
	}
}

func TestListAuditLogs_OperatorForbidden(t *testing.T) {
	repo := &fakeRepo{}
	svc := NewService(fakeTxManager{}, repo)

	_, _, err := svc.ListAuditLogs(context.Background(), Actor{ID: 10, Role: RoleOperator}, AuditListFilter{Page: 1, Limit: 10})
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if err != domain.ErrForbiddenOperation {
		t.Fatalf("expected ErrForbiddenOperation, got %v", err)
	}
}

func TestListAuditLogs_InvalidActionValidation(t *testing.T) {
	repo := &fakeRepo{}
	svc := NewService(fakeTxManager{}, repo)

	_, _, err := svc.ListAuditLogs(context.Background(), Actor{ID: 1, Role: RoleAdmin}, AuditListFilter{Action: "DELETE", Page: 1, Limit: 10})
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("expected ErrValidation, got %v", err)
	}
}
