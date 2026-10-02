package etl

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/shared/testutil"
)

func TestAplicarSectoresNoPisaUnSectorYaAsignado(t *testing.T) {
	_, gdb := testutil.MigrarDBTemporal(t, "campus_verde_etl_sector_a")

	geom := json.RawMessage(`{"type":"MultiPolygon","coordinates":[[[[-77.08,-12.07],[-77.079,-12.07],[-77.079,-12.069],[-77.08,-12.069],[-77.08,-12.07]]]]}`)
	nombre := "Polígono de prueba"
	zona := Record{FeatureID: "Z-0001", SourceIndex: 0, Nombre: &nombre, Geometry: geom}
	if err := Load(gdb, nil, []Record{zona}, map[string][]Record{}); err != nil {
		t.Fatal(err)
	}

	// La migración 044 ya trae source_index=0 → bosque-humedo en
	// poligonos_sector_ref; Load ya debió completarlo.
	var sector string
	if err := gdb.Raw(`SELECT sector FROM poligonos_cuadrilla WHERE feature_id = 'Z-0001'`).Scan(&sector).Error; err != nil {
		t.Fatal(err)
	}
	if sector != "bosque-humedo" {
		t.Fatalf("sector tras Load = %q, se esperaba bosque-humedo", sector)
	}

	// Se fija a mano un sector distinto: AplicarSectores no debe pisarlo.
	if err := gdb.Exec(`UPDATE poligonos_cuadrilla SET sector = 'cua-mateo' WHERE feature_id = 'Z-0001'`).Error; err != nil {
		t.Fatal(err)
	}
	n, err := AplicarSectores(gdb)
	if err != nil {
		t.Fatal(err)
	}
	if n != 0 {
		t.Fatalf("AplicarSectores tocó %d fila(s); no debía tocar ninguna", n)
	}
	if err := gdb.Raw(`SELECT sector FROM poligonos_cuadrilla WHERE feature_id = 'Z-0001'`).Scan(&sector).Error; err != nil {
		t.Fatal(err)
	}
	if sector != "cua-mateo" {
		t.Fatalf("sector = %q, se esperaba conservar el valor fijado a mano", sector)
	}
}

func TestCargarCuadrillasYPoligonosConservaSector(t *testing.T) {
	_, gdb := testutil.MigrarDBTemporal(t, "campus_verde_etl_sector_b")

	var usuario int64
	if err := gdb.Raw(`
		INSERT INTO usuarios (usuario, nombre, rol, password_hash)
		VALUES ('lote.sector', 'Lote de prueba', 'coordinacion', 'no-es-clave')
		RETURNING id`).Row().Scan(&usuario); err != nil {
		t.Fatal(err)
	}
	for _, id := range []int64{1, 2} {
		if err := gdb.Exec(`
			INSERT INTO lotes_importacion (id, entidad, usuario_id) VALUES ($1, 'poligonos_cuadrilla', $2)`,
			id, usuario).Error; err != nil {
			t.Fatal(err)
		}
	}

	fake := fragmentosPersonales()[0]
	geom := `{"type":"MultiPolygon","coordinates":[[[[-77.08,-12.07],[-77.079,-12.07],[-77.079,-12.069],[-77.08,-12.069],[-77.08,-12.07]]]]}`
	body := []byte(fmt.Sprintf(
		`{"type":"FeatureCollection","features":[{"type":"Feature","properties":{"jefes":%q,"Nombre":"Sector de prueba"},"geometry":%s}]}`,
		fake, geom))
	pols, rechazos, err := leerPoligonos(body)
	if err != nil {
		t.Fatal(err)
	}
	if len(rechazos) != 0 || len(pols) != 1 {
		t.Fatalf("pols=%+v rechazos=%+v", pols, rechazos)
	}

	rep := ReporteLote{Cargados: map[string]int{}}
	if err := cargarCuadrillasYPoligonos(gdb, pols, 1, &rep); err != nil {
		t.Fatal(err)
	}
	if rep.Cargados["sectores"] != 1 {
		t.Fatalf("sectores cargados = %d, se esperaba 1 (source_index 0 → bosque-humedo)", rep.Cargados["sectores"])
	}
	var sector string
	if err := gdb.Raw(`SELECT sector FROM poligonos_cuadrilla WHERE feature_id = 'PC-0001'`).Scan(&sector).Error; err != nil {
		t.Fatal(err)
	}
	if sector != "bosque-humedo" {
		t.Fatalf("sector = %q, se esperaba bosque-humedo", sector)
	}

	// Un sector fijado a mano no debe perderse en un segundo lote (el upsert
	// del ON CONFLICT no incluye sector en su SET).
	if err := gdb.Exec(`UPDATE poligonos_cuadrilla SET sector = 'cua-renato' WHERE feature_id = 'PC-0001'`).Error; err != nil {
		t.Fatal(err)
	}
	rep2 := ReporteLote{Cargados: map[string]int{}}
	if err := cargarCuadrillasYPoligonos(gdb, pols, 2, &rep2); err != nil {
		t.Fatal(err)
	}
	if err := gdb.Raw(`SELECT sector FROM poligonos_cuadrilla WHERE feature_id = 'PC-0001'`).Scan(&sector).Error; err != nil {
		t.Fatal(err)
	}
	if sector != "cua-renato" {
		t.Fatalf("segundo lote: sector = %q, el upsert no debía pisarlo", sector)
	}

	texto, err := TextoCargado(gdb, 1)
	if err != nil {
		t.Fatal(err)
	}
	for _, frag := range fragmentosPersonales() {
		if strings.Contains(texto, frag) {
			t.Fatalf("TextoCargado contiene un fragmento personal %q", frag)
		}
	}
}
