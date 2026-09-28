// Package postgres provides PostgreSQL implementations of persistence contracts.
package postgres

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"gorm.io/gorm"

	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/contracts"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/entities"
	domainErrors "github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/errors"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/persistence/mapper"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/persistence/models"
)

// CapasEditables lists the six editable auxiliary cadastral layer tables.
var CapasEditables = []string{"fauna", "puertas", "playas_estacionamiento", "veredas_riesgo", "xerofiticas", "jardines_reserva"}

type inventarioCampoRepository struct {
	db *gorm.DB
}

// NewInventarioCampoRepository creates a new IInventarioCampoRepository instance.
func NewInventarioCampoRepository(db *gorm.DB) contracts.IInventarioCampoRepository {
	return &inventarioCampoRepository{db: db}
}

// ClavesPatch returns a map of present keys, a clean JSON string, or ErrPatchVacio.
// Notice: _todo is stripped out so clients cannot bypass partial patch safety.
func ClavesPatch(doc json.RawMessage) (map[string]bool, string, error) {
	limpio := bytes.TrimSpace(doc)
	if len(limpio) == 0 || string(limpio) == "null" || string(limpio) == "{}" {
		return nil, "", domainErrors.ErrPatchVacio
	}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(limpio, &raw); err != nil {
		return nil, "", fmt.Errorf("JSON inválido")
	}
	delete(raw, "_todo")
	if len(raw) == 0 {
		return nil, "", domainErrors.ErrPatchVacio
	}
	out := make(map[string]bool, len(raw))
	for k := range raw {
		out[k] = true
	}
	cuerpo, err := json.Marshal(raw)
	if err != nil {
		return nil, "", err
	}
	return out, string(cuerpo), nil
}

func tablaPermitida(tabla string) bool {
	switch tabla {
	case "tachos", "bebederos", "puntos_pucp", "reservas_jardin", "fauna", "puertas", "playas_estacionamiento", "veredas_riesgo", "xerofiticas", "jardines_reserva":
		return true
	default:
		return false
	}
}

func geomFnCapa(capa string) string {
	if capa == "xerofiticas" || capa == "jardines_reserva" {
		return "catastro_geom_4326"
	}
	return "inventario_geom_4326"
}

func (r *inventarioCampoRepository) devolverID(ctx context.Context, q string, args ...any) (int64, error) {
	sqlDB, err := r.db.DB()
	if err != nil {
		return 0, err
	}
	var id int64
	err = sqlDB.QueryRowContext(ctx, q, args...).Scan(&id)
	if err != nil {
		return 0, err
	}
	if id == 0 {
		return 0, domainErrors.ErrRegistroNoEncontrado
	}
	return id, nil
}

func (r *inventarioCampoRepository) ListarTachos(ctx context.Context) ([]entities.Tacho, error) {
	rows, err := r.db.WithContext(ctx).Raw(`
		SELECT id, codigo, lat, lon, COALESCE(nota,''), COALESCE(lugar,''), COALESCE(espacios,''),
		       COALESCE(accion,''), COALESCE(tacho_actual,''), COALESCE(tacho_nuevo,''), COALESCE(recomendaciones,''),
		       no_aprovechables, papel_carton, plastico, vidrio, pilas, peligrosos, raee, metales, aniquem,
		       intermedios_plastico, intermedios_metal, activo
		FROM tachos WHERE activo ORDER BY codigo`).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []entities.Tacho{}
	for rows.Next() {
		var (
			m                                                                       models.TachoModel
			nota, lugar, espacios, accion, tachoActual, tachoNuevo, recomendaciones string
		)
		if err := rows.Scan(
			&m.ID, &m.Codigo, &m.Lat, &m.Lon, &nota, &lugar, &espacios,
			&accion, &tachoActual, &tachoNuevo, &recomendaciones,
			&m.NoAprovechables, &m.PapelCarton, &m.Plastico, &m.Vidrio, &m.Pilas,
			&m.Peligrosos, &m.RAEE, &m.Metales, &m.Aniquem,
			&m.IntermediosPlastico, &m.IntermediosMetal, &m.Activo,
		); err != nil {
			return nil, err
		}
		m.Nota = &nota
		m.Lugar = &lugar
		m.Espacios = &espacios
		m.Accion = &accion
		m.TachoActual = &tachoActual
		m.TachoNuevo = &tachoNuevo
		m.Recomendaciones = &recomendaciones
		out = append(out, *mapper.TachoModelToEntity(&m))
	}
	return out, rows.Err()
}

