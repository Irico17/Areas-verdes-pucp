package usecases_test

import (
	"context"
	"errors"
	"testing"

	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/dto"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/usecases"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/entities"
	domainErrors "github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/errors"
)

type mockZonaRepo struct {
	items []entities.ZonaSupervision
	err   error
}

func (m *mockZonaRepo) Listar(_ context.Context) ([]entities.ZonaSupervision, error) {
	return m.items, m.err
}

func (m *mockZonaRepo) Crear(_ context.Context, codigo, nombre, _ string, area *float64) (entities.ZonaSupervision, error) {
	if m.err != nil {
		return entities.ZonaSupervision{}, m.err
	}
	return entities.ZonaSupervision{ID: 1, Codigo: codigo, Nombre: nombre, AreaM2: area, ConGeom: true, Activo: true}, nil
}

func TestZonaSupervisionUseCase(t *testing.T) {
	ctx := context.Background()
	repo := &mockZonaRepo{items: []entities.ZonaSupervision{{ID: 1, Codigo: "Z1", Nombre: "Zona 1"}}}
	uc := usecases.NewZonaSupervisionUseCase(repo)

	// 1. Listar
	list, err := uc.Listar(ctx)
	if err != nil || len(list) != 1 {
		t.Fatalf("Listar failed: %v", err)
	}

	// 2. Crear valid
	area := 1500.0
	created, err := uc.Crear(ctx, dto.CrearZonaSupervisionDTO{Codigo: "Z1", Nombre: "Zona Uno", AreaM2: &area, GeoJSON: `{"type":"Polygon"}`})
	if err != nil || created.Codigo != "Z1" {
		t.Fatalf("Crear failed: %v", err)
	}

	// 3. Crear invalid code
	_, err = uc.Crear(ctx, dto.CrearZonaSupervisionDTO{Codigo: "Z9", Nombre: "Invalida", GeoJSON: `{"type":"Polygon"}`})
	if !errors.Is(err, domainErrors.ErrEntrada) {
		t.Fatalf("expected ErrEntrada, got: %v", err)
	}
}

type mockCuadrillaRepo struct {
	items []entities.Cuadrilla
	err   error
}

func (m *mockCuadrillaRepo) Listar(_ context.Context) ([]entities.Cuadrilla, error) {
	return m.items, m.err
}

func (m *mockCuadrillaRepo) Crear(_ context.Context, id, nombre, turno string) (entities.Cuadrilla, error) {
	if m.err != nil {
		return entities.Cuadrilla{}, m.err
	}
	return entities.Cuadrilla{ID: id, NombreFicticio: nombre, Turno: turno, Activo: true}, nil
}

func TestCuadrillaUseCase(t *testing.T) {
	ctx := context.Background()
	repo := &mockCuadrillaRepo{items: []entities.Cuadrilla{{ID: "C1", NombreFicticio: "Equipo 1", Turno: "manana", Activo: true}}}
	uc := usecases.NewCuadrillaUseCase(repo)

	list, err := uc.Listar(ctx)
	if err != nil || len(list) != 1 {
		t.Fatalf("Listar failed: %v", err)
	}

	created, err := uc.Crear(ctx, dto.CrearCuadrillaDTO{ID: "C2", Nombre: "Equipo 2", Turno: "tarde"})
	if err != nil || created.ID != "C2" {
		t.Fatalf("Crear failed: %v", err)
	}

	_, err = uc.Crear(ctx, dto.CrearCuadrillaDTO{ID: "C3", Nombre: "Equipo 3", Turno: "noche"})
	if !errors.Is(err, domainErrors.ErrEntrada) {
		t.Fatalf("expected ErrEntrada, got: %v", err)
	}
}

type mockLugarRepo struct {
	items []entities.Lugar
	err   error
}

func (m *mockLugarRepo) Listar(_ context.Context) ([]entities.Lugar, error) {
	return m.items, m.err
}

func (m *mockLugarRepo) Crear(_ context.Context, nombre string, lat, lon float64, zonaID *int64) (entities.Lugar, error) {
	if m.err != nil {
		return entities.Lugar{}, m.err
	}
	return entities.Lugar{ID: 1, Nombre: nombre, NombreNorm: entities.NormalizarNombre(nombre), Lat: lat, Lon: lon, ZonaSupervisionID: zonaID, Activo: true}, nil
}

func TestLugarUseCase(t *testing.T) {
	ctx := context.Background()
	repo := &mockLugarRepo{items: []entities.Lugar{{ID: 1, Nombre: "Jardín Central"}}}
	uc := usecases.NewLugarUseCase(repo)

	list, err := uc.Listar(ctx)
	if err != nil || len(list) != 1 {
		t.Fatalf("Listar failed: %v", err)
	}

	created, err := uc.Crear(ctx, dto.CrearLugarDTO{Nombre: "Pabellón Z", Lat: -12.07, Lon: -77.08})
	if err != nil || created.Nombre != "Pabellón Z" {
		t.Fatalf("Crear failed: %v", err)
	}

	_, err = uc.Crear(ctx, dto.CrearLugarDTO{Nombre: "Lejos", Lat: 0.0, Lon: 0.0})
	if !errors.Is(err, domainErrors.ErrEntrada) {
		t.Fatalf("expected ErrEntrada, got: %v", err)
	}
}

type mockEspecieRepo struct {
	items []entities.Especie
	err   error
}

func (m *mockEspecieRepo) Listar(_ context.Context) ([]entities.Especie, error) {
	return m.items, m.err
}

