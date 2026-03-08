package main

import (
	"log"

	"e-plan-ai/internal/bootstrap"
	"e-plan-ai/internal/config"

	"github.com/gin-gonic/gin"
)

func main() {
	cfg := config.Load()
	gin.SetMode(cfg.GinMode)
	r := bootstrap.NewRouter(cfg)

	log.Printf("e-plan-ai API running on %s", cfg.HTTPAddr)
	if err := r.Run(cfg.HTTPAddr); err != nil {
		log.Fatalf("failed to run server: %v", err)
	}
}
