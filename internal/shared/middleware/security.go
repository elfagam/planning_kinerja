package middleware

import (
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// Security middleware blocks common bot scanners and malicious patterns
func Security() gin.HandlerFunc {
	// List of forbidden patterns that indicate automated scanning
	forbiddenPatterns := []string{
		"/wordpress",
		"/wp-admin",
		"/wp-content",
		"/wp-includes",
		"/setup-config.php",
		"wp-login.php",
		"xmlrpc.php",
		"wlwmanifest.xml",
		"autodiscover.xml",
		"/phpmyadmin",
		"/.env",
		"/.git",
		"/.vscode",
		"/.aws",
		"/.ssh",
		"/.bash_history",
		"/debug/default/view",
		"/invoker/",
		"/jmx-console/",
		"/web-console/",
		"/console/",
		"/actuator/",
		"/metrics",
	}

	forbiddenExtensions := []string{
		".php",
		".asp",
		".aspx",
		".jsp",
		".jspx",
		".cgi",
		".pl",
		".py",
		".sh",
		".bat",
	}

	return func(c *gin.Context) {
		path := strings.ToLower(c.Request.URL.Path)

		// Check for forbidden keywords/paths
		for _, pattern := range forbiddenPatterns {
			if strings.Contains(path, strings.ToLower(pattern)) {
				log.Printf("[SECURITY BLOCK] Path pattern match: %s from IP: %s", path, c.ClientIP())
				c.AbortWithStatus(http.StatusNotFound) 
				return
			}
		}

		// Check for forbidden extensions
		for _, ext := range forbiddenExtensions {
			if strings.HasSuffix(path, ext) {
				log.Printf("[SECURITY BLOCK] Extension match: %s from IP: %s", path, c.ClientIP())
				c.AbortWithStatus(http.StatusNotFound)
				return
			}
		}

		c.Next()
	}
}
