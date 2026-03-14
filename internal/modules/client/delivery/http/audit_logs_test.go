package http

import (
	"context"
	"encoding/json"
	"net/http/httptest"
	"testing"

	"e-plan-ai/internal/modules/client/domain"
	"e-plan-ai/internal/modules/client/usecase"

	"github.com/gin-gonic/gin"
)

type auditTestTxManager struct{}

func (auditTestTxManager) WithinTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	return fn(ctx)
}

type auditTestRepo struct {
	auditLogs []usecase.AuditLog
}

func (r *auditTestRepo) List(ctx context.Context, filter usecase.ListFilter) ([]domain.Client, int64, error) {
	return nil, 0, nil
}

func (r *auditTestRepo) ListAuditLogs(ctx context.Context, filter usecase.AuditListFilter) ([]usecase.AuditLog, int64, error) {
	return r.auditLogs, int64(len(r.auditLogs)), nil
}

func (r *auditTestRepo) GetByID(ctx context.Context, id uint64) (domain.Client, error) {
	return domain.Client{}, domain.ErrClientNotFound
}

func (r *auditTestRepo) GetByIDForUpdate(ctx context.Context, id uint64) (domain.Client, error) {
	return domain.Client{}, domain.ErrClientNotFound
}

func (r *auditTestRepo) Create(ctx context.Context, client domain.Client) (domain.Client, error) {
	return domain.Client{}, nil
}

func (r *auditTestRepo) Update(ctx context.Context, client domain.Client) (domain.Client, error) {
	return client, nil
}

func (r *auditTestRepo) SoftDelete(ctx context.Context, id uint64) error {
	return nil
}

func (r *auditTestRepo) CreateHistory(ctx context.Context, history domain.StatusHistory) error {
	return nil
}

func (r *auditTestRepo) ListHistory(ctx context.Context, clientID uint64) ([]domain.StatusHistory, error) {
	return nil, nil
}

func (r *auditTestRepo) AppendAudit(ctx context.Context, entry usecase.AuditLogEntry) error {
	return nil
}

func TestAuditLogs_AdminAllowed_Returns200(t *testing.T) {
	gin.SetMode(gin.TestMode)

	repo := &auditTestRepo{
		auditLogs: []usecase.AuditLog{{
			ID:           1,
			Action:       "CLIENT_CREATE",
			ResourceType: "CLIENT",
			CreatedAt:    "2026-03-14T12:00:00Z",
		}},
	}
	h := &Handler{service: usecase.NewService(auditTestTxManager{}, repo)}

	r := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(r)
	c.Request = httptest.NewRequest("GET", "/api/v1/clients/audit-logs?action=client_create&page=1&limit=10", nil)
	c.Set("auth.user_id", uint64(1))
	c.Set("auth.role", "ADMIN")
	c.Set("auth.full_name", "Admin")

	h.AuditLogs(c)

	if r.Code != 200 {
		t.Fatalf("expected 200, got %d", r.Code)
	}

	var body map[string]any
	if err := json.Unmarshal(r.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to decode response body: %v", err)
	}
	if body["success"] != true {
		t.Fatalf("expected success=true, got %v", body["success"])
	}
}

func TestAuditLogs_NonAdmin_Forbidden403(t *testing.T) {
	gin.SetMode(gin.TestMode)

	repo := &auditTestRepo{}
	h := &Handler{service: usecase.NewService(auditTestTxManager{}, repo)}

	r := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(r)
	c.Request = httptest.NewRequest("GET", "/api/v1/clients/audit-logs?page=1&limit=10", nil)
	c.Set("auth.user_id", uint64(10))
	c.Set("auth.role", "OPERATOR")
	c.Set("auth.full_name", "Operator")

	h.AuditLogs(c)

	if r.Code != 403 {
		t.Fatalf("expected 403, got %d", r.Code)
	}
}

func TestAuditLogs_InvalidUserID_BadRequest400(t *testing.T) {
	gin.SetMode(gin.TestMode)

	repo := &auditTestRepo{}
	h := &Handler{service: usecase.NewService(auditTestTxManager{}, repo)}

	r := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(r)
	c.Request = httptest.NewRequest("GET", "/api/v1/clients/audit-logs?user_id=abc", nil)
	c.Set("auth.user_id", uint64(1))
	c.Set("auth.role", "ADMIN")
	c.Set("auth.full_name", "Admin")

	h.AuditLogs(c)

	if r.Code != 400 {
		t.Fatalf("expected 400, got %d", r.Code)
	}
}

func TestAuditLogs_InvalidResourceID_BadRequest400(t *testing.T) {
	gin.SetMode(gin.TestMode)

	repo := &auditTestRepo{}
	h := &Handler{service: usecase.NewService(auditTestTxManager{}, repo)}

	r := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(r)
	c.Request = httptest.NewRequest("GET", "/api/v1/clients/audit-logs?resource_id=0", nil)
	c.Set("auth.user_id", uint64(1))
	c.Set("auth.role", "ADMIN")
	c.Set("auth.full_name", "Admin")

	h.AuditLogs(c)

	if r.Code != 400 {
		t.Fatalf("expected 400, got %d", r.Code)
	}
}
