package atencion

import (
	"encoding/xml"
	"strings"
)

// CSV arma la exportación básica, con BOM para Excel en español.
func CSV(filas []Fila) string {
	var b strings.Builder
	b.WriteString("\uFEFF")
	b.WriteString("VerdePUCP — Gestión de Áreas Verdes\n")
	b.WriteString("id,titulo,tipo,estado,ejecutor,equipo,zona,codigo_externo,fuente,creada\n")
	for _, f := range filas {
		b.WriteString(csv(f.ID))
		b.WriteByte(',')
		b.WriteString(csv(f.Titulo))
		b.WriteByte(',')
		b.WriteString(csv(f.Tipo))
		b.WriteByte(',')
		b.WriteString(csv(f.Estado))
		b.WriteByte(',')
		b.WriteString(csv(f.Ejecutor))
		b.WriteByte(',')
		b.WriteString(csv(f.Equipo))
		b.WriteByte(',')
		b.WriteString(csv(f.Zona))
		b.WriteByte(',')
		b.WriteString(csv(f.CodigoExterno))
		b.WriteByte(',')
		b.WriteString(csv(f.Fuente))
		b.WriteByte(',')
		b.WriteString(csv(f.CreatedAt))
		b.WriteByte('\n')
	}
	return b.String()
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
	b.WriteString(`<?xml version="1.0" encoding="UTF-8"?>` + "\n")
	b.WriteString(`<?mso-application progid="Excel.Sheet"?>` + "\n")
	b.WriteString(`<Workbook xmlns="urn:schemas-microsoft-com:office:spreadsheet"><Worksheet ss:Name="Labores" xmlns:ss="urn:schemas-microsoft-com:office:spreadsheet"><Table>`)
	b.WriteString(`<Row><Cell><Data ss:Type="String">VerdePUCP — Gestión de Áreas Verdes</Data></Cell></Row>`)
	cabeceras := []string{"id", "titulo", "tipo", "estado", "ejecutor", "equipo", "zona", "codigo_externo", "fuente", "creada"}
	b.WriteString("<Row>")
	for _, h := range cabeceras {
		b.WriteString("<Cell><Data ss:Type=\"String\">" + xmlEscape(h) + "</Data></Cell>")
	}
	b.WriteString("</Row>")
	for _, f := range filas {
		vals := []string{f.ID, f.Titulo, f.Tipo, f.Estado, f.Ejecutor, f.Equipo, f.Zona, f.CodigoExterno, f.Fuente, f.CreatedAt}
		b.WriteString("<Row>")
		for _, v := range vals {
			b.WriteString("<Cell><Data ss:Type=\"String\">" + xmlEscape(v) + "</Data></Cell>")
		}
		b.WriteString("</Row>")
	}
	b.WriteString("</Table></Worksheet></Workbook>")
	return b.String()
}

func xmlEscape(s string) string {
	var b strings.Builder
	_ = xml.EscapeText(&b, []byte(s))
	return b.String()
}
