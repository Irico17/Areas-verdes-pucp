package etl

import (
	"strings"
	"testing"
)

func TestCommaFloatYLatLon(t *testing.T) {
	v, ok := commaFloat("-12,067020")
	if !ok || v > -12 || v < -12.1 {
		t.Fatalf("lat %v %v", v, ok)
	}
	lat, lon, ok := latLonFrom([]string{"1", "PT_b1", "-12,067020", "-77,078135", "nota"})
	if !ok || lat > -12 || lon > -77 {
		t.Fatalf("punto %v %v %v", lat, lon, ok)
	}
}

func TestPlainPlaceIgnoraHTML(t *testing.T) {
	lugar := plainPlace(map[string]any{
		"Name":       "PT_bbn23",
		"descriptio": "<table><td>PT_bbn23</td></table>",
		"nota":       "Jardín frontal de Artes, junto a las mesas",
	}, "PT_bbn23")
	if strings.Contains(lugar, "<") || strings.Contains(lugar, "table") {
		t.Fatalf("se coló HTML: %q", lugar)
	}
	if !strings.Contains(lugar, "Jardín") {
		t.Fatalf("lugar %q", lugar)
	}
}

func TestUniqueID(t *testing.T) {
	seen := map[string]int{}
	if uniqueID(seen, "BB-a") != "BB-a" || uniqueID(seen, "BB-a") != "BB-a-2" {
		t.Fatalf("%v", seen)
	}
}
