// Package main runs database migrations, seeds initial accounts, and checks ETL readiness.
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"gorm.io/gorm"

	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/cmd/ioc"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/contracts"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/infrastructure/etl"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/persistence/database"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/shared/config"
)

func main() {
	cfg := config.GetConfig()
	if err := cfg.Validar(); err != nil {
		log.Fatalf("configuración: %v", err)
	}

	necesitaETL := flag.Bool("necesita-etl", false, "comprueba si la base necesita carga inicial de ETL (exit 0: vacía, exit 10: con datos)")
	catastroIncompleto := flag.Bool("catastro-incompleto", false, "exit 0 si faltan áreas con geometría o sectores; exit 10 si el catastro publicado ya está")
	semillaFicticia := flag.Bool("semilla-ficticia", false, "aplica los INSERT de la semilla ficticia; ON CONFLICT no pisa filas")
	flag.Parse()

	container, err := ioc.BuildContainer()
	if err != nil {
		log.Fatalf("contenedor: %v", err)
	}

	err = container.Invoke(func(gdb *gorm.DB, cfg *config.Config, semillaUC contracts.ISemillaAccesosUseCase) error {
		if *necesitaETL {
			sqlDB, err := gdb.DB()
			if err != nil {
				log.Fatalf("obtener conexion sql: %v", err)
			}
			necesita, conDatos, err := database.ComprobarNecesitaETL(sqlDB)
			if err != nil {
				log.Fatalf("comprobar datos para ETL: %v", err)
			}
			if !necesita {
				fmt.Printf("BD con datos en: %s\n", strings.Join(conDatos, ", "))
				hasAreas := false
				hasInv := false
				for _, item := range conDatos {
					if strings.HasPrefix(item, "areas_verdes ") {
						hasAreas = true
					}
					if strings.HasPrefix(item, "inventario ") {
						hasInv = true
					}
				}
				if hasAreas && !hasInv {
					log.Println("ADVERTENCIA: areas_verdes contiene datos pero inventario está vacío")
				}
				os.Exit(10)
			}
			fmt.Println("BD vacía: requiere carga inicial de ETL")
			os.Exit(0)
		}

		if *catastroIncompleto {
			sqlDB, err := gdb.DB()
			if err != nil {
				log.Fatalf("obtener conexion sql: %v", err)
			}
			areas, zonas, err := database.ConteosVisibles(sqlDB)
			if err != nil {
				log.Fatalf("conteos del catastro: %v", err)
			}
			fmt.Printf("catastro visible: areas_con_geom=%d zonas_con_sector=%d (mínimo %d/%d)\n",
				areas, zonas, etl.ExpectedAreas, etl.ExpectedZonas)
			if database.CatastroIncompleto(areas, zonas, etl.ExpectedAreas, etl.ExpectedZonas) {
				os.Exit(0)
			}
			os.Exit(10)
		}

		if err := database.Apply(gdb, cfg.Migraciones.Dir); err != nil {
			log.Fatal(err)
		}
		if err := semillaUC.Ensure(context.Background(), cfg.Accesos.DevPassword); err != nil {
			log.Fatalf("cuentas locales: %v", err)
		}
		if *semillaFicticia {
			if err := database.AplicarSemillaFicticia(gdb, cfg.Migraciones.Semilla); err != nil {
				log.Fatalf("semilla ficticia: %v", err)
			}
		}
		log.Println("migraciones al día")
		return nil
	})
	if err != nil {
		log.Fatalf("base de datos: %v", err)
	}
}
