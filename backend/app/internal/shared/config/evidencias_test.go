package config

import "testing"

func TestEvidenciasBucketVacioORecortado(t *testing.T) {
	t.Setenv("EVIDENCIAS_BUCKET", "   ")
	vacio := New()
	if vacio.Evidencias.Bucket != "" || vacio.Evidencias.EnCubo() {
		t.Fatalf("espacios no son un cubo: %+v", vacio.Evidencias)
	}

	t.Setenv("EVIDENCIAS_BUCKET", " cubo-de-prueba ")
	conCubo := New()
	if conCubo.Evidencias.Bucket != "cubo-de-prueba" || !conCubo.Evidencias.EnCubo() {
		t.Fatalf("bucket recortado: %+v", conCubo.Evidencias)
	}
}
