package etl

import (
	"crypto/sha256"
	"encoding/hex"
	"sort"
	"strings"
	"unicode"
)

// Ficticios fijos del campus (MAPA-DATOS-Y-EDICION.md §2).
// La clave es el conteo del hash, no el nombre de origen.
const (
	ficticio259 = "Valeria Quispe"
	ficticio168 = "Mateo Salazar"
	ficticio104 = "Renato Cárdenas"
)

var ficticiosPorFrecuencia = []string{
	"Nora Beltrán",
	"Iván Paredes",
	"Lucía Mendoza",
}

// etiquetaSector conserva rótulos que no son personas.
func etiquetaSector(normalizado string) (visible string, ok bool) {
	switch normalizado {
	case "campo depo":
		return "campo depo", true
	case "bosque hume":
		return "Bosque húme", true
	default:
		return "", false
	}
}

func normalizarPersona(s string) string {
	s = strings.TrimSpace(strings.ToLower(s))
	var b strings.Builder
	for _, r := range s {
		switch r {
		case 'á', 'à':
			r = 'a'
		case 'é', 'è':
			r = 'e'
		case 'í', 'ì':
			r = 'i'
		case 'ó', 'ò':
			r = 'o'
		case 'ú', 'ù', 'ü':
			r = 'u'
		case 'ñ':
			r = 'n'
		}
		if unicode.Is(unicode.Mn, r) {
			continue
		}
		b.WriteRune(r)
	}
	return strings.Join(strings.Fields(b.String()), " ")
}

func hashPersona(normalizado string) string {
	sum := sha256.Sum256([]byte(normalizado))
	return hex.EncodeToString(sum[:])
}

// esHashPersona reconoce la clave ya anonimizada que se guarda en data/raw.
func esHashPersona(s string) bool {
	if len(s) != 64 {
		return false
	}
	for _, r := range s {
		if (r < '0' || r > '9') && (r < 'a' || r > 'f') {
			return false
		}
	}
	return true
}

// clavePersona devuelve el hash estable. No devuelve el texto de origen.
func clavePersona(valor string) (clave string, etiqueta bool, visible string) {
	valor = strings.TrimSpace(valor)
	if valor == "" {
		return "", false, ""
	}
	if esHashPersona(valor) {
		return valor, false, ""
	}
	norm := normalizarPersona(valor)
	if vis, ok := etiquetaSector(norm); ok {
		return "", true, vis
	}
	return hashPersona(norm), false, ""
}

type cuadrillaFicticia struct {
	ID       string
	Nombre   string
	Clave    string
	Etiqueta bool
}

// asignarCuadrillas reparte ficticios por frecuencia del hash.
// 259, 168 y 104 usan la tabla fija. El resto, en orden de frecuencia.
func asignarCuadrillas(valores []string) map[string]cuadrillaFicticia {
	type grupo struct {
		clave string
		n     int
	}
	counts := map[string]int{}
	etiquetas := map[string]string{}
	for _, valor := range valores {
		clave, etiqueta, visible := clavePersona(valor)
		if etiqueta {
			etiquetas[visible] = visible
			continue
		}
		if clave == "" {
			continue
		}
		counts[clave]++
	}
	grupos := make([]grupo, 0, len(counts))
	for clave, n := range counts {
		grupos = append(grupos, grupo{clave: clave, n: n})
	}
	sort.Slice(grupos, func(i, j int) bool {
		if grupos[i].n != grupos[j].n {
			return grupos[i].n > grupos[j].n
		}
		return grupos[i].clave < grupos[j].clave
	})
	out := map[string]cuadrillaFicticia{}
	resto := 0
	for _, g := range grupos {
		nombre := ficticioPorConteo(g.n)
		if nombre == "" {
			if resto < len(ficticiosPorFrecuencia) {
				nombre = ficticiosPorFrecuencia[resto]
				resto++
			} else {
				nombre = "Cuadrilla sin serie"
			}
		}
		id := "cf-" + g.clave[:12]
		out[g.clave] = cuadrillaFicticia{ID: id, Nombre: nombre, Clave: g.clave}
	}
	for visible := range etiquetas {
		id := "cf-" + strings.ReplaceAll(normalizarPersona(visible), " ", "-")
		out["etiqueta:"+visible] = cuadrillaFicticia{ID: id, Nombre: visible, Etiqueta: true}
	}
	return out
}

func ficticioPorConteo(n int) string {
	switch n {
	case 259:
		return ficticio259
	case 168:
		return ficticio168
	case 104:
		return ficticio104
	default:
		return ""
	}
}

func cuadrillaDe(tabla map[string]cuadrillaFicticia, valor string) (cuadrillaFicticia, bool) {
	clave, etiqueta, visible := clavePersona(valor)
	if etiqueta {
		c, ok := tabla["etiqueta:"+visible]
		return c, ok
	}
	if clave == "" {
		return cuadrillaFicticia{}, false
	}
	c, ok := tabla[clave]
	return c, ok
}
