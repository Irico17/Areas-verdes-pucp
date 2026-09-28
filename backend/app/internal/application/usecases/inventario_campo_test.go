package usecases_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/dto"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/usecases"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/entities"
	domainErrors "github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/errors"
)

type mockInventarioCampoRepo struct {
	tachos    []entities.Tacho
	bebederos []entities.Bebedero
	puntos    []entities.PuntoPUCP
	reservas  []entities.ReservaJardin
	fichas    map[string][]entities.FichaCapa
	lastID    int64
	failErr   error
}

func newMockInventarioCampoRepo() *mockInventarioCampoRepo {
	return &mockInventarioCampoRepo{
		fichas: make(map[string][]entities.FichaCapa),
	}
}

func (m *mockInventarioCampoRepo) ListarTachos(_ context.Context) ([]entities.Tacho, error) {
	if m.failErr != nil {
		return nil, m.failErr
	}
	return m.tachos, nil
}

func (m *mockInventarioCampoRepo) GuardarTacho(_ context.Context, t entities.Tacho) (*entities.Tacho, error) {
	if m.failErr != nil {
		return nil, m.failErr
	}
	m.lastID++
	t.ID = m.lastID
	t.Activo = true
	m.tachos = append(m.tachos, t)
	return &t, nil
}

func (m *mockInventarioCampoRepo) ActualizarTacho(_ context.Context, id int64, t entities.Tacho, _ []byte) (*entities.Tacho, error) {
	if m.failErr != nil {
		return nil, m.failErr
	}
	t.ID = id
	t.Activo = true
	return &t, nil
}

func (m *mockInventarioCampoRepo) ListarBebederos(_ context.Context) ([]entities.Bebedero, error) {
	if m.failErr != nil {
		return nil, m.failErr
	}
	return m.bebederos, nil
}

func (m *mockInventarioCampoRepo) GuardarBebedero(_ context.Context, b entities.Bebedero) (*entities.Bebedero, error) {
	if m.failErr != nil {
		return nil, m.failErr
	}
	m.lastID++
	b.ID = m.lastID
	b.Activo = true
	m.bebederos = append(m.bebederos, b)
	return &b, nil
}

func (m *mockInventarioCampoRepo) ActualizarBebedero(_ context.Context, id int64, b entities.Bebedero, _ []byte) (*entities.Bebedero, error) {
	if m.failErr != nil {
		return nil, m.failErr
	}
	b.ID = id
	b.Activo = true
	return &b, nil
}

func (m *mockInventarioCampoRepo) ListarPuntos(_ context.Context, q string) ([]entities.PuntoPUCP, error) {
	if m.failErr != nil {
		return nil, m.failErr
	}
	if q == "" {
		return m.puntos, nil
	}
	var out []entities.PuntoPUCP
	for _, p := range m.puntos {
		if strings.Contains(strings.ToLower(p.Titulo), strings.ToLower(q)) {
			out = append(out, p)
		}
	}
	return out, nil
}

func (m *mockInventarioCampoRepo) GuardarPunto(_ context.Context, p entities.PuntoPUCP) (*entities.PuntoPUCP, error) {
	if m.failErr != nil {
		return nil, m.failErr
	}
	m.lastID++
	p.ID = m.lastID
	p.Activo = true
	m.puntos = append(m.puntos, p)
	return &p, nil
}

func (m *mockInventarioCampoRepo) ActualizarPunto(_ context.Context, id int64, p entities.PuntoPUCP, _ []byte) (*entities.PuntoPUCP, error) {
	if m.failErr != nil {
		return nil, m.failErr
	}
	p.ID = id
	p.Activo = true
	return &p, nil
}

func (m *mockInventarioCampoRepo) ListarReservas(_ context.Context, _, _ string) ([]entities.ReservaJardin, error) {
	if m.failErr != nil {
		return nil, m.failErr
	}
	return m.reservas, nil
}

