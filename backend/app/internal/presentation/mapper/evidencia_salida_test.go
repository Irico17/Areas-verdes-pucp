package mapper

import "testing"

func TestSalidaSubidaNoPrometeS3SinCubo(t *testing.T) {
	sinCubo := SalidaSubida("id-1", false, "  ")
	if sinCubo.Almacen != "disco" || sinCubo.ID != "id-1" || sinCubo.Idempotente {
		t.Fatalf("%+v", sinCubo)
	}
	conCubo := SalidaSubida("id-2", true, "cubo-de-prueba")
	if conCubo.Almacen != "s3" || !conCubo.Idempotente {
		t.Fatalf("%+v", conCubo)
	}
}
