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

type mockAreaVerdeRepo struct {
	fichas        map[string]entities.AreaVerdeFicha
	crearError    error
	actualizarErr error
}

func newMockAreaVerdeRepo() *mockAreaVerdeRepo {
	return &mockAreaVerdeRepo{
		fichas: map[string]entities.AreaVerdeFicha{
			"AV-0001": {
				FeatureID:  "AV-0001",
				Nombre:     "Bosque Húmedo",
				Uso:        "Uso Institucional",
				RiegoAct:   "Riego por aspersión",
				Referencia: "",
				ConGeom:    true,
			},
		},
	}
}

func (m *mockAreaVerdeRepo) Fichas(_ context.Context, q string) ([]entities.AreaVerdeFicha, error) {
	out := []entities.AreaVerdeFicha{}
	for _, f := range m.fichas {
		if q == "" || strings.Contains(f.Nombre, q) || strings.Contains(f.FeatureID, q) {
			out = append(out, f)
		}
	}
	return out, nil
}

func (m *mockAreaVerdeRepo) ObtenerFichaPorFeatureID(_ context.Context, featureID string) (*entities.AreaVerdeFicha, error) {
	if f, ok := m.fichas[featureID]; ok {
		return &f, nil
	}
	return nil, domainErrors.ErrFichaNoEncontrada
}

func (m *mockAreaVerdeRepo) ActualizarFicha(_ context.Context, featureID, nombre, uso, riego, referencia string) (*entities.AreaVerdeFicha, error) {
	if m.actualizarErr != nil {
		return nil, m.actualizarErr
	}
	f, ok := m.fichas[featureID]
	if !ok {
		return nil, domainErrors.ErrFichaNoEncontrada
	}
	f.Nombre = nombre
	f.Uso = uso
	f.RiegoAct = riego
	f.Referencia = referencia
	m.fichas[featureID] = f
	return &f, nil
}

func (m *mockAreaVerdeRepo) CrearSinGeom(_ context.Context, featureID, nombre, uso string) (*entities.AreaVerdeFicha, error) {
	if m.crearError != nil {
		return nil, m.crearError
	}
	f := entities.AreaVerdeFicha{
		FeatureID: featureID,
		Nombre:    nombre,
		Uso:       uso,
		ConGeom:   false,
	}
	m.fichas[featureID] = f
	return &f, nil
}

func TestAreaVerdeUseCase_ValidacionesReferenciaYNombre(t *testing.T) {
	repo := newMockAreaVerdeRepo()
	uc := usecases.NewAreaVerdeUseCase(repo)
	ctx := context.Background()
	var userID int64 = 1

	// Referencia larga: la API vieja no la limita en fichas (columna text).
	refLarga := strings.Repeat("r", 501)
	if _, err := uc.ActualizarFicha(ctx, "AV-0001", dto.ActualizarFichaDTO{
		Nombre:     "Nombre Válido",
		Referencia: refLarga,
	}, &userID); err != nil {
		t.Fatalf("referencia > 500 caracteres debía aceptarse como en la API vieja, obtenido: %v", err)
	}

	// Nombre mayor a 160 caracteres
	nombreLargo := string(make([]rune, 161))
	_, err := uc.ActualizarFicha(ctx, "AV-0001", dto.ActualizarFichaDTO{
		Nombre: nombreLargo,
	}, &userID)
	if err != domainErrors.ErrEntrada {
		t.Fatalf("nombre > 160 caracteres debía fallar con ErrEntrada, obtenido: %v", err)
	}

	// Crear sin nombre
	_, err = uc.CrearSinGeom(ctx, dto.CrearAreaSinGeomDTO{
		Nombre: "   ",
	}, &userID)
	if err != domainErrors.ErrEntrada {
		t.Fatalf("crear sin nombre debía fallar con ErrEntrada, obtenido: %v", err)
	}

	// Crear con feature_id inválido (con espacios o > 40)
	_, err = uc.CrearSinGeom(ctx, dto.CrearAreaSinGeomDTO{
		FeatureID: "AV CON ESPACIOS",
		Nombre:    "Nombre OK",
	}, &userID)
	if err != domainErrors.ErrEntrada {
		t.Fatalf("feature_id con espacios debía fallar con ErrEntrada, obtenido: %v", err)
	}

	// Crear con nombre > 160: se rechaza antes de insertar (la vieja insertaba y luego daba 400).
	antes := len(repo.fichas)
	_, err = uc.CrearSinGeom(ctx, dto.CrearAreaSinGeomDTO{Nombre: strings.Repeat("n", 161)}, &userID)
	if err != domainErrors.ErrEntrada {
		t.Fatalf("crear con nombre > 160 debía fallar con ErrEntrada, obtenido: %v", err)
	}
	if len(repo.fichas) != antes {
		t.Fatalf("crear con nombre > 160 no debía insertar filas")
	}
}

func TestAreaVerdeUseCase_Actualizar(t *testing.T) {
	repo := newMockAreaVerdeRepo()
	uc := usecases.NewAreaVerdeUseCase(repo)
	ctx := context.Background()
	var userID int64 = 5

	updated, err := uc.ActualizarFicha(ctx, "AV-0001", dto.ActualizarFichaDTO{
		Nombre:     "Bosque Tropical",
		Uso:        "Uso Científico",
		RiegoAct:   "goteo",
		Referencia: "Junto al río",
	}, &userID)
	if err != nil {
		t.Fatalf("error actualizando ficha: %v", err)
	}
	if updated.Nombre != "Bosque Tropical" {
		t.Fatalf("nombre no actualizado: %s", updated.Nombre)
	}
}

func TestAreaVerdeUseCase_Crear(t *testing.T) {
	repo := newMockAreaVerdeRepo()
	uc := usecases.NewAreaVerdeUseCase(repo)
	ctx := context.Background()
	var userID int64 = 6

	creada, err := uc.CrearSinGeom(ctx, dto.CrearAreaSinGeomDTO{
		FeatureID: "AV-NUEVA-01",
		Nombre:    "Jardín Nuevo",
		Uso:       "Recreativo",
	}, &userID)
	if err != nil {
		t.Fatalf("error creando ficha: %v", err)
	}
	if creada.FeatureID != "AV-NUEVA-01" || creada.Nombre != "Jardín Nuevo" {
		t.Fatalf("datos incorrectos en ficha creada: %+v", creada)
	}
}

func (m *mockAreaVerdeRepo) Baja(_ context.Context, featureID string, _ *int64) error {
	if _, ok := m.fichas[featureID]; !ok {
		return domainErrors.ErrFichaNoEncontrada
	}
	delete(m.fichas, featureID)
	return nil
}

func TestAreaVerdeUseCase_Baja(t *testing.T) {
	repo := newMockAreaVerdeRepo()
	uc := usecases.NewAreaVerdeUseCase(repo)
	ctx := context.Background()
	var userID int64 = 3
	if err := uc.Baja(ctx, "AV-0001", &userID); err != nil {
		t.Fatalf("baja: %v", err)
	}
	if _, ok := repo.fichas["AV-0001"]; ok {
		t.Fatal("la ficha sigue en el mock")
	}
	if err := uc.Baja(ctx, "AV-NO", &userID); !errors.Is(err, domainErrors.ErrFichaNoEncontrada) {
		t.Fatalf("esperado no encontrada, obtenido %v", err)
	}
}
