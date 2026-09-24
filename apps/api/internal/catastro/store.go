package catastro

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"campusverde/api/internal/geojson"
	"campusverde/api/internal/models"

	"gorm.io/gorm"
)

// ErrCapaDesconocida se devuelve cuando la capa no está en el catálogo de lectura.
var ErrCapaDesconocida = errors.New("capa desconocida")

// CapasConocidas son las capas opcionales de Fase B.
var CapasConocidas = []string{"jardines_reserva", "xerofitica"}

// Filter acota una lectura geo. BBox nil y Limit 0 devuelven la colección completa.
type Filter struct {
	BBox  *geojson.BBox
	Limit int
}

// Resumen son conteos del catastro cargado.
type Resumen struct {
	CRS               string      `json:"crs"`
	Areas             int64       `json:"areas"`
	AreasConGeometria int64       `json:"areas_con_geometria"`
	Zonas             int64       `json:"zonas"`
	ZonasConGeometria int64       `json:"zonas_con_geometria"`
	Capas             []CapaCount `json:"capas"`
}

// CapaCount es el tamaño de una capa auxiliar.
type CapaCount struct {
	Capa     string `json:"capa"`
	Features int64  `json:"features"`
}

// CapasIndex lista el catálogo y lo que hay cargado.
type CapasIndex struct {
	CapasConocidas []string    `json:"capas_conocidas"`
	Cargadas       []CapaCount `json:"cargadas"`
}

// Store lee catastro desde PostGIS vía GORM.
type Store struct {
	db *gorm.DB
}

func NewStore(db *gorm.DB) *Store {
	return &Store{db: db}
}

func (s *Store) Areas(ctx context.Context, f Filter) (geojson.FeatureCollection, error) {
	q, args := spatialSelect(`
		SELECT id, feature_id, source_index, codigo, nombre, uso, proy_riego, riego_act,
		       referencia, perimetro_m, area_m2, ST_AsGeoJSON(geom, 9)
		FROM areas_verdes`, "", nil, f)
	return s.scanCatastro(ctx, "areas_verdes", q, args)
}

func (s *Store) Zonas(ctx context.Context, f Filter) (geojson.FeatureCollection, error) {
	q, args := spatialSelect(`
		SELECT id, feature_id, source_index, codigo, nombre, uso, proy_riego, riego_act,
		       referencia, perimetro_m, area_m2, ST_AsGeoJSON(geom, 9)
		FROM zonas`, "", nil, f)
	return s.scanCatastro(ctx, "zonas", q, args)
}

func (s *Store) Capa(ctx context.Context, capa string, f Filter) (geojson.FeatureCollection, error) {
	if !knownCapa(capa) {
		return geojson.FeatureCollection{}, ErrCapaDesconocida
	}
	q, args := spatialSelect(`
		SELECT id, feature_id, source_index, codigo, nombre, uso, proy_riego, riego_act,
		       referencia, perimetro_m, area_m2, clase, pertenecen, ST_AsGeoJSON(geom, 9)
		FROM capas_auxiliares`, "capa = $1", []any{capa}, f)
	return s.scanCapa(ctx, capa, q, args)
}

func (s *Store) Resumen(ctx context.Context) (Resumen, error) {
	var out Resumen
	out.CRS = "EPSG:4326"
	db := s.db.WithContext(ctx)
	if err := db.Model(&models.AreaVerde{}).Count(&out.Areas).Error; err != nil {
		return out, err
	}
	if err := db.Model(&models.AreaVerde{}).Where("geom IS NOT NULL").Count(&out.AreasConGeometria).Error; err != nil {
		return out, err
	}
	if err := db.Model(&models.Zona{}).Count(&out.Zonas).Error; err != nil {
		return out, err
	}
	if err := db.Model(&models.Zona{}).Where("geom IS NOT NULL").Count(&out.ZonasConGeometria).Error; err != nil {
		return out, err
	}
	capas, err := s.capaCounts(ctx)
	if err != nil {
		return out, err
	}
	out.Capas = capas
	return out, nil
}

