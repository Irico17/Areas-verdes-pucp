// Package postgres provides PostgreSQL implementations of persistence contracts.
package postgres

import (
	"context"
	"strings"

	"gorm.io/gorm"

	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/contracts"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/entities"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/infrastructure/etl"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/persistence/database"
)

type cargaLoteRepository struct {
	db *gorm.DB
}

// NewCargaLoteRepository creates a new ICargaLoteRepository instance.
func NewCargaLoteRepository(db *gorm.DB) contracts.ICargaLoteRepository {
	return &cargaLoteRepository{db: db}
}

// CargarLote executes upsert operations for batch source files and records changes in auditoria.
func (r *cargaLoteRepository) CargarLote(ctx context.Context, fuentes entities.FuentesLote) (entities.ReporteLote, error) {
	db := database.DBFromContext(ctx, r.db)
	rep, err := etl.CargarLote(db.WithContext(ctx), fuentes)
	if err != nil {
		return entities.ReporteLote{}, err
	}
	rech := make([]entities.Rechazo, len(rep.Rechazados))
	for i, rc := range rep.Rechazados {
		rech[i] = entities.Rechazo{
			Fuente: rc.Fuente,
			Fila:   rc.Fila,
			Campo:  rc.Campo,
			Motivo: rc.Motivo,
		}
	}
	return entities.ReporteLote{
		LoteID:     rep.LoteID,
		Origen:     rep.Origen,
		Cargados:   rep.Cargados,
		Rechazados: rech,
		Avisos:     rep.Avisos,
	}, nil
}

// CargarLoteCompleto coordinates full batch source resolution, load, and layer updates.
func (r *cargaLoteRepository) CargarLoteCompleto(ctx context.Context, rawDir string, soloLectura bool) (*entities.ReporteLote, error) {
	fuentes, err := etl.ResolverFuentes(rawDir, nil)
	if err != nil {
		return nil, err
	}

	if soloLectura {
		return &entities.ReporteLote{
			Origen:   fuentes.Origen,
			Cargados: map[string]int{},
		}, nil
	}

	db := database.DBFromContext(ctx, r.db)
	rep, err := etl.CargarLote(db.WithContext(ctx), fuentes)
	if err != nil {
		return nil, err
	}

	areas, err := etl.CargarAreasVerdes(db.WithContext(ctx), rawDir)
	if err != nil {
		return nil, err
	}
	rep.Cargados["areas_verdes"] = areas
	rep.Avisos = append(rep.Avisos, "areas_verdes: upsert por feature_id, sin TRUNCATE. cmd/etl sigue truncando; este camino reconcilia las 521 del catastro.")

	capas, err := etl.CargarFrente2B(db.WithContext(ctx), rawDir)
	if err != nil {
		return nil, err
	}
	for k, v := range capas.Cargados {
		rep.Cargados[k] = v
	}
	rep.Rechazados = append(rep.Rechazados, capas.Rechazados...)
	rep.Avisos = append(rep.Avisos, capas.Avisos...)
	if len(capas.ColumnasOmitidas) > 0 {
		rep.Avisos = append(rep.Avisos, "columnas omitidas por datos personales: "+strings.Join(capas.ColumnasOmitidas, ", "))
	}

	rech := make([]entities.Rechazo, len(rep.Rechazados))
	for i, rc := range rep.Rechazados {
		rech[i] = entities.Rechazo{
			Fuente: rc.Fuente,
			Fila:   rc.Fila,
			Campo:  rc.Campo,
			Motivo: rc.Motivo,
		}
	}

	return &entities.ReporteLote{
		LoteID:     rep.LoteID,
		Origen:     rep.Origen,
		Cargados:   rep.Cargados,
		Rechazados: rech,
		Avisos:     rep.Avisos,
	}, nil
}

// TextoCargado returns concatenated text columns of the loaded batch for PII verification.
func (r *cargaLoteRepository) TextoCargado(ctx context.Context, loteID int64) (string, error) {
	db := database.DBFromContext(ctx, r.db)
	return etl.TextoCargado(db.WithContext(ctx), loteID)
}
