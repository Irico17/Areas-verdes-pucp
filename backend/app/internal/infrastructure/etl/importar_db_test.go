package etl_test

import (
	"context"
	"testing"

	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/infrastructure/etl"
	postgresRepo "github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/persistence/repository/postgres"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/shared/testutil"
)

func TestConfirmarEscribeYRevertirDeshace(t *testing.T) {
	_, gdb := testutil.MigrarDBTemporal(t, "campus_verde_import_3a")

	var usuario int64
	if err := gdb.Raw(`
		INSERT INTO usuarios (usuario, nombre, rol, password_hash)
		VALUES ('import.coord', 'Coordinación', 'coordinacion', 'no-es-clave')
		RETURNING id`).Row().Scan(&usuario); err != nil {
		t.Fatal(err)
	}

	invalido := []byte("lugar,latitud,longitud\nAfuera,-1.20,-7.70\n")
	vista, err := etl.Previsualizar("lugares", "lugares.csv", invalido)
	if err != nil {
		t.Fatal(err)
	}
	if vista.Validas != 0 {
		t.Fatalf("el CSV inválido no debía traer filas: %+v", vista)
	}
	var idMalo int64
	if err := gdb.Raw(`
		INSERT INTO lotes_importacion (entidad, estado, usuario_id, filas, contenido, nombre_archivo)
		VALUES ('lugares', 'vista_previa', $1, 0, $2, 'lugares.csv')
		RETURNING id`, usuario, invalido).Row().Scan(&idMalo); err != nil {
		t.Fatal(err)
	}
	if _, _, err := etl.Confirmar(gdb, idMalo, usuario); err == nil {
		t.Fatal("confirmar un CSV inválido no debía escribir")
	}
	var n int
	if err := gdb.Raw(`SELECT count(*) FROM lugares`).Scan(&n).Error; err != nil {
		t.Fatal(err)
	}
	if n != 0 {
		t.Fatalf("lugares = %d", n)
	}

	valido := []byte("lugar,latitud,longitud\nBiblioteca,-12.0704,-77.0808\n")
	var id int64
	if err := gdb.Raw(`
		INSERT INTO lotes_importacion (entidad, estado, usuario_id, filas, contenido, nombre_archivo)
		VALUES ('lugares', 'vista_previa', $1, 1, $2, 'lugares.csv')
		RETURNING id`, usuario, valido).Row().Scan(&id); err != nil {
		t.Fatal(err)
	}
	if _, _, err := etl.Confirmar(gdb, id, usuario); err != nil {
		t.Fatal(err)
	}
	if err := gdb.Raw(`SELECT count(*) FROM lugares WHERE activo`).Scan(&n).Error; err != nil {
		t.Fatal(err)
	}
	if n != 1 {
		t.Fatalf("tras confirmar, lugares activos = %d", n)
	}
	rep, err := postgresRepo.NewLoteRepository(gdb).Revertir(context.Background(), id, usuario, false)
	if err != nil {
		t.Fatal(err)
	}
	if len(rep.Revertidas) != 1 {
		t.Fatalf("revertidas %+v", rep)
	}
	if err := gdb.Raw(`SELECT count(*) FROM lugares WHERE activo`).Scan(&n).Error; err != nil {
		t.Fatal(err)
	}
	if n != 0 {
		t.Fatalf("tras revertir, lugares activos = %d", n)
	}
}
