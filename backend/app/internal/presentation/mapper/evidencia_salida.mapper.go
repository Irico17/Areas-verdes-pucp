// Package mapper builds HTTP payloads without touching persistence.
package mapper

import "strings"

// RespuestaSubida es el cuerpo de POST /evidencias.
// Almacen vale "s3" solo cuando este proceso tiene cubo configurado.
// Si no, vale "disco". No es una URL ni una columna nueva.
type RespuestaSubida struct {
	ID          string `json:"id"`
	Idempotente bool   `json:"idempotente"`
	Almacen     string `json:"almacen"`
}

// SalidaSubida arma la respuesta de la subida. No escribe el INSERT.
func SalidaSubida(id string, idempotente bool, bucket string) RespuestaSubida {
	return RespuestaSubida{
		ID:          id,
		Idempotente: idempotente,
		Almacen:     ModoAlmacen(bucket),
	}
}

// ModoAlmacen informa el almacén activo. "s3" solo si hay cubo.
func ModoAlmacen(bucket string) string {
	if strings.TrimSpace(bucket) == "" {
		return "disco"
	}
	return "s3"
}
