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
	"xerofitica",
	"jardines_reserva",
	"puntos_pucp",
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
	q, args := consultaCapa(capa)
	rows, err := s.db.WithContext(ctx).Raw(q, args...).Rows()
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
