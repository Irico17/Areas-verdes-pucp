package etl

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm"
)

// SectoresValidos son los sectores operativos que puede llevar
// poligonos_cuadrilla.sector: 3 cuadrillas ficticias, por frecuencia de mayor
// a menor, y 2 rótulos de lugar. Mismo orden que el CHECK de la migración 044
// y que MAPA-DATOS-Y-EDICION.md §2.
var SectoresValidos = []string{"cua-valeria", "cua-mateo", "cua-renato", "campo-deportivo", "bosque-humedo"}

// SectorPoligono es una fila de data/v1/zonas_sector.json.
type SectorPoligono struct {
	SourceIndex int    `json:"source_index"`
	FeatureID   string `json:"feature_id"`
	Sector      string `json:"sector"`
}

// ArchivoSectores es el contenido de data/v1/zonas_sector.json. No lleva
// hashes ni ninguna huella de los nombres de origen: solo slugs, índices y
// conteos.
type ArchivoSectores struct {
	Version   int              `json:"version"`
	Fuente    string           `json:"fuente"`
	Clave     string           `json:"clave"`
	Nota      string           `json:"nota"`
	Conteos   map[string]int   `json:"conteos"`
	Poligonos []SectorPoligono `json:"poligonos"`
}

// AnonimizarJefes expone anonimizarJefes para cmd/sectores: CompararOrden
// necesita el mismo cuerpo anonimizado que ya calcula SectoresDeJefes por
// dentro, y no vale la pena hacer público un segundo camino que anonimice dos
// veces con lógicas distintas.
func AnonimizarJefes(body []byte) ([]byte, error) {
	return anonimizarJefes(body)
}

// FuenteSectores da el cuerpo de jefe_de_grupo.json para calcular sectores.
// Sin vivo, usa solo la copia versionada data/raw/lote/jefe_de_grupo.json (ya
// anonimizada, sin red). Con vivo, intenta la fuente en línea primero y cae a
// esa copia si falla. Nunca lee data/raw/jefe_de_grupo.json: esa copia base
// sustituye todos los jefes por "responsable" y junta los tres grupos en uno solo.
func FuenteSectores(rawDir string, vivo bool, client *http.Client) (body []byte, desde string, err error) {
	cache := filepath.Join(rawDir, "lote", "jefe_de_grupo.json")
	if vivo {
		if client == nil {
			client = &http.Client{Timeout: 45 * time.Second}
		}
		if b, derr := descargar(client, urlJefes); derr == nil && len(bytes.TrimSpace(b)) > 0 {
			return b, "vivo", nil
		}
	}
	b, rerr := os.ReadFile(cache)
	if rerr != nil || len(bytes.TrimSpace(b)) == 0 {
		return nil, "", fmt.Errorf("sectores: sin copia local en %s: %w", cache, rerr)
	}
	return b, "lote", nil
}

type asignado struct {
	clave    string
	etiqueta bool
	visible  string
}

// SectoresDeJefes agrupa jefe_de_grupo.json por sector. Primero anonimiza el
// cuerpo (es idempotente si ya venía anonimizado), así el resto de la función
// solo ve hashes o etiquetas, nunca nombres. Exige exactamente 3 grupos de
// hash sin empate entre sus conteos: son los que reciben cua-valeria (el más
// numeroso), cua-mateo y cua-renato, en ese orden. No exige un conteo fijo,
// porque la fuente viva cambia con el tiempo (ver MAPA-DATOS-Y-EDICION.md §2).
func SectoresDeJefes(body []byte, desde string) (ArchivoSectores, error) {
	anon, err := anonimizarJefes(body)
	if err != nil {
		return ArchivoSectores{}, err
	}
	fc, err := parseFC(anon)
	if err != nil {
		return ArchivoSectores{}, fmt.Errorf("sectores: %w", err)
	}

	conteoHash := map[string]int{}
	porIndice := make([]asignado, len(fc.Features))
	for i, ft := range fc.Features {
		if ft.Properties == nil {
			continue
		}
		raw, _ := ft.Properties["jefes"].(string)
		if raw == "" {
			continue
		}
		clave, etiqueta, visible := clavePersona(raw)
		if etiqueta {
			porIndice[i] = asignado{etiqueta: true, visible: visible}
			continue
		}
		if clave == "" {
			continue
		}
		conteoHash[clave]++
		porIndice[i] = asignado{clave: clave}
	}

	type grupo struct {
		clave string
		n     int
	}
	grupos := make([]grupo, 0, len(conteoHash))
	for clave, n := range conteoHash {
		grupos = append(grupos, grupo{clave: clave, n: n})
	}
	sort.Slice(grupos, func(i, j int) bool { return grupos[i].n > grupos[j].n })
	if len(grupos) != 3 {
		return ArchivoSectores{}, fmt.Errorf("sectores: se esperaban 3 grupos de hash, hubo %d", len(grupos))
	}
	if grupos[0].n == grupos[1].n || grupos[1].n == grupos[2].n {
		return ArchivoSectores{}, fmt.Errorf(
			"sectores: empate entre grupos de hash (conteos %d, %d, %d); el orden por frecuencia queda ambiguo",
			grupos[0].n, grupos[1].n, grupos[2].n)
	}

	slugPorHash := map[string]string{
		grupos[0].clave: "cua-valeria",
		grupos[1].clave: "cua-mateo",
		grupos[2].clave: "cua-renato",
	}
	slugPorEtiqueta := map[string]string{
		"campo depo":  "campo-deportivo",
		"Bosque húme": "bosque-humedo",
	}

	poligonos := make([]SectorPoligono, 0, len(fc.Features))
	conteos := map[string]int{}
	for i, a := range porIndice {
		var slug string
		switch {
		case a.etiqueta:
			slug = slugPorEtiqueta[a.visible]
		case a.clave != "":
			slug = slugPorHash[a.clave]
		}
		if slug == "" {
			continue
		}
		poligonos = append(poligonos, SectorPoligono{
			SourceIndex: i,
			FeatureID:   fmt.Sprintf("Z-%04d", i+1),
			Sector:      slug,
		})
		conteos[slug]++
	}

	return ArchivoSectores{
		Version:   1,
		Fuente:    fmt.Sprintf("jefe_de_grupo.json (fuente: %s)", desde),
		Clave:     "source_index",
		Nota:      "sector es un id de cuadrilla ficticia (cua-*) o un rótulo de lugar; no identifica personas",
		Conteos:   conteos,
		Poligonos: poligonos,
	}, nil
}

