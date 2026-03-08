package main

import (
	"net/http"
	"os"

	"eplan/services/performance-service/internal/controller"
	"eplan/services/performance-service/internal/repository"
	"eplan/services/performance-service/internal/service"

	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"service": "performance-service", "status": "ok"})
	})

	repo := repository.InMemoryPerformanceRepository{}
	svc := service.NewPerformanceService(repo)
	ctrl := controller.NewPerformanceController(svc)
	ctrl.Register(r)

	_ = r.Run(getenv("HTTP_ADDR", ":8083"))
}

func getenv(key, fallback string) string {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	return v
}