func (m *mockInventarioCampoRepo) GuardarReserva(_ context.Context, r entities.ReservaJardin) (*entities.ReservaJardin, error) {
	if m.failErr != nil {
		return nil, m.failErr
	}
	m.lastID++
	r.ID = m.lastID
	r.Activo = true
	m.reservas = append(m.reservas, r)
	return &r, nil
}

func (m *mockInventarioCampoRepo) ActualizarReserva(_ context.Context, id int64, r entities.ReservaJardin, _ []byte) (*entities.ReservaJardin, error) {
	if m.failErr != nil {
		return nil, m.failErr
	}
	r.ID = id
	r.Activo = true
	return &r, nil
}

func (m *mockInventarioCampoRepo) ListarFichas(_ context.Context, capa string) ([]entities.FichaCapa, error) {
	if m.failErr != nil {
		return nil, m.failErr
	}
	return m.fichas[capa], nil
}

func (m *mockInventarioCampoRepo) GuardarFicha(_ context.Context, capa string, f entities.FichaCapa) (*entities.FichaCapa, error) {
	if m.failErr != nil {
		return nil, m.failErr
	}
	m.lastID++
	f.ID = m.lastID
	f.Activo = true
	m.fichas[capa] = append(m.fichas[capa], f)
	return &f, nil
}

func (m *mockInventarioCampoRepo) ActualizarFicha(_ context.Context, _ string, id int64, f entities.FichaCapa, _ []byte) (*entities.FichaCapa, error) {
	if m.failErr != nil {
		return nil, m.failErr
	}
	f.ID = id
	f.Activo = true
	return &f, nil
}

func (m *mockInventarioCampoRepo) Baja(_ context.Context, _ string, id int64) error {
	if m.failErr != nil {
		return m.failErr
	}
	if id < 1 {
		return domainErrors.ErrRegistroNoEncontrado
	}
	return nil
}

type mockParser struct {
	failErr bool
}

func (p *mockParser) LeerPuntosPUCP(_ []byte) ([]entities.PuntoCarga, []entities.RechazoFormato, []string, error) {
	if p.failErr {
		return nil, nil, nil, errors.New("bad csv")
	}
	return []entities.PuntoCarga{
		{Titulo: "Punto 1", Lat: -12.07, Lon: -77.08, URL: "https://example.com"},
	}, []entities.RechazoFormato{}, []string{"phone", "placeId"}, nil
}

func TestInventarioCampoUseCase_Tachos(t *testing.T) {
	repo := newMockInventarioCampoRepo()
	uc := usecases.NewInventarioCampoUseCase(repo, &mockParser{})
	ctx := context.Background()

	// 1. Validacion: codigo sin PT
	_, err := uc.GuardarTacho(ctx, dto.TachoDTO{Codigo: "INVALID", NoAprovechables: 1})
	if !errors.Is(err, domainErrors.ErrTachoInvalido) {
		t.Fatalf("se esperaba ErrTachoInvalido, obtenido %v", err)
	}

	// 2. Validacion: conteo negativo
	_, err = uc.GuardarTacho(ctx, dto.TachoDTO{Codigo: "PT-01", NoAprovechables: -1})
	if !errors.Is(err, domainErrors.ErrTachoInvalido) {
		t.Fatalf("se esperaba ErrTachoInvalido por conteo negativo, obtenido %v", err)
	}

	// 3. Guardar exitoso
	saved, err := uc.GuardarTacho(ctx, dto.TachoDTO{Codigo: "PT-01", NoAprovechables: 5, PapelCarton: 2})
	if err != nil {
		t.Fatalf("GuardarTacho fallo: %v", err)
	}
	if saved.ID != 1 || saved.Codigo != "PT-01" {
		t.Fatalf("saved inesperado: %+v", saved)
	}

	// 4. Listar
	res, err := uc.ListarTachos(ctx)
	if err != nil || len(res.Tachos) != 1 {
		t.Fatalf("ListarTachos inesperado: %+v, err: %v", res, err)
	}

	// 5. CSV export
	csvStr, err := uc.CSVTachos(ctx)
	if err != nil {
		t.Fatalf("CSVTachos fallo: %v", err)
	}
	if !strings.Contains(csvStr, "codigo,lugar") || !strings.Contains(csvStr, "PT-01") {
		t.Fatalf("CSV inesperado: %s", csvStr)
	}

	// 6. Actualizar parcial
	updated, err := uc.ActualizarTacho(ctx, saved.ID, dto.TachoDTO{Nota: "nueva nota"}, []byte(`{"nota":"nueva nota"}`))
	if err != nil {
		t.Fatalf("ActualizarTacho fallo: %v", err)
	}
	if updated.ID != saved.ID {
		t.Fatalf("updated ID inesperado: %d", updated.ID)
	}

	// 7. Baja
	if err := uc.BajaTacho(ctx, saved.ID); err != nil {
		t.Fatalf("BajaTacho fallo: %v", err)
	}
}

