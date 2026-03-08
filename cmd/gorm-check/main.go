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
			log.Fatalf("gorm connection failed: %v; %s", err, database.ConnectionFailureHint())
		}
		log.Fatalf("gorm connection failed: %v", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("unable to get sql db: %v", err)
	}
	defer sqlDB.Close()

	log.Println("GORM MySQL connection established")
}
