package database

import "context"

type TxManager interface {
	WithinTransaction(ctx context.Context, fn func(ctx context.Context) error) error
}

// NoopTxManager is useful for early development before real DB transaction wiring exists.
type NoopTxManager struct{}

func (n NoopTxManager) WithinTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	return fn(ctx)
}
