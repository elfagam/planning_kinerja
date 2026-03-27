package main

import (
	"log"
	"os"

	"e-plan-ai/internal/bootstrap"
	"e-plan-ai/internal/config"

	"github.com/gin-gonic/gin"
)

func main() {
	cfg := config.Load()
	gin.SetMode(cfg.GinMode)
	r := bootstrap.NewRouter(cfg)

	// 1. Ambil port dari Railway (Variable "PORT")
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080" // Fallback jika dijalankan di laptop sendiri
	}

	// 2. Jalankan di "0.0.0.0" (BUKAN localhost) agar bisa diakses dari luar kontainer
	// Cukup gunakan ":" + port
	log.Printf("e-plan-ai API running on port %s", port)
	if err := r.Run("0.0.0.0:" + port); err != nil {
		log.Fatalf("failed to run server: %v", err)
	}
}

