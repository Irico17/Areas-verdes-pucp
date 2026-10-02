// Package postgres implements database repositories for PostgreSQL/PostGIS.
package postgres

import (
	"context"
	"database/sql"
	"strings"
	"time"

	"gorm.io/gorm"

	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/contracts"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/usecases"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/entities"
	domainErrors "github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/errors"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/persistence/mapper"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/persistence/models"
)

const selectActividadFeature = `
SELECT a.id::text, a.tipo, a.estado, a.titulo, a.detalle,
       a.area_feature_id, a.zona_feature_id, a.assigned_capataz_id, c.equipo,
       a.archivada_en IS NOT NULL, a.created_at, a.updated_at, COALESCE(a.ejecutor, 'propia'),
       COALESCE(
         ST_AsGeoJSON(a.geom, 6),
         (SELECT ST_AsGeoJSON(ST_SetSRID(ST_MakePoint(l.lon, l.lat), 4326), 6)
            FROM lugares l WHERE l.id = a.lugar_id),
         (SELECT ST_AsGeoJSON(ST_PointOnSurface(z.geom), 6)
            FROM zonas_supervision z WHERE z.id = a.zona_supervision_id)
       ),
       COALESCE(a.origen, 'interna'),
       a.codigo_externo,
       a.unidad_solicitante,
       a.nivel_riesgo,
       to_char(a.fecha_programada, 'YYYY-MM-DD'),
       a.cantidad::float8,
       a.subtipo,
       a.clase_codigo,
       COALESCE((
         SELECT json_agg(p.nombre_ficticio ORDER BY p.nombre_ficticio)::text
         FROM personal_labor p WHERE p.actividad_id = a.id
       ), '[]')
FROM actividades a
LEFT JOIN capataces c ON c.id = a.assigned_capataz_id
`

type intervencionRepository struct {
	db *gorm.DB
}

// NewIntervencionRepository creates a new IIntervencionRepository instance.
func NewIntervencionRepository(db *gorm.DB) contracts.IIntervencionRepository {
	return &intervencionRepository{db: db}
}

func (r *intervencionRepository) Capataces(ctx context.Context) ([]entities.Capataz, error) {
	var dbModels []models.CapatazModel
	err := r.db.WithContext(ctx).Raw(`
		SELECT id, equipo, turno, activo FROM capataces WHERE activo ORDER BY equipo
	`).Scan(&dbModels).Error
	if err != nil {
		return nil, err
	}
	out := make([]entities.Capataz, len(dbModels))
	for i, m := range dbModels {
		out[i] = mapper.CapatazModelToEntity(m)
	}
	return out, nil
}

func (r *intervencionRepository) List(ctx context.Context, f entities.FiltroIntervenciones) (entities.FeatureCollection, error) {
	fc := entities.Collection("actividades")
	where, args := clausulasActividades(f)
	query := selectActividadFeature + " WHERE " + where + " ORDER BY a.created_at DESC"
	rows, err := r.db.WithContext(ctx).Raw(query, args...).Rows()
	if err != nil {
		return fc, err
	}
	defer rows.Close()

	for rows.Next() {
		feat, err := mapper.ScanFeature(rows)
		if err != nil {
			return fc, err
		}
		fc.Features = append(fc.Features, feat)
	}
	return fc, rows.Err()
}

