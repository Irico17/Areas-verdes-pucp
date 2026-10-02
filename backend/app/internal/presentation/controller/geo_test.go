package controller_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"

	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/dto"
	domainEntities "github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/entities"
	domainErrors "github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/errors"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/presentation/controller"
)

type fakeGeoUseCase struct {
	areas        domainEntities.FeatureCollection
	edificios    []byte
	edificiosErr error
	zonasErr     error
	resumenErr   error
	capasErr     error
	capaErr      error
	areasErr     error
}

func (f fakeGeoUseCase) Areas(_ context.Context, _ dto.FiltroGeoDTO) (domainEntities.FeatureCollection, error) {
	if f.areasErr != nil {
		return domainEntities.FeatureCollection{}, f.areasErr
	}
	return f.areas, nil
}

func (f fakeGeoUseCase) Zonas(_ context.Context, _ dto.FiltroGeoDTO) (domainEntities.FeatureCollection, error) {
	if f.zonasErr != nil {
		return domainEntities.FeatureCollection{}, f.zonasErr
	}
	sector := "cua-mateo"
	etiqueta := "Cuadrilla Mateo Salazar"
	fc := domainEntities.Collection("zonas")
	fc.Features = append(fc.Features, domainEntities.Feature{
		Type:     "Feature",
		ID:       "Z-0001",
		Geometry: json.RawMessage(`{"type":"MultiPolygon","coordinates":[]}`),
		Properties: dto.ZonaPropertiesDTO{
			CatastroPropertiesDTO: dto.CatastroPropertiesDTO{ID: 1, FeatureID: "Z-0001"},
			Sector:                &sector,
			SectorEtiqueta:        &etiqueta,
		},
	})
	return fc, nil
}

func (f fakeGeoUseCase) Capa(_ context.Context, capa string, _ dto.FiltroGeoDTO) (domainEntities.FeatureCollection, error) {
	if f.capaErr != nil {
		return domainEntities.FeatureCollection{}, f.capaErr
	}
	if capa != "xerofitica" && capa != "jardines_reserva" {
		return domainEntities.FeatureCollection{}, domainErrors.ErrCapaDesconocida
	}
	return domainEntities.Collection(capa), nil
}

func (f fakeGeoUseCase) Capas(context.Context) (dto.CapasIndexDTO, error) {
	if f.capasErr != nil {
		return dto.CapasIndexDTO{}, f.capasErr
	}
	return dto.CapasIndexDTO{
		CapasConocidas: []string{"jardines_reserva", "xerofitica"},
		Cargadas:       []dto.CapaCountDTO{},
	}, nil
}

func (f fakeGeoUseCase) Resumen(context.Context) (dto.ResumenDTO, error) {
	if f.resumenErr != nil {
		return dto.ResumenDTO{}, f.resumenErr
	}
	return dto.ResumenDTO{CRS: "EPSG:4326", Areas: 521, Zonas: 534}, nil
}

func (f fakeGeoUseCase) Edificios(context.Context) ([]byte, error) {
	if f.edificiosErr != nil {
		return nil, f.edificiosErr
	}
	if f.edificios != nil {
		return f.edificios, nil
	}
	return []byte(`{"type":"FeatureCollection","name":"edificios","features":[]}`), nil
}

func TestGeoController_Areas(t *testing.T) {
	gin.SetMode(gin.TestMode)
	fc := domainEntities.Collection("areas_verdes")
	fc.Features = append(fc.Features, domainEntities.Feature{
		Type:       "Feature",
		ID:         "AV-0001",
		Geometry:   json.RawMessage(`{"type":"MultiPolygon","coordinates":[]}`),
		Properties: dto.CatastroPropertiesDTO{ID: 1, FeatureID: "AV-0001"},
	})

	ctrl := controller.NewGeoController(fakeGeoUseCase{areas: fc}, zerolog.Nop())
	r := gin.New()
	r.GET("/api/v1/geo/areas", ctrl.Areas)
	r.GET("/api/v1/geo/capas/:capa", ctrl.Capa)

	// GET /areas -> 200
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/v1/geo/areas", nil))
	if w.Code != http.StatusOK {
		t.Fatalf("status %d", w.Code)
	}
	if ct := w.Header().Get("Content-Type"); ct != "application/geo+json; charset=utf-8" {
		t.Fatalf("content-type %s", ct)
	}
	var got domainEntities.FeatureCollection
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	if got.Type != "FeatureCollection" || len(got.Features) != 1 || got.Features[0].ID != "AV-0001" {
		t.Fatalf("body %s", w.Body.Bytes())
	}

	// BBox inválido -> 400
	w = httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/v1/geo/areas?bbox=nope", nil))
	if w.Code != http.StatusBadRequest {
		t.Fatalf("bbox status %d", w.Code)
	}

	// Capa desconocida -> 404
	w = httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/v1/geo/capas/fauna", nil))
	if w.Code != http.StatusNotFound {
		t.Fatalf("capa status %d", w.Code)
	}
	var capaErrResp map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &capaErrResp); err != nil {
		t.Fatal(err)
	}
	if capaErrResp["error"] != "capa desconocida" {
		t.Fatalf("error mensaje: %v", capaErrResp["error"])
	}
}