func TestInventarioCampoUseCase_Bebederos(t *testing.T) {
	repo := newMockInventarioCampoRepo()
	uc := usecases.NewInventarioCampoUseCase(repo, &mockParser{})
	ctx := context.Background()

	// 1. Subtipo invalido
	_, err := uc.GuardarBebedero(ctx, dto.BebederoDTO{Codigo: "PT_01", Subtipo: "invalido", Estado: "ok"})
	if !errors.Is(err, domainErrors.ErrBebederoInvalido) {
		t.Fatalf("se esperaba ErrBebederoInvalido por subtipo, obtenido %v", err)
	}

	// 2. Codigo sin PT_
	_, err = uc.GuardarBebedero(ctx, dto.BebederoDTO{Codigo: "NO_PREFIJO", Subtipo: "fuente", Estado: "ok"})
	if !errors.Is(err, domainErrors.ErrBebederoInvalido) {
		t.Fatalf("se esperaba ErrBebederoInvalido por codigo, obtenido %v", err)
	}

	// 3. Guardar exitoso
	saved, err := uc.GuardarBebedero(ctx, dto.BebederoDTO{Codigo: "PT_BEB_1", Subtipo: "fuente", Estado: "operativo"})
	if err != nil {
		t.Fatalf("GuardarBebedero fallo: %v", err)
	}
	if saved.ID != 1 {
		t.Fatalf("ID inesperado: %d", saved.ID)
	}

	// 4. Listar
	listRes, err := uc.ListarBebederos(ctx)
	if err != nil || len(listRes.Bebederos) != 1 {
		t.Fatalf("ListarBebederos inesperado: %+v", listRes)
	}

	// 5. Baja
	if err := uc.BajaBebedero(ctx, saved.ID); err != nil {
		t.Fatalf("BajaBebedero fallo: %v", err)
	}
}

func TestInventarioCampoUseCase_Puntos(t *testing.T) {
	repo := newMockInventarioCampoRepo()
	uc := usecases.NewInventarioCampoUseCase(repo, &mockParser{})
	ctx := context.Background()

	// 1. Datos personales (phone/placeId)
	rawBody := []byte(`{"titulo":"Comedor","phone":"999999999"}`)
	_, err := uc.GuardarPunto(ctx, dto.PuntoDTO{Titulo: "Comedor"}, rawBody)
	if !errors.Is(err, domainErrors.ErrPuntoContacto) {
		t.Fatalf("se esperaba ErrPuntoContacto, obtenido %v", err)
	}

	// 2. Coordenadas fuera de campus
	_, err = uc.GuardarPunto(ctx, dto.PuntoDTO{Titulo: "Comedor", Lat: -10.0, Lon: -75.0}, []byte(`{}`))
	if !errors.Is(err, domainErrors.ErrPuntoInvalido) {
		t.Fatalf("se esperaba ErrPuntoInvalido por lat/lon fuera de campus, obtenido %v", err)
	}

	// 3. Guardar exitoso
	saved, err := uc.GuardarPunto(ctx, dto.PuntoDTO{Titulo: "Comedor Central", Lat: -12.071, Lon: -77.081, URL: "https://maps.example.com"}, []byte(`{}`))
	if err != nil {
		t.Fatalf("GuardarPunto fallo: %v", err)
	}
	if saved.ID != 1 {
		t.Fatalf("ID inesperado: %d", saved.ID)
	}

	// 4. Formato puntos
	formatoRes, err := uc.FormatoPuntos(ctx, []byte(`csv body`))
	if err != nil {
		t.Fatalf("FormatoPuntos fallo: %v", err)
	}
	if formatoRes.Filas != 1 || len(formatoRes.ColumnasOmitidas) != 2 {
		t.Fatalf("formatoRes inesperado: %+v", formatoRes)
	}
}