func (r *intervencionRepository) Create(ctx context.Context, in entities.NuevaIntervencion) (entities.Feature, bool, error) {
	var zero entities.Feature
	in.Titulo = strings.TrimSpace(in.Titulo)
	in.Detalle = strings.TrimSpace(in.Detalle)
	in.AssignedCapatazID = strings.TrimSpace(in.AssignedCapatazID)
	in.AreaFeatureID = strings.TrimSpace(in.AreaFeatureID)
	in.ZonaFeatureID = strings.TrimSpace(in.ZonaFeatureID)

	var created bool
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		saved, found, err := loadSavedActividad(tx, in.ID)
		if err != nil {
			return err
		}
		if found {
			if !usecases.SamePayload(saved, in) {
				return domainErrors.ErrLaborConflicto
			}
			created = false
			return nil
		}
		if in.AssignedCapatazID != "" {
			ok, err := capatazExiste(tx, in.AssignedCapatazID)
			if err != nil {
				return err
			}
			if !ok {
				return domainErrors.InputError{Reason: "capataz_id no existe"}
			}
		}
		var nTipo int
		if err := tx.Raw(`SELECT count(*) FROM catalogos WHERE clase = 'tipo_actividad' AND codigo = $1 AND activo`, in.Tipo).Scan(&nTipo).Error; err != nil {
			return err
		}
		if nTipo != 1 {
			return domainErrors.InputError{Reason: "tipo no está en el catálogo activo"}
		}
		if err := validarAltaCatalogo(tx, in); err != nil {
			return err
		}
		sinPunto := strings.TrimSpace(in.LugarID) != "" || strings.TrimSpace(in.ZonaSupervisionID) != ""
		var cantidad any
		if in.Cantidad != nil {
			cantidad = *in.Cantidad
		}
		if err := tx.Exec(`
			INSERT INTO actividades (
			  id, tipo, estado, titulo, detalle, area_feature_id, zona_feature_id,
			  assigned_capataz_id, geom, ejecutor, lugar_id, zona_supervision_id,
			  origen, codigo_externo, unidad_solicitante, nivel_riesgo, fecha_programada,
			  cantidad, subtipo, clase_codigo
			) VALUES (
			  $1, $2, 'pendiente', $3, $4, NULLIF($5, ''), NULLIF($6, ''), NULLIF($7, ''),
			  CASE WHEN $11 THEN NULL ELSE ST_SetSRID(ST_MakePoint($8, $9), 4326) END,
			  $10,
			  (CASE WHEN $12 ~ '^[0-9]+$' THEN $12 END)::bigint,
			  (SELECT id FROM zonas_supervision WHERE codigo = NULLIF($13, '') LIMIT 1),
			  COALESCE(NULLIF($14, ''), 'interna'),
			  NULLIF($15, ''), NULLIF($16, ''), NULLIF($17, ''), NULLIF($18, '')::date,
			  $19, NULLIF($20, ''), NULLIF($21, '')
			)`,
			in.ID, in.Tipo, in.Titulo, in.Detalle, in.AreaFeatureID, in.ZonaFeatureID,
			in.AssignedCapatazID, in.Lon, in.Lat, in.Ejecutor, sinPunto && in.Lon == 0 && in.Lat == 0,
			strings.TrimSpace(in.LugarID), strings.TrimSpace(in.ZonaSupervisionID),
			strings.TrimSpace(in.Origen), strings.TrimSpace(in.CodigoExterno), strings.TrimSpace(in.UnidadSolicitante),
			strings.TrimSpace(in.NivelRiesgo), strings.TrimSpace(in.FechaProgramada), cantidad,
			strings.TrimSpace(in.Subtipo), strings.TrimSpace(in.Clase),
		).Error; err != nil {
			return err
		}
		if err := insertarPersonal(tx, in.ID, in.Personal); err != nil {
			return err
		}
		if err := insertEventoDB(tx, in.ID, "creada", "pendiente", in.AssignedCapatazID, in.ActorRol, "Alta desde el mapa", in.UsuarioID); err != nil {
			return err
		}
		if in.AssignedCapatazID != "" {
			if err := insertEventoDB(tx, in.ID, "asignada", "pendiente", in.AssignedCapatazID, in.ActorRol, "Asignación en el alta", in.UsuarioID); err != nil {
				return err
			}
		}
		created = true
		return nil
	})

	if err != nil {
		if strings.Contains(err.Error(), "duplicate key") {
			saved, found, err2 := loadSavedActividad(r.db.WithContext(ctx), in.ID)
			if err2 != nil {
				return zero, false, err2
			}
			if found && usecases.SamePayload(saved, in) {
				f, err3 := r.One(ctx, in.ID)
				return f, false, err3
			}
			return zero, false, domainErrors.ErrLaborConflicto
		}
		return zero, false, err
	}

	f, err := r.One(ctx, in.ID)
	return f, created, err
}

func (r *intervencionRepository) Assign(ctx context.Context, in entities.AsignarIntervencion) (entities.Feature, error) {
	var zero entities.Feature
	capatazID := strings.TrimSpace(in.CapatazID)

	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		row, err := lockActividadDB(tx, in.ID)
		if err != nil {
			return err
		}
		if row.Archivada {
			return domainErrors.InputError{Reason: "la labor está archivada"}
		}
		ok, err := capatazExiste(tx, capatazID)
		if err != nil {
			return err
		}
		if !ok {
			return domainErrors.InputError{Reason: "capataz_id no existe"}
		}
		if row.Capataz == capatazID {
			return nil
		}
		tipo := "asignada"
		nota := "Asignación"
		if row.Capataz != "" {
			tipo = "reasignada"
			nota = "Reasignación"
		}
		if err := tx.Exec(`
			UPDATE actividades
			SET assigned_capataz_id = $2, updated_at = now()
			WHERE id = $1`, in.ID, capatazID).Error; err != nil {
			return err
		}
		return insertEventoDB(tx, in.ID, tipo, row.Estado, capatazID, in.ActorRol, nota, in.UsuarioID)
	})
	if err != nil {
		return zero, err
	}
	return r.One(ctx, in.ID)
}

