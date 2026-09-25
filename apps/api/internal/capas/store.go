package capas

import (
	"context"
	"encoding/csv"
	"fmt"
	"strconv"
	"strings"

	"gorm.io/gorm"
)

// Store edita las capas del frente 2B. La baja es lógica.
type Store struct {
	db *gorm.DB
}

func NewStore(db *gorm.DB) *Store { return &Store{db: db} }

// Tacho es un punto de residuos con los 11 conteos.
type Tacho struct {
	ID                  int64    `json:"id"`
	Codigo              string   `json:"codigo"`
	Lat                 *float64 `json:"lat,omitempty"`
	Lon                 *float64 `json:"lon,omitempty"`
	Nota                string   `json:"nota"`
	Lugar               string   `json:"lugar"`
	Espacios            string   `json:"espacios"`
	Accion              string   `json:"accion"`
	TachoActual         string   `json:"tacho_actual"`
	TachoNuevo          string   `json:"tacho_nuevo"`
	Recomendaciones     string   `json:"recomendaciones"`
	NoAprovechables     int      `json:"no_aprovechables"`
	PapelCarton         int      `json:"papel_carton"`
	Plastico            int      `json:"plastico"`
	Vidrio              int      `json:"vidrio"`
	Pilas               int      `json:"pilas"`
	Peligrosos          int      `json:"peligrosos"`
	RAEE                int      `json:"raee"`
	Metales             int      `json:"metales"`
	Aniquem             int      `json:"aniquem"`
	IntermediosPlastico int      `json:"intermedios_plastico"`
	IntermediosMetal    int      `json:"intermedios_metal"`
	Activo              bool     `json:"activo"`
}

// Bebedero tiene estado y sede propios.
type Bebedero struct {
	ID      int64    `json:"id"`
	Codigo  string   `json:"codigo"`
	Subtipo string   `json:"subtipo"`
	Estado  string   `json:"estado"`
	Sede    string   `json:"sede"`
	Lat     *float64 `json:"lat,omitempty"`
	Lon     *float64 `json:"lon,omitempty"`
	Activo  bool     `json:"activo"`
}

// Punto es un lugar público. No tiene teléfono ni placeId.
type Punto struct {
	ID     int64   `json:"id"`
	Titulo string  `json:"titulo"`
	Lat    float64 `json:"lat"`
	Lon    float64 `json:"lon"`
	URL    string  `json:"url"`
	Activo bool    `json:"activo"`
}

// Reserva es ficticia mientras la hoja institucional responda 401.
type Reserva struct {
	ID         int64  `json:"id"`
	JardinID   *int64 `json:"jardin_id,omitempty"`
	Fecha      string `json:"fecha"`
	HoraInicio string `json:"hora_inicio"`
	HoraFin    string `json:"hora_fin"`
	Estado     string `json:"estado"`
	Evento     string `json:"evento"`
	Unidad     string `json:"unidad"`
	Origen     string `json:"origen"`
	Activo     bool   `json:"activo"`
}

// Ficha es una capa de la sección 4.13.
type Ficha struct {
	ID         int64    `json:"id"`
	FeatureID  string   `json:"feature_id"`
	Nombre     string   `json:"nombre"`
	Codigo     string   `json:"codigo"`
	Nota       string   `json:"nota"`
	Clase      string   `json:"clase"`
	Riego      string   `json:"riego"`
	AreaM2     *float64 `json:"area_m2,omitempty"`
	PerimetroM *float64 `json:"perimetro_m,omitempty"`
	Pertenecen string   `json:"pertenecen"`
	Uso        string   `json:"uso"`
	GeoJSON    string   `json:"geojson,omitempty"`
	Activo     bool     `json:"activo"`
}

var CapasEditables = []string{"fauna", "puertas", "playas_estacionamiento", "veredas_riesgo", "xerofiticas", "jardines_reserva"}

