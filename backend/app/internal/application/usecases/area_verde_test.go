package usecases_test

import (
	"context"
	"strings"
	"testing"

	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/dto"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/usecases"
	domainErrors "github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/errors"
)

type mockAuditoriaService struct {
	registrados []dto.RegistrarCambioDTO
}

func (m *mockAuditoriaService) RegistrarCambio(_ context.Context, req dto.RegistrarCambioDTO) error {
	m.registrados = append(m.registrados, req)
	return nil
}

type mockAreaVerdeRepo struct {
	fichas        map[string]dto.FichaDTO
	crearError    error
	actualizarErr error
}

func newMockAreaVerdeRepo() *mockAreaVerdeRepo {
	return &mockAreaVerdeRepo{
		fichas: map[string]dto.FichaDTO{
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

func (m *mockAreaVerdeRepo) Fichas(_ context.Context, q string) ([]dto.FichaDTO, error) {
	out := []dto.FichaDTO{}
	for _, f := range m.fichas {
		if q == "" || strings.Contains(f.Nombre, q) || strings.Contains(f.FeatureID, q) {
			out = append(out, f)
		}
	}
	return out, nil
}

func (m *mockAreaVerdeRepo) ObtenerFichaPorFeatureID(_ context.Context, featureID string) (*dto.FichaDTO, error) {
	if f, ok := m.fichas[featureID]; ok {
		return &f, nil
	}
	return nil, domainErrors.ErrFichaNoEncontrada
}

func (m *mockAreaVerdeRepo) ActualizarFicha(_ context.Context, featureID, nombre, uso, riego, referencia string) (*dto.FichaDTO, error) {
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

func (m *mockAreaVerdeRepo) CrearSinGeom(_ context.Context, featureID, nombre, uso string) (*dto.FichaDTO, error) {
	if m.crearError != nil {
		return nil, m.crearError
	}
	f := dto.FichaDTO{
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
	aud := &mockAuditoriaService{}
	uc := usecases.NewAreaVerdeUseCase(repo, aud)
	ctx := context.Background()
	var userID int64 = 1

	// Referencia mayor a 500 caracteres (portado de catastro/modelo_test.go: TestValidacionesDeCatastro)
	refLarga := string(make([]rune, 501))
	_, err := uc.ActualizarFicha(ctx, "AV-0001", dto.ActualizarFichaDTO{
		Nombre:     "Nombre Válido",
		Referencia: refLarga,
	}, &userID)
	if err != domainErrors.ErrEntrada {
		t.Fatalf("referencia > 500 caracteres debía fallar con ErrEntrada, obtenido: %v", err)
	}

	// Nombre mayor a 160 caracteres
	nombreLargo := string(make([]rune, 161))
	_, err = uc.ActualizarFicha(ctx, "AV-0001", dto.ActualizarFichaDTO{
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
}

func TestAreaVerdeUseCase_ActualizarRegistraCambio(t *testing.T) {
	repo := newMockAreaVerdeRepo()
	aud := &mockAuditoriaService{}
	uc := usecases.NewAreaVerdeUseCase(repo, aud)
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

	if len(aud.registrados) != 1 {
		t.Fatalf("se esperaba 1 cambio registrado, obtenido: %d", len(aud.registrados))
	}
	cambio := aud.registrados[0]
	if cambio.Entidad != "areas_verdes" || cambio.EntidadID != "AV-0001" || cambio.Accion != "edicion" {
		t.Fatalf("metadatos de cambio incorrectos: %+v", cambio)
	}
	if cambio.UsuarioID == nil || *cambio.UsuarioID != 5 {
		t.Fatalf("usuario_id esperado 5, obtenido: %v", cambio.UsuarioID)
	}
}

func TestAreaVerdeUseCase_CrearRegistraCambio(t *testing.T) {
	repo := newMockAreaVerdeRepo()
	aud := &mockAuditoriaService{}
	uc := usecases.NewAreaVerdeUseCase(repo, aud)
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

	if len(aud.registrados) != 1 {
		t.Fatalf("se esperaba 1 cambio registrado en creación, obtenido: %d", len(aud.registrados))
	}
	cambio := aud.registrados[0]
	if cambio.Entidad != "areas_verdes" || cambio.EntidadID != "AV-NUEVA-01" || cambio.Accion != "alta" {
		t.Fatalf("metadatos de alta incorrectos: %+v", cambio)
	}
}
