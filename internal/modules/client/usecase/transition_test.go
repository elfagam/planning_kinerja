package usecase

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"e-plan-ai/internal/modules/client/domain"
)

type fakeTxManager struct{}

func (f fakeTxManager) WithinTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	return fn(ctx)
}

type fakeRepo struct {
	client    domain.Client
	updated   bool
	histories []domain.StatusHistory
	audits    []AuditLogEntry
	auditLogs []AuditLog
}

func (f *fakeRepo) List(ctx context.Context, filter ListFilter) ([]domain.Client, int64, error) {
	return []domain.Client{f.client}, 1, nil
}

func (f *fakeRepo) ListAuditLogs(ctx context.Context, filter AuditListFilter) ([]AuditLog, int64, error) {
	return f.auditLogs, int64(len(f.auditLogs)), nil
}

func (f *fakeRepo) GetByID(ctx context.Context, id uint64) (domain.Client, error) {
	if id != f.client.ID {
		return domain.Client{}, domain.ErrClientNotFound
	}
	return f.client, nil
}

func (f *fakeRepo) GetByIDForUpdate(ctx context.Context, id uint64) (domain.Client, error) {
	return f.GetByID(ctx, id)
}

func (f *fakeRepo) Create(ctx context.Context, client domain.Client) (domain.Client, error) {
	f.client = client
	return f.client, nil
}

func (f *fakeRepo) Update(ctx context.Context, client domain.Client) (domain.Client, error) {
	f.updated = true
	f.client = client
	return client, nil
}

func (f *fakeRepo) SoftDelete(ctx context.Context, id uint64) error {
	return nil
}

func (f *fakeRepo) CreateHistory(ctx context.Context, history domain.StatusHistory) error {
	f.histories = append(f.histories, history)
	return nil
}

func (f *fakeRepo) ListHistory(ctx context.Context, clientID uint64) ([]domain.StatusHistory, error) {
	return f.histories, nil
}

func (f *fakeRepo) AppendAudit(ctx context.Context, entry AuditLogEntry) error {
	f.audits = append(f.audits, entry)
	return nil
}

func TestSubmit_ForbiddenRole_ReturnsForbidden(t *testing.T) {
	creatorID := uint64(10)
	repo := &fakeRepo{client: domain.Client{ID: 1, Status: domain.StatusDraft, CreatedBy: &creatorID, CreatedAt: time.Now()}}
	svc := NewService(fakeTxManager{}, repo)

	err := svc.Submit(context.Background(), 1, Actor{ID: 10, Role: RoleVerifikator, Name: "Verifier"}, "submit")
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if err != domain.ErrForbiddenOperation {
		t.Fatalf("expected ErrForbiddenOperation, got %v", err)
	}
	if repo.updated {
		t.Fatalf("expected no update when forbidden")
	}
}

func TestSubmit_InvalidTransition_ReturnsConflictError(t *testing.T) {
	creatorID := uint64(10)
	repo := &fakeRepo{client: domain.Client{ID: 1, Status: domain.StatusDitolak, CreatedBy: &creatorID, CreatedAt: time.Now()}}
	svc := NewService(fakeTxManager{}, repo)

	err := svc.Submit(context.Background(), 1, Actor{ID: 10, Role: RoleOperator, Name: "Operator"}, "submit")
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if err != domain.ErrInvalidTransition {
		t.Fatalf("expected ErrInvalidTransition, got %v", err)
	}
	if repo.updated {
		t.Fatalf("expected no update when transition invalid")
	}
}

func TestSubmit_AppendsAuditPayload(t *testing.T) {
	creatorID := uint64(10)
	repo := &fakeRepo{client: domain.Client{ID: 1, Status: domain.StatusDraft, CreatedBy: &creatorID, CreatedAt: time.Now()}}
	svc := NewService(fakeTxManager{}, repo)

	err := svc.Submit(context.Background(), 1, Actor{ID: 10, Role: RoleOperator, Name: "Operator", IPAddress: "10.0.0.7", UserAgent: "api-test"}, "catatan submit")
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if len(repo.audits) != 1 {
		t.Fatalf("expected 1 audit entry, got %d", len(repo.audits))
	}

	audit := repo.audits[0]
	if audit.Action != "CLIENT_SUBMIT" {
		t.Fatalf("expected action CLIENT_SUBMIT, got %s", audit.Action)
	}
	if audit.ResourceType != clientResourceType {
		t.Fatalf("expected resource type %s, got %s", clientResourceType, audit.ResourceType)
	}
	if audit.ResourceID == nil || *audit.ResourceID != 1 {
		t.Fatalf("expected resource id 1, got %v", audit.ResourceID)
	}
	if audit.IPAddress == nil || *audit.IPAddress != "10.0.0.7" {
		t.Fatalf("expected ip 10.0.0.7, got %v", audit.IPAddress)
	}
	if audit.UserAgent == nil || *audit.UserAgent != "api-test" {
		t.Fatalf("expected user agent api-test, got %v", audit.UserAgent)
	}
	if audit.RequestPayload == nil {
		t.Fatalf("expected request payload, got nil")
	}

	var payload map[string]any
	if err := json.Unmarshal([]byte(*audit.RequestPayload), &payload); err != nil {
		t.Fatalf("failed to unmarshal payload: %v", err)
	}
	if payload["action"] != "submit" {
		t.Fatalf("expected action submit in payload, got %v", payload["action"])
	}
	if payload["from_status"] != string(domain.StatusDraft) {
		t.Fatalf("expected from_status %s, got %v", domain.StatusDraft, payload["from_status"])
	}
	if payload["to_status"] != string(domain.StatusDiajukan) {
		t.Fatalf("expected to_status %s, got %v", domain.StatusDiajukan, payload["to_status"])
	}
	if payload["note"] != "catatan submit" {
		t.Fatalf("expected note catatan submit, got %v", payload["note"])
	}
}
