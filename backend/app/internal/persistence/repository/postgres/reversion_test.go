package postgres_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"testing"

	"gorm.io/gorm"

	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/entities"
	apperrors "github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/errors"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/persistence/repository/postgres"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/shared/testutil"
)

func TestRevertirLoteRestauraElAntesYNoPisaEdicionPosterior(t *testing.T) {
	rawDB, gdb := testutil.MigrarDBTemporal(t, "lotes_2c")
	_ = rawDB

	ctx := context.Background()
	var sesion, otro int64
	if err := gdb.Raw(`
		INSERT INTO usuarios (usuario, nombre, rol, password_hash)
		VALUES ('sesion.coord', 'Coordinación de prueba', 'coordinacion', 'no-es-clave')
		RETURNING id`).Row().Scan(&sesion); err != nil {
		t.Fatal(err)
	}
	if err := gdb.Raw(`
		INSERT INTO usuarios (usuario, nombre, rol, password_hash)
		VALUES ('otra.persona', 'Otra persona', 'admin', 'no-es-clave')
		RETURNING id`).Row().Scan(&otro); err != nil {
		t.Fatal(err)
	}

	var idA, idB int64
	if err := gdb.Raw(`
		INSERT INTO catalogos (clase, codigo, nombre, activo, orden)
		VALUES ('lugar', 'lote-a', 'Antes A', true, 10)
		RETURNING id`).Row().Scan(&idA); err != nil {
		t.Fatal(err)
	}
	if err := gdb.Raw(`
		INSERT INTO catalogos (clase, codigo, nombre, activo, orden)
		VALUES ('lugar', 'lote-b', 'Antes B', true, 11)
		RETURNING id`).Row().Scan(&idB); err != nil {
		t.Fatal(err)
	}

	store := postgres.NewLoteRepository(gdb)
	antesA := json.RawMessage(`{"clase":"lugar","codigo":"lote-a","nombre":"Antes A","activo":true,"orden":10,"zona":"Z1","origen":"ficticio"}`)
	despuesA := json.RawMessage(`{"clase":"lugar","codigo":"lote-a","nombre":"Lote A","activo":true,"orden":10,"zona":"Z1","origen":"ficticio"}`)
	antesB := json.RawMessage(`{"clase":"lugar","codigo":"lote-b","nombre":"Antes B","activo":true,"orden":11,"zona":"Z1","origen":"ficticio"}`)
	despuesB := json.RawMessage(`{"clase":"lugar","codigo":"lote-b","nombre":"Lote B","activo":true,"orden":11,"zona":"Z1","origen":"ficticio"}`)
	loteID, err := store.Importar(ctx, sesion, "catalogos", []entities.FilaLote{
		{EntidadID: fmt.Sprint(idA), Accion: "edicion", Antes: antesA, Despues: despuesA},
		{EntidadID: fmt.Sprint(idB), Accion: "edicion", Antes: antesB, Despues: despuesB},
	})
	if err != nil {
		t.Fatal(err)
	}
	if nombreCatalogo(t, gdb, idA) != "Lote A" || nombreCatalogo(t, gdb, idB) != "Lote B" {
		t.Fatal("la importación no aplicó las dos filas")
	}

	tl, err := store.Timeline(ctx, entities.FiltroAuditoria{Entidad: "catalogos", EntidadID: fmt.Sprint(idA)})
	if err != nil {
		t.Fatal(err)
	}
	if len(tl) != 1 || tl[0].Usuario != "sesion.coord" || tl[0].Nombre != "Coordinación de prueba" {
		t.Fatalf("timeline %#v", tl)
	}
	if tl[0].Usuario == "coordinacion" {
		t.Fatal("el timeline no debe usar el rol como actor")
	}

	if err := store.Editar(ctx, otro, "catalogos", fmt.Sprint(idA), []byte(`{"clase":"lugar","codigo":"lote-a","nombre":"Manual A","activo":true,"orden":10,"zona":"Z1","origen":"ficticio"}`)); err != nil {
		t.Fatal(err)
	}

	if _, err := store.Revertir(ctx, loteID, sesion, false); !errors.Is(err, apperrors.ErrConfirmacion) {
		t.Fatalf("se esperaba confirmación, fue %v", err)
	}
	if nombreCatalogo(t, gdb, idA) != "Manual A" || nombreCatalogo(t, gdb, idB) != "Lote B" {
		t.Fatal("la reversión sin confirmar no debe cambiar filas")
	}

	rep, err := store.Revertir(ctx, loteID, sesion, true)
	if err != nil {
		t.Fatal(err)
	}
	if nombreCatalogo(t, gdb, idA) != "Manual A" {
		t.Fatalf("la fila editada después quedó %q", nombreCatalogo(t, gdb, idA))
	}
	if nombreCatalogo(t, gdb, idB) != "Antes B" {
		t.Fatalf("la otra fila no volvió al antes: %q", nombreCatalogo(t, gdb, idB))
	}
	if len(rep.Excluidas) != 1 || rep.Excluidas[0].EntidadID != fmt.Sprint(idA) {
		t.Fatalf("reporte %+v", rep)
	}
	if len(rep.Revertidas) != 1 || rep.Revertidas[0] != fmt.Sprint(idB) {
		t.Fatalf("revertidas %+v", rep.Revertidas)
	}

	filtrado, err := store.Historial(ctx, entities.FiltroAuditoria{Zona: "Z1", Origen: "ficticio"})
	if err != nil {
		t.Fatal(err)
	}
	if len(filtrado) < 2 {
		t.Fatalf("historial filtrable = %d", len(filtrado))
	}

	op := postgres.NewIntervencionRepository(gdb)
	labor := "77777777-7777-4777-8777-777777777777"
	if _, _, err := op.Create(ctx, entities.NuevaIntervencion{
		ID: labor, Tipo: "riego", Titulo: "Riego de lote", Lon: -77.08, Lat: -12.07,
		ActorRol: "coordinacion", UsuarioID: sesion, Ejecutor: "propia",
	}); err != nil {
		t.Fatal(err)
	}
	bitacora, err := op.Timeline(ctx, labor)
	if err != nil {
		t.Fatal(err)
	}
	if len(bitacora) == 0 || bitacora[0].Usuario != "sesion.coord" {
		t.Fatalf("bitácora %#v", bitacora)
	}
}

