// Package postgres provides PostgreSQL implementations of persistence contracts.
package postgres

import (
	"context"

	"gorm.io/gorm"

	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/contracts"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/entities"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/persistence/mapper"
)

type cambioRepository struct {
	db *gorm.DB
}

// NewCambioRepository creates a new cambio repository.
func NewCambioRepository(db *gorm.DB) contracts.ICambioRepository {
	return &cambioRepository{db: db}
}

// Crear persists an audit record into the cambios table.
func (r *cambioRepository) Crear(ctx context.Context, c *entities.Cambio) error {
	m := mapper.CambioEntityToModel(c)
	if m == nil {
		return nil
	}
	var antesVal, despuesVal any
	if m.Antes != nil {
		antesVal = *m.Antes
	}
	if m.Despues != nil {
		despuesVal = *m.Despues
	}

	return r.db.WithContext(ctx).Exec(`
		INSERT INTO cambios (entidad, entidad_id, accion, antes, despues, usuario_id, lote_id, created_at)
		VALUES ($1, $2, $3, $4::jsonb, $5::jsonb, $6, $7, COALESCE(NULLIF($8, '0001-01-01 00:00:00+00'::timestamptz), now()))`,
		m.Entidad, m.EntidadID, m.Accion, antesVal, despuesVal, m.UsuarioID, m.LoteID, m.CreatedAt,
	).Error
}
