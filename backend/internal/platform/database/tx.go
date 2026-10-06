package database

import (
	"context"

	"gorm.io/gorm"
)

// TxManager runs a function inside a database transaction. Services depend on
// this interface, never on GORM directly.
type TxManager interface {
	WithTx(ctx context.Context, fn func(ctx context.Context) error) error
}

type txKey struct{}

type txManager struct {
	db *gorm.DB
}

// NewTxManager returns the production TxManager.
func NewTxManager(db *gorm.DB) TxManager { return &txManager{db: db} }

// WithTx begins a transaction, puts it in the context and commits or rolls
// back based on the returned error.
func (m *txManager) WithTx(ctx context.Context, fn func(ctx context.Context) error) error {
	return m.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(context.WithValue(ctx, txKey{}, tx))
	})
}

// Connection returns the transaction stored in ctx when present, otherwise
// falls back to the regular connection.
func Connection(ctx context.Context, fallback *gorm.DB) *gorm.DB {
	if tx, ok := ctx.Value(txKey{}).(*gorm.DB); ok {
		return tx
	}
	return fallback
}
