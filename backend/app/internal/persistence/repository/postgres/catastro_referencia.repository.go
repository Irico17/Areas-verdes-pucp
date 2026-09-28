package postgres

import (
	"context"

	"gorm.io/gorm"

	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/contracts"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/entities"
	domainErrors "github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/errors"
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
		SELECT id, feature_id, COALESCE(codigo, ''), COALESCE(nombre, ''),
		       COALESCE(cuadrilla_id, ''), zona_supervision_id, geom IS NOT NULL, activo
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
		var p entities.PoligonoCuadrilla
		if err := rows.Scan(&p.ID, &p.FeatureID, &p.Codigo, &p.Nombre, &p.CuadrillaID, &p.ZonaSupervisionID, &p.ConGeom, &p.Activo); err != nil {
			return nil, err
		}
		out = append(out, p)
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
		var f entities.CapaFicha
		if err := rows.Scan(&f.ID, &f.FeatureID, &f.Nombre, &f.Codigo, &f.Nota, &f.Clase, &f.Riego, &f.Pertenecen, &f.Activo); err != nil {
			return nil, err
		}
		out = append(out, f)
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
