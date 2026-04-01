package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"e-plan-ai/internal/config"

	_ "github.com/go-sql-driver/mysql"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// NewGormMySQL creates a GORM connection backed by MySQL.
func NewGormMySQL(cfg *config.Config) (*gorm.DB, error) {
	if cfg.MySQLDSN == "" {
		return nil, fmt.Errorf("MySQL DSN is empty, check your environment variables")
	}
	dbCfg := NewGormMySQLConfig(cfg)
	gormCfg := &gorm.Config{Logger: logger.Default.LogMode(dbCfg.LogMode)}

	maxRetries := cfg.DBConnectMaxRetries
	if maxRetries < 0 {
		maxRetries = 0
	}

	retryDelay := time.Duration(cfg.DBConnectRetryDelaySecs) * time.Second
	if retryDelay <= 0 {
		retryDelay = time.Second
	}

	var lastErr error
	for attempt := 0; attempt <= maxRetries; attempt++ {
		db, err := openAndPingGorm(dbCfg, gormCfg)
		if err == nil {
			return db, nil
		}

		lastErr = err
		if attempt < maxRetries {
			time.Sleep(retryDelay)
		}
	}

	return nil, lastErr
}

func openAndPingGorm(dbCfg GormMySQLConfig, gormCfg *gorm.Config) (*gorm.DB, error) {
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
		_ = sqlDB.Close()
		return nil, WrapConnectionError("gorm ping", fmt.Errorf("ping gorm mysql: %w", err))
	}

	return db, nil
}

// PingMySQL checks database availability without creating a long-lived pool.
func PingMySQL(cfg *config.Config) error {
	db, err := sql.Open("mysql", cfg.MySQLDSN)
	if err != nil {
		return WrapConnectionError("sql open", fmt.Errorf("open mysql: %w", err))
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return WrapConnectionError("sql ping", fmt.Errorf("ping mysql: %w", err))
	}

	return nil
}
