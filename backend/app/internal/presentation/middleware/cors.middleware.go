package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// CORS reflects allowed origins and handles preflight OPTIONS requests without using wildcards.
func CORS(origenes []string) gin.HandlerFunc {
	permitidos := map[string]struct{}{}
	for _, o := range origenes {
		if o == "" || o == "*" {
			continue
		}
		permitidos[o] = struct{}{}
	}
	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		if _, ok := permitidos[origin]; ok {
			c.Header("Access-Control-Allow-Origin", origin)
			c.Header("Vary", "Origin")
			c.Header("Access-Control-Allow-Credentials", "true")
		}
		c.Header("Access-Control-Allow-Methods", "GET, POST, PATCH, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type")
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}
