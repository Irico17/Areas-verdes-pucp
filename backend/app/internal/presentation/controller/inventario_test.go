package controller_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"

	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/dto"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/entities"
	domainErrors "github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/errors"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/presentation/controller"
)

type mockInventarioUC struct {
	indexFunc func(ctx context.Context) (dto.IndiceInventarioDTO, error)
	capaFunc  func(ctx context.Context, capa string) (entities.FeatureCollection, error)
	fotoFunc  func(ctx context.Context, name string) (string, error)
}

func (m *mockInventarioUC) Index(ctx context.Context) (dto.IndiceInventarioDTO, error) {
	if m.indexFunc != nil {
		return m.indexFunc(ctx)
	}
	return dto.IndiceInventarioDTO{}, nil
}

func (m *mockInventarioUC) Capa(ctx context.Context, capa string) (entities.FeatureCollection, error) {
	if m.capaFunc != nil {
		return m.capaFunc(ctx, capa)
	}
	return entities.Collection(capa), nil
}

func (m *mockInventarioUC) Foto(ctx context.Context, name string) (string, error) {
	if m.fotoFunc != nil {
		return m.fotoFunc(ctx, name)
	}
	return "", nil
}

func TestInventarioController_Index(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Success
	uc := &mockInventarioUC{
		indexFunc: func(_ context.Context) (dto.IndiceInventarioDTO, error) {
			return dto.IndiceInventarioDTO{
				Capas: entities.CapasConocidas,
				Cargadas: []dto.CapaCountDTO{
					{Capa: "bebederos", Features: 5},
				},
			}, nil
		},
	}
	ctrl := controller.NewInventarioController(uc, zerolog.Nop())

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/geo/inventario", nil)
	ctrl.Index(c)

	if w.Code != http.StatusOK {
		t.Fatalf("se esperaba 200, obtenido %d", w.Code)
	}
	var res dto.IndiceInventarioDTO
	if err := json.Unmarshal(w.Body.Bytes(), &res); err != nil {
		t.Fatal(err)
	}
	if len(res.Capas) != 11 || len(res.Cargadas) != 1 {
		t.Fatalf("respuesta inesperada: %+v", res)
	}

	// Error 500
	ucErr := &mockInventarioUC{
		indexFunc: func(_ context.Context) (dto.IndiceInventarioDTO, error) {
			return dto.IndiceInventarioDTO{}, errors.New("db error")
		},
	}
	ctrlErr := controller.NewInventarioController(ucErr, zerolog.Nop())
	w = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/geo/inventario", nil)
	ctrlErr.Index(c)
	if w.Code != http.StatusInternalServerError {
		t.Fatalf("se esperaba 500, obtenido %d", w.Code)
	}
	var errBody map[string]string
	_ = json.Unmarshal(w.Body.Bytes(), &errBody)
	if errBody["error"] != "no se pudo leer el inventario" {
		t.Fatalf("error inesperado: %v", errBody)
	}
}

