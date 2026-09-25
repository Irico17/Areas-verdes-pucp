package server

import (
	"crypto/rand"
	"encoding/hex"
	"log"
	"time"

	"github.com/gin-gonic/gin"
)

// Bitacora anota método, ruta, estado y un id de petición.
// No escribe el cuerpo, la query ni cabeceras de autorización o cookie.
func Bitacora() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.GetHeader("X-Request-ID")
		if id == "" {
			id = nuevoID()
		}
		c.Header("X-Request-ID", id)
		inicio := time.Now()
		c.Next()
		log.Printf("request_id=%s method=%s path=%s status=%d latency_ms=%d bytes=%d",
			id, c.Request.Method, c.Request.URL.Path, c.Writer.Status(),
			time.Since(inicio).Milliseconds(), c.Writer.Size())
	}
}

func nuevoID() string {
	var b [8]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "sin-id"
	}
	return hex.EncodeToString(b[:])
}