func (r *intervencionRepository) SetEstado(ctx context.Context, in entities.CambiarEstadoIntervencion) (entities.Feature, error) {
	var zero entities.Feature

	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		row, err := lockActividadDB(tx, in.ID)
		if err != nil {
			return err
		}
		if row.Archivada {
			return domainErrors.InputError{Reason: "la labor está archivada"}
		}
		if in.ActorRol == usecases.RolCapataz && row.Capataz != strings.TrimSpace(in.CapatazID) {
			return domainErrors.ForbiddenError{Reason: "el capataz solo puede modificar sus propias labores"}
		}
		if in.ActorRol == usecases.RolCapataz && (row.Estado == "cerrada" || row.Estado == "cancelada") {
			return domainErrors.ForbiddenError{Reason: "el capataz no puede modificar labores cerradas o canceladas"}
		}
		if row.Estado == in.Estado {
			return nil
		}
		if in.Estado == "cerrada" {
			var ordenes, avances int
			var fecha sql.NullTime
			if err := tx.Raw(`SELECT count(*) FROM ordenes_servicio WHERE actividad_id = $1`, in.ID).Scan(&ordenes).Error; err != nil {
				return err
			}
			if err := tx.Raw(`SELECT fecha_atencion FROM actividades WHERE id = $1`, in.ID).Scan(&fecha).Error; err != nil {
				return err
			}
			if err := tx.Raw(`SELECT count(*) FROM actividad_avances WHERE actividad_id = $1`, in.ID).Scan(&avances).Error; err != nil {
				return err
			}
			if err := usecases.PuedeCerrar(row.Ejecutor, ordenes > 0, fecha.Valid || avances > 0); err != nil {
				return err
			}
		}
		if err := tx.Exec(`
			UPDATE actividades SET estado = $2, updated_at = now() WHERE id = $1`, in.ID, in.Estado).Error; err != nil {
			return err
		}
		evTipo := "estado"
		if in.Estado == "cancelada" {
			evTipo = "cancelada"
		}
		return insertEventoDB(tx, in.ID, evTipo, in.Estado, row.Capataz, in.ActorRol, "Cambio de estado", in.UsuarioID)
	})
	if err != nil {
		return zero, err
	}
	return r.One(ctx, in.ID)
}

func (r *intervencionRepository) EstadoActual(ctx context.Context, id string) (string, error) {
	var estado string
	err := r.db.WithContext(ctx).Raw(`SELECT estado FROM actividades WHERE id = $1`, id).Row().Scan(&estado)
	if err == sql.ErrNoRows {
		return "", domainErrors.ErrLaborNoEncontrada
	}
	if err != nil {
		return "", err
	}
	return estado, nil
}

func (r *intervencionRepository) Archive(ctx context.Context, in entities.ArchivarIntervencion) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		row, err := lockActividadDB(tx, in.ID)
		if err != nil {
			return err
		}
		if row.Archivada {
			return nil
		}
		nota := "Baja lógica"
		motivo := strings.TrimSpace(in.Motivo)
		if motivo != "" {
			nota = nota + ": " + motivo
		}
		if err := tx.Exec(`
			UPDATE actividades SET archivada_en = now(), motivo_archivo = NULLIF($2, ''), updated_at = now() WHERE id = $1`, in.ID, motivo).Error; err != nil {
			return err
		}
		return insertEventoDB(tx, in.ID, "archivada", row.Estado, row.Capataz, in.ActorRol, nota, in.UsuarioID)
	})
}