func (r *inventarioCampoRepository) GuardarTacho(ctx context.Context, t entities.Tacho) (*entities.Tacho, error) {
	var id int64
	err := r.db.WithContext(ctx).Raw(`
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
		return nil, err
	}
	t.ID = id
	t.Activo = true
	return &t, nil
}

func conteosTocadosValidos(claves map[string]bool, t entities.Tacho) bool {
	pares := []struct {
		clave string
		n     int
	}{
		{"no_aprovechables", t.NoAprovechables},
		{"papel_carton", t.PapelCarton},
		{"plastico", t.Plastico},
		{"vidrio", t.Vidrio},
		{"pilas", t.Pilas},
		{"peligrosos", t.Peligrosos},
		{"raee", t.RAEE},
		{"metales", t.Metales},
		{"aniquem", t.Aniquem},
		{"intermedios_plastico", t.IntermediosPlastico},
		{"intermedios_metal", t.IntermediosMetal},
	}
	for _, par := range pares {
		if claves[par.clave] && par.n < 0 {
			return false
		}
	}
	return true
}

func (r *inventarioCampoRepository) ActualizarTacho(ctx context.Context, id int64, t entities.Tacho, rawJSON []byte) (*entities.Tacho, error) {
	claves, docJSON, err := ClavesPatch(rawJSON)
	if err != nil {
		return nil, err
	}
	if id < 1 {
		return nil, domainErrors.ErrTachoInvalido
	}
	if claves["codigo"] && !strings.HasPrefix(strings.TrimSpace(t.Codigo), "PT") {
		return nil, domainErrors.ErrTachoInvalido
	}
	if !conteosTocadosValidos(claves, t) {
		return nil, domainErrors.ErrTachoInvalido
	}
	nid, err := r.devolverID(ctx, `
		UPDATE tachos SET
		  codigo = CASE WHEN $23::jsonb ? 'codigo' THEN $1 ELSE codigo END,
		  lat = CASE WHEN $23::jsonb ? 'lat' THEN $2 ELSE lat END,
		  lon = CASE WHEN $23::jsonb ? 'lon' THEN $3 ELSE lon END,
		  nota = CASE WHEN $23::jsonb ? 'nota' THEN NULLIF($4,'') ELSE nota END,
		  lugar = CASE WHEN $23::jsonb ? 'lugar' THEN NULLIF($5,'') ELSE lugar END,
		  espacios = CASE WHEN $23::jsonb ? 'espacios' THEN NULLIF($6,'') ELSE espacios END,
		  accion = CASE WHEN $23::jsonb ? 'accion' THEN NULLIF($7,'') ELSE accion END,
		  tacho_actual = CASE WHEN $23::jsonb ? 'tacho_actual' THEN NULLIF($8,'') ELSE tacho_actual END,
		  tacho_nuevo = CASE WHEN $23::jsonb ? 'tacho_nuevo' THEN NULLIF($9,'') ELSE tacho_nuevo END,
		  recomendaciones = CASE WHEN $23::jsonb ? 'recomendaciones' THEN NULLIF($10,'') ELSE recomendaciones END,
		  no_aprovechables = CASE WHEN $23::jsonb ? 'no_aprovechables' THEN $11 ELSE no_aprovechables END,
		  papel_carton = CASE WHEN $23::jsonb ? 'papel_carton' THEN $12 ELSE papel_carton END,
		  plastico = CASE WHEN $23::jsonb ? 'plastico' THEN $13 ELSE plastico END,
		  vidrio = CASE WHEN $23::jsonb ? 'vidrio' THEN $14 ELSE vidrio END,
		  pilas = CASE WHEN $23::jsonb ? 'pilas' THEN $15 ELSE pilas END,
		  peligrosos = CASE WHEN $23::jsonb ? 'peligrosos' THEN $16 ELSE peligrosos END,
		  raee = CASE WHEN $23::jsonb ? 'raee' THEN $17 ELSE raee END,
		  metales = CASE WHEN $23::jsonb ? 'metales' THEN $18 ELSE metales END,
		  aniquem = CASE WHEN $23::jsonb ? 'aniquem' THEN $19 ELSE aniquem END,
		  intermedios_plastico = CASE WHEN $23::jsonb ? 'intermedios_plastico' THEN $20 ELSE intermedios_plastico END,
		  intermedios_metal = CASE WHEN $23::jsonb ? 'intermedios_metal' THEN $21 ELSE intermedios_metal END,
		  geom = CASE
		    WHEN NOT ($23::jsonb ? 'lat' OR $23::jsonb ? 'lon') THEN geom
		    ELSE ST_SetSRID(ST_MakePoint(
		      CASE WHEN $23::jsonb ? 'lon' THEN $3 ELSE lon END,
		      CASE WHEN $23::jsonb ? 'lat' THEN $2 ELSE lat END
		    ), 4326)
		  END,
		  updated_at = now()
		WHERE id = $22 AND activo
		RETURNING id`,
		strings.TrimSpace(t.Codigo), t.Lat, t.Lon, t.Nota, t.Lugar, t.Espacios, t.Accion, t.TachoActual, t.TachoNuevo, t.Recomendaciones,
		t.NoAprovechables, t.PapelCarton, t.Plastico, t.Vidrio, t.Pilas, t.Peligrosos, t.RAEE, t.Metales, t.Aniquem, t.IntermediosPlastico, t.IntermediosMetal,
		id, docJSON)
	if err != nil {
		return nil, err
	}
	t.ID = nid
	t.Activo = true
	return &t, nil
}

func (r *inventarioCampoRepository) ListarBebederos(ctx context.Context) ([]entities.Bebedero, error) {
	var modelsList []models.BebederoModel
	err := r.db.WithContext(ctx).Raw(`
		SELECT id, codigo, subtipo, estado, COALESCE(sede,'') AS sede, lat, lon, activo
		FROM bebederos WHERE activo ORDER BY codigo`).Scan(&modelsList).Error
	if err != nil {
		return nil, err
	}
	out := make([]entities.Bebedero, 0, len(modelsList))
	for i := range modelsList {
		out = append(out, *mapper.BebederoModelToEntity(&modelsList[i]))
	}
	return out, nil
}

func (r *inventarioCampoRepository) GuardarBebedero(ctx context.Context, b entities.Bebedero) (*entities.Bebedero, error) {
	var id int64
	err := r.db.WithContext(ctx).Raw(`
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
		return nil, err
	}
	b.ID = id
	b.Activo = true
	return &b, nil
}

