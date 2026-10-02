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

type riegoRepoFijo struct{}

func (riegoRepoFijo) Listar(context.Context, string) ([]*entities.TurnoRiego, error) {
	return nil, nil
}

func (riegoRepoFijo) Crear(context.Context, entities.NuevoTurnoRiego) error {
	return errors.New("no debía persistir un sector de texto libre")
}

func (riegoRepoFijo) CoberturaMes(context.Context) (int, int, error) {
	return 0, 0, nil
}

func TestConsultaRiegoSoloElEquipoDelCapataz(t *testing.T) {
	q, args := usecases.ConsultaRiego("cap-norte")
	if !strings.Contains(q, "WHERE r.capataz_id = $1") {
		t.Fatalf("el capataz debe filtrar por equipo: %s", q)
	}
	if len(args) != 1 || args[0] != "cap-norte" {
		t.Fatalf("argumento: %#v", args)
	}
}

func TestCrearRiegoRechazaSectorLibre(t *testing.T) {
	uc := usecases.NewRiegoUseCase(riegoRepoFijo{})
	_, err := uc.Crear(context.Background(), dto.CrearRiegoDTO{
		ID:     "11111111-1111-4111-8111-111111111111",
		Sector: "Eje central",
		Turno:  "manana",
		ZonaID: "Z1",
		Fecha:  "2026-10-02",
	}, "coordinacion", "")
	var input domainErrors.InputError
	if !errors.As(err, &input) {
		t.Fatalf("se esperaba rechazo del texto libre, obtuve %v", err)
	}
}

func TestConsultaRiegoOficinaVeTodos(t *testing.T) {
	q, args := usecases.ConsultaRiego("  ")
	if strings.Contains(q, "WHERE") {
		t.Fatalf("oficina no debe filtrar: %s", q)
	}
	if len(args) != 0 {
		t.Fatalf("sin argumentos: %#v", args)
	}
}
