// Package main runs batch loading into PostGIS or inspects batch source status in read-only mode.
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/cmd/ioc"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/contracts"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/dto"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/infrastructure/etl"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/shared/config"
)

func main() {
	soloLectura := flag.Bool("solo-lectura", false, "resuelve fuentes y cuenta filas, sin escribir en PostGIS")
	flag.Parse()

	cfg := config.GetConfig()
	if *soloLectura {
		fuentes, err := etl.ResolverFuentes(cfg.Datos.RawDir, nil)
		if err != nil {
			log.Fatalf("fuentes: %v", err)
		}
		fmt.Printf("origen=%v\n", fuentes.Origen)
		fmt.Println("copia en", cfg.Datos.RawDir+"/lote")
		return
	}

	container, err := ioc.BuildContainer()
	if err != nil {
		log.Fatalf("contenedor: %v", err)
	}

	err = container.Invoke(func(uc contracts.ICargaLoteUseCase) error {
		rep, err := uc.Ejecutar(context.Background(), dto.OpcionesCargaLoteDTO{
			SoloLectura: false,
			RawDir:      cfg.Datos.RawDir,
		})
		if err != nil {
			return err
		}
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(rep)
	})
	if err != nil {
		log.Fatal(err)
	}
}
