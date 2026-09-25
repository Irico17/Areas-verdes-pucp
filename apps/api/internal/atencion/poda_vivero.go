package atencion

import (
	"context"
	"database/sql"
	"strings"
	"time"

	"campusverde/api/internal/operacion"
)

type Poda struct {
	ID                string  `json:"id"`
	Codigo            string  `json:"codigo"`
	CodigoExterno     string  `json:"codigo_externo,omitempty"`
	Tipo              string  `json:"tipo"`
	TipoActividad     string  `json:"tipo_actividad"`
	FechaReporte      string  `json:"fecha_reporte,omitempty"`
	FechaEjecucion    string  `json:"fecha_ejecucion,omitempty"`
	Personal          string  `json:"personal"`
	Ubicacion         string  `json:"ubicacion"`
	Unidad            string  `json:"unidad"`
	CantidadPedida    float64 `json:"cantidad_pedida"`
	CantidadEjecutada float64 `json:"cantidad_ejecutada"`
	Prioridad         string  `json:"prioridad"`
	Comentario        string  `json:"comentario,omitempty"`
	NombreComun       string  `json:"nombre_comun,omitempty"`
	NombreCientifico  string  `json:"nombre_cientifico,omitempty"`
}

type PodaInput struct {
	ID                string  `json:"id"`
	Codigo            string  `json:"codigo"`
	CodigoExterno     string  `json:"codigo_externo"`
	Tipo              string  `json:"tipo"`
	TipoActividad     string  `json:"tipo_actividad"`
	FechaReporte      string  `json:"fecha_reporte"`
	FechaEjecucion    string  `json:"fecha_ejecucion"`
	Personal          string  `json:"personal"`
	Ubicacion         string  `json:"ubicacion"`
	Unidad            string  `json:"unidad"`
	CantidadPedida    float64 `json:"cantidad_pedida"`
	CantidadEjecutada float64 `json:"cantidad_ejecutada"`
	Prioridad         string  `json:"prioridad"`
	Comentario        string  `json:"comentario"`
	NombreComun       string  `json:"nombre_comun"`
	NombreCientifico  string  `json:"nombre_cientifico"`
}

func validarPoda(in PodaInput) error {
	in.Codigo = strings.TrimSpace(in.Codigo)
	if len(in.Codigo) < 4 || !strings.HasPrefix(in.Codigo, "PO-") {
		return operacion.InputError{Reason: "el código de poda es PO-n"}
	}
	if !operacionUUID(in.ID) {
		return operacion.InputError{Reason: "id debe ser un UUID"}
	}
	switch strings.TrimSpace(in.Prioridad) {
	case "baja", "media", "alta":
	default:
		return operacion.InputError{Reason: "prioridad no reconocida"}
	}
	if in.CantidadPedida < 0 || in.CantidadEjecutada < 0 {
		return operacion.InputError{Reason: "las cantidades son cero o más"}
	}
	return nil
}