func (s *Store) ListarTachos(ctx context.Context) ([]Tacho, error) {
	rows, err := s.db.WithContext(ctx).Raw(`
		SELECT id, codigo, lat, lon, COALESCE(nota,''), COALESCE(lugar,''), COALESCE(espacios,''),
		       COALESCE(accion,''), COALESCE(tacho_actual,''), COALESCE(tacho_nuevo,''), COALESCE(recomendaciones,''),
		       no_aprovechables, papel_carton, plastico, vidrio, pilas, peligrosos, raee, metales, aniquem,
		       intermedios_plastico, intermedios_metal, activo
		FROM tachos WHERE activo ORDER BY codigo`).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []Tacho{}
	for rows.Next() {
		var t Tacho
		if err := rows.Scan(&t.ID, &t.Codigo, &t.Lat, &t.Lon, &t.Nota, &t.Lugar, &t.Espacios, &t.Accion, &t.TachoActual, &t.TachoNuevo, &t.Recomendaciones, &t.NoAprovechables, &t.PapelCarton, &t.Plastico, &t.Vidrio, &t.Pilas, &t.Peligrosos, &t.RAEE, &t.Metales, &t.Aniquem, &t.IntermediosPlastico, &t.IntermediosMetal, &t.Activo); err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	return out, rows.Err()
}

func ConteosNoNegativos(t Tacho) bool {
	for _, n := range []int{t.NoAprovechables, t.PapelCarton, t.Plastico, t.Vidrio, t.Pilas, t.Peligrosos, t.RAEE, t.Metales, t.Aniquem, t.IntermediosPlastico, t.IntermediosMetal} {
		if n < 0 {
			return false
		}
	}
	return strings.HasPrefix(strings.TrimSpace(t.Codigo), "PT")
}

func (s *Store) GuardarTacho(ctx context.Context, t Tacho) (Tacho, error) {
	if !ConteosNoNegativos(t) {
		return Tacho{}, fmt.Errorf("tacho inválido")
	}
	var id int64
	err := s.db.WithContext(ctx).Raw(`
		INSERT INTO tachos (
		  codigo, lat, lon, nota, lugar, espacios, accion, tacho_actual, tacho_nuevo, recomendaciones,
		  no_aprovechables, papel_carton, plastico, vidrio, pilas, peligrosos, raee, metales, aniquem,
		  intermedios_plastico, intermedios_metal, geom, origen_ref
		) VALUES (
		  $1,$2,$3,NULLIF($4,''),NULLIF($5,''),NULLIF($6,''),NULLIF($7,''),NULLIF($8,''),NULLIF($9,''),NULLIF($10,''),
		  $11,$12,$13,$14,$15,$16,$17,$18,$19,$20,$21,
		  CASE WHEN $2::float8 IS NULL THEN NULL ELSE ST_SetSRID(ST_MakePoint($3,$2),4326) END,
		  'manual:' || $1
		)
		ON CONFLICT (origen_ref) DO UPDATE SET
		  nota = EXCLUDED.nota, lugar = EXCLUDED.lugar, recomendaciones = EXCLUDED.recomendaciones,
		  no_aprovechables = EXCLUDED.no_aprovechables, papel_carton = EXCLUDED.papel_carton,
		  plastico = EXCLUDED.plastico, vidrio = EXCLUDED.vidrio, pilas = EXCLUDED.pilas,
		  peligrosos = EXCLUDED.peligrosos, raee = EXCLUDED.raee, metales = EXCLUDED.metales,
		  aniquem = EXCLUDED.aniquem, intermedios_plastico = EXCLUDED.intermedios_plastico,
		  intermedios_metal = EXCLUDED.intermedios_metal, geom = EXCLUDED.geom, updated_at = now()
		RETURNING id`,
		strings.TrimSpace(t.Codigo), t.Lat, t.Lon, t.Nota, t.Lugar, t.Espacios, t.Accion, t.TachoActual, t.TachoNuevo, t.Recomendaciones,
		t.NoAprovechables, t.PapelCarton, t.Plastico, t.Vidrio, t.Pilas, t.Peligrosos, t.RAEE, t.Metales, t.Aniquem, t.IntermediosPlastico, t.IntermediosMetal,
	).Scan(&id).Error
	if err != nil {
		return Tacho{}, err
	}
	t.ID = id
	t.Activo = true
	return t, nil
}

func (s *Store) Baja(ctx context.Context, tabla string, id int64) error {
	if !tablaPermitida(tabla) || id < 1 {
		return fmt.Errorf("baja")
	}
	res := s.db.WithContext(ctx).Exec(fmt.Sprintf(`UPDATE %s SET activo = FALSE, updated_at = now() WHERE id = $1 AND activo`, tabla), id)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected != 1 {
		return fmt.Errorf("no encontrado")
	}
	return nil
}

func tablaPermitida(tabla string) bool {
	switch tabla {
	case "tachos", "bebederos", "puntos_pucp", "reservas_jardin", "fauna", "puertas", "playas_estacionamiento", "veredas_riesgo", "xerofiticas", "jardines_reserva":
		return true
	default:
		return false
	}
}

func (s *Store) ListarBebederos(ctx context.Context) ([]Bebedero, error) {
	var out []Bebedero
	err := s.db.WithContext(ctx).Raw(`
		SELECT id, codigo, subtipo, estado, COALESCE(sede,''), lat, lon, activo
		FROM bebederos WHERE activo ORDER BY codigo`).Scan(&out).Error
	if out == nil {
		out = []Bebedero{}
	}
	return out, err
}

func (s *Store) GuardarBebedero(ctx context.Context, b Bebedero) (Bebedero, error) {
	switch b.Subtipo {
	case "fuente", "llenador", "nuevo", "deterioro", "baja":
	default:
		return Bebedero{}, fmt.Errorf("subtipo")
	}
	if b.Estado == "" || !strings.HasPrefix(b.Codigo, "PT_") {
		return Bebedero{}, fmt.Errorf("bebedero")
	}
	var id int64
	err := s.db.WithContext(ctx).Raw(`
		INSERT INTO bebederos (codigo, subtipo, estado, sede, lat, lon, geom, origen_ref)
		VALUES ($1,$2,$3,NULLIF($4,''),$5,$6,
		  CASE WHEN $5::float8 IS NULL THEN NULL ELSE ST_SetSRID(ST_MakePoint($6,$5),4326) END,
		  'manual:' || $1)
		ON CONFLICT (origen_ref) DO UPDATE SET
		  subtipo = EXCLUDED.subtipo, estado = EXCLUDED.estado, sede = EXCLUDED.sede,
		  lat = EXCLUDED.lat, lon = EXCLUDED.lon, geom = EXCLUDED.geom, updated_at = now()
		RETURNING id`,
		b.Codigo, b.Subtipo, b.Estado, b.Sede, b.Lat, b.Lon).Scan(&id).Error
	if err != nil {
		return Bebedero{}, err
	}
	b.ID = id
	b.Activo = true
	return b, nil
}

func (s *Store) ListarPuntos(ctx context.Context, q string) ([]Punto, error) {
	var out []Punto
	err := s.db.WithContext(ctx).Raw(`
		SELECT id, titulo, lat, lon, COALESCE(url,''), activo
		FROM puntos_pucp
		WHERE activo AND ($1 = '' OR lower(titulo) LIKE '%' || lower($1) || '%')
		ORDER BY titulo`, q).Scan(&out).Error
	if out == nil {
		out = []Punto{}
	}
	return out, err
}

func (s *Store) GuardarPunto(ctx context.Context, p Punto) (Punto, error) {
	p.Titulo = strings.TrimSpace(p.Titulo)
	p.URL = urlLimpia(p.URL)
	if p.Titulo == "" || p.Lat < -12.20 || p.Lat > -11.90 || p.Lon < -77.30 || p.Lon > -76.90 {
		return Punto{}, fmt.Errorf("punto")
	}
	var id int64
	err := s.db.WithContext(ctx).Raw(`
		INSERT INTO puntos_pucp (titulo, lat, lon, url, geom, origen_ref)
		VALUES ($1,$2,$3,NULLIF($4,''), ST_SetSRID(ST_MakePoint($3,$2),4326), 'manual:' || md5($1 || $2::text || $3::text))
		ON CONFLICT (origen_ref) DO UPDATE SET titulo = EXCLUDED.titulo, url = EXCLUDED.url, geom = EXCLUDED.geom, updated_at = now()
		RETURNING id`, p.Titulo, p.Lat, p.Lon, p.URL).Scan(&id).Error
	if err != nil {
		return Punto{}, err
	}
	p.ID = id
	p.Activo = true
	return p, nil
}

func urlLimpia(raw string) string {
	raw = strings.TrimSpace(raw)
	if strings.Contains(raw, "placeId") || strings.Contains(raw, "place_id") || strings.Contains(strings.ToLower(raw), "phone") {
		return ""
	}
	return raw
}

// consultaReservas arma el listado. Un query vacío no se castea a date.
func consultaReservas(desde, hasta string) (string, []any) {
	q := `
		SELECT id, jardin_id, to_char(fecha,'YYYY-MM-DD'), to_char(hora_inicio,'HH24:MI'), to_char(hora_fin,'HH24:MI'),
		       estado, evento, COALESCE(unidad,''), origen, activo
		FROM reservas_jardin
		WHERE activo AND origen = 'ficticio'`
	args := []any{}
	if desde != "" {
		args = append(args, desde)
		q += fmt.Sprintf(" AND fecha >= $%d::date", len(args))
	}
	if hasta != "" {
		args = append(args, hasta)
		q += fmt.Sprintf(" AND fecha <= $%d::date", len(args))
	}
	q += " ORDER BY fecha, hora_inicio"
	return q, args
}

func (s *Store) ListarReservas(ctx context.Context, desde, hasta string) ([]Reserva, error) {
	q, args := consultaReservas(desde, hasta)
	var out []Reserva
	err := s.db.WithContext(ctx).Raw(q, args...).Scan(&out).Error
	if out == nil {
		out = []Reserva{}
	}
	return out, err
}

func (s *Store) GuardarReserva(ctx context.Context, r Reserva) (Reserva, error) {
	if r.Origen != "" && r.Origen != "ficticio" {
		return Reserva{}, fmt.Errorf("origen")
	}
	switch r.Estado {
	case "reservado", "realizado", "cancelado":
	default:
		return Reserva{}, fmt.Errorf("estado")
	}
	if r.HoraFin <= r.HoraInicio || r.Fecha == "" || strings.TrimSpace(r.Evento) == "" {
		return Reserva{}, fmt.Errorf("reserva")
	}
	var id int64
	err := s.db.WithContext(ctx).Raw(`
		INSERT INTO reservas_jardin (jardin_id, fecha, hora_inicio, hora_fin, estado, evento, unidad, origen, origen_ref)
		VALUES ($1,$2,$3,$4,$5,$6,NULLIF($7,''),'ficticio', 'manual:' || $2 || $3 || $5 || left($6, 40))
		ON CONFLICT (origen_ref) DO UPDATE SET
		  estado = EXCLUDED.estado, evento = EXCLUDED.evento, unidad = EXCLUDED.unidad,
		  hora_inicio = EXCLUDED.hora_inicio, hora_fin = EXCLUDED.hora_fin, updated_at = now()
		RETURNING id`,
		r.JardinID, r.Fecha, r.HoraInicio, r.HoraFin, r.Estado, r.Evento, r.Unidad).Scan(&id).Error
	if err != nil {
		return Reserva{}, err
	}
	r.ID = id
	r.Origen = "ficticio"
	r.Activo = true
	return r, nil
}

func (s *Store) ListarFichas(ctx context.Context, capa string) ([]Ficha, error) {
	q, ok := fichaSelect(capa)
	if !ok {
		return nil, fmt.Errorf("capa")
	}
	rows, err := s.db.WithContext(ctx).Raw(q).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []Ficha{}
	for rows.Next() {
		var f Ficha
		if err := rows.Scan(&f.ID, &f.FeatureID, &f.Nombre, &f.Codigo, &f.Nota, &f.Clase, &f.Riego, &f.AreaM2, &f.PerimetroM, &f.Pertenecen, &f.Uso, &f.GeoJSON, &f.Activo); err != nil {
			return nil, err
		}
		out = append(out, f)
	}
	return out, rows.Err()
}

func fichaSelect(capa string) (string, bool) {
	switch capa {
	case "fauna":
		return `SELECT id, feature_id, COALESCE(nombre,''), '', '', '', '', NULL::float8, NULL::float8, '', '', COALESCE(ST_AsGeoJSON(geom),''), activo FROM fauna WHERE activo ORDER BY feature_id`, true
	case "puertas":
		return `SELECT id, feature_id, COALESCE(nombre,''), COALESCE(codigo,''), '', '', '', NULL::float8, NULL::float8, '', '', COALESCE(ST_AsGeoJSON(geom),''), activo FROM puertas WHERE activo ORDER BY feature_id`, true
	case "playas_estacionamiento":
		return `SELECT id, feature_id, '', COALESCE(codigo,''), '', '', '', NULL::float8, NULL::float8, '', '', COALESCE(ST_AsGeoJSON(geom),''), activo FROM playas_estacionamiento WHERE activo ORDER BY feature_id`, true
	case "veredas_riesgo":
		return `SELECT id, feature_id, '', '', COALESCE(nota,''), '', '', NULL::float8, NULL::float8, '', '', COALESCE(ST_AsGeoJSON(geom),''), activo FROM veredas_riesgo WHERE activo ORDER BY feature_id`, true
	case "xerofiticas":
		return `SELECT id, feature_id, '', '', '', COALESCE(clase,''), COALESCE(riego,''), area_m2, perimetro_m, '', '', COALESCE(ST_AsGeoJSON(geom),''), activo FROM xerofiticas WHERE activo ORDER BY feature_id`, true
	case "jardines_reserva":
		return `SELECT id, feature_id, COALESCE(nombre,''), COALESCE(codigo,''), COALESCE(referencia,''), '', COALESCE(riego_act,''), area_m2, perimetro_m, COALESCE(pertenecen,''), COALESCE(uso,''), COALESCE(ST_AsGeoJSON(geom),''), activo FROM jardines_reserva WHERE activo ORDER BY feature_id`, true
	default:
		return "", false
	}
}

func geomFnCapa(capa string) string {
	if capa == "xerofiticas" || capa == "jardines_reserva" {
		return "catastro_geom_4326"
	}
	return "inventario_geom_4326"
}

func (s *Store) devolverID(ctx context.Context, q string, args ...any) (int64, error) {
	var id int64
	err := s.db.WithContext(ctx).Raw(q, args...).Scan(&id).Error
	if err != nil {
		return 0, err
	}
	if id == 0 {
		return 0, fmt.Errorf("no encontrado")
	}
	return id, nil
}

func (s *Store) GuardarFicha(ctx context.Context, capa string, f Ficha) (Ficha, error) {
	if !tablaPermitida(capa) {
		return Ficha{}, fmt.Errorf("capa")
	}
	f.FeatureID = strings.TrimSpace(f.FeatureID)
	if f.FeatureID == "" {
		return Ficha{}, fmt.Errorf("feature")
	}
	geomFn := geomFnCapa(capa)
	var q string
	var args []any
	switch capa {
	case "fauna":
		q = fmt.Sprintf(`INSERT INTO fauna (feature_id, nombre, geom, origen_ref) VALUES ($1, NULLIF($2,''), %s(NULLIF($3,'')), $1)
			ON CONFLICT (feature_id) DO UPDATE SET nombre = EXCLUDED.nombre, geom = COALESCE(EXCLUDED.geom, fauna.geom), updated_at = now()
			RETURNING id`, geomFn)
		args = []any{f.FeatureID, f.Nombre, f.GeoJSON}
	case "puertas":
		q = fmt.Sprintf(`INSERT INTO puertas (feature_id, codigo, nombre, geom, origen_ref) VALUES ($1, NULLIF($2,''), NULLIF($3,''), %s(NULLIF($4,'')), $1)
			ON CONFLICT (feature_id) DO UPDATE SET codigo = EXCLUDED.codigo, nombre = EXCLUDED.nombre, geom = COALESCE(EXCLUDED.geom, puertas.geom), updated_at = now()
			RETURNING id`, geomFn)
		args = []any{f.FeatureID, f.Codigo, f.Nombre, f.GeoJSON}
	case "playas_estacionamiento":
		q = fmt.Sprintf(`INSERT INTO playas_estacionamiento (feature_id, codigo, geom, origen_ref) VALUES ($1, NULLIF($2,''), %s(NULLIF($3,'')), $1)
			ON CONFLICT (feature_id) DO UPDATE SET codigo = EXCLUDED.codigo, geom = COALESCE(EXCLUDED.geom, playas_estacionamiento.geom), updated_at = now()
			RETURNING id`, geomFn)
		args = []any{f.FeatureID, f.Codigo, f.GeoJSON}
	case "veredas_riesgo":
		q = fmt.Sprintf(`INSERT INTO veredas_riesgo (feature_id, nota, geom, origen_ref) VALUES ($1, NULLIF($2,''), %s(NULLIF($3,'')), $1)
			ON CONFLICT (feature_id) DO UPDATE SET nota = EXCLUDED.nota, geom = COALESCE(EXCLUDED.geom, veredas_riesgo.geom), updated_at = now()
			RETURNING id`, geomFn)
		args = []any{f.FeatureID, f.Nota, f.GeoJSON}
	case "xerofiticas":
		q = fmt.Sprintf(`INSERT INTO xerofiticas (feature_id, clase, riego, area_m2, perimetro_m, geom, origen_ref) VALUES ($1, NULLIF($2,''), NULLIF($3,''), $4, $5, %s(NULLIF($6,'')), $1)
			ON CONFLICT (feature_id) DO UPDATE SET clase = EXCLUDED.clase, riego = EXCLUDED.riego, area_m2 = EXCLUDED.area_m2, perimetro_m = EXCLUDED.perimetro_m, geom = COALESCE(EXCLUDED.geom, xerofiticas.geom), updated_at = now()
			RETURNING id`, geomFn)
		args = []any{f.FeatureID, f.Clase, f.Riego, f.AreaM2, f.PerimetroM, f.GeoJSON}
	case "jardines_reserva":
		q = fmt.Sprintf(`INSERT INTO jardines_reserva (feature_id, codigo, nombre, uso, riego_act, referencia, pertenecen, area_m2, perimetro_m, geom, origen_ref)
			VALUES ($1, NULLIF($2,''), NULLIF($3,''), NULLIF($4,''), NULLIF($5,''), NULLIF($6,''), NULLIF($7,''), $8, $9, %s(NULLIF($10,'')), $1)
			ON CONFLICT (feature_id) DO UPDATE SET codigo = EXCLUDED.codigo, nombre = EXCLUDED.nombre, uso = EXCLUDED.uso, riego_act = EXCLUDED.riego_act, referencia = EXCLUDED.referencia, pertenecen = EXCLUDED.pertenecen, area_m2 = EXCLUDED.area_m2, perimetro_m = EXCLUDED.perimetro_m, geom = COALESCE(EXCLUDED.geom, jardines_reserva.geom), updated_at = now()
			RETURNING id`, geomFn)
		args = []any{f.FeatureID, f.Codigo, f.Nombre, f.Uso, f.Riego, f.Nota, f.Pertenecen, f.AreaM2, f.PerimetroM, f.GeoJSON}
	default:
		return Ficha{}, fmt.Errorf("capa")
	}
	id, err := s.devolverID(ctx, q, args...)
	if err != nil {
		return Ficha{}, err
	}
	f.ID = id
	f.Activo = true
	return f, nil
}

func (s *Store) ActualizarTacho(ctx context.Context, id int64, t Tacho) (Tacho, error) {
	if id < 1 || !ConteosNoNegativos(t) {
		return Tacho{}, fmt.Errorf("tacho inválido")
	}
	nid, err := s.devolverID(ctx, `
		UPDATE tachos SET
		  codigo = $1, lat = $2, lon = $3,
		  nota = NULLIF($4,''), lugar = NULLIF($5,''), espacios = NULLIF($6,''),
		  accion = NULLIF($7,''), tacho_actual = NULLIF($8,''), tacho_nuevo = NULLIF($9,''),
		  recomendaciones = NULLIF($10,''),
		  no_aprovechables = $11, papel_carton = $12, plastico = $13, vidrio = $14, pilas = $15,
		  peligrosos = $16, raee = $17, metales = $18, aniquem = $19,
		  intermedios_plastico = $20, intermedios_metal = $21,
		  geom = CASE WHEN $2::float8 IS NULL THEN NULL ELSE ST_SetSRID(ST_MakePoint($3,$2),4326) END,
		  updated_at = now()
		WHERE id = $22 AND activo
		RETURNING id`,
		strings.TrimSpace(t.Codigo), t.Lat, t.Lon, t.Nota, t.Lugar, t.Espacios, t.Accion, t.TachoActual, t.TachoNuevo, t.Recomendaciones,
		t.NoAprovechables, t.PapelCarton, t.Plastico, t.Vidrio, t.Pilas, t.Peligrosos, t.RAEE, t.Metales, t.Aniquem, t.IntermediosPlastico, t.IntermediosMetal,
		id)
	if err != nil {
		return Tacho{}, err
	}
	t.ID = nid
	t.Activo = true
	return t, nil
}

func (s *Store) ActualizarBebedero(ctx context.Context, id int64, b Bebedero) (Bebedero, error) {
	switch b.Subtipo {
	case "fuente", "llenador", "nuevo", "deterioro", "baja":
	default:
		return Bebedero{}, fmt.Errorf("subtipo")
	}
	if id < 1 || b.Estado == "" || !strings.HasPrefix(b.Codigo, "PT_") {
		return Bebedero{}, fmt.Errorf("bebedero")
	}
	nid, err := s.devolverID(ctx, `
		UPDATE bebederos SET
		  codigo = $1, subtipo = $2, estado = $3, sede = NULLIF($4,''), lat = $5, lon = $6,
		  geom = CASE WHEN $5::float8 IS NULL THEN NULL ELSE ST_SetSRID(ST_MakePoint($6,$5),4326) END,
		  updated_at = now()
		WHERE id = $7 AND activo
		RETURNING id`,
		b.Codigo, b.Subtipo, b.Estado, b.Sede, b.Lat, b.Lon, id)
	if err != nil {
		return Bebedero{}, err
	}
	b.ID = nid
	b.Activo = true
	return b, nil
}

func (s *Store) ActualizarPunto(ctx context.Context, id int64, p Punto) (Punto, error) {
	p.Titulo = strings.TrimSpace(p.Titulo)
	p.URL = urlLimpia(p.URL)
	if id < 1 || p.Titulo == "" || p.Lat < -12.20 || p.Lat > -11.90 || p.Lon < -77.30 || p.Lon > -76.90 {
		return Punto{}, fmt.Errorf("punto")
	}
	nid, err := s.devolverID(ctx, `
		UPDATE puntos_pucp SET
		  titulo = $1, lat = $2, lon = $3, url = NULLIF($4,''),
		  geom = ST_SetSRID(ST_MakePoint($3,$2),4326), updated_at = now()
		WHERE id = $5 AND activo
		RETURNING id`,
		p.Titulo, p.Lat, p.Lon, p.URL, id)
	if err != nil {
		return Punto{}, err
	}
	p.ID = nid
	p.Activo = true
	return p, nil
}

func (s *Store) ActualizarReserva(ctx context.Context, id int64, r Reserva) (Reserva, error) {
	if id < 1 || (r.Origen != "" && r.Origen != "ficticio") {
		return Reserva{}, fmt.Errorf("origen")
	}
	switch r.Estado {
	case "reservado", "realizado", "cancelado":
	default:
		return Reserva{}, fmt.Errorf("estado")
	}
	if r.HoraFin <= r.HoraInicio || r.Fecha == "" || strings.TrimSpace(r.Evento) == "" {
		return Reserva{}, fmt.Errorf("reserva")
	}
	nid, err := s.devolverID(ctx, `
		UPDATE reservas_jardin SET
		  jardin_id = $1, fecha = $2, hora_inicio = $3, hora_fin = $4,
		  estado = $5, evento = $6, unidad = NULLIF($7,''), origen = 'ficticio', updated_at = now()
		WHERE id = $8 AND activo
		RETURNING id`,
		r.JardinID, r.Fecha, r.HoraInicio, r.HoraFin, r.Estado, r.Evento, r.Unidad, id)
	if err != nil {
		return Reserva{}, err
	}
	r.ID = nid
	r.Origen = "ficticio"
	r.Activo = true
	return r, nil
}

func (s *Store) ActualizarFicha(ctx context.Context, capa string, id int64, f Ficha) (Ficha, error) {
	if id < 1 || !tablaPermitida(capa) {
		return Ficha{}, fmt.Errorf("capa")
	}
	f.FeatureID = strings.TrimSpace(f.FeatureID)
	if f.FeatureID == "" {
		return Ficha{}, fmt.Errorf("feature")
	}
	geomFn := geomFnCapa(capa)
	var q string
	var args []any
	switch capa {
	case "fauna":
		q = fmt.Sprintf(`UPDATE fauna SET feature_id = $1, nombre = NULLIF($2,''), geom = COALESCE(%s(NULLIF($3,'')), geom), updated_at = now() WHERE id = $4 AND activo RETURNING id`, geomFn)
		args = []any{f.FeatureID, f.Nombre, f.GeoJSON, id}
	case "puertas":
		q = fmt.Sprintf(`UPDATE puertas SET feature_id = $1, codigo = NULLIF($2,''), nombre = NULLIF($3,''), geom = COALESCE(%s(NULLIF($4,'')), geom), updated_at = now() WHERE id = $5 AND activo RETURNING id`, geomFn)
		args = []any{f.FeatureID, f.Codigo, f.Nombre, f.GeoJSON, id}
	case "playas_estacionamiento":
		q = fmt.Sprintf(`UPDATE playas_estacionamiento SET feature_id = $1, codigo = NULLIF($2,''), geom = COALESCE(%s(NULLIF($3,'')), geom), updated_at = now() WHERE id = $4 AND activo RETURNING id`, geomFn)
		args = []any{f.FeatureID, f.Codigo, f.GeoJSON, id}
	case "veredas_riesgo":
		q = fmt.Sprintf(`UPDATE veredas_riesgo SET feature_id = $1, nota = NULLIF($2,''), geom = COALESCE(%s(NULLIF($3,'')), geom), updated_at = now() WHERE id = $4 AND activo RETURNING id`, geomFn)
		args = []any{f.FeatureID, f.Nota, f.GeoJSON, id}
	case "xerofiticas":
		q = fmt.Sprintf(`UPDATE xerofiticas SET feature_id = $1, clase = NULLIF($2,''), riego = NULLIF($3,''), area_m2 = $4, perimetro_m = $5, geom = COALESCE(%s(NULLIF($6,'')), geom), updated_at = now() WHERE id = $7 AND activo RETURNING id`, geomFn)
		args = []any{f.FeatureID, f.Clase, f.Riego, f.AreaM2, f.PerimetroM, f.GeoJSON, id}
	case "jardines_reserva":
		q = fmt.Sprintf(`UPDATE jardines_reserva SET feature_id = $1, codigo = NULLIF($2,''), nombre = NULLIF($3,''), uso = NULLIF($4,''), riego_act = NULLIF($5,''), referencia = NULLIF($6,''), pertenecen = NULLIF($7,''), area_m2 = $8, perimetro_m = $9, geom = COALESCE(%s(NULLIF($10,'')), geom), updated_at = now() WHERE id = $11 AND activo RETURNING id`, geomFn)
		args = []any{f.FeatureID, f.Codigo, f.Nombre, f.Uso, f.Riego, f.Nota, f.Pertenecen, f.AreaM2, f.PerimetroM, f.GeoJSON, id}
	default:
		return Ficha{}, fmt.Errorf("capa")
	}
	nid, err := s.devolverID(ctx, q, args...)
	if err != nil {
		return Ficha{}, err
	}
	f.ID = nid
	f.Activo = true
	return f, nil
}

func CSVTachos(rows []Tacho) string {
	var b strings.Builder
	w := csv.NewWriter(&b)
	_ = w.Write([]string{"codigo", "lugar", "recomendaciones", "no_aprovechables", "papel_carton", "plastico", "vidrio", "pilas", "peligrosos", "raee", "metales", "aniquem", "intermedios_plastico", "intermedios_metal"})
	for _, t := range rows {
		_ = w.Write([]string{t.Codigo, t.Lugar, t.Recomendaciones, itoa(t.NoAprovechables), itoa(t.PapelCarton), itoa(t.Plastico), itoa(t.Vidrio), itoa(t.Pilas), itoa(t.Peligrosos), itoa(t.RAEE), itoa(t.Metales), itoa(t.Aniquem), itoa(t.IntermediosPlastico), itoa(t.IntermediosMetal)})
	}
	w.Flush()
	return b.String()
}

func itoa(n int) string { return strconv.Itoa(n) }

// CuerpoConContacto es true si el JSON trae teléfono o placeId.
func CuerpoConContacto(raw string) bool {
	low := strings.ToLower(raw)
	return strings.Contains(low, "phone") || strings.Contains(raw, "placeId") || strings.Contains(low, "place_id") || strings.Contains(low, "website")
}
