package etl

import (
	"strings"
)

// Ficticios de campus, en el orden de la sección 2 del mapa de datos.
// La clave interna es el hash; el nombre real no se guarda.
var ficticiosJefes = []string{"Valeria Quispe", "Mateo Salazar", "Renato Cárdenas"}

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

// ficticiosEstables es la lista completa. El índice sale del hash, no del conteo del lote.
var ficticiosEstables = append(append([]string{}, ficticiosJefes...), ficticiosMonitoreo...)

// Aplicar asigna un ficticio por hash del nombre. El mismo responsable
// recibe el mismo nombre en cualquier lote, aunque cambie cuántas filas trae.
func (t *Tabla) Aplicar(nombres []string) {
	for _, nombre := range nombres {
		norma := normalizarPersona(nombre)
		if norma == "" || EsEtiqueta(nombre) {
			continue
		}
		h := hashPersona(norma)
		if _, ya := t.porHash[h]; ya {
			continue
		}
		t.porHash[h] = ficticioDeHash(h)
	}
}

func ficticioDeHash(hash string) string {
	if len(ficticiosEstables) == 0 {
		return ""
	}
	return ficticiosEstables[indiceEstable(hash)%len(ficticiosEstables)]
}

// indiceEstable elige el ficticio con el hash del nombre, no con el orden del lote.
func indiceEstable(hash string) int {
	n := 0
	for i := 0; i < len(hash) && i < 8; i++ {
		c := hash[i]
		n *= 16
		switch {
		case c >= '0' && c <= '9':
			n += int(c - '0')
		case c >= 'a' && c <= 'f':
			n += int(c-'a') + 10
		}
	}
	if n < 0 {
		n = -n
	}
	return n
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
