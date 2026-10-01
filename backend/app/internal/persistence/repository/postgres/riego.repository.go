// Package postgres implements repository interfaces using GORM and PostgreSQL.
package postgres

import (
	"context"
	"database/sql"
	"strings"

	"gorm.io/gorm"

	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/contracts"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/usecases"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/entities"
	domainErrors "github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/errors"
)

type riegoRepository struct {
	db *gorm.DB
}

// NewRiegoRepository creates a new IRiegoRepository.
func NewRiegoRepository(db *gorm.DB) contracts.IRiegoRepository {
	return &riegoRepository{db: db}
}

func (r *riegoRepository) Listar(ctx context.Context, capatazID string) ([]*entities.TurnoRiego, error) {
	q, args := usecases.ConsultaRiego(capatazID)
	rows, err := r.db.WithContext(ctx).Raw(q, args...).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []*entities.TurnoRiego{}
	for rows.Next() {
		var item entities.TurnoRiego
		var capID, equipo, zonaCode, ciclo sql.NullString
		if err := rows.Scan(
			&item.ID, &item.Sector, &item.Turno, &capID, &equipo,
			&item.Fecha, &item.Nota, &zonaCode, &ciclo,
		); err != nil {
			return nil, err
		}
		item.CapatazID = nullStringPtr(capID)
		item.Equipo = nullStringPtr(equipo)
		item.ZonaSupervisionCode = nullStringPtr(zonaCode)
		if ciclo.Valid {
			item.Ciclo = ciclo.String
		}
		out = append(out, &item)
	}
	return out, rows.Err()
}

func (r *riegoRepository) Crear(ctx context.Context, in entities.NuevoTurnoRiego) error {
	var zonaNum int64
	if err := r.db.WithContext(ctx).Raw(`SELECT COALESCE((SELECT id FROM zonas_supervision WHERE codigo = $1), 0)`, in.ZonaID).Scan(&zonaNum).Error; err != nil {
		return err
	}
	if zonaNum == 0 {
		return domainErrors.InputError{Reason: "la zona de supervisión no existe"}
	}

	err := r.db.WithContext(ctx).Exec(`
		INSERT INTO riego_registros (id, sector, turno, capataz_id, fecha, nota, zona_supervision_id, ciclo, superficie_m2)
		VALUES ($1, $2, $3, NULLIF($4, ''), $5::date, $6, $7, $8, NULLIF($9, 0))`,
		in.ID, in.Sector, in.Turno, strings.TrimSpace(in.CapatazID), in.Fecha, strings.TrimSpace(in.Nota), zonaNum, strings.TrimSpace(in.Ciclo), in.Superficie,
	).Error
	if err != nil && (strings.Contains(err.Error(), "duplicate") || strings.Contains(err.Error(), "unique")) {
		return domainErrors.InputError{Reason: "ya hay un riego de esa zona, turno y fecha"}
	}
	return err
}
