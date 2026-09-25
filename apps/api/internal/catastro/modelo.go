package catastro

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"unicode"

	"golang.org/x/text/unicode/norm"
	"gorm.io/gorm"
)

// ErrEntrada es una validación de catastro que no llega a la base.
var ErrEntrada = errors.New("entrada")

// ErrNoEncontrado es una fila que no existe o está dada de baja.
var ErrNoEncontrado = errors.New("no encontrado")

// TiposVegetacion son los valores que acepta ejemplares.tipo_vegetacion.
var TiposVegetacion = []string{
	"Árbol", "Palmera", "Arbusto", "Herbácea", "Trepadora", "Suculenta", "cafeto",
}

// CodigosZonaSupervision son Z1–Z4.
var CodigosZonaSupervision = []string{"Z1", "Z2", "Z3", "Z4"}

// NormalizarNombre deja un nombre de lugar en minúsculas y sin tildes.
func NormalizarNombre(s string) string {
	s = strings.TrimSpace(strings.ToLower(s))
	s = norm.NFD.String(s)
	var b strings.Builder
	for _, r := range s {
		if unicode.Is(unicode.Mn, r) {
			continue
		}
		b.WriteRune(r)
	}
	return strings.Join(strings.Fields(b.String()), " ")
}

// ValidarCodigoZona acepta solo Z1–Z4.
func ValidarCodigoZona(codigo string) error {
	codigo = strings.TrimSpace(codigo)
	for _, c := range CodigosZonaSupervision {
		if codigo == c {
			return nil
		}
	}
	return ErrEntrada
}

// ValidarPuntoCampus exige latitud y longitud dentro del campus.
func ValidarPuntoCampus(lat, lon float64) error {
	if lat < -12.20 || lat > -11.90 || lon < -77.30 || lon > -76.90 {
		return ErrEntrada
	}
	return nil
}

// ValidarTipoVegetacion acepta vacío o un tipo del catálogo.
func ValidarTipoVegetacion(tipo string) error {
	tipo = strings.TrimSpace(tipo)
	if tipo == "" {
		return nil
	}
	for _, t := range TiposVegetacion {
		if tipo == t {
			return nil
		}
	}
	return ErrEntrada
}

// ValidarCantidad exige un entero mayor o igual que 1.
func ValidarCantidad(n int) error {
	if n < 1 {
		return ErrEntrada
	}
	return nil
}

// ValidarReferencia corta textos de más de 500 caracteres.
func ValidarReferencia(s string) error {
	if len([]rune(strings.TrimSpace(s))) > 500 {
		return ErrEntrada
	}
	return nil
}

// ZonaSupervision es una de las cuatro zonas Z1–Z4.
type ZonaSupervision struct {
	ID      int64    `json:"id"`
	Codigo  string   `json:"codigo"`
	Nombre  string   `json:"nombre"`
	AreaM2  *float64 `json:"area_m2,omitempty"`
	ConGeom bool     `json:"con_geometria"`
	Activo  bool     `json:"activo"`
}

// Cuadrilla es un equipo ficticio.
type Cuadrilla struct {
	ID             string `json:"id"`
	NombreFicticio string `json:"nombre_ficticio"`
	Turno          string `json:"turno"`
	Activo         bool   `json:"activo"`
}

// Lugar es un punto con nombre normalizado.
type Lugar struct {
	ID                int64   `json:"id"`
	Nombre            string  `json:"nombre"`
	NombreNorm        string  `json:"nombre_norm"`
	Lat               float64 `json:"lat"`
	Lon               float64 `json:"lon"`
	ZonaSupervisionID *int64  `json:"zona_supervision_id,omitempty"`
	Activo            bool    `json:"activo"`
}

// Especie clasifica ejemplares por nombre científico.
type Especie struct {
	ID               int64  `json:"id"`
	NombreCientifico string `json:"nombre_cientifico"`
	NombreComun      string `json:"nombre_comun"`
	Activo           bool   `json:"activo"`
}

