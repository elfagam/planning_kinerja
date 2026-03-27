package http

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"e-plan-ai/internal/config"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestCreateIndikatorKegiatan_RequiredValidation(t *testing.T) {
	gin.SetMode(gin.TestMode)
	cfg := config.Config{ /* ...isi config test... */ }
	// Handler setup
	h := NewHandler(&cfg)
	if !h.ready {
		t.Skip("Handler not ready: " + h.reason)
	}

	router := gin.New()
	v1 := router.Group("/api/v1")
	h.RegisterRoutes(v1)

	// Payload tanpa indikator_program_id
	payload := map[string]interface{}{
		"kegiatan_id": 1,
		"kode": "IK-001",
		"nama": "Indikator Test",
	}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/indikator_kegiatan", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusBadRequest, resp.Code)
	assert.Contains(t, resp.Body.String(), "indikator_program_id is required")	
}