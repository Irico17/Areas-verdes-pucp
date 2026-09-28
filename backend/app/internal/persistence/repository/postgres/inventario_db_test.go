package postgres_test

import (
	"context"
	"testing"

	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/persistence/repository/postgres"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/shared/testutil"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func TestInventarioRepository_Database(t *testing.T) {
	rawDB, gdb := testutil.MigrarDBTemporal(t, "inventario")
	defer rawDB.Close()

	ctx := context.Background()
	repo := postgres.NewInventarioRepository(gdb)

	// 1. Index inicialmente vacío
	idx, err := repo.Index(ctx)
	if err != nil {
		t.Fatalf("error en Index inicial: %v", err)
	}
	if len(idx.Capas) != 11 {
		t.Errorf("se esperaban 11 capas conocidas, obtenidas %d", len(idx.Capas))
	}
	if len(idx.Cargadas) != 0 {
		t.Errorf("se esperaban 0 cargadas inicialmente, obtenidas %d", len(idx.Cargadas))
	}

	// 2. Insertamos registros en inventario
	query := `
		INSERT INTO inventario (capa, feature_id, nombre, subtipo, detalle, lugar, foto, geom)
		VALUES
		('bebederos', 'BB-TEST-1', 'Bebedero 1', 'fuente', 'Detalle 1', 'Lugar 1', 'foto1.jpg', ST_SetSRID(ST_MakePoint(-77.08, -12.07), 4326)),
		('bebederos', 'BB-TEST-2', 'Bebedero 2', 'llenador', 'Detalle 2', 'Lugar 2', 'foto2.jpg', ST_SetSRID(ST_MakePoint(-77.081, -12.071), 4326)),
		('fauna', 'FAU-TEST-1', 'Ardilla', 'roedor', 'En árbol', 'Bosque', '', ST_SetSRID(ST_MakePoint(-77.082, -12.072), 4326))`
	if err := gdb.Exec(query).Error; err != nil {
		t.Fatalf("error insertando semillas en inventario: %v", err)
	}

	// 3. Verificar Index con capas cargadas
	idx, err = repo.Index(ctx)
	if err != nil {
		t.Fatalf("error en Index tras insertar: %v", err)
	}
	if len(idx.Cargadas) != 2 {
		t.Fatalf("se esperaban 2 capas cargadas, obtenidas %d", len(idx.Cargadas))
	}
	for _, c := range idx.Cargadas {
		if c.Capa == "bebederos" && c.Features != 2 {
			t.Errorf("se esperaban 2 features en bebederos, obtenidos %d", c.Features)
		}
		if c.Capa == "fauna" && c.Features != 1 {
			t.Errorf("se esperaba 1 feature en fauna, obtenido %d", c.Features)
		}
	}

	// 4. Consultar Capa conocida ("bebederos")
	fc, err := repo.Capa(ctx, "bebederos")
	if err != nil {
		t.Fatalf("error consultando capa bebederos: %v", err)
	}
	if fc.Type != "FeatureCollection" || fc.Name != "bebederos" {
		t.Errorf("metadatos de FeatureCollection inesperados: %+v", fc)
	}
	if len(fc.Features) != 2 {
		t.Fatalf("se esperaban 2 features en bebederos, obtenidos %d", len(fc.Features))
	}

	// 5. Consultar Capa desconocida
	_, errDesc := repo.Capa(ctx, "desconocida")
	if errDesc == nil {
		t.Fatal("se esperaba error consultando capa desconocida")
	}
}
