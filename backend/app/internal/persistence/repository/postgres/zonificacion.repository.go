package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"strconv"

	"gorm.io/gorm"

	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/contracts"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/dto"
	apperrors "github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/errors"
)

type zonificacionRepository struct {
	db *gorm.DB
}

// NewZonificacionRepository creates the zoning repository.
func NewZonificacionRepository(db *gorm.DB) contracts.IZonificacionRepository {
	return &zonificacionRepository{db: db}
}

func (r *zonificacionRepository) ListarSectores(ctx context.Context, soloActivos bool) ([]dto.SectorCapatazDTO, error) {
	rows, err := r.db.WithContext(ctx).Raw(`
		SELECT id, codigo, nombre, color, activo
		FROM sectores_capataz
		WHERE ($1 = FALSE OR activo = TRUE)
		ORDER BY nombre, codigo`, soloActivos).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []dto.SectorCapatazDTO{}
	for rows.Next() {
		var item dto.SectorCapatazDTO
		if err := rows.Scan(&item.ID, &item.Codigo, &item.Nombre, &item.Color, &item.Activo); err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

func (r *zonificacionRepository) CrearSector(ctx context.Context, in dto.CrearSectorDTO) (dto.SectorCapatazDTO, error) {
	var item dto.SectorCapatazDTO
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var id int64
		err := tx.Raw(`SELECT id FROM sectores_capataz WHERE codigo = $1`, in.Codigo).Row().Scan(&id)
		if err == nil {
			return apperrors.ErrSectorDuplicado
		}
		if !errors.Is(err, sql.ErrNoRows) {
			return err
		}
		if err := tx.Raw(`
			INSERT INTO sectores_capataz (codigo, nombre, color, activo)
			VALUES ($1, $2, $3, TRUE)
			RETURNING id, codigo, nombre, color, activo`,
			in.Codigo, in.Nombre, in.Color,
		).Row().Scan(&item.ID, &item.Codigo, &item.Nombre, &item.Color, &item.Activo); err != nil {
			return err
		}
		return registrarCambioTx(tx, "sectores_capataz", itoa(item.ID), "alta", nil, snapshotSector(item), usuarioOpcional(in.UsuarioID))
	})
	return item, err
}

func (r *zonificacionRepository) ActualizarSector(ctx context.Context, in dto.ActualizarSectorDTO) (dto.SectorCapatazDTO, error) {
	var item dto.SectorCapatazDTO
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		prev, ok, err := leerSector(tx, in.Codigo)
		if err != nil {
			return err
		}
		if !ok {
			return apperrors.ErrSectorNoExiste
		}
		activo := prev.Activo
		if in.Activo != nil {
			activo = *in.Activo
		}
		if prev.Nombre == in.Nombre && prev.Color == in.Color && prev.Activo == activo {
			item = prev
			return nil
		}
		if err := tx.Raw(`
			UPDATE sectores_capataz
			SET nombre = $2, color = $3, activo = $4, updated_at = now()
			WHERE codigo = $1
			RETURNING id, codigo, nombre, color, activo`,
			in.Codigo, in.Nombre, in.Color, activo,
		).Row().Scan(&item.ID, &item.Codigo, &item.Nombre, &item.Color, &item.Activo); err != nil {
			return err
		}
		accion := "edicion"
		if prev.Activo && !item.Activo {
			accion = "baja"
		}
		return registrarCambioTx(tx, "sectores_capataz", itoa(item.ID), accion, snapshotSector(prev), snapshotSector(item), usuarioOpcional(in.UsuarioID))
	})
	return item, err
}

func (r *zonificacionRepository) DesactivarSector(ctx context.Context, codigo string, usuarioID int64) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		prev, ok, err := leerSector(tx, codigo)
		if err != nil {
			return err
		}
		if !ok {
			return apperrors.ErrSectorNoExiste
		}
		if !prev.Activo {
			return nil
		}
		if err := tx.Exec(`
			UPDATE sectores_capataz SET activo = FALSE, updated_at = now() WHERE id = $1`, prev.ID).Error; err != nil {
			return err
		}
		despues := snapshotSector(prev)
		despues["activo"] = false
		return registrarCambioTx(tx, "sectores_capataz", itoa(prev.ID), "baja", snapshotSector(prev), despues, usuarioOpcional(usuarioID))
	})
}

