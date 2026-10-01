package postgres_test

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/entities"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/persistence/repository/postgres"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/shared/testutil"
)

func TestReporteNoRecortaEn300(t *testing.T) {
	body, err := os.ReadFile("reporte.repository.go")
	if err != nil {
		t.Fatal(err)
	}
	src := string(body)
	inicio := strings.Index(src, "func (r *reporteRepository) ObtenerReporte")
	fin := strings.Index(src, "func ClausulasReporte")
	if inicio < 0 || fin < inicio {
		t.Fatal("no está el reporte en el repositorio")
	}
	bloque := src[inicio:fin]
	if strings.Contains(bloque, "LIMIT 300") {
		t.Fatal("la exportación no puede cortar en 300 si los conteos incluyen todo el filtro")
	}
	if !strings.Contains(bloque, "ORDER BY a.created_at DESC") {
		t.Fatal("el reporte perdió el orden")
	}
}

func TestReporteRepository_DB(t *testing.T) {
	_, gdb := testutil.MigrarDBTemporal(t, "reporte_repo")
	repo := postgres.NewReporteRepository(gdb)
	ctx := context.Background()

	// 1. Obtener reporte completo sin filtros
	rep, err := repo.ObtenerReporte(ctx, entities.FiltroReporte{})
	if err != nil {
		t.Fatalf("error al obtener reporte: %v", err)
	}
	if rep.Aviso == "" {
		t.Fatal("aviso no debe estar vacío")
	}
	if len(rep.Pendientes) != 3 {
		t.Fatalf("esperado 3 indicadores pendientes, obtenido %d", len(rep.Pendientes))
	}
	if rep.PorEstado == nil {
		t.Fatal("por_estado no debe ser nil")
	}
	if rep.Filas == nil {
		t.Fatal("filas no debe ser nil")
	}

	// 2. Validar streaming fila a fila
	filasCount := 0
	err = repo.RecorrerFilas(ctx, entities.FiltroReporte{}, func(f entities.FilaReporte) error {
		filasCount++
		if f.ID == "" {
			t.Error("fila sin ID")
		}
		return nil
	})
	if err != nil {
		t.Fatalf("error en RecorrerFilas: %v", err)
	}
	if filasCount != len(rep.Filas) {
		t.Errorf("conteos difieren: RecorrerFilas=%d, ObtenerReporte=%d", filasCount, len(rep.Filas))
	}

	// 3. Validar filtros
	err = repo.ValidarFiltro(entities.FiltroReporte{Desde: "invalido"})
	if err == nil {
		t.Error("esperado error con fecha inválida")
	}
	err = repo.ValidarFiltro(entities.FiltroReporte{Hasta: "invalido"})
	if err == nil {
		t.Error("esperado error con fecha inválida")
	}
	err = repo.ValidarFiltro(entities.FiltroReporte{Desde: "2026-01-01", Hasta: "2026-12-31"})
	if err != nil {
		t.Errorf("error inesperado con fechas válidas: %v", err)
	}
}
