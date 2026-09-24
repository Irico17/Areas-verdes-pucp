package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"campusverde/api/internal/catastro"
	"campusverde/api/internal/geojson"

	"github.com/gin-gonic/gin"
)

type fakeGeo struct {
	areas geojson.FeatureCollection
}

func (f fakeGeo) Areas(context.Context, catastro.Filter) (geojson.FeatureCollection, error) {
	return f.areas, nil
}
func (f fakeGeo) Zonas(context.Context, catastro.Filter) (geojson.FeatureCollection, error) {
	return geojson.Collection("zonas"), nil
}
func (f fakeGeo) Capa(_ context.Context, capa string, _ catastro.Filter) (geojson.FeatureCollection, error) {
	if capa != "xerofitica" && capa != "jardines_reserva" {
		return geojson.FeatureCollection{}, catastro.ErrCapaDesconocida
	}
	return geojson.Collection(capa), nil
}
func (f fakeGeo) Capas(context.Context) (catastro.CapasIndex, error) {
	return catastro.CapasIndex{CapasConocidas: catastro.CapasConocidas, Cargadas: []catastro.CapaCount{}}, nil
}
func (f fakeGeo) Resumen(context.Context) (catastro.Resumen, error) {
	return catastro.Resumen{CRS: "EPSG:4326", Areas: 521, Zonas: 534}, nil
}

func TestGeoAreas(t *testing.T) {
	gin.SetMode(gin.TestMode)
	fc := geojson.Collection("areas_verdes")
	fc.Features = append(fc.Features, geojson.Feature{
		Type:       "Feature",
		ID:         "AV-0001",
		Geometry:   json.RawMessage(`{"type":"MultiPolygon","coordinates":[]}`),
		Properties: geojson.CatastroProperties{ID: 1, FeatureID: "AV-0001"},
	})

	r := gin.New()
	h := Geo{Source: fakeGeo{areas: fc}}
	r.GET("/api/v1/geo/areas", h.Areas)
	r.GET("/api/v1/geo/capas/:capa", h.Capa)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/v1/geo/areas", nil))
	if w.Code != 200 {
		t.Fatalf("status %d", w.Code)
	}
	if ct := w.Header().Get("Content-Type"); ct != "application/geo+json; charset=utf-8" {
		t.Fatalf("content-type %s", ct)
	}
	var got geojson.FeatureCollection
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	if got.Type != "FeatureCollection" || len(got.Features) != 1 || got.Features[0].ID != "AV-0001" {
		t.Fatalf("body %s", w.Body.Bytes())
	}

	w = httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/v1/geo/areas?bbox=nope", nil))
	if w.Code != 400 {
		t.Fatalf("bbox status %d", w.Code)
	}

	w = httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/v1/geo/capas/fauna", nil))
	if w.Code != 404 {
		t.Fatalf("capa status %d", w.Code)
	}
}

func TestGeoSinDB(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := Geo{}
	r.GET("/api/v1/geo/zonas", h.Zonas)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/v1/geo/zonas", nil))
	if w.Code != 503 {
		t.Fatalf("status %d", w.Code)
	}
}
