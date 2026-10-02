package postgres

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"gorm.io/gorm"

	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/contracts"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/entities"
	domainErrors "github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/errors"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/persistence/mapper"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/persistence/models"
)

type zonaSupervisionRepository struct {
	db *gorm.DB
}

// NewZonaSupervisionRepository creates a new IZonaSupervisionRepository instance.
func NewZonaSupervisionRepository(db *gorm.DB) contracts.IZonaSupervisionRepository {
	return &zonaSupervisionRepository{db: db}
}

func (r *zonaSupervisionRepository) Listar(ctx context.Context) ([]entities.ZonaSupervision, error) {
	rows, err := r.db.WithContext(ctx).Raw(`
		SELECT id, codigo, nombre, area_m2, activo, geom IS NOT NULL
		FROM zonas_supervision
		WHERE activo
		ORDER BY codigo`).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []entities.ZonaSupervision{}
	for rows.Next() {
		var (
			m       models.ZonaSupervisionModel
			conGeom bool
		)
		if err := rows.Scan(&m.ID, &m.Codigo, &m.Nombre, &m.AreaM2, &m.Activo, &conGeom); err != nil {
			return nil, err
		}
		out = append(out, *mapper.ZonaSupervisionModelToEntity(&m, conGeom))
	}
	return out, rows.Err()
}

func (r *zonaSupervisionRepository) Crear(ctx context.Context, codigo, nombre, geojson string, area *float64) (entities.ZonaSupervision, error) {
	var id int64
	err := r.db.WithContext(ctx).Raw(`
		INSERT INTO zonas_supervision (codigo, nombre, area_m2, geom)
		VALUES ($1, $2, $3, catastro_geom_4326($4))
		RETURNING id`, codigo, nombre, area, geojson).Scan(&id).Error
	if err != nil {
		return entities.ZonaSupervision{}, domainErrors.ErrEntrada
	}
	list, err := r.Listar(ctx)
	if err != nil {
		return entities.ZonaSupervision{}, err
	}
	for _, z := range list {
		if z.ID == id {
			return z, nil
		}
	}
	return entities.ZonaSupervision{}, domainErrors.ErrNoEncontrado
}

func (r *zonaSupervisionRepository) obtener(ctx context.Context, codigo string) (entities.ZonaSupervision, error) {
	var (
		m       models.ZonaSupervisionModel
		conGeom bool
	)
	err := r.db.WithContext(ctx).Raw(`
		SELECT id, codigo, nombre, area_m2, activo, geom IS NOT NULL
		FROM zonas_supervision
		WHERE codigo = $1`, codigo).Row().Scan(&m.ID, &m.Codigo, &m.Nombre, &m.AreaM2, &m.Activo, &conGeom)
	if errors.Is(err, sql.ErrNoRows) {
		return entities.ZonaSupervision{}, domainErrors.ErrNoEncontrado
	}
	if err != nil {
		return entities.ZonaSupervision{}, err
	}
	return *mapper.ZonaSupervisionModelToEntity(&m, conGeom), nil
}

// Actualizar edita nombre, área y, si viene geometría, el polígono. El código no cambia.
func (r *zonaSupervisionRepository) Actualizar(ctx context.Context, codigo, nombre, geojson string, area *float64, usuarioID *int64) (entities.ZonaSupervision, error) {
	codigo = strings.TrimSpace(codigo)
	var actualizada entities.ZonaSupervision
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var (
			id         int64
			nombrePrev string
			areaPrev   *float64
			activo     bool
		)
		err := tx.Raw(`
			SELECT id, nombre, area_m2, activo
			FROM zonas_supervision
			WHERE codigo = $1`, codigo).Row().Scan(&id, &nombrePrev, &areaPrev, &activo)
		if errors.Is(err, sql.ErrNoRows) || (err == nil && !activo) {
			return domainErrors.ErrNoEncontrado
		}
		if err != nil {
			return err
		}
		res := tx.Exec(`
			UPDATE zonas_supervision
			SET nombre = $2,
			    area_m2 = $3,
			    geom = CASE WHEN $4 = '' THEN geom ELSE catastro_geom_4326($4) END,
			    updated_at = now()
			WHERE codigo = $1 AND activo`,
			codigo, nombre, area, geojson,
		)
		if res.Error != nil {
			return domainErrors.ErrEntrada
		}
		if res.RowsAffected != 1 {
			return domainErrors.ErrNoEncontrado
		}
		if err := registrarCambioTx(tx, "zonas_supervision", codigo, "edicion",
			map[string]any{"codigo": codigo, "nombre": nombrePrev, "area_m2": areaPrev, "activo": true},
			map[string]any{"codigo": codigo, "nombre": nombre, "area_m2": area, "activo": true, "geom_actualizada": geojson != ""},
			usuarioID,
		); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return entities.ZonaSupervision{}, err
	}
	actualizada, err = r.obtener(ctx, codigo)
	return actualizada, err
}

// Baja deja la zona inactiva. La fila permanece. Una segunda baja no agrega otro cambio.
func (r *zonaSupervisionRepository) Baja(ctx context.Context, codigo string, usuarioID *int64) (entities.ZonaSupervision, error) {
	codigo = strings.TrimSpace(codigo)
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var (
			id     int64
			nombre string
			activo bool
		)
		err := tx.Raw(`SELECT id, nombre, activo FROM zonas_supervision WHERE codigo = $1`, codigo).Row().Scan(&id, &nombre, &activo)
		if errors.Is(err, sql.ErrNoRows) {
			return domainErrors.ErrNoEncontrado
		}
		if err != nil {
			return err
		}
		if !activo {
			return nil
		}
		res := tx.Exec(`UPDATE zonas_supervision SET activo = FALSE, updated_at = now() WHERE id = $1 AND activo`, id)
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected != 1 {
			return domainErrors.ErrNoEncontrado
		}
		return registrarCambioTx(tx, "zonas_supervision", codigo, "baja",
			map[string]any{"codigo": codigo, "nombre": nombre, "activo": true},
			map[string]any{"codigo": codigo, "nombre": nombre, "activo": false},
			usuarioID,
		)
	})
	if err != nil {
		return entities.ZonaSupervision{}, err
	}
	return r.obtener(ctx, codigo)
}
