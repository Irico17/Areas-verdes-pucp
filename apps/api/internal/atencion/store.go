package atencion

import (
	"context"
	"database/sql"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"campusverde/api/internal/operacion"

	"gorm.io/gorm"
)

// Store cubre solicitudes, órdenes, riego, evidencias y el reporte básico.
type Store struct{ db *gorm.DB }

func NewStore(db *gorm.DB) *Store { return &Store{db: db} }

type Solicitud struct {
	ID            string  `json:"id"`
	CodigoExterno *string `json:"codigo_externo,omitempty"`
	Fuente        string  `json:"fuente"`
	Titulo        string  `json:"titulo"`
	Detalle       string  `json:"detalle,omitempty"`
	Prioridad     string  `json:"prioridad"`
	Estado        string  `json:"estado"`
	Lugar         *string `json:"lugar,omitempty"`
	Cantidad      *int    `json:"cantidad,omitempty"`
	ActividadID   *string `json:"actividad_id,omitempty"`
	CreatedAt     string  `json:"created_at"`
}

type SolicitudInput struct {
	ID            string `json:"id"`
	CodigoExterno string `json:"codigo_externo"`
	Fuente        string `json:"fuente"`
	Titulo        string `json:"titulo"`
	Detalle       string `json:"detalle"`
	Prioridad     string `json:"prioridad"`
	Lugar         string `json:"lugar"`
	Cantidad      int    `json:"cantidad"`
	ActividadID   string `json:"actividad_id"`
}

func (s *Store) ListarSolicitudes(ctx context.Context) ([]Solicitud, error) {
	rows, err := s.db.WithContext(ctx).Raw(`
		SELECT id::text, codigo_externo, fuente, titulo, detalle, prioridad, estado,
		       lugar, cantidad, actividad_id::text, created_at
		FROM solicitudes ORDER BY created_at DESC LIMIT 200`).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanSolicitudes(rows)
}

func (s *Store) CrearSolicitud(ctx context.Context, in SolicitudInput) (Solicitud, error) {
	var zero Solicitud
	in.Titulo = strings.TrimSpace(in.Titulo)
	in.CodigoExterno = strings.TrimSpace(in.CodigoExterno)
	in.Fuente = strings.TrimSpace(in.Fuente)
	in.Prioridad = strings.TrimSpace(in.Prioridad)
	if in.Prioridad == "" {
		in.Prioridad = "media"
	}
	if in.Titulo == "" || utf8.RuneCountInString(in.Titulo) > 160 {
		return zero, operacion.InputError{Reason: "el título de la solicitud es obligatorio"}
	}
	switch in.Fuente {
	case "centuria", "osg", "correo", "interna":
	default:
		return zero, operacion.InputError{Reason: "la fuente debe ser centuria, osg, correo o interna"}
	}
	switch in.Prioridad {
	case "baja", "media", "alta":
	default:
		return zero, operacion.InputError{Reason: "prioridad no reconocida"}
	}
	if strings.TrimSpace(in.ActividadID) != "" && !operacionUUID(in.ActividadID) {
		return zero, operacion.InputError{Reason: "actividad_id debe ser un UUID"}
	}
	if !operacionUUID(in.ID) {
		return zero, operacion.InputError{Reason: "id debe ser un UUID"}
	}
	err := s.db.WithContext(ctx).Exec(`
		INSERT INTO solicitudes (
		  id, codigo_externo, fuente, titulo, detalle, prioridad, lugar, cantidad, actividad_id
		) VALUES (
		  $1, NULLIF($2, ''), $3, $4, $5, $6, NULLIF($7, ''), NULLIF($8, 0), NULLIF($9, '')::uuid
		)`,
		in.ID, in.CodigoExterno, in.Fuente, in.Titulo, strings.TrimSpace(in.Detalle),
		in.Prioridad, strings.TrimSpace(in.Lugar), in.Cantidad, strings.TrimSpace(in.ActividadID),
	).Error
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key") || strings.Contains(err.Error(), "unique") {
			return zero, operacion.InputError{Reason: "ese código externo ya está registrado"}
		}
		return zero, err
	}
	return s.unaSolicitud(ctx, in.ID)
}

type Orden struct {
	ID          string `json:"id"`
	ActividadID string `json:"actividad_id"`
	Empresa     string `json:"empresa"`
	Referencia  string `json:"referencia"`
	Frecuencia  string `json:"frecuencia,omitempty"`
	Estado      string `json:"estado"`
	Conformidad string `json:"conformidad,omitempty"`
	CreatedAt   string `json:"created_at"`
}

