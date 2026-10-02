package usecases

import (
	"context"
	"errors"
	"testing"

	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/dto"
	apperrors "github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/errors"
)

type fakeZonificacionRepo struct {
	sectores map[string]dto.SectorCapatazDTO
	lugares  []dto.LugarCatalogoDTO
	altas    int
}

func (f *fakeZonificacionRepo) ListarSectores(context.Context, bool) ([]dto.SectorCapatazDTO, error) {
	out := make([]dto.SectorCapatazDTO, 0, len(f.sectores))
	for _, item := range f.sectores {
		out = append(out, item)
	}
	return out, nil
}

func (f *fakeZonificacionRepo) CrearSector(_ context.Context, in dto.CrearSectorDTO) (dto.SectorCapatazDTO, error) {
	if f.sectores == nil {
		f.sectores = map[string]dto.SectorCapatazDTO{}
	}
	if _, ok := f.sectores[in.Codigo]; ok {
		return dto.SectorCapatazDTO{}, apperrors.ErrSectorDuplicado
	}
	item := dto.SectorCapatazDTO{ID: int64(len(f.sectores) + 1), Codigo: in.Codigo, Nombre: in.Nombre, Color: in.Color, Activo: true}
	f.sectores[in.Codigo] = item
	f.altas++
	return item, nil
}

func (f *fakeZonificacionRepo) ActualizarSector(context.Context, dto.ActualizarSectorDTO) (dto.SectorCapatazDTO, error) {
	return dto.SectorCapatazDTO{}, nil
}
func (f *fakeZonificacionRepo) DesactivarSector(context.Context, string, int64) error { return nil }
func (f *fakeZonificacionRepo) ImportarSectores(_ context.Context, filas []dto.CrearSectorDTO, _ int64) (dto.ImportacionSectorDTO, error) {
	var res dto.ImportacionSectorDTO
	for _, fila := range filas {
		if _, ok := f.sectores[fila.Codigo]; ok {
			f.sectores[fila.Codigo] = dto.SectorCapatazDTO{Codigo: fila.Codigo, Nombre: fila.Nombre, Color: fila.Color, Activo: true}
			res.Actualizados++
			continue
		}
		if _, err := f.CrearSector(context.Background(), fila); err != nil {
			return res, err
		}
		res.Creados++
	}
	return res, nil
}
func (f *fakeZonificacionRepo) ListarLugares(context.Context) ([]dto.LugarCatalogoDTO, error) {
	return f.lugares, nil
}
func (f *fakeZonificacionRepo) LugarPorID(_ context.Context, id int64) (dto.LugarCatalogoDTO, bool, error) {
	for _, lugar := range f.lugares {
		if lugar.ID == id {
			return lugar, true, nil
		}
	}
	return dto.LugarCatalogoDTO{}, false, nil
}
func (f *fakeZonificacionRepo) LugarPorNorm(_ context.Context, nombreNorm string) (dto.LugarCatalogoDTO, bool, error) {
	for _, lugar := range f.lugares {
		if lugar.Nombre == nombreNorm || lugar.Nombre == "Jardín de prueba" && nombreNorm == "jardin de prueba" {
			return lugar, true, nil
		}
	}
	return dto.LugarCatalogoDTO{}, false, nil
}
func (f *fakeZonificacionRepo) ContarLugares(context.Context) (int64, error) {
	return int64(len(f.lugares)), nil
}
func (f *fakeZonificacionRepo) ViasGeoJSON(context.Context) ([]byte, error) {
	return []byte(`{"type":"FeatureCollection","features":[]}`), nil
}
func (f *fakeZonificacionRepo) ImportarVias(context.Context, []dto.ViaAltaDTO, int64) (dto.ImportacionViaDTO, error) {
	return dto.ImportacionViaDTO{}, nil
}
func (f *fakeZonificacionRepo) CuartelesGeoJSON(context.Context) ([]byte, error) {
	return []byte(`{"type":"FeatureCollection","aviso":"sin archivo de cuarteles","features":[]}`), nil
}
func (f *fakeZonificacionRepo) CrearReferente(context.Context, dto.CrearReferenteDTO) (dto.ReferenteDTO, error) {
	return dto.ReferenteDTO{}, nil
}

