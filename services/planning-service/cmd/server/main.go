package main

import (
	"net/http"
	"os"

	"eplan/services/planning-service/internal/controller"
	"eplan/services/planning-service/internal/repository"
	"eplan/services/planning-service/internal/service"

	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"service": "planning-service", "status": "ok"})
	})

	repo := repository.InMemoryStrategicRepository{}
	svc := service.NewStrategicService(repo)
	ctrl := controller.NewStrategicController(svc)
	ctrl.Register(r)

	_ = r.Run(getenv("HTTP_ADDR", ":8081"))
}

func getenv(key, fallback string) string {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	return v
}
