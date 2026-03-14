package usecase

import (
	"context"
	"encoding/json"
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
