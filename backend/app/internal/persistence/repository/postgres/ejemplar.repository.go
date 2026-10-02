package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"reflect"
	"strings"

	"gorm.io/gorm"

	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/contracts"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/entities"
	domainErrors "github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/errors"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/persistence/mapper"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/persistence/models"
)

type ejemplarRepository struct {
	db *gorm.DB
}

const selectEjemplar = `
	e.id, e.numero_origen, e.codigo, e.especie_id, e.nombre_comun,
	e.tipo_vegetacion, e.cantidad, e.ubicacion_lugar_id, e.referencia,
	e.lat, e.lon, e.observacion_fen_2026, e.salud, e.sector_cuartel_id,
	c.nombre AS sector_cuartel_nombre, c.clase AS sector_cuartel_clase,
	esp.nombre_cientifico AS especie_cientifico, lug.nombre AS lugar_nombre,
	e.activo, e.created_at, e.updated_at`

const fromEjemplar = `
	FROM ejemplares e
	LEFT JOIN catalogos c ON c.id = e.sector_cuartel_id
	LEFT JOIN especies esp ON esp.id = e.especie_id
	LEFT JOIN lugares lug ON lug.id = e.ubicacion_lugar_id`

// NewEjemplarRepository creates a new IEjemplarRepository instance.
func NewEjemplarRepository(db *gorm.DB) contracts.IEjemplarRepository {
	return &ejemplarRepository{db: db}
}

func (r *ejemplarRepository) Listar(ctx context.Context, limit, offset int, q string) ([]entities.Ejemplar, int, error) {
	q = strings.TrimSpace(q)
	filtro := `
		WHERE e.activo
		  AND (
		    $1 = ''
		    OR strpos(lower(coalesce(e.codigo, '')), lower($1)) > 0
		    OR strpos(lower(coalesce(e.nombre_comun, '')), lower($1)) > 0
		    OR coalesce(e.numero_origen::text, '') = $1
		  )`
	var total int
	if err := r.db.WithContext(ctx).Raw(`SELECT count(*) `+fromEjemplar+filtro, q).Scan(&total).Error; err != nil {
		return nil, 0, err
	}

	consulta := `SELECT ` + selectEjemplar + fromEjemplar + filtro + `
		ORDER BY e.numero_origen NULLS LAST, e.id`
	args := []any{q}
	if limit > 0 {
		if offset < 0 {
			offset = 0
		}
		consulta += ` LIMIT $2 OFFSET $3`
		args = append(args, limit, offset)
	}

	var modelsList []models.EjemplarModel
	if err := r.db.WithContext(ctx).Raw(consulta, args...).Scan(&modelsList).Error; err != nil {
		return nil, 0, err
	}

	out := make([]entities.Ejemplar, len(modelsList))
	for i, m := range modelsList {
		out[i] = *mapper.EjemplarModelToEntity(&m)
	}
	return out, total, nil
}

func (r *ejemplarRepository) Crear(ctx context.Context, e entities.Ejemplar) (entities.Ejemplar, error) {
	var m models.EjemplarModel
	err := r.db.WithContext(ctx).Raw(`
		INSERT INTO ejemplares (
		  numero_origen, codigo, especie_id, nombre_comun, tipo_vegetacion, cantidad,
		  ubicacion_lugar_id, referencia, lat, lon, observacion_fen_2026, salud, geom
		) VALUES (
		  $1, NULLIF($2, ''), $3, NULLIF($4, ''), NULLIF($5, ''), $6,
		  $7, NULLIF($8, ''), $9, $10, NULLIF($11, ''), NULL,
		  CASE WHEN $9::float8 IS NULL THEN NULL
		       ELSE ST_SetSRID(ST_MakePoint($10, $9), 4326) END
		)
		RETURNING id, numero_origen, codigo, especie_id, nombre_comun, tipo_vegetacion, cantidad,
		          ubicacion_lugar_id, referencia, lat, lon, observacion_fen_2026, salud, activo, created_at, updated_at`,
		e.NumeroOrigen, strings.TrimSpace(e.Codigo), e.EspecieID, strings.TrimSpace(e.NombreComun),
		strings.TrimSpace(e.TipoVegetacion), e.Cantidad, e.UbicacionLugarID, strings.TrimSpace(e.Referencia),
		e.Lat, e.Lon, strings.TrimSpace(e.ObservacionFen2026),
	).Scan(&m).Error
	if err != nil {
		return entities.Ejemplar{}, domainErrors.ErrEntrada
	}
	return *mapper.EjemplarModelToEntity(&m), nil
}

