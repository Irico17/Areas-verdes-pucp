// Package contracts defines interfaces implemented across layers.
package contracts

import (
	"context"
	"io"
)

// IAlmacenArchivos defines the contract for storing and retrieving evidence blobs.
type IAlmacenArchivos interface {
	Put(ctx context.Context, name string, r io.Reader, mime string) (string, error)
	Open(ctx context.Context, ref string) (io.ReadCloser, error)
}
