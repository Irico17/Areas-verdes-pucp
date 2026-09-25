package auditoria

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"
)

// ErrConfirmacion: hay filas editadas después del lote y falta confirmar.
var ErrConfirmacion = errors.New("confirmacion")

// ErrNoEncontrado: el lote no existe.
var ErrNoEncontrado = errors.New("no encontrado")

// InputError es un 400.
type InputError struct{ Reason string }

func (e InputError) Error() string { return e.Reason }

// Fila es una fila del lote. Antes es el estado previo; null en un alta.
type Fila struct {
	EntidadID string          `json:"entidad_id"`
	Accion    string          `json:"accion"`
	Antes     json.RawMessage `json:"antes"`
	Despues   json.RawMessage `json:"despues"`
}

// Excluida es una fila que la reversión no pisa.
type Excluida struct {
	EntidadID string `json:"entidad_id"`
	Motivo    string `json:"motivo"`
}

// Reporte es el resultado de revertir.
type Reporte struct {
	LoteID     int64      `json:"lote_id"`
	Revertidas []string   `json:"revertidas"`
	Excluidas  []Excluida `json:"excluidas"`
}

// Evento es una entrada del timeline. El actor es el usuario de sesión.
type Evento struct {
	ID        int64  `json:"id"`
	Entidad   string `json:"entidad"`
	EntidadID string `json:"entidad_id"`
	Accion    string `json:"accion"`
	UsuarioID int64  `json:"usuario_id"`
	Usuario   string `json:"usuario"`
	Nombre    string `json:"nombre"`
	LoteID    *int64 `json:"lote_id,omitempty"`
	CreatedAt string `json:"created_at"`
}

// Filtro acota el historial (zona, origen, fechas).
type Filtro struct {
	Entidad   string
	EntidadID string
	Zona      string
	Origen    string
	Desde     string
	Hasta     string
}

// Store persiste lotes sobre la tabla cambios de la ola 0.
type Store struct {
	db *gorm.DB
}

func NewStore(db *gorm.DB) *Store { return &Store{db: db} }

// Importar escribe el lote y un cambio por fila, y aplica el después.
func (s *Store) Importar(ctx context.Context, usuarioID int64, entidad string, filas []Fila) (int64, error) {
	entidad = strings.TrimSpace(entidad)
	if err := validarEntidad(entidad); err != nil {
		return 0, err
	}
	if usuarioID < 1 {
		return 0, InputError{Reason: "usuario de sesión obligatorio"}
	}
	if len(filas) == 0 {
		return 0, InputError{Reason: "el lote no tiene filas"}
	}
	var loteID int64
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Raw(`
			INSERT INTO lotes_importacion (entidad, estado, usuario_id, filas)
			VALUES ($1, 'confirmado', $2, $3)
			RETURNING id`, entidad, usuarioID, len(filas)).Row().Scan(&loteID); err != nil {
			return err
		}
		for _, fila := range filas {
			id, accion, antes, despues, err := normalizar(fila)
			if err != nil {
				return err
			}
			if err := aplicar(tx, entidad, id, accion, despues, false); err != nil {
				return err
			}
			if err := insertarCambio(tx, entidad, id, "importacion", antes, despues, usuarioID, &loteID); err != nil {
				return err
			}
		}
		return nil
	})
	return loteID, err
}

// Editar registra un cambio posterior de un usuario. La reversión del lote no lo pisa.
func (s *Store) Editar(ctx context.Context, usuarioID int64, entidad, entidadID string, despues json.RawMessage) error {
	if err := validarEntidad(entidad); err != nil {
		return err
	}
	entidadID = strings.TrimSpace(entidadID)
	if entidadID == "" || usuarioID < 1 {
		return InputError{Reason: "fila y usuario de sesión son obligatorios"}
	}
	if len(despues) == 0 || string(despues) == "null" {
		return InputError{Reason: "despues es obligatorio"}
	}
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		antes, err := leerActual(tx, entidad, entidadID)
		if err != nil {
			return err
		}
		if err := aplicar(tx, entidad, entidadID, "edicion", despues, false); err != nil {
			return err
		}
		return insertarCambio(tx, entidad, entidadID, "edicion", antes, despues, usuarioID, nil)
	})
}

