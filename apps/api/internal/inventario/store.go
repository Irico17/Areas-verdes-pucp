package inventario

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	"campusverde/api/internal/geojson"

	"gorm.io/gorm"
)

// Capas son los overlays opcionales. Si el archivo raw no existe, el ETL no las carga.
var Capas = []string{
	"bebederos",
	"fauna",
	"playas_estacionamiento",
	"puertas",
	"area_vereda_peligro",
	"flora",
	"cafetos",
	"tachos",
}

// ErrCapa is a unknown layer name.
var ErrCapa = errors.New("capa de inventario desconocida")

// CapaCount is one loaded layer.
type CapaCount struct {
	Capa     string `json:"capa"`
	Features int64  `json:"features"`
}

// Index lists the catalog and what is loaded.
type Index struct {
	Capas    []string    `json:"capas"`
	Cargadas []CapaCount `json:"cargadas"`
}

// Props travel on each feature. Foto is a local file name, or empty.
type Props struct {
	FeatureID string `json:"feature_id"`
	Capa      string `json:"capa"`
	Nombre    string `json:"nombre,omitempty"`
	Subtipo   string `json:"subtipo,omitempty"`
	Detalle   string `json:"detalle,omitempty"`
	Lugar     string `json:"lugar,omitempty"`
	Foto      string `json:"foto,omitempty"`
}

// Store reads inventario.
type Store struct {
	db *gorm.DB
}

func NewStore(db *gorm.DB) *Store { return &Store{db: db} }

func Known(capa string) bool {
	for _, name := range Capas {
		if name == capa {
			return true
		}
	}
	return false
}

func (s *Store) Index(ctx context.Context) (Index, error) {
	out := Index{Capas: Capas, Cargadas: []CapaCount{}}
	err := s.db.WithContext(ctx).Raw(`
		SELECT capa, count(*) AS features
		FROM inventario
		GROUP BY capa
		ORDER BY capa`).Scan(&out.Cargadas).Error
	if err != nil {
		return out, err
	}
	if out.Cargadas == nil {
		out.Cargadas = []CapaCount{}
	}
	return out, nil
}

func (s *Store) Capa(ctx context.Context, capa string) (geojson.FeatureCollection, error) {
	fc := geojson.Collection(capa)
	if !Known(capa) {
		return fc, ErrCapa
	}
	rows, err := s.db.WithContext(ctx).Raw(`
		SELECT feature_id, nombre, subtipo, detalle, lugar, foto, ST_AsGeoJSON(geom, 7)
		FROM inventario
		WHERE capa = $1 AND geom IS NOT NULL
		ORDER BY feature_id`, capa).Rows()
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
		props := Props{FeatureID: id, Capa: capa}
		if nombre.Valid {
			props.Nombre = nombre.String
		}
		if sub.Valid {
			props.Subtipo = sub.String
		}
		if det.Valid {
			props.Detalle = det.String
		}
		if lugar.Valid {
			props.Lugar = lugar.String
		}
		if foto.Valid {
			props.Foto = foto.String
		}
		fc.Features = append(fc.Features, geojson.Feature{
			Type: "Feature", ID: id, Geometry: raw, Properties: props,
		})
	}
	return fc, rows.Err()
}