func (r *zonificacionRepository) ImportarSectores(ctx context.Context, filas []dto.CrearSectorDTO, usuarioID int64) (dto.ImportacionSectorDTO, error) {
	var res dto.ImportacionSectorDTO
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, fila := range filas {
			prev, ok, err := leerSector(tx, fila.Codigo)
			if err != nil {
				return err
			}
			if !ok {
				var id int64
				if err := tx.Raw(`
					INSERT INTO sectores_capataz (codigo, nombre, color, activo)
					VALUES ($1, $2, $3, TRUE)
					RETURNING id`, fila.Codigo, fila.Nombre, fila.Color).Row().Scan(&id); err != nil {
					return err
				}
				item := dto.SectorCapatazDTO{ID: id, Codigo: fila.Codigo, Nombre: fila.Nombre, Color: fila.Color, Activo: true}
				if err := registrarCambioTx(tx, "sectores_capataz", itoa(id), "importacion", nil, snapshotSector(item), usuarioOpcional(usuarioID)); err != nil {
					return err
				}
				res.Creados++
				continue
			}
			if prev.Nombre == fila.Nombre && prev.Color == fila.Color && prev.Activo {
				continue
			}
			var item dto.SectorCapatazDTO
			if err := tx.Raw(`
				UPDATE sectores_capataz
				SET nombre = $2, color = $3, activo = TRUE, updated_at = now()
				WHERE codigo = $1
				RETURNING id, codigo, nombre, color, activo`,
				fila.Codigo, fila.Nombre, fila.Color,
			).Row().Scan(&item.ID, &item.Codigo, &item.Nombre, &item.Color, &item.Activo); err != nil {
				return err
			}
			if err := registrarCambioTx(tx, "sectores_capataz", itoa(item.ID), "importacion", snapshotSector(prev), snapshotSector(item), usuarioOpcional(usuarioID)); err != nil {
				return err
			}
			res.Actualizados++
		}
		return nil
	})
	return res, err
}

func (r *zonificacionRepository) ListarLugares(ctx context.Context) ([]dto.LugarCatalogoDTO, error) {
	rows, err := r.db.WithContext(ctx).Raw(`
		SELECT id, nombre FROM lugares WHERE activo = TRUE ORDER BY nombre_norm`).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []dto.LugarCatalogoDTO{}
	for rows.Next() {
		var item dto.LugarCatalogoDTO
		if err := rows.Scan(&item.ID, &item.Nombre); err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

func (r *zonificacionRepository) LugarPorID(ctx context.Context, id int64) (dto.LugarCatalogoDTO, bool, error) {
	var item dto.LugarCatalogoDTO
	err := r.db.WithContext(ctx).Raw(`
		SELECT id, nombre FROM lugares WHERE id = $1 AND activo = TRUE`, id).Row().Scan(&item.ID, &item.Nombre)
	if errors.Is(err, sql.ErrNoRows) {
		return dto.LugarCatalogoDTO{}, false, nil
	}
	if err != nil {
		return dto.LugarCatalogoDTO{}, false, err
	}
	return item, true, nil
}

func (r *zonificacionRepository) LugarPorNorm(ctx context.Context, nombreNorm string) (dto.LugarCatalogoDTO, bool, error) {
	var item dto.LugarCatalogoDTO
	err := r.db.WithContext(ctx).Raw(`
		SELECT id, nombre FROM lugares WHERE nombre_norm = $1 AND activo = TRUE`, nombreNorm).Row().Scan(&item.ID, &item.Nombre)
	if errors.Is(err, sql.ErrNoRows) {
		return dto.LugarCatalogoDTO{}, false, nil
	}
	if err != nil {
		return dto.LugarCatalogoDTO{}, false, err
	}
	return item, true, nil
}

func (r *zonificacionRepository) ContarLugares(ctx context.Context) (int64, error) {
	var n int64
	err := r.db.WithContext(ctx).Raw(`SELECT count(*) FROM lugares`).Row().Scan(&n)
	return n, err
}

func (r *zonificacionRepository) ViasGeoJSON(ctx context.Context) ([]byte, error) {
	var body string
	err := r.db.WithContext(ctx).Raw(`
		SELECT jsonb_build_object(
			'type', 'FeatureCollection',
			'name', 'vias',
			'features', COALESCE((
				SELECT jsonb_agg(jsonb_build_object(
					'type', 'Feature',
					'id', feature_id,
					'geometry', ST_AsGeoJSON(geom)::jsonb,
					'properties', jsonb_build_object('feature_id', feature_id, 'nombre', nombre)
				) ORDER BY id)
				FROM vias
				WHERE activo = TRUE AND geom IS NOT NULL
			), '[]'::jsonb)
		)::text`).Row().Scan(&body)
	if err != nil {
		return nil, err
	}
	return []byte(body), nil
}

func (r *zonificacionRepository) ImportarVias(ctx context.Context, filas []dto.ViaAltaDTO, usuarioID int64) (dto.ImportacionViaDTO, error) {
	var res dto.ImportacionViaDTO
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, fila := range filas {
			var id int64
			var nombre sql.NullString
			err := tx.Raw(`SELECT id, nombre FROM vias WHERE feature_id = $1`, fila.FeatureID).Row().Scan(&id, &nombre)
			if errors.Is(err, sql.ErrNoRows) {
				if err := tx.Raw(`
					INSERT INTO vias (feature_id, nombre, geom, origen_ref, activo)
					VALUES (
						$1, NULLIF($2, ''),
						ST_Multi(ST_Force2D(ST_SetSRID(ST_GeomFromGeoJSON($3), 4326))),
						$1, TRUE
					)
					RETURNING id`, fila.FeatureID, fila.Nombre, fila.GeoJSON).Row().Scan(&id); err != nil {
					return err
				}
				if err := registrarCambioTx(tx, "vias", itoa(id), "importacion", nil, map[string]any{
					"feature_id": fila.FeatureID,
					"nombre":     fila.Nombre,
				}, usuarioOpcional(usuarioID)); err != nil {
					return err
				}
				res.Creadas++
				continue
			}
			if err != nil {
				return err
			}
			if err := tx.Exec(`
				UPDATE vias
				SET nombre = NULLIF($2, ''),
				    geom = ST_Multi(ST_Force2D(ST_SetSRID(ST_GeomFromGeoJSON($3), 4326))),
				    activo = TRUE,
				    updated_at = now()
				WHERE id = $1`, id, fila.Nombre, fila.GeoJSON).Error; err != nil {
				return err
			}
			if err := registrarCambioTx(tx, "vias", itoa(id), "importacion", map[string]any{
				"feature_id": fila.FeatureID,
				"nombre":     nombre.String,
			}, map[string]any{
				"feature_id": fila.FeatureID,
				"nombre":     fila.Nombre,
			}, usuarioOpcional(usuarioID)); err != nil {
				return err
			}
			res.Actualizadas++
		}
		return nil
	})
	return res, err
}