func (r *intervencionRepository) Timeline(ctx context.Context, id string) ([]entities.ActividadEvento, error) {
	var n int
	if err := r.db.WithContext(ctx).Raw(`SELECT count(*) FROM actividades WHERE id = $1`, id).Scan(&n).Error; err != nil {
		return nil, err
	}
	if n == 0 {
		return nil, domainErrors.ErrLaborNoEncontrada
	}
	rows, err := r.db.WithContext(ctx).Raw(`
		SELECT e.id, e.tipo, e.estado, e.capataz_id, c.equipo, e.actor_rol, e.nota, e.created_at,
		       e.usuario_id, u.usuario, u.nombre, e.uuid_cliente
		FROM actividad_eventos e
		LEFT JOIN capataces c ON c.id = e.capataz_id
		LEFT JOIN usuarios u ON u.id = e.usuario_id
		WHERE e.actividad_id = $1
		ORDER BY e.id`, id).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []entities.ActividadEvento{}
	for rows.Next() {
		var (
			ev      entities.ActividadEvento
			estado  sql.NullString
			capataz sql.NullString
			equipo  sql.NullString
			nota    string
			when    time.Time
			usuario sql.NullInt64
			login   sql.NullString
			nombre  sql.NullString
			uuidCli sql.NullString
		)
		if err := rows.Scan(&ev.ID, &ev.Tipo, &estado, &capataz, &equipo, &ev.ActorRol, &nota, &when, &usuario, &login, &nombre, &uuidCli); err != nil {
			return nil, err
		}
		ev.ActividadID = id
		ev.Estado = mapper.NullString(estado)
		ev.CapatazID = mapper.NullString(capataz)
		ev.Equipo = mapper.NullString(equipo)
		ev.Nota = nota
		if usuario.Valid {
			uID := usuario.Int64
			ev.UsuarioID = &uID
		}
		if login.Valid {
			ev.Usuario = login.String
		}
		if nombre.Valid {
			ev.Nombre = nombre.String
		}
		if uuidCli.Valid {
			idCli := uuidCli.String
			ev.UUIDCliente = &idCli
		}
		ev.CreatedAt = when
		out = append(out, ev)
	}
	return out, rows.Err()
}

func (r *intervencionRepository) GuardarFicha(ctx context.Context, in entities.FichaIntervencion) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		row, err := lockActividadDB(tx, in.ID)
		if err != nil {
			return err
		}
		if row.Archivada {
			return domainErrors.InputError{Reason: "la labor está archivada"}
		}
		if in.ActorRol == usecases.RolCapataz && row.Capataz != strings.TrimSpace(in.CapatazID) {
			return domainErrors.ForbiddenError{Reason: "el capataz solo puede editar la ficha de sus propias labores"}
		}
		var lugarID any
		lugar := strings.TrimSpace(in.Lugar)
		if lugar != "" {
			var id int64
			err := tx.Raw(`SELECT COALESCE((SELECT id FROM lugares WHERE id::text = $1 LIMIT 1), 0)`, lugar).Scan(&id).Error
			if err != nil {
				return err
			}
			if id != 0 {
				lugarID = id
			}
		}
		res := tx.Exec(`
			UPDATE actividades SET
			  clase_codigo = NULLIF($2, ''),
			  fecha_solicitud = NULLIF($3, '')::date,
			  fecha_atencion = NULLIF($4, '')::date,
			  lugar_id = $5,
			  lugar_libre = $6,
			  comentario = $7,
			  updated_at = now()
			WHERE id = $1 AND archivada_en IS NULL`,
			in.ID, strings.TrimSpace(in.Clase), strings.TrimSpace(in.FechaSolicitud), strings.TrimSpace(in.FechaAtencion),
			lugarID, lugar, strings.TrimSpace(in.Comentario),
		)
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return domainErrors.ErrLaborNoEncontrada
		}
		return nil
	})
}

func (r *intervencionRepository) CrearAvance(ctx context.Context, in entities.NuevoAvance) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		row, err := lockActividadDB(tx, in.ActividadID)
		if err != nil {
			return err
		}
		if row.Archivada {
			return domainErrors.InputError{Reason: "la labor está archivada"}
		}
		if in.ActorRol == usecases.RolCapataz && row.Capataz != strings.TrimSpace(in.CapatazID) {
			return domainErrors.ForbiddenError{Reason: "el capataz solo puede registrar avances en sus propias labores"}
		}
		return tx.Exec(`
			INSERT INTO actividad_avances (id, actividad_id, fecha, nota, area_feature_id, ejemplar_ref)
			VALUES ($1, $2, $3::date, $4, NULLIF($5, ''), NULLIF($6, ''))`,
			in.ID, in.ActividadID, in.Fecha, strings.TrimSpace(in.Nota), strings.TrimSpace(in.AreaFeatureID), strings.TrimSpace(in.EjemplarRef),
		).Error
	})
}

func (r *intervencionRepository) One(ctx context.Context, id string) (entities.Feature, error) {
	rows, err := r.db.WithContext(ctx).Raw(selectActividadFeature+` WHERE a.id = $1`, id).Rows()
	if err != nil {
		return entities.Feature{}, err
	}
	defer rows.Close()
	if !rows.Next() {
		return entities.Feature{}, domainErrors.ErrLaborNoEncontrada
	}
	return mapper.ScanFeature(rows)
}

