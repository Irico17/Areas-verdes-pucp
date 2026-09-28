package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/contracts"
)

// Auth checks the session cookie cv_sesion and sets the user in gin.Context.
func Auth(sesionUC contracts.ISesionUseCase) gin.HandlerFunc {
	return func(c *gin.Context) {
		if sesionUC != nil {
			cookie, err := c.Request.Cookie(CookieSesion)
			if err == nil && strings.TrimSpace(cookie.Value) != "" {
				if u, err := sesionUC.Resolver(c.Request.Context(), cookie.Value); err == nil && u != nil {
					c.Set("usuario", *u)
				}
			}
		}
		c.Next()
	}
}
