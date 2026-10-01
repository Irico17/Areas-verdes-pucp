// Package postgres implements repository interfaces using PostgreSQL and GORM.
package postgres

import (
	"bytes"
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"

	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/contracts"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/entities"
	domainErrors "github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/errors"
)

type evidenciaRepository struct {
	db *gorm.DB
}

// NewEvidenciaRepository creates a new instance of IEvidenciaRepository.
func NewEvidenciaRepository(db *gorm.DB) contracts.IEvidenciaRepository {
	return &evidenciaRepository{db: db}
}

// Listar retrieves the latest 100 evidence items, optionally filtered by activity ID.
func (r *evidenciaRepository) Listar(ctx context.Context, actividadID string) ([]entities.Evidencia, error) {
	rows, err := r.db.WithContext(ctx).Raw(`
		SELECT id::text, COALESCE(actividad_id::text, ''), nombre, mime, bytes, nota, created_at
		FROM evidencias
		WHERE ($1 = '' OR actividad_id::text = $1)
		ORDER BY created_at DESC LIMIT 100`, actividadID).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []entities.Evidencia{}
	for rows.Next() {
		var e entities.Evidencia
		var when time.Time
		if err := rows.Scan(&e.ID, &e.ActividadID, &e.Nombre, &e.Mime, &e.Bytes, &e.Nota, &when); err != nil {
			return nil, err
		}
		e.CreatedAt = when
		out = append(out, e)
	}
	return out, rows.Err()
}

// ObtenerRutaYMime returns the file storage path and MIME type for an evidence ID.
func (r *evidenciaRepository) ObtenerRutaYMime(ctx context.Context, id string) (string, string, error) {
	var ruta, mime string
	err := r.db.WithContext(ctx).Raw(`SELECT ruta, mime FROM evidencias WHERE id = $1`, id).Row().Scan(&ruta, &mime)
	if errors.Is(err, sql.ErrNoRows) {
		return "", "", domainErrors.ErrLaborNoEncontrada
	}
	return ruta, mime, err
}

// Guardar atomically stores evidence metadata and file blob with idempotency semantics.
func (r *evidenciaRepository) Guardar(ctx context.Context, in entities.GuardarEvidencia, storage contracts.IAlmacenArchivos) (*entities.ResultadoEvidencia, error) {
	if r.db == nil || storage == nil {
		return nil, errors.New("almacén no disponible")
	}

	var out entities.ResultadoEvidencia
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 1. Verificar permiso sobre la labor
		var asignado string
		err := tx.Raw(`SELECT COALESCE(assigned_capataz_id, '') FROM actividades WHERE id = $1`, in.ActividadID).Row().Scan(&asignado)
		if errors.Is(err, sql.ErrNoRows) {
			return domainErrors.ErrLaborNoEncontrada
		}
		if err != nil {
			return err
		}
		if in.Rol == "capataz" && strings.TrimSpace(asignado) != strings.TrimSpace(in.CapatazID) {
			return domainErrors.ErrOperacionProhibido
		}

		// 2. Si hay orden_id, verificar que pertenezca a la labor
		if in.OrdenID != "" {
			var act string
			err := tx.Raw(`SELECT actividad_id::text FROM ordenes_servicio WHERE id = $1`, in.OrdenID).Row().Scan(&act)
			if errors.Is(err, sql.ErrNoRows) {
				return domainErrors.InputError{Reason: "la orden no existe"}
			}
			if err != nil {
				return err
			}
			if act != in.ActividadID {
				return domainErrors.InputError{Reason: "la orden no es de esta labor"}
			}
		}

		// 3. Chequeo de idempotencia previo por ID
		var previo sql.NullString
		scanErr := tx.Raw(`SELECT sha256 FROM evidencias WHERE id = $1`, in.ID).Row().Scan(&previo)
		if scanErr == nil {
			if previo.Valid && strings.EqualFold(previo.String, in.Hash) {
				out = entities.ResultadoEvidencia{ID: in.ID, Idempotente: true}
				return nil
			}
			return domainErrors.ErrLaborConflicto
		}
		if !errors.Is(scanErr, sql.ErrNoRows) {
			return scanErr
		}

		// 4. Guardar archivo en storage
		ref, err := storage.Put(ctx, in.Hash+in.Ext, bytes.NewReader(in.Contenido), in.Mime)
		if err != nil {
			return err
		}

		// 5. Insertar en evidencias
		res := tx.Exec(`
			INSERT INTO evidencias (id, actividad_id, orden_id, nombre, mime, bytes, ruta, nota, sha256, lat, lon, exif)
			VALUES ($1, $2, NULLIF($3, '')::uuid, $4, $5, $6, $7, $8, $9, $10, $11, $12::jsonb)`,
			in.ID, in.ActividadID, in.OrdenID, in.Nombre, in.Mime, in.Bytes, ref, in.Nota, in.Hash, in.Lat, in.Lon, in.Exif,
		)
		if res.Error != nil {
			if esDuplicado(res.Error) {
				var ya sql.NullString
				if err := tx.Raw(`SELECT sha256 FROM evidencias WHERE id = $1`, in.ID).Row().Scan(&ya); err != nil {
					return err
				}
				if ya.Valid && strings.EqualFold(ya.String, in.Hash) {
					out = entities.ResultadoEvidencia{ID: in.ID, Idempotente: true}
					return nil
				}
				return domainErrors.ErrLaborConflicto
			}
			return res.Error
		}

		// 6. Registrar en actividad_eventos
		var usuario any
		if in.UsuarioID > 0 {
			usuario = in.UsuarioID
		}
		if err := tx.Exec(`
			INSERT INTO actividad_eventos (actividad_id, tipo, actor_rol, nota, usuario_id)
			VALUES ($1, 'evidencia', $2, $3, $4)`,
			in.ActividadID, rolEvento(in.Rol), "Evidencia: "+in.Nombre, usuario,
		).Error; err != nil {
			return err
		}

		out = entities.ResultadoEvidencia{ID: in.ID, Idempotente: false}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &out, nil
}

func rolEvento(rol string) string {
	switch rol {
	case "capataz", "coordinacion", "admin", "jefatura":
		return rol
	default:
		return "sesion"
	}
}

func esDuplicado(err error) bool {
	var pg *pgconn.PgError
	if errors.As(err, &pg) && pg.Code == "23505" {
		return true
	}
	return strings.Contains(err.Error(), "duplicate key")
}
