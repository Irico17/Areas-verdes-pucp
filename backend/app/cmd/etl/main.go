// Package main runs the ETL data normalization and loading pipeline.
package main

import (
	"context"
	"flag"
	"fmt"
	"log"

	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/cmd/ioc"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/contracts"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/dto"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/infrastructure/etl"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/shared/config"
)

func main() {
	skipLoad := flag.Bool("skip-load", false, "solo normaliza data/raw hacia data/v1")
	noStrict := flag.Bool("no-strict", false, "no exigir los conteos del baseline (521 áreas, 534 zonas)")
	rawDirFlag := flag.String("raw-dir", "", "directorio de datos raw (opcional)")
	v1DirFlag := flag.String("v1-dir", "", "directorio de salida v1 (opcional)")
	flag.Parse()

	cfg := config.GetConfig()
	rawDir := cfg.Datos.RawDir
	if *rawDirFlag != "" {
		rawDir = *rawDirFlag
	}
	v1Dir := cfg.Datos.V1Dir
	if *v1DirFlag != "" {
		v1Dir = *v1DirFlag
	}

	if *skipLoad {
		opt := etl.Options{
			RawDir:   rawDir,
			V1Dir:    v1Dir,
			SkipLoad: true,
			Strict:   !*noStrict,
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
		fmt.Printf("data/v1 escrito en %s\n", v1Dir)
		return
	}

	container, err := ioc.BuildContainer()
	if err != nil {
		log.Fatalf("contenedor: %v", err)
	}

	err = container.Invoke(func(uc contracts.ICargaInicialUseCase) error {
		opt := dto.OpcionesETLDTO{
			RawDir:   rawDir,
			V1Dir:    v1Dir,
			SkipLoad: false,
			Strict:   !*noStrict,
		}
		rep, err := uc.Ejecutar(context.Background(), opt)
		if err != nil {
			return err
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
		fmt.Printf("data/v1 escrito en %s\n", v1Dir)
		return nil
	})
	if err != nil {
		log.Fatal(err)
	}
}
