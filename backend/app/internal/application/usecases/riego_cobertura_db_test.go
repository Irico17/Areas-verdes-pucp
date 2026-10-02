package usecases_test

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/usecases"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/persistence/repository/postgres"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/shared/testutil"
)

func TestCoberturaProvisionalDosSectoresUnoRegado(t *testing.T) {
	_, gdb := testutil.MigrarDBTemporal(t, "riego_cob")
	ctx := context.Background()

	if err := gdb.Exec(`UPDATE sectores_capataz SET activo = FALSE`).Error; err != nil {
		t.Fatal(err)
	}
	if err := gdb.Exec(`
		INSERT INTO sectores_capataz (codigo, nombre, color, activo) VALUES
		  ('sector-lago', 'Sector lago (ficticio)', '#3f73b0', TRUE),
		  ('sector-olivo', 'Sector olivo (ficticio)', '#6b5596', TRUE),
		  ('sector-seco', 'Sector seco (ficticio)', '#8a9a62', FALSE)
	`).Error; err != nil {
		t.Fatal(err)
	}
	if err := gdb.Exec(`
		INSERT INTO riego_registros (id, sector, turno, fecha, nota, sector_id)
		SELECT '11111111-1111-4111-8111-111111111111', s.nombre, 'manana',
		       (timezone('America/Lima', now()))::date, 'Ronda del mes', s.id
		FROM sectores_capataz s WHERE s.codigo = 'sector-lago'
	`).Error; err != nil {
		t.Fatal(err)
	}
	if err := gdb.Exec(`
		INSERT INTO riego_registros (id, sector, turno, fecha, nota, sector_id)
		SELECT '11111111-1111-4111-8111-111111111112', s.nombre, 'tarde',
		       (timezone('America/Lima', now()))::date, 'Segunda ronda del mismo sector', s.id
		FROM sectores_capataz s WHERE s.codigo = 'sector-lago'
	`).Error; err != nil {
		t.Fatal(err)
	}
	if err := gdb.Exec(`
		INSERT INTO riego_registros (id, sector, turno, fecha, nota, sector_id)
		SELECT '11111111-1111-4111-8111-111111111113', s.nombre, 'manana',
		       (date_trunc('month', timezone('America/Lima', now())) - interval '1 day')::date,
		       'Ronda del mes anterior', s.id
		FROM sectores_capataz s WHERE s.codigo = 'sector-olivo'
	`).Error; err != nil {
		t.Fatal(err)
	}
	if err := gdb.Exec(`
		INSERT INTO riego_registros (id, sector, turno, fecha, nota, sector_id)
		SELECT '11111111-1111-4111-8111-111111111114', s.nombre, 'manana',
		       (timezone('America/Lima', now()))::date, 'Sector inactivo', s.id
		FROM sectores_capataz s WHERE s.codigo = 'sector-seco'
	`).Error; err != nil {
		t.Fatal(err)
	}
	if err := gdb.Exec(`
		INSERT INTO riego_registros (id, sector, turno, fecha, nota)
		VALUES (
		  '22222222-2222-4222-8222-222222222222',
		  'Eje central',
		  'tarde',
		  (timezone('America/Lima', now()))::date,
		  'Texto anterior, sin sector de catálogo'
		)
	`).Error; err != nil {
		t.Fatal(err)
	}

	var texto string
	var sectorNulo bool
	if err := gdb.Raw(`
		SELECT sector, sector_id IS NULL
		FROM riego_registros
		WHERE id = '22222222-2222-4222-8222-222222222222'
	`).Row().Scan(&texto, &sectorNulo); err != nil {
		t.Fatal(err)
	}
	if texto != "Eje central" || !sectorNulo {
		t.Fatalf("el texto viejo debe quedar: %q nulo=%v", texto, sectorNulo)
	}

	res, err := usecases.NewRiegoUseCase(postgres.NewRiegoRepository(gdb)).Listar(ctx, "")
	if err != nil {
		t.Fatal(err)
	}
	if res.Cobertura != 50 {
		t.Fatalf("cobertura = %d, se esperaba 50", res.Cobertura)
	}
	if !res.Provisional {
		t.Fatal("la cobertura debe marcarse provisional")
	}
	raw, err := json.Marshal(res)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(raw), `"provisional":true`) {
		t.Fatalf("JSON sin provisional true: %s", raw)
	}
	if !strings.Contains(string(raw), `"cobertura":50`) {
		t.Fatalf("JSON sin cobertura 50: %s", raw)
	}
}
