package entities_test

import (
	"testing"

	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/entities"
)

func TestCapasConocidasYEsCapaConocida(t *testing.T) {
	if len(entities.CapasConocidas) != 11 {
		t.Fatalf("se esperaban 11 capas conocidas, obtenidas %d", len(entities.CapasConocidas))
	}

	knownCases := []string{
		"bebederos", "fauna", "playas_estacionamiento", "puertas",
		"area_vereda_peligro", "flora", "cafetos", "tachos",
		"xerofitica", "jardines_reserva", "puntos_pucp",
	}
	for _, c := range knownCases {
		if !entities.EsCapaConocida(c) {
			t.Errorf("se esperaba que %q fuera conocida", c)
		}
	}

	unknownCases := []string{"", "desconocida", "arboles", "areas", "edificios"}
	for _, c := range unknownCases {
		if entities.EsCapaConocida(c) {
			t.Errorf("se esperaba que %q fuera desconocida", c)
		}
	}
}
