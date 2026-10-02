// Package mapper provides converters between persistence models and domain entities.
package mapper

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/entities"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/persistence/models"
)

// Scanner abstracts row scanning for database results.
type Scanner interface {
	Scan(dest ...any) error
}

// CapatazModelToEntity converts a CapatazModel to domain entity.
func CapatazModelToEntity(m models.CapatazModel) entities.Capataz {
	return entities.Capataz{
		ID:     m.ID,
		Equipo: m.Equipo,
		Turno:  m.Turno,
		Activo: m.Activo,
	}
}

// ScanFeature scans an activity row including geometry into an entities.Feature.
func ScanFeature(rows Scanner) (entities.Feature, error) {
	var (
		id, tipo, estado, titulo, detalle string
		area, zona, capataz, equipo       sql.NullString
		archivada                         bool
		created, updated                  time.Time
		ejecutor                          string
		geom                              sql.NullString
		origen                            string
		codigo, unidad, riesgo, fecha     sql.NullString
		cantidad                          sql.NullFloat64
		subtipo, clase, personalJSON      sql.NullString
	)
	if err := rows.Scan(
		&id, &tipo, &estado, &titulo, &detalle, &area, &zona, &capataz, &equipo,
		&archivada, &created, &updated, &ejecutor, &geom,
		&origen, &codigo, &unidad, &riesgo, &fecha, &cantidad, &subtipo, &clase, &personalJSON,
	); err != nil {
		return entities.Feature{}, err
	}
	personal, err := personalDeJSON(personalJSON)
	if err != nil {
		return entities.Feature{}, err
	}
	props := entities.ActividadProperties{
		ID:                id,
		Tipo:              tipo,
		Estado:            estado,
		Titulo:            titulo,
		Detalle:           detalle,
		AreaFeatureID:     NullString(area),
		ZonaFeatureID:     NullString(zona),
		AssignedCapatazID: NullString(capataz),
		Equipo:            NullString(equipo),
		Ejecutor:          ejecutor,
		Origen:            origen,
		CodigoExterno:     NullString(codigo),
		UnidadSolicitante: NullString(unidad),
		NivelRiesgo:       NullString(riesgo),
		FechaProgramada:   NullString(fecha),
		Cantidad:          nullFloat(cantidad),
		Subtipo:           NullString(subtipo),
		Clase:             NullString(clase),
		Personal:          personal,
		Archivada:         archivada,
		CreatedAt:         created.UTC().Format(time.RFC3339),
		UpdatedAt:         updated.UTC().Format(time.RFC3339),
	}
	raw := json.RawMessage("null")
	if geom.Valid && strings.TrimSpace(geom.String) != "" {
		raw = json.RawMessage(geom.String)
	}
	if !json.Valid(raw) {
		return entities.Feature{}, fmt.Errorf("geometría inválida para %s", id)
	}
	return entities.Feature{
		Type:       "Feature",
		ID:         id,
		Geometry:   raw,
		Properties: props,
	}, nil
}

func nullFloat(v sql.NullFloat64) *float64 {
	if !v.Valid {
		return nil
	}
	n := v.Float64
	return &n
}

func personalDeJSON(v sql.NullString) ([]string, error) {
	if !v.Valid || strings.TrimSpace(v.String) == "" || v.String == "[]" || v.String == "null" {
		return nil, nil
	}
	var names []string
	if err := json.Unmarshal([]byte(v.String), &names); err != nil {
		return nil, err
	}
	if len(names) == 0 {
		return nil, nil
	}
	return names, nil
}

// NullString converts sql.NullString to *string.
func NullString(v sql.NullString) *string {
	if !v.Valid || strings.TrimSpace(v.String) == "" {
		return nil
	}
	s := v.String
	return &s
}