func (s *Store) ListarPodas(ctx context.Context) ([]Poda, error) {
	rows, err := s.db.WithContext(ctx).Raw(`
		SELECT id::text, codigo, COALESCE(codigo_externo, ''), tipo, tipo_actividad,
		       COALESCE(to_char(fecha_reporte, 'YYYY-MM-DD'), ''),
		       COALESCE(to_char(fecha_ejecucion, 'YYYY-MM-DD'), ''),
		       personal_ficticio, ubicacion, unidad, cantidad_pedida, cantidad_ejecutada,
		       prioridad, comentario, nombre_comun, nombre_cientifico
		FROM podas WHERE archivada_en IS NULL ORDER BY codigo`).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []Poda{}
	for rows.Next() {
		var p Poda
		if err := rows.Scan(&p.ID, &p.Codigo, &p.CodigoExterno, &p.Tipo, &p.TipoActividad, &p.FechaReporte, &p.FechaEjecucion, &p.Personal, &p.Ubicacion, &p.Unidad, &p.CantidadPedida, &p.CantidadEjecutada, &p.Prioridad, &p.Comentario, &p.NombreComun, &p.NombreCientifico); err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

func (s *Store) GuardarPoda(ctx context.Context, in PodaInput) (Poda, error) {
	var zero Poda
	if err := validarPoda(in); err != nil {
		return zero, err
	}
	prioridad := strings.TrimSpace(in.Prioridad)
	err := s.db.WithContext(ctx).Exec(`
		INSERT INTO podas (
		  id, codigo, codigo_externo, tipo, tipo_actividad, fecha_reporte, fecha_ejecucion,
		  personal_ficticio, ubicacion, unidad, cantidad_pedida, cantidad_ejecutada,
		  prioridad, comentario, nombre_comun, nombre_cientifico, origen_ref
		) VALUES (
		  $1, $2, NULLIF($3, ''), $4, $5, NULLIF($6, '')::date, NULLIF($7, '')::date,
		  $8, $9, $10, $11, $12, $13, $14, $15, $16, $2
		)
		ON CONFLICT (codigo) DO UPDATE SET
		  codigo_externo = EXCLUDED.codigo_externo,
		  tipo = EXCLUDED.tipo,
		  tipo_actividad = EXCLUDED.tipo_actividad,
		  fecha_reporte = EXCLUDED.fecha_reporte,
		  fecha_ejecucion = EXCLUDED.fecha_ejecucion,
		  personal_ficticio = EXCLUDED.personal_ficticio,
		  ubicacion = EXCLUDED.ubicacion,
		  unidad = EXCLUDED.unidad,
		  cantidad_pedida = EXCLUDED.cantidad_pedida,
		  cantidad_ejecutada = EXCLUDED.cantidad_ejecutada,
		  prioridad = EXCLUDED.prioridad,
		  comentario = EXCLUDED.comentario,
		  nombre_comun = EXCLUDED.nombre_comun,
		  nombre_cientifico = EXCLUDED.nombre_cientifico,
		  updated_at = now()`,
		in.ID, strings.TrimSpace(in.Codigo), strings.TrimSpace(in.CodigoExterno), strings.TrimSpace(in.Tipo),
		strings.TrimSpace(in.TipoActividad), strings.TrimSpace(in.FechaReporte), strings.TrimSpace(in.FechaEjecucion),
		strings.TrimSpace(in.Personal), strings.TrimSpace(in.Ubicacion), strings.TrimSpace(in.Unidad),
		in.CantidadPedida, in.CantidadEjecutada, prioridad, strings.TrimSpace(in.Comentario),
		strings.TrimSpace(in.NombreComun), strings.TrimSpace(in.NombreCientifico),
	).Error
	if err != nil {
		return zero, err
	}
	return Poda{
		ID: in.ID, Codigo: strings.TrimSpace(in.Codigo), CodigoExterno: strings.TrimSpace(in.CodigoExterno),
		Tipo: in.Tipo, TipoActividad: in.TipoActividad, Personal: in.Personal, Ubicacion: in.Ubicacion,
		Unidad: in.Unidad, CantidadPedida: in.CantidadPedida, CantidadEjecutada: in.CantidadEjecutada,
		Prioridad: prioridad, Comentario: in.Comentario,
	}, nil
}

func (s *Store) ArchivarPoda(ctx context.Context, id string) error {
	if !operacionUUID(id) {
		return operacion.InputError{Reason: "id debe ser un UUID"}
	}
	res := s.db.WithContext(ctx).Exec(`UPDATE podas SET archivada_en = now(), updated_at = now() WHERE id = $1 AND archivada_en IS NULL`, id)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return operacion.ErrNoEncontrada
	}
	return nil
}

type Vivero struct {
	ID            string `json:"id"`
	Fecha         string `json:"fecha,omitempty"`
	Area          string `json:"area"`
	Subproceso    string `json:"subproceso"`
	Etapa         string `json:"etapa"`
	Descripcion   string `json:"descripcion"`
	Observaciones string `json:"observaciones,omitempty"`
	Responsables  string `json:"responsables,omitempty"`
	LugarID       string `json:"lugar_id,omitempty"`
	LugarLibre    string `json:"lugar_libre,omitempty"`
}

type ViveroInput struct {
	ID            string `json:"id"`
	Fecha         string `json:"fecha"`
	Area          string `json:"area"`
	Subproceso    string `json:"subproceso"`
	Etapa         string `json:"etapa"`
	Descripcion   string `json:"descripcion"`
	Observaciones string `json:"observaciones"`
	Responsables  string `json:"responsables"`
	LugarID       string `json:"lugar_id"`
	LugarLibre    string `json:"lugar_libre"`
}

func (s *Store) ListarVivero(ctx context.Context, mes string) ([]Vivero, error) {
	mes = strings.TrimSpace(mes)
	if mes != "" && (len(mes) != 7 || mes[4] != '-') {
		return nil, operacion.InputError{Reason: "el mes usa AAAA-MM"}
	}
	q := `
		SELECT id::text, COALESCE(to_char(fecha, 'YYYY-MM-DD'), ''), area, subproceso, etapa,
		       descripcion, observaciones, responsables, COALESCE(lugar_id, ''), lugar_libre
		FROM vivero_registros
		WHERE archivada_en IS NULL`
	args := []any{}
	if mes != "" {
		q += ` AND to_char(fecha, 'YYYY-MM') = $1`
		args = append(args, mes)
	}
	q += ` ORDER BY fecha DESC NULLS LAST LIMIT 300`
	rows, err := s.db.WithContext(ctx).Raw(q, args...).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []Vivero{}
	for rows.Next() {
		var v Vivero
		if err := rows.Scan(&v.ID, &v.Fecha, &v.Area, &v.Subproceso, &v.Etapa, &v.Descripcion, &v.Observaciones, &v.Responsables, &v.LugarID, &v.LugarLibre); err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	return out, rows.Err()
}

func (s *Store) GuardarVivero(ctx context.Context, in ViveroInput) (Vivero, error) {
	var zero Vivero
	if !operacionUUID(in.ID) {
		return zero, operacion.InputError{Reason: "id debe ser un UUID"}
	}
	area := strings.TrimSpace(in.Area)
	switch area {
	case "Fauna", "Flora", "Ambiental", "Otros", "":
	default:
		return zero, operacion.InputError{Reason: "el área debe estar en el catálogo"}
	}
	err := s.db.WithContext(ctx).Exec(`
		INSERT INTO vivero_registros (
		  id, fecha, area, subproceso, etapa, descripcion, observaciones, responsables, lugar_id, lugar_libre
		) VALUES (
		  $1, NULLIF($2, '')::date, $3, $4, $5, $6, $7, $8, NULLIF($9, ''), $10
		)
		ON CONFLICT (id) DO UPDATE SET
		  fecha = EXCLUDED.fecha,
		  area = EXCLUDED.area,
		  subproceso = EXCLUDED.subproceso,
		  etapa = EXCLUDED.etapa,
		  descripcion = EXCLUDED.descripcion,
		  observaciones = EXCLUDED.observaciones,
		  responsables = EXCLUDED.responsables,
		  lugar_id = EXCLUDED.lugar_id,
		  lugar_libre = EXCLUDED.lugar_libre,
		  updated_at = now()`,
		in.ID, strings.TrimSpace(in.Fecha), area, strings.TrimSpace(in.Subproceso), strings.TrimSpace(in.Etapa),
		strings.TrimSpace(in.Descripcion), strings.TrimSpace(in.Observaciones), strings.TrimSpace(in.Responsables),
		strings.TrimSpace(in.LugarID), strings.TrimSpace(in.LugarLibre),
	).Error
	if err != nil {
		return zero, err
	}
	return Vivero{ID: in.ID, Fecha: in.Fecha, Area: area, Subproceso: in.Subproceso, Etapa: in.Etapa, Descripcion: in.Descripcion, Responsables: in.Responsables, LugarID: in.LugarID, LugarLibre: in.LugarLibre}, nil
}

func (s *Store) ArchivarVivero(ctx context.Context, id string) error {
	if !operacionUUID(id) {
		return operacion.InputError{Reason: "id debe ser un UUID"}
	}
	res := s.db.WithContext(ctx).Exec(`UPDATE vivero_registros SET archivada_en = now() WHERE id = $1`, id)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return operacion.ErrNoEncontrada
	}
	return nil
}

func (s *Store) EditarSolicitud(ctx context.Context, in SolicitudInput) (Solicitud, error) {
	var zero Solicitud
	if !operacionUUID(in.ID) {
		return zero, operacion.InputError{Reason: "id debe ser un UUID"}
	}
	in.Titulo = strings.TrimSpace(in.Titulo)
	if in.Titulo == "" {
		return zero, operacion.InputError{Reason: "el título de la solicitud es obligatorio"}
	}
	res := s.db.WithContext(ctx).Exec(`
		UPDATE solicitudes SET
		  codigo_externo = NULLIF($2, ''),
		  titulo = $3,
		  detalle = $4,
		  prioridad = $5,
		  lugar = NULLIF($6, ''),
		  updated_at = now()
		WHERE id = $1 AND archivada_en IS NULL`,
		in.ID, strings.TrimSpace(in.CodigoExterno), in.Titulo, strings.TrimSpace(in.Detalle),
		strings.TrimSpace(in.Prioridad), strings.TrimSpace(in.Lugar),
	)
	if res.Error != nil {
		return zero, res.Error
	}
	if res.RowsAffected == 0 {
		return zero, operacion.ErrNoEncontrada
	}
	return s.unaSolicitud(ctx, in.ID)
}

type OrdenInput struct {
	ID               string `json:"id"`
	Conformidad      string `json:"conformidad"`
	PeriodoInicio    string `json:"periodo_inicio"`
	PeriodoFin       string `json:"periodo_fin"`
	ReporteProveedor string `json:"reporte_proveedor"`
	Estado           string `json:"estado"`
}

func (s *Store) EditarOrden(ctx context.Context, in OrdenInput) (Orden, error) {
	var zero Orden
	if !operacionUUID(in.ID) {
		return zero, operacion.InputError{Reason: "id debe ser un UUID"}
	}
	switch strings.TrimSpace(in.Estado) {
	case "", "en_proceso", "ejecutada", "conforme":
	default:
		return zero, operacion.InputError{Reason: "estado de orden no reconocido"}
	}
	estado := strings.TrimSpace(in.Estado)
	if estado == "" {
		estado = "en_proceso"
	}
	res := s.db.WithContext(ctx).Exec(`
		UPDATE ordenes_servicio SET
		  conformidad = $2,
		  periodo_inicio = NULLIF($3, '')::date,
		  periodo_fin = NULLIF($4, '')::date,
		  reporte_proveedor = $5,
		  estado = $6
		WHERE id = $1`,
		in.ID, strings.TrimSpace(in.Conformidad), strings.TrimSpace(in.PeriodoInicio),
		strings.TrimSpace(in.PeriodoFin), strings.TrimSpace(in.ReporteProveedor), estado,
	)
	if res.Error != nil {
		return zero, res.Error
	}
	if res.RowsAffected == 0 {
		return zero, operacion.ErrNoEncontrada
	}
	var o Orden
	var when time.Time
	err := s.db.WithContext(ctx).Raw(`
		SELECT id::text, actividad_id::text, empresa, referencia, frecuencia, estado, conformidad, created_at
		FROM ordenes_servicio WHERE id = $1`, in.ID).Row().Scan(
		&o.ID, &o.ActividadID, &o.Empresa, &o.Referencia, &o.Frecuencia, &o.Estado, &o.Conformidad, &when,
	)
	if err == sql.ErrNoRows {
		return zero, operacion.ErrNoEncontrada
	}
	o.CreatedAt = when.UTC().Format(time.RFC3339)
	return o, err
}

func (s *Store) CrearAvance(ctx context.Context, actividadID, id, fecha, nota, area, ejemplar string) error {
	if !operacionUUID(actividadID) || !operacionUUID(id) {
		return operacion.InputError{Reason: "id debe ser un UUID"}
	}
	if _, err := time.Parse("2006-01-02", fecha); err != nil {
		return operacion.InputError{Reason: "la fecha usa el formato AAAA-MM-DD"}
	}
	if strings.TrimSpace(area) == "" && strings.TrimSpace(ejemplar) == "" {
		return operacion.InputError{Reason: "el avance se liga a un área o a un ejemplar"}
	}
	return s.db.WithContext(ctx).Exec(`
		INSERT INTO actividad_avances (id, actividad_id, fecha, nota, area_feature_id, ejemplar_ref)
		VALUES ($1, $2, $3::date, $4, NULLIF($5, ''), NULLIF($6, ''))`,
		id, actividadID, fecha, strings.TrimSpace(nota), strings.TrimSpace(area), strings.TrimSpace(ejemplar),
	).Error
}
