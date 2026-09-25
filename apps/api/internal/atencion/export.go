package atencion

import (
	"encoding/xml"
	"io"
	"strings"
)

// CSV arma la exportación básica, con BOM para Excel en español.
func CSV(filas []Fila) string {
	var b strings.Builder
	if err := EscribirCSV(&b, filas); err != nil {
		return ""
	}
	return b.String()
}

// EscribirCSV escribe la cabecera y cada fila en w, sin juntar el archivo antes.
func EscribirCSV(w io.Writer, filas []Fila) error {
	if err := abrirCSV(w); err != nil {
		return err
	}
	for _, f := range filas {
		if err := escribirFilaCSV(w, f); err != nil {
			return err
		}
	}
	return nil
}

func abrirCSV(w io.Writer) error {
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

func escribirFilaCSV(w io.Writer, f Fila) error {
	var b strings.Builder
	for i, v := range valoresFila(f) {
		if i > 0 {
			b.WriteByte(',')
		}
		b.WriteString(csv(v))
	}
	b.WriteByte('\n')
	_, err := io.WriteString(w, b.String())
	return err
}

func csv(v string) string {
	if strings.ContainsAny(v, ",\"\n") {
		return `"` + strings.ReplaceAll(v, `"`, `""`) + `"`
	}
	return v
}

// ExcelXML es una hoja SpreadsheetML que Excel abre sin una librería binaria.
func ExcelXML(filas []Fila) string {
	var b strings.Builder
	if err := EscribirExcelXML(&b, filas); err != nil {
		return ""
	}
	return b.String()
}

// EscribirExcelXML abre la hoja, escribe cada fila y cierra el libro.
func EscribirExcelXML(w io.Writer, filas []Fila) error {
	if err := abrirExcel(w); err != nil {
		return err
	}
	for _, f := range filas {
		if err := escribirFilaExcel(w, f); err != nil {
			return err
		}
	}
	return cerrarExcel(w)
}

func abrirExcel(w io.Writer) error {
	_, err := io.WriteString(w, `<?xml version="1.0" encoding="UTF-8"?>`+"\n"+
		`<?mso-application progid="Excel.Sheet"?>`+"\n"+
		`<Workbook xmlns="urn:schemas-microsoft-com:office:spreadsheet"><Worksheet ss:Name="Labores" xmlns:ss="urn:schemas-microsoft-com:office:spreadsheet"><Table>`+
		`<Row><Cell><Data ss:Type="String">VerdePUCP — Gestión de Áreas Verdes</Data></Cell></Row>`)
	if err != nil {
		return err
	}
	if _, err := io.WriteString(w, "<Row>"); err != nil {
		return err
	}
	for _, h := range ColumnasBasicas() {
		if _, err := io.WriteString(w, `<Cell><Data ss:Type="String">`+xmlEscape(h)+`</Data></Cell>`); err != nil {
			return err
		}
	}
	_, err = io.WriteString(w, "</Row>")
	return err
}

func escribirFilaExcel(w io.Writer, f Fila) error {
	var b strings.Builder
	b.WriteString("<Row>")
	for _, v := range valoresFila(f) {
		b.WriteString(`<Cell><Data ss:Type="String">` + xmlEscape(v) + `</Data></Cell>`)
	}
	b.WriteString("</Row>")
	_, err := io.WriteString(w, b.String())
	return err
}

func cerrarExcel(w io.Writer) error {
	_, err := io.WriteString(w, "</Table></Worksheet></Workbook>")
	return err
}

// ColumnasBasicas son las del reporte básico ya definidas. No incluye horas,
// superficie, cobertura ni el nombre real de una persona.
func ColumnasBasicas() []string {
	return []string{
		"id", "titulo", "tipo", "estado", "ejecutor", "equipo", "zona",
		"codigo_externo", "fuente", "creada", "clase", "lugar", "cuadrilla",
		"fecha_solicitud", "fecha_atencion",
	}
}

func valoresFila(f Fila) []string {
	return []string{
		f.ID, f.Titulo, f.Tipo, f.Estado, f.Ejecutor, f.Equipo, f.Zona,
		f.CodigoExterno, f.Fuente, f.CreatedAt, f.Clase, f.Lugar, f.Cuadrilla,
		f.FechaSolicitud, f.FechaAtencion,
	}
}

func xmlEscape(s string) string {
	var b strings.Builder
	_ = xml.EscapeText(&b, []byte(s))
	return b.String()
}
