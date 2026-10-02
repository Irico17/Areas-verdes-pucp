package controller

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"

	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/dto"
	apperrors "github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/errors"
)

type fakeZonificacionHTTP struct {
	creados int
}

func (f *fakeZonificacionHTTP) ListarSectores(context.Context, bool) (*dto.SectorListDTO, error) {
	return &dto.SectorListDTO{Sectores: []dto.SectorCapatazDTO{}}, nil
}
func (f *fakeZonificacionHTTP) CrearSector(_ context.Context, in dto.CrearSectorDTO) (dto.SectorCapatazDTO, error) {
	if in.Codigo == "sector-lago" && f.creados > 0 {
		return dto.SectorCapatazDTO{}, apperrors.ErrSectorDuplicado
	}
	f.creados++
	return dto.SectorCapatazDTO{ID: 9, Codigo: in.Codigo, Nombre: in.Nombre, Color: in.Color, Activo: true}, nil
}
func (f *fakeZonificacionHTTP) ActualizarSector(context.Context, dto.ActualizarSectorDTO) (dto.SectorCapatazDTO, error) {
	return dto.SectorCapatazDTO{}, nil
}
func (f *fakeZonificacionHTTP) DesactivarSector(context.Context, string, int64) error { return nil }
func (f *fakeZonificacionHTTP) ImportarSectores(context.Context, []dto.CrearSectorDTO, int64) (dto.ImportacionSectorDTO, error) {
	return dto.ImportacionSectorDTO{}, nil
}
func (f *fakeZonificacionHTTP) ListarLugares(context.Context) ([]dto.LugarCatalogoDTO, error) {
	return nil, nil
}
func (f *fakeZonificacionHTTP) ResolverLugar(context.Context, dto.ResolverLugarDTO) (dto.LugarResueltoDTO, error) {
	return dto.LugarResueltoDTO{}, apperrors.ErrLugarDesconocido
}
func (f *fakeZonificacionHTTP) Vias(context.Context) ([]byte, error) {
	return []byte(`{"type":"FeatureCollection","features":[]}`), nil
}
func (f *fakeZonificacionHTTP) ImportarVias(context.Context, []byte, int64) (dto.ImportacionViaDTO, error) {
	return dto.ImportacionViaDTO{}, nil
}
func (f *fakeZonificacionHTTP) Cuarteles(context.Context) ([]byte, error) {
	return []byte(`{"type":"FeatureCollection","aviso":"sin archivo de cuarteles","features":[]}`), nil
}
func (f *fakeZonificacionHTTP) Edificios(context.Context) ([]dto.EdificioRefDTO, error) {
	return nil, nil
}
func (f *fakeZonificacionHTTP) CrearReferente(context.Context, dto.CrearReferenteDTO) (dto.ReferenteDTO, error) {
	return dto.ReferenteDTO{}, nil
}

func TestSectorPost201Y409SinDelete(t *testing.T) {
	gin.SetMode(gin.TestMode)
	uc := &fakeZonificacionHTTP{}
	ctrl := NewZonificacionController(uc, zerolog.Nop())
	r := gin.New()
	r.POST("/zonificacion/sectores", ctrl.CrearSector)

	cuerpo := `{"codigo":"sector-lago","nombre":"Sector lago (ficticio)","color":"#3f73b0"}`
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/zonificacion/sectores", bytes.NewBufferString(cuerpo)))
	if w.Code != http.StatusCreated {
		t.Fatalf("alta %d %s", w.Code, w.Body.String())
	}
	w = httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/zonificacion/sectores", bytes.NewBufferString(cuerpo)))
	if w.Code != http.StatusConflict {
		t.Fatalf("duplicado %d %s", w.Code, w.Body.String())
	}
	w = httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodDelete, "/zonificacion/sectores/sector-lago", nil))
	if w.Code != http.StatusNotFound {
		t.Fatalf("DELETE no debe existir, obtuvo %d", w.Code)
	}
}
