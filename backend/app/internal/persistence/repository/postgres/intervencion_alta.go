package postgres

import (
	"context"
	"errors"
	"strings"

	"gorm.io/gorm"

	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/entities"
	domainErrors "github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/errors"
)

func validarAltaCatalogo(tx *gorm.DB, in entities.NuevaIntervencion) error {
	if strings.TrimSpace(in.LugarLibre) != "" || strings.TrimSpace(in.LugarTexto) != "" {
		return domainErrors.InputError{Reason: "el lugar no se escribe a mano; elija uno del catálogo"}
	}
	if v := strings.TrimSpace(in.Origen); v != "" {
		ok, err := catalogoActivoTx(tx, "fuente", v)
		if err != nil {
			return err
		}
		if !ok {
			return domainErrors.InputError{Reason: "origen no está en el catálogo activo"}
		}
	}
	if v := strings.TrimSpace(in.NivelRiesgo); v != "" {
		ok, err := catalogoActivoTx(tx, "nivel_riesgo", v)
		if err != nil {
			return err
		}
		if !ok {
			return domainErrors.InputError{Reason: "nivel_riesgo no está en el catálogo activo"}
		}
	}
	clase := strings.TrimSpace(in.Clase)
	subtipo := strings.TrimSpace(in.Subtipo)
	if subtipo != "" && clase == "" {
		return domainErrors.InputError{Reason: "el tipo de actividad exige su clase"}
	}
	if clase != "" {
		ok, err := catalogoActivoTx(tx, "clase_actividad", clase)
		if err != nil {
			return err
		}
		if !ok {
			return domainErrors.InputError{Reason: "clase no está en el catálogo activo"}
		}
	}
	if subtipo != "" {
		var n int
		if err := tx.Raw(`
			SELECT count(*) FROM catalogos
			WHERE clase = 'subtipo_actividad' AND codigo = $1 AND activo AND padre_codigo = $2`,
			subtipo, clase).Scan(&n).Error; err != nil {
			return err
		}
		if n != 1 {
			return domainErrors.InputError{Reason: "el tipo de actividad no pertenece a esa clase"}
		}
	}
	if lugar := strings.TrimSpace(in.LugarID); lugar != "" {
		var n int
		if err := tx.Raw(`SELECT count(*) FROM lugares WHERE id::text = $1 AND activo`, lugar).Scan(&n).Error; err != nil {
			return err
		}
		if n != 1 {
			return domainErrors.InputError{Reason: "el lugar no está en el catálogo"}
		}
	}
	if zona := strings.TrimSpace(in.ZonaSupervisionID); zona != "" {
		var n int
		if err := tx.Raw(`SELECT count(*) FROM zonas_supervision WHERE codigo = $1`, zona).Scan(&n).Error; err != nil {
			return err
		}
		if n != 1 {
			return domainErrors.InputError{Reason: "la zona de supervisión no está en el catálogo"}
		}
	}
	for _, nombre := range in.Personal {
		if _, err := nombreFicticio(tx, nombre); err != nil {
			return err
		}
	}
	return nil
}

func catalogoActivoTx(tx *gorm.DB, clase, codigo string) (bool, error) {
	var n int
	err := tx.Raw(`SELECT count(*) FROM catalogos WHERE clase = $1 AND codigo = $2 AND activo`, clase, codigo).Scan(&n).Error
	return n == 1, err
}

func nombreFicticio(tx *gorm.DB, nombre string) (string, error) {
	var canonical string
	err := tx.Raw(`
		SELECT nombre_ficticio FROM personal_ficticio
		WHERE activo AND lower(nombre_ficticio) = lower($1)`, strings.TrimSpace(nombre)).Scan(&canonical).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return "", err
	}
	if strings.TrimSpace(canonical) == "" {
		return "", domainErrors.InputError{Reason: "el personal debe ser un nombre ficticio del catálogo"}
	}
	return canonical, nil
}

func insertarPersonal(tx *gorm.DB, actividadID string, nombres []string) error {
	vistos := map[string]struct{}{}
	for _, nombre := range nombres {
		canonical, err := nombreFicticio(tx, nombre)
		if err != nil {
			return err
		}
		if _, ok := vistos[canonical]; ok {
			continue
		}
		vistos[canonical] = struct{}{}
		if err := tx.Exec(`
			INSERT INTO personal_labor (id, actividad_id, nombre_ficticio, rol_campo)
			VALUES (gen_random_uuid(), $1, $2, 'personal de cuadrilla')
			ON CONFLICT (actividad_id, nombre_ficticio) DO NOTHING`,
			actividadID, canonical).Error; err != nil {
			return err
		}
	}
	return nil
}

