package database

import (
	"context"

	"gorm.io/gorm"

	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/contracts"
)

type txContextKey struct{}

// ContextWithDB stores a *gorm.DB (such as an active transaction) into the context.
func ContextWithDB(ctx context.Context, db *gorm.DB) context.Context {
	return context.WithValue(ctx, txContextKey{}, db)
}

// DBFromContext retrieves a *gorm.DB from the context if present, or returns the fallback *gorm.DB.
func DBFromContext(ctx context.Context, fallback *gorm.DB) *gorm.DB {
	if val := ctx.Value(txContextKey{}); val != nil {
		if tx, ok := val.(*gorm.DB); ok && tx != nil {
			return tx
		}
	}
	return fallback
}

type transaccion struct {
	db *gorm.DB
}

// NewTransaccion creates a new ITransaccion provider wrapping *gorm.DB.
func NewTransaccion(db *gorm.DB) contracts.ITransaccion {
	return &transaccion{db: db}
}

// Ejecutar runs the given function inside a database transaction, passing the transaction through the context.
func (t *transaccion) Ejecutar(ctx context.Context, fn func(ctx context.Context) error) error {
	return t.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		txCtx := ContextWithDB(ctx, tx)
		return fn(txCtx)
	})
}
