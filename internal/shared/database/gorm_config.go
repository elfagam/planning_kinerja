package database

import (
	"time"

	"e-plan-ai/internal/config"

	"gorm.io/gorm/logger"
)

// GormMySQLConfig centralizes GORM MySQL connection settings.
type GormMySQLConfig struct {
	DSN             string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	ConnMaxIdleTime time.Duration
	LogMode         logger.LogLevel
}

// NewGormMySQLConfig builds GORM DB config from app config and optional env overrides.
func NewGormMySQLConfig(cfg config.Config) GormMySQLConfig {
	logMode := logger.Warn
	if cfg.AppEnv == "production" {
		logMode = logger.Error
	}

	maxOpen := cfg.DBMaxOpenConns
	if maxOpen <= 0 {
		maxOpen = 20
	}

	maxIdle := cfg.DBMaxIdleConns
	if maxIdle < 0 {
		maxIdle = 10
	}
	if maxIdle > maxOpen {
		maxIdle = maxOpen
	}

	lifeMinutes := cfg.DBConnMaxLifetimeMins
	if lifeMinutes <= 0 {
		lifeMinutes = 30
	}

	idleMinutes := cfg.DBConnMaxIdleTimeMins
	if idleMinutes <= 0 {
		idleMinutes = 10
	}

	return GormMySQLConfig{
		DSN:             cfg.MySQLDSN,
		MaxOpenConns:    maxOpen,
		MaxIdleConns:    maxIdle,
		ConnMaxLifetime: time.Duration(lifeMinutes) * time.Minute,
		ConnMaxIdleTime: time.Duration(idleMinutes) * time.Minute,
		LogMode:         logMode,
	}
}