func (r *intervencionRepository) Taxonomia(ctx context.Context) (entities.TaxonomiaActividad, error) {
	var out entities.TaxonomiaActividad
	var clases []struct {
		Codigo string
		Nombre string
	}
	if err := r.db.WithContext(ctx).Raw(`
		SELECT codigo, nombre FROM catalogos
		WHERE clase = 'clase_actividad' AND activo
		ORDER BY orden, codigo`).Scan(&clases).Error; err != nil {
		return out, err
	}
	var tipos []struct {
		Codigo string
		Nombre string
		Padre  string
	}
	if err := r.db.WithContext(ctx).Raw(`
		SELECT codigo, nombre, COALESCE(padre_codigo, '') AS padre
		FROM catalogos
		WHERE clase = 'subtipo_actividad' AND activo
		ORDER BY orden, nombre`).Scan(&tipos).Error; err != nil {
		return out, err
	}
	out.Clases = make([]entities.ClaseActividad, 0, len(clases))
	for _, clase := range clases {
		item := entities.ClaseActividad{Codigo: clase.Codigo, Nombre: clase.Nombre, Tipos: []entities.OpcionCatalogo{}}
		for _, tipo := range tipos {
			if tipo.Padre == clase.Codigo {
				item.Tipos = append(item.Tipos, entities.OpcionCatalogo{Codigo: tipo.Codigo, Nombre: tipo.Nombre, Padre: tipo.Padre})
			}
		}
		out.Clases = append(out.Clases, item)
	}
	riesgos, err := opcionesCatalogo(r.db.WithContext(ctx), "nivel_riesgo")
	if err != nil {
		return out, err
	}
	origenes, err := opcionesCatalogo(r.db.WithContext(ctx), "fuente")
	if err != nil {
		return out, err
	}
	out.Riesgos = riesgos
	out.Origenes = origenes
	if err := r.db.WithContext(ctx).Raw(`
		SELECT id, nombre_ficticio FROM personal_ficticio WHERE activo ORDER BY nombre_ficticio`).Scan(&out.Personal).Error; err != nil {
		return out, err
	}
	if out.Personal == nil {
		out.Personal = []entities.PersonalFicticio{}
	}
	return out, nil
}

func opcionesCatalogo(tx *gorm.DB, clase string) ([]entities.OpcionCatalogo, error) {
	var rows []entities.OpcionCatalogo
	err := tx.Raw(`
		SELECT codigo, nombre FROM catalogos
		WHERE clase = $1 AND activo
		ORDER BY orden, codigo`, clase).Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	if rows == nil {
		rows = []entities.OpcionCatalogo{}
	}
	return rows, nil
}

func (r *intervencionRepository) ListarPersonal(ctx context.Context, actividadID string) ([]entities.PersonalLabor, error) {
	if err := actividadExiste(r.db.WithContext(ctx), actividadID); err != nil {
		return nil, err
	}
	var rows []entities.PersonalLabor
	err := r.db.WithContext(ctx).Raw(`
		SELECT id::text, actividad_id::text, nombre_ficticio, rol_campo
		FROM personal_labor WHERE actividad_id = $1
		ORDER BY nombre_ficticio`, actividadID).Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	if rows == nil {
		rows = []entities.PersonalLabor{}
	}
	return rows, nil
}

func (r *intervencionRepository) RegistrarPersonal(ctx context.Context, actividadID string, nombres []string) ([]entities.PersonalLabor, error) {
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := actividadExiste(tx, actividadID); err != nil {
			return err
		}
		return insertarPersonal(tx, actividadID, nombres)
	})
	if err != nil {
		return nil, err
	}
	return r.ListarPersonal(ctx, actividadID)
}

func actividadExiste(tx *gorm.DB, id string) error {
	var n int
	if err := tx.Raw(`SELECT count(*) FROM actividades WHERE id = $1`, id).Scan(&n).Error; err != nil {
		return err
	}
	if n != 1 {
		return domainErrors.ErrLaborNoEncontrada
	}
	return nil
}
