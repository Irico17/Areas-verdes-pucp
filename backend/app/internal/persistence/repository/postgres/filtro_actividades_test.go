package postgres

import (
	"context"
	"strings"
	"testing"

	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/entities"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/shared/testutil"
)

func TestClausulasCombinanRiesgoFechasYDescartanCuadrillaAjena(t *testing.T) {
	consulta, args := clausulasActividades(entities.FiltroIntervenciones{
		Rol:          "capataz",
		CapatazID:    "cap-norte",
		CuadrillaID:  "cap-sur",
		NivelRiesgo:  "alto",
		Desde:        "2026-01-01",
		Hasta:        "2026-12-31",
		SoloAbiertas: true,
	})
	for _, arg := range args {
		if arg == "cap-sur" {
			t.Fatalf("la cuadrilla ajena entró en la consulta: %v", args)
		}
	}
	for _, parte := range []string{"nivel_riesgo", "fecha_programada >=", "fecha_programada <=", "cerrada"} {
		if !strings.Contains(consulta, parte) {
			t.Fatalf("falta %q en %s", parte, consulta)
		}
	}
	for _, esperado := range []string{"cap-norte", "alto", "2026-01-01", "2026-12-31"} {
		if !tieneArg(args, esperado) {
			t.Fatalf("falta %q en %v", esperado, args)
		}
	}
}

