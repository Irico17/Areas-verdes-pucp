package atencion

import "testing"

func TestHuecosSinFormula(t *testing.T) {
	vistos := map[string]bool{}
	for _, h := range HuecosIndicador() {
		vistos[h.Clave] = true
		if h.Estado != "definición pendiente" {
			t.Fatalf("%s: %q", h.Clave, h.Estado)
		}
		if h.Nota == "" || h.Nombre == "" {
			t.Fatalf("hueco incompleto %+v", h)
		}
	}
	for _, clave := range []string{"cobertura", "rendimiento", "metricas_proveedor"} {
		if !vistos[clave] {
			t.Fatalf("falta %s", clave)
		}
	}
	if len(vistos) != 3 {
		t.Fatalf("no se inventan más indicadores: %d", len(vistos))
	}
}
