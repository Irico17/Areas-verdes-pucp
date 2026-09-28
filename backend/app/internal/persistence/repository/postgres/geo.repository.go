package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"gorm.io/gorm"

	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/contracts"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/dto"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/constants/enums"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/entities"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/persistence/mapper"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/persistence/models"
)

type geoRepository struct {
	db *gorm.DB
}

// NewGeoRepository creates a new GeoRepository instance.
func NewGeoRepository(db *gorm.DB) contracts.IGeoRepository {
	return &geoRepository{db: db}
}

func (r *geoRepository) Areas(ctx context.Context, f dto.FiltroGeoDTO) (entities.FeatureCollection, error) {
	q, args := spatialSelect(`
		SELECT id, feature_id, source_index, codigo, nombre, uso, proy_riego, riego_act,
		       referencia, perimetro_m, area_m2, ST_AsGeoJSON(geom, 9)
		FROM areas_verdes`, "", nil, f)
	return r.scanCatastro(ctx, "areas_verdes", q, args)
}

func (r *geoRepository) Zonas(ctx context.Context, f dto.FiltroGeoDTO) (entities.FeatureCollection, error) {
	q, args := spatialSelect(`
		SELECT id, feature_id, source_index, codigo, nombre, uso, proy_riego, riego_act,
		       referencia, perimetro_m, area_m2, sector, ST_AsGeoJSON(geom, 9)
		FROM zonas`, "", nil, f)
	return r.scanZonas(ctx, "zonas", q, args)
}

func (r *geoRepository) Capa(ctx context.Context, capa string, f dto.FiltroGeoDTO) (entities.FeatureCollection, error) {
	q, args := spatialSelect(`
		SELECT id, feature_id, source_index, codigo, nombre, uso, proy_riego, riego_act,
		       referencia, perimetro_m, area_m2, clase, pertenecen, ST_AsGeoJSON(geom, 9)
		FROM capas_auxiliares`, "capa = $1", []any{capa}, f)
	return r.scanCapa(ctx, capa, q, args)
}

func (r *geoRepository) Capas(ctx context.Context) (dto.CapasIndexDTO, error) {
	cargadas, err := r.capaCounts(ctx)
	if err != nil {
		return dto.CapasIndexDTO{}, err
	}
	if cargadas == nil {
		cargadas = []dto.CapaCountDTO{}
	}
	return dto.CapasIndexDTO{Cargadas: cargadas}, nil
}

func (r *geoRepository) Resumen(ctx context.Context) (dto.ResumenDTO, error) {
	var out dto.ResumenDTO
	out.CRS = "EPSG:4326"
	db := r.db.WithContext(ctx)
	if err := db.Model(&models.AreaVerdeModel{}).Count(&out.Areas).Error; err != nil {
		return out, err
	}
	if err := db.Model(&models.AreaVerdeModel{}).Where("geom IS NOT NULL").Count(&out.AreasConGeometria).Error; err != nil {
		return out, err
	}
	if err := db.Model(&models.PoligonoCuadrillaModel{}).Count(&out.Zonas).Error; err != nil {
		return out, err
	}
	if err := db.Model(&models.PoligonoCuadrillaModel{}).Where("geom IS NOT NULL").Count(&out.ZonasConGeometria).Error; err != nil {
		return out, err
	}
	capas, err := r.capaCounts(ctx)
	if err != nil {
		return out, err
	}
	out.Capas = capas
	return out, nil
}

func (r *geoRepository) capaCounts(ctx context.Context) ([]dto.CapaCountDTO, error) {
	var rows []dto.CapaCountDTO
	err := r.db.WithContext(ctx).Raw(`
		SELECT capa, count(*) AS features
		FROM capas_auxiliares
		GROUP BY capa
		ORDER BY capa`).Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	if rows == nil {
		rows = []dto.CapaCountDTO{}
	}
	return rows, nil
}

func (r *geoRepository) scanCatastro(ctx context.Context, name, query string, args []any) (entities.FeatureCollection, error) {
	fc := entities.Collection(name)
	rows, err := r.db.WithContext(ctx).Raw(query, args...).Rows()
	if err != nil {
		return fc, err
	}
	defer rows.Close()

	for rows.Next() {
		var (
			id          int64
			featureID   string
			sourceIndex int
			codigo      sql.NullString
			nombre      sql.NullString
			uso         sql.NullString
			proy        sql.NullString
			riego       sql.NullString
			ref         sql.NullString
			per         sql.NullFloat64
			area        sql.NullFloat64
			geom        sql.NullString
		)
		if err := rows.Scan(&id, &featureID, &sourceIndex, &codigo, &nombre, &uso, &proy, &riego, &ref, &per, &area, &geom); err != nil {
			return fc, err
		}
		fc.Features = append(fc.Features, entities.Feature{
			Type:     "Feature",
			ID:       featureID,
			Geometry: mapper.GeomJSON(geom),
			Properties: dto.CatastroPropertiesDTO{
				ID:          id,
				FeatureID:   featureID,
				SourceIndex: sourceIndex,
				Codigo:      mapper.NullStr(codigo),
				Nombre:      mapper.NullStr(nombre),
				Uso:         mapper.NullStr(uso),
				ProyRiego:   mapper.NullStr(proy),
				RiegoAct:    mapper.NullStr(riego),
				Referencia:  mapper.NullStr(ref),
				PerimetroM:  dto.FloatPtr(mapper.NullFloat(per)),
				AreaM2:      dto.FloatPtr(mapper.NullFloat(area)),
			},
		})
	}
	return fc, rows.Err()
}