func (r *inventarioCampoRepository) ActualizarBebedero(ctx context.Context, id int64, b entities.Bebedero, rawJSON []byte) (*entities.Bebedero, error) {
	claves, docJSON, err := ClavesPatch(rawJSON)
	if err != nil {
		return nil, err
	}
	if claves["subtipo"] {
		switch b.Subtipo {
		case "fuente", "llenador", "nuevo", "deterioro", "baja":
		default:
			return nil, domainErrors.ErrBebederoInvalido
		}
	}
	if id < 1 || (claves["estado"] && b.Estado == "") || (claves["codigo"] && !strings.HasPrefix(b.Codigo, "PT_")) {
		return nil, domainErrors.ErrBebederoInvalido
	}
	nid, err := r.devolverID(ctx, `
		UPDATE bebederos SET
		  codigo = CASE WHEN $8::jsonb ? 'codigo' THEN $1 ELSE codigo END,
		  subtipo = CASE WHEN $8::jsonb ? 'subtipo' THEN $2 ELSE subtipo END,
		  estado = CASE WHEN $8::jsonb ? 'estado' THEN $3 ELSE estado END,
		  sede = CASE WHEN $8::jsonb ? 'sede' THEN NULLIF($4,'') ELSE sede END,
		  lat = CASE WHEN $8::jsonb ? 'lat' THEN $5 ELSE lat END,
		  lon = CASE WHEN $8::jsonb ? 'lon' THEN $6 ELSE lon END,
		  geom = CASE
		    WHEN NOT ($8::jsonb ? 'lat' OR $8::jsonb ? 'lon') THEN geom
		    ELSE ST_SetSRID(ST_MakePoint(
		      CASE WHEN $8::jsonb ? 'lon' THEN $6 ELSE lon END,
		      CASE WHEN $8::jsonb ? 'lat' THEN $5 ELSE lat END
		    ), 4326)
		  END,
		  updated_at = now()
		WHERE id = $7 AND activo
		RETURNING id`,
		b.Codigo, b.Subtipo, b.Estado, b.Sede, b.Lat, b.Lon, id, docJSON)
	if err != nil {
		return nil, err
	}
	b.ID = nid
	b.Activo = true
	return &b, nil
}

