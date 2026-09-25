package operacion

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"campusverde/api/internal/geojson"

	"gorm.io/gorm"
)

// Store lee y escribe labores en PostGIS.
type Store struct {
	db *gorm.DB
}

func NewStore(db *gorm.DB) *Store {
	return &Store{db: db}
}

const selectFeature = `
SELECT a.id::text, a.tipo, a.estado, a.titulo, a.detalle,
       a.area_feature_id, a.zona_feature_id, a.assigned_capataz_id, c.equipo,
       a.archivada_en IS NOT NULL, a.created_at, a.updated_at, COALESCE(a.ejecutor, 'propia'),
       COALESCE(
         ST_AsGeoJSON(a.geom, 6),
         (SELECT ST_AsGeoJSON(ST_SetSRID(ST_MakePoint(l.lon, l.lat), 4326), 6)
            FROM lugares l WHERE l.id = a.lugar_id),
         (SELECT ST_AsGeoJSON(ST_PointOnSurface(z.geom), 6)
            FROM zonas_supervision z WHERE z.id = a.zona_supervision_id)
       )
FROM actividades a
LEFT JOIN capataces c ON c.id = a.assigned_capataz_id
`

func (s *Store) Capataces(ctx context.Context) ([]Capataz, error) {
	var out []Capataz
	err := s.db.WithContext(ctx).Raw(`
		SELECT id, equipo, turno FROM capataces WHERE activo ORDER BY equipo
	`).Scan(&out).Error
	if err != nil {
		return nil, err
	}
	if out == nil {
		out = []Capataz{}
	}
	return out, nil
}

func (s *Store) List(ctx context.Context, q Query) (geojson.FeatureCollection, error) {
	fc := geojson.Collection("actividades")
	if err := ValidateQuery(q); err != nil {
		return fc, err
	}
	where := []string{"1=1"}
	args := []any{}
	n := 1
	if q.SoloAbiertas {
		where = append(where, "a.archivada_en IS NULL AND a.estado NOT IN ('cancelada', 'cerrada')")
	} else {
		where = append(where, "a.archivada_en IS NULL")
	}
	if q.Rol == RolCapataz {
		where = append(where, fmt.Sprintf("a.assigned_capataz_id = $%d", n))
		args = append(args, q.CapatazID)
		n++
	}
	if q.Estado != "" {
		where = append(where, fmt.Sprintf("a.estado = $%d", n))
		args = append(args, q.Estado)
		n++
	}
	if q.Tipo != "" {
		where = append(where, fmt.Sprintf("a.tipo = $%d", n))
		args = append(args, q.Tipo)
		n++
	}
	if q.ZonaSupervisionID != "" {
		where = append(where, fmt.Sprintf("a.zona_supervision_id = $%d", n))
		args = append(args, q.ZonaSupervisionID)
		n++
	}
	if q.CuadrillaID != "" {
		where = append(where, fmt.Sprintf("a.cuadrilla_id = $%d", n))
		args = append(args, q.CuadrillaID)
		n++
	}
	if q.Origen != "" {
		where = append(where, fmt.Sprintf("a.origen = $%d", n))
		args = append(args, q.Origen)
		n++
	}
	_ = n
	query := selectFeature + " WHERE " + strings.Join(where, " AND ") + " ORDER BY a.created_at DESC"
	rows, err := s.db.WithContext(ctx).Raw(query, args...).Rows()
	if err != nil {
		return fc, err
	}
	defer rows.Close()
	for rows.Next() {
		f, err := scanFeature(rows)
		if err != nil {
			return fc, err
		}
		fc.Features = append(fc.Features, f)
	}
	return fc, rows.Err()
}

