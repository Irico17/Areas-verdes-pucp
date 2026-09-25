package etl

import (
	"archive/zip"
	"bytes"
	"strconv"
	"strings"
	"testing"
)

func TestCSVInvalidoNoTieneFilasParaEscribir(t *testing.T) {
	body := []byte("lugar,latitud,longitud\nAfuera,-1.20,-7.70\n")
	vista, err := Previsualizar("lugares", "lugares.csv", body)
	if err != nil {
		t.Fatal(err)
	}
	if vista.Escrito {
		t.Fatal("la vista previa no debe marcar escritura")
	}
	if vista.Validas != 0 {
		t.Fatalf("válidas %d", vista.Validas)
	}
	if len(vista.Errores) != 1 || vista.Errores[0].Campo == "" || vista.Errores[0].Motivo == "" {
		t.Fatalf("errores %+v", vista.Errores)
	}
}

func TestLugarEnCampusQuedaEnLaVista(t *testing.T) {
	body := []byte("lugar,latitud,longitud\nBiblioteca,-12.0704,-77.0808\n")
	vista, err := Previsualizar("lugares", "lugares.csv", body)
	if err != nil {
		t.Fatal(err)
	}
	if vista.Validas != 1 || vista.Escrito {
		t.Fatalf("%+v", vista)
	}
	if len(vista.Filas) != 1 {
		t.Fatalf("filas %d", len(vista.Filas))
	}
}

func TestXLSXExigeHojaConElNombreDeLaEntidad(t *testing.T) {
	malo := xlsxMinimo(t, "otra", [][]string{{"lugar", "latitud", "longitud"}, {"Biblioteca", "-12.07", "-77.08"}})
	if _, err := Previsualizar("lugares", "lugares.xlsx", malo); err == nil || !strings.Contains(err.Error(), "lugares") {
		t.Fatalf("err %v", err)
	}
	bueno := xlsxMinimo(t, "lugares", [][]string{{"lugar", "latitud", "longitud"}, {"Biblioteca", "-12.07", "-77.08"}})
	vista, err := Previsualizar("lugares", "lugares.xlsx", bueno)
	if err != nil {
		t.Fatal(err)
	}
	if vista.Formato != "xlsx" || vista.Validas != 1 {
		t.Fatalf("%+v", vista)
	}
}

func TestEntidadFueraDelMapa(t *testing.T) {
	_, err := Previsualizar("indicadores", "x.csv", []byte("a\n1\n"))
	if err == nil {
		t.Fatal("debía rechazar la entidad")
	}
}

func TestPuntosAnunciaColumnasOmitidas(t *testing.T) {
	body := []byte("title,location/lat,location/lng,phone,placeId,url\nBiblioteca,-12.0704,-77.0808,999,abc,https://maps.example/p\n")
	vista, err := Previsualizar("puntos_pucp", "puntos_pucp.csv", body)
	if err != nil {
		t.Fatal(err)
	}
	if vista.AvisoOmitidas != "columnas omitidas por datos personales" {
		t.Fatalf("aviso %q", vista.AvisoOmitidas)
	}
	if vista.Validas != 1 {
		t.Fatalf("válidas %d errores %+v", vista.Validas, vista.Errores)
	}
}

func TestReservaNoFicticiaNoEsValida(t *testing.T) {
	body := []byte("origen,fecha,hora_inicio,hora_fin,estado,evento,unidad,jardin_id\nreal,2026-05-01,08:00,09:00,reservado,acto,DAF,JR-1\n")
	vista, err := Previsualizar("reservas", "reservas.csv", body)
	if err != nil {
		t.Fatal(err)
	}
	if vista.Validas != 0 || len(vista.Errores) == 0 {
		t.Fatalf("%+v", vista)
	}
}

func xlsxMinimo(t *testing.T, hoja string, filas [][]string) []byte {
	t.Helper()
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	escribir := func(nombre, contenido string) {
		t.Helper()
		w, err := zw.Create(nombre)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := w.Write([]byte(contenido)); err != nil {
			t.Fatal(err)
		}
	}
	var celdas strings.Builder
	for r, fila := range filas {
		celdas.WriteString(`<row>`)
		for c, valor := range fila {
			ref := string(rune('A'+c)) + strconv.Itoa(r+1)
			celdas.WriteString(`<c r="` + ref + `" t="inlineStr"><is><t>` + valor + `</t></is></c>`)
		}
		celdas.WriteString(`</row>`)
	}
	escribir("[Content_Types].xml", `<?xml version="1.0"?><Types></Types>`)
	escribir("xl/workbook.xml", `<?xml version="1.0"?><workbook xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships"><sheets><sheet name="`+hoja+`" sheetId="1" r:id="rId1"/></sheets></workbook>`)
	escribir("xl/_rels/workbook.xml.rels", `<?xml version="1.0"?><Relationships><Relationship Id="rId1" Type="worksheet" Target="worksheets/sheet1.xml"/></Relationships>`)
	escribir("xl/worksheets/sheet1.xml", `<?xml version="1.0"?><worksheet><sheetData>`+celdas.String()+`</sheetData></worksheet>`)
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}