func (s *Store) ListarOrdenes(ctx context.Context) ([]Orden, error) {
	rows, err := s.db.WithContext(ctx).Raw(`
		SELECT id::text, actividad_id::text, empresa, referencia, frecuencia, estado, conformidad, created_at
		FROM ordenes_servicio ORDER BY created_at DESC LIMIT 200`).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []Orden{}
	for rows.Next() {
		var o Orden
		var when time.Time
		if err := rows.Scan(&o.ID, &o.ActividadID, &o.Empresa, &o.Referencia, &o.Frecuencia, &o.Estado, &o.Conformidad, &when); err != nil {
			return nil, err
		}
		o.CreatedAt = when.UTC().Format(time.RFC3339)
		out = append(out, o)
	}
	return out, rows.Err()
}

func (s *Store) CrearOrden(ctx context.Context, id, actividadID, empresa, referencia, frecuencia string) (Orden, error) {
	var zero Orden
	empresa = strings.TrimSpace(empresa)
	referencia = strings.TrimSpace(referencia)
	if empresa == "" || referencia == "" {
		return zero, operacion.InputError{Reason: "la empresa y la referencia son obligatorias"}
	}
	if !operacionUUID(id) || !operacionUUID(actividadID) {
		return zero, operacion.InputError{Reason: "id y actividad_id deben ser UUID"}
	}
	var ejecutor string
	err := s.db.WithContext(ctx).Raw(`SELECT ejecutor FROM actividades WHERE id = $1 AND archivada_en IS NULL`, actividadID).Row().Scan(&ejecutor)
	if err == sql.ErrNoRows {
		return zero, operacion.ErrNoEncontrada
	}
	if err != nil {
		return zero, err
	}
	if ejecutor != "tercerizada" {
		return zero, operacion.InputError{Reason: "la orden solo se vincula a una labor tercerizada"}
	}
	if err := s.db.WithContext(ctx).Exec(`
		INSERT INTO ordenes_servicio (id, actividad_id, empresa, referencia, frecuencia)
		VALUES ($1, $2, $3, $4, $5)`,
		id, actividadID, empresa, referencia, strings.TrimSpace(frecuencia),
	).Error; err != nil {
		return zero, err
	}
	var o Orden
	var when time.Time
	err = s.db.WithContext(ctx).Raw(`
		SELECT id::text, actividad_id::text, empresa, referencia, frecuencia, estado, conformidad, created_at
		FROM ordenes_servicio WHERE id = $1`, id).Row().Scan(
		&o.ID, &o.ActividadID, &o.Empresa, &o.Referencia, &o.Frecuencia, &o.Estado, &o.Conformidad, &when,
	)
	o.CreatedAt = when.UTC().Format(time.RFC3339)
	return o, err
}

type Riego struct {
	ID        string `json:"id"`
	Sector    string `json:"sector"`
	Turno     string `json:"turno"`
	CapatazID string `json:"capataz_id,omitempty"`
	Equipo    string `json:"equipo,omitempty"`
	Fecha     string `json:"fecha"`
	Nota      string `json:"nota,omitempty"`
}

// consultaRiego arma el listado. Un capataz solo recibe las filas de su equipo.
func consultaRiego(capatazID string) (string, []any) {
	q := `
		SELECT r.id::text, r.sector, r.turno, COALESCE(r.capataz_id, ''), COALESCE(c.equipo, ''),
		       to_char(r.fecha, 'YYYY-MM-DD'), r.nota
		FROM riego_registros r
		LEFT JOIN capataces c ON c.id = r.capataz_id`
	var args []any
	if id := strings.TrimSpace(capatazID); id != "" {
		q += ` WHERE r.capataz_id = $1`
		args = append(args, id)
	}
	q += ` ORDER BY r.fecha DESC, r.created_at DESC LIMIT 100`
	return q, args
}

