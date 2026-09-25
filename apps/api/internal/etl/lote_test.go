package etl

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoteBaselineConteos(t *testing.T) {
	dir := rawDir(t)

	lug, rech, err := leerLugares(mustRead(t, filepath.Join(dir, "sheets", "lugares.csv")))
	if err != nil {
		t.Fatal(err)
	}
	if len(lug) != 74 || len(rech) != 2 {
		t.Fatalf("lugares=%d rechazos=%+v", len(lug), rech)
	}

	tipos, rech, err := leerActividades(mustRead(t, filepath.Join(dir, "sheets", "actividades.csv")))
	if err != nil {
		t.Fatal(err)
	}
	if len(rech) != 0 {
		t.Fatalf("rechazos de catálogo: %+v", rech)
	}
	clases := map[string]int{}
	for _, tipo := range tipos {
		clases[tipo.Clase]++
	}
	if len(tipos) != 45 || len(clases) != 7 {
		t.Fatalf("tipos=%d clases=%d (%v)", len(tipos), len(clases), clases)
	}

	med, rech, err := leerMedidasPalmera(mustRead(t, filepath.Join(dir, "sheets", "flora.csv")))
	if err != nil {
		t.Fatal(err)
	}
	if len(med) != 74 || len(rech) != 2 {
		t.Fatalf("palmeras válidas=%d rechazadas=%d", len(med), len(rech))
	}
	motivos := map[string]int{}
	for _, r := range rech {
		motivos[r.Campo]++
		if r.Motivo == "" {
			t.Fatalf("rechazo sin motivo: %+v", r)
		}
	}
	if motivos["latitud"] != 1 || motivos["longitud"] != 1 {
		t.Fatalf("motivos: %+v rechazos=%+v", motivos, rech)
	}

	caf, _, rech, err := leerCafetos(mustRead(t, filepath.Join(dir, "sheets", "cafetos.csv")))
	if err != nil {
		t.Fatal(err)
	}
	if len(caf) != 53 {
		t.Fatalf("cafetos válidos=%d, rechazos=%d", len(caf), len(rech))
	}
	if len(rech) < 2 {
		t.Fatalf("se esperaban al menos 2 cafetos rechazados, hubo %d", len(rech))
	}
}

func TestGvizCanonTipo(t *testing.T) {
	body := []byte(`{"table":{"cols":[{"label":"N°"},{"label":"Latitud"},{"label":"Longitud"},{"label":"Tipo de vegetación (Formas biológicas)"},{"label":"Nombre científico"}],"rows":[{"c":[{"v":1},{"v":-12.07},{"v":-77.08},{"v":"Planta herbácea"},{"v":"Species demo"}]}]}}`)
	rows, rech, err := leerGvizFlora(body)
	if err != nil {
		t.Fatal(err)
	}
	if len(rech) != 0 || len(rows) != 1 || rows[0].Tipo != "Herbácea" {
		t.Fatalf("rows=%+v rech=%+v", rows, rech)
	}
}

func TestAnonimizaSinFragmentosPersonales(t *testing.T) {
	personas := fragmentosPersonales()
	if len(personas) == 0 {
		t.Fatal("baseline.go no declara fragmentos personales")
	}
	var valores []string
	for i := 0; i < 259; i++ {
		valores = append(valores, personas[0])
	}
	if len(personas) > 1 {
		for i := 0; i < 168; i++ {
			valores = append(valores, personas[1])
		}
	}
	valores = append(valores, "campo depo", "Bosque húme", "campo depo")
	tabla := asignarCuadrillas(valores)
	var blob strings.Builder
	for _, c := range tabla {
		blob.WriteString(c.Nombre)
		blob.WriteByte('\n')
		blob.WriteString(c.ID)
		blob.WriteByte('\n')
	}
	texto := blob.String()
	for _, frag := range personas {
		if strings.Contains(texto, frag) {
			t.Fatalf("fragmento personal %q en %q", frag, texto)
		}
	}
	if !strings.Contains(texto, "Valeria Quispe") || !strings.Contains(texto, "campo depo") || !strings.Contains(texto, "Bosque húme") {
		t.Fatalf("faltan ficticio o etiquetas: %s", texto)
	}
	clave, etiqueta, _ := clavePersona(personas[0])
	if etiqueta || !esHashPersona(clave) {
		t.Fatal("la clave de una persona debe ser un hash")
	}
	otra, _, _ := clavePersona("  " + strings.ToUpper(personas[0]) + " ")
	if otra != clave {
		t.Fatal("el hash no es estable ante mayúsculas y espacios")
	}
}

func TestCacheJefesNoGuardaFragmentos(t *testing.T) {
	personas := fragmentosPersonales()
	body := []byte(`{"type":"FeatureCollection","features":[{"type":"Feature","properties":{"jefes":"` + personas[0] + `","Nombre":"Jardín"},"geometry":{"type":"Polygon","coordinates":[[[-77.08,-12.07],[-77.079,-12.07],[-77.079,-12.069],[-77.08,-12.07]]]}}]}`)
	out, err := anonimizarJefes(body)
	if err != nil {
		t.Fatal(err)
	}
	for _, frag := range personas {
		if strings.Contains(string(out), frag) {
			t.Fatalf("la copia conserva %q", frag)
		}
	}
	pols, _, err := leerPoligonos(out)
	if err != nil {
		t.Fatal(err)
	}
	if len(pols) != 1 || !esHashPersona(pols[0].JefesRaw) {
		t.Fatalf("jefes=%q", pols[0].JefesRaw)
	}
}

func fragmentosPersonales() []string {
	vistos := map[string]bool{}
	var out []string
	for _, frag := range PIIFragments {
		if frag == "" || vistos[frag] {
			continue
		}
		if _, ok := etiquetaSector(normalizarPersona(frag)); ok {
			continue
		}
		if frag == "jefes" {
			continue
		}
		vistos[frag] = true
		out = append(out, frag)
	}
	return out
}

func mustRead(t *testing.T, path string) []byte {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return b
}
