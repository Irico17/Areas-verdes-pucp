// Package postgres provides PostgreSQL implementations of persistence contracts.
package postgres

import (
	"context"

	"gorm.io/gorm"

	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/contracts"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/infrastructure/etl"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/persistence/database"
)

type cargaWriteRepository struct {
	db *gorm.DB
}

// NewCargaWriteRepository creates a new ICargaWriteRepository instance.
func NewCargaWriteRepository(db *gorm.DB) contracts.ICargaWriteRepository {
	return &cargaWriteRepository{db: db}
}

// GuardarVistaPrevia records an import batch in 'vista_previa' state.
func (r *cargaWriteRepository) GuardarVistaPrevia(ctx context.Context, entidad string, usuarioID int64, validas int, contenido []byte, nombreArchivo string) (int64, error) {
	db := database.DBFromContext(ctx, r.db)
	var id int64
	err := db.WithContext(ctx).Raw(`
		INSERT INTO lotes_importacion (entidad, estado, usuario_id, filas, contenido, nombre_archivo)
		VALUES ($1, 'vista_previa', $2, $3, $4, $5)
		RETURNING id`, entidad, usuarioID, validas, contenido, nombreArchivo).Row().Scan(&id)
	return id, err
}

// Confirmar executes the confirmation and persistence of an import preview batch.
func (r *cargaWriteRepository) Confirmar(ctx context.Context, loteID, usuarioID int64) (int64, int, error) {
	db := database.DBFromContext(ctx, r.db)
	return etl.Confirmar(db.WithContext(ctx), loteID, usuarioID)
}
