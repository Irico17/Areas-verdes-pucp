// Package exportacion implements report exporters for CSV and Excel XML formats.
package exportacion

import (
	"io"
	"strings"

	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/entities"
)

// CSV generates the complete CSV string with UTF-8 BOM.
func CSV(filas []entities.FilaReporte) string {
	var b strings.Builder
	if err := EscribirCSV(&b, filas); err != nil {
		return ""
	}
	return b.String()
}

// EscribirCSV writes header and rows directly to w.
func EscribirCSV(w io.Writer, filas []entities.FilaReporte) error {
	if err := AbrirCSV(w); err != nil {
		return err
	}
	for _, f := range filas {
		if err := EscribirFilaCSV(w, f); err != nil {
			return err
		}
	}
	return nil
}

// AbrirCSV writes the UTF-8 BOM, title banner, and column header row.
func AbrirCSV(w io.Writer) error {
	if _, err := io.WriteString(w, "\uFEFF"); err != nil {
		return err
	}
	if _, err := io.WriteString(w, "VerdePUCP — Gestión de Áreas Verdes\n"); err != nil {
		return err
	}
	if _, err := io.WriteString(w, strings.Join(ColumnasBasicas(), ",")); err != nil {
		return err
	}
	_, err := io.WriteString(w, "\n")
	return err
}

// EscribirFilaCSV writes a single report row in CSV format.
func EscribirFilaCSV(w io.Writer, f entities.FilaReporte) error {
	var b strings.Builder
	for i, v := range ValoresFila(f) {
		if i > 0 {
			b.WriteByte(',')
		}
		b.WriteString(csvEscape(v))
	}
	b.WriteByte('\n')
	_, err := io.WriteString(w, b.String())
	return err
}

func csvEscape(v string) string {
	if strings.ContainsAny(v, ",\"\n") {
		return `"` + strings.ReplaceAll(v, `"`, `""`) + `"`
	}
	return v
}

// ColumnasBasicas returns the standard column headers for basic reports.
func ColumnasBasicas() []string {
	return []string{
		"id", "titulo", "tipo", "estado", "ejecutor", "equipo", "zona",
		"codigo_externo", "fuente", "creada", "clase", "lugar", "cuadrilla",
		"fecha_solicitud", "fecha_atencion",
	}
}

// ValoresFila extracts row values in the column order defined by ColumnasBasicas.
func ValoresFila(f entities.FilaReporte) []string {
	return []string{
		f.ID, f.Titulo, f.Tipo, f.Estado, f.Ejecutor, f.Equipo, f.Zona,
		f.CodigoExterno, f.Fuente, f.CreatedAt, f.Clase, f.Lugar, f.Cuadrilla,
		f.FechaSolicitud, f.FechaAtencion,
	}
}