func TestAltaInsertaYReimportacionPosteriorNoSePisa(t *testing.T) {
	_, gdb := testutil.MigrarDBTemporal(t, "lotes_alta")
	ctx := context.Background()
	var sesion int64
	if err := gdb.Raw(`
		INSERT INTO usuarios (usuario, nombre, rol, password_hash)
		VALUES ('alta.coord', 'Coordinación de alta', 'coordinacion', 'no-es-clave')
		RETURNING id`).Row().Scan(&sesion); err != nil {
		t.Fatal(err)
	}
	store := postgres.NewLoteRepository(gdb)
	nuevo := "88001"
	loteAlta, err := store.Importar(ctx, sesion, "catalogos", []entities.FilaLote{{
		EntidadID: nuevo,
		Accion:    "alta",
		Antes:     json.RawMessage("null"),
		Despues:   json.RawMessage(`{"clase":"lugar","codigo":"alta-nueva","nombre":"Fila nueva","activo":true,"orden":3}`),
	}})
	if err != nil {
		t.Fatal(err)
	}
	if nombreCatalogo(t, gdb, 88001) != "Fila nueva" {
		t.Fatal("el alta hizo UPDATE sobre una fila que no existía")
	}

	var id int64
	if err := gdb.Raw(`
		INSERT INTO catalogos (clase, codigo, nombre, activo, orden)
		VALUES ('lugar', 'reimp', 'Original', true, 4)
		RETURNING id`).Row().Scan(&id); err != nil {
		t.Fatal(err)
	}
	primero, err := store.Importar(ctx, sesion, "catalogos", []entities.FilaLote{{
		EntidadID: fmt.Sprint(id),
		Accion:    "edicion",
		Antes:     json.RawMessage(`{"clase":"lugar","codigo":"reimp","nombre":"Original","activo":true,"orden":4}`),
		Despues:   json.RawMessage(`{"clase":"lugar","codigo":"reimp","nombre":"Lote uno","activo":true,"orden":4}`),
	}})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := store.Importar(ctx, sesion, "catalogos", []entities.FilaLote{{
		EntidadID: fmt.Sprint(id),
		Accion:    "edicion",
		Antes:     json.RawMessage(`{"clase":"lugar","codigo":"reimp","nombre":"Lote uno","activo":true,"orden":4}`),
		Despues:   json.RawMessage(`{"clase":"lugar","codigo":"reimp","nombre":"Lote dos","activo":true,"orden":4}`),
	}}); err != nil {
		t.Fatal(err)
	}
	if _, err := store.Revertir(ctx, primero, sesion, false); !errors.Is(err, apperrors.ErrConfirmacion) {
		t.Fatalf("la reimportación posterior debía frenar el revert, fue %v", err)
	}
	rep, err := store.Revertir(ctx, primero, sesion, true)
	if err != nil {
		t.Fatal(err)
	}
	if nombreCatalogo(t, gdb, id) != "Lote dos" {
		t.Fatalf("el revert del primer lote pisó la reimportación: %q", nombreCatalogo(t, gdb, id))
	}
	if len(rep.Excluidas) != 1 {
		t.Fatalf("excluidas %+v", rep.Excluidas)
	}
	if loteAlta < 1 {
		t.Fatal("lote de alta")
	}
	var seq, maxID int64
	if err := gdb.Raw(`SELECT last_value FROM catalogos_id_seq`).Row().Scan(&seq); err != nil {
		t.Fatal(err)
	}
	if err := gdb.Raw(`SELECT MAX(id) FROM catalogos`).Row().Scan(&maxID); err != nil {
		t.Fatal(err)
	}
	if seq < maxID {
		t.Fatalf("la secuencia %d quedó detrás del id %d", seq, maxID)
	}
	var siguiente int64
	if err := gdb.Raw(`
		INSERT INTO catalogos (clase, codigo, nombre, activo, orden)
		VALUES ('lugar', 'seq-ok', 'Después del alta', true, 5)
		RETURNING id`).Row().Scan(&siguiente); err != nil {
		t.Fatal(err)
	}
	if siguiente <= 88001 {
		t.Fatalf("el alta normal chocó con el id explícito: %d", siguiente)
	}
}

