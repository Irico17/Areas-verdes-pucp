// Package requests defines request payload structures for HTTP endpoints.
package requests

// SugerirTipoRequest represents the JSON request body for /ia/sugerir-tipo.
type SugerirTipoRequest struct {
	Titulo string `json:"titulo"`
}
