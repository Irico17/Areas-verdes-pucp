package etl

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"gorm.io/gorm"
)

const insertArea = `
INSERT INTO areas_verdes (
  feature_id, source_index, codigo, nombre, uso, proy_riego, riego_act, referencia,
  perimetro_m, area_m2, geom
) VALUES (
  $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, catastro_geom_4326($11)
) RETURNING id`

// motivoCodigoDuplicado es el mismo texto que usa la migración 014 para este
// tipo de corrección, así el historial de cambios queda consistente.
const motivoCodigoDuplicado = "codigo duplicado; se conservo el de menor id. Corrija el codigo desde la ficha"

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
		if err := negarSiHayDependientes(tx); err != nil {
			return err
		}
		// zonas es una vista sobre poligonos_cuadrilla. zonas_origen no se toca.
		if err := tx.Exec(`TRUNCATE areas_verdes, poligonos_cuadrilla, capas_auxiliares RESTART IDENTITY CASCADE`).Error; err != nil {
			return fmt.Errorf("truncar catastro: %w", err)
		}
		if err := insertAreas(tx, areas); err != nil {
			return err
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
		// El TRUNCATE anterior deja sector en NULL; poligonos_sector_ref no
		// tiene FK a poligonos_cuadrilla, así que sobrevive y esto lo vuelve a
		// completar en cada corrida (ver internal/infrastructure/etl/sector.go).
		if _, err := AplicarSectores(tx); err != nil {
			return fmt.Errorf("aplicar sectores: %w", err)
		}
		return verifyLoaded(tx, len(areas), len(zonas), capas)
	})
}

// TablasGuardLoad son las tablas que deben estar vacías para que Load proceda.
var TablasGuardLoad = []string{
	"areas_verdes",
	"poligonos_cuadrilla",
	"capas_auxiliares",
	"ejemplares",
	"asignaciones_poligono",
}

// NegarSiHayDependientes evita que el TRUNCATE ... CASCADE de Load borre datos
// existentes: Load solo procede sobre un catastro vacío (areas_verdes,
// poligonos_cuadrilla, capas_auxiliares, ejemplares y asignaciones_poligono sin
// filas). En una base con datos, se niega en vez de borrar filas y remite a
// etl-lote (upsert).
func NegarSiHayDependientes(tx *gorm.DB) error {
	for _, tabla := range TablasGuardLoad {
		var n int
		if err := tx.Raw(fmt.Sprintf(`SELECT count(*) FROM %s`, tabla)).Scan(&n).Error; err != nil {
			return err
		}
		if n > 0 {
			return fmt.Errorf(
				"Load: %s tiene %d fila(s); TRUNCATE las borraría. Load solo procede sobre un catastro vacío. Usa etl-lote (upsert, sin TRUNCATE) en una base con datos",
				tabla, n)
		}
	}
	return nil
}

var negarSiHayDependientes = NegarSiHayDependientes

// insertAreas inserta el catastro de áreas verdes. El catastro de origen repite
// código a veces (p. ej. "D 20" en dos features): la unicidad de codigo la exige
// areas_verdes_codigo_uidx (migración 014), así que antes de insertar se detectan
// los duplicados y, tras conocer el id de cada fila repetida, se renombran con el
// mismo criterio que usan las migraciones 014/034 (menor id conserva el código).
func insertAreas(tx *gorm.DB, areas []Record) error {
	duplicados := duplicadosPorIndice(areas)
	for i, r := range areas {
		args := argsCatastro(r)
		if _, esDuplicado := duplicados[i]; esDuplicado {
			args[2] = nil // codigo se asigna luego de conocer el id
		}
		var id int64
		if err := tx.Raw(insertArea, args...).Row().Scan(&id); err != nil {
			return fmt.Errorf("área %s: %w", r.FeatureID, err)
		}
		if codigoOriginal, esDuplicado := duplicados[i]; esDuplicado {
			if err := renombrarCodigoDuplicado(tx, id, codigoOriginal); err != nil {
				return fmt.Errorf("área %s: %w", r.FeatureID, err)
			}
		}
	}
	return nil
}

// duplicadosPorIndice devuelve, para cada área cuyo código ya apareció antes en la
// lista, el código original. Las áreas se insertan en este mismo orden, así que la
// primera aparición recibe el id menor y conserva el código sin cambios. La
// comparación usa el código tal cual (sin recortar), igual que el índice único de
// la base; TrimSpace solo sirve para descartar códigos en blanco.
func duplicadosPorIndice(areas []Record) map[int]string {
	vistos := map[string]bool{}
	duplicados := map[int]string{}
	for i, r := range areas {
		if r.Codigo == nil {
			continue
		}
		codigo := *r.Codigo
		if strings.TrimSpace(codigo) == "" {
			continue
		}
		if vistos[codigo] {
			duplicados[i] = codigo
			continue
		}
		vistos[codigo] = true
	}
	return duplicados
}

// renombrarCodigoDuplicado aplica "codigo · id" (con sufijo -n si también choca) y dejar
// constancia en cambios, igual que las migraciones 014 y 034. No se salta si el cambio
// ya está registrado, para que una segunda corrida del ETL no duplique el historial.
func renombrarCodigoDuplicado(tx *gorm.DB, id int64, codigoOriginal string) error {
	idStr := strconv.FormatInt(id, 10)
	nuevo := codigoOriginal + " ·" + idStr
	for n := 2; ; n++ {
		var existe bool
		if err := tx.Raw(`SELECT EXISTS (SELECT 1 FROM areas_verdes WHERE codigo = $1 AND id <> $2)`,
			nuevo, id).Row().Scan(&existe); err != nil {
			return err
		}
		if !existe {
			break
		}
		nuevo = fmt.Sprintf("%s ·%s-%d", codigoOriginal, idStr, n)
	}

	if err := tx.Exec(`UPDATE areas_verdes SET codigo = $1, updated_at = now() WHERE id = $2`,
		nuevo, id).Error; err != nil {
		return err
	}

	antes, _ := json.Marshal(map[string]any{"codigo": codigoOriginal})
	despues, _ := json.Marshal(map[string]any{"codigo": nuevo, "motivo": motivoCodigoDuplicado})

	var yaRegistrado bool
	if err := tx.Raw(`
		SELECT EXISTS (
			SELECT 1 FROM cambios
			WHERE entidad = 'areas_verdes' AND entidad_id = $1 AND accion = 'edicion'
			  AND despues->>'motivo' = $2 AND despues->>'codigo' = $3
		)`, idStr, motivoCodigoDuplicado, nuevo).Row().Scan(&yaRegistrado); err != nil {
		return err
	}
	if yaRegistrado {
		return nil
	}
	return tx.Exec(`
		INSERT INTO cambios (entidad, entidad_id, accion, antes, despues)
		VALUES ('areas_verdes', $1, 'edicion', $2::jsonb, $3::jsonb)`,
		idStr, string(antes), string(despues)).Error
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

	if zonas == ExpectedZonas {
		var sinSector int
		if err := tx.Raw(`
			SELECT count(*) FROM poligonos_cuadrilla
			WHERE sector IS NULL AND source_index IN (SELECT source_index FROM poligonos_sector_ref)
		`).Scan(&sinSector).Error; err != nil {
			return err
		}
		if sinSector != 0 {
			return fmt.Errorf("AplicarSectores dejó %d polígono(s) con sector de referencia sin completar", sinSector)
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
