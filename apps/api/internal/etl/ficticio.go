package etl

import (
	"crypto/sha256"
	"encoding/hex"
	"sort"
	"strings"
)

// Ficticios de campus, en el orden de la sección 2 del mapa de datos.
// La clave interna es el hash; el nombre real no se guarda.
var ficticiosJefes = []string{"Valeria Quispe", "Mateo Salazar", "Renato Cárdenas"}
var conteosJefes = []int{259, 168, 104}

var ficticiosMonitoreo = []string{"Nora Beltrán", "Iván Paredes", "Lucía Mendoza"}

// etiquetasConservadas no son personas.
var etiquetasConservadas = map[string]string{
	"campo depo":     "campo depo",
	"bosque hume":    "Bosque húme",
	"supervision":    "Supervisión",
	"unidad":         "Unidad",
	"ningun reporte": "Ningun reporte",
}

// Tabla asigna un ficticio estable por hash normalizado.
type Tabla struct {
	porHash map[string]string
}

func NuevaTabla() *Tabla {
	return &Tabla{porHash: map[string]string{}}
}

func normalizarPersona(s string) string {
	s = strings.TrimSpace(strings.ToLower(s))
	r := strings.NewReplacer(
		"á", "a", "é", "e", "í", "i", "ó", "o", "ú", "u", "ü", "u", "ñ", "n",
	)
	return strings.Join(strings.Fields(r.Replace(s)), " ")
}

func hashPersona(normalizado string) string {
	sum := sha256.Sum256([]byte(normalizado))
	return hex.EncodeToString(sum[:])
}

// EsEtiqueta indica un rótulo que se conserva tal cual.
func EsEtiqueta(nombre string) bool {
	_, ok := etiquetasConservadas[normalizarPersona(nombre)]
	return ok
}

// ConservarEtiqueta devuelve el rótulo canónico.
func ConservarEtiqueta(nombre string) string {
	if v, ok := etiquetasConservadas[normalizarPersona(nombre)]; ok {
		return v
	}
	return ""
}

type grupo struct {
	hash  string
	norma string
	n     int
}

// Aplicar asigna ficticios por frecuencia. Los conteos 259, 168 y 104
// usan la terna de jefes. El resto, la terna de monitoreo, en ese orden.
// El mismo hash siempre cae en el mismo ficticio dentro de la tabla.
func (t *Tabla) Aplicar(nombres []string) {
	counts := map[string]*grupo{}
	for _, nombre := range nombres {
		norma := normalizarPersona(nombre)
		if norma == "" || EsEtiqueta(nombre) {
			continue
		}
		h := hashPersona(norma)
		g, ok := counts[h]
		if !ok {
			g = &grupo{hash: h, norma: norma}
			counts[h] = g
		}
		g.n++
	}
	var grupos []grupo
	for _, g := range counts {
		if _, ya := t.porHash[g.hash]; ya {
			continue
		}
		grupos = append(grupos, *g)
	}
	sort.Slice(grupos, func(i, j int) bool {
		if grupos[i].n != grupos[j].n {
			return grupos[i].n > grupos[j].n
		}
		return grupos[i].hash < grupos[j].hash
	})
	jefes := map[int]string{}
	for i, n := range conteosJefes {
		jefes[n] = ficticiosJefes[i]
	}
	usoMon := 0
	for _, g := range grupos {
		if nombre, ok := jefes[g.n]; ok {
			t.porHash[g.hash] = nombre
			continue
		}
		if usoMon < len(ficticiosMonitoreo) {
			t.porHash[g.hash] = ficticiosMonitoreo[usoMon]
			usoMon++
			continue
		}
		t.porHash[g.hash] = ficticiosMonitoreo[len(ficticiosMonitoreo)-1]
	}
}

// Ficticio sustituye un nombre de persona. Vacío sigue vacío.
// No devuelve el nombre de entrada ni el hash.
func (t *Tabla) Ficticio(nombre string) string {
	norma := normalizarPersona(nombre)
	if norma == "" {
		return ""
	}
	if et := ConservarEtiqueta(nombre); et != "" {
		return et
	}
	h := hashPersona(norma)
	if v, ok := t.porHash[h]; ok {
		return v
	}
	t.Aplicar([]string{nombre})
	if v, ok := t.porHash[h]; ok {
		return v
	}
	return ficticiosMonitoreo[0]
}

// Lista separa por coma y anonimiza cada tramo.
func (t *Tabla) Lista(valor string) string {
	if strings.TrimSpace(valor) == "" {
		return ""
	}
	partes := strings.Split(valor, ",")
	var out []string
	for _, p := range partes {
		if f := t.Ficticio(p); f != "" {
			out = append(out, f)
		}
	}
	return strings.Join(out, ", ")
}
