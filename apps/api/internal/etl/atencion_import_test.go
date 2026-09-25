package etl

import (
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func rawSheets(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("sin ruta del test")
	}
	return filepath.Join(filepath.Dir(file), "..", "..", "..", "..", "data", "raw", "sheets")
}

func TestImportarMonitoreo283(t *testing.T) {
	tabla := NuevaTabla()
	rows, rep, err := ImportarMonitoreo(filepath.Join(rawSheets(t), "monitoreo_2026.csv"), tabla)
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != MonitoreoN || rep.Labores != MonitoreoN {
		t.Fatalf("labores %d, reporte %d", len(rows), rep.Labores)
	}
	if rep.SinEstado != 231 {
		t.Fatalf("sin_estado %d", rep.SinEstado)
	}
	if rep.ConCoordenada != 102 {
		t.Fatalf("coordenadas %d", rep.ConCoordenada)
	}
	if rep.SinResponsable != 110 {
		t.Fatalf("sin responsable %d", rep.SinResponsable)
	}
	for _, row := range rows {
		if row.Ejecutor != "propia" {
			t.Fatalf("ejecutor %s", row.Ejecutor)
		}
		if row.Estado == "" {
			t.Fatal("estado vacío")
		}
		if row.Cuadrilla != "" && row.Cuadrilla != "Valeria Quispe" && row.Cuadrilla != "Mateo Salazar" && row.Cuadrilla != "Renato Cárdenas" && row.Cuadrilla != "Nora Beltrán" && row.Cuadrilla != "Iván Paredes" && row.Cuadrilla != "Lucía Mendoza" && row.Cuadrilla != "campo depo" && row.Cuadrilla != "Bosque húme" {
			t.Fatalf("responsable no anonimizado")
		}
		if strings.Contains(row.Cuadrilla, " ") && row.Cuadrilla != "" {
			for _, frag := range []string{"@", "http"} {
				if strings.Contains(strings.ToLower(row.Cuadrilla), frag) {
					t.Fatalf("dato personal en cuadrilla")
				}
			}
		}
	}
}

func TestImportarPoda25(t *testing.T) {
	tabla := NuevaTabla()
	rows, _, err := ImportarPoda(filepath.Join(rawSheets(t), "poda.csv"), tabla, map[string]string{})
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != PodaN {
		t.Fatalf("podas %d", len(rows))
	}
	externos := 0
	for _, row := range rows {
		if row.CodigoExterno != "" {
			if !strings.HasPrefix(row.CodigoExterno, "OSG-") {
				t.Fatalf("código generado o inválido: %s", row.CodigoExterno)
			}
			externos++
		}
		if strings.Contains(strings.ToLower(row.Personal), "http") {
			t.Fatal("url en personal")
		}
	}
	if externos == 0 {
		t.Fatal("se esperaban códigos OSG presentes en la hoja")
	}
	if CodigoExterno("aun no tiene codigo") != "" || CodigoExterno("") != "" {
		t.Fatal("el código OSG no debe autogenerarse")
	}
}

func TestImportarViveroDelta(t *testing.T) {
	tabla := NuevaTabla()
	rows, rep, err := ImportarVivero(filepath.Join(rawSheets(t), "vivero.csv"), tabla, map[string]string{})
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != ViveroLocal || rep.Vivero != ViveroLocal {
		t.Fatalf("vivero %d", len(rows))
	}
	if rep.ViveroVivo != ViveroVivo {
		t.Fatalf("vivo %d", rep.ViveroVivo)
	}
	if !strings.Contains(rep.ViveroDelta, "689") || !strings.Contains(rep.ViveroDelta, "683") {
		t.Fatalf("delta sin explicar: %s", rep.ViveroDelta)
	}
	if rep.LugarLibre == 0 {
		t.Fatal("se esperaban lugares sin match")
	}
}

func TestFicticioEstable(t *testing.T) {
	a := NuevaTabla()
	b := NuevaTabla()
	nombres := []string{"Ana Ejemplo", "Ana Ejemplo", "Luis Ejemplo", "campo depo", "Bosque húme"}
	a.Aplicar(nombres)
	b.Aplicar(nombres)
	if a.Ficticio("Ana Ejemplo") != b.Ficticio("Ana Ejemplo") {
		t.Fatal("el ficticio no es estable")
	}
	if a.Ficticio("campo depo") != "campo depo" || a.Ficticio("Bosque húme") != "Bosque húme" {
		t.Fatal("se anonimizó una etiqueta")
	}
	if a.Ficticio("Ana Ejemplo") == "Ana Ejemplo" {
		t.Fatal("se conservó un nombre de persona")
	}
}