func (m *mockEspecieRepo) Crear(_ context.Context, cientifico, comun string) (entities.Especie, error) {
	if m.err != nil {
		return entities.Especie{}, m.err
	}
	return entities.Especie{ID: 1, NombreCientifico: cientifico, NombreComun: comun, Activo: true}, nil
}

func TestEspecieUseCase(t *testing.T) {
	ctx := context.Background()
	repo := &mockEspecieRepo{items: []entities.Especie{{ID: 1, NombreCientifico: "Tipuana tipu"}}}
	uc := usecases.NewEspecieUseCase(repo)

	list, err := uc.Listar(ctx)
	if err != nil || len(list) != 1 {
		t.Fatalf("Listar failed: %v", err)
	}

	created, err := uc.Crear(ctx, dto.CrearEspecieDTO{Cientifico: "Jacaranda mimosifolia", Comun: "Jacarandá"})
	if err != nil || created.NombreCientifico != "Jacaranda mimosifolia" {
		t.Fatalf("Crear failed: %v", err)
	}

	_, err = uc.Crear(ctx, dto.CrearEspecieDTO{Cientifico: "", Comun: "Solo común"})
	if !errors.Is(err, domainErrors.ErrEntrada) {
		t.Fatalf("expected ErrEntrada, got: %v", err)
	}
}

type mockEjemplarRepo struct {
	items   []entities.Ejemplar
	total   int
	codigos []entities.CodigoHistorico
	err     error
}

func (m *mockEjemplarRepo) Listar(_ context.Context, _, _ int) ([]entities.Ejemplar, int, error) {
	return m.items, m.total, m.err
}

func (m *mockEjemplarRepo) Crear(_ context.Context, e entities.Ejemplar) (entities.Ejemplar, error) {
	if m.err != nil {
		return entities.Ejemplar{}, m.err
	}
	e.ID = 100
	e.Activo = true
	return e, nil
}

func (m *mockEjemplarRepo) Recodificar(_ context.Context, ejemplarID int64, nuevo string) (entities.CodigoHistorico, error) {
	if m.err != nil {
		return entities.CodigoHistorico{}, m.err
	}
	return entities.CodigoHistorico{ID: 1, EjemplarID: ejemplarID, CodigoAnterior: "AV-OLD", CodigoNuevo: nuevo}, nil
}

func (m *mockEjemplarRepo) ListarCodigos(_ context.Context, ejemplarID int64) ([]entities.CodigoHistorico, error) {
	return m.codigos, m.err
}

func TestEjemplarUseCase(t *testing.T) {
	ctx := context.Background()
	repo := &mockEjemplarRepo{
		items:   []entities.Ejemplar{{ID: 1, Codigo: "AV-01", TipoVegetacion: "Árbol", Cantidad: 1}},
		total:   1,
		codigos: []entities.CodigoHistorico{{ID: 1, EjemplarID: 1, CodigoAnterior: "OLD", CodigoNuevo: "AV-01"}},
	}
	uc := usecases.NewEjemplarUseCase(repo)

	// Listar
	res, err := uc.Listar(ctx, 10, 0)
	if err != nil || res.Total != 1 || len(res.Ejemplares) != 1 {
		t.Fatalf("Listar failed: %v", err)
	}

	// Crear
	lat := -12.07
	lon := -77.08
	created, err := uc.Crear(ctx, dto.EjemplarDTO{Codigo: "AV-02", TipoVegetacion: "Palmera", Cantidad: 1, Lat: &lat, Lon: &lon})
	if err != nil || created.ID != 100 {
		t.Fatalf("Crear failed: %v", err)
	}

	// Crear invalid veg
	_, err = uc.Crear(ctx, dto.EjemplarDTO{Codigo: "AV-03", TipoVegetacion: "musgo"})
	if !errors.Is(err, domainErrors.ErrEntrada) {
		t.Fatalf("expected ErrEntrada, got: %v", err)
	}

	// Recodificar
	hist, err := uc.Recodificar(ctx, 1, dto.RecodificarDTO{Codigo: "AV-NEW"})
	if err != nil || hist.CodigoNuevo != "AV-NEW" {
		t.Fatalf("Recodificar failed: %v", err)
	}

	// ListarCodigos
	cods, err := uc.ListarCodigos(ctx, 1)
	if err != nil || len(cods) != 1 {
		t.Fatalf("ListarCodigos failed: %v", err)
	}
}

type mockReferenciaRepo struct {
	poligonos []entities.PoligonoCuadrilla
	capa      []entities.CapaFicha
	err       error
}

func (m *mockReferenciaRepo) ListarPoligonos(_ context.Context) ([]entities.PoligonoCuadrilla, error) {
	return m.poligonos, m.err
}

func (m *mockReferenciaRepo) ListarCapa(_ context.Context, _ string) ([]entities.CapaFicha, error) {
	return m.capa, m.err
}

func TestCatastroReferenciaUseCase(t *testing.T) {
	ctx := context.Background()
	repo := &mockReferenciaRepo{
		poligonos: []entities.PoligonoCuadrilla{{ID: 1, FeatureID: "P-1"}},
		capa:      []entities.CapaFicha{{ID: 1, FeatureID: "F-1"}},
	}
	uc := usecases.NewCatastroReferenciaUseCase(repo)

	poligonos, err := uc.ListarPoligonos(ctx)
	if err != nil || len(poligonos) != 1 {
		t.Fatalf("ListarPoligonos failed: %v", err)
	}

	capa, err := uc.ListarCapa(ctx, "fauna")
	if err != nil || len(capa) != 1 {
		t.Fatalf("ListarCapa failed: %v", err)
	}
}
