package etl

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestTachosSumasPublicadas(t *testing.T) {
	body := mustRead(t, filepath.Join(rawDir(t), "sheets", "tachos.csv"))
	rows, rech, sumas, err := LeerTachos(body)
	if err != nil {
		t.Fatal(err)
	}
	conPunto := 0
	for _, row := range rows {
		if !row.SinPunto {
			conPunto++
		}
	}
	if conPunto != 184 {
		t.Fatalf("tachos con coordenada = %d", conPunto)
	}
	sinCoord := 0
	for _, r := range rech {
		if r.Campo == "latitud" {
			sinCoord++
		}
	}
	if sinCoord > 1 {
		t.Fatalf("rechazos de coordenada: %+v", rech)
	}
	for campo, esperado := range ConteosTachosEsperados {
		if sumas[campo] != esperado {
			t.Fatalf("%s = %d, mapa de datos %d", campo, sumas[campo], esperado)
		}
	}
	for _, row := range rows {
		if strings.Contains(row.Nota, "http") || strings.Contains(row.Recomendaciones, "drive.google") {
			t.Fatalf("tacho %s guardó un enlace", row.Codigo)
		}
	}
}

func TestBebederosTienenEstadoYSede(t *testing.T) {
	rows, rech, err := LeerBebederos(rawDir(t))
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 67 {
		t.Fatalf("bebederos = %d, rechazos = %+v", len(rows), rech)
	}
	subtipos := map[string]int{}
	for _, row := range rows {
		if row.Estado == "" || row.Sede == "" {
			t.Fatalf("%s sin estado o sede", row.Codigo)
		}
		subtipos[row.Subtipo]++
		if row.Estado == "sin_dato" {
			continue
		}
		if row.Estado == row.Subtipo && row.Subtipo != "nuevo" {
			t.Fatalf("%s usó el nombre del archivo como estado", row.Codigo)
		}
		if strings.Contains(row.Estado, "<") || strings.Contains(row.Sede, "http") {
			t.Fatalf("%s guardó HTML", row.Codigo)
		}
	}
	if subtipos["fuente"] != 40 || subtipos["llenador"] != 10 || subtipos["nuevo"] != 8 || subtipos["deterioro"] != 7 || subtipos["baja"] != 2 {
		t.Fatalf("subtipos %+v rechazos %+v", subtipos, rech)
	}
}

func TestPuntosSinContacto(t *testing.T) {
	body := mustRead(t, filepath.Join(rawDir(t), "sheets", "puntos_pucp.csv"))
	rows, _, omitidas, err := LeerPuntosPUCP(body)
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 153 {
		t.Fatalf("puntos = %d", len(rows))
	}
	texto := strings.Join(omitidas, ",")
	if !strings.Contains(texto, "phone") || !strings.Contains(texto, "placeId") {
		t.Fatalf("columnas omitidas: %v", omitidas)
	}
	for _, row := range rows {
		if strings.Contains(row.URL, "placeId") || strings.Contains(row.URL, "place_id") || strings.Contains(strings.ToLower(row.URL), "phone") {
			t.Fatalf("punto %s guardó contacto: %s", row.Titulo, row.URL)
		}
	}
}

func TestReservasSoloFicticias(t *testing.T) {
	body, err := os.ReadFile(filepath.Join(rawDir(t), "..", "mocks", "reservas_agenda.mock.json"))
	if err != nil {
		t.Fatal(err)
	}
	rows, rech, err := LeerReservasFicticias(body)
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 71 || len(rech) != 0 {
		t.Fatalf("reservas=%d rechazos=%+v", len(rows), rech)
	}
	for _, row := range rows {
		if !strings.HasPrefix(row.Origen, "FAKE-") {
			t.Fatalf("origen %s", row.Origen)
		}
		if row.Estado != "reservado" && row.Estado != "realizado" && row.Estado != "cancelado" {
			t.Fatalf("estado %s", row.Estado)
		}
	}
	ajeno := []byte(`{"reservas":[{"id":"REAL-1","_fake":false,"fecha":"2026-09-22","hora":"09:00-11:00","estado":"Reservado","evento":"x"}]}`)
	_, rech, err = LeerReservasFicticias(ajeno)
	if err != nil {
		t.Fatal(err)
	}
	if len(rech) != 1 || rech[0].Campo != "origen" {
		t.Fatalf("debía rechazar origen no ficticio: %+v", rech)
	}
}

func TestURLSinPlaceID(t *testing.T) {
	in := "https://www.google.com/maps/search/?api=1&query=Biblioteca&query_place_id=ChIJvQZ"
	out := urlSinContacto(in)
	if strings.Contains(out, "place") || strings.Contains(out, "ChIJ") {
		t.Fatalf("url %s", out)
	}
	if !strings.Contains(out, "query=Biblioteca") {
		t.Fatalf("perdió la url pública: %s", out)
	}
}