// Revertir aplica los antes en orden inverso. Una fila con cambio posterior se excluye.
func (s *Store) Revertir(ctx context.Context, loteID, usuarioID int64, confirmar bool) (Reporte, error) {
	rep := Reporte{LoteID: loteID, Revertidas: []string{}, Excluidas: []Excluida{}}
	if loteID < 1 || usuarioID < 1 {
		return rep, InputError{Reason: "lote y usuario de sesión son obligatorios"}
	}
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var estado, entidad string
		err := tx.Raw(`
			SELECT estado, entidad FROM lotes_importacion WHERE id = $1 FOR UPDATE`, loteID).
			Row().Scan(&estado, &entidad)
		if errors.Is(err, sql.ErrNoRows) {
			return ErrNoEncontrado
		}
		if err != nil {
			return err
		}
		if estado != "confirmado" {
			return InputError{Reason: "el lote ya fue revertido"}
		}
		type cambio struct {
			id        int64
			entidadID string
			accion    string
			antes     []byte
		}
		rows, err := tx.Raw(`
			SELECT id, entidad_id, accion, antes
			FROM cambios
			WHERE lote_id = $1 AND accion = 'importacion'
			ORDER BY id DESC`, loteID).Rows()
		if err != nil {
			return err
		}
		defer rows.Close()
		var lista []cambio
		for rows.Next() {
			var c cambio
			if err := rows.Scan(&c.id, &c.entidadID, &c.accion, &c.antes); err != nil {
				return err
			}
			lista = append(lista, c)
		}
		if err := rows.Err(); err != nil {
			return err
		}
		type trabajo struct {
			c     cambio
			salta bool
		}
		plan := make([]trabajo, 0, len(lista))
		for _, c := range lista {
			var n int
			if err := tx.Raw(`
				SELECT count(*) FROM cambios
				WHERE entidad = $1 AND entidad_id = $2 AND id > $3
				  AND accion IN ('alta', 'edicion', 'baja')
				  AND (lote_id IS NULL OR lote_id <> $4)`,
				entidad, c.entidadID, c.id, loteID).Scan(&n).Error; err != nil {
				return err
			}
			if n > 0 {
				rep.Excluidas = append(rep.Excluidas, Excluida{
					EntidadID: c.entidadID,
					Motivo:    "hay una edición posterior; no se pisa",
				})
				plan = append(plan, trabajo{c: c, salta: true})
				continue
			}
			plan = append(plan, trabajo{c: c})
		}
		if len(rep.Excluidas) > 0 && !confirmar {
			return ErrConfirmacion
		}
		for _, item := range plan {
			if item.salta {
				continue
			}
			if err := aplicar(tx, entidad, item.c.entidadID, "reversion", item.c.antes, true); err != nil {
				return err
			}
			despues, err := leerActual(tx, entidad, item.c.entidadID)
			if err != nil {
				return err
			}
			lote := loteID
			if err := insertarCambio(tx, entidad, item.c.entidadID, "reversion", item.c.antes, despues, usuarioID, &lote); err != nil {
				return err
			}
			rep.Revertidas = append(rep.Revertidas, item.c.entidadID)
		}
		return tx.Exec(`
			UPDATE lotes_importacion
			SET estado = 'revertido', revertido_en = now()
			WHERE id = $1`, loteID).Error
	})
	if errors.Is(err, ErrConfirmacion) {
		return rep, err
	}
	return rep, err
}

// Timeline lista los cambios de una fila. El actor es el usuario de sesión.
func (s *Store) Timeline(ctx context.Context, f Filtro) ([]Evento, error) {
	return s.listar(ctx, f, true)
}

