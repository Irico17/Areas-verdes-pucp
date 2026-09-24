package main

import (
	"log"
	"path/filepath"

	"campusverde/api/internal/accesos"
	"campusverde/api/internal/config"
	"campusverde/api/internal/db"
	"campusverde/api/internal/server"
)

func main() {
	cfg := config.Load()

	gdb, err := db.Open(cfg.DatabaseURL)
	if err != nil {
		log.Printf("aviso: sin base de datos (%v); /health reportará database=down", err)
		gdb = nil
	}

	if gdb != nil {
		if err := accesos.Ensure(gdb, cfg.DevPassword); err != nil {
			log.Fatalf("cuentas locales: %v", err)
		}
	}

	engine := server.New(server.Deps{
		DB:               gdb,
		OpenAPIPath:      cfg.OpenAPIPath,
		EdificiosPath:    cfg.EdificiosPath,
		ReservasPath:     cfg.ReservasPath,
		FotosDir:         filepath.Join(cfg.RawDir, "drive_fotos"),
		EvidenciasDir:    cfg.EvidenciasDir,
		EvidenciasBucket: cfg.EvidenciasBucket,
	})
	log.Printf("campus-verde-api escuchando en %s", cfg.APIAddr)
	if err := engine.Run(cfg.APIAddr); err != nil {
		log.Fatal(err)
	}
}
