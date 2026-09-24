package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHealthYIndiceSinDB(t *testing.T) {
	r := New(Deps{OpenAPIPath: "no-existe.yaml"})

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/health", nil))
	if w.Code != 503 {
		t.Fatalf("health %d %s", w.Code, w.Body.String())
	}

	w = httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/v1", nil))
	if w.Code != 200 {
		t.Fatalf("indice %d", w.Code)
	}
	var body struct {
		CRS   string `json:"crs"`
		Rutas []any  `json:"rutas"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body.CRS != "EPSG:4326" || len(body.Rutas) < 4 {
		t.Fatalf("indice %#v", body)
	}

	w = httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/v1/geo/areas", nil))
	if w.Code != 503 {
		t.Fatalf("areas sin db %d", w.Code)
	}
}
