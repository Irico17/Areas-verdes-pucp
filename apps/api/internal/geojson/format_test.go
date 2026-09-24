package geojson

import "testing"

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