// Ejemplar es un individuo de flora.
type Ejemplar struct {
	ID                 int64    `json:"id"`
	NumeroOrigen       *int     `json:"numero_origen,omitempty"`
	Codigo             string   `json:"codigo"`
	EspecieID          *int64   `json:"especie_id,omitempty"`
	NombreComun        string   `json:"nombre_comun"`
	TipoVegetacion     string   `json:"tipo_vegetacion"`
	Cantidad           int      `json:"cantidad"`
	UbicacionLugarID   *int64   `json:"ubicacion_lugar_id,omitempty"`
	Referencia         string   `json:"referencia"`
	Lat                *float64 `json:"lat,omitempty"`
	Lon                *float64 `json:"lon,omitempty"`
	ObservacionFen2026 string   `json:"observacion_fen_2026"`
	Salud              *string  `json:"salud"`
	Activo             bool     `json:"activo"`
}

// CodigoHistorico conserva el código anterior de un ejemplar.
type CodigoHistorico struct {
	ID             int64  `json:"id"`
	EjemplarID     int64  `json:"ejemplar_id"`
	CodigoAnterior string `json:"codigo_anterior"`
	CodigoNuevo    string `json:"codigo_nuevo"`
}

// PoligonoCuadrilla es la entidad que antes se leía como zona.
type PoligonoCuadrilla struct {
	ID                int64  `json:"id"`
	FeatureID         string `json:"feature_id"`
	Codigo            string `json:"codigo"`
	Nombre            string `json:"nombre"`
	CuadrillaID       string `json:"cuadrilla_id"`
	ZonaSupervisionID *int64 `json:"zona_supervision_id,omitempty"`
	ConGeom           bool   `json:"con_geometria"`
	Activo            bool   `json:"activo"`
}

// CapaFicha es una fila editable de fauna, puerta, playa, vereda, xerofítica o jardín.
type CapaFicha struct {
	ID         int64  `json:"id"`
	FeatureID  string `json:"feature_id"`
	Nombre     string `json:"nombre,omitempty"`
	Codigo     string `json:"codigo,omitempty"`
	Nota       string `json:"nota,omitempty"`
	Clase      string `json:"clase,omitempty"`
	Riego      string `json:"riego,omitempty"`
	Pertenecen string `json:"pertenecen,omitempty"`
	Activo     bool   `json:"activo"`
}

// ListarZonasSupervision devuelve las zonas activas.
func (s *Store) ListarZonasSupervision(ctx context.Context) ([]ZonaSupervision, error) {
	rows, err := s.db.WithContext(ctx).Raw(`
		SELECT id, codigo, nombre, area_m2, geom IS NOT NULL, activo
		FROM zonas_supervision
		WHERE activo
		ORDER BY codigo`).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []ZonaSupervision{}
	for rows.Next() {
		var z ZonaSupervision
		var area *float64
		if err := rows.Scan(&z.ID, &z.Codigo, &z.Nombre, &area, &z.ConGeom, &z.Activo); err != nil {
			return nil, err
		}
		z.AreaM2 = area
		out = append(out, z)
	}
	return out, rows.Err()
}

// CrearZonaSupervision inserta una zona con geometría.
func (s *Store) CrearZonaSupervision(ctx context.Context, codigo, nombre, geojson string, area *float64) (ZonaSupervision, error) {
	codigo = strings.TrimSpace(codigo)
	nombre = strings.TrimSpace(nombre)
	if err := ValidarCodigoZona(codigo); err != nil || nombre == "" || strings.TrimSpace(geojson) == "" {
		return ZonaSupervision{}, ErrEntrada
	}
	var id int64
	err := s.db.WithContext(ctx).Raw(`
		INSERT INTO zonas_supervision (codigo, nombre, area_m2, geom)
		VALUES ($1, $2, $3, catastro_geom_4326($4))
		RETURNING id`, codigo, nombre, area, geojson).Scan(&id).Error
	if err != nil {
		return ZonaSupervision{}, ErrEntrada
	}
	list, err := s.ListarZonasSupervision(ctx)
	if err != nil {
		return ZonaSupervision{}, err
	}
	for _, z := range list {
		if z.ID == id {
			return z, nil
		}
	}
	return ZonaSupervision{}, ErrNoEncontrado
}

// ListarCuadrillas devuelve los equipos activos.
func (s *Store) ListarCuadrillas(ctx context.Context) ([]Cuadrilla, error) {
	var out []Cuadrilla
	err := s.db.WithContext(ctx).Raw(`
		SELECT id, nombre_ficticio, turno, activo
		FROM cuadrillas
		WHERE activo
		ORDER BY id`).Scan(&out).Error
	if err != nil {
		return nil, err
	}
	if out == nil {
		out = []Cuadrilla{}
	}
	return out, nil
}

