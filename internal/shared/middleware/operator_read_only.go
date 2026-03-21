package middleware

import (
	"github.com/gin-gonic/gin"
)

// OperatorReadOnly is kept for backward compatibility and currently acts as a no-op.
func OperatorReadOnly(authEnabled bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
	}
}