// Historial es el listado filtrable por zona, origen y fechas.
func (s *Store) Historial(ctx context.Context, f Filtro) ([]Evento, error) {
	return s.listar(ctx, f, false)
}

func (s *Store) listar(ctx context.Context, f Filtro, unaFila bool) ([]Evento, error) {
	if unaFila {
		if strings.TrimSpace(f.Entidad) == "" || strings.TrimSpace(f.EntidadID) == "" {
			return nil, InputError{Reason: "entidad y entidad_id son obligatorios"}
		}
	}
	desde, hasta, err := rango(f.Desde, f.Hasta)
	if err != nil {
		return nil, err
	}
	rows, err := s.db.WithContext(ctx).Raw(`
		SELECT c.id, c.entidad, c.entidad_id, c.accion, c.usuario_id,
		       u.usuario, u.nombre, c.lote_id, c.created_at
		FROM cambios c
		JOIN usuarios u ON u.id = c.usuario_id
		WHERE ($1 = '' OR c.entidad = $1)
		  AND ($2 = '' OR c.entidad_id = $2)
		  AND ($3 = '' OR c.antes->>'zona' = $3 OR c.despues->>'zona' = $3)
		  AND ($4 = '' OR c.antes->>'origen' = $4 OR c.despues->>'origen' = $4)
		  AND ($5::timestamptz IS NULL OR c.created_at >= $5)
		  AND ($6::timestamptz IS NULL OR c.created_at < $6)
		ORDER BY c.id`,
		strings.TrimSpace(f.Entidad), strings.TrimSpace(f.EntidadID),
		strings.TrimSpace(f.Zona), strings.TrimSpace(f.Origen),
		desde, hasta,
	).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []Evento{}
	for rows.Next() {
		var ev Evento
		var lote *int64
		var when time.Time
		if err := rows.Scan(&ev.ID, &ev.Entidad, &ev.EntidadID, &ev.Accion, &ev.UsuarioID, &ev.Usuario, &ev.Nombre, &lote, &when); err != nil {
			return nil, err
		}
		ev.LoteID = lote
		ev.CreatedAt = when.UTC().Format(time.RFC3339)
		out = append(out, ev)
	}
	return out, rows.Err()
}

func validarEntidad(entidad string) error {
	switch entidad {
	case "catalogos", "areas_verdes", "actividades":
		return nil
	default:
		return InputError{Reason: "entidad no importable"}
	}
}

func normalizar(fila Fila) (string, string, json.RawMessage, json.RawMessage, error) {
	id := strings.TrimSpace(fila.EntidadID)
	accion := strings.TrimSpace(fila.Accion)
	if id == "" {
		return "", "", nil, nil, InputError{Reason: "entidad_id es obligatorio"}
	}
	switch accion {
	case "alta", "edicion", "baja":
	default:
		return "", "", nil, nil, InputError{Reason: "accion debe ser alta, edicion o baja"}
	}
	antes := fila.Antes
	if len(antes) == 0 {
		antes = json.RawMessage("null")
	}
	despues := fila.Despues
	if accion != "baja" && (len(despues) == 0 || string(despues) == "null") {
		return "", "", nil, nil, InputError{Reason: "despues es obligatorio"}
	}
	if len(despues) == 0 {
		despues = json.RawMessage("null")
	}
	if !json.Valid(antes) || !json.Valid(despues) {
		return "", "", nil, nil, InputError{Reason: "antes y despues deben ser JSON"}
	}
	return id, accion, antes, despues, nil
}

func insertarCambio(tx *gorm.DB, entidad, entidadID, accion string, antes, despues json.RawMessage, usuarioID int64, loteID *int64) error {
	return tx.Exec(`
		INSERT INTO cambios (entidad, entidad_id, accion, antes, despues, usuario_id, lote_id)
		VALUES ($1, $2, $3, $4::jsonb, $5::jsonb, $6, $7)`,
		entidad, entidadID, accion, string(antes), string(despues), usuarioID, loteID,
	).Error
}

