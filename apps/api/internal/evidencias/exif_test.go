package evidencias

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"campusverde/api/internal/operacion"
)

func TestFiltrarExifDejaSoloLaListaBlanca(t *testing.T) {
	raw := json.RawMessage(`{"fecha":"2026:09:25 10:00:00","lat":-12.07,"lon":-77.08,"orientación":1,"modelo":"no","gps":{"a":1}}`)
	got, err := filtrarExif(raw)
	if err != nil {
		t.Fatal(err)
	}
	s, _ := got.(string)
	if strings.Contains(s, "modelo") || strings.Contains(s, "gps") {
		t.Fatalf("se coló un campo: %s", s)
	}
	for _, clave := range []string{"fecha", "lat", "lon", "orientacion"} {
		if !strings.Contains(s, clave) {
			t.Fatalf("falta %s en %s", clave, s)
		}
	}
}

func TestFiltrarExifRechazaElTope(t *testing.T) {
	raw := json.RawMessage(`{"fecha":"` + strings.Repeat("a", topeExif) + `"}`)
	_, err := filtrarExif(raw)
	var input operacion.InputError
	if !errors.As(err, &input) || !strings.Contains(input.Reason, "tamaño") {
		t.Fatalf("esperaba tope, obtuvo %v", err)
	}
}