func TestInventarioCampoUseCase_Reservas(t *testing.T) {
	repo := newMockInventarioCampoRepo()
	uc := usecases.NewInventarioCampoUseCase(repo, &mockParser{})
	ctx := context.Background()

	// 1. Origen no ficticio
	_, err := uc.GuardarReserva(ctx, dto.ReservaDTO{Origen: "real", Estado: "reservado", Fecha: "2026-10-01", HoraInicio: "09:00", HoraFin: "10:00", Evento: "Test"})
	if !errors.Is(err, domainErrors.ErrReservaOrigen) {
		t.Fatalf("se esperaba ErrReservaOrigen, obtenido %v", err)
	}

	// 2. HoraFin <= HoraInicio
	_, err = uc.GuardarReserva(ctx, dto.ReservaDTO{Estado: "reservado", Fecha: "2026-10-01", HoraInicio: "11:00", HoraFin: "10:00", Evento: "Test"})
	if !errors.Is(err, domainErrors.ErrReservaInvalida) {
		t.Fatalf("se esperaba ErrReservaInvalida, obtenido %v", err)
	}

	// 3. Guardar exitoso
	saved, err := uc.GuardarReserva(ctx, dto.ReservaDTO{Estado: "reservado", Fecha: "2026-10-01", HoraInicio: "09:00", HoraFin: "11:00", Evento: "Test"})
	if err != nil {
		t.Fatalf("GuardarReserva fallo: %v", err)
	}
	if saved.ID != 1 {
		t.Fatalf("ID inesperado: %d", saved.ID)
	}

	// 4. Listar
	res, err := uc.ListarReservas(ctx, "", "")
	if err != nil || len(res.Reservas) != 1 {
		t.Fatalf("ListarReservas fallo: %+v", res)
	}
	if res.Origen != "ficticio" || !strings.Contains(res.Aviso, "ficticia") {
		t.Fatalf("Aviso u origen inesperado: %+v", res)
	}
}

func TestInventarioCampoUseCase_Capas(t *testing.T) {
	repo := newMockInventarioCampoRepo()
	uc := usecases.NewInventarioCampoUseCase(repo, &mockParser{})
	ctx := context.Background()

	// 1. FeatureID vacio
	_, err := uc.GuardarCapa(ctx, "fauna", dto.FichaCapaDTO{FeatureID: "   "})
	if !errors.Is(err, domainErrors.ErrFichaInvalida) {
		t.Fatalf("se esperaba ErrFichaInvalida por FeatureID vacio, obtenido %v", err)
	}

	// 2. Guardar exitoso
	saved, err := uc.GuardarCapa(ctx, "fauna", dto.FichaCapaDTO{FeatureID: "FAU-01", Nombre: "Ardilla"})
	if err != nil {
		t.Fatalf("GuardarCapa fallo: %v", err)
	}
	if saved.ID != 1 {
		t.Fatalf("ID inesperado: %d", saved.ID)
	}

	// 3. Listar
	listRes, err := uc.ListarCapa(ctx, "fauna")
	if err != nil || len(listRes.Filas) != 1 {
		t.Fatalf("ListarCapa fallo: %+v", listRes)
	}

	// 4. CSV export
	csvStr, err := uc.CSVCapa(ctx, "fauna")
	if err != nil {
		t.Fatalf("CSVCapa fallo: %v", err)
	}
	if !strings.Contains(csvStr, "feature_id,nombre") || !strings.Contains(csvStr, "FAU-01") {
		t.Fatalf("CSVCapa inesperado: %s", csvStr)
	}
}