type snap struct {
	Clase   string `json:"clase"`
	Codigo  string `json:"codigo"`
	Nombre  string `json:"nombre"`
	Activo  *bool  `json:"activo"`
	Orden   *int   `json:"orden"`
	Titulo  string `json:"titulo"`
	Detalle string `json:"detalle"`
	Estado  string `json:"estado"`
	Zona    string `json:"zona"`
	Origen  string `json:"origen"`
}

func aplicar(tx *gorm.DB, entidad, entidadID, accion string, doc json.RawMessage, reversion bool) error {
	vacio := len(doc) == 0 || string(doc) == "null"
	if reversion && vacio {
		return bajaLogica(tx, entidad, entidadID)
	}
	if accion == "baja" {
		return bajaLogica(tx, entidad, entidadID)
	}
	var s snap
	if err := json.Unmarshal(doc, &s); err != nil {
		return InputError{Reason: "el estado no es un objeto"}
	}
	switch entidad {
	case "catalogos":
		return aplicarCatalogo(tx, entidadID, s)
	case "actividades":
		return aplicarActividad(tx, entidadID, s)
	case "areas_verdes":
		return aplicarArea(tx, entidadID, s)
	default:
		return InputError{Reason: "entidad no importable"}
	}
}

func bajaLogica(tx *gorm.DB, entidad, entidadID string) error {
	switch entidad {
	case "catalogos":
		return tx.Exec(`UPDATE catalogos SET activo = false WHERE id = $1`, entidadID).Error
	case "actividades":
		return tx.Exec(`
			UPDATE actividades
			SET archivada_en = COALESCE(archivada_en, now()), updated_at = now()
			WHERE id = $1`, entidadID).Error
	case "podas":
		return tx.Exec(`
			UPDATE podas
			SET archivada_en = COALESCE(archivada_en, now()), updated_at = now()
			WHERE id::text = $1`, entidadID).Error
	case "vivero_registros":
		return tx.Exec(`
			UPDATE vivero_registros
			SET archivada_en = COALESCE(archivada_en, now()), updated_at = now()
			WHERE id::text = $1`, entidadID).Error
	case "areas_verdes":
		return tx.Exec(`
			UPDATE areas_verdes
			SET activo = false, referencia = COALESCE(referencia, 'baja lógica'), updated_at = now()
			WHERE feature_id = $1`, entidadID).Error
	case "medidas_palmera":
		return tx.Exec(`DELETE FROM medidas_palmera WHERE ejemplar_id::text = $1`, entidadID).Error
	default:
		if spec, ok := bajaPorActivo[entidad]; ok {
			q := fmt.Sprintf(`UPDATE %s SET activo = false WHERE %s = $1`, spec.tabla, spec.col)
			return tx.Exec(q, entidadID).Error
		}
		return InputError{Reason: "entidad no importable"}
	}
}

// bajaPorActivo deshace un alta del importador (frente 3A) sin borrar la fila.
var bajaPorActivo = map[string]struct{ tabla, col string }{
	"lugares":                {"lugares", "id::text"},
	"zonas_supervision":      {"zonas_supervision", "id::text"},
	"poligonos_cuadrilla":    {"poligonos_cuadrilla", "id::text"},
	"cuadrillas":             {"cuadrillas", "id"},
	"ejemplares":             {"ejemplares", "id::text"},
	"especies":               {"especies", "id::text"},
	"tachos":                 {"tachos", "id::text"},
	"bebederos":              {"bebederos", "id::text"},
	"puntos_pucp":            {"puntos_pucp", "id::text"},
	"reservas_jardin":        {"reservas_jardin", "id::text"},
	"fauna":                  {"fauna", "feature_id"},
	"puertas":                {"puertas", "feature_id"},
	"playas_estacionamiento": {"playas_estacionamiento", "feature_id"},
	"veredas_riesgo":         {"veredas_riesgo", "feature_id"},
	"xerofiticas":            {"xerofiticas", "feature_id"},
	"jardines_reserva":       {"jardines_reserva", "feature_id"},
}

