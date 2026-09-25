package atencion

import (
	"os"
	"strings"
	"testing"
)

func TestReporteNoRecortaEn300(t *testing.T) {
	body, err := os.ReadFile("store.go")
	if err != nil {
		t.Fatal(err)
	}
	src := string(body)
	inicio := strings.Index(src, "func (s *Store) Reporte")
	fin := strings.Index(src, "func clausulasReporte")
	if inicio < 0 || fin < inicio {
		t.Fatal("no está el reporte")
	}
	bloque := src[inicio:fin]
	if strings.Contains(bloque, "LIMIT 300") {
		t.Fatal("la exportación no puede cortar en 300 si los conteos incluyen todo el filtro")
	}
	if !strings.Contains(bloque, "ORDER BY a.created_at DESC") {
		t.Fatal("el reporte perdió el orden")
	}
}

func TestExportacionEscribeCadaFila(t *testing.T) {
	filas := []Fila{
		{ID: "1", Titulo: "Poda", Estado: "cerrada"},
		{ID: "2", Titulo: "Riego", Estado: "pendiente"},
	}
	var csvBuf, xlsBuf strings.Builder
	if err := EscribirCSV(&csvBuf, filas); err != nil {
		t.Fatal(err)
	}
	if err := EscribirExcelXML(&xlsBuf, filas); err != nil {
		t.Fatal(err)
	}
	if csvBuf.String() != CSV(filas) || xlsBuf.String() != ExcelXML(filas) {
		t.Fatal("el escritor por filas no coincide con el archivo completo")
	}
	if strings.Count(csvBuf.String(), "\n") < 4 {
		t.Fatal("el CSV no escribió las dos filas")
	}
	handler, err := os.ReadFile("../handlers/producto.go")
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(handler), "CSV(rep.Filas)") || strings.Contains(string(handler), "ExcelXML(rep.Filas)") {
		t.Fatal("el handler sigue armando el archivo en memoria")
	}
}

func TestExportacionBasica(t *testing.T) {
	filas := []Fila{{
		ID: "1", Titulo: "Poda", Tipo: "poda", Estado: "sin_estado", Ejecutor: "propia",
		Zona: "Z1", Clase: "Poda", Lugar: "Tinkuy", Cuadrilla: "Nora Beltrán",
		FechaSolicitud: "2026-05-01", FechaAtencion: "2026-05-03", CreatedAt: "2026-05-01T12:00:00Z",
	}}
	csv := CSV(filas)
	xls := ExcelXML(filas)
	for _, col := range []string{"clase", "lugar", "cuadrilla", "fecha_solicitud", "fecha_atencion", "Nora Beltrán", "Tinkuy"} {
		if !strings.Contains(csv, col) || !strings.Contains(xls, col) {
			t.Fatalf("falta %q en csv o spreadsheetml", col)
		}
	}
	for _, ausente := range []string{"horas", "superficie", "cobertura", "rendimiento", "responsable"} {
		cabecera := strings.Split(strings.TrimPrefix(csv, "\uFEFF"), "\n")[1]
		if strings.Contains(cabecera, ausente) {
			t.Fatalf("columna no definida en la cabecera: %s", ausente)
		}
	}
	if strings.Contains(csv, "application/pdf") || strings.Contains(xls, ".pdf") {
		t.Fatal("el básico no exporta PDF")
	}
}

func TestFiltroFechasYZona(t *testing.T) {
	where, args, err := clausulasReporte(FiltroReporte{
		Zona: "Z2", Cuadrilla: "cua-nora", Origen: "monitoreo",
		Desde: "2026-05-01", Hasta: "2026-05-31", Estado: "cerrada",
	}, true)
	if err != nil {
		t.Fatal(err)
	}
	sql := strings.Join(where, " ")
	for _, parte := range []string{"z.codigo", "cuadrilla_id", "nombre_ficticio", "a.origen", "fecha_solicitud", "fecha_atencion", "a.estado"} {
		if !strings.Contains(sql, parte) {
			t.Fatalf("falta %q en %s", parte, sql)
		}
	}
	if len(args) != 6 {
		t.Fatalf("args = %d, quiero 6", len(args))
	}
	_, _, err = clausulasReporte(FiltroReporte{Desde: "05/01/2026"}, false)
	if err == nil {
		t.Fatal("desde inválido debe fallar")
	}
}

// TestHastaUsaElMismoIntervaloQueDesde exige el COALESCE completo de hasta.
// El bug anterior filtraba solo fecha_solicitud y created_at, e ignoraba fecha_atencion.
func TestHastaUsaElMismoIntervaloQueDesde(t *testing.T) {
	const intervalo = "COALESCE(a.fecha_atencion, a.fecha_solicitud, a.created_at::date)"

	where, args, err := clausulasReporte(FiltroReporte{Hasta: "2026-05-31"}, false)
	if err != nil {
		t.Fatal(err)
	}
	sql := strings.Join(where, " AND ")
	wantHasta := intervalo + " <= $1::date"
	if !strings.Contains(sql, wantHasta) {
		t.Fatalf("hasta = %q, quiero %q", sql, wantHasta)
	}
	if strings.Contains(sql, "COALESCE(a.fecha_solicitud, a.created_at::date)") {
		t.Fatal("hasta ignora fecha_atencion")
	}
	if !strings.Contains(sql, "a.archivada_en IS NULL") {
		t.Fatalf("el filtro compartido no excluye archivadas: %s", sql)
	}
	if len(args) != 1 || args[0] != "2026-05-31" {
		t.Fatalf("args = %#v", args)
	}

	where, _, err = clausulasReporte(FiltroReporte{Desde: "2026-05-01", Hasta: "2026-05-31"}, false)
	if err != nil {
		t.Fatal(err)
	}
	sql = strings.Join(where, " AND ")
	if !strings.Contains(sql, intervalo+" >= $1::date") || !strings.Contains(sql, intervalo+" <= $2::date") {
		t.Fatalf("desde y hasta no comparten el intervalo: %s", sql)
	}
}
