package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func performOperatorMiddlewareRequest(t *testing.T, authEnabled bool, method string, requestPath string) *httptest.ResponseRecorder {
	t.Helper()

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("auth.role", "OPERATOR")
		c.Next()
	})
	r.Use(OperatorReadOnly(authEnabled))
	r.Handle(method, requestPath, func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(method, requestPath, nil)
	r.ServeHTTP(w, req)
	return w
}

func TestOperatorReadOnly_BlocksPlanningWrite(t *testing.T) {
	w := performOperatorMiddlewareRequest(t, true, http.MethodPost, "/api/v1/program")
	if w.Code != http.StatusForbidden {
		t.Fatalf("expected status %d, got %d", http.StatusForbidden, w.Code)
	}
}

func TestOperatorReadOnly_AllowsClientWrite(t *testing.T) {
	w := performOperatorMiddlewareRequest(t, true, http.MethodPost, "/api/v1/clients")
	if w.Code != http.StatusNoContent {
		t.Fatalf("expected status %d, got %d", http.StatusNoContent, w.Code)
	}
}

func TestOperatorReadOnly_DisabledAuthSkipsBlock(t *testing.T) {
	w := performOperatorMiddlewareRequest(t, false, http.MethodPost, "/api/v1/program")
	if w.Code != http.StatusNoContent {
		t.Fatalf("expected status %d, got %d", http.StatusNoContent, w.Code)
	}
}