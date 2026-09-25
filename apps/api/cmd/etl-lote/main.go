package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"

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
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(rep); err != nil {
		log.Fatal(err)
	}
}
