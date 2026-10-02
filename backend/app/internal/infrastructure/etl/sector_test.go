package etl

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"testing"
)

// featureJefes arma una FeatureCollection sintética con nombres de ejemplo,
// nunca reales, para probar el agrupamiento sin tocar datos de personas.
func featureJefes(valores []string) []byte {
	geom := `{"type":"Polygon","coordinates":[[[-77.08,-12.07],[-77.079,-12.07],[-77.079,-12.069],[-77.08,-12.07]]]}`
	var b strings.Builder
	b.WriteString(`{"type":"FeatureCollection","features":[`)
	for i, v := range valores {
		if i > 0 {
			b.WriteByte(',')
		}
		props := map[string]any{"código": fmt.Sprintf("X-%02d", i)}
		if v != "" {
			props["jefes"] = v
		}
		propsJSON, _ := json.Marshal(props)
		fmt.Fprintf(&b, `{"type":"Feature","geometry":%s,"properties":%s}`, geom, propsJSON)
	}
	b.WriteString(`]}`)
	return []byte(b.String())
}

func TestSectoresDeJefesAgrupaPorHash(t *testing.T) {
	valores := []string{}
	for i := 0; i < 5; i++ {
		valores = append(valores, "Ana Ejemplo")
	}
	for i := 0; i < 3; i++ {
		valores = append(valores, "Luis Ejemplo")
	}
	for i := 0; i < 2; i++ {
		valores = append(valores, "Rosa Ejemplo")
	}
	valores = append(valores, "campo depo", "campo depo", "Bosque húme", "")

	body := featureJefes(valores)
	out, err := SectoresDeJefes(body, "test")
	if err != nil {
		t.Fatal(err)
	}
	if out.Conteos["cua-valeria"] != 5 || out.Conteos["cua-mateo"] != 3 || out.Conteos["cua-renato"] != 2 {
		t.Fatalf("conteos por rango: %+v", out.Conteos)
	}
	if out.Conteos["campo-deportivo"] != 2 || out.Conteos["bosque-humedo"] != 1 {
		t.Fatalf("conteos por etiqueta: %+v", out.Conteos)
	}
	if len(out.Poligonos) != len(valores)-1 { // el vacío ("") no genera fila
		t.Fatalf("filas = %d, se esperaban %d", len(out.Poligonos), len(valores)-1)
	}
	for i, p := range out.Poligonos {
		if p.SourceIndex != i {
			t.Fatalf("fila %d: source_index=%d", i, p.SourceIndex)
		}
		if p.FeatureID != fmt.Sprintf("Z-%04d", i+1) {
			t.Fatalf("fila %d: feature_id=%s", i, p.FeatureID)
		}
	}
	if out.Poligonos[0].Sector != "cua-valeria" || out.Poligonos[5].Sector != "cua-mateo" || out.Poligonos[8].Sector != "cua-renato" {
		t.Fatalf("orden de sectores: %+v", out.Poligonos)
	}

	blob, err := json.Marshal(out)
	if err != nil {
		t.Fatal(err)
	}
	texto := string(blob)
	for _, prohibido := range []string{"Ejemplo", "jefes", "responsable"} {
		if strings.Contains(texto, prohibido) {
			t.Fatalf("el JSON contiene %q", prohibido)
		}
	}
	if regexp.MustCompile(`[0-9a-f]{64}`).MatchString(texto) {
		t.Fatalf("el JSON contiene un hash de 64 caracteres: %s", texto)
	}
}

func TestSectoresFallaConMasDeTresGrupos(t *testing.T) {
	valores := []string{}
	for i := 0; i < 5; i++ {
		valores = append(valores, "Persona Uno")
	}
	for i := 0; i < 4; i++ {
		valores = append(valores, "Persona Dos")
	}
	for i := 0; i < 3; i++ {
		valores = append(valores, "Persona Tres")
	}
	for i := 0; i < 2; i++ {
		valores = append(valores, "Persona Cuatro")
	}
	body := featureJefes(valores)
	if _, err := SectoresDeJefes(body, "test"); err == nil {
		t.Fatal("se esperaba error con 4 grupos de hash")
	}
}

func TestSectoresFallaConEmpate(t *testing.T) {
	valores := []string{}
	for i := 0; i < 5; i++ {
		valores = append(valores, "Persona Uno")
	}
	for i := 0; i < 5; i++ {
		valores = append(valores, "Persona Dos")
	}
	for i := 0; i < 3; i++ {
		valores = append(valores, "Persona Tres")
	}
	body := featureJefes(valores)
	_, err := SectoresDeJefes(body, "test")
	if err == nil {
		t.Fatal("se esperaba error por empate")
	}
	if strings.Contains(err.Error(), "Persona") {
		t.Fatalf("el mensaje de error no debe mostrar nombres: %v", err)
	}
}

// transporteQueFalla simula una red caída: FuenteSectores no debe caer a
// data/raw/jefe_de_grupo.json (la copia base, con "responsable").
type transporteQueFalla struct{}

func (transporteQueFalla) RoundTrip(*http.Request) (*http.Response, error) {
	return nil, fmt.Errorf("red no disponible en la prueba")
}

func TestFuenteSectoresNoUsaBaseline(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "jefe_de_grupo.json"), []byte(`{"type":"FeatureCollection","features":[{"type":"Feature","properties":{"jefes":"responsable"},"geometry":null}]}`), 0o644); err != nil {
		t.Fatal(err)
	}
	client := &http.Client{Transport: transporteQueFalla{}}

	if _, _, err := FuenteSectores(dir, true, client); err == nil {
		t.Fatal("se esperaba error: sin lote/ y con red caída, no debe caer a la copia base")
	}
	if _, _, err := FuenteSectores(dir, false, client); err == nil {
		t.Fatal("se esperaba error: sin -vivo y sin copia en lote/, no debe caer a la copia base")
	}
}