type countingGeoUseCase struct {
	fakeGeoUseCase
	limit int
	areas int
}

func (c countingGeoUseCase) Areas(_ context.Context, f dto.FiltroGeoDTO) (domainEntities.FeatureCollection, error) {
	fc := domainEntities.Collection("areas_verdes")
	for i := 0; i < c.areas; i++ {
		fc.Features = append(fc.Features, domainEntities.Feature{
			Type:     "Feature",
			ID:       strconv.Itoa(i + 1),
			Geometry: json.RawMessage(`{"type":"Polygon","coordinates":[]}`),
		})
	}
	if f.Limit != c.limit {
		return fc, context.Canceled
	}
	return fc, nil
}

func TestGeoController_AreasSinLimite(t *testing.T) {
	gin.SetMode(gin.TestMode)
	const n = 521
	ctrl := controller.NewGeoController(countingGeoUseCase{areas: n, limit: 0}, zerolog.Nop())
	r := gin.New()
	r.GET("/api/v1/geo/areas", ctrl.Areas)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/v1/geo/areas", nil))
	if w.Code != http.StatusOK {
		t.Fatalf("status %d body %s", w.Code, w.Body.Bytes())
	}
	var got domainEntities.FeatureCollection
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	if len(got.Features) != n {
		t.Fatalf("features %d, se esperaban %d", len(got.Features), n)
	}
}

func TestGeoController_ZonasIncluyeSector(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctrl := controller.NewGeoController(fakeGeoUseCase{}, zerolog.Nop())
	r := gin.New()
	r.GET("/api/v1/geo/zonas", ctrl.Zonas)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/v1/geo/zonas", nil))
	if w.Code != http.StatusOK {
		t.Fatalf("status %d", w.Code)
	}
	var got struct {
		Features []struct {
			Properties struct {
				Sector         *string `json:"sector"`
				SectorEtiqueta *string `json:"sector_etiqueta"`
			} `json:"properties"`
		} `json:"features"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	if len(got.Features) != 1 || got.Features[0].Properties.Sector == nil || *got.Features[0].Properties.Sector != "cua-mateo" {
		t.Fatalf("body %s", w.Body.Bytes())
	}
	if got.Features[0].Properties.SectorEtiqueta == nil || *got.Features[0].Properties.SectorEtiqueta != "Cuadrilla Mateo Salazar" {
		t.Fatalf("etiqueta esperada 'Cuadrilla Mateo Salazar', obtenida %v", got.Features[0].Properties.SectorEtiqueta)
	}
}

func TestGeoController_Edificios(t *testing.T) {
	gin.SetMode(gin.TestMode)
	edificiosJSON := []byte(`{"type":"FeatureCollection","name":"edificios","features":[{"type":"Feature","id":"b-1"}]}`)
	ctrl := controller.NewGeoController(fakeGeoUseCase{edificios: edificiosJSON}, zerolog.Nop())
	r := gin.New()
	r.GET("/api/v1/geo/edificios", ctrl.Edificios)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/v1/geo/edificios", nil))
	if w.Code != http.StatusOK {
		t.Fatalf("status %d", w.Code)
	}
	if ct := w.Header().Get("Content-Type"); ct != "application/geo+json; charset=utf-8" {
		t.Fatalf("content-type %s", ct)
	}
	if w.Body.String() != string(edificiosJSON) {
		t.Fatalf("edificios body %s", w.Body.Bytes())
	}
}

func TestGeoController_CapasYResumen(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctrl := controller.NewGeoController(fakeGeoUseCase{}, zerolog.Nop())
	r := gin.New()
	r.GET("/api/v1/geo/capas", ctrl.Capas)
	r.GET("/api/v1/geo/resumen", ctrl.Resumen)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/v1/geo/capas", nil))
	if w.Code != http.StatusOK {
		t.Fatalf("capas status %d", w.Code)
	}

	w = httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/v1/geo/resumen", nil))
	if w.Code != http.StatusOK {
		t.Fatalf("resumen status %d", w.Code)
	}
	var res dto.ResumenDTO
	if err := json.Unmarshal(w.Body.Bytes(), &res); err != nil {
		t.Fatal(err)
	}
	if res.Areas != 521 || res.Zonas != 534 {
		t.Fatalf("resumen data %+v", res)
	}
}
