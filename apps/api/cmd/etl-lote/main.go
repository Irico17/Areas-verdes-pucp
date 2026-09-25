package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"campusverde/api/internal/config"
	"campusverde/api/internal/db"
	"campusverde/api/internal/etl"
)

func main() {
	soloLectura := flag.Bool("solo-lectura", false, "resuelve fuentes y cuenta filas, sin escribir en PostGIS")
	flag.Parse()

	cfg := config.Load()
	fuentes, err := etl.ResolverFuentes(cfg.RawDir, nil)
	if err != nil {
		log.Fatalf("fuentes: %v", err)
	}
	if *soloLectura {
		fmt.Printf("origen=%v\n", fuentes.Origen)
		fmt.Println("copia en", cfg.RawDir+"/lote")
		return
	}
	gdb, err := db.Open(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("base de datos: %v", err)
	}
	rep, err := etl.CargarLote(gdb, fuentes)
	if err != nil {
		log.Fatalf("carga: %v", err)
	}
	areas, err := etl.CargarAreasVerdes(gdb, cfg.RawDir)
	if err != nil {
		log.Fatalf("áreas verdes: %v", err)
	}
	rep.Cargados["areas_verdes"] = areas
	rep.Avisos = append(rep.Avisos, "areas_verdes: upsert por feature_id, sin TRUNCATE. cmd/etl sigue truncando; este camino reconcilia las 521 del catastro.")
	capas, err := etl.CargarFrente2B(gdb, cfg.RawDir)
	if err != nil {
		log.Fatalf("inventario 2B: %v", err)
	}
	for k, v := range capas.Cargados {
		rep.Cargados[k] = v
	}
	rep.Rechazados = append(rep.Rechazados, capas.Rechazados...)
	rep.Avisos = append(rep.Avisos, capas.Avisos...)
	if len(capas.ColumnasOmitidas) > 0 {
		rep.Avisos = append(rep.Avisos, "columnas omitidas por datos personales: "+strings.Join(capas.ColumnasOmitidas, ", "))
	}
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(rep); err != nil {
		log.Fatal(err)
	}
}
