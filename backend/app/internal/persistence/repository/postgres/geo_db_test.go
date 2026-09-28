package postgres_test

import (
	"context"
	"testing"

	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/entities"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/persistence/repository/postgres"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/shared/testutil"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func TestGeoRepository_Database(t *testing.T) {
	_, gdb := testutil.MigrarDBTemporal(t, "geo_repo")
	ctx := context.Background()
	repo := postgres.NewGeoRepository(gdb)

	// Resumen on freshly migrated DB
	res, err := repo.Resumen(ctx)
	if err != nil {
		t.Fatalf("error en Resumen: %v", err)
	}
	if res.CRS != "EPSG:4326" {
		t.Fatalf("CRS esperado EPSG:4326, obtenido: %s", res.CRS)
	}

	// Capas on freshly migrated DB
	capas, err := repo.Capas(ctx)
	if err != nil {
		t.Fatalf("error en Capas: %v", err)
	}
	if capas.Cargadas == nil {
		t.Fatal("cargadas no debe ser nil")
	}

	// Areas on empty DB
	areas, err := repo.Areas(ctx, entities.FiltroGeo{})
	if err != nil {
		t.Fatalf("error en Areas: %v", err)
	}
	if areas.Type != "FeatureCollection" || areas.Name != "areas_verdes" {
		t.Fatalf("FeatureCollection de areas inválida: %+v", areas)
	}

	// Zonas on empty DB
	zonas, err := repo.Zonas(ctx, entities.FiltroGeo{})
	if err != nil {
		t.Fatalf("error en Zonas: %v", err)
	}
	if zonas.Type != "FeatureCollection" || zonas.Name != "zonas" {
		t.Fatalf("FeatureCollection de zonas inválida: %+v", zonas)
	}
}
