package middleware

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	domainErrors "github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/errors"
)

// Errores translates errors attached to gin.Context into standard {"error": "..."} responses.
func Errores() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
		if len(c.Errors) > 0 && !c.Writer.Written() {
			err := c.Errors.Last().Err
			var appErr *domainErrors.AppError
			if errors.As(err, &appErr) {
				c.JSON(appErr.StatusCode, gin.H{"error": appErr.Message})
				return
			}

			status := c.Writer.Status()
			if status == http.StatusOK {
				status = http.StatusInternalServerError
			}
			c.JSON(status, gin.H{"error": err.Error()})
		}
	}
}