func (r *geoRepository) scanZonas(ctx context.Context, name, query string, args []any) (entities.FeatureCollection, error) {
	fc := entities.Collection(name)
	rows, err := r.db.WithContext(ctx).Raw(query, args...).Rows()
	if err != nil {
		return fc, err
	}
	defer rows.Close()

	for rows.Next() {
		var (
			id          int64
			featureID   string
			sourceIndex int
			codigo      sql.NullString
			nombre      sql.NullString
			uso         sql.NullString
			proy        sql.NullString
			riego       sql.NullString
			ref         sql.NullString
			per         sql.NullFloat64
			area        sql.NullFloat64
			sector      sql.NullString
			geom        sql.NullString
		)
		if err := rows.Scan(&id, &featureID, &sourceIndex, &codigo, &nombre, &uso, &proy, &riego, &ref, &per, &area, &sector, &geom); err != nil {
			return fc, err
		}
		sectorPtr := mapper.NullStr(sector)
		var etiquetaPtr *string
		if sectorPtr != nil {
			if etiqueta := enums.EtiquetaSector(*sectorPtr); etiqueta != "" {
				etiquetaPtr = &etiqueta
			}
		}
		fc.Features = append(fc.Features, entities.Feature{
			Type:     "Feature",
			ID:       featureID,
			Geometry: mapper.GeomJSON(geom),
			Properties: dto.ZonaPropertiesDTO{
				CatastroPropertiesDTO: dto.CatastroPropertiesDTO{
					ID:          id,
					FeatureID:   featureID,
					SourceIndex: sourceIndex,
					Codigo:      mapper.NullStr(codigo),
					Nombre:      mapper.NullStr(nombre),
					Uso:         mapper.NullStr(uso),
					ProyRiego:   mapper.NullStr(proy),
					RiegoAct:    mapper.NullStr(riego),
					Referencia:  mapper.NullStr(ref),
					PerimetroM:  dto.FloatPtr(mapper.NullFloat(per)),
					AreaM2:      dto.FloatPtr(mapper.NullFloat(area)),
				},
				Sector:         sectorPtr,
				SectorEtiqueta: etiquetaPtr,
			},
		})
	}
	return fc, rows.Err()
}

func (r *geoRepository) scanCapa(ctx context.Context, name, query string, args []any) (entities.FeatureCollection, error) {
	fc := entities.Collection(name)
	rows, err := r.db.WithContext(ctx).Raw(query, args...).Rows()
	if err != nil {
		return fc, err
	}
	defer rows.Close()

	for rows.Next() {
		var (
			id          int64
			featureID   string
			sourceIndex int
			codigo      sql.NullString
			nombre      sql.NullString
			uso         sql.NullString
			proy        sql.NullString
			riego       sql.NullString
			ref         sql.NullString
			per         sql.NullFloat64
			area        sql.NullFloat64
			clase       sql.NullString
			pertenecen  sql.NullString
			geom        sql.NullString
		)
		if err := rows.Scan(
			&id, &featureID, &sourceIndex, &codigo, &nombre, &uso, &proy, &riego, &ref,
			&per, &area, &clase, &pertenecen, &geom,
		); err != nil {
			return fc, err
		}
		fc.Features = append(fc.Features, entities.Feature{
			Type:     "Feature",
			ID:       featureID,
			Geometry: mapper.GeomJSON(geom),
			Properties: dto.CapaPropertiesDTO{
				CatastroPropertiesDTO: dto.CatastroPropertiesDTO{
					ID:          id,
					FeatureID:   featureID,
					SourceIndex: sourceIndex,
					Codigo:      mapper.NullStr(codigo),
					Nombre:      mapper.NullStr(nombre),
					Uso:         mapper.NullStr(uso),
					ProyRiego:   mapper.NullStr(proy),
					RiegoAct:    mapper.NullStr(riego),
					Referencia:  mapper.NullStr(ref),
					PerimetroM:  dto.FloatPtr(mapper.NullFloat(per)),
					AreaM2:      dto.FloatPtr(mapper.NullFloat(area)),
				},
				Capa:       name,
				Clase:      mapper.NullStr(clase),
				Pertenecen: mapper.NullStr(pertenecen),
			},
		})
	}
	return fc, rows.Err()
}

func spatialSelect(base, extraWhere string, extraArgs []any, f dto.FiltroGeoDTO) (string, []any) {
	args := append([]any{}, extraArgs...)
	conds := make([]string, 0, 2)
	if extraWhere != "" {
		conds = append(conds, extraWhere)
	}
	if f.BBox != nil {
		args = append(args, f.BBox.MinX, f.BBox.MinY, f.BBox.MaxX, f.BBox.MaxY)
		n := len(args)
		conds = append(conds, fmt.Sprintf(
			"geom IS NOT NULL AND ST_Intersects(geom, ST_MakeEnvelope($%d,$%d,$%d,$%d,4326))",
			n-3, n-2, n-1, n,
		))
	}
	q := base
	if len(conds) > 0 {
		q += " WHERE " + strings.Join(conds, " AND ")
	}
	q += " ORDER BY source_index"
	if f.Limit > 0 {
		args = append(args, f.Limit)
		q += fmt.Sprintf(" LIMIT $%d", len(args))
	}
	return q, args
}
