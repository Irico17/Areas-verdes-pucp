package etl

import (
	"fmt"

	"gorm.io/gorm"
)

const insertArea = `
INSERT INTO areas_verdes (
  feature_id, source_index, codigo, nombre, uso, proy_riego, riego_act, referencia,
  perimetro_m, area_m2, geom
) VALUES (
  $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, catastro_geom_4326($11)
)`

const insertZona = `
INSERT INTO zonas (
  feature_id, source_index, codigo, nombre, uso, proy_riego, riego_act, referencia,
  perimetro_m, area_m2, geom
) VALUES (
  $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, catastro_geom_4326($11)
)`

const insertCapa = `
INSERT INTO capas_auxiliares (
  capa, feature_id, source_index, codigo, nombre, uso, proy_riego, clase, riego_act,
  referencia, pertenecen, perimetro_m, area_m2, geom
) VALUES (
  $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, catastro_geom_4326($14)
)`

// Load reemplaza el catastro semilla en una transacción.
func Load(db *gorm.DB, areas, zonas []Record, capas map[string][]Record) error {
	return db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec(`TRUNCATE areas_verdes, zonas, capas_auxiliares RESTART IDENTITY`).Error; err != nil {
			return fmt.Errorf("truncar catastro: %w", err)
		}
		for _, r := range areas {
			if err := tx.Exec(insertArea, argsCatastro(r)...).Error; err != nil {
				return fmt.Errorf("área %s: %w", r.FeatureID, err)
			}
		}
		for _, r := range zonas {
			if err := tx.Exec(insertZona, argsCatastro(r)...).Error; err != nil {
				return fmt.Errorf("zona %s: %w", r.FeatureID, err)
			}
		}
		for name, rows := range capas {
			for _, r := range rows {
				if err := tx.Exec(insertCapa, argsCapa(r)...).Error; err != nil {
					return fmt.Errorf("capa %s %s: %w", name, r.FeatureID, err)
				}
			}
		}
		return verifyLoaded(tx, len(areas), len(zonas), capas)
	})
}

func argsCatastro(r Record) []any {
	return []any{
		r.FeatureID,
		r.SourceIndex,
		r.Codigo,
		r.Nombre,
		r.Uso,
		r.ProyRiego,
		r.RiegoAct,
		r.Referencia,
		r.PerimetroM,
		r.AreaM2,
		geomArg(r),
	}
}

func argsCapa(r Record) []any {
	return []any{
		r.Capa,
		r.FeatureID,
		r.SourceIndex,
		r.Codigo,
		r.Nombre,
		r.Uso,
		r.ProyRiego,
		r.Clase,
		r.RiegoAct,
		r.Referencia,
		r.Pertenecen,
		r.PerimetroM,
		r.AreaM2,
		geomArg(r),
	}
}

func geomArg(r Record) any {
	if len(r.Geometry) == 0 || string(r.Geometry) == "null" {
		return nil
	}
	return string(r.Geometry)
}

type geomStats struct {
	N     int
	Con   int
	SRID  int
	Multi int
}

func verifyLoaded(tx *gorm.DB, areas, zonas int, capas map[string][]Record) error {
	var a geomStats
	if err := scanStats(tx, `SELECT count(*), count(geom),
		count(*) FILTER (WHERE geom IS NOT NULL AND ST_SRID(geom) = 4326),
		count(*) FILTER (WHERE geom IS NOT NULL AND GeometryType(geom) = 'MULTIPOLYGON')
		FROM areas_verdes`, &a); err != nil {
		return err
	}
	if a.N != areas || a.Con != areas || a.SRID != areas || a.Multi != areas {
		return fmt.Errorf("áreas cargadas=%+v, esperadas=%d con geometría MultiPolygon 4326", a, areas)
	}

	var z geomStats
	if err := scanStats(tx, `SELECT count(*), count(geom),
		count(*) FILTER (WHERE geom IS NOT NULL AND ST_SRID(geom) = 4326),
		count(*) FILTER (WHERE geom IS NOT NULL AND GeometryType(geom) = 'MULTIPOLYGON')
		FROM zonas`, &z); err != nil {
		return err
	}
	if z.N != zonas || z.Con != zonas || z.SRID != zonas || z.Multi != zonas {
		return fmt.Errorf("zonas cargadas=%+v, esperadas=%d con geometría MultiPolygon 4326", z, zonas)
	}

	for name, rows := range capas {
		var c geomStats
		if err := scanStats(tx, `SELECT count(*), count(geom),
			count(*) FILTER (WHERE geom IS NOT NULL AND ST_SRID(geom) = 4326),
			count(*) FILTER (WHERE geom IS NOT NULL AND GeometryType(geom) = 'MULTIPOLYGON')
			FROM capas_auxiliares WHERE capa = $1`, &c, name); err != nil {
			return err
		}
		if c.N != len(rows) || c.Con != len(rows) || c.SRID != len(rows) || c.Multi != len(rows) {
			return fmt.Errorf("capa %s cargada=%+v, esperadas=%d", name, c, len(rows))
		}
	}
	return nil
}

func scanStats(tx *gorm.DB, query string, dest *geomStats, args ...any) error {
	row := tx.Raw(query, args...).Row()
	if err := row.Scan(&dest.N, &dest.Con, &dest.SRID, &dest.Multi); err != nil {
		return err
	}
	return nil
}