func TestFuenteSectoresUsaLaCopiaDeLote(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "lote"), 0o755); err != nil {
		t.Fatal(err)
	}
	contenido := []byte(`{"type":"FeatureCollection","features":[]}`)
	if err := os.WriteFile(filepath.Join(dir, "lote", "jefe_de_grupo.json"), contenido, 0o644); err != nil {
		t.Fatal(err)
	}
	client := &http.Client{Transport: transporteQueFalla{}}
	body, desde, err := FuenteSectores(dir, false, client)
	if err != nil {
		t.Fatal(err)
	}
	if desde != "lote" || string(body) != string(contenido) {
		t.Fatalf("desde=%s body=%s", desde, body)
	}
}

func TestZonasSectorV1(t *testing.T) {
	path := filepath.Join(v1Dir(t), "zonas_sector.json")
	body, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var out ArchivoSectores
	if err := json.Unmarshal(body, &out); err != nil {
		t.Fatal(err)
	}
	if len(out.Poligonos) != ExpectedZonas {
		t.Fatalf("filas = %d, se esperaban %d", len(out.Poligonos), ExpectedZonas)
	}
	vistos := map[int]bool{}
	for i, p := range out.Poligonos {
		if vistos[p.SourceIndex] {
			t.Fatalf("source_index repetido: %d", p.SourceIndex)
		}
		vistos[p.SourceIndex] = true
		if p.FeatureID != fmt.Sprintf("Z-%04d", p.SourceIndex+1) {
			t.Fatalf("fila %d: feature_id=%s source_index=%d", i, p.FeatureID, p.SourceIndex)
		}
	}
	esperado := map[string]int{"cua-valeria": 259, "cua-mateo": 168, "cua-renato": 104, "campo-deportivo": 2, "bosque-humedo": 1}
	for sector, n := range esperado {
		if out.Conteos[sector] != n {
			t.Fatalf("conteo[%s] = %d, se esperaba %d", sector, out.Conteos[sector], n)
		}
	}

	texto := string(body)
	for _, frag := range PIIFragments {
		if strings.Contains(texto, frag) {
			t.Fatalf("data/v1/zonas_sector.json contiene %q", frag)
		}
	}
	if regexp.MustCompile(`[0-9a-f]{64}`).MatchString(texto) {
		t.Fatal("data/v1/zonas_sector.json contiene un hash de 64 caracteres")
	}
}

var reArregloSector = regexp.MustCompile(`unnest\('\{([0-9,]*)\}'::int\[\]\), '([a-z-]+)'`)

func TestMigracion044CoincideConV1(t *testing.T) {
	sqlBody, err := os.ReadFile(filepath.Join(migrationsDir(t), "044_sector_poligonos.sql"))
	if err != nil {
		t.Fatal(err)
	}
	sql := string(sqlBody)

	jsonBody, err := os.ReadFile(filepath.Join(v1Dir(t), "zonas_sector.json"))
	if err != nil {
		t.Fatal(err)
	}
	var out ArchivoSectores
	if err := json.Unmarshal(jsonBody, &out); err != nil {
		t.Fatal(err)
	}
	deV1 := map[string]bool{}
	for _, p := range out.Poligonos {
		deV1[strconv.Itoa(p.SourceIndex)+":"+p.Sector] = true
	}

	deSQL := map[string]bool{}
	sectoresEnSQL := map[string]bool{}
	for _, m := range reArregloSector.FindAllStringSubmatch(sql, -1) {
		indices, sector := m[1], m[2]
		sectoresEnSQL[sector] = true
		if indices == "" {
			continue
		}
		for _, idx := range strings.Split(indices, ",") {
			deSQL[idx+":"+sector] = true
		}
	}
	if len(deSQL) != len(deV1) {
		t.Fatalf("la migración trae %d pares (índice, sector); data/v1/zonas_sector.json trae %d", len(deSQL), len(deV1))
	}
	for par := range deV1 {
		if !deSQL[par] {
			t.Fatalf("falta en la migración: %s", par)
		}
	}

	for _, sector := range SectoresValidos {
		if !sectoresEnSQL[sector] {
			t.Fatalf("la migración no trae ningún arreglo para %s", sector)
		}
	}

	checks := regexp.MustCompile(`CHECK \(sector IS NULL OR sector IN \(([^)]*)\)\)|CHECK\s*\n?\s*\(sector IN \(([^)]*)\)\)`).FindAllStringSubmatch(sql, -1)
	if len(checks) != 2 {
		t.Fatalf("se esperaban 2 CHECK con la lista de sectores, hubo %d", len(checks))
	}
	for _, m := range checks {
		lista := m[1]
		if lista == "" {
			lista = m[2]
		}
		var got []string
		for _, s := range strings.Split(lista, ",") {
			got = append(got, strings.Trim(strings.TrimSpace(s), "'"))
		}
		sort.Strings(got)
		want := append([]string{}, SectoresValidos...)
		sort.Strings(want)
		if strings.Join(got, ",") != strings.Join(want, ",") {
			t.Fatalf("CHECK trae %v, se esperaba %v", got, want)
		}
	}
}

func v1Dir(t *testing.T) string {
	t.Helper()
	return filepath.Join(rawDir(t), "..", "v1")
}

func migrationsDir(t *testing.T) string {
	t.Helper()
	return filepath.Join(rawDir(t), "..", "..", "apps", "api", "migrations")
}
