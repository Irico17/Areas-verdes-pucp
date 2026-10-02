// Package postgres provides PostgreSQL implementations of persistence contracts.
package postgres

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"

	"gorm.io/gorm"

	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/contracts"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/entities"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/infrastructure/etl"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/persistence/database"
)

type cargaSectorRepository struct {
	db *gorm.DB
}

// NewCargaSectorRepository creates a new ICargaSectorRepository instance.
func NewCargaSectorRepository(db *gorm.DB) contracts.ICargaSectorRepository {
	return &cargaSectorRepository{db: db}
}

// AplicarSectores populates the sector column for polygons from reference data where NULL.
func (r *cargaSectorRepository) AplicarSectores(ctx context.Context) (int64, error) {
	db := database.DBFromContext(ctx, r.db)
	return etl.AplicarSectores(db.WithContext(ctx))
}

// GenerarSectores generates sector groupings from jefe_de_grupo.json and writes JSON or returns SQL.
func (r *cargaSectorRepository) GenerarSectores(ctx context.Context, rawDir, v1Dir, salida string, imprimirSQL, vivo bool) (*entities.ArchivoSectores, string, error) {
	body, desde, err := etl.FuenteSectores(rawDir, vivo, nil)
	if err != nil {
		return nil, "", err
	}

	zonasV1, err := os.ReadFile(filepath.Join(v1Dir, "zonas.geojson"))
	if err != nil {
		return nil, "", err
	}
	anonBody, err := etl.AnonimizarJefes(body)
	if err != nil {
		return nil, "", err
	}
	if err := etl.CompararOrden(anonBody, zonasV1); err != nil {
		return nil, "", err
	}

	out, err := etl.SectoresDeJefes(body, desde)
	if err != nil {
		return nil, "", err
	}

	if imprimirSQL {
		sqlStr := etl.SQLSectores(out)
		return &out, sqlStr, nil
	}

	salidaRuta := salida
	if salidaRuta == "" {
		salidaRuta = filepath.Join(v1Dir, "zonas_sector.json")
	}

	rawOut, err := json.MarshalIndent(out, "", "  ")
	if err != nil {
		return nil, "", err
	}
	rawOut = append(rawOut, '\n')
	if err := os.MkdirAll(filepath.Dir(salidaRuta), 0o755); err != nil {
		return nil, "", err
	}
	if err := os.WriteFile(salidaRuta, rawOut, 0o644); err != nil {
		return nil, "", err
	}

	return &out, "", nil
}