func (r *ejemplarRepository) Actualizar(ctx context.Context, id int64, p entities.ParcheEjemplar, usuarioID *int64) (entities.Ejemplar, error) {
	var out entities.Ejemplar
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		antes, err := leerEjemplarTx(tx, id, true)
		if err != nil {
			return err
		}
		despues := aplicarParche(antes, p)
		if reflect.DeepEqual(instantanea(antes), instantanea(despues)) {
			out = antes
			return nil
		}
		if antes.Codigo != despues.Codigo {
			if err := tx.Exec(`
				INSERT INTO codigos_historicos (ejemplar_id, codigo_anterior, codigo_nuevo)
				VALUES ($1, $2, $3)`,
				id, antes.Codigo, despues.Codigo).Error; err != nil {
				return err
			}
		}
		if err := escribirEjemplar(tx, despues); err != nil {
			return traducirEjemplar(err)
		}
		if err := registrarCambioTx(tx, "ejemplares", fmt.Sprint(id), "edicion", instantanea(antes), instantanea(despues), usuarioID); err != nil {
			return err
		}
		out, err = leerEjemplarTx(tx, id, false)
		return err
	})
	if err != nil {
		return entities.Ejemplar{}, err
	}
	return out, nil
}

func (r *ejemplarRepository) Recodificar(ctx context.Context, ejemplarID int64, codigoNuevo string, usuarioID *int64) (entities.CodigoHistorico, error) {
	var out entities.CodigoHistorico
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var anterior sql.NullString
		row := tx.Raw(`SELECT codigo FROM ejemplares WHERE id = $1 AND activo`, ejemplarID).Row()
		if err := row.Scan(&anterior); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return domainErrors.ErrNoEncontrado
			}
			return err
		}
		previo := ""
		if anterior.Valid {
			previo = anterior.String
		}
		if previo == codigoNuevo {
			out = entities.CodigoHistorico{EjemplarID: ejemplarID, CodigoAnterior: previo, CodigoNuevo: codigoNuevo}
			return nil
		}
		var m models.CodigoHistoricoModel
		if err := tx.Raw(`
			INSERT INTO codigos_historicos (ejemplar_id, codigo_anterior, codigo_nuevo)
			VALUES ($1, $2, $3)
			RETURNING id, ejemplar_id, codigo_anterior, codigo_nuevo, created_at`,
			ejemplarID, previo, codigoNuevo).Scan(&m).Error; err != nil {
			return err
		}
		res := tx.Exec(`UPDATE ejemplares SET codigo = $2, updated_at = now() WHERE id = $1 AND activo`, ejemplarID, codigoNuevo)
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected != 1 {
			return domainErrors.ErrNoEncontrado
		}
		if err := registrarCambioTx(tx, "ejemplares", fmt.Sprint(ejemplarID), "edicion",
			map[string]any{"codigo": previo},
			map[string]any{"codigo": codigoNuevo},
			usuarioID); err != nil {
			return err
		}
		out = *mapper.CodigoHistoricoModelToEntity(&m)
		return nil
	})
	if err != nil {
		return entities.CodigoHistorico{}, err
	}
	return out, nil
}

func (r *ejemplarRepository) ListarCodigos(ctx context.Context, ejemplarID int64) ([]entities.CodigoHistorico, error) {
	var list []models.CodigoHistoricoModel
	err := r.db.WithContext(ctx).Where("ejemplar_id = ?", ejemplarID).Order("id").Find(&list).Error
	if err != nil {
		return nil, err
	}
	out := make([]entities.CodigoHistorico, len(list))
	for i, m := range list {
		out[i] = *mapper.CodigoHistoricoModelToEntity(&m)
	}
	return out, nil
}