func aplicarCatalogo(tx *gorm.DB, id string, s snap) error {
	activo := true
	if s.Activo != nil {
		activo = *s.Activo
	}
	orden := 0
	if s.Orden != nil {
		orden = *s.Orden
	}
	res := tx.Exec(`
		UPDATE catalogos
		SET clase = COALESCE(NULLIF($2, ''), clase),
		    codigo = COALESCE(NULLIF($3, ''), codigo),
		    nombre = COALESCE(NULLIF($4, ''), nombre),
		    activo = $5,
		    orden = $6
		WHERE id = $1`, id, s.Clase, s.Codigo, s.Nombre, activo, orden)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return InputError{Reason: "fila de catálogo no encontrada"}
	}
	return nil
}

func aplicarActividad(tx *gorm.DB, id string, s snap) error {
	res := tx.Exec(`
		UPDATE actividades
		SET titulo = COALESCE(NULLIF($2, ''), titulo),
		    detalle = CASE WHEN $3 = '' THEN detalle ELSE $3 END,
		    estado = COALESCE(NULLIF($4, ''), estado),
		    zona_feature_id = COALESCE(NULLIF($5, ''), zona_feature_id),
		    updated_at = now()
		WHERE id = $1`, id, s.Titulo, s.Detalle, s.Estado, s.Zona)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return InputError{Reason: "labor no encontrada"}
	}
	return nil
}

func aplicarArea(tx *gorm.DB, featureID string, s snap) error {
	res := tx.Exec(`
		UPDATE areas_verdes
		SET nombre = COALESCE(NULLIF($2, ''), nombre),
		    referencia = COALESCE(NULLIF($3, ''), referencia),
		    updated_at = now()
		WHERE feature_id = $1`, featureID, s.Nombre, s.Detalle)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return InputError{Reason: "área no encontrada"}
	}
	return nil
}

func leerActual(tx *gorm.DB, entidad, entidadID string) (json.RawMessage, error) {
	var raw string
	var err error
	switch entidad {
	case "catalogos":
		err = tx.Raw(`
			SELECT json_build_object(
				'clase', clase, 'codigo', codigo, 'nombre', nombre,
				'activo', activo, 'orden', orden
			)::text
			FROM catalogos WHERE id = $1`, entidadID).Row().Scan(&raw)
	case "actividades":
		err = tx.Raw(`
			SELECT json_build_object(
				'titulo', titulo, 'detalle', detalle, 'estado', estado,
				'zona', zona_feature_id
			)::text
			FROM actividades WHERE id = $1`, entidadID).Row().Scan(&raw)
	case "areas_verdes":
		err = tx.Raw(`
			SELECT json_build_object('nombre', nombre, 'detalle', referencia)::text
			FROM areas_verdes WHERE feature_id = $1`, entidadID).Row().Scan(&raw)
	default:
		return nil, InputError{Reason: "entidad no importable"}
	}
	if err != nil {
		return nil, err
	}
	return json.RawMessage(raw), nil
}

func rango(desde, hasta string) (*time.Time, *time.Time, error) {
	d, err := parseCuando(desde)
	if err != nil {
		return nil, nil, InputError{Reason: "desde no es una fecha"}
	}
	h, err := parseCuando(hasta)
	if err != nil {
		return nil, nil, InputError{Reason: "hasta no es una fecha"}
	}
	return d, h, nil
}

func parseCuando(v string) (*time.Time, error) {
	v = strings.TrimSpace(v)
	if v == "" {
		return nil, nil
	}
	t, err := time.Parse(time.RFC3339, v)
	if err != nil {
		t, err = time.Parse("2006-01-02", v)
		if err != nil {
			return nil, err
		}
	}
	u := t.UTC()
	return &u, nil
}
