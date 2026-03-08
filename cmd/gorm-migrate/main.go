package main

import (
	"log"

	"e-plan-ai/internal/config"
	"e-plan-ai/internal/shared/database"
)

func main() {
	cfg := config.Load()

	db, err := database.NewGormMySQL(cfg)
	if err != nil {
		if database.IsConnectionError(err) {
			log.Fatalf("gorm mysql connection failed: %v; %s", err, database.ConnectionFailureHint())
		}
		log.Fatalf("gorm mysql connection failed: %v", err)
	}

	if err := database.AutoMigrateAll(db); err != nil {
		log.Fatalf("gorm automigrate failed: %v", err)
	}

	log.Println("GORM auto migration completed")
}
