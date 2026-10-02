package enums_test

import (
	"testing"

	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/constants/enums"
)

func TestClaseCatalogo_ValoresYValidez(t *testing.T) {
	expectedClases := []string{
		"tipo_actividad",
		"estado",
		"prioridad",
		"lugar",
		"especie",
		"motivo_archivo",
		"turno",
		"fuente",
		"clase_actividad",
		"plaga",
		"producto_fitosanitario",
		"frecuencia",
		"sede",
		"cuartel",
		"sector_capataz",
		"nivel_riesgo",
		"subtipo_actividad",
	}

	clases := enums.ClasesCatalogoValidas()
	if len(clases) != len(expectedClases) {
		t.Fatalf("se esperaban %d clases, se obtuvieron %d", len(expectedClases), len(clases))
	}

	for i, c := range expectedClases {
		if clases[i] != c {
			t.Errorf("clase en posición %d: esperada %q, obtenida %q", i, c, clases[i])
		}
		if !enums.ClaseCatalogo(c).EsValida() {
			t.Errorf("clase %q debería ser válida", c)
		}
	}

	invalidas := []string{"", "no_existe", "TIPO_ACTIVIDAD", "rol", "usuario"}
	for _, inv := range invalidas {
		if enums.ClaseCatalogo(inv).EsValida() {
			t.Errorf("clase %q no debería ser válida", inv)
		}
	}
}

func TestClaseCatalogo_String(t *testing.T) {
	c := enums.ClaseTipoActividad
	if c.String() != "tipo_actividad" {
		t.Errorf("esperado 'tipo_actividad', obtenido %q", c.String())
	}
}
