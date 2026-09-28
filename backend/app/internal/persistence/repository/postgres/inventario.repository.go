// Package postgres implements repository interfaces using PostgreSQL and GORM.
package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"gorm.io/gorm"

	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/contracts"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/dto"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/entities"
	domainErrors "github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/errors"
)

type inventarioRepository struct {
	db *gorm.DB
}

// NewInventarioRepository creates a new PostgreSQL repository for inventario.
func NewInventarioRepository(db *gorm.DB) contracts.IInventarioRepository {
	return &inventarioRepository{db: db}
}

// Index lists the known inventory catalogue and counts of loaded features.
func (r *inventarioRepository) Index(ctx context.Context) (dto.IndiceInventarioDTO, error) {
	out := dto.IndiceInventarioDTO{
		Capas:    entities.CapasConocidas,
		Cargadas: []dto.CapaCountDTO{},
	}
	err := r.db.WithContext(ctx).Raw(`
		SELECT capa, count(*) AS features
		FROM inventario
		GROUP BY capa
		ORDER BY capa`).Scan(&out.Cargadas).Error
	if err != nil {
		return out, err
	}
	if out.Cargadas == nil {
		out.Cargadas = []dto.CapaCountDTO{}
	}
	return out, nil
}

// Capa returns the GeoJSON FeatureCollection for a requested inventory layer.
func (r *inventarioRepository) Capa(ctx context.Context, capa string) (entities.FeatureCollection, error) {
	fc := entities.Collection(capa)
	if !entities.EsCapaConocida(capa) {
		return fc, domainErrors.ErrCapaInventarioDesconocida
	}

	q, args := consultaCapa(capa)
	rows, err := r.db.WithContext(ctx).Raw(q, args...).Rows()
	if err != nil {
		return fc, err
	}
	defer rows.Close()

	for rows.Next() {
		var (
			id, geom                string
			nombre, sub, det, lugar sql.NullString
			foto                    sql.NullString
		)
		if err := rows.Scan(&id, &nombre, &sub, &det, &lugar, &foto, &geom); err != nil {
			return fc, err
		}
		raw := json.RawMessage(geom)
		if !json.Valid(raw) {
			return fc, fmt.Errorf("geometría inválida %s", id)
		}
		props := dto.InventarioPropsDTO{FeatureID: id, Capa: capa}
		if nombre.Valid && nombre.String != "" {
			props.Nombre = nombre.String
		}
		if sub.Valid && sub.String != "" {
			props.Subtipo = sub.String
		}
		if det.Valid && det.String != "" {
			props.Detalle = det.String
		}
		if lugar.Valid && lugar.String != "" {
			props.Lugar = lugar.String
		}
		if foto.Valid && foto.String != "" {
			props.Foto = foto.String
		}
		fc.Features = append(fc.Features, entities.Feature{
			Type:       "Feature",
			ID:         id,
			Geometry:   raw,
			Properties: props,
		})
	}
	return fc, rows.Err()
}

func consultaCapa(capa string) (string, []any) {
	legacy := `
		SELECT feature_id, COALESCE(nombre,''), COALESCE(subtipo,''), COALESCE(detalle,''), COALESCE(lugar,''), COALESCE(foto,''), ST_AsGeoJSON(geom, 7)
		FROM inventario WHERE capa = $1 AND geom IS NOT NULL`
	switch capa {
	case "tachos":
		return `SELECT codigo, codigo, '', COALESCE(recomendaciones,''), COALESCE(lugar,''), COALESCE(foto,''), ST_AsGeoJSON(geom, 7)
			FROM tachos WHERE activo AND geom IS NOT NULL
			UNION ALL
			SELECT feature_id, COALESCE(nombre,''), COALESCE(subtipo,''), COALESCE(detalle,''), COALESCE(lugar,''), COALESCE(foto,''), ST_AsGeoJSON(geom, 7)
			FROM inventario WHERE capa = 'tachos' AND geom IS NOT NULL
			AND NOT EXISTS (SELECT 1 FROM tachos WHERE geom IS NOT NULL)`, nil
	case "bebederos":
		return `SELECT codigo, codigo, subtipo, estado, COALESCE(sede,''), COALESCE(foto,''), ST_AsGeoJSON(geom, 7)
			FROM bebederos WHERE activo AND geom IS NOT NULL
			UNION ALL
			SELECT feature_id, COALESCE(nombre,''), COALESCE(subtipo,''), COALESCE(detalle,''), COALESCE(lugar,''), COALESCE(foto,''), ST_AsGeoJSON(geom, 7)
			FROM inventario WHERE capa = 'bebederos' AND geom IS NOT NULL
			AND NOT EXISTS (SELECT 1 FROM bebederos WHERE geom IS NOT NULL)`, nil
	case "fauna":
		return unionCapa("fauna", "fauna", "COALESCE(nombre,'')", "''", "''", "''")
	case "puertas":
		return unionCapa("puertas", "puertas", "COALESCE(codigo,'')", "''", "''", "COALESCE(codigo,'')")
	case "playas_estacionamiento":
		return unionCapa("playas_estacionamiento", "playas_estacionamiento", "COALESCE(codigo,'')", "''", "''", "''")
	case "area_vereda_peligro":
		return unionCapa("veredas_riesgo", "area_vereda_peligro", "COALESCE(nota,'')", "''", "COALESCE(nota,'')", "''")
	case "xerofitica":
		return `SELECT feature_id, COALESCE(clase,''), COALESCE(riego,''), '', '', '', ST_AsGeoJSON(geom, 7)
			FROM xerofiticas WHERE activo AND geom IS NOT NULL`, nil
	case "jardines_reserva":
		return `SELECT feature_id, COALESCE(nombre,''), COALESCE(pertenecen,''), COALESCE(referencia,''), COALESCE(codigo,''), '', ST_AsGeoJSON(geom, 7)
			FROM jardines_reserva WHERE activo AND geom IS NOT NULL`, nil
	case "puntos_pucp":
		return `SELECT origen_ref, titulo, '', '', '', '', ST_AsGeoJSON(geom, 7)
			FROM puntos_pucp WHERE activo AND geom IS NOT NULL`, nil
	default:
		return legacy, []any{capa}
	}
}

func unionCapa(tabla, capaLegacy, nombre, subtipo, detalle, lugar string) (string, []any) {
	q := fmt.Sprintf(`
		SELECT feature_id, %s, %s, %s, %s, '', ST_AsGeoJSON(geom, 7)
		FROM %s WHERE activo AND geom IS NOT NULL
		UNION ALL
		SELECT feature_id, COALESCE(nombre,''), COALESCE(subtipo,''), COALESCE(detalle,''), COALESCE(lugar,''), COALESCE(foto,''), ST_AsGeoJSON(geom, 7)
		FROM inventario WHERE capa = '%s' AND geom IS NOT NULL
		AND NOT EXISTS (SELECT 1 FROM %s WHERE geom IS NOT NULL)`,
		nombre, subtipo, detalle, lugar, tabla, capaLegacy, tabla)
	return q, nil
}
