package middleware

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"

	domainErrors "github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/errors"
)

// Errores translates errors attached to gin.Context into standard {"error": "..."} responses.
// It handles both domain AppError and team ApplicationError.
// Non-AppError errors are logged with the injected logger and returned as 500 {"error":"error interno"}.
func Errores(logger zerolog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
		if len(c.Errors) > 0 && !c.Writer.Written() {
			err := c.Errors.Last().Err

			// 1. Check for domain AppError
			var appErr *domainErrors.AppError
			if errors.As(err, &appErr) {
				c.JSON(appErr.StatusCode, gin.H{"error": appErr.Message})
				return
			}

			// 2. Check for team ApplicationError
			var applicationErr *domainErrors.ApplicationError
			if errors.As(err, &applicationErr) {
				statusCode, message := mapApplicationError(applicationErr)
				if statusCode >= 500 {
					logger.Error().Err(applicationErr).
						Str("code", applicationErr.Code).
						Str("path", c.Request.URL.Path).
						Msg("error de aplicacion en handler")
				}
				c.JSON(statusCode, gin.H{"error": message})
				return
			}

			// 3. Generic unknown error
			logger.Error().Err(err).Str("path", c.Request.URL.Path).Msg("error interno en handler")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "error interno"})
		}
	}
}

func mapApplicationError(appErr *domainErrors.ApplicationError) (int, string) {
	status := appErr.StatusCode
	msg := appErr.Message

	switch {
	case strings.HasPrefix(appErr.Code, "DB-7000") || appErr.Code == domainErrors.DBDatabaseConnection:
		status = http.StatusServiceUnavailable
		if msg == "" {
			msg = "base de datos no disponible"
		}
	case strings.HasPrefix(appErr.Code, "DB-7001") || appErr.Code == domainErrors.DBDatabaseQuery:
		if status == 0 {
			status = http.StatusInternalServerError
		}
		if msg == "" {
			msg = "error de base de datos"
		}
	case strings.HasPrefix(appErr.Code, "DB-7002") || appErr.Code == domainErrors.DBDatabaseError:
		if status == 0 {
			status = http.StatusInternalServerError
		}
		if msg == "" {
			msg = "error de base de datos"
		}
	case strings.HasPrefix(appErr.Code, "SRV-1000") || appErr.Code == domainErrors.SrvInternalServer:
		if status == 0 {
			status = http.StatusInternalServerError
		}
		if msg == "" {
			msg = "error interno"
		}
	default:
		if status == 0 {
			status = http.StatusInternalServerError
		}
		if msg == "" {
			msg = "error interno"
		}
	}

	return status, msg
}