// CrearCuadrilla registra un equipo ficticio.
func (s *Store) CrearCuadrilla(ctx context.Context, id, nombre, turno string) (Cuadrilla, error) {
	id = strings.TrimSpace(id)
	nombre = strings.TrimSpace(nombre)
	turno = strings.TrimSpace(turno)
	if id == "" || nombre == "" || (turno != "manana" && turno != "tarde") {
		return Cuadrilla{}, ErrEntrada
	}
	err := s.db.WithContext(ctx).Exec(`
		INSERT INTO cuadrillas (id, nombre_ficticio, turno) VALUES ($1, $2, $3)`,
		id, nombre, turno).Error
	if err != nil {
		return Cuadrilla{}, ErrEntrada
	}
	return Cuadrilla{ID: id, NombreFicticio: nombre, Turno: turno, Activo: true}, nil
}

// ListarLugares devuelve lugares activos.
func (s *Store) ListarLugares(ctx context.Context) ([]Lugar, error) {
	rows, err := s.db.WithContext(ctx).Raw(`
		SELECT id, nombre, nombre_norm, lat, lon, zona_supervision_id, activo
		FROM lugares
		WHERE activo
		ORDER BY nombre_norm`).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []Lugar{}
	for rows.Next() {
		var l Lugar
		var zona *int64
		if err := rows.Scan(&l.ID, &l.Nombre, &l.NombreNorm, &l.Lat, &l.Lon, &zona, &l.Activo); err != nil {
			return nil, err
		}
		l.ZonaSupervisionID = zona
		out = append(out, l)
	}
	return out, rows.Err()
}

// CrearLugar inserta un lugar con nombre normalizado único.
func (s *Store) CrearLugar(ctx context.Context, nombre string, lat, lon float64, zonaID *int64) (Lugar, error) {
	nombre = strings.TrimSpace(nombre)
	norm := NormalizarNombre(nombre)
	if nombre == "" || norm == "" {
		return Lugar{}, ErrEntrada
	}
	if err := ValidarPuntoCampus(lat, lon); err != nil {
		return Lugar{}, err
	}
	var id int64
	err := s.db.WithContext(ctx).Raw(`
		INSERT INTO lugares (nombre, nombre_norm, lat, lon, zona_supervision_id)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id`, nombre, norm, lat, lon, zonaID).Scan(&id).Error
	if err != nil {
		return Lugar{}, ErrEntrada
	}
	return Lugar{ID: id, Nombre: nombre, NombreNorm: norm, Lat: lat, Lon: lon, ZonaSupervisionID: zonaID, Activo: true}, nil
}

// ListarEspecies devuelve especies activas.
func (s *Store) ListarEspecies(ctx context.Context) ([]Especie, error) {
	rows, err := s.db.WithContext(ctx).Raw(`
		SELECT id, nombre_cientifico, COALESCE(nombre_comun, ''), activo
		FROM especies WHERE activo ORDER BY nombre_cientifico`).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []Especie{}
	for rows.Next() {
		var e Especie
		if err := rows.Scan(&e.ID, &e.NombreCientifico, &e.NombreComun, &e.Activo); err != nil {
			return nil, err
		}
		out = append(out, e)
	}
	return out, rows.Err()
}

// CrearEspecie da de alta un nombre científico.
func (s *Store) CrearEspecie(ctx context.Context, cientifico, comun string) (Especie, error) {
	cientifico = strings.TrimSpace(cientifico)
	comun = strings.TrimSpace(comun)
	if cientifico == "" {
		return Especie{}, ErrEntrada
	}
	var id int64
	err := s.db.WithContext(ctx).Raw(`
		INSERT INTO especies (nombre_cientifico, nombre_comun)
		VALUES ($1, NULLIF($2, ''))
		RETURNING id`, cientifico, comun).Scan(&id).Error
	if err != nil {
		return Especie{}, ErrEntrada
	}
	return Especie{ID: id, NombreCientifico: cientifico, NombreComun: comun, Activo: true}, nil
}

