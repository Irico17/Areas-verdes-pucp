package etl

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/shared/testutil"
)

// TestCargarAreasVerdesResuelveDuplicadosYNoPisaEdiciones reproduce el catastro real
// (código repetido, p. ej. "D 20") por el camino de etl-lote (CargarAreasVerdes, que
// hace upsert sin TRUNCATE). No debe fallar por areas_verdes_codigo_uidx en una base
// vacía, debe ser idempotente, y no debe pisar un código ya renombrado por
// deduplicación ni uno editado a mano desde la ficha.
func TestCargarAreasVerdesResuelveDuplicadosYNoPisaEdiciones(t *testing.T) {
	_, gdb := testutil.MigrarDBTemporal(t, "campus_verde_etl_2b_upsert")

	rawDir := t.TempDir()
	geojson := `{
		"type": "FeatureCollection",
		"features": [
			{"type":"Feature","properties":{"código":"D 20","Nombre":"Área A"},"geometry":{"type":"Polygon","coordinates":[[[-77.08,-12.07],[-77.079,-12.07],[-77.079,-12.069],[-77.08,-12.069],[-77.08,-12.07]]]}},
			{"type":"Feature","properties":{"código":"G 13","Nombre":"Área B"},"geometry":{"type":"Polygon","coordinates":[[[-77.08,-12.07],[-77.079,-12.07],[-77.079,-12.069],[-77.08,-12.069],[-77.08,-12.07]]]}},
			{"type":"Feature","properties":{"código":"D 20","Nombre":"Área C"},"geometry":{"type":"Polygon","coordinates":[[[-77.08,-12.07],[-77.079,-12.07],[-77.079,-12.069],[-77.08,-12.069],[-77.08,-12.07]]]}}
		]
	}`
	if err := os.WriteFile(filepath.Join(rawDir, "areas_verdes.geojson"), []byte(geojson), 0o644); err != nil {
		t.Fatal(err)
	}

	// Primera corrida sobre base vacía: no debe fallar por el índice único de codigo.
	n, err := CargarAreasVerdes(gdb, rawDir)
	if err != nil {
		t.Fatalf("primera carga: %v", err)
	}
	if n != 3 {
		t.Fatalf("cargadas = %d, se esperaban 3", n)
	}

	var codigoA string
	if err := gdb.Raw(`SELECT codigo FROM areas_verdes WHERE feature_id = 'AV-0001'`).Scan(&codigoA).Error; err != nil {
		t.Fatal(err)
	}
	if codigoA != "D 20" {
		t.Fatalf("AV-0001 (menor id) codigo = %q, se esperaba sin cambios %q", codigoA, "D 20")
	}
	var idC int64
	var codigoC string
	if err := gdb.Raw(`SELECT id, codigo FROM areas_verdes WHERE feature_id = 'AV-0003'`).Row().Scan(&idC, &codigoC); err != nil {
		t.Fatal(err)
	}
	esperado := fmt.Sprintf("D 20 ·%d", idC)
	if codigoC != esperado {
		t.Fatalf("AV-0003 (mayor id, duplicado) codigo = %q, se esperaba %q", codigoC, esperado)
	}

	var totalCambios int
	if err := gdb.Raw(`SELECT count(*) FROM cambios WHERE entidad = 'areas_verdes'`).Scan(&totalCambios).Error; err != nil {
		t.Fatal(err)
	}
	if totalCambios != 1 {
		t.Fatalf("cambios registrados = %d, se esperaba 1", totalCambios)
	}

	// Edición manual desde la ficha sobre una fila que no es duplicada.
	if err := gdb.Exec(`UPDATE areas_verdes SET codigo = 'G-EDITADO' WHERE feature_id = 'AV-0002'`).Error; err != nil {
		t.Fatal(err)
	}

	// Segunda corrida (base ya migrada): no debe pisar el renombre por duplicado
	// ni la edición manual, y no debe duplicar el registro en cambios.
	if _, err := CargarAreasVerdes(gdb, rawDir); err != nil {
		t.Fatalf("segunda carga: %v", err)
	}

	var totalAreas int
	if err := gdb.Raw(`SELECT count(*) FROM areas_verdes`).Scan(&totalAreas).Error; err != nil {
		t.Fatal(err)
	}
	if totalAreas != 3 {
		t.Fatalf("segunda corrida: areas_verdes = %d, se esperaban 3 (ninguna fila borrada)", totalAreas)
	}

	if err := gdb.Raw(`SELECT codigo FROM areas_verdes WHERE feature_id = 'AV-0001'`).Scan(&codigoA).Error; err != nil {
		t.Fatal(err)
	}
	if codigoA != "D 20" {
		t.Fatalf("segunda corrida: AV-0001 codigo = %q, se esperaba seguir en %q", codigoA, "D 20")
	}
	if err := gdb.Raw(`SELECT codigo FROM areas_verdes WHERE feature_id = 'AV-0003'`).Scan(&codigoC).Error; err != nil {
		t.Fatal(err)
	}
	if codigoC != esperado {
		t.Fatalf("segunda corrida: AV-0003 codigo = %q, se esperaba seguir en %q", codigoC, esperado)
	}
	var codigoB string
	if err := gdb.Raw(`SELECT codigo FROM areas_verdes WHERE feature_id = 'AV-0002'`).Scan(&codigoB).Error; err != nil {
		t.Fatal(err)
	}
	if codigoB != "G-EDITADO" {
		t.Fatalf("segunda corrida: AV-0002 codigo = %q, se esperaba conservar la edición manual %q", codigoB, "G-EDITADO")
	}

	if err := gdb.Raw(`SELECT count(*) FROM cambios WHERE entidad = 'areas_verdes'`).Scan(&totalCambios).Error; err != nil {
		t.Fatal(err)
	}
	if totalCambios != 1 {
		t.Fatalf("segunda corrida: cambios registrados = %d, se esperaba seguir en 1 (sin duplicar el historial)", totalCambios)
	}
}
