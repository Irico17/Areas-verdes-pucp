// Package services contains cross-cutting domain services.
package services

import (
	"strings"

	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/contracts"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/entities"
)

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

type sugeridorTipoService struct{}

// NewSugeridorTipoService creates a new ISugeridorTipo instance using local heuristics.
func NewSugeridorTipoService() contracts.ISugeridorTipo {
	return &sugeridorTipoService{}
}

// SugerirTipo applies keyword heuristics over the title.
func (s *sugeridorTipoService) SugerirTipo(titulo string) entities.SugerenciaIA {
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
		return entities.SugerenciaIA{
			Explicacion:    "No hay una pista suficiente en el título. Elija el tipo.",
			Confianza:      "nula",
			RequiereHumano: true,
		}
	}
	return entities.SugerenciaIA{
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
