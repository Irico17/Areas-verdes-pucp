package usecases_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/usecases"
	domainErrors "github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/errors"
)

type mockReservasAdapter struct {
	readFunc func(ctx context.Context) ([]byte, error)
}

func (m *mockReservasAdapter) LeerReservas(ctx context.Context) ([]byte, error) {
	if m.readFunc != nil {
		return m.readFunc(ctx)
	}
	return nil, nil
}

// TestReservasSinSheet ports apps/api/internal/handlers/reservas_test.go:TestReservasSinSheet
func TestReservasSinSheet(t *testing.T) {
	body := []byte(`{
	  "_meta": {"fake": true, "warning": "SHEET_ID_RESERVAS=1R3Xz8A5xIVm-s0duMQhav2YAJrQdlSAYgKYeoEvoT8s"},
	  "reservas": [{
	    "id": "FAKE-RES-001",
	    "jardin": "Jardín Rosales",
	    "fecha": "2026-09-22",
	    "hora": "15:00-17:00",
	    "evento": "Feria",
	    "estado": "Confirmado",
	    "notas": "viene del Sheet real"
	  }]
	}`)

	adapter := &mockReservasAdapter{
		readFunc: func(_ context.Context) ([]byte, error) {
			return body, nil
		},
	}
	uc := usecases.NewReservasMockUseCase(adapter)

	resp, encoded, err := uc.ObtenerAgenda(context.Background())
	if err != nil {
		t.Fatalf("error inesperado: %v", err)
	}

	if !resp.Fake || resp.Total != 1 || len(resp.Reservas) != 1 {
		t.Fatalf("respuesta estructurada inesperada: %+v", resp)
	}

	text := string(encoded)
	for _, forbidden := range []string{"SHEET", "1R3Xz", "docs.google.com", "viene del Sheet"} {
		if strings.Contains(text, forbidden) {
			t.Fatalf("se filtró %q en %s", forbidden, text)
		}
	}
	if !strings.Contains(text, "Agenda ficticia") || !strings.Contains(text, "Jardín Rosales") {
		t.Fatalf("respuesta %s", text)
	}
	if resp.Reservas[0].Notas != "Dato ficticio de demostración." || !resp.Reservas[0].Fake {
		t.Fatalf("notas o fake no sanitizados: %+v", resp.Reservas[0])
	}
}

func TestReservas_ArchivoInexistente(t *testing.T) {
	adapter := &mockReservasAdapter{
		readFunc: func(_ context.Context) ([]byte, error) {
			return nil, errors.New("file not found")
		},
	}
	uc := usecases.NewReservasMockUseCase(adapter)

	resp, encoded, err := uc.ObtenerAgenda(context.Background())
	if err != nil {
		t.Fatalf("error inesperado con archivo ausente: %v", err)
	}
	if !resp.Fake || resp.Total != 0 || len(resp.Reservas) != 0 {
		t.Fatalf("se esperaba agenda vacía, obtenido %+v", resp)
	}
	if !strings.Contains(string(encoded), `"total":0`) {
		t.Fatalf("json inesperado: %s", string(encoded))
	}
}

func TestReservas_FiltraReferenciaExterna(t *testing.T) {
	body := []byte(`{
	  "reservas": [{
	    "id": "RES-001",
	    "jardin": "Jardín con enlace a docs.google.com/spreadsheets"
	  }]
	}`)

	adapter := &mockReservasAdapter{
		readFunc: func(_ context.Context) ([]byte, error) {
			return body, nil
		},
	}
	uc := usecases.NewReservasMockUseCase(adapter)

	_, _, err := uc.ObtenerAgenda(context.Background())
	if !errors.Is(err, domainErrors.ErrAgendaFicticiaReferenciaExterna) {
		t.Fatalf("se esperaba ErrAgendaFicticiaReferenciaExterna, obtenido %v", err)
	}
}