// ListarEjemplares devuelve ejemplares activos, hasta 100.
func (s *Store) ListarEjemplares(ctx context.Context) ([]Ejemplar, error) {
	rows, err := s.db.WithContext(ctx).Raw(`
		SELECT id, numero_origen, COALESCE(codigo, ''), especie_id, COALESCE(nombre_comun, ''),
		       COALESCE(tipo_vegetacion, ''), cantidad, ubicacion_lugar_id, COALESCE(referencia, ''),
		       lat, lon, COALESCE(observacion_fen_2026, ''), salud, activo
		FROM ejemplares
		WHERE activo
		ORDER BY numero_origen NULLS LAST, id
		LIMIT 100`).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []Ejemplar{}
	for rows.Next() {
		var e Ejemplar
		if err := rows.Scan(
			&e.ID, &e.NumeroOrigen, &e.Codigo, &e.EspecieID, &e.NombreComun,
			&e.TipoVegetacion, &e.Cantidad, &e.UbicacionLugarID, &e.Referencia,
			&e.Lat, &e.Lon, &e.ObservacionFen2026, &e.Salud, &e.Activo,
		); err != nil {
			return nil, err
		}
		out = append(out, e)
	}
	return out, rows.Err()
}

// CrearEjemplar registra un ejemplar. salud no se acepta: queda NULL.
func (s *Store) CrearEjemplar(ctx context.Context, e Ejemplar) (Ejemplar, error) {
	if e.Cantidad == 0 {
		e.Cantidad = 1
	}
	if err := ValidarCantidad(e.Cantidad); err != nil {
		return Ejemplar{}, err
	}
	if err := ValidarTipoVegetacion(e.TipoVegetacion); err != nil {
		return Ejemplar{}, err
	}
	if err := ValidarReferencia(e.Referencia); err != nil {
		return Ejemplar{}, err
	}
	if e.Lat != nil || e.Lon != nil {
		if e.Lat == nil || e.Lon == nil {
			return Ejemplar{}, ErrEntrada
		}
		if err := ValidarPuntoCampus(*e.Lat, *e.Lon); err != nil {
			return Ejemplar{}, err
		}
	}
	var id int64
	err := s.db.WithContext(ctx).Raw(`
		INSERT INTO ejemplares (
		  numero_origen, codigo, especie_id, nombre_comun, tipo_vegetacion, cantidad,
		  ubicacion_lugar_id, referencia, lat, lon, observacion_fen_2026, salud, geom
		) VALUES (
		  $1, NULLIF($2, ''), $3, NULLIF($4, ''), NULLIF($5, ''), $6,
		  $7, NULLIF($8, ''), $9, $10, NULLIF($11, ''), NULL,
		  CASE WHEN $9::float8 IS NULL THEN NULL
		       ELSE ST_SetSRID(ST_MakePoint($10, $9), 4326) END
		)
		RETURNING id`,
		e.NumeroOrigen, strings.TrimSpace(e.Codigo), e.EspecieID, strings.TrimSpace(e.NombreComun),
		strings.TrimSpace(e.TipoVegetacion), e.Cantidad, e.UbicacionLugarID, strings.TrimSpace(e.Referencia),
		e.Lat, e.Lon, strings.TrimSpace(e.ObservacionFen2026),
	).Scan(&id).Error
	if err != nil {
		return Ejemplar{}, ErrEntrada
	}
	e.ID = id
	e.Salud = nil
	e.Activo = true
	return e, nil
}

// Recodificar guarda el código anterior y escribe el nuevo.
func (s *Store) Recodificar(ctx context.Context, ejemplarID int64, codigoNuevo string) (CodigoHistorico, error) {
	codigoNuevo = strings.TrimSpace(codigoNuevo)
	if ejemplarID < 1 || codigoNuevo == "" {
		return CodigoHistorico{}, ErrEntrada
	}
	var out CodigoHistorico
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var anterior sql.NullString
		row := tx.Raw(`SELECT codigo FROM ejemplares WHERE id = $1 AND activo`, ejemplarID).Row()
		if err := row.Scan(&anterior); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return ErrNoEncontrado
			}
			return err
		}
		previo := ""
		if anterior.Valid {
			previo = anterior.String
		}
		if previo == codigoNuevo {
			out = CodigoHistorico{EjemplarID: ejemplarID, CodigoAnterior: previo, CodigoNuevo: codigoNuevo}
			return nil
		}
		if err := tx.Raw(`
			INSERT INTO codigos_historicos (ejemplar_id, codigo_anterior, codigo_nuevo)
			VALUES ($1, $2, $3)
			RETURNING id, ejemplar_id, codigo_anterior, codigo_nuevo`,
			ejemplarID, previo, codigoNuevo).Scan(&out).Error; err != nil {
			return err
		}
		res := tx.Exec(`UPDATE ejemplares SET codigo = $2, updated_at = now() WHERE id = $1 AND activo`, ejemplarID, codigoNuevo)
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected != 1 {
			return ErrNoEncontrado
		}
		return nil
	})
	if err != nil {
		return CodigoHistorico{}, err
	}
	return out, nil
}

