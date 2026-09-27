package catastro

import "testing"

func TestEtiquetaSector(t *testing.T) {
	casos := map[string]string{
		"cua-valeria":     "Cuadrilla Valeria Quispe",
		"cua-mateo":       "Cuadrilla Mateo Salazar",
		"cua-renato":      "Cuadrilla Renato Cárdenas",
		"campo-deportivo": "Campo deportivo",
		"bosque-humedo":   "Bosque húmedo",
	}
	for slug, want := range casos {
		if got := EtiquetaSector(slug); got != want {
			t.Fatalf("EtiquetaSector(%q) = %q, se esperaba %q", slug, got, want)
		}
	}
	if got := EtiquetaSector("desconocido"); got != "" {
		t.Fatalf("EtiquetaSector(desconocido) = %q, se esperaba vacío", got)
	}
	if got := EtiquetaSector(""); got != "" {
		t.Fatalf("EtiquetaSector(\"\") = %q, se esperaba vacío", got)
	}
}
