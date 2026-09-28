package seguridad_test

import (
	"testing"

	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/infrastructure/seguridad"
)

func TestBcryptHasher(t *testing.T) {
	hasher := seguridad.NewBcryptHasher()

	pwd := "pando-local"
	hash, err := hasher.Hash(pwd)
	if err != nil {
		t.Fatalf("Hash falló: %v", err)
	}
	if hash == "" || hash == pwd {
		t.Fatalf("hash inválido: %q", hash)
	}

	if err := hasher.Compare(hash, pwd); err != nil {
		t.Fatalf("Compare clave correcta falló: %v", err)
	}

	if err := hasher.Compare(hash, "clave-incorrecta"); err == nil {
		t.Fatal("Compare con clave errónea debió fallar")
	}
}