func (r *inventarioCampoRepository) ListarPuntos(ctx context.Context, q string) ([]entities.PuntoPUCP, error) {
	var modelsList []models.PuntoPUCPModel
	err := r.db.WithContext(ctx).Raw(`
		SELECT id, titulo, lat, lon, COALESCE(url,'') AS url, activo
		FROM puntos_pucp
		WHERE activo AND ($1 = '' OR lower(titulo) LIKE '%' || lower($1) || '%')
		ORDER BY titulo`, q).Scan(&modelsList).Error
	if err != nil {
		return nil, err
	}
	out := make([]entities.PuntoPUCP, 0, len(modelsList))
	for i := range modelsList {
		out = append(out, *mapper.PuntoPUCPModelToEntity(&modelsList[i]))
	}
	return out, nil
}

func (r *inventarioCampoRepository) GuardarPunto(ctx context.Context, p entities.PuntoPUCP) (*entities.PuntoPUCP, error) {
	var id int64
	err := r.db.WithContext(ctx).Raw(`
		INSERT INTO puntos_pucp (titulo, lat, lon, url, geom, origen_ref)
		VALUES ($1,$2,$3,NULLIF($4,''), ST_SetSRID(ST_MakePoint($3,$2),4326), 'manual:' || md5($1 || $2::text || $3::text))
		ON CONFLICT (origen_ref) DO UPDATE SET titulo = EXCLUDED.titulo, url = EXCLUDED.url, geom = EXCLUDED.geom, updated_at = now()
		RETURNING id`, p.Titulo, p.Lat, p.Lon, p.URL).Scan(&id).Error
	if err != nil {
		return nil, err
	}
	p.ID = id
	p.Activo = true
	return &p, nil
}

func urlLimpia(raw string) string {
	raw = strings.TrimSpace(raw)
	if strings.Contains(raw, "placeId") || strings.Contains(raw, "place_id") || strings.Contains(strings.ToLower(raw), "phone") {
		return ""
	}
	return raw
}

func (r *inventarioCampoRepository) ActualizarPunto(ctx context.Context, id int64, p entities.PuntoPUCP, rawJSON []byte) (*entities.PuntoPUCP, error) {
	claves, docJSON, err := ClavesPatch(rawJSON)
	if err != nil {
		return nil, err
	}
	p.Titulo = strings.TrimSpace(p.Titulo)
	p.URL = urlLimpia(p.URL)
	if id < 1 || (claves["titulo"] && p.Titulo == "") {
		return nil, domainErrors.ErrPuntoInvalido
	}
	if claves["lat"] && (p.Lat < -12.20 || p.Lat > -11.90) {
		return nil, domainErrors.ErrPuntoInvalido
	}
	if claves["lon"] && (p.Lon < -77.30 || p.Lon > -76.90) {
		return nil, domainErrors.ErrPuntoInvalido
	}
	nid, err := r.devolverID(ctx, `
		UPDATE puntos_pucp SET
		  titulo = CASE WHEN $6::jsonb ? 'titulo' THEN $1 ELSE titulo END,
		  lat = CASE WHEN $6::jsonb ? 'lat' THEN $2 ELSE lat END,
		  lon = CASE WHEN $6::jsonb ? 'lon' THEN $3 ELSE lon END,
		  url = CASE WHEN $6::jsonb ? 'url' THEN NULLIF($4,'') ELSE url END,
		  geom = CASE
		    WHEN NOT ($6::jsonb ? 'lat' OR $6::jsonb ? 'lon') THEN geom
		    ELSE ST_SetSRID(ST_MakePoint(
		      CASE WHEN $6::jsonb ? 'lon' THEN $3 ELSE lon END,
		      CASE WHEN $6::jsonb ? 'lat' THEN $2 ELSE lat END
		    ), 4326)
		  END,
		  updated_at = now()
		WHERE id = $5 AND activo
		RETURNING id`,
		p.Titulo, p.Lat, p.Lon, p.URL, id, docJSON)
	if err != nil {
		return nil, err
	}
	p.ID = nid
	p.Activo = true
	return &p, nil
}

