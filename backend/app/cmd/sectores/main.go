// Package main generates data/v1/zonas_sector.json from jefe_de_grupo.json or prints SQL inserts.
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"path/filepath"

	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/cmd/ioc"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/contracts"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/dto"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/shared/config"
)

func main() {
	cfg := config.GetConfig()
	defaultRaw := cfg.Datos.RawDir
	if defaultRaw == "" {
		defaultRaw = "../../data/raw"
	}
	defaultV1 := cfg.Datos.V1Dir
	if defaultV1 == "" {
		defaultV1 = "../../data/v1"
	}

	raw := flag.String("raw", defaultRaw, "directorio data/raw")
	v1 := flag.String("v1", defaultV1, "directorio data/v1")
	salida := flag.String("salida", "", "ruta de salida (por defecto <v1>/zonas_sector.json)")
	imprimirSQL := flag.Bool("sql", false, "imprime por stdout los INSERT de poligonos_sector_ref en vez de escribir el JSON")
	vivo := flag.Bool("vivo", false, "intenta la fuente en línea antes de la copia de data/raw/lote (nunca por defecto)")
	flag.Parse()

	salidaRuta := *salida
	if salidaRuta == "" {
		salidaRuta = filepath.Join(*v1, "zonas_sector.json")
	}

	container, err := ioc.BuildContainer()
	if err != nil {
		log.Fatalf("contenedor: %v", err)
	}

	err = container.Invoke(func(uc contracts.ISectoresUseCase) error {
		opt := dto.OpcionesSectoresDTO{
			RawDir:      *raw,
			V1Dir:       *v1,
			Salida:      salidaRuta,
			ImprimirSQL: *imprimirSQL,
			Vivo:        *vivo,
		}
		res, err := uc.Ejecutar(context.Background(), opt)
		if err != nil {
			return err
		}
		if *imprimirSQL {
			fmt.Print(res.SQL)
			return nil
		}
		fmt.Printf("sectores: fuente=%s filas=%d conteos=%v\n", res.Desde, res.Filas, res.Conteos)
		return nil
	})
	if err != nil {
		log.Fatal(err)
	}
}
