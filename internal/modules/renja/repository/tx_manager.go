package repository

import (
	"context"
	"fmt"

	"gorm.io/gorm"
)

type gormTxContextKey struct{}

type GormTxManager struct {
	db *gorm.DB
}

func NewGormTxManager(db *gorm.DB) *GormTxManager {
	return &GormTxManager{db: db}
}

func (m *GormTxManager) WithinTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	if m == nil || m.db == nil {
		return fmt.Errorf("nil gorm tx manager")
	}

	return m.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		txCtx := context.WithValue(ctx, gormTxContextKey{}, tx)
		return fn(txCtx)
	})
}

func GormTxFromContext(ctx context.Context) *gorm.DB {
	if ctx == nil {
		return nil
	}
	tx, ok := ctx.Value(gormTxContextKey{}).(*gorm.DB)
	if !ok {
		return nil
	}
	return tx
}