type fakeArchivos struct {
	body []byte
	err  error
}

func (f fakeArchivos) LeerEdificios(context.Context) ([]byte, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.body, nil
}

func TestCrearSectorNoDuplica(t *testing.T) {
	repo := &fakeZonificacionRepo{}
	uc := NewZonificacionUseCase(repo, fakeArchivos{body: []byte(`{"features":[]}`)})
	primero, err := uc.CrearSector(context.Background(), dto.CrearSectorDTO{
		Codigo: "sector-lago", Nombre: "Sector lago (ficticio)", Color: "#3F73B0",
	})
	if err != nil {
		t.Fatal(err)
	}
	if primero.Color != "#3f73b0" || !primero.Activo {
		t.Fatalf("sector %+v", primero)
	}
	_, err = uc.CrearSector(context.Background(), dto.CrearSectorDTO{
		Codigo: "sector-lago", Nombre: "Otro nombre", Color: "#3f73b0",
	})
	if !errors.Is(err, apperrors.ErrSectorDuplicado) {
		t.Fatalf("se esperaba duplicado, obtuvo %v", err)
	}
	if repo.altas != 1 {
		t.Fatalf("altas %d", repo.altas)
	}
}

func TestLugarDesconocidoNoSeInserta(t *testing.T) {
	repo := &fakeZonificacionRepo{lugares: []dto.LugarCatalogoDTO{{ID: 4, Nombre: "Jardín de prueba"}}}
	antes := len(repo.lugares)
	uc := NewZonificacionUseCase(repo, fakeArchivos{})
	_, err := uc.ResolverLugar(context.Background(), dto.ResolverLugarDTO{Nombre: "Un sitio que nadie cargó"})
	if !errors.Is(err, apperrors.ErrLugarDesconocido) {
		t.Fatalf("se esperaba lugar desconocido, obtuvo %v", err)
	}
	if len(repo.lugares) != antes || repo.altas != 0 {
		t.Fatal("un lugar desconocido no debe insertarse")
	}
	res, err := uc.ResolverLugar(context.Background(), dto.ResolverLugarDTO{
		Nombre:     "Jardín de prueba",
		LugarLibre: "texto ya guardado",
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.LugarID == nil || *res.LugarID != 4 || res.LugarLibre != "texto ya guardado" {
		t.Fatalf("resuelto %+v", res)
	}
	if len(repo.lugares) != antes {
		t.Fatal("resolver un lugar conocido tampoco inserta filas")
	}
}

func TestViaSinLineaNoSeInventa(t *testing.T) {
	uc := NewZonificacionUseCase(&fakeZonificacionRepo{}, fakeArchivos{})
	_, err := uc.ImportarVias(context.Background(), []byte(`{"type":"FeatureCollection","features":[]}`), 1)
	if !errors.Is(err, apperrors.ErrViaVacia) {
		t.Fatalf("archivo vacío: %v", err)
	}
}

func TestEdificioSoloPorID(t *testing.T) {
	uc := NewZonificacionUseCase(
		&fakeZonificacionRepo{lugares: []dto.LugarCatalogoDTO{{ID: 1, Nombre: "Jardín de prueba"}}},
		fakeArchivos{body: []byte(`{"features":[{"id":"osm-way-1","properties":{"nombre":"CIGA"}}]}`)},
	)
	_, err := uc.CrearReferente(context.Background(), dto.CrearReferenteDTO{LugarID: 1, EdificioID: "el pabellón nuevo"})
	if !errors.Is(err, apperrors.ErrEdificioDesconocido) {
		t.Fatalf("texto suelto: %v", err)
	}
	item, err := uc.CrearReferente(context.Background(), dto.CrearReferenteDTO{LugarID: 1, EdificioID: "osm-way-1"})
	if err != nil {
		t.Fatal(err)
	}
	_ = item
}
