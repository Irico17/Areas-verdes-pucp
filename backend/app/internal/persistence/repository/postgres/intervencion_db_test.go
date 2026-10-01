package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"testing"
	"time"

	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/entities"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/persistence/mapper"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/shared/testutil"
)

type filaNula struct{}

func (filaNula) Scan(dest ...any) error {
	*dest[0].(*string) = "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaa2"
	*dest[1].(*string) = "riego"
	*dest[2].(*string) = "pendiente"
	*dest[3].(*string) = "Con lugar"
	*dest[4].(*string) = ""
	*dest[9].(*bool) = false
	*dest[10].(*time.Time) = time.Date(2026, 9, 25, 0, 0, 0, 0, time.UTC)
	*dest[11].(*time.Time) = time.Date(2026, 9, 25, 0, 0, 0, 0, time.UTC)
	*dest[12].(*string) = "propia"
	*dest[13].(*sql.NullString) = sql.NullString{}
	return nil
}

func TestScanFeatureAceptaGeomNula(t *testing.T) {
	f, err := mapper.ScanFeature(filaNula{})
	if err != nil {
		t.Fatal(err)
	}
	if string(f.Geometry) != "null" {
		t.Fatalf("geometry = %s", f.Geometry)
	}
	body, err := json.Marshal(f)
	if err != nil {
		t.Fatal(err)
	}
	if !json.Valid(body) {
		t.Fatalf("json inválido: %s", body)
	}
}

func TestListLaborSinGeom(t *testing.T) {
	_, gdb := testutil.MigrarDBTemporal(t, "labor_geom")

	if err := gdb.Exec(`
		INSERT INTO lugares (nombre, nombre_norm, lat, lon) VALUES ('Eje', 'eje', -12.07, -77.08);
		INSERT INTO actividades (id, tipo, estado, titulo, detalle, geom, lugar_id)
		VALUES (
		  'aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaa2', 'riego', 'pendiente', 'Con lugar', '', NULL,
		  (SELECT id FROM lugares WHERE nombre_norm = 'eje')
		)`).Error; err != nil {
		t.Fatal(err)
	}

	repo := NewIntervencionRepository(gdb)
	fc, err := repo.List(context.Background(), entities.FiltroIntervenciones{Rol: "coordinacion"})
	if err != nil {
		t.Fatal(err)
	}
	var found bool
	for _, f := range fc.Features {
		if f.ID != "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaa2" {
			continue
		}
		found = true
		if !json.Valid(f.Geometry) || string(f.Geometry) == "" {
			t.Fatalf("geometría = %s", f.Geometry)
		}
	}
	if !found {
		t.Fatal("la labor sin geom no salió en el listado")
	}
}

func TestIntervencionRepository_MutacionesYEventos(t *testing.T) {
	_, gdb := testutil.MigrarDBTemporal(t, "op_mut")
	repo := NewIntervencionRepository(gdb)
	ctx := context.Background()

	laborID := "98765432-1111-4111-8111-111111111111"

	if err := gdb.Exec(`
		INSERT INTO usuarios (id, usuario, nombre, rol, password_hash)
		VALUES (4, 'coordinacion', 'Coordinación', 'coordinacion', 'hash')
		ON CONFLICT (id) DO NOTHING`).Error; err != nil {
		t.Fatal(err)
	}

	// 1. Create labor
	in := entities.NuevaIntervencion{
		ID:                laborID,
		Tipo:              "riego",
		Titulo:            "Riego sector norte",
		Detalle:           "Riego manual",
		Lon:               -77.08,
		Lat:               -12.07,
		AssignedCapatazID: "cap-norte",
		ActorRol:          "coordinacion",
		Ejecutor:          "propia",
		UsuarioID:         4,
	}
	feat, created, err := repo.Create(ctx, in)
	if err != nil {
		t.Fatalf("error al crear labor: %v", err)
	}
	if !created || feat.ID != laborID {
		t.Fatalf("esperado created=true y id=%s, obtuve created=%v id=%s", laborID, created, feat.ID)
	}

	// 2. Create again (idempotent)
	feat2, created2, err := repo.Create(ctx, in)
	if err != nil {
		t.Fatalf("error reintento idempotente: %v", err)
	}
	if created2 || feat2.ID != laborID {
		t.Fatalf("esperado created=false en reintento, obtuve created=%v", created2)
	}

	// 3. Assign to cap-sur
	feat3, err := repo.Assign(ctx, entities.AsignarIntervencion{
		ID:        laborID,
		CapatazID: "cap-sur",
		ActorRol:  "coordinacion",
		UsuarioID: 4,
	})
	if err != nil {
		t.Fatalf("error al reasignar: %v", err)
	}
	if feat3.ID != laborID {
		t.Fatalf("id inesperado tras reasignar: %s", feat3.ID)
	}

	// 4. Change state to en_proceso
	feat4, err := repo.SetEstado(ctx, entities.CambiarEstadoIntervencion{
		ID:        laborID,
		Estado:    "en_proceso",
		ActorRol:  "coordinacion",
		CapatazID: "cap-sur",
		UsuarioID: 4,
	})
	if err != nil {
		t.Fatalf("error al cambiar estado: %v", err)
	}
	if feat4.ID != laborID {
		t.Fatalf("id inesperado tras cambiar estado: %s", feat4.ID)
	}

	// 5. Guardar ficha
	err = repo.GuardarFicha(ctx, entities.FichaIntervencion{
		ID:             laborID,
		Comentario:     "Comentario ficha",
		FechaSolicitud: "2026-09-20",
		FechaAtencion:  "2026-09-22",
		ActorRol:       "coordinacion",
	})
	if err != nil {
		t.Fatalf("error al guardar ficha: %v", err)
	}

	// 6. Crear avance
	avanceID := "55555555-5555-4555-8555-555555555555"
	err = repo.CrearAvance(ctx, entities.NuevoAvance{
		ActividadID:   laborID,
		ID:            avanceID,
		Fecha:         "2026-09-28",
		Nota:          "Avance 1",
		AreaFeatureID: "AV-0001",
		ActorRol:      "coordinacion",
	})
	if err != nil {
		t.Fatalf("error al crear avance: %v", err)
	}

	// 7. Timeline
	eventos, err := repo.Timeline(ctx, laborID)
	if err != nil {
		t.Fatalf("error al leer timeline: %v", err)
	}
	if len(eventos) < 4 {
		t.Fatalf("timeline inesperado: %+v", eventos)
	}
	// Verify usuario_id was populated in eventos
	for _, ev := range eventos {
		if ev.ActividadID != laborID || ev.UsuarioID == nil || *ev.UsuarioID != 4 {
			t.Fatalf("evento sin usuario_id esperado: %+v", ev)
		}
	}

	// 8. Archivar labor
	err = repo.Archive(ctx, entities.ArchivarIntervencion{
		ID:        laborID,
		ActorRol:  "coordinacion",
		Motivo:    "duplicada",
		UsuarioID: 4,
	})
	if err != nil {
		t.Fatalf("error al archivar: %v", err)
	}
}
