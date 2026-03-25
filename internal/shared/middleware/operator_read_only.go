package middleware

import (
	"fmt"
	"net/http"
	"strings"

	"e-plan-ai/internal/shared/response"

	"github.com/gin-gonic/gin"
)

// OperatorReadOnly intercepts and blocks mutating HTTP requests (POST, PUT, DELETE, PATCH)
// if the authenticated user has a restricted role (e.g., OPERATOR, PERENCANA, VERIFIKATOR).
// Explicit exceptions are made for endpoints they are authorized to modify.
func OperatorReadOnly(authEnabled bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !authEnabled {
			c.Next()
			return
		}

		method := c.Request.Method
		if method == http.MethodGet || method == http.MethodOptions || method == http.MethodHead {
			c.Next()
			return
		}

		rawRole, ok := c.Get("auth.role")
		if !ok {
			c.Next()
			return
		}

		role := strings.ToUpper(strings.TrimSpace(fmt.Sprintf("%v", rawRole)))

		// Admins and Pimpinan are fully unrestricted
		if role == "ADMIN" || role == "PIMPINAN" {
			c.Next()
			return
		}

		// Role is restricted, secure the mutation pathways
		path := c.Request.URL.Path

		allowedPrefixes := []string{
			"/api/v1/rencana_kerja",
			"/api/v1/indikator_rencana_kerja",
			"/api/v1/realisasi_rencana_kerja",
			"/api/v1/target-realisasi",
			"/api/v1/indikator-kinerja",
			"/api/v1/clients",
			"/api/v1/dokumen_pdf",
			"/api/v1/performance/target-realisasi",
			"/api/v1/performance/calculate-achievement",
		}

		isAllowed := false
		for _, prefix := range allowedPrefixes {
			if strings.HasPrefix(path, prefix) {
				isAllowed = true
				break
			}
		}

		if !isAllowed {
			response.Error(c, http.StatusForbidden, "role anda dilarang untuk mengubah data pada endpoint ini")
			c.Abort()
			return
		}

		c.Next()
	}
}
