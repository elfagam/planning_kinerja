package database

import (
	"fmt"

	"e-plan-ai/internal/config"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// NewGormMySQL creates a GORM connection backed by MySQL.
func NewGormMySQL(cfg config.Config) (*gorm.DB, error) {
	dbCfg := NewGormMySQLConfig(cfg)
	gormCfg := &gorm.Config{Logger: logger.Default.LogMode(dbCfg.LogMode)}

	db, err := gorm.Open(mysql.Open(dbCfg.DSN), gormCfg)
	if err != nil {
		return nil, WrapConnectionError("gorm open", fmt.Errorf("open gorm mysql: %w", err))
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("extract sql db from gorm: %w", err)
	}

	sqlDB.SetMaxOpenConns(dbCfg.MaxOpenConns)
	sqlDB.SetMaxIdleConns(dbCfg.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(dbCfg.ConnMaxLifetime)
	sqlDB.SetConnMaxIdleTime(dbCfg.ConnMaxIdleTime)
	if err := sqlDB.Ping(); err != nil {
		return nil, WrapConnectionError("gorm ping", fmt.Errorf("ping gorm mysql: %w", err))
	}

	return db, nil
}
