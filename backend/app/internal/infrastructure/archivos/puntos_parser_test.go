package archivos_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/infrastructure/archivos"
)

func TestURLSinPlaceID(t *testing.T) {
	in := "https://www.google.com/maps/search/?api=1&query=Biblioteca&query_place_id=ChIJvQZ"
	out := archivos.UrlSinContacto(in)
	if strings.Contains(out, "place") || strings.Contains(out, "ChIJ") {
		t.Fatalf("url %s", out)
	}
	if !strings.Contains(out, "query=Biblioteca") {
		t.Fatalf("perdió la url pública: %s", out)
	}
}

func TestPuntosSinContacto(t *testing.T) {
	parser := archivos.NewPuntosParser()

	// Try reading actual sheets/puntos_pucp.csv if present
	csvPath := filepath.Join("..", "..", "..", "..", "..", "data", "raw", "sheets", "puntos_pucp.csv")
	body, err := os.ReadFile(csvPath)
	if err != nil {
		// Fallback synthetic CSV
		body = []byte(`title,location/lat,location/lng,url,phone,placeId
Comedor Central,"-12.071234","-77.081234","https://www.google.com/maps/search/?api=1&query=Comedor&query_place_id=ChIJ123",999999999,ChIJ123
Fuera de Campus,"-10.000000","-75.000000","https://example.com",,
`)
	}

	rows, rech, omitidas, err := parser.LeerPuntosPUCP(body)
	if err != nil {
		t.Fatal(err)
	}

	texto := strings.Join(omitidas, ",")
	if !strings.Contains(texto, "phone") || !strings.Contains(texto, "placeId") {
		t.Fatalf("columnas omitidas: %v", omitidas)
	}

	for _, row := range rows {
		if strings.Contains(row.URL, "placeId") || strings.Contains(row.URL, "place_id") || strings.Contains(strings.ToLower(row.URL), "phone") {
			t.Fatalf("punto %s guardó contacto: %s", row.Titulo, row.URL)
		}
	}

	if len(rows) == 153 {
		// Loaded from real data/raw/sheets/puntos_pucp.csv
		if len(rech) > 0 {
			t.Logf("rechazos: %d", len(rech))
		}
	}
}