// ConsultaReservas builds the SQL query and arguments for reservas_jardin.
// An empty filter does not cast to date.
func ConsultaReservas(desde, hasta string) (string, []any) {
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

func (r *inventarioCampoRepository) ListarReservas(ctx context.Context, desde, hasta string) ([]entities.ReservaJardin, error) {
	q, args := ConsultaReservas(desde, hasta)
	rows, err := r.db.WithContext(ctx).Raw(q, args...).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []entities.ReservaJardin{}
	for rows.Next() {
		var item entities.ReservaJardin
		if err := rows.Scan(
			&item.ID, &item.JardinID, &item.Fecha, &item.HoraInicio, &item.HoraFin,
			&item.Estado, &item.Evento, &item.Unidad, &item.Origen, &item.Activo,
		); err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

func (r *inventarioCampoRepository) GuardarReserva(ctx context.Context, res entities.ReservaJardin) (*entities.ReservaJardin, error) {
	var id int64
	err := r.db.WithContext(ctx).Raw(`
		INSERT INTO reservas_jardin (jardin_id, fecha, hora_inicio, hora_fin, estado, evento, unidad, origen, origen_ref)
		VALUES ($1,$2,$3,$4,$5,$6,NULLIF($7,''),'ficticio', 'manual:' || $2 || $3 || $5 || left($6, 40))
		ON CONFLICT (origen_ref) DO UPDATE SET
		  estado = EXCLUDED.estado, evento = EXCLUDED.evento, unidad = EXCLUDED.unidad,
		  hora_inicio = EXCLUDED.hora_inicio, hora_fin = EXCLUDED.hora_fin, updated_at = now()
		RETURNING id`,
		res.JardinID, res.Fecha, res.HoraInicio, res.HoraFin, res.Estado, res.Evento, res.Unidad).Scan(&id).Error
	if err != nil {
		return nil, err
	}
	res.ID = id
	res.Origen = "ficticio"
	res.Activo = true
	return &res, nil
}

func (r *inventarioCampoRepository) ActualizarReserva(ctx context.Context, id int64, res entities.ReservaJardin, rawJSON []byte) (*entities.ReservaJardin, error) {
	claves, docJSON, err := ClavesPatch(rawJSON)
	if err != nil {
		return nil, err
	}
	if claves["origen"] && res.Origen != "" && res.Origen != "ficticio" {
		return nil, domainErrors.ErrReservaOrigen
	}
	if id < 1 {
		return nil, domainErrors.ErrReservaInvalida
	}
	if claves["estado"] {
		switch res.Estado {
		case "reservado", "realizado", "cancelado":
		default:
			return nil, domainErrors.ErrReservaInvalida
		}
	}
	if claves["hora_fin"] || claves["hora_inicio"] || claves["fecha"] || claves["evento"] {
		if (claves["hora_fin"] || claves["hora_inicio"]) && res.HoraFin <= res.HoraInicio {
			return nil, domainErrors.ErrReservaInvalida
		}
		if claves["fecha"] && res.Fecha == "" {
			return nil, domainErrors.ErrReservaInvalida
		}
		if claves["evento"] && strings.TrimSpace(res.Evento) == "" {
			return nil, domainErrors.ErrReservaInvalida
		}
	}
	nid, err := r.devolverID(ctx, `
		UPDATE reservas_jardin SET
		  jardin_id = CASE WHEN $9::jsonb ? 'jardin_id' THEN $1 ELSE jardin_id END,
		  fecha = CASE WHEN $9::jsonb ? 'fecha' THEN NULLIF($2,'')::date ELSE fecha END,
		  hora_inicio = CASE WHEN $9::jsonb ? 'hora_inicio' THEN NULLIF($3,'')::time ELSE hora_inicio END,
		  hora_fin = CASE WHEN $9::jsonb ? 'hora_fin' THEN NULLIF($4,'')::time ELSE hora_fin END,
		  estado = CASE WHEN $9::jsonb ? 'estado' THEN $5 ELSE estado END,
		  evento = CASE WHEN $9::jsonb ? 'evento' THEN $6 ELSE evento END,
		  unidad = CASE WHEN $9::jsonb ? 'unidad' THEN NULLIF($7,'') ELSE unidad END,
		  origen = 'ficticio', updated_at = now()
		WHERE id = $8 AND activo
		RETURNING id`,
		res.JardinID, res.Fecha, res.HoraInicio, res.HoraFin, res.Estado, res.Evento, res.Unidad, id, docJSON)
	if err != nil {
		return nil, err
	}
	res.ID = nid
	res.Origen = "ficticio"
	res.Activo = true
	return &res, nil
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

func (r *inventarioCampoRepository) ListarFichas(ctx context.Context, capa string) ([]entities.FichaCapa, error) {
	q, ok := fichaSelect(capa)
	if !ok {
		return nil, domainErrors.ErrCapaDesconocida
	}
	rows, err := r.db.WithContext(ctx).Raw(q).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []entities.FichaCapa{}
	for rows.Next() {
		var f models.FichaCapaModel
		if err := rows.Scan(&f.ID, &f.FeatureID, &f.Nombre, &f.Codigo, &f.Nota, &f.Clase, &f.Riego, &f.AreaM2, &f.PerimetroM, &f.Pertenecen, &f.Uso, &f.GeoJSON, &f.Activo); err != nil {
			return nil, err
		}
		out = append(out, *mapper.FichaCapaModelToEntity(&f))
	}
	return out, rows.Err()
}

func (r *inventarioCampoRepository) GuardarFicha(ctx context.Context, capa string, f entities.FichaCapa) (*entities.FichaCapa, error) {
	if !tablaPermitida(capa) {
		return nil, domainErrors.ErrCapaDesconocida
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
		return nil, domainErrors.ErrCapaDesconocida
	}
	id, err := r.devolverID(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	f.ID = id
	f.Activo = true
	return &f, nil
}

func (r *inventarioCampoRepository) ActualizarFicha(ctx context.Context, capa string, id int64, f entities.FichaCapa, rawJSON []byte) (*entities.FichaCapa, error) {
	claves, docJSON, err := ClavesPatch(rawJSON)
	if err != nil {
		return nil, err
	}
	if !tablaPermitida(capa) {
		return nil, domainErrors.ErrCapaDesconocida
	}
	f.FeatureID = strings.TrimSpace(f.FeatureID)
	if id < 1 || (claves["feature_id"] && f.FeatureID == "") {
		return nil, domainErrors.ErrFichaInvalida
	}
	geomFn := geomFnCapa(capa)
	var q string
	var args []any
	switch capa {
	case "fauna":
		q = fmt.Sprintf(`UPDATE fauna SET
		  feature_id = CASE WHEN $5::jsonb ? 'feature_id' THEN $1 ELSE feature_id END,
		  nombre = CASE WHEN $5::jsonb ? 'nombre' THEN NULLIF($2,'') ELSE nombre END,
		  geom = CASE WHEN $5::jsonb ? 'geojson' THEN COALESCE(%s(NULLIF($3,'')), geom) ELSE geom END,
		  updated_at = now() WHERE id = $4 AND activo RETURNING id`, geomFn)
		args = []any{f.FeatureID, f.Nombre, f.GeoJSON, id, docJSON}
	case "puertas":
		q = fmt.Sprintf(`UPDATE puertas SET
		  feature_id = CASE WHEN $6::jsonb ? 'feature_id' THEN $1 ELSE feature_id END,
		  codigo = CASE WHEN $6::jsonb ? 'codigo' THEN NULLIF($2,'') ELSE codigo END,
		  nombre = CASE WHEN $6::jsonb ? 'nombre' THEN NULLIF($3,'') ELSE nombre END,
		  geom = CASE WHEN $6::jsonb ? 'geojson' THEN COALESCE(%s(NULLIF($4,'')), geom) ELSE geom END,
		  updated_at = now() WHERE id = $5 AND activo RETURNING id`, geomFn)
		args = []any{f.FeatureID, f.Codigo, f.Nombre, f.GeoJSON, id, docJSON}
	case "playas_estacionamiento":
		q = fmt.Sprintf(`UPDATE playas_estacionamiento SET
		  feature_id = CASE WHEN $5::jsonb ? 'feature_id' THEN $1 ELSE feature_id END,
		  codigo = CASE WHEN $5::jsonb ? 'codigo' THEN NULLIF($2,'') ELSE codigo END,
		  geom = CASE WHEN $5::jsonb ? 'geojson' THEN COALESCE(%s(NULLIF($3,'')), geom) ELSE geom END,
		  updated_at = now() WHERE id = $4 AND activo RETURNING id`, geomFn)
		args = []any{f.FeatureID, f.Codigo, f.GeoJSON, id, docJSON}
	case "veredas_riesgo":
		q = fmt.Sprintf(`UPDATE veredas_riesgo SET
		  feature_id = CASE WHEN $5::jsonb ? 'feature_id' THEN $1 ELSE feature_id END,
		  nota = CASE WHEN $5::jsonb ? 'nota' THEN NULLIF($2,'') ELSE nota END,
		  geom = CASE WHEN $5::jsonb ? 'geojson' THEN COALESCE(%s(NULLIF($3,'')), geom) ELSE geom END,
		  updated_at = now() WHERE id = $4 AND activo RETURNING id`, geomFn)
		args = []any{f.FeatureID, f.Nota, f.GeoJSON, id, docJSON}
	case "xerofiticas":
		q = fmt.Sprintf(`UPDATE xerofiticas SET
		  feature_id = CASE WHEN $8::jsonb ? 'feature_id' THEN $1 ELSE feature_id END,
		  clase = CASE WHEN $8::jsonb ? 'clase' THEN NULLIF($2,'') ELSE clase END,
		  riego = CASE WHEN $8::jsonb ? 'riego' THEN NULLIF($3,'') ELSE riego END,
		  area_m2 = CASE WHEN $8::jsonb ? 'area_m2' THEN $4 ELSE area_m2 END,
		  perimetro_m = CASE WHEN $8::jsonb ? 'perimetro_m' THEN $5 ELSE perimetro_m END,
		  geom = CASE WHEN $8::jsonb ? 'geojson' THEN COALESCE(%s(NULLIF($6,'')), geom) ELSE geom END,
		  updated_at = now() WHERE id = $7 AND activo RETURNING id`, geomFn)
		args = []any{f.FeatureID, f.Clase, f.Riego, f.AreaM2, f.PerimetroM, f.GeoJSON, id, docJSON}
	case "jardines_reserva":
		q = fmt.Sprintf(`UPDATE jardines_reserva SET
		  feature_id = CASE WHEN $12::jsonb ? 'feature_id' THEN $1 ELSE feature_id END,
		  codigo = CASE WHEN $12::jsonb ? 'codigo' THEN NULLIF($2,'') ELSE codigo END,
		  nombre = CASE WHEN $12::jsonb ? 'nombre' THEN NULLIF($3,'') ELSE nombre END,
		  uso = CASE WHEN $12::jsonb ? 'uso' THEN NULLIF($4,'') ELSE uso END,
		  riego_act = CASE WHEN $12::jsonb ? 'riego' THEN NULLIF($5,'') ELSE riego_act END,
		  referencia = CASE WHEN $12::jsonb ? 'nota' THEN NULLIF($6,'') ELSE referencia END,
		  pertenecen = CASE WHEN $12::jsonb ? 'pertenecen' THEN NULLIF($7,'') ELSE pertenecen END,
		  area_m2 = CASE WHEN $12::jsonb ? 'area_m2' THEN $8 ELSE area_m2 END,
		  perimetro_m = CASE WHEN $12::jsonb ? 'perimetro_m' THEN $9 ELSE perimetro_m END,
		  geom = CASE WHEN $12::jsonb ? 'geojson' THEN COALESCE(%s(NULLIF($10,'')), geom) ELSE geom END,
		  updated_at = now() WHERE id = $11 AND activo RETURNING id`, geomFn)
		args = []any{f.FeatureID, f.Codigo, f.Nombre, f.Uso, f.Riego, f.Nota, f.Pertenecen, f.AreaM2, f.PerimetroM, f.GeoJSON, id, docJSON}
	default:
		return nil, domainErrors.ErrCapaDesconocida
	}
	nid, err := r.devolverID(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	f.ID = nid
	f.Activo = true
	return &f, nil
}

func (r *inventarioCampoRepository) Baja(ctx context.Context, tabla string, id int64) error {
	if !tablaPermitida(tabla) || id < 1 {
		return domainErrors.ErrRegistroNoEncontrado
	}
	res := r.db.WithContext(ctx).Exec(fmt.Sprintf(`UPDATE %s SET activo = FALSE, updated_at = now() WHERE id = $1 AND activo`, tabla), id)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected != 1 {
		return domainErrors.ErrRegistroNoEncontrado
	}
	return nil
}
