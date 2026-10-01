// Package storage implements storage adapters for blob files.
package storage

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/contracts"
)

type discoStorage struct {
	dir string
}

// NewDiscoStorage creates a disk-based blob storage adapter.
func NewDiscoStorage(dir string) contracts.IAlmacenArchivos {
	return &discoStorage{dir: dir}
}

// Put writes the contents of r to a file named after the base of name inside dir.
func (d *discoStorage) Put(_ context.Context, name string, r io.Reader, _ string) (string, error) {
	if err := os.MkdirAll(d.dir, 0o755); err != nil {
		return "", err
	}
	path := filepath.Join(d.dir, filepath.Base(name))
	f, err := os.Create(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	if _, err := io.Copy(f, r); err != nil {
		_ = os.Remove(path)
		return "", err
	}
	return path, nil
}

// Open opens the file referenced by ref, ensuring path traversal outside dir is rejected.
func (d *discoStorage) Open(_ context.Context, ref string) (io.ReadCloser, error) {
	clean := filepath.Clean(ref)
	base := filepath.Clean(d.dir)
	if base == "." || base == "" || !strings.HasPrefix(clean, base+string(os.PathSeparator)) {
		return nil, os.ErrNotExist
	}
	return os.Open(clean)
}
