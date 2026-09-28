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

// ActividadModelToEntity converts an ActividadModel to domain entity.
func ActividadModelToEntity(m models.ActividadModel) entities.Intervencion {
	return entities.Intervencion{
		ID:                m.ID,
		Tipo:              m.Tipo,
		Estado:            m.Estado,
		Titulo:            m.Titulo,
		Detalle:           m.Detalle,
		AreaFeatureID:     m.AreaFeatureID,
		ZonaFeatureID:     m.ZonaFeatureID,
		AssignedCapatazID: m.AssignedCapatazID,
		ArchivadaEn:       m.ArchivadaEn,
		CreatedAt:         m.CreatedAt,
		UpdatedAt:         m.UpdatedAt,
		Ejecutor:          m.Ejecutor,
		MotivoArchivo:     m.MotivoArchivo,
		LugarID:           m.LugarID,
		ZonaSupervisionID: m.ZonaSupervisionID,
		FechaSolicitud:    m.FechaSolicitud,
		FechaAtencion:     m.FechaAtencion,
		CuadrillaID:       m.CuadrillaID,
		ClaseCodigo:       m.ClaseCodigo,
		TipoCodigo:        m.TipoCodigo,
		Comentario:        m.Comentario,
		LugarLibre:        m.LugarLibre,
		OrigenRef:         m.OrigenRef,
		Origen:            m.Origen,
	}
}

// ActividadEventoModelToEntity converts an ActividadEventoModel to domain entity.
func ActividadEventoModelToEntity(m models.ActividadEventoModel) entities.ActividadEvento {
	return entities.ActividadEvento{
		ID:          m.ID,
		ActividadID: m.ActividadID,
		Tipo:        m.Tipo,
		Estado:      m.Estado,
		CapatazID:   m.CapatazID,
		ActorRol:    m.ActorRol,
		Nota:        m.Nota,
		CreatedAt:   m.CreatedAt,
		UsuarioID:   m.UsuarioID,
	}
}

// ActividadAvanceModelToEntity converts an ActividadAvanceModel to domain entity.
func ActividadAvanceModelToEntity(m models.ActividadAvanceModel) entities.Avance {
	return entities.Avance{
		ID:            m.ID,
		ActividadID:   m.ActividadID,
		Fecha:         m.Fecha,
		Nota:          m.Nota,
		AreaFeatureID: m.AreaFeatureID,
		EjemplarRef:   m.EjemplarRef,
		CreatedAt:     m.CreatedAt,
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
	)
	if err := rows.Scan(&id, &tipo, &estado, &titulo, &detalle, &area, &zona, &capataz, &equipo, &archivada, &created, &updated, &ejecutor, &geom); err != nil {
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

// NullString converts sql.NullString to *string.
func NullString(v sql.NullString) *string {
	if !v.Valid || strings.TrimSpace(v.String) == "" {
		return nil
	}
	s := v.String
	return &s
}
