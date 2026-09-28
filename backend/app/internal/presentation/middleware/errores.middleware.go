package middleware

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"

	domainErrors "github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/errors"
)

// Errores translates errors attached to gin.Context into standard {"error": "..."} responses.
// Non-AppError errors are logged with the injected logger and returned as 500 {"error":"error interno"}.
func Errores(logger zerolog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
		if len(c.Errors) > 0 && !c.Writer.Written() {
			err := c.Errors.Last().Err
			var appErr *domainErrors.AppError
			if errors.As(err, &appErr) {
				c.JSON(appErr.StatusCode, gin.H{"error": appErr.Message})
				return
			}

			logger.Error().Err(err).Str("path", c.Request.URL.Path).Msg("error interno en handler")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "error interno"})
		}
	}
}
