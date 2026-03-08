package main

import (
	"net/http"
	"os"

	"eplan/services/renja-service/internal/controller"
	"eplan/services/renja-service/internal/repository"
	"eplan/services/renja-service/internal/service"

	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"service": "renja-service", "status": "ok"})
	})

	repo := repository.InMemoryRenjaRepository{}
	svc := service.NewRenjaService(repo)
	ctrl := controller.NewRenjaController(svc)
	ctrl.Register(r)

	_ = r.Run(getenv("HTTP_ADDR", ":8082"))
}

func getenv(key, fallback string) string {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	return v
}