func (s *Store) Create(ctx context.Context, in CreateInput) (geojson.Feature, bool, error) {
	var zero geojson.Feature
	if err := ValidateCreate(in); err != nil {
		return zero, false, err
	}
	in.Titulo = strings.TrimSpace(in.Titulo)
	in.Detalle = strings.TrimSpace(in.Detalle)
	in.AssignedCapatazID = strings.TrimSpace(in.AssignedCapatazID)
	in.AreaFeatureID = strings.TrimSpace(in.AreaFeatureID)
	in.ZonaFeatureID = strings.TrimSpace(in.ZonaFeatureID)

	var created bool
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		saved, found, err := loadSaved(tx, in.ID)
		if err != nil {
			return err
		}
		if found {
			if !SamePayload(saved, in) {
				return ErrConflicto
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
				return InputError{Reason: "capataz_id no existe"}
			}
		}
		var nTipo int
		if err := tx.Raw(`SELECT count(*) FROM catalogos WHERE clase = 'tipo_actividad' AND codigo = $1 AND activo`, in.Tipo).Scan(&nTipo).Error; err != nil {
			return err
		}
		if nTipo != 1 {
			return InputError{Reason: "tipo no está en el catálogo activo"}
		}
		sinPunto := strings.TrimSpace(in.LugarID) != "" || strings.TrimSpace(in.ZonaSupervisionID) != ""
		if err := tx.Exec(`
			INSERT INTO actividades (
			  id, tipo, estado, titulo, detalle, area_feature_id, zona_feature_id,
			  assigned_capataz_id, geom, ejecutor, lugar_id, zona_supervision_id
			) VALUES (
			  $1, $2, 'pendiente', $3, $4, NULLIF($5, ''), NULLIF($6, ''), NULLIF($7, ''),
			  CASE WHEN $11 THEN NULL ELSE ST_SetSRID(ST_MakePoint($8, $9), 4326) END,
			  $10,
			  CASE WHEN $12 ~ '^[0-9]+$' THEN $12::bigint ELSE NULL END,
			  (SELECT id FROM zonas_supervision WHERE codigo = NULLIF($13, '') LIMIT 1)
			)`,
			in.ID, in.Tipo, in.Titulo, in.Detalle, in.AreaFeatureID, in.ZonaFeatureID,
			in.AssignedCapatazID, in.Lon, in.Lat, ejecutorDe(in.Ejecutor), sinPunto && in.Lon == 0 && in.Lat == 0,
			strings.TrimSpace(in.LugarID), strings.TrimSpace(in.ZonaSupervisionID),
		).Error; err != nil {
			return err
		}
		if err := insertEvento(tx, in.ID, "creada", "pendiente", in.AssignedCapatazID, in.ActorRol, "Alta desde el mapa", in.UsuarioID); err != nil {
			return err
		}
		if in.AssignedCapatazID != "" {
			if err := insertEvento(tx, in.ID, "asignada", "pendiente", in.AssignedCapatazID, in.ActorRol, "Asignación en el alta", in.UsuarioID); err != nil {
				return err
			}
		}
		created = true
		return nil
	})
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key") {
			saved, found, err2 := loadSaved(s.db.WithContext(ctx), in.ID)
			if err2 != nil {
				return zero, false, err2
			}
			if found && SamePayload(saved, in) {
				f, err3 := s.one(ctx, in.ID)
				return f, false, err3
			}
			return zero, false, ErrConflicto
		}
		return zero, false, err
	}
	f, err := s.one(ctx, in.ID)
	return f, created, err
}

