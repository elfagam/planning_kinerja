package middleware

import (
	"net/http"
	"strings"

	"e-plan-ai/internal/shared/response"

	"github.com/gin-gonic/gin"
)

// Auth validates Bearer token for protected routes when enabled.
func Auth(enabled bool, expectedToken string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !enabled {
			c.Next()
			return
		}

		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			response.Error(c, http.StatusUnauthorized, "missing authorization header")
			c.Abort()
			return
		}

		const prefix = "Bearer "
		if !strings.HasPrefix(authHeader, prefix) {
			response.Error(c, http.StatusUnauthorized, "invalid authorization scheme")
			c.Abort()
			return
		}

		token := strings.TrimPrefix(authHeader, prefix)
		if token == "" || token != expectedToken {
			response.Error(c, http.StatusUnauthorized, "invalid token")
			c.Abort()
			return
		}

		c.Next()
	}
}
