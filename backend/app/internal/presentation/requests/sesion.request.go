// Package requests defines incoming HTTP request body structures.
package requests

// LoginRequest contains credentials submitted to POST /sesion.
type LoginRequest struct {
	Usuario string `json:"usuario"`
	Clave   string `json:"clave"`
}