func TestRevertirMedidaPalmeraConservaLaFila(t *testing.T) {
	_, gdb := testutil.MigrarDBTemporal(t, "medida_baja")
	ctx := context.Background()
	var sesion int64
	if err := gdb.Raw(`
		INSERT INTO usuarios (usuario, nombre, rol, password_hash)
		VALUES ('medida.coord', 'Coordinación de medida', 'coordinacion', 'no-es-clave')
		RETURNING id`).Row().Scan(&sesion); err != nil {
		t.Fatal(err)
	}
	var ejemplarID int64
	if err := gdb.Raw(`
		INSERT INTO ejemplares (nombre_comun, tipo_vegetacion, cantidad)
		VALUES ('Palmera de prueba', 'Palmera', 1)
		RETURNING id`).Row().Scan(&ejemplarID); err != nil {
		t.Fatal(err)
	}
	if err := gdb.Exec(`
		INSERT INTO medidas_palmera (ejemplar_id, altura, dap)
		VALUES ($1, 8, 0.4)`, ejemplarID).Error; err != nil {
		t.Fatal(err)
	}
	var antes int
	if err := gdb.Raw(`SELECT count(*) FROM medidas_palmera`).Row().Scan(&antes); err != nil {
		t.Fatal(err)
	}
	var loteID int64
	if err := gdb.Raw(`
		INSERT INTO lotes_importacion (entidad, estado, usuario_id, filas)
		VALUES ('medidas_palmera', 'confirmado', $1, 1)
		RETURNING id`, sesion).Row().Scan(&loteID); err != nil {
		t.Fatal(err)
	}
	if err := gdb.Exec(`
		INSERT INTO cambios (entidad, entidad_id, accion, antes, despues, usuario_id, lote_id)
		VALUES ('medidas_palmera', $1, 'importacion', 'null'::jsonb, '{"altura":8}'::jsonb, $2, $3)`,
		fmt.Sprint(ejemplarID), sesion, loteID).Error; err != nil {
		t.Fatal(err)
	}

	store := postgres.NewLoteRepository(gdb)
	rep, err := store.Revertir(ctx, loteID, sesion, true)
	if err != nil {
		t.Fatal(err)
	}
	var despues, bajas int
	if err := gdb.Raw(`SELECT count(*) FROM medidas_palmera`).Row().Scan(&despues); err != nil {
		t.Fatal(err)
	}
	if err := gdb.Raw(`SELECT count(*) FROM medidas_palmera WHERE baja_en IS NOT NULL`).Row().Scan(&bajas); err != nil {
		t.Fatal(err)
	}
	if antes != 1 || despues != antes {
		t.Fatalf("filas antes=%d despues=%d", antes, despues)
	}
	if bajas != 1 {
		t.Fatalf("la medida no quedó en baja lógica: %d", bajas)
	}
	if len(rep.Revertidas) != 1 || rep.Revertidas[0] != fmt.Sprint(ejemplarID) {
		t.Fatalf("revertidas %+v", rep.Revertidas)
	}
}

func nombreCatalogo(t *testing.T, gdb *gorm.DB, id int64) string {
	t.Helper()
	var n string
	if err := gdb.Raw(`SELECT nombre FROM catalogos WHERE id = $1`, id).Row().Scan(&n); err != nil {
		t.Fatal(err)
	}
	return n
}
