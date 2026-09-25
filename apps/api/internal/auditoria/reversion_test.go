package auditoria

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"campusverde/api/internal/db"
	"campusverde/api/internal/migrate"
	"campusverde/api/internal/operacion"

	_ "github.com/jackc/pgx/v5/stdlib"
	"gorm.io/gorm"
)

func TestRevertirLoteRestauraElAntesYNoPisaEdicionPosterior(t *testing.T) {
	base := os.Getenv("MIGRATE_TEST_URL")
	if base == "" {
		base = "postgres://campus:campus@127.0.0.1:5432/postgres?sslmode=disable"
	}
	admin, err := sql.Open("pgx", base)
	if err != nil {
		t.Fatal(err)
	}
	defer admin.Close()
	if err := admin.Ping(); err != nil {
		t.Skipf("sin postgres de prueba: %v", err)
	}

	name := "campus_verde_lotes_2c"
	if _, err := admin.Exec(`SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE datname = $1 AND pid <> pg_backend_pid()`, name); err != nil {
		t.Fatal(err)
	}
	if _, err := admin.Exec("DROP DATABASE IF EXISTS " + name); err != nil {
		t.Fatal(err)
	}
	if _, err := admin.Exec("CREATE DATABASE " + name); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_, _ = admin.Exec(`SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE datname = $1 AND pid <> pg_backend_pid()`, name)
		_, _ = admin.Exec("DROP DATABASE IF EXISTS " + name)
	})

	gdb, err := db.Open(fmt.Sprintf("postgres://campus:campus@127.0.0.1:5432/%s?sslmode=disable", name))
	if err != nil {
		t.Fatal(err)
	}
	sqlDB, err := gdb.DB()
	if err != nil {
		t.Fatal(err)
	}
	defer sqlDB.Close()
	dir := filepath.Join("..", "..", "migrations")
	if err := migrate.Apply(gdb, dir); err != nil {
		t.Fatal(err)
	}

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

	store := NewStore(gdb)
	antesA := json.RawMessage(`{"clase":"lugar","codigo":"lote-a","nombre":"Antes A","activo":true,"orden":10,"zona":"Z1","origen":"ficticio"}`)
	despuesA := json.RawMessage(`{"clase":"lugar","codigo":"lote-a","nombre":"Lote A","activo":true,"orden":10,"zona":"Z1","origen":"ficticio"}`)
	antesB := json.RawMessage(`{"clase":"lugar","codigo":"lote-b","nombre":"Antes B","activo":true,"orden":11,"zona":"Z1","origen":"ficticio"}`)
	despuesB := json.RawMessage(`{"clase":"lugar","codigo":"lote-b","nombre":"Lote B","activo":true,"orden":11,"zona":"Z1","origen":"ficticio"}`)
	loteID, err := store.Importar(ctx, sesion, "catalogos", []Fila{
		{EntidadID: fmt.Sprint(idA), Accion: "edicion", Antes: antesA, Despues: despuesA},
		{EntidadID: fmt.Sprint(idB), Accion: "edicion", Antes: antesB, Despues: despuesB},
	})
	if err != nil {
		t.Fatal(err)
	}
	if nombre(t, gdb, idA) != "Lote A" || nombre(t, gdb, idB) != "Lote B" {
		t.Fatal("la importación no aplicó las dos filas")
	}

	tl, err := store.Timeline(ctx, Filtro{Entidad: "catalogos", EntidadID: fmt.Sprint(idA)})
	if err != nil {
		t.Fatal(err)
	}
	if len(tl) != 1 || tl[0].Usuario != "sesion.coord" || tl[0].Nombre != "Coordinación de prueba" {
		t.Fatalf("timeline %#v", tl)
	}
	if tl[0].Usuario == "coordinacion" {
		t.Fatal("el timeline no debe usar el rol como actor")
	}

	if err := store.Editar(ctx, otro, "catalogos", fmt.Sprint(idA), json.RawMessage(`{"clase":"lugar","codigo":"lote-a","nombre":"Manual A","activo":true,"orden":10,"zona":"Z1","origen":"ficticio"}`)); err != nil {
		t.Fatal(err)
	}

	if _, err := store.Revertir(ctx, loteID, sesion, false); !errors.Is(err, ErrConfirmacion) {
		t.Fatalf("se esperaba confirmación, fue %v", err)
	}
	if nombre(t, gdb, idA) != "Manual A" || nombre(t, gdb, idB) != "Lote B" {
		t.Fatal("la reversión sin confirmar no debe cambiar filas")
	}

	rep, err := store.Revertir(ctx, loteID, sesion, true)
	if err != nil {
		t.Fatal(err)
	}
	if nombre(t, gdb, idA) != "Manual A" {
		t.Fatalf("la fila editada después quedó %q", nombre(t, gdb, idA))
	}
	if nombre(t, gdb, idB) != "Antes B" {
		t.Fatalf("la otra fila no volvió al antes: %q", nombre(t, gdb, idB))
	}
	if len(rep.Excluidas) != 1 || rep.Excluidas[0].EntidadID != fmt.Sprint(idA) {
		t.Fatalf("reporte %+v", rep)
	}
	if len(rep.Revertidas) != 1 || rep.Revertidas[0] != fmt.Sprint(idB) {
		t.Fatalf("revertidas %+v", rep.Revertidas)
	}

	filtrado, err := store.Historial(ctx, Filtro{Zona: "Z1", Origen: "ficticio"})
	if err != nil {
		t.Fatal(err)
	}
	if len(filtrado) < 2 {
		t.Fatalf("historial filtrable = %d", len(filtrado))
	}

	op := operacion.NewStore(gdb)
	labor := "77777777-7777-4777-8777-777777777777"
	if _, _, err := op.Create(ctx, operacion.CreateInput{
		ID: labor, Tipo: "riego", Titulo: "Riego de lote", Lon: -77.08, Lat: -12.07,
		ActorRol: "coordinacion", UsuarioID: sesion,
	}); err != nil {
		t.Fatal(err)
	}
	bitacora, err := op.Timeline(ctx, labor)
	if err != nil {
		t.Fatal(err)
	}
	if len(bitacora.Eventos) == 0 || bitacora.Eventos[0].Usuario != "sesion.coord" {
		t.Fatalf("bitácora %#v", bitacora.Eventos)
	}
}

func nombre(t *testing.T, gdb *gorm.DB, id int64) string {
	t.Helper()
	var n string
	if err := gdb.Raw(`SELECT nombre FROM catalogos WHERE id = $1`, id).Row().Scan(&n); err != nil {
		t.Fatal(err)
	}
	return n
}