func TestListRiesgoFechasNoDevuelveOtraCuadrillaNiCerradas(t *testing.T) {
	_, gdb := testutil.MigrarDBTemporal(t, "filtro_act")
	repo := NewIntervencionRepository(gdb)
	ctx := context.Background()

	if err := gdb.Exec(`
		INSERT INTO actividades (
		  id, tipo, estado, titulo, detalle, assigned_capataz_id, cuadrilla_id,
		  geom, nivel_riesgo, fecha_programada, origen, ejecutor
		) VALUES
		  ('a1111111-1111-4111-8111-111111111111', 'riego', 'pendiente', 'Riego alto norte', '',
		   'cap-norte', 'cap-norte', ST_SetSRID(ST_MakePoint(-77.080, -12.070), 4326), 'alto', DATE '2026-03-01', 'interna', 'propia'),
		  ('a2222222-2222-4222-8222-222222222222', 'riego', 'pendiente', 'Riego alto sur', '',
		   'cap-sur', 'cap-sur', ST_SetSRID(ST_MakePoint(-77.081, -12.071), 4326), 'alto', DATE '2026-03-02', 'interna', 'propia'),
		  ('a3333333-3333-4333-8333-333333333333', 'riego', 'pendiente', 'Riego bajo norte', '',
		   'cap-norte', 'cap-norte', ST_SetSRID(ST_MakePoint(-77.079, -12.069), 4326), 'bajo', DATE '2026-03-03', 'interna', 'propia'),
		  ('a4444444-4444-4444-8444-444444444444', 'riego', 'pendiente', 'Riego alto viejo', '',
		   'cap-norte', 'cap-norte', ST_SetSRID(ST_MakePoint(-77.0795, -12.0695), 4326), 'alto', DATE '2025-06-01', 'interna', 'propia'),
		  ('a5555555-5555-4555-8555-555555555555', 'riego', 'cerrada', 'Riego alto cerrado', '',
		   'cap-norte', 'cap-norte', ST_SetSRID(ST_MakePoint(-77.0802, -12.0702), 4326), 'alto', DATE '2026-04-01', 'interna', 'propia'),
		  ('a6666666-6666-4666-8666-666666666666', 'poda', 'pendiente', 'Poda en cuadrilla ajena', '',
		   'cap-norte', 'cap-sur', ST_SetSRID(ST_MakePoint(-77.0804, -12.0704), 4326), 'alto', DATE '2026-05-01', 'interna', 'propia')
	`).Error; err != nil {
		t.Fatal(err)
	}

	abiertas, err := repo.List(ctx, entities.FiltroIntervenciones{
		Rol:          "capataz",
		CapatazID:    "cap-norte",
		CuadrillaID:  "cap-sur",
		NivelRiesgo:  "alto",
		Desde:        "2026-01-01",
		Hasta:        "2026-12-31",
		SoloAbiertas: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	ids := idsDe(abiertas)
	if len(ids) != 1 || ids[0] != "a1111111-1111-4111-8111-111111111111" {
		t.Fatalf("el capataz recibió %v", ids)
	}

	historico, err := repo.List(ctx, entities.FiltroIntervenciones{
		Rol:          "coordinacion",
		NivelRiesgo:  "alto",
		Desde:        "2026-01-01",
		Hasta:        "2026-12-31",
		SoloAbiertas: false,
	})
	if err != nil {
		t.Fatal(err)
	}
	conCerrada := idsDe(historico)
	if !contieneID(conCerrada, "a5555555-5555-4555-8555-555555555555") {
		t.Fatalf("el histórico no trajo la cerrada: %v", conCerrada)
	}
	if contieneID(conCerrada, "a3333333-3333-4333-8333-333333333333") || contieneID(conCerrada, "a4444444-4444-4444-8444-444444444444") {
		t.Fatalf("entró una fila fuera de riesgo o de fecha: %v", conCerrada)
	}

	oficina, err := repo.List(ctx, entities.FiltroIntervenciones{
		Rol:          "coordinacion",
		NivelRiesgo:  "alto",
		Desde:        "2026-01-01",
		SoloAbiertas: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if contieneID(idsDe(oficina), "a5555555-5555-4555-8555-555555555555") {
		t.Fatal("abiertas=1 incluyó una cerrada")
	}
}

func TestClausulasFiltranEjemplarYResponsable(t *testing.T) {
	consulta, args := clausulasActividades(entities.FiltroIntervenciones{
		EjemplarID:  "42",
		Responsable: "Lucía Mendoza",
	})
	for _, parte := range []string{"ejemplar_ref", "nombre_ficticio", "cuadrilla_id", "assigned_capataz_id"} {
		if !strings.Contains(consulta, parte) {
			t.Fatalf("falta %q en %s", parte, consulta)
		}
	}
	if !tieneArg(args, "42") || !tieneArg(args, "Lucía Mendoza") {
		t.Fatalf("argumentos %v", args)
	}
}

func TestListCerradasPorEjemplarYResponsable(t *testing.T) {
	_, gdb := testutil.MigrarDBTemporal(t, "filtro_hist")
	repo := NewIntervencionRepository(gdb)
	ctx := context.Background()
	if err := gdb.Exec(`
		INSERT INTO actividades (
		  id, tipo, estado, titulo, geom, origen, ejecutor, fecha_programada
		) VALUES
		  ('b1111111-1111-4111-8111-111111111111', 'riego', 'cerrada', 'Cerrada del ejemplar',
		   ST_SetSRID(ST_MakePoint(-77.080, -12.070), 4326), 'interna', 'propia', DATE '2026-04-01'),
		  ('b2222222-2222-4222-8222-222222222222', 'riego', 'pendiente', 'Abierta del mismo ejemplar',
		   ST_SetSRID(ST_MakePoint(-77.081, -12.071), 4326), 'interna', 'propia', DATE '2026-04-02'),
		  ('b3333333-3333-4333-8333-333333333333', 'riego', 'cerrada', 'Cerrada de otro ejemplar',
		   ST_SetSRID(ST_MakePoint(-77.082, -12.072), 4326), 'externa', 'propia', DATE '2026-04-03');
		INSERT INTO actividad_avances (id, actividad_id, fecha, ejemplar_ref) VALUES
		  ('c1111111-1111-4111-8111-111111111111', 'b1111111-1111-4111-8111-111111111111', DATE '2026-04-01', '42'),
		  ('c2222222-2222-4222-8222-222222222222', 'b2222222-2222-4222-8222-222222222222', DATE '2026-04-02', '42'),
		  ('c3333333-3333-4333-8333-333333333333', 'b3333333-3333-4333-8333-333333333333', DATE '2026-04-03', '99');
		INSERT INTO personal_labor (id, actividad_id, nombre_ficticio) VALUES
		  ('d1111111-1111-4111-8111-111111111111', 'b1111111-1111-4111-8111-111111111111', 'Lucía Mendoza'),
		  ('d2222222-2222-4222-8222-222222222222', 'b2222222-2222-4222-8222-222222222222', 'Lucía Mendoza'),
		  ('d3333333-3333-4333-8333-333333333333', 'b3333333-3333-4333-8333-333333333333', 'Mateo Salazar')
	`).Error; err != nil {
		t.Fatal(err)
	}

	historico, err := repo.List(ctx, entities.FiltroIntervenciones{
		Rol:          "jefatura",
		Estado:       "cerrada",
		Origen:       "interna",
		Desde:        "2026-01-01",
		Hasta:        "2026-12-31",
		EjemplarID:   "42",
		Responsable:  "Lucía Mendoza",
		SoloAbiertas: false,
	})
	if err != nil {
		t.Fatal(err)
	}
	ids := idsDe(historico)
	if len(ids) != 1 || ids[0] != "b1111111-1111-4111-8111-111111111111" {
		t.Fatalf("el histórico cerrado devolvió %v", ids)
	}

	abiertas, err := repo.List(ctx, entities.FiltroIntervenciones{
		Rol:          "jefatura",
		EjemplarID:   "42",
		Responsable:  "Lucía Mendoza",
		SoloAbiertas: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if contieneID(idsDe(abiertas), "b1111111-1111-4111-8111-111111111111") {
		t.Fatal("abiertas=1 incluyó la cerrada del ejemplar")
	}
}

func idsDe(fc entities.FeatureCollection) []string {
	out := make([]string, 0, len(fc.Features))
	for _, f := range fc.Features {
		out = append(out, f.ID)
	}
	return out
}

func contieneID(ids []string, id string) bool {
	for _, item := range ids {
		if item == id {
			return true
		}
	}
	return false
}

func tieneArg(args []any, want string) bool {
	for _, arg := range args {
		if arg == want {
			return true
		}
	}
	return false
}

func TestSoloAbiertasExcluyeCanceladaCerradaYArchivada(t *testing.T) {
	consulta, _ := clausulasActividades(entities.FiltroIntervenciones{SoloAbiertas: true})
	for _, estado := range []string{"'cancelada'", "'cerrada'", "'archivada'"} {
		if !strings.Contains(consulta, estado) {
			t.Fatalf("abiertas no excluye %s: %s", estado, consulta)
		}
	}
	historico, _ := clausulasActividades(entities.FiltroIntervenciones{SoloAbiertas: false})
	if strings.Contains(historico, "'archivada'") {
		t.Fatalf("el histórico no debe excluir por estado: %s", historico)
	}
}
