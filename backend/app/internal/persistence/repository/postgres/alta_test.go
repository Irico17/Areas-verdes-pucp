package postgres_test

import (
	"errors"
	"testing"

	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/entities"
	apperrors "github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/errors"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/persistence/repository/postgres"
)

func TestAltaDeLaborExigeTituloYEstado(t *testing.T) {
	_, _, err := postgres.CamposAltaActividad(entities.SnapAuditoria{Titulo: "", Estado: "pendiente"})
	var inputErr apperrors.InputError
	if err == nil || !errors.As(err, &inputErr) {
		t.Fatalf("sin título: %v", err)
	}
	_, _, err = postgres.CamposAltaActividad(entities.SnapAuditoria{Titulo: "Poda", Estado: "  "})
	if err == nil {
		t.Fatal("sin estado debía rechazarse")
	}
	titulo, estado, err := postgres.CamposAltaActividad(entities.SnapAuditoria{Titulo: " Poda ", Estado: "cerrada"})
	if err != nil || titulo != "Poda" || estado != "cerrada" {
		t.Fatalf("alta válida: %q %q %v", titulo, estado, err)
	}
}
