package blobs

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// Disk guarda evidencias en una carpeta local.
type Disk struct {
	Dir string
}

func (d Disk) Put(_ context.Context, name string, r io.Reader, _ string) (string, error) {
	if err := os.MkdirAll(d.Dir, 0o755); err != nil {
		return "", err
	}
	path := filepath.Join(d.Dir, filepath.Base(name))
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

func (d Disk) Open(_ context.Context, ref string) (io.ReadCloser, error) {
	clean := filepath.Clean(ref)
	base := filepath.Clean(d.Dir)
	if base == "." || base == "" || !strings.HasPrefix(clean, base+string(os.PathSeparator)) {
		return nil, os.ErrNotExist
	}
	return os.Open(clean)
}
