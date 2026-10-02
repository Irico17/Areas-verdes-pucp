package etl

import (
	"strings"
)

func duplicadosPorIndice(areas []Record) map[int]string {
	vistos := map[string]bool{}
	duplicados := map[int]string{}
	for i, r := range areas {
		if r.Codigo == nil {
			continue
		}
		codigo := *r.Codigo
		if strings.TrimSpace(codigo) == "" {
			continue
		}
		if vistos[codigo] {
			duplicados[i] = codigo
			continue
		}
		vistos[codigo] = true
	}
	return duplicados
}
