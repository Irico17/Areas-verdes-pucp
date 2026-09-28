package contracts

import "context"

// ITransaccion defines operations to execute units of work inside a database transaction.
type ITransaccion interface {
	Ejecutar(ctx context.Context, fn func(ctx context.Context) error) error
}
