package enums_test

import (
	"testing"

	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/constants/enums"
)

func TestEtiquetaSector(t *testing.T) {
	casos := map[string]string{
		"cua-valeria":     "Cuadrilla Valeria Quispe",
		"cua-mateo":       "Cuadrilla Mateo Salazar",
		"cua-renato":      "Cuadrilla Renato Cárdenas",
		"campo-deportivo": "Campo deportivo",
		"bosque-humedo":   "Bosque húmedo",
	}

	for slug, want := range casos {
		if got := enums.EtiquetaSector(slug); got != want {
			t.Fatalf("EtiquetaSector(%q) = %q, se esperaba %q", slug, got, want)
		}
	}

	if got := enums.EtiquetaSector("desconocido"); got != "" {
		t.Fatalf("EtiquetaSector(desconocido) = %q, se esperaba vacío", got)
	}

	if got := enums.EtiquetaSector(""); got != "" {
		t.Fatalf("EtiquetaSector(\"\") = %q, se esperaba vacío", got)
	}
}
