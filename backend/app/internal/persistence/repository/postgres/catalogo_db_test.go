package postgres_test

import (
	"context"
	"errors"
	"strconv"
	"testing"

	apperrors "github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/errors"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/persistence/repository/postgres"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/shared/testutil"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func TestCatalogoRepository_CRUD(t *testing.T) {
	rawDB, gdb := testutil.MigrarDBTemporal(t, "catalogo_repo")
	_ = rawDB
	ctx := context.Background()
	repo := postgres.NewCatalogoRepository(gdb)

	// 1. Create new item
	item, err := repo.Create(ctx, "tipo_actividad", "nuevo_riego", "Riego Tecnificado")
	if err != nil {
		t.Fatalf("error creando item de catalogo: %v", err)
	}
	if item.ID == 0 || item.Clase != "tipo_actividad" || item.Codigo != "nuevo_riego" || item.Nombre != "Riego Tecnificado" || !item.Activo {
		t.Fatalf("item creado inesperado: %+v", item)
	}

	// 2. Activo returns true
	activo, err := repo.Activo(ctx, "tipo_actividad", "nuevo_riego")
	if err != nil {
		t.Fatalf("error verificando activo: %v", err)
	}
	if !activo {
		t.Fatal("se esperaba que el item estuviera activo")
	}

	// 3. Deactivate sets activo = false
	if err := repo.Deactivate(ctx, item.ID, 0); err != nil {
		t.Fatalf("error desactivando item: %v", err)
	}

	activo, err = repo.Activo(ctx, "tipo_actividad", "nuevo_riego")
	if err != nil {
		t.Fatalf("error verificando activo post-desactivar: %v", err)
	}
	if activo {
		t.Fatal("se esperaba que el item no estuviera activo")
	}

	// 4. Create on conflict updates name and reactivates (activo = true)
	reactivado, err := repo.Create(ctx, "tipo_actividad", "nuevo_riego", "Riego Actualizado")
	if err != nil {
		t.Fatalf("error reactivando item: %v", err)
	}
	if reactivado.ID != item.ID {
		t.Fatalf("se esperaba mantener id %d, se obtuvo %d", item.ID, reactivado.ID)
	}
	if reactivado.Nombre != "Riego Actualizado" || !reactivado.Activo {
		t.Fatalf("item reactivado inesperado: %+v", reactivado)
	}

	// 5. Deactivate on non-existent id returns ErrItemNoExiste
	err = repo.Deactivate(ctx, 99999999, 0)
	if !errors.Is(err, apperrors.ErrItemNoExiste) {
		t.Fatalf("se esperaba ErrItemNoExiste, se obtuvo %v", err)
	}

	// 6. List with filter
	list, err := repo.List(ctx, "tipo_actividad", false)
	if err != nil {
		t.Fatalf("error listando: %v", err)
	}
	found := false
	for _, it := range list {
		if it.Codigo == "nuevo_riego" {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("item creado no fue encontrado en la lista")
	}

	// 7. Verify no physical deletes occur: count total rows before and after deactivation
	var countBefore int64
	if err := gdb.Raw("SELECT count(*) FROM catalogos").Scan(&countBefore).Error; err != nil {
		t.Fatal(err)
	}
	if err := repo.Deactivate(ctx, item.ID, 0); err != nil {
		t.Fatal(err)
	}
	var countAfter int64
	if err := gdb.Raw("SELECT count(*) FROM catalogos").Scan(&countAfter).Error; err != nil {
		t.Fatal(err)
	}
	if countBefore != countAfter {
		t.Fatalf("el conteo total de filas cambió tras desactivar: antes=%d, despues=%d", countBefore, countAfter)
	}
}

func TestCatalogoRepository_RenombrarConservaFilaYHistorial(t *testing.T) {
	_, gdb := testutil.MigrarDBTemporal(t, "cat_nom")
	ctx := context.Background()
	repo := postgres.NewCatalogoRepository(gdb)

	item, err := repo.Create(ctx, "lugar", "sector_demo", "Sector demo")
	if err != nil {
		t.Fatal(err)
	}
	primero, err := repo.Renombrar(ctx, item.ID, "Sector de demostración", 0)
	if err != nil {
		t.Fatal(err)
	}
	if primero.ID != item.ID || primero.Nombre != "Sector de demostración" || primero.Codigo != "sector_demo" {
		t.Fatalf("renombre inesperado: %+v", primero)
	}
	segundo, err := repo.Renombrar(ctx, item.ID, "Sector norte ficticio", 0)
	if err != nil {
		t.Fatal(err)
	}
	if segundo.ID != item.ID {
		t.Fatalf("el id cambió: %d → %d", item.ID, segundo.ID)
	}

	var filas int
	if err := gdb.Raw(`SELECT count(*) FROM catalogos WHERE clase = 'lugar' AND codigo = 'sector_demo'`).Scan(&filas).Error; err != nil {
		t.Fatal(err)
	}
	if filas != 1 {
		t.Fatalf("filas del ítem = %d, se esperaba 1", filas)
	}
	var ediciones int
	id := strconv.FormatInt(item.ID, 10)
	if err := gdb.Raw(`SELECT count(*) FROM cambios WHERE entidad = 'catalogos' AND entidad_id = $1 AND accion = 'edicion'`, id).Scan(&ediciones).Error; err != nil {
		t.Fatal(err)
	}
	if ediciones != 2 {
		t.Fatalf("ediciones = %d, se esperaban 2", ediciones)
	}
	var antes string
	if err := gdb.Raw(`SELECT antes->>'nombre' FROM cambios WHERE entidad = 'catalogos' AND entidad_id = $1 AND accion = 'edicion' ORDER BY id`, id).Row().Scan(&antes); err != nil {
		t.Fatal(err)
	}
	if antes != "Sector demo" {
		t.Fatalf("la edición anterior no conservó el nombre %q", antes)
	}
}

func TestMigracionEstadosYClases(t *testing.T) {
	_, gdb := testutil.MigrarDBTemporal(t, "cat_seed")

	var nombre string
	if err := gdb.Raw(`SELECT nombre FROM catalogos WHERE clase = 'estado' AND codigo = 'pendiente'`).Scan(&nombre).Error; err != nil {
		t.Fatal(err)
	}
	if nombre != "Por iniciar" {
		t.Fatalf("etiqueta pendiente = %q", nombre)
	}
	var activo, provisional bool
	if err := gdb.Raw(`SELECT activo, provisional FROM catalogos WHERE clase = 'estado' AND codigo = 'bloqueada'`).Row().Scan(&activo, &provisional); err != nil {
		t.Fatal(err)
	}
	if activo || !provisional {
		t.Fatalf("bloqueada activo=%v provisional=%v", activo, provisional)
	}
	var n int
	if err := gdb.Raw(`SELECT count(*) FROM catalogos WHERE clase = 'clase_actividad' AND codigo IN ('habilitacion', 'fitosanitario', 'inspeccion_monitoreo')`).Scan(&n).Error; err != nil {
		t.Fatal(err)
	}
	if n != 3 {
		t.Fatalf("clases de actividad conservadas y nuevas = %d", n)
	}
	if err := gdb.Raw(`SELECT count(*) FROM catalogos WHERE clase = 'estado' AND codigo = 'ejecutado' AND nombre = 'Ejecutado' AND activo`).Scan(&n).Error; err != nil {
		t.Fatal(err)
	}
	if n != 1 {
		t.Fatal("falta Ejecutado activo")
	}
	for _, clase := range []string{"plaga", "producto_fitosanitario", "frecuencia", "sede", "cuartel", "sector_capataz", "clase_actividad"} {
		var c int
		if err := gdb.Raw(`SELECT count(*) FROM catalogos WHERE clase = $1`, clase).Scan(&c).Error; err != nil {
			t.Fatal(err)
		}
		if c == 0 {
			t.Fatalf("la clase %s no tiene ítems", clase)
		}
	}
}