// CompararOrden exige que la feature i de la fuente anonimizada coincida con
// data/v1/zonas.geojson[i] en código, nombre y área (redondeada a 3
// decimales). Si no coinciden, el origen ya no está alineado por índice con
// el catastro normalizado y asignar sector por source_index sería incorrecto.
func CompararOrden(anon []byte, zonasV1 []byte) error {
	fc, err := parseFC(anon)
	if err != nil {
		return fmt.Errorf("comparar orden: %w", err)
	}
	var v1 struct {
		Features []struct {
			Properties struct {
				SourceIndex int         `json:"source_index"`
				Codigo      *string     `json:"codigo"`
				Nombre      *string     `json:"nombre"`
				AreaM2      json.Number `json:"area_m2"`
			} `json:"properties"`
		} `json:"features"`
	}
	if err := json.Unmarshal(zonasV1, &v1); err != nil {
		return fmt.Errorf("comparar orden: zonas.geojson: %w", err)
	}
	if len(fc.Features) != len(v1.Features) {
		return fmt.Errorf("comparar orden: %d features en la fuente, %d en zonas.geojson", len(fc.Features), len(v1.Features))
	}
	for i, ft := range fc.Features {
		props := ft.Properties
		codigo := strPtr(props["código"])
		nombre := strPtr(props["Nombre"])
		_, area, _ := numRaw(props["Área"])
		other := v1.Features[i].Properties
		if other.SourceIndex != i {
			return fmt.Errorf("comparar orden: índice %d: zonas.geojson trae source_index %d", i, other.SourceIndex)
		}
		if !strEq(codigo, other.Codigo) || !strEq(nombre, other.Nombre) {
			return fmt.Errorf("comparar orden: índice %d: código o nombre no coinciden con zonas.geojson", i)
		}
		if !areaEq(area, other.AreaM2) {
			return fmt.Errorf("comparar orden: índice %d: área no coincide con zonas.geojson", i)
		}
	}
	return nil
}

func strEq(a, b *string) bool {
	if a == nil || b == nil {
		return a == nil && b == nil
	}
	return *a == *b
}

func areaEq(a *float64, b json.Number) bool {
	if a == nil {
		return b == ""
	}
	bf, err := b.Float64()
	if err != nil {
		return false
	}
	return math.Round(*a*1000) == math.Round(bf*1000)
}

// AplicarSectores completa sector solo donde está NULL, desde
// poligonos_sector_ref (migración 044). No pisa un sector asignado a mano ni
// toca polígonos del editor que coincidan por source_index pero no por
// feature_id (Z-NNNN o PC-NNNN, los prefijos que usan el ETL normal y el de
// lote). Devuelve las filas tocadas.
func AplicarSectores(tx *gorm.DB) (int64, error) {
	res := tx.Exec(`
		UPDATE poligonos_cuadrilla p
		SET sector = r.sector, updated_at = now()
		FROM poligonos_sector_ref r
		WHERE p.source_index = r.source_index
		  AND p.sector IS NULL
		  AND p.feature_id IN (
		    'Z-' || lpad((r.source_index + 1)::text, 4, '0'),
		    'PC-' || lpad((r.source_index + 1)::text, 4, '0'))`)
	return res.RowsAffected, res.Error
}

// SQLSectores arma los INSERT ... unnest(...) para poligonos_sector_ref, uno
// por sector en el orden de SectoresValidos. Se pega una sola vez en la
// migración 044 (make sectores -sql). Los literales solo llevan dígitos y
// comas: son seguros para el separador de sentencias de internal/migrate.
func SQLSectores(a ArchivoSectores) string {
	porSector := map[string][]int{}
	for _, p := range a.Poligonos {
		porSector[p.Sector] = append(porSector[p.Sector], p.SourceIndex)
	}
	var b strings.Builder
	for _, sector := range SectoresValidos {
		indices := porSector[sector]
		if len(indices) == 0 {
			continue
		}
		sort.Ints(indices)
		partes := make([]string, len(indices))
		for i, idx := range indices {
			partes[i] = strconv.Itoa(idx)
		}
		fmt.Fprintf(&b, "INSERT INTO poligonos_sector_ref (source_index, sector)\nSELECT unnest('{%s}'::int[]), '%s'\nON CONFLICT (source_index) DO NOTHING;\n\n",
			strings.Join(partes, ","), sector)
	}
	return b.String()
}
