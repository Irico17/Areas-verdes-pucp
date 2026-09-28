package postgres_test

import (
	"context"
	"errors"
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
	if err := repo.Deactivate(ctx, item.ID); err != nil {
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
	err = repo.Deactivate(ctx, 99999999)
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
	if err := repo.Deactivate(ctx, item.ID); err != nil {
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
