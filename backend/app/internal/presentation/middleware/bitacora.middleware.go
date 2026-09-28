// Package middleware provides HTTP middlewares for presentation layer.
package middleware

import (
	"crypto/rand"
	"encoding/hex"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
)

// Bitacora logs HTTP requests using zerolog without exposing sensitive data.
func Bitacora(logger zerolog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.GetHeader("X-Request-ID")
		if id == "" {
			id = nuevoID()
		}
		c.Header("X-Request-ID", id)
		inicio := time.Now()
		c.Next()
		logger.Info().
			Str("request_id", id).
			Str("method", c.Request.Method).
			Str("path", c.Request.URL.Path).
			Int("status", c.Writer.Status()).
			Int64("latency_ms", time.Since(inicio).Milliseconds()).
			Int("bytes", c.Writer.Size()).
			Msg("http_request")
	}
}

func nuevoID() string {
	var b [8]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "sin-id"
	}
	return hex.EncodeToString(b[:])
}