type lockedActividad struct {
	Estado    string
	Capataz   string
	Ejecutor  string
	Archivada bool
}

func lockActividadDB(tx *gorm.DB, id string) (lockedActividad, error) {
	var row lockedActividad
	var cap sql.NullString
	var archivada sql.NullTime
	err := tx.Raw(`
		SELECT estado, assigned_capataz_id, archivada_en, COALESCE(ejecutor, 'propia')
		FROM actividades WHERE id = $1 FOR UPDATE`, id).Row().Scan(&row.Estado, &cap, &archivada, &row.Ejecutor)
	if err == sql.ErrNoRows {
		return row, domainErrors.ErrLaborNoEncontrada
	}
	if err != nil {
		return row, err
	}
	if cap.Valid {
		row.Capataz = cap.String
	}
	row.Archivada = archivada.Valid
	return row, nil
}

func loadSavedActividad(tx *gorm.DB, id string) (usecases.SavedPayload, bool, error) {
	var saved usecases.SavedPayload
	var cap, area, zona sql.NullString
	var cantidad sql.NullFloat64
	var origen, codigo, unidad, riesgo, fecha, subtipo, clase, lugar, zonaCod, personal string
	err := tx.Raw(`
		SELECT tipo, titulo, detalle, assigned_capataz_id, area_feature_id, zona_feature_id,
		       COALESCE(ST_X(geom), 0), COALESCE(ST_Y(geom), 0), created_at, COALESCE(ejecutor, 'propia'),
		       COALESCE(origen, 'interna'), COALESCE(codigo_externo, ''), COALESCE(unidad_solicitante, ''),
		       COALESCE(nivel_riesgo, ''), COALESCE(to_char(fecha_programada, 'YYYY-MM-DD'), ''),
		       cantidad::float8, COALESCE(subtipo, ''), COALESCE(clase_codigo, ''),
		       COALESCE(lugar_id::text, ''),
		       COALESCE((SELECT z.codigo FROM zonas_supervision z WHERE z.id = actividades.zona_supervision_id), ''),
		       COALESCE((SELECT string_agg(p.nombre_ficticio, '|' ORDER BY p.nombre_ficticio) FROM personal_labor p WHERE p.actividad_id = actividades.id), '')
		FROM actividades WHERE id = $1`, id).Row().Scan(
		&saved.Tipo, &saved.Titulo, &saved.Detalle, &cap, &area, &zona, &saved.Lon, &saved.Lat, &saved.CreatedAt, &saved.Ejecutor,
		&origen, &codigo, &unidad, &riesgo, &fecha, &cantidad, &subtipo, &clase, &lugar, &zonaCod, &personal,
	)
	if err == sql.ErrNoRows {
		return usecases.SavedPayload{}, false, nil
	}
	if err != nil {
		return usecases.SavedPayload{}, false, err
	}
	if cap.Valid {
		saved.AssignedCapatazID = cap.String
	}
	if area.Valid {
		saved.AreaFeatureID = area.String
	}
	if zona.Valid {
		saved.ZonaFeatureID = zona.String
	}
	saved.Origen = origen
	saved.CodigoExterno = codigo
	saved.UnidadSolicitante = unidad
	saved.NivelRiesgo = riesgo
	saved.FechaProgramada = fecha
	if cantidad.Valid {
		n := cantidad.Float64
		saved.Cantidad = &n
	}
	saved.Subtipo = subtipo
	saved.Clase = clase
	saved.LugarID = lugar
	saved.ZonaSupervisionID = zonaCod
	if personal != "" {
		saved.Personal = strings.Split(personal, "|")
	}
	return saved, true, nil
}

func capatazExiste(tx *gorm.DB, id string) (bool, error) {
	var n int
	err := tx.Raw(`SELECT count(*) FROM capataces WHERE id = $1 AND activo`, id).Scan(&n).Error
	return n == 1, err
}

func insertEventoDB(tx *gorm.DB, id, tipo, estado, capataz, actor, nota string, usuarioID int64) error {
	return tx.Exec(`
		INSERT INTO actividad_eventos (actividad_id, tipo, estado, capataz_id, actor_rol, nota, usuario_id)
		VALUES ($1, $2, NULLIF($3, ''), NULLIF($4, ''), $5, $6, NULLIF($7, 0))`,
		id, tipo, estado, capataz, actor, nota, usuarioID,
	).Error
}
