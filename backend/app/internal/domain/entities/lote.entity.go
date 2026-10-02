package entities

import (
	"encoding/json"
	"time"
)

// FilaLote represents an input row in an import batch.
type FilaLote struct {
	EntidadID string
	Accion    string
	Antes     json.RawMessage
	Despues   json.RawMessage
}

// Excluida represents an entity excluded from reversion due to subsequent edits.
type Excluida struct {
	EntidadID string
	Motivo    string
}

// ReporteReversion represents the summary result of reverting a batch.
type ReporteReversion struct {
	LoteID     int64
	Revertidas []string
	Excluidas  []Excluida
}

// EventoAuditoria represents an audit log entry in the timeline or history.
type EventoAuditoria struct {
	ID        int64
	Entidad   string
	EntidadID string
	Accion    string
	UsuarioID int64
	Usuario   string
	Nombre    string
	LoteID    *int64
	CreatedAt time.Time
}

// SnapAuditoria holds fields captured or updated in snapshots.
type SnapAuditoria struct {
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

// FiltroAuditoria defines filtering criteria for audit events.
type FiltroAuditoria struct {
	Entidad   string
	EntidadID string
	Zona      string
	Origen    string
	Desde     string
	Hasta     string
}
