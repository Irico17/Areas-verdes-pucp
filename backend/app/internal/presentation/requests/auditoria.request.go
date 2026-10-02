// Package requests defines incoming HTTP payload schemas.
package requests

import "encoding/json"

// FilaLoteRequest represents an entity change within a batch import payload.
type FilaLoteRequest struct {
	EntidadID string          `json:"entidad_id"`
	Accion    string          `json:"accion"`
	Antes     json.RawMessage `json:"antes"`
	Despues   json.RawMessage `json:"despues"`
}

// LoteRequest represents the payload for importing a batch of changes.
type LoteRequest struct {
	Entidad string            `json:"entidad"`
	Filas   []FilaLoteRequest `json:"filas"`
}

// RevertirLoteRequest represents the payload for reverting a batch.
type RevertirLoteRequest struct {
	Confirmar bool `json:"confirmar"`
}

// EditarAuditoriaRequest represents the payload for manually auditing an edit.
type EditarAuditoriaRequest struct {
	Entidad   string         `json:"entidad"`
	EntidadID string         `json:"entidad_id"`
	Despues   map[string]any `json:"despues"`
}