// ListarCodigos devuelve el historial de un ejemplar.
func (s *Store) ListarCodigos(ctx context.Context, ejemplarID int64) ([]CodigoHistorico, error) {
	var out []CodigoHistorico
	err := s.db.WithContext(ctx).Raw(`
		SELECT id, ejemplar_id, codigo_anterior, COALESCE(codigo_nuevo, '')
		FROM codigos_historicos
		WHERE ejemplar_id = $1
		ORDER BY id`, ejemplarID).Scan(&out).Error
	if err != nil {
		return nil, err
	}
	if out == nil {
		out = []CodigoHistorico{}
	}
	return out, nil
}

// ListarPoligonos devuelve polígonos activos. La vista zonas sigue leyendo los mismos.
func (s *Store) ListarPoligonos(ctx context.Context) ([]PoligonoCuadrilla, error) {
	rows, err := s.db.WithContext(ctx).Raw(`
		SELECT id, feature_id, COALESCE(codigo, ''), COALESCE(nombre, ''),
		       COALESCE(cuadrilla_id, ''), zona_supervision_id, geom IS NOT NULL, activo
		FROM poligonos_cuadrilla
		WHERE activo
		ORDER BY source_index
		LIMIT 100`).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []PoligonoCuadrilla{}
	for rows.Next() {
		var p PoligonoCuadrilla
		if err := rows.Scan(&p.ID, &p.FeatureID, &p.Codigo, &p.Nombre, &p.CuadrillaID, &p.ZonaSupervisionID, &p.ConGeom, &p.Activo); err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

// ListarCapa lee una tabla de la sección 4.13.
func (s *Store) ListarCapa(ctx context.Context, tabla string) ([]CapaFicha, error) {
	q, ok := capaQuery(tabla)
	if !ok {
		return nil, ErrEntrada
	}
	rows, err := s.db.WithContext(ctx).Raw(q).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []CapaFicha{}
	for rows.Next() {
		var f CapaFicha
		if err := rows.Scan(&f.ID, &f.FeatureID, &f.Nombre, &f.Codigo, &f.Nota, &f.Clase, &f.Riego, &f.Pertenecen, &f.Activo); err != nil {
			return nil, err
		}
		out = append(out, f)
	}
	return out, rows.Err()
}

func capaQuery(tabla string) (string, bool) {
	switch tabla {
	case "fauna":
		return `SELECT id, feature_id, COALESCE(nombre, ''), '', '', '', '', '', activo FROM fauna WHERE activo ORDER BY feature_id`, true
	case "puertas":
		return `SELECT id, feature_id, COALESCE(nombre, ''), COALESCE(codigo, ''), '', '', '', '', activo FROM puertas WHERE activo ORDER BY feature_id`, true
	case "playas_estacionamiento":
		return `SELECT id, feature_id, '', COALESCE(codigo, ''), '', '', '', '', activo FROM playas_estacionamiento WHERE activo ORDER BY feature_id`, true
	case "veredas_riesgo":
		return `SELECT id, feature_id, '', '', COALESCE(nota, ''), '', '', '', activo FROM veredas_riesgo WHERE activo ORDER BY feature_id`, true
	case "xerofiticas":
		return `SELECT id, feature_id, '', '', '', COALESCE(clase, ''), COALESCE(riego, ''), '', activo FROM xerofiticas WHERE activo ORDER BY feature_id`, true
	case "jardines_reserva":
		return `SELECT id, feature_id, COALESCE(nombre, ''), COALESCE(codigo, ''), '', '', COALESCE(riego_act, ''), COALESCE(pertenecen, ''), activo FROM jardines_reserva WHERE activo ORDER BY feature_id`, true
	default:
		return "", false
	}
}
