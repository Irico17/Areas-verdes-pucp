package postgres_test

import (
	"context"
	"errors"
	"testing"

	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/entities"
	domainErrors "github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/errors"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/persistence/repository/postgres"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/shared/testutil"
)

func TestEjemplarPatchYRecodificacion(t *testing.T) {
	_, gdb := testutil.MigrarDBTemporal(t, "ejemplar_patch")
	ctx := context.Background()
	repo := postgres.NewEjemplarRepository(gdb)

	var uid int64
	if err := gdb.Raw(`
		INSERT INTO usuarios (usuario, nombre, rol, password_hash)
		VALUES ('coord.patch', 'Coordinación de ensayo', 'coordinacion', 'no-es-clave')
		RETURNING id`).Row().Scan(&uid); err != nil {
		t.Fatalf("usuario: %v", err)
	}

	ej, err := repo.Crear(ctx, entities.Ejemplar{
		Codigo:         "FL-100",
		NombreComun:    "Tipuana de ensayo",
		TipoVegetacion: "Árbol",
		Cantidad:       1,
	})
	if err != nil {
		t.Fatalf("crear: %v", err)
	}

	var sectorID int64
	if err := gdb.Raw(`SELECT id FROM catalogos WHERE clase = 'sector_capataz' AND codigo = 'sector_norte'`).Scan(&sectorID).Error; err != nil || sectorID == 0 {
		t.Fatalf("sector de catálogo: id=%d err=%v", sectorID, err)
	}
	var lugarCat int64
	if err := gdb.Raw(`SELECT id FROM catalogos WHERE clase = 'lugar' ORDER BY id LIMIT 1`).Scan(&lugarCat).Error; err != nil || lugarCat == 0 {
		t.Fatalf("catálogo lugar: id=%d err=%v", lugarCat, err)
	}

	salud := "bueno"
	lat := -12.072
	lon := -77.082
	parche := entities.ParcheEjemplar{
		TieneSalud:      true,
		Salud:           &salud,
		TieneSector:     true,
		SectorCuartelID: &sectorID,
		TieneLat:        true,
		TieneLon:        true,
		Lat:             &lat,
		Lon:             &lon,
	}
	editado, err := repo.Actualizar(ctx, ej.ID, parche, &uid)
	if err != nil {
		t.Fatalf("patch salud: %v", err)
	}
	if editado.Salud == nil || *editado.Salud != "bueno" {
		t.Fatalf("salud = %v", editado.Salud)
	}
	if editado.SectorCuartelID == nil || *editado.SectorCuartelID != sectorID {
		t.Fatalf("sector = %v", editado.SectorCuartelID)
	}
	if editado.SectorCuartelClase != "sector_capataz" || editado.SectorCuartelNombre == "" {
		t.Fatalf("nombre de sector = %q %q", editado.SectorCuartelClase, editado.SectorCuartelNombre)
	}

	var cambios int
	if err := gdb.Raw(`SELECT count(*) FROM cambios WHERE entidad = 'ejemplares' AND entidad_id = ? AND accion = 'edicion'`, ej.ID).Scan(&cambios).Error; err != nil {
		t.Fatal(err)
	}
	if cambios != 1 {
		t.Fatalf("cambios de edición = %d", cambios)
	}

	malo := entities.ParcheEjemplar{TieneSector: true, SectorCuartelID: &lugarCat}
	if _, err := repo.Actualizar(ctx, ej.ID, malo, &uid); !errors.Is(err, domainErrors.ErrEntrada) {
		t.Fatalf("un lugar de catálogo no es sector ni cuartel: %v", err)
	}

	hist, err := repo.Recodificar(ctx, ej.ID, "FL-200", &uid)
	if err != nil {
		t.Fatalf("recodificar: %v", err)
	}
	if hist.CodigoAnterior != "FL-100" || hist.CodigoNuevo != "FL-200" {
		t.Fatalf("historial = %+v", hist)
	}
	codigos, err := repo.ListarCodigos(ctx, ej.ID)
	if err != nil || len(codigos) != 1 || codigos[0].CodigoAnterior != "FL-100" {
		t.Fatalf("códigos = %+v err=%v", codigos, err)
	}
	var codigo string
	if err := gdb.Raw(`SELECT codigo FROM ejemplares WHERE id = ?`, ej.ID).Scan(&codigo).Error; err != nil {
		t.Fatal(err)
	}
	if codigo != "FL-200" {
		t.Fatalf("código vigente = %s", codigo)
	}

	// El upsert por origen_ref no menciona la columna nueva y no la borra.
	if err := gdb.Exec(`
		INSERT INTO ejemplares (codigo, tipo_vegetacion, cantidad, origen_ref, sector_cuartel_id)
		VALUES ('FL-UPSERT', 'Árbol', 1, 'gviz:ensayo-patch', $1)
		ON CONFLICT (origen_ref) DO UPDATE SET
		  codigo = EXCLUDED.codigo,
		  tipo_vegetacion = EXCLUDED.tipo_vegetacion,
		  cantidad = EXCLUDED.cantidad,
		  updated_at = now()`, sectorID).Error; err != nil {
		t.Fatalf("alta con origen_ref: %v", err)
	}
	if err := gdb.Exec(`
		INSERT INTO ejemplares (codigo, tipo_vegetacion, cantidad, origen_ref)
		VALUES ('FL-UPSERT-2', 'Arbusto', 2, 'gviz:ensayo-patch')
		ON CONFLICT (origen_ref) DO UPDATE SET
		  codigo = EXCLUDED.codigo,
		  tipo_vegetacion = EXCLUDED.tipo_vegetacion,
		  cantidad = EXCLUDED.cantidad,
		  updated_at = now()`).Error; err != nil {
		t.Fatalf("upsert origen_ref: %v", err)
	}
	var quedo struct {
		Codigo string
		Tipo   string
		Sector *int64
	}
	if err := gdb.Raw(`
		SELECT codigo, tipo_vegetacion, sector_cuartel_id
		FROM ejemplares WHERE origen_ref = 'gviz:ensayo-patch'`).Row().Scan(&quedo.Codigo, &quedo.Tipo, &quedo.Sector); err != nil {
		t.Fatal(err)
	}
	if quedo.Codigo != "FL-UPSERT-2" || quedo.Tipo != "Arbusto" || quedo.Sector == nil || *quedo.Sector != sectorID {
		t.Fatalf("upsert pisó el sector o no actualizó la ficha: %+v", quedo)
	}
}
