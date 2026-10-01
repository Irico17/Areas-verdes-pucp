package postgres

import (
	"context"
	"errors"
	"testing"

	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/entities"
	domainErrors "github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/errors"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/shared/testutil"
)

func TestPodaRepository_Database(t *testing.T) {
	_, gdb := testutil.MigrarDBTemporal(t, "poda_repo")
	repo := NewPodaRepository(gdb)
	ctx := context.Background()

	podaID := "aaaaaaaa-1111-2222-3333-444444444444"
	in := entities.GuardarPoda{
		ID:                podaID,
		Codigo:            "PO-100",
		CodigoExterno:     "OSG-001",
		Tipo:              "Mantenimiento",
		TipoActividad:     "Descope",
		FechaReporte:      "2026-06-01",
		FechaEjecucion:    "2026-06-02",
		Personal:          "Cuadrilla 1",
		Ubicacion:         "Jardín Central",
		Unidad:            "árboles",
		CantidadPedida:    3,
		CantidadEjecutada: 3,
		Prioridad:         "alta",
		Comentario:        "Poda inicial",
		NombreComun:       "Ficus",
		NombreCientifico:  "Ficus benjamina",
	}

	// 1. Guardar (creación)
	guardada, err := repo.Guardar(ctx, in)
	if err != nil {
		t.Fatalf("error al guardar poda: %v", err)
	}
	if guardada.ID != podaID || guardada.Codigo != "PO-100" || guardada.CodigoExterno == nil || *guardada.CodigoExterno != "OSG-001" {
		t.Fatalf("poda guardada inesperada: %+v", guardada)
	}

	// 2. Listar
	items, err := repo.Listar(ctx)
	if err != nil {
		t.Fatalf("error al listar podas: %v", err)
	}
	if len(items) != 1 || items[0].Codigo != "PO-100" {
		t.Fatalf("lista de podas inesperada: %v", items)
	}

	// 3. Guardar (actualización ON CONFLICT)
	in.Comentario = "Poda modificada"
	in.Prioridad = "media"
	actualizada, err := repo.Guardar(ctx, in)
	if err != nil {
		t.Fatalf("error al actualizar poda: %v", err)
	}
	if actualizada.Comentario != "Poda modificada" || actualizada.Prioridad != "media" {
		t.Fatalf("poda actualizada inesperada: %+v", actualizada)
	}

	// 4. Archivar
	if err := repo.Archivar(ctx, podaID); err != nil {
		t.Fatalf("error al archivar poda: %v", err)
	}

	// 5. Listar excluye archivadas
	items, err = repo.Listar(ctx)
	if err != nil {
		t.Fatalf("error al listar podas tras archivar: %v", err)
	}
	if len(items) != 0 {
		t.Fatalf("se esperaban 0 podas activas, obtenidas %d", len(items))
	}

	// 6. Archivar nuevamente debe retornar ErrLaborNoEncontrada
	err = repo.Archivar(ctx, podaID)
	if !errors.Is(err, domainErrors.ErrLaborNoEncontrada) {
		t.Fatalf("se esperaba ErrLaborNoEncontrada al archivar por segunda vez, obtenido: %v", err)
	}

	// 7. La fila sigue existiendo en base de datos con archivada_en no nulo
	var count int64
	var archivadaEnVal *string
	err = gdb.Raw("SELECT count(*), max(archivada_en::text) FROM podas WHERE id = ?", podaID).Row().Scan(&count, &archivadaEnVal)
	if err != nil {
		t.Fatalf("error al consultar BD directa: %v", err)
	}
	if count != 1 || archivadaEnVal == nil {
		t.Fatalf("la fila de poda debió mantenerse con archivada_en seteado: count=%d, archivada_en=%v", count, archivadaEnVal)
	}
}

func TestViveroRepository_Database(t *testing.T) {
	_, gdb := testutil.MigrarDBTemporal(t, "vivero_repo")
	repo := NewViveroRepository(gdb)
	ctx := context.Background()

	viveroID := "bbbbbbbb-1111-2222-3333-444444444444"
	in := entities.GuardarVivero{
		ID:            viveroID,
		Fecha:         "2026-05-15",
		Area:          "Flora",
		Subproceso:    "Propagación",
		Etapa:         "Enraizamiento",
		Descripcion:   "Esquejes de crotón",
		Observaciones: "Bajo sombra",
		Responsables:  "Equipo Vivero",
		LugarID:       "VIV-ZONA-A",
		LugarLibre:    "Bancal 3",
	}

	// 1. Guardar (creación)
	guardada, err := repo.Guardar(ctx, in)
	if err != nil {
		t.Fatalf("error al guardar vivero: %v", err)
	}
	if guardada.ID != viveroID || guardada.Area != "Flora" || guardada.Fecha == nil || *guardada.Fecha != "2026-05-15" {
		t.Fatalf("registro vivero guardado inesperado: %+v", guardada)
	}

	// 2. Listar sin mes
	items, err := repo.Listar(ctx, "")
	if err != nil {
		t.Fatalf("error al listar vivero: %v", err)
	}
	if len(items) != 1 || items[0].ID != viveroID {
		t.Fatalf("lista de vivero inesperada: %v", items)
	}

	// 3. Listar con mes coincidente
	itemsMes, err := repo.Listar(ctx, "2026-05")
	if err != nil {
		t.Fatalf("error al listar vivero por mes: %v", err)
	}
	if len(itemsMes) != 1 {
		t.Fatalf("se esperaba 1 registro para 2026-05, obtenidos %d", len(itemsMes))
	}

	// 4. Listar con mes no coincidente
	itemsOtroMes, err := repo.Listar(ctx, "2026-06")
	if err != nil {
		t.Fatalf("error al listar vivero por otro mes: %v", err)
	}
	if len(itemsOtroMes) != 0 {
		t.Fatalf("se esperaban 0 registros para 2026-06, obtenidos %d", len(itemsOtroMes))
	}

	// 5. Guardar (actualización ON CONFLICT)
	in.Etapa = "Trasplante"
	actualizada, err := repo.Guardar(ctx, in)
	if err != nil {
		t.Fatalf("error al actualizar vivero: %v", err)
	}
	if actualizada.Etapa != "Trasplante" {
		t.Fatalf("registro vivero actualizado inesperado: %+v", actualizada)
	}

	// 6. Archivar
	if err := repo.Archivar(ctx, viveroID); err != nil {
		t.Fatalf("error al archivar vivero: %v", err)
	}

	// 7. Listar excluye archivadas
	items, err = repo.Listar(ctx, "")
	if err != nil {
		t.Fatalf("error al listar vivero tras archivar: %v", err)
	}
	if len(items) != 0 {
		t.Fatalf("se esperaban 0 registros activos, obtenidos %d", len(items))
	}

	// 8. La fila sigue existiendo en base de datos con archivada_en no nulo
	var count int64
	var archivadaEnVal *string
	err = gdb.Raw("SELECT count(*), max(archivada_en::text) FROM vivero_registros WHERE id = ?", viveroID).Row().Scan(&count, &archivadaEnVal)
	if err != nil {
		t.Fatalf("error al consultar BD directa: %v", err)
	}
	if count != 1 || archivadaEnVal == nil {
		t.Fatalf("la fila de vivero debió mantenerse con archivada_en seteado: count=%d, archivada_en=%v", count, archivadaEnVal)
	}
}
