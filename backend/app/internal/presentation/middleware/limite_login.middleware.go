package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/contracts"
)

// LimiteLogin limits POST login requests per client IP.
func LimiteLogin(lim contracts.ILimitador) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Method == http.MethodPost && isLoginPath(c.Request.URL.Path) {
			if !lim.Permitir(c.ClientIP()) {
				c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": "demasiados intentos de ingreso"})
				return
			}
		}
		c.Next()
	}
}

func isLoginPath(path string) bool {
	return path == "/api/v1/sesion" || path == "/areas-verdes/v1/sesion"
}