func (r *zonificacionRepository) CuartelesGeoJSON(ctx context.Context) ([]byte, error) {
	rows, err := r.db.WithContext(ctx).Raw(`
		SELECT codigo, nombre, ST_AsGeoJSON(geom)
		FROM cuarteles_historico
		WHERE geom IS NOT NULL
		ORDER BY codigo`).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	type feature struct {
		Type       string          `json:"type"`
		ID         string          `json:"id"`
		Geometry   json.RawMessage `json:"geometry"`
		Properties map[string]any  `json:"properties"`
	}
	feats := []feature{}
	for rows.Next() {
		var codigo, nombre, geom string
		if err := rows.Scan(&codigo, &nombre, &geom); err != nil {
			return nil, err
		}
		if geom == "" || geom == "null" {
			continue
		}
		feats = append(feats, feature{
			Type:     "Feature",
			ID:       codigo,
			Geometry: json.RawMessage(geom),
			Properties: map[string]any{
				"codigo": codigo,
				"nombre": nombre,
			},
		})
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	payload := map[string]any{
		"type":     "FeatureCollection",
		"name":     "cuarteles",
		"features": feats,
	}
	if len(feats) == 0 {
		payload["aviso"] = "sin archivo de cuarteles"
	}
	return json.Marshal(payload)
}

func (r *zonificacionRepository) CrearReferente(ctx context.Context, in dto.CrearReferenteDTO) (dto.ReferenteDTO, error) {
	var item dto.ReferenteDTO
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		err := tx.Raw(`
			SELECT id, lugar_id, edificio_id
			FROM referentes_edificio
			WHERE lugar_id = $1 AND edificio_id = $2`, in.LugarID, in.EdificioID,
		).Row().Scan(&item.ID, &item.LugarID, &item.EdificioID)
		if err == nil {
			item.Creado = false
			return nil
		}
		if !errors.Is(err, sql.ErrNoRows) {
			return err
		}
		if err := tx.Raw(`
			INSERT INTO referentes_edificio (lugar_id, edificio_id)
			VALUES ($1, $2)
			RETURNING id, lugar_id, edificio_id`, in.LugarID, in.EdificioID,
		).Row().Scan(&item.ID, &item.LugarID, &item.EdificioID); err != nil {
			return err
		}
		item.Creado = true
		return registrarCambioTx(tx, "referentes_edificio", itoa(item.ID), "alta", nil, map[string]any{
			"lugar_id":    item.LugarID,
			"edificio_id": item.EdificioID,
		}, usuarioOpcional(in.UsuarioID))
	})
	return item, err
}

func leerSector(tx *gorm.DB, codigo string) (dto.SectorCapatazDTO, bool, error) {
	var item dto.SectorCapatazDTO
	err := tx.Raw(`
		SELECT id, codigo, nombre, color, activo
		FROM sectores_capataz WHERE codigo = $1`, codigo).Row().Scan(
		&item.ID, &item.Codigo, &item.Nombre, &item.Color, &item.Activo)
	if errors.Is(err, sql.ErrNoRows) {
		return dto.SectorCapatazDTO{}, false, nil
	}
	if err != nil {
		return dto.SectorCapatazDTO{}, false, err
	}
	return item, true, nil
}

func snapshotSector(item dto.SectorCapatazDTO) map[string]any {
	return map[string]any{
		"codigo": item.Codigo,
		"nombre": item.Nombre,
		"color":  item.Color,
		"activo": item.Activo,
	}
}

func itoa(id int64) string {
	return strconv.FormatInt(id, 10)
}