func (s *Store) Capas(ctx context.Context) (CapasIndex, error) {
	cargadas, err := s.capaCounts(ctx)
	if err != nil {
		return CapasIndex{}, err
	}
	if cargadas == nil {
		cargadas = []CapaCount{}
	}
	return CapasIndex{CapasConocidas: CapasConocidas, Cargadas: cargadas}, nil
}

func (s *Store) PostGISVersion(ctx context.Context) (string, error) {
	sqlDB, err := s.db.DB()
	if err != nil {
		return "", err
	}
	if err := sqlDB.PingContext(ctx); err != nil {
		return "", err
	}
	var version string
	err = s.db.WithContext(ctx).Raw("SELECT PostGIS_Version()").Scan(&version).Error
	return version, err
}

func (s *Store) capaCounts(ctx context.Context) ([]CapaCount, error) {
	var rows []CapaCount
	err := s.db.WithContext(ctx).Raw(`
		SELECT capa, count(*) AS features
		FROM capas_auxiliares
		GROUP BY capa
		ORDER BY capa`).Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	if rows == nil {
		rows = []CapaCount{}
	}
	return rows, nil
}

func (s *Store) scanCatastro(ctx context.Context, name, query string, args []any) (geojson.FeatureCollection, error) {
	fc := geojson.Collection(name)
	rows, err := s.db.WithContext(ctx).Raw(query, args...).Rows()
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
		fc.Features = append(fc.Features, geojson.Feature{
			Type:     "Feature",
			ID:       featureID,
			Geometry: geomJSON(geom),
			Properties: geojson.CatastroProperties{
				ID:          id,
				FeatureID:   featureID,
				SourceIndex: sourceIndex,
				Codigo:      nullStr(codigo),
				Nombre:      nullStr(nombre),
				Uso:         nullStr(uso),
				ProyRiego:   nullStr(proy),
				RiegoAct:    nullStr(riego),
				Referencia:  nullStr(ref),
				PerimetroM:  geojson.FloatPtr(nullFloat(per)),
				AreaM2:      geojson.FloatPtr(nullFloat(area)),
			},
		})
	}
	return fc, rows.Err()
}

func (s *Store) scanCapa(ctx context.Context, name, query string, args []any) (geojson.FeatureCollection, error) {
	fc := geojson.Collection(name)
	rows, err := s.db.WithContext(ctx).Raw(query, args...).Rows()
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
		fc.Features = append(fc.Features, geojson.Feature{
			Type:     "Feature",
			ID:       featureID,
			Geometry: geomJSON(geom),
			Properties: geojson.CapaProperties{
				CatastroProperties: geojson.CatastroProperties{
					ID:          id,
					FeatureID:   featureID,
					SourceIndex: sourceIndex,
					Codigo:      nullStr(codigo),
					Nombre:      nullStr(nombre),
					Uso:         nullStr(uso),
					ProyRiego:   nullStr(proy),
					RiegoAct:    nullStr(riego),
					Referencia:  nullStr(ref),
					PerimetroM:  geojson.FloatPtr(nullFloat(per)),
					AreaM2:      geojson.FloatPtr(nullFloat(area)),
				},
				Capa:       name,
				Clase:      nullStr(clase),
				Pertenecen: nullStr(pertenecen),
			},
		})
	}
	return fc, rows.Err()
}

func spatialSelect(base, extraWhere string, extraArgs []any, f Filter) (string, []any) {
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

func knownCapa(name string) bool {
	for _, c := range CapasConocidas {
		if c == name {
			return true
		}
	}
	return false
}

func nullStr(v sql.NullString) *string {
	if !v.Valid {
		return nil
	}
	s := strings.TrimSpace(v.String)
	if s == "" {
		return nil
	}
	return &s
}

func nullFloat(v sql.NullFloat64) *float64 {
	if !v.Valid {
		return nil
	}
	f := v.Float64
	return &f
}

func geomJSON(v sql.NullString) []byte {
	if !v.Valid || strings.TrimSpace(v.String) == "" {
		return []byte("null")
	}
	return []byte(v.String)
}
