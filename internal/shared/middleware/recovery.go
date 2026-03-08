package middleware

import (
	"log"

	"e-plan-ai/internal/shared/response"

	"github.com/gin-gonic/gin"
)

func Recovery() gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered any) {
		log.Printf("panic recovered: %v", recovered)
		response.Error(c, 500, "internal server error")
	})
}
