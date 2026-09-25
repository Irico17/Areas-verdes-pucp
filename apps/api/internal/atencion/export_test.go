package atencion

import (
	"strings"
	"testing"
)

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
