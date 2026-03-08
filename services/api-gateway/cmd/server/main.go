package main

import (
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"service": "api-gateway", "status": "ok"})
	})

	r.GET("/v1/planning/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"target_service": "planning-service", "status": "proxy-placeholder"})
	})

	r.GET("/v1/renja/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"target_service": "renja-service", "status": "proxy-placeholder"})
	})

	r.GET("/v1/performance/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"target_service": "performance-service", "status": "proxy-placeholder"})
	})

	_ = r.Run(getenv("HTTP_ADDR", ":8080"))
}

func getenv(key, fallback string) string {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	return v
}