func (s *Store) GuardarFicha(ctx context.Context, in FichaInput) error {
	if !uuidRe.MatchString(in.ID) {
		return InputError{Reason: "id debe ser un UUID"}
	}
	if in.FechaSolicitud != "" && in.FechaAtencion != "" && in.FechaAtencion < in.FechaSolicitud {
		return InputError{Reason: "la atención no puede ser anterior a la solicitud"}
	}
	var lugarID any
	lugar := strings.TrimSpace(in.Lugar)
	if lugar != "" {
		var id int64
		err := s.db.WithContext(ctx).Raw(`SELECT COALESCE((SELECT id FROM lugares WHERE id::text = $1 LIMIT 1), 0)`, lugar).Scan(&id).Error
		if err != nil {
			return err
		}
		if id != 0 {
			lugarID = id
		}
	}
	res := s.db.WithContext(ctx).Exec(`
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
		return ErrNoEncontrada
	}
	return nil
}

func (s *Store) Assign(ctx context.Context, id, capatazID, actorRol string, usuarioID int64) (geojson.Feature, error) {
	var zero geojson.Feature
	if err := ValidateAsignacion(actorRol, capatazID); err != nil {
		return zero, err
	}
	capatazID = strings.TrimSpace(capatazID)
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		row, err := lockActividad(tx, id)
		if err != nil {
			return err
		}
		if row.archivada {
			return InputError{Reason: "la labor está archivada"}
		}
		ok, err := capatazExiste(tx, capatazID)
		if err != nil {
			return err
		}
		if !ok {
			return InputError{Reason: "capataz_id no existe"}
		}
		if row.capataz == capatazID {
			return nil
		}
		tipo := "asignada"
		nota := "Asignación"
		if row.capataz != "" {
			tipo = "reasignada"
			nota = "Reasignación"
		}
		if err := tx.Exec(`
			UPDATE actividades
			SET assigned_capataz_id = $2, updated_at = now()
			WHERE id = $1`, id, capatazID).Error; err != nil {
			return err
		}
		return insertEvento(tx, id, tipo, row.estado, capatazID, actorRol, nota, usuarioID)
	})
	if err != nil {
		return zero, err
	}
	return s.one(ctx, id)
}

func (s *Store) SetEstado(ctx context.Context, id, estado, actorRol, capatazID string, usuarioID int64) (geojson.Feature, error) {
	var zero geojson.Feature
	if err := ValidateEstado(estado, actorRol, capatazID); err != nil {
		return zero, err
	}
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		row, err := lockActividad(tx, id)
		if err != nil {
			return err
		}
		if row.archivada {
			return InputError{Reason: "la labor está archivada"}
		}
		if actorRol == RolCapataz && row.capataz != strings.TrimSpace(capatazID) {
			return ErrProhibido
		}
		if row.estado == estado {
			return nil
		}
		if estado == "cerrada" {
			var ordenes, avances int
			var fecha sql.NullTime
			if err := tx.Raw(`SELECT count(*) FROM ordenes_servicio WHERE actividad_id = $1`, id).Scan(&ordenes).Error; err != nil {
				return err
			}
			if err := tx.Raw(`SELECT fecha_atencion FROM actividades WHERE id = $1`, id).Scan(&fecha).Error; err != nil {
				return err
			}
			if err := tx.Raw(`SELECT count(*) FROM actividad_avances WHERE actividad_id = $1`, id).Scan(&avances).Error; err != nil {
				return err
			}
			if err := PuedeCerrar(row.ejecutor, ordenes > 0, fecha.Valid || avances > 0); err != nil {
				return err
			}
		}
		if err := tx.Exec(`
			UPDATE actividades SET estado = $2, updated_at = now() WHERE id = $1`, id, estado).Error; err != nil {
			return err
		}
		return insertEvento(tx, id, eventoEstado(estado), estado, row.capataz, actorRol, "Cambio de estado", usuarioID)
	})
	if err != nil {
		return zero, err
	}
	return s.one(ctx, id)
}

func (s *Store) Archive(ctx context.Context, id, actorRol, motivo string, usuarioID int64) error {
	if err := ValidateArchivo(actorRol); err != nil {
		return err
	}
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		row, err := lockActividad(tx, id)
		if err != nil {
			return err
		}
		if row.archivada {
			return nil
		}
		nota := "Baja lógica"
		motivo = strings.TrimSpace(motivo)
		if motivo != "" {
			nota = nota + ": " + motivo
		}
		if err := tx.Exec(`
			UPDATE actividades SET archivada_en = now(), motivo_archivo = NULLIF($2, ''), updated_at = now() WHERE id = $1`, id, motivo).Error; err != nil {
			return err
		}
		return insertEvento(tx, id, "archivada", row.estado, row.capataz, actorRol, nota, usuarioID)
	})
}

func (s *Store) Timeline(ctx context.Context, id string) (Timeline, error) {
	out := Timeline{ActividadID: id, Eventos: []Evento{}}
	if !uuidRe.MatchString(id) {
		return out, InputError{Reason: "id debe ser un UUID"}
	}
	var n int
	if err := s.db.WithContext(ctx).Raw(`SELECT count(*) FROM actividades WHERE id = $1`, id).Scan(&n).Error; err != nil {
		return out, err
	}
	if n == 0 {
		return out, ErrNoEncontrada
	}
	rows, err := s.db.WithContext(ctx).Raw(`
		SELECT e.id, e.tipo, e.estado, e.capataz_id, c.equipo, e.actor_rol, e.nota, e.created_at,
		       e.usuario_id, u.usuario, u.nombre
		FROM actividad_eventos e
		LEFT JOIN capataces c ON c.id = e.capataz_id
		LEFT JOIN usuarios u ON u.id = e.usuario_id
		WHERE e.actividad_id = $1
		ORDER BY e.id`, id).Rows()
	if err != nil {
		return out, err
	}
	defer rows.Close()
	for rows.Next() {
		var (
			ev      Evento
			estado  sql.NullString
			capataz sql.NullString
			equipo  sql.NullString
			nota    string
			when    time.Time
			usuario sql.NullInt64
			login   sql.NullString
			nombre  sql.NullString
		)
		if err := rows.Scan(&ev.ID, &ev.Tipo, &estado, &capataz, &equipo, &ev.ActorRol, &nota, &when, &usuario, &login, &nombre); err != nil {
			return out, err
		}
		ev.Estado = nullString(estado)
		ev.CapatazID = nullString(capataz)
		ev.Equipo = nullString(equipo)
		ev.Nota = nota
		if usuario.Valid {
			id := usuario.Int64
			ev.UsuarioID = &id
		}
		if login.Valid {
			ev.Usuario = login.String
		}
		if nombre.Valid {
			ev.Nombre = nombre.String
		}
		ev.CreatedAt = when.UTC().Format(time.RFC3339)
		out.Eventos = append(out.Eventos, ev)
	}
	return out, rows.Err()
}

func (s *Store) one(ctx context.Context, id string) (geojson.Feature, error) {
	rows, err := s.db.WithContext(ctx).Raw(selectFeature+` WHERE a.id = $1`, id).Rows()
	if err != nil {
		return geojson.Feature{}, err
	}
	defer rows.Close()
	if !rows.Next() {
		return geojson.Feature{}, ErrNoEncontrada
	}
	return scanFeature(rows)
}

type locked struct {
	estado    string
	capataz   string
	ejecutor  string
	archivada bool
}

func lockActividad(tx *gorm.DB, id string) (locked, error) {
	var row locked
	if !uuidRe.MatchString(id) {
		return row, InputError{Reason: "id debe ser un UUID"}
	}
	var cap sql.NullString
	var archivada sql.NullTime
	err := tx.Raw(`
		SELECT estado, assigned_capataz_id, archivada_en, COALESCE(ejecutor, 'propia')
		FROM actividades WHERE id = $1 FOR UPDATE`, id).Row().Scan(&row.estado, &cap, &archivada, &row.ejecutor)
	if err == sql.ErrNoRows {
		return row, ErrNoEncontrada
	}
	if err != nil {
		return row, err
	}
	if cap.Valid {
		row.capataz = cap.String
	}
	row.archivada = archivada.Valid
	return row, nil
}

func loadSaved(tx *gorm.DB, id string) (Saved, bool, error) {
	var saved Saved
	var cap, area, zona sql.NullString
	err := tx.Raw(`
		SELECT tipo, titulo, detalle, assigned_capataz_id, area_feature_id, zona_feature_id,
		       COALESCE(ST_X(geom), 0), COALESCE(ST_Y(geom), 0), created_at, COALESCE(ejecutor, 'propia')
		FROM actividades WHERE id = $1`, id).Row().Scan(
		&saved.Tipo, &saved.Titulo, &saved.Detalle, &cap, &area, &zona, &saved.Lon, &saved.Lat, &saved.CreatedAt, &saved.Ejecutor,
	)
	if err == sql.ErrNoRows {
		return Saved{}, false, nil
	}
	if err != nil {
		return Saved{}, false, err
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
	return saved, true, nil
}

func capatazExiste(tx *gorm.DB, id string) (bool, error) {
	var n int
	err := tx.Raw(`SELECT count(*) FROM capataces WHERE id = $1 AND activo`, id).Scan(&n).Error
	return n == 1, err
}

func insertEvento(tx *gorm.DB, id, tipo, estado, capataz, actor, nota string, usuarioID int64) error {
	return tx.Exec(`
		INSERT INTO actividad_eventos (actividad_id, tipo, estado, capataz_id, actor_rol, nota, usuario_id)
		VALUES ($1, $2, NULLIF($3, ''), NULLIF($4, ''), $5, $6, NULLIF($7, 0))`,
		id, tipo, estado, capataz, actor, nota, usuarioID,
	).Error
}

type scanner interface {
	Scan(dest ...any) error
}

func scanFeature(rows scanner) (geojson.Feature, error) {
	var (
		id, tipo, estado, titulo, detalle string
		area, zona, capataz, equipo       sql.NullString
		archivada                         bool
		created, updated                  time.Time
		ejecutor                          string
		geom                              sql.NullString
	)
	if err := rows.Scan(&id, &tipo, &estado, &titulo, &detalle, &area, &zona, &capataz, &equipo, &archivada, &created, &updated, &ejecutor, &geom); err != nil {
		return geojson.Feature{}, err
	}
	props := ActividadProperties{
		ID:                id,
		Tipo:              tipo,
		Estado:            estado,
		Titulo:            titulo,
		Detalle:           detalle,
		AreaFeatureID:     nullString(area),
		ZonaFeatureID:     nullString(zona),
		AssignedCapatazID: nullString(capataz),
		Equipo:            nullString(equipo),
		Ejecutor:          ejecutor,
		Archivada:         archivada,
		CreatedAt:         created.UTC().Format(time.RFC3339),
		UpdatedAt:         updated.UTC().Format(time.RFC3339),
	}
	raw := json.RawMessage("null")
	if geom.Valid && strings.TrimSpace(geom.String) != "" {
		raw = json.RawMessage(geom.String)
	}
	if !json.Valid(raw) {
		return geojson.Feature{}, fmt.Errorf("geometría inválida para %s", id)
	}
	return geojson.Feature{Type: "Feature", ID: id, Geometry: raw, Properties: props}, nil
}

func nullString(v sql.NullString) *string {
	if !v.Valid || strings.TrimSpace(v.String) == "" {
		return nil
	}
	s := v.String
	return &s
}
