package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
)

type DevelopmentActorContext struct {
	UserID   uint64
	Email    string
	FullName string
	Role     string
}

func (a DevelopmentActorContext) Enabled() bool {
	return a.UserID > 0 && strings.TrimSpace(a.Role) != ""
}

func DevelopmentActor(authEnabled bool, actor DevelopmentActorContext) gin.HandlerFunc {
	return func(c *gin.Context) {
		if authEnabled || !actor.Enabled() {
			c.Next()
			return
		}

		if _, exists := c.Get("auth.user_id"); !exists {
			c.Set("auth.user_id", actor.UserID)
		}
		if _, exists := c.Get("auth.email"); !exists && strings.TrimSpace(actor.Email) != "" {
			c.Set("auth.email", actor.Email)
		}
		if _, exists := c.Get("auth.username"); !exists {
			username := strings.TrimSpace(actor.Email)
			if username == "" {
				username = strings.TrimSpace(actor.FullName)
			}
			if username != "" {
				c.Set("auth.username", username)
			}
		}
		if _, exists := c.Get("auth.full_name"); !exists && strings.TrimSpace(actor.FullName) != "" {
			c.Set("auth.full_name", actor.FullName)
		}
		if _, exists := c.Get("auth.role"); !exists {
			c.Set("auth.role", actor.Role)
		}
		if _, exists := c.Get("auth.subject"); !exists {
			c.Set("auth.subject", actor.Email)
		}
		if _, exists := c.Get("auth.token_type"); !exists {
			c.Set("auth.token_type", "development")
		}

		c.Next()
	}
}
