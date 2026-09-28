package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"campusverde/api/internal/accesos"
	"campusverde/api/internal/config"
	"campusverde/api/internal/db"
	"campusverde/api/internal/migrate"
)

func main() {
	necesitaETL := flag.Bool("necesita-etl", false, "comprueba si la base necesita carga inicial de ETL (exit 0: vacía, exit 10: con datos)")
	flag.Parse()

	cfg := config.Load()
	gdb, err := db.Open(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("base de datos: %v", err)
	}

	if *necesitaETL {
		sqlDB, err := gdb.DB()
		if err != nil {
			log.Fatalf("obtener conexion sql: %v", err)
		}
		necesita, conDatos, err := migrate.ComprobarNecesitaETL(sqlDB)
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

	if err := migrate.Apply(gdb, cfg.MigrationsDir); err != nil {
		log.Fatal(err)
	}
	if err := accesos.Ensure(gdb, cfg.DevPassword); err != nil {
		log.Fatalf("cuentas locales: %v", err)
	}
	log.Println("migraciones al día")
}
