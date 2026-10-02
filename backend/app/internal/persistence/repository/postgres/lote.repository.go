// Package postgres provides PostgreSQL implementations of persistence contracts.
package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"

	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/contracts"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/entities"
	apperrors "github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/errors"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/persistence/database"
)

type loteRepository struct {
	db *gorm.DB
}

// NewLoteRepository creates a new ILoteRepository instance.
func NewLoteRepository(db *gorm.DB) contracts.ILoteRepository {
	return &loteRepository{db: db}
}

// Importar writes the batch and an audit change per row, applying the "despues" state.
func (s *loteRepository) Importar(ctx context.Context, usuarioID int64, entidad string, filas []entities.FilaLote) (int64, error) {
	entidad = strings.TrimSpace(entidad)
	if err := validarEntidad(entidad); err != nil {
		return 0, err
	}
	if usuarioID < 1 {
		return 0, apperrors.InputError{Reason: "usuario de sesión obligatorio"}
	}
	if len(filas) == 0 {
		return 0, apperrors.InputError{Reason: "el lote no tiene filas"}
	}
	var loteID int64
	db := database.DBFromContext(ctx, s.db)
	err := db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
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

// Editar records a subsequent manual change by a user. Later reversions will not overwrite it.
func (s *loteRepository) Editar(ctx context.Context, usuarioID int64, entidad, entidadID string, despues []byte) error {
	if err := validarEntidad(entidad); err != nil {
		return err
	}
	entidadID = strings.TrimSpace(entidadID)
	if entidadID == "" || usuarioID < 1 {
		return apperrors.InputError{Reason: "fila y usuario de sesión son obligatorios"}
	}
	if len(despues) == 0 || string(despues) == "null" {
		return apperrors.InputError{Reason: "despues es obligatorio"}
	}
	db := database.DBFromContext(ctx, s.db)
	return db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
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

// Revertir applies the "antes" snapshots in reverse order. Rows with subsequent edits are excluded.
func (s *loteRepository) Revertir(ctx context.Context, loteID, usuarioID int64, confirmar bool) (*entities.ReporteReversion, error) {
	rep := &entities.ReporteReversion{LoteID: loteID, Revertidas: []string{}, Excluidas: []entities.Excluida{}}
	if loteID < 1 || usuarioID < 1 {
		return rep, apperrors.InputError{Reason: "lote y usuario de sesión son obligatorios"}
	}
	db := database.DBFromContext(ctx, s.db)
	err := db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var estado, entidad string
		err := tx.Raw(`
			SELECT estado, entidad FROM lotes_importacion WHERE id = $1 FOR UPDATE`, loteID).
			Row().Scan(&estado, &entidad)
		if errors.Is(err, sql.ErrNoRows) {
			return apperrors.ErrLoteNoEncontrado
		}
		if err != nil {
			return err
		}
		if estado != "confirmado" {
			return apperrors.InputError{Reason: "el lote ya fue revertido"}
		}
		type cambioRow struct {
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
		var lista []cambioRow
		for rows.Next() {
			var c cambioRow
			if err := rows.Scan(&c.id, &c.entidadID, &c.accion, &c.antes); err != nil {
				return err
			}
			lista = append(lista, c)
		}
		if err := rows.Err(); err != nil {
			return err
		}
		type trabajo struct {
			c     cambioRow
			salta bool
		}
		plan := make([]trabajo, 0, len(lista))
		for _, c := range lista {
			var n int
			if err := tx.Raw(`
				SELECT count(*) FROM cambios
				WHERE entidad = $1 AND entidad_id = $2 AND id > $3
				  AND (
				    accion IN ('alta', 'edicion', 'baja')
				    OR (accion = 'importacion' AND lote_id IS DISTINCT FROM $4)
				  )`,
				entidad, c.entidadID, c.id, loteID).Scan(&n).Error; err != nil {
				return err
			}
			if n > 0 {
				rep.Excluidas = append(rep.Excluidas, entities.Excluida{
					EntidadID: c.entidadID,
					Motivo:    "hay una edición posterior; no se pisa",
				})
				plan = append(plan, trabajo{c: c, salta: true})
				continue
			}
			plan = append(plan, trabajo{c: c})
		}
		if len(rep.Excluidas) > 0 && !confirmar {
			return apperrors.ErrConfirmacion
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
	return rep, err
}

// Timeline returns audit events for a specific row.
func (s *loteRepository) Timeline(ctx context.Context, f entities.FiltroAuditoria) ([]entities.EventoAuditoria, error) {
	return s.listar(ctx, f, true)
}

// Historial returns audit events filtered by criteria.
func (s *loteRepository) Historial(ctx context.Context, f entities.FiltroAuditoria) ([]entities.EventoAuditoria, error) {
	return s.listar(ctx, f, false)
}

func (s *loteRepository) listar(ctx context.Context, f entities.FiltroAuditoria, unaFila bool) ([]entities.EventoAuditoria, error) {
	if unaFila {
		if strings.TrimSpace(f.Entidad) == "" || strings.TrimSpace(f.EntidadID) == "" {
			return nil, apperrors.InputError{Reason: "entidad y entidad_id son obligatorios"}
		}
	}
	desde, hasta, err := rango(f.Desde, f.Hasta)
	if err != nil {
		return nil, err
	}
	db := database.DBFromContext(ctx, s.db)
	rows, err := db.WithContext(ctx).Raw(`
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
	out := []entities.EventoAuditoria{}
	for rows.Next() {
		var ev entities.EventoAuditoria
		var lote *int64
		var when time.Time
		if err := rows.Scan(&ev.ID, &ev.Entidad, &ev.EntidadID, &ev.Accion, &ev.UsuarioID, &ev.Usuario, &ev.Nombre, &lote, &when); err != nil {
			return nil, err
		}
		ev.LoteID = lote
		ev.CreatedAt = when.UTC()
		out = append(out, ev)
	}
	return out, rows.Err()
}

func validarEntidad(entidad string) error {
	switch entidad {
	case "catalogos", "areas_verdes", "actividades":
		return nil
	default:
		return apperrors.InputError{Reason: "entidad no importable"}
	}
}

func normalizar(fila entities.FilaLote) (string, string, json.RawMessage, json.RawMessage, error) {
	id := strings.TrimSpace(fila.EntidadID)
	accion := strings.TrimSpace(fila.Accion)
	if id == "" {
		return "", "", nil, nil, apperrors.InputError{Reason: "entidad_id es obligatorio"}
	}
	switch accion {
	case "alta", "edicion", "baja":
	default:
		return "", "", nil, nil, apperrors.InputError{Reason: "accion debe ser alta, edicion o baja"}
	}
	antes := fila.Antes
	if len(antes) == 0 {
		antes = json.RawMessage("null")
	}
	despues := fila.Despues
	if accion != "baja" && (len(despues) == 0 || string(despues) == "null") {
		return "", "", nil, nil, apperrors.InputError{Reason: "despues es obligatorio"}
	}
	if len(despues) == 0 {
		despues = json.RawMessage("null")
	}
	if !json.Valid(antes) || !json.Valid(despues) {
		return "", "", nil, nil, apperrors.InputError{Reason: "antes y despues deben ser JSON"}
	}
	return id, accion, antes, despues, nil
}

func insertarCambio(tx *gorm.DB, entidad, entidadID, accion string, antes, despues json.RawMessage, usuarioID int64, loteID *int64) error {
	var antesVal, despuesVal any
	if len(antes) > 0 {
		antesVal = string(antes)
	}
	if len(despues) > 0 {
		despuesVal = string(despues)
	}
	return tx.Exec(`
		INSERT INTO cambios (entidad, entidad_id, accion, antes, despues, usuario_id, lote_id)
		VALUES ($1, $2, $3, $4::jsonb, $5::jsonb, $6, $7)`,
		entidad, entidadID, accion, antesVal, despuesVal, usuarioID, loteID,
	).Error
}

func aplicar(tx *gorm.DB, entidad, entidadID, accion string, doc json.RawMessage, reversion bool) error {
	vacio := len(doc) == 0 || string(doc) == "null"
	if reversion && vacio {
		return bajaLogica(tx, entidad, entidadID)
	}
	if accion == "baja" {
		return bajaLogica(tx, entidad, entidadID)
	}
	var s entities.SnapAuditoria
	if err := json.Unmarshal(doc, &s); err != nil {
		return apperrors.InputError{Reason: "el estado no es un objeto"}
	}
	switch entidad {
	case "catalogos":
		return aplicarCatalogo(tx, entidadID, accion, s)
	case "actividades":
		return aplicarActividad(tx, entidadID, accion, s)
	case "areas_verdes":
		return aplicarArea(tx, entidadID, accion, s)
	default:
		return apperrors.InputError{Reason: "entidad no importable"}
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
		// Same as the old API (auditoria/store.go): reverting an import removes the
		// measurement rows that the import itself created. This is not loaded
		// catastro data; every other entity reverts by logical deactivation.
		return tx.Exec(`DELETE FROM medidas_palmera WHERE ejemplar_id::text = $1`, entidadID).Error
	default:
		if spec, ok := bajaPorActivo[entidad]; ok {
			q := fmt.Sprintf(`UPDATE %s SET activo = false WHERE %s = $1`, spec.tabla, spec.col)
			return tx.Exec(q, entidadID).Error
		}
		return apperrors.InputError{Reason: "entidad no importable"}
	}
}

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
	"sectores_capataz":       {"sectores_capataz", "id::text"},
	"vias":                   {"vias", "id::text"},
}

func aplicarCatalogo(tx *gorm.DB, id, accion string, s entities.SnapAuditoria) error {
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
	if res.RowsAffected > 0 {
		return nil
	}
	if accion != "alta" {
		return apperrors.InputError{Reason: "fila de catálogo no encontrada"}
	}
	if strings.TrimSpace(s.Clase) == "" || strings.TrimSpace(s.Codigo) == "" || strings.TrimSpace(s.Nombre) == "" {
		return apperrors.InputError{Reason: "el alta de catálogo exige clase, código y nombre"}
	}
	if err := tx.Exec(`
		INSERT INTO catalogos (id, clase, codigo, nombre, activo, orden)
		VALUES ($1::bigint, $2, $3, $4, $5, $6)`, id, s.Clase, s.Codigo, s.Nombre, activo, orden).Error; err != nil {
		return err
	}
	return tx.Exec(`
		SELECT setval(
			pg_get_serial_sequence('catalogos', 'id'),
			(SELECT MAX(id) FROM catalogos)
		)`).Error
}

func aplicarActividad(tx *gorm.DB, id, accion string, s entities.SnapAuditoria) error {
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
	if res.RowsAffected > 0 {
		return nil
	}
	if accion != "alta" {
		return apperrors.InputError{Reason: "labor no encontrada"}
	}
	titulo, estado, err := CamposAltaActividad(s)
	if err != nil {
		return err
	}
	return tx.Exec(`
		INSERT INTO actividades (id, tipo, estado, titulo, detalle, zona_feature_id, origen_ref, ejecutor)
		VALUES ($1::uuid, 'inspeccion', $2, $3, $4, NULLIF($5, ''), $1, 'propia')`,
		id, estado, titulo, s.Detalle, s.Zona).Error
}

// CamposAltaActividad validates required fields for a new labor and returns (titulo, estado, err).
func CamposAltaActividad(s entities.SnapAuditoria) (string, string, error) {
	titulo := strings.TrimSpace(s.Titulo)
	estado := strings.TrimSpace(s.Estado)
	if titulo == "" || estado == "" {
		return "", "", apperrors.InputError{Reason: "el alta de una labor exige título y estado"}
	}
	return titulo, estado, nil
}

func aplicarArea(tx *gorm.DB, featureID, accion string, s entities.SnapAuditoria) error {
	res := tx.Exec(`
		UPDATE areas_verdes
		SET nombre = COALESCE(NULLIF($2, ''), nombre),
		    referencia = COALESCE(NULLIF($3, ''), referencia),
		    updated_at = now()
		WHERE feature_id = $1`, featureID, s.Nombre, s.Detalle)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected > 0 {
		return nil
	}
	if accion != "alta" {
		return apperrors.InputError{Reason: "área no encontrada"}
	}
	nombre := strings.TrimSpace(s.Nombre)
	if nombre == "" {
		nombre = featureID
	}
	return tx.Exec(`
		INSERT INTO areas_verdes (feature_id, source_index, nombre, referencia, origen_ref, activo)
		SELECT $1, COALESCE((SELECT MAX(source_index) FROM areas_verdes), 0) + 1,
		       $2, NULLIF($3, ''), $1, TRUE`,
		featureID, nombre, s.Detalle).Error
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
		spec, ok := bajaPorActivo[entidad]
		if !ok {
			return nil, apperrors.InputError{Reason: "entidad no importable"}
		}
		q := fmt.Sprintf(`SELECT to_jsonb(t)::text FROM %s t WHERE %s = $1`, spec.tabla, spec.col)
		err = tx.Raw(q, entidadID).Row().Scan(&raw)
	}
	if err != nil {
		return nil, err
	}
	return json.RawMessage(raw), nil
}

func rango(desde, hasta string) (*time.Time, *time.Time, error) {
	d, err := parseCuando(desde)
	if err != nil {
		return nil, nil, apperrors.InputError{Reason: "desde no es una fecha"}
	}
	h, err := parseCuando(hasta)
	if err != nil {
		return nil, nil, apperrors.InputError{Reason: "hasta no es una fecha"}
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
