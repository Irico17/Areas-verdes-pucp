// sectores genera data/v1/zonas_sector.json (y, con -sql, el bloque de
// migración) a partir de jefe_de_grupo.json. Por defecto usa la copia
// versionada y anonimizada data/raw/lote/jefe_de_grupo.json, sin red. Nunca
// escribe, imprime ni loguea nombres reales ni hashes de personas.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"campusverde/api/internal/etl"
)

func main() {
	raw := flag.String("raw", "../../data/raw", "directorio data/raw")
	v1 := flag.String("v1", "../../data/v1", "directorio data/v1")
	salida := flag.String("salida", "", "ruta de salida (por defecto <v1>/zonas_sector.json)")
	imprimirSQL := flag.Bool("sql", false, "imprime por stdout los INSERT de poligonos_sector_ref en vez de escribir el JSON")
	vivo := flag.Bool("vivo", false, "intenta la fuente en línea antes de la copia de data/raw/lote (nunca por defecto)")
	flag.Parse()

	if *salida == "" {
		*salida = filepath.Join(*v1, "zonas_sector.json")
	}

	var client *http.Client
	if *vivo {
		client = &http.Client{}
	}
	body, desde, err := etl.FuenteSectores(*raw, *vivo, client)
	if err != nil {
		log.Fatalf("fuente: %v", err)
	}

	zonasV1, err := os.ReadFile(filepath.Join(*v1, "zonas.geojson"))
	if err != nil {
		log.Fatalf("leer %s: %v", filepath.Join(*v1, "zonas.geojson"), err)
	}
	anonBody, err := etl.AnonimizarJefes(body)
	if err != nil {
		log.Fatalf("anonimizar: %v", err)
	}
	if err := etl.CompararOrden(anonBody, zonasV1); err != nil {
		log.Fatalf("orden: %v", err)
	}

	out, err := etl.SectoresDeJefes(body, desde)
	if err != nil {
		log.Fatalf("agrupar: %v", err)
	}

	if *imprimirSQL {
		fmt.Print(etl.SQLSectores(out))
		return
	}

	if err := escribir(*salida, out); err != nil {
		log.Fatalf("escribir %s: %v", *salida, err)
	}
	fmt.Printf("sectores: fuente=%s filas=%d conteos=%v\n", desde, len(out.Poligonos), out.Conteos)
}

func escribir(path string, out etl.ArchivoSectores) error {
	body, err := json.MarshalIndent(out, "", "  ")
	if err != nil {
		return err
	}
	body = append(body, '\n')
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, body, 0o644)
}
