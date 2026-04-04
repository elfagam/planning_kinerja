package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
)

// StaticCache provides a middleware to add Cache-Control headers for static assets.
// This is specifically optimized for PDF files which are immutable (named with unique timestamps).
func StaticCache() gin.HandlerFunc {
	return func(c *gin.Context) {
		path := c.Request.URL.Path

		// Only apply to PDF files in the assets/pdf directory
		if strings.HasSuffix(strings.ToLower(path), ".pdf") && strings.Contains(path, "/assets/pdf/") {
			// public: can be cached by browser and CDNs
			// max-age: 1 year (31536000 seconds)
			// immutable: file will never change, skip revalidation
			c.Header("Cache-Control", "public, max-age=31536000, immutable")
		}

		c.Next()
	}
}
