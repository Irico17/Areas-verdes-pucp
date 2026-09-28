package controller_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"

	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/dto"
	domainErrors "github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/errors"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/presentation/controller"
)

type fakeAreaVerdeUseCase struct {
	fichas        []dto.FichaDTO
	fichasErr     error
	actualizarRes *dto.FichaDTO
	actualizarErr error
	crearRes      *dto.FichaDTO
	crearErr      error
}

func (f fakeAreaVerdeUseCase) Fichas(_ context.Context, _ string) ([]dto.FichaDTO, error) {
	if f.fichasErr != nil {
		return nil, f.fichasErr
	}
	if f.fichas == nil {
		return []dto.FichaDTO{}, nil
	}
	return f.fichas, nil
}

func (f fakeAreaVerdeUseCase) ActualizarFicha(_ context.Context, _ string, _ dto.ActualizarFichaDTO, _ *int64) (*dto.FichaDTO, error) {
	if f.actualizarErr != nil {
		return nil, f.actualizarErr
	}
	return f.actualizarRes, nil
}

func (f fakeAreaVerdeUseCase) CrearSinGeom(_ context.Context, _ dto.CrearAreaSinGeomDTO, _ *int64) (*dto.FichaDTO, error) {
	if f.crearErr != nil {
		return nil, f.crearErr
	}
	return f.crearRes, nil
}

func TestAreaVerdeController_Listar(t *testing.T) {
	gin.SetMode(gin.TestMode)
	area := 11147.353
	fichas := []dto.FichaDTO{
		{
			FeatureID:  "AV-0001",
			Nombre:     "Bosque Húmedo",
			Uso:        "Uso Institucional",
			RiegoAct:   "Riego por aspersión",
			Referencia: "",
			AreaM2:     &area,
			ConGeom:    true,
		},
	}
	ctrl := controller.NewAreaVerdeController(fakeAreaVerdeUseCase{fichas: fichas}, zerolog.Nop())
	r := gin.New()
	r.GET("/api/v1/catastro/areas", ctrl.Listar)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/v1/catastro/areas", nil))
	if w.Code != http.StatusOK {
		t.Fatalf("status %d", w.Code)
	}

	var got struct {
		Areas []dto.FichaDTO `json:"areas"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	if len(got.Areas) != 1 || got.Areas[0].FeatureID != "AV-0001" {
		t.Fatalf("body %s", w.Body.Bytes())
	}
}

func TestAreaVerdeController_Actualizar(t *testing.T) {
	gin.SetMode(gin.TestMode)
	area := 11147.353
	res := &dto.FichaDTO{
		FeatureID:  "AV-0001",
		Nombre:     "Área Modificada",
		Uso:        "jardín",
		RiegoAct:   "goteo",
		Referencia: "cerca a EEGGCC",
		AreaM2:     &area,
		ConGeom:    true,
	}

	ctrl := controller.NewAreaVerdeController(fakeAreaVerdeUseCase{actualizarRes: res}, zerolog.Nop())
	r := gin.New()
	r.PATCH("/api/v1/catastro/areas/:id", ctrl.Actualizar)

	// JSON inválido -> 400
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodPatch, "/api/v1/catastro/areas/AV-0001", bytes.NewBufferString("{bad")))
	if w.Code != http.StatusBadRequest {
		t.Fatalf("status %d", w.Code)
	}

	// Éxito -> 200
	bodyJSON := `{"nombre":"Área Modificada","uso":"jardín","riego_act":"goteo","referencia":"cerca a EEGGCC"}`
	w = httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodPatch, "/api/v1/catastro/areas/AV-0001", bytes.NewBufferString(bodyJSON)))
	if w.Code != http.StatusOK {
		t.Fatalf("status %d body %s", w.Code, w.Body.Bytes())
	}

	// No encontrada -> 404
	ctrlNotFound := controller.NewAreaVerdeController(fakeAreaVerdeUseCase{actualizarErr: domainErrors.ErrFichaNoEncontrada}, zerolog.Nop())
	rNotFound := gin.New()
	rNotFound.PATCH("/api/v1/catastro/areas/:id", ctrlNotFound.Actualizar)
	w = httptest.NewRecorder()
	rNotFound.ServeHTTP(w, httptest.NewRequest(http.MethodPatch, "/api/v1/catastro/areas/AV-9999", bytes.NewBufferString(bodyJSON)))
	if w.Code != http.StatusNotFound {
		t.Fatalf("status %d", w.Code)
	}

	// Error validación / actualización -> 400
	ctrlErr := controller.NewAreaVerdeController(fakeAreaVerdeUseCase{actualizarErr: errors.New("entrada")}, zerolog.Nop())
	rErr := gin.New()
	rErr.PATCH("/api/v1/catastro/areas/:id", ctrlErr.Actualizar)
	w = httptest.NewRecorder()
	rErr.ServeHTTP(w, httptest.NewRequest(http.MethodPatch, "/api/v1/catastro/areas/AV-0001", bytes.NewBufferString(bodyJSON)))
	if w.Code != http.StatusBadRequest {
		t.Fatalf("status %d", w.Code)
	}
}

func TestAreaVerdeController_Crear(t *testing.T) {
	gin.SetMode(gin.TestMode)
	res := &dto.FichaDTO{
		FeatureID:  "AV-P123",
		Nombre:     "Nueva Área",
		Uso:        "jardín",
		RiegoAct:   "",
		Referencia: "",
		AreaM2:     nil,
		ConGeom:    false,
	}

	ctrl := controller.NewAreaVerdeController(fakeAreaVerdeUseCase{crearRes: res}, zerolog.Nop())
	r := gin.New()
	r.POST("/api/v1/catastro/areas", ctrl.Crear)

	// JSON inválido -> 400
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/api/v1/catastro/areas", bytes.NewBufferString("{bad")))
	if w.Code != http.StatusBadRequest {
		t.Fatalf("status %d", w.Code)
	}

	// Éxito -> 201
	bodyJSON := `{"feature_id":"AV-P123","nombre":"Nueva Área","uso":"jardín"}`
	w = httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/api/v1/catastro/areas", bytes.NewBufferString(bodyJSON)))
	if w.Code != http.StatusCreated {
		t.Fatalf("status %d body %s", w.Code, w.Body.Bytes())
	}

	// Error -> 400
	ctrlErr := controller.NewAreaVerdeController(fakeAreaVerdeUseCase{crearErr: errors.New("error al crear")}, zerolog.Nop())
	rErr := gin.New()
	rErr.POST("/api/v1/catastro/areas", ctrlErr.Crear)
	w = httptest.NewRecorder()
	rErr.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/api/v1/catastro/areas", bytes.NewBufferString(bodyJSON)))
	if w.Code != http.StatusBadRequest {
		t.Fatalf("status %d", w.Code)
	}
}
