package blobs

import (
	"bytes"
	"context"
	"io"
	"os"
	"path/filepath"
	"testing"
)

func TestDiskRoundTrip(t *testing.T) {
	dir := t.TempDir()
	store := Disk{Dir: dir}
	ref, err := store.Put(context.Background(), "foto.jpg", bytes.NewBufferString("hola"), "image/jpeg")
	if err != nil {
		t.Fatal(err)
	}
	rc, err := store.Open(context.Background(), ref)
	if err != nil {
		t.Fatal(err)
	}
	defer rc.Close()
	body, _ := io.ReadAll(rc)
	if string(body) != "hola" {
		t.Fatalf("cuerpo %q", body)
	}
	if _, err := store.Open(context.Background(), filepath.Join(dir, "..", "otro")); err == nil {
		t.Fatal("no debe abrir fuera de la carpeta")
	}
	if _, err := os.Stat(ref); err != nil {
		t.Fatal(err)
	}
}
