package geojson

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestParseBBox(t *testing.T) {
	b, err := ParseBBox("-77.09,-12.08,-77.07,-12.06")
	if err != nil {
		t.Fatal(err)
	}
	if b.MinX != -77.09 || b.MaxY != -12.06 {
		t.Fatalf("%+v", b)
	}
	if _, err := ParseBBox("1,2,3"); err == nil {
		t.Fatal("esperaba error")
	}
	if _, err := ParseBBox("5,0,1,2"); err == nil {
		t.Fatal("esperaba min < max")
	}
	empty, err := ParseBBox("  ")
	if err != nil || empty != nil {
		t.Fatalf("vacío: %v %v", empty, err)
	}
}

func TestRoundedFloat(t *testing.T) {
	b, err := RoundedFloat(11147.353).MarshalJSON()
	if err != nil {
		t.Fatal(err)
	}
	if string(b) != "11147.353" {
		t.Fatalf("%s", b)
	}
}

func TestZonaPropertiesOmiteSectorVacio(t *testing.T) {
	sinSector := ZonaProperties{CatastroProperties: CatastroProperties{ID: 1, FeatureID: "Z-0001"}}
	b, err := json.Marshal(sinSector)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(b), `"sector"`) || strings.Contains(string(b), `"sector_etiqueta"`) {
		t.Fatalf("sin sector no debía serializar esas claves: %s", b)
	}

	sector := "cua-mateo"
	etiqueta := "Cuadrilla Mateo Salazar"
	conSector := ZonaProperties{
		CatastroProperties: CatastroProperties{ID: 1, FeatureID: "Z-0001"},
		Sector:             &sector,
		SectorEtiqueta:     &etiqueta,
	}
	b, err = json.Marshal(conSector)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(b), `"sector":"cua-mateo"`) || !strings.Contains(string(b), `"sector_etiqueta":"Cuadrilla Mateo Salazar"`) {
		t.Fatalf("con sector debía serializar ambas claves: %s", b)
	}
}
