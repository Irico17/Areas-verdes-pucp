package main

import (
	"log"

	"campusverde/api/internal/accesos"
	"campusverde/api/internal/config"
	"campusverde/api/internal/db"
	"campusverde/api/internal/migrate"
)

func main() {
	cfg := config.Load()
	gdb, err := db.Open(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("base de datos: %v", err)
	}
	if err := migrate.Apply(gdb, cfg.MigrationsDir); err != nil {
		log.Fatal(err)
	}
	if err := accesos.Ensure(gdb, cfg.DevPassword); err != nil {
		log.Fatalf("cuentas locales: %v", err)
	}
	log.Println("migraciones al día")
}