func leerEjemplarTx(tx *gorm.DB, id int64, bloquear bool) (entities.Ejemplar, error) {
	consulta := `SELECT ` + selectEjemplar + fromEjemplar + ` WHERE e.id = $1 AND e.activo`
	if bloquear {
		consulta += ` FOR UPDATE OF e`
	}
	var m models.EjemplarModel
	res := tx.Raw(consulta, id).Scan(&m)
	if res.Error != nil {
		return entities.Ejemplar{}, res.Error
	}
	if m.ID == 0 {
		return entities.Ejemplar{}, domainErrors.ErrNoEncontrado
	}
	return *mapper.EjemplarModelToEntity(&m), nil
}

func escribirEjemplar(tx *gorm.DB, e entities.Ejemplar) error {
	var codigo, salud, especie, lugar, sector, lat, lon any
	if e.Codigo != "" {
		codigo = e.Codigo
	}
	if e.Salud != nil && *e.Salud != "" {
		salud = *e.Salud
	}
	if e.EspecieID != nil {
		especie = *e.EspecieID
	}
	if e.UbicacionLugarID != nil {
		lugar = *e.UbicacionLugarID
	}
	if e.SectorCuartelID != nil {
		sector = *e.SectorCuartelID
	}
	if e.Lat != nil {
		lat = *e.Lat
	}
	if e.Lon != nil {
		lon = *e.Lon
	}
	res := tx.Exec(`
		UPDATE ejemplares SET
		  codigo = $2,
		  especie_id = $3,
		  salud = $4,
		  lat = $5,
		  lon = $6,
		  ubicacion_lugar_id = $7,
		  sector_cuartel_id = $8,
		  geom = CASE
		    WHEN $5::float8 IS NULL OR $6::float8 IS NULL THEN NULL
		    ELSE ST_SetSRID(ST_MakePoint($6, $5), 4326)
		  END,
		  updated_at = now()
		WHERE id = $1 AND activo`,
		e.ID, codigo, especie, salud, lat, lon, lugar, sector)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected != 1 {
		return domainErrors.ErrNoEncontrado
	}
	return nil
}

func aplicarParche(e entities.Ejemplar, p entities.ParcheEjemplar) entities.Ejemplar {
	if p.TieneCodigo && p.Codigo != nil {
		e.Codigo = strings.TrimSpace(*p.Codigo)
	}
	if p.TieneEspecie {
		e.EspecieID = p.EspecieID
	}
	if p.TieneSalud {
		e.Salud = p.Salud
	}
	if p.TieneLat && p.TieneLon {
		e.Lat = p.Lat
		e.Lon = p.Lon
	}
	if p.TieneLugar {
		e.UbicacionLugarID = p.UbicacionLugarID
	}
	if p.TieneSector {
		e.SectorCuartelID = p.SectorCuartelID
	}
	return e
}

func instantanea(e entities.Ejemplar) map[string]any {
	return map[string]any{
		"codigo":             e.Codigo,
		"especie_id":         e.EspecieID,
		"salud":              e.Salud,
		"lat":                e.Lat,
		"lon":                e.Lon,
		"ubicacion_lugar_id": e.UbicacionLugarID,
		"sector_cuartel_id":  e.SectorCuartelID,
	}
}

func traducirEjemplar(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, domainErrors.ErrNoEncontrado) || errors.Is(err, domainErrors.ErrEntrada) {
		return err
	}
	msg := strings.ToLower(err.Error())
	if strings.Contains(msg, "sector o cuartel desconocido") ||
		strings.Contains(msg, "foreign key") ||
		strings.Contains(msg, "23503") ||
		strings.Contains(msg, "23514") ||
		strings.Contains(msg, "ejemplares_lat_chk") ||
		strings.Contains(msg, "ejemplares_lon_chk") {
		return domainErrors.ErrEntrada
	}
	return err
}
