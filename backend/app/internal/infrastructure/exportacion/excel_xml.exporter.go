// Package exportacion implements report exporters for CSV and Excel XML formats.
package exportacion

import (
	"encoding/xml"
	"io"
	"strings"

	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/entities"
)

// ExcelXML returns the complete SpreadsheetML workbook as a string.
func ExcelXML(filas []entities.FilaReporte) string {
	var b strings.Builder
	if err := EscribirExcelXML(&b, filas); err != nil {
		return ""
	}
	return b.String()
}

// EscribirExcelXML writes the SpreadsheetML workbook directly to w.
func EscribirExcelXML(w io.Writer, filas []entities.FilaReporte) error {
	if err := AbrirExcel(w); err != nil {
		return err
	}
	for _, f := range filas {
		if err := EscribirFilaExcel(w, f); err != nil {
			return err
		}
	}
	return CerrarExcel(w)
}

// AbrirExcel writes the SpreadsheetML XML declaration, workbook, worksheet, and header row.
func AbrirExcel(w io.Writer) error {
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

// EscribirFilaExcel writes a single report row as an XML SpreadsheetML row.
func EscribirFilaExcel(w io.Writer, f entities.FilaReporte) error {
	var b strings.Builder
	b.WriteString("<Row>")
	for _, v := range ValoresFila(f) {
		b.WriteString(`<Cell><Data ss:Type="String">` + xmlEscape(v) + `</Data></Cell>`)
	}
	b.WriteString("</Row>")
	_, err := io.WriteString(w, b.String())
	return err
}

// CerrarExcel writes the closing tags for Table, Worksheet, and Workbook.
func CerrarExcel(w io.Writer) error {
	_, err := io.WriteString(w, "</Table></Worksheet></Workbook>")
	return err
}

func xmlEscape(s string) string {
	var b strings.Builder
	_ = xml.EscapeText(&b, []byte(s))
	return b.String()
}
