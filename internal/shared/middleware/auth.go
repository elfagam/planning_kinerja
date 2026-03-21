package middleware

import (
	"fmt"
	"net/http"
	"strings"

	"e-plan-ai/internal/shared/response"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	UserID    uint64 `json:"user_id"`
	Username  string `json:"username"`
	Email     string `json:"email"`
	FullName  string `json:"full_name"`
	Role      string `json:"role"`
	TokenType string `json:"token_type"`
	jwt.RegisteredClaims
}

// Auth validates Bearer JWT for protected routes when enabled.
func Auth(enabled bool, signingKey string) gin.HandlerFunc {
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
		if token == "" {
			response.Error(c, http.StatusUnauthorized, "invalid token")
			c.Abort()
			return
		}

		claims, err := parseToken(token, signingKey)
		if err != nil {
			response.Error(c, http.StatusUnauthorized, "invalid token")
			c.Abort()
			return
		}

		if claims.TokenType != "access" {
			response.Error(c, http.StatusUnauthorized, "invalid token type")
			c.Abort()
			return
		}

		c.Set("auth.username", claims.Username)
		c.Set("auth.user_id", claims.UserID)
		c.Set("auth.email", claims.Email)
		c.Set("auth.full_name", claims.FullName)
		c.Set("auth.role", claims.Role)
		c.Set("auth.subject", claims.Subject)
		c.Set("auth.token_type", claims.TokenType)

		c.Next()
	}
}

func parseToken(tokenString string, signingKey string) (*Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(signingKey), nil
	})
	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	return claims, nil
}
