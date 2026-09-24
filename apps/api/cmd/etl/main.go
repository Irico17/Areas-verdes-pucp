package main

import (
	"flag"
	"fmt"
	"log"

	"campusverde/api/internal/config"
	"campusverde/api/internal/db"
	"campusverde/api/internal/etl"
)

func main() {
	skipLoad := flag.Bool("skip-load", false, "solo normaliza data/raw hacia data/v1")
	noStrict := flag.Bool("no-strict", false, "no exigir los conteos del baseline (521 áreas, 534 zonas)")
	flag.Parse()

	cfg := config.Load()
	opt := etl.Options{
		RawDir:   cfg.RawDir,
		V1Dir:    cfg.V1Dir,
		SkipLoad: *skipLoad,
		Strict:   !*noStrict,
	}
	if !*skipLoad {
		gdb, err := db.Open(cfg.DatabaseURL)
		if err != nil {
			log.Fatalf("base de datos: %v", err)
		}
		opt.DB = gdb
	}

	rep, err := etl.Run(opt)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("ETL_COUNTS areas=%d zonas=%d", rep.Areas, rep.Zonas)
	for name, n := range rep.Capas {
		fmt.Printf(" %s=%d", name, n)
	}
	fmt.Println()
	fmt.Printf("ETL_INVENTARIO")
	for name, n := range rep.Inventario {
		fmt.Printf(" %s=%d", name, n)
	}
	fmt.Println()
	fmt.Printf("data/v1 escrito en %s\n", cfg.V1Dir)
}
