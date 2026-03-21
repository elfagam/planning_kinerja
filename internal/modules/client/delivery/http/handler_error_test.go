package http

import (
	"net/http/httptest"
	"testing"

	"e-plan-ai/internal/modules/client/domain"

	"github.com/gin-gonic/gin"
)

func TestHandleError_ForbiddenMapsTo403(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(r)

	h := &Handler{}
	h.handleError(c, domain.ErrForbiddenOperation)

	if r.Code != 403 {
		t.Fatalf("expected 403, got %d", r.Code)
	}
}

func TestHandleError_InvalidTransitionMapsTo409(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(r)

	h := &Handler{}
	h.handleError(c, domain.ErrInvalidTransition)

	if r.Code != 409 {
		t.Fatalf("expected 409, got %d", r.Code)
	}
}
