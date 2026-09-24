package atencion

import "strings"

// Sugerencia es una pista local a partir del título. No sale de este proceso
// y no llama a un modelo externo.
type Sugerencia struct {
	Codigo         string `json:"codigo"`
	Etiqueta       string `json:"etiqueta"`
	Explicacion    string `json:"explicacion"`
	Confianza      string `json:"confianza"`
	RequiereHumano bool   `json:"requiere_humano"`
}

var pistas = []struct {
	codigo, etiqueta string
	palabras         []string
}{
	{"riego", "Riego", []string{"riego", "aspersor", "agua", "regador"}},
	{"poda", "Poda", []string{"poda", "seto", "cortar", "ramas"}},
	{"limpieza", "Limpieza", []string{"limpieza", "barrido", "residuo", "hojas"}},
	{"incidencia", "Incidencia", []string{"fuga", "rotura", "dano", "danada"}},
	{"inspeccion", "Inspección", []string{"inspeccion", "recorrido", "control", "revisar"}},
}

// SugerirTipo aplica reglas sobre el título. Si no hay pista, el código queda vacío.
func SugerirTipo(titulo string) Sugerencia {
	texto := fold(titulo)
	mejor := ""
	etiqueta := ""
	puntos := 0
	for _, pista := range pistas {
		n := 0
		for _, palabra := range pista.palabras {
			if strings.Contains(texto, palabra) {
				n++
			}
		}
		if n > puntos {
			puntos = n
			mejor = pista.codigo
			etiqueta = pista.etiqueta
		}
	}
	if puntos == 0 {
		return Sugerencia{
			Explicacion:    "No hay una pista suficiente en el título. Elija el tipo.",
			Confianza:      "nula",
			RequiereHumano: true,
		}
	}
	return Sugerencia{
		Codigo:         mejor,
		Etiqueta:       etiqueta,
		Explicacion:    "Regla local sobre el título. No es un modelo externo: confirme antes de guardar.",
		Confianza:      "baja",
		RequiereHumano: true,
	}
}

func fold(s string) string {
	s = strings.ToLower(s)
	r := strings.NewReplacer("á", "a", "é", "e", "í", "i", "ó", "o", "ú", "u", "ü", "u", "ñ", "n")
	return r.Replace(s)
}
