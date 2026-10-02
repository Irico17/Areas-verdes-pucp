package dto

import "encoding/json"

// RegistrarCambioDTO holds parameters to record a change in the audit trail.
type RegistrarCambioDTO struct {
	Entidad   string
	EntidadID string
	Accion    string
	Antes     any
	Despues   any
	UsuarioID *int64
	LoteID    *int64
}

// FilaLoteDTO represents a row within an import batch request.
type FilaLoteDTO struct {
	EntidadID string          `json:"entidad_id"`
	Accion    string          `json:"accion"`
	Antes     json.RawMessage `json:"antes"`
	Despues   json.RawMessage `json:"despues"`
}

// ImportarLoteDTO holds parameters for importing a batch of changes.
type ImportarLoteDTO struct {
	Entidad string        `json:"entidad"`
	Filas   []FilaLoteDTO `json:"filas"`
}

// ImportarLoteResponseDTO is the response after successfully importing a batch.
type ImportarLoteResponseDTO struct {
	LoteID int64 `json:"lote_id"`
	Filas  int   `json:"filas"`
}

// RevertirLoteDTO holds parameters for reverting an import batch.
type RevertirLoteDTO struct {
	Confirmar bool `json:"confirmar"`
}

// ExcluidaDTO represents an entity excluded from reversion.
type ExcluidaDTO struct {
	EntidadID string `json:"entidad_id"`
	Motivo    string `json:"motivo"`
}

// ReporteReversionDTO is the response returned when reverting a batch.
type ReporteReversionDTO struct {
	LoteID     int64         `json:"lote_id"`
	Revertidas []string      `json:"revertidas"`
	Excluidas  []ExcluidaDTO `json:"excluidas"`
}

// EditarAuditoriaDTO holds parameters for manually editing an entity with audit trail.
type EditarAuditoriaDTO struct {
	Entidad   string         `json:"entidad"`
	EntidadID string         `json:"entidad_id"`
	Despues   map[string]any `json:"despues"`
}

// EditarAuditoriaResponseDTO is the response after a manual edit.
type EditarAuditoriaResponseDTO struct {
	Editada   bool   `json:"editada"`
	EntidadID string `json:"entidad_id"`
}

// FiltroAuditoriaDTO holds parameters for filtering audit logs.
type FiltroAuditoriaDTO struct {
	Entidad   string
	EntidadID string
	Zona      string
	Origen    string
	Desde     string
	Hasta     string
}

// EventoAuditoriaDTO represents an audit event entry.
type EventoAuditoriaDTO struct {
	ID        int64  `json:"id"`
	Entidad   string `json:"entidad"`
	EntidadID string `json:"entidad_id"`
	Accion    string `json:"accion"`
	UsuarioID int64  `json:"usuario_id"`
	Usuario   string `json:"usuario"`
	Nombre    string `json:"nombre"`
	LoteID    *int64 `json:"lote_id,omitempty"`
	CreatedAt string `json:"created_at"`
	Antes     string `json:"antes,omitempty"`
	Despues   string `json:"despues,omitempty"`
}

// HistorialAuditoriaResponseDTO represents the list of audit events.
type HistorialAuditoriaResponseDTO struct {
	Eventos []EventoAuditoriaDTO `json:"eventos"`
}