func (s *Store) ListarRiego(ctx context.Context, capatazID string) ([]Riego, error) {
	q, args := consultaRiego(capatazID)
	rows, err := s.db.WithContext(ctx).Raw(q, args...).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []Riego{}
	for rows.Next() {
		var r Riego
		if err := rows.Scan(&r.ID, &r.Sector, &r.Turno, &r.CapatazID, &r.Equipo, &r.Fecha, &r.Nota); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

func (s *Store) CrearRiego(ctx context.Context, id, sector, turno, capatazID, fecha, nota string) error {
	sector = strings.TrimSpace(sector)
	turno = strings.TrimSpace(turno)
	if sector == "" || utf8.RuneCountInString(sector) > 80 {
		return operacion.InputError{Reason: "el sector es obligatorio"}
	}
	if turno != "manana" && turno != "tarde" {
		return operacion.InputError{Reason: "el turno es mañana o tarde"}
	}
	if _, err := time.Parse("2006-01-02", fecha); err != nil {
		return operacion.InputError{Reason: "la fecha usa el formato AAAA-MM-DD"}
	}
	if !operacionUUID(id) {
		return operacion.InputError{Reason: "id debe ser un UUID"}
	}
	return s.db.WithContext(ctx).Exec(`
		INSERT INTO riego_registros (id, sector, turno, capataz_id, fecha, nota)
		VALUES ($1, $2, $3, NULLIF($4, ''), $5::date, $6)`,
		id, sector, turno, strings.TrimSpace(capatazID), fecha, strings.TrimSpace(nota),
	).Error
}

type Evidencia struct {
	ID          string `json:"id"`
	ActividadID string `json:"actividad_id,omitempty"`
	Nombre      string `json:"nombre"`
	Mime        string `json:"mime"`
	Bytes       int    `json:"bytes"`
	Nota        string `json:"nota,omitempty"`
	CreatedAt   string `json:"created_at"`
}

func (s *Store) ListarEvidencias(ctx context.Context, actividadID string) ([]Evidencia, error) {
	rows, err := s.db.WithContext(ctx).Raw(`
		SELECT id::text, COALESCE(actividad_id::text, ''), nombre, mime, bytes, nota, created_at
		FROM evidencias
		WHERE ($1 = '' OR actividad_id::text = $1)
		ORDER BY created_at DESC LIMIT 100`, actividadID).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []Evidencia{}
	for rows.Next() {
		var e Evidencia
		var when time.Time
		if err := rows.Scan(&e.ID, &e.ActividadID, &e.Nombre, &e.Mime, &e.Bytes, &e.Nota, &when); err != nil {
			return nil, err
		}
		e.CreatedAt = when.UTC().Format(time.RFC3339)
		out = append(out, e)
	}
	return out, rows.Err()
}

func (s *Store) GuardarEvidencia(ctx context.Context, id, actividadID, nombre, mime, ruta, nota string, nBytes int) error {
	if !operacionUUID(id) || !operacionUUID(actividadID) {
		return operacion.InputError{Reason: "id y actividad_id deben ser UUID"}
	}
	res := s.db.WithContext(ctx).Exec(`
		INSERT INTO evidencias (id, actividad_id, nombre, mime, bytes, ruta, nota)
		SELECT $1, $2, $3, $4, $5, $6, $7
		WHERE EXISTS (SELECT 1 FROM actividades WHERE id = $2)`,
		id, actividadID, nombre, mime, nBytes, ruta, nota,
	)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return operacion.ErrNoEncontrada
	}
	return s.db.WithContext(ctx).Exec(`
		INSERT INTO actividad_eventos (actividad_id, tipo, actor_rol, nota)
		VALUES ($1, 'evidencia', 'sesion', $2)`,
		actividadID, "Evidencia: "+nombre,
	).Error
}

func (s *Store) RutaEvidencia(ctx context.Context, id string) (string, string, error) {
	var ruta, mime string
	err := s.db.WithContext(ctx).Raw(`SELECT ruta, mime FROM evidencias WHERE id = $1`, id).Row().Scan(&ruta, &mime)
	if err == sql.ErrNoRows {
		return "", "", operacion.ErrNoEncontrada
	}
	return ruta, mime, err
}

type Fila struct {
	ID            string `json:"id"`
	Titulo        string `json:"titulo"`
	Tipo          string `json:"tipo"`
	Estado        string `json:"estado"`
	Ejecutor      string `json:"ejecutor"`
	Equipo        string `json:"equipo,omitempty"`
	Zona          string `json:"zona,omitempty"`
	CodigoExterno string `json:"codigo_externo,omitempty"`
	Fuente        string `json:"fuente,omitempty"`
	CreatedAt     string `json:"created_at"`
}

type Conteo struct {
	Estado string `json:"estado"`
	N      int    `json:"n"`
}

type Reporte struct {
	Aviso     string   `json:"aviso"`
	PorEstado []Conteo `json:"por_estado"`
	Filas     []Fila   `json:"filas"`
}

func (s *Store) Reporte(ctx context.Context, estado, desde, hasta string) (Reporte, error) {
	out := Reporte{
		Aviso:     "Conteos de labores. No son el indicador oficial de cobertura: esa definición sigue pendiente de validación.",
		PorEstado: []Conteo{},
		Filas:     []Fila{},
	}
	if err := s.db.WithContext(ctx).Raw(`
		SELECT estado, count(*)::int AS n FROM actividades
		WHERE archivada_en IS NULL GROUP BY estado ORDER BY estado`).Scan(&out.PorEstado).Error; err != nil {
		return out, err
	}
	where := []string{"1=1"}
	args := []any{}
	n := 1
	if estado != "" {
		where = append(where, "$"+strconv.Itoa(n)+" = a.estado")
		args = append(args, estado)
		n++
	}
	if desde != "" {
		if _, err := time.Parse("2006-01-02", desde); err != nil {
			return out, operacion.InputError{Reason: "desde usa AAAA-MM-DD"}
		}
		where = append(where, "a.created_at::date >= $"+strconv.Itoa(n)+"::date")
		args = append(args, desde)
		n++
	}
	if hasta != "" {
		if _, err := time.Parse("2006-01-02", hasta); err != nil {
			return out, operacion.InputError{Reason: "hasta usa AAAA-MM-DD"}
		}
		where = append(where, "a.created_at::date <= $"+strconv.Itoa(n)+"::date")
		args = append(args, hasta)
		n++
	}
	_ = n
	q := `
		SELECT a.id::text, a.titulo, a.tipo, a.estado, a.ejecutor,
		       COALESCE(c.equipo, ''), COALESCE(a.zona_feature_id, ''),
		       COALESCE(s.codigo_externo, ''), COALESCE(s.fuente, ''), a.created_at
		FROM actividades a
		LEFT JOIN capataces c ON c.id = a.assigned_capataz_id
		LEFT JOIN solicitudes s ON s.actividad_id = a.id
		WHERE ` + strings.Join(where, " AND ") + `
		ORDER BY a.created_at DESC LIMIT 300`
	rows, err := s.db.WithContext(ctx).Raw(q, args...).Rows()
	if err != nil {
		return out, err
	}
	defer rows.Close()
	for rows.Next() {
		var f Fila
		var when time.Time
		if err := rows.Scan(&f.ID, &f.Titulo, &f.Tipo, &f.Estado, &f.Ejecutor, &f.Equipo, &f.Zona, &f.CodigoExterno, &f.Fuente, &when); err != nil {
			return out, err
		}
		f.CreatedAt = when.UTC().Format(time.RFC3339)
		out.Filas = append(out.Filas, f)
	}
	return out, rows.Err()
}

func (s *Store) unaSolicitud(ctx context.Context, id string) (Solicitud, error) {
	rows, err := s.db.WithContext(ctx).Raw(`
		SELECT id::text, codigo_externo, fuente, titulo, detalle, prioridad, estado,
		       lugar, cantidad, actividad_id::text, created_at
		FROM solicitudes WHERE id = $1`, id).Rows()
	if err != nil {
		return Solicitud{}, err
	}
	defer rows.Close()
	list, err := scanSolicitudes(rows)
	if err != nil {
		return Solicitud{}, err
	}
	if len(list) == 0 {
		return Solicitud{}, operacion.ErrNoEncontrada
	}
	return list[0], nil
}

type rower interface {
	Next() bool
	Scan(dest ...any) error
	Err() error
}

func scanSolicitudes(rows rower) ([]Solicitud, error) {
	out := []Solicitud{}
	for rows.Next() {
		var item Solicitud
		var codigo, lugar, actividad sql.NullString
		var cantidad sql.NullInt64
		var when time.Time
		if err := rows.Scan(&item.ID, &codigo, &item.Fuente, &item.Titulo, &item.Detalle, &item.Prioridad, &item.Estado, &lugar, &cantidad, &actividad, &when); err != nil {
			return nil, err
		}
		item.CodigoExterno = nullStr(codigo)
		item.Lugar = nullStr(lugar)
		item.ActividadID = nullStr(actividad)
		if cantidad.Valid {
			n := int(cantidad.Int64)
			item.Cantidad = &n
		}
		item.CreatedAt = when.UTC().Format(time.RFC3339)
		out = append(out, item)
	}
	return out, rows.Err()
}

func nullStr(v sql.NullString) *string {
	if !v.Valid || strings.TrimSpace(v.String) == "" {
		return nil
	}
	s := v.String
	return &s
}

func operacionUUID(id string) bool {
	if len(id) != 36 {
		return false
	}
	for i, r := range id {
		switch i {
		case 8, 13, 18, 23:
			if r != '-' {
				return false
			}
		default:
			if (r < '0' || r > '9') && (r < 'a' || r > 'f') && (r < 'A' || r > 'F') {
				return false
			}
		}
	}
	return true
}
