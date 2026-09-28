package postgres

import (
	"context"

	"gorm.io/gorm"

	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/contracts"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/entities"
	domainErrors "github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/errors"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/persistence/mapper"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/persistence/models"
)

type catastroReferenciaRepository struct {
	db *gorm.DB
}

// NewCatastroReferenciaRepository creates a new ICatastroReferenciaRepository instance.
func NewCatastroReferenciaRepository(db *gorm.DB) contracts.ICatastroReferenciaRepository {
	return &catastroReferenciaRepository{db: db}
}

func (r *catastroReferenciaRepository) ListarPoligonos(ctx context.Context) ([]entities.PoligonoCuadrilla, error) {
	rows, err := r.db.WithContext(ctx).Raw(`
		SELECT id, feature_id, source_index, codigo, nombre, cuadrilla_id,
		       zona_supervision_id, activo, geom IS NOT NULL
		FROM poligonos_cuadrilla
		WHERE activo
		ORDER BY source_index
		LIMIT 100`).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []entities.PoligonoCuadrilla{}
	for rows.Next() {
		var (
			m       models.PoligonoCuadrillaModel
			conGeom bool
		)
		if err := rows.Scan(&m.ID, &m.FeatureID, &m.SourceIndex, &m.Codigo, &m.Nombre, &m.CuadrillaID, &m.ZonaSupervisionID, &m.Activo, &conGeom); err != nil {
			return nil, err
		}
		out = append(out, *mapper.PoligonoCuadrillaModelToEntity(&m, conGeom))
	}
	return out, rows.Err()
}

func (r *catastroReferenciaRepository) ListarCapa(ctx context.Context, tabla string) ([]entities.CapaFicha, error) {
	q, ok := capaQuery(tabla)
	if !ok {
		return nil, domainErrors.ErrEntrada
	}
	rows, err := r.db.WithContext(ctx).Raw(q).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []entities.CapaFicha{}
	for rows.Next() {
		var m models.CapaAuxiliarModel
		if err := rows.Scan(&m.ID, &m.FeatureID, &m.Nombre, &m.Codigo, &m.Referencia, &m.Clase, &m.RiegoAct, &m.Pertenecen, &m.Capa); err != nil {
			return nil, err
		}
		out = append(out, *mapper.CapaAuxiliarModelToEntity(&m))
	}
	return out, rows.Err()
}

func capaQuery(tabla string) (string, bool) {
	switch tabla {
	case "fauna":
		return `SELECT id, feature_id, COALESCE(nombre, ''), '', '', '', '', '', activo FROM fauna WHERE activo ORDER BY feature_id`, true
	case "puertas":
		return `SELECT id, feature_id, COALESCE(nombre, ''), COALESCE(codigo, ''), '', '', '', '', activo FROM puertas WHERE activo ORDER BY feature_id`, true
	case "playas_estacionamiento":
		return `SELECT id, feature_id, '', COALESCE(codigo, ''), '', '', '', '', activo FROM playas_estacionamiento WHERE activo ORDER BY feature_id`, true
	case "veredas_riesgo":
		return `SELECT id, feature_id, '', '', COALESCE(nota, ''), '', '', '', activo FROM veredas_riesgo WHERE activo ORDER BY feature_id`, true
	case "xerofiticas":
		return `SELECT id, feature_id, '', '', '', COALESCE(clase, ''), COALESCE(riego, ''), '', activo FROM xerofiticas WHERE activo ORDER BY feature_id`, true
	case "jardines_reserva":
		return `SELECT id, feature_id, COALESCE(nombre, ''), COALESCE(codigo, ''), '', '', COALESCE(riego_act, ''), COALESCE(pertenecen, ''), activo FROM jardines_reserva WHERE activo ORDER BY feature_id`, true
	default:
		return "", false
	}
}
