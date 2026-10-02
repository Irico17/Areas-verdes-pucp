// Package postgres provides PostgreSQL implementations of persistence contracts.
package postgres

import (
	"context"
	"encoding/json"

	"gorm.io/gorm"

	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/contracts"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/entities"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/persistence/database"
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

	db := database.DBFromContext(ctx, r.db)
	return db.WithContext(ctx).Exec(`
		INSERT INTO cambios (entidad, entidad_id, accion, antes, despues, usuario_id, lote_id)
		VALUES ($1, $2, $3, $4::jsonb, $5::jsonb, $6, $7)`,
		m.Entidad, m.EntidadID, m.Accion, antesVal, despuesVal, m.UsuarioID, m.LoteID,
	).Error
}

// registrarCambioTx escribe el historial dentro de la misma transacción que la edición.
// usuarioID nulo o no positivo se guarda como NULL (la fila de cambios no exige actor).
func registrarCambioTx(tx *gorm.DB, entidad, entidadID, accion string, antes, despues any, usuarioID *int64) error {
	antesJSON, err := json.Marshal(antes)
	if err != nil {
		return err
	}
	despuesJSON, err := json.Marshal(despues)
	if err != nil {
		return err
	}
	var uid any
	if usuarioID != nil && *usuarioID > 0 {
		uid = *usuarioID
	}
	return tx.Exec(`
		INSERT INTO cambios (entidad, entidad_id, accion, antes, despues, usuario_id)
		VALUES ($1, $2, $3, $4::jsonb, $5::jsonb, $6)`,
		entidad, entidadID, accion, string(antesJSON), string(despuesJSON), uid,
	).Error
}