func TestInventarioController_Capa(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// 1. Success
	uc := &mockInventarioUC{
		capaFunc: func(_ context.Context, capa string) (entities.FeatureCollection, error) {
			fc := entities.Collection(capa)
			fc.Features = append(fc.Features, entities.Feature{
				Type: "Feature",
				ID:   "BB-1",
			})
			return fc, nil
		},
	}
	ctrl := controller.NewInventarioController(uc, zerolog.Nop())
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "capa", Value: "bebederos"}}
	c.Request = httptest.NewRequest(http.MethodGet, "/geo/inventario/bebederos", nil)
	ctrl.Capa(c)

	if w.Code != http.StatusOK {
		t.Fatalf("se esperaba 200, obtenido %d", w.Code)
	}
	if ct := w.Header().Get("Content-Type"); ct != "application/geo+json; charset=utf-8" {
		t.Fatalf("Content-Type esperado 'application/geo+json; charset=utf-8', obtenido %q", ct)
	}

	// 2. Unknown layer -> 404 with both error and capas list
	ucUnknown := &mockInventarioUC{
		capaFunc: func(_ context.Context, capa string) (entities.FeatureCollection, error) {
			return entities.Collection(capa), domainErrors.ErrCapaInventarioDesconocida
		},
	}
	ctrlUnknown := controller.NewInventarioController(ucUnknown, zerolog.Nop())
	w = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "capa", Value: "desconocida"}}
	c.Request = httptest.NewRequest(http.MethodGet, "/geo/inventario/desconocida", nil)
	ctrlUnknown.Capa(c)

	if w.Code != http.StatusNotFound {
		t.Fatalf("se esperaba 404, obtenido %d", w.Code)
	}
	var notFoundRes struct {
		Error string   `json:"error"`
		Capas []string `json:"capas"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &notFoundRes); err != nil {
		t.Fatal(err)
	}
	if notFoundRes.Error != "capa de inventario desconocida" || len(notFoundRes.Capas) != 11 {
		t.Fatalf("cuerpo 404 inesperado: %+v", notFoundRes)
	}

	// 3. Error 500
	ucErr := &mockInventarioUC{
		capaFunc: func(_ context.Context, capa string) (entities.FeatureCollection, error) {
			return entities.Collection(capa), errors.New("db error")
		},
	}
	ctrlErr := controller.NewInventarioController(ucErr, zerolog.Nop())
	w = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "capa", Value: "bebederos"}}
	c.Request = httptest.NewRequest(http.MethodGet, "/geo/inventario/bebederos", nil)
	ctrlErr.Capa(c)
	if w.Code != http.StatusInternalServerError {
		t.Fatalf("se esperaba 500, obtenido %d", w.Code)
	}
}

func TestInventarioController_Foto(t *testing.T) {
	gin.SetMode(gin.TestMode)

	dir := t.TempDir()
	samplePhoto := filepath.Join(dir, "bbr63.jpg")
	if err := os.WriteFile(samplePhoto, []byte("fake-jpeg-data"), 0o644); err != nil {
		t.Fatal(err)
	}

	uc := &mockInventarioUC{
		fotoFunc: func(_ context.Context, name string) (string, error) {
			switch name {
			case "bbr63.jpg":
				return samplePhoto, nil
			case "invalido.png":
				return "", domainErrors.ErrFotografiaNoDisponible
			default:
				return "", domainErrors.ErrFotografiaNoRecuperada
			}
		},
	}
	ctrl := controller.NewInventarioController(uc, zerolog.Nop())

	// 1. Success serving JPEG
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "name", Value: "bbr63.jpg"}}
	c.Request = httptest.NewRequest(http.MethodGet, "/geo/inventario/fotos/bbr63.jpg", nil)
	ctrl.Foto(c)

	if w.Code != http.StatusOK {
		t.Fatalf("se esperaba 200, obtenido %d", w.Code)
	}
	if w.Body.String() != "fake-jpeg-data" {
		t.Fatalf("cuerpo de imagen inesperado: %s", w.Body.String())
	}

	// 2. Extensión inválida -> 404 "fotografía no disponible"
	w = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "name", Value: "invalido.png"}}
	c.Request = httptest.NewRequest(http.MethodGet, "/geo/inventario/fotos/invalido.png", nil)
	ctrl.Foto(c)

	if w.Code != http.StatusNotFound {
		t.Fatalf("se esperaba 404, obtenido %d", w.Code)
	}
	var errBody map[string]string
	_ = json.Unmarshal(w.Body.Bytes(), &errBody)
	if errBody["error"] != "fotografía no disponible" {
		t.Fatalf("error esperado 'fotografía no disponible', obtenido %v", errBody)
	}

	// 3. Foto inexistente -> 404 "fotografía no recuperada"
	w = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "name", Value: "no_existe.jpg"}}
	c.Request = httptest.NewRequest(http.MethodGet, "/geo/inventario/fotos/no_existe.jpg", nil)
	ctrl.Foto(c)

	if w.Code != http.StatusNotFound {
		t.Fatalf("se esperaba 404, obtenido %d", w.Code)
	}
	_ = json.Unmarshal(w.Body.Bytes(), &errBody)
	if errBody["error"] != "fotografía no recuperada" {
		t.Fatalf("error esperado 'fotografía no recuperada', obtenido %v", errBody)
	}
}
