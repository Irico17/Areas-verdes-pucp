package etl

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestPromotePolygon(t *testing.T) {
	in := json.RawMessage(`{"type":"Polygon","coordinates":[[[1,2],[3,4],[5,6],[1,2]]]}`)
	out, err := promoteMultiPolygon(in)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(out), `"type":"MultiPolygon"`) {
		t.Fatalf("tipo: %s", out)
	}
	if !strings.Contains(string(out), "[[[1,2],[3,4],[5,6],[1,2]]]") {
		t.Fatalf("coordenadas reescritas: %s", out)
	}
}

func TestNormalizeBaseline(t *testing.T) {
	dir := rawDir(t)

	areasBody, err := os.ReadFile(filepath.Join(dir, "areas_verdes.geojson"))
	if err != nil {
		t.Fatal(err)
	}
	areas, err := NormalizeAreas(areasBody)
	if err != nil {
		t.Fatal(err)
	}
	if len(areas) != ExpectedAreas {
		t.Fatalf("áreas = %d, want %d", len(areas), ExpectedAreas)
	}
	if areas[0].FeatureID != "AV-0001" || areas[0].Nombre == nil || *areas[0].Nombre != "Bosque Húmedo" {
		t.Fatalf("primera área: %+v", areas[0])
	}
	if areas[0].Codigo == nil || *areas[0].Codigo != "G 13" {
		t.Fatalf("código: %v", areas[0].Codigo)
	}
	if string(areas[0].AreaRaw) != "11147.353" {
		t.Fatalf("área m2: %s", areas[0].AreaRaw)
	}
	if areas[1].Nombre != nil || areas[1].Codigo != nil {
		t.Fatalf("feature vacía debería ser nil: nombre=%v codigo=%v", areas[1].Nombre, areas[1].Codigo)
	}
	assertMulti(t, areas[0].Geometry, "-77.08072672456687")

	zonasBody, err := os.ReadFile(filepath.Join(dir, "jefe_de_grupo.json"))
	if err != nil {
		t.Fatal(err)
	}
	zonas, err := NormalizeZonas(zonasBody)
	if err != nil {
		t.Fatal(err)
	}
	if len(zonas) != ExpectedZonas {
		t.Fatalf("zonas = %d, want %d", len(zonas), ExpectedZonas)
	}
	if zonas[0].FeatureID != "Z-0001" {
		t.Fatalf("zone id: %s", zonas[0].FeatureID)
	}
	assertMulti(t, zonas[0].Geometry, "-77.08072672456687")
	if err := assertNoPII(zonas); err != nil {
		t.Fatal(err)
	}
	blob, _ := json.Marshal(zonasPropsForTest(zonas))
	for _, frag := range PIIFragments {
		if strings.Contains(string(blob), frag) {
			t.Fatalf("PII %q en propiedades", frag)
		}
	}

	jrBody, err := os.ReadFile(filepath.Join(dir, "jardines_reserva.geojson"))
	if err != nil {
		t.Fatal(err)
	}
	jr, err := NormalizeCapa(jrBody, "jardines_reserva", "JR")
	if err != nil {
		t.Fatal(err)
	}
	if len(jr) != ExpectedJardines {
		t.Fatalf("jardines = %d", len(jr))
	}
	if jr[0].Pertenecen == nil || *jr[0].Pertenecen != "DAF" {
		t.Fatalf("pertenecen: %v", jr[0].Pertenecen)
	}

	xeBody, err := os.ReadFile(filepath.Join(dir, "xerofitica.geojson"))
	if err != nil {
		t.Fatal(err)
	}
	xe, err := NormalizeCapa(xeBody, "xerofitica", "XE")
	if err != nil {
		t.Fatal(err)
	}
	if len(xe) != ExpectedXerofitica {
		t.Fatalf("xerofitica = %d", len(xe))
	}
	if xe[0].Clase == nil || *xe[0].Clase != "Ornato" {
		t.Fatalf("clase: %v", xe[0].Clase)
	}
}

func zonasPropsForTest(rows []Record) []string {
	out := make([]string, 0, len(rows))
	for _, z := range rows {
		out = append(out, strings.Join([]string{
			z.FeatureID, str(z.Codigo), str(z.Nombre), str(z.Uso), str(z.ProyRiego), str(z.RiegoAct), str(z.Referencia),
		}, "|"))
	}
	return out
}

func assertMulti(t *testing.T, geom json.RawMessage, coord string) {
	t.Helper()
	var head struct {
		Type string `json:"type"`
	}
	if err := json.Unmarshal(geom, &head); err != nil {
		t.Fatal(err)
	}
	if head.Type != "MultiPolygon" {
		t.Fatalf("tipo %s", head.Type)
	}
	if !strings.Contains(string(geom), coord) {
		t.Fatalf("falta coordenada original %s", coord)
	}
}

func rawDir(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("caller")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", "..", "..", "..", "data", "raw"))
}
