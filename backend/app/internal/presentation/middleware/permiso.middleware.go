package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/contracts"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/dto"
	domainErrors "github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/errors"
)

// UsuarioEn extracts the authenticated user from gin.Context if present.
func UsuarioEn(c *gin.Context) (dto.UsuarioSesionDTO, bool) {
	v, ok := c.Get("usuario")
	if !ok {
		return dto.UsuarioSesionDTO{}, false
	}
	switch u := v.(type) {
	case dto.UsuarioSesionDTO:
		return u, true
	case *dto.UsuarioSesionDTO:
		if u != nil {
			return *u, true
		}
	}
	return dto.UsuarioSesionDTO{}, false
}

// RequierePermiso requires that the authenticated user's role has the specified permission.
func RequierePermiso(permisosSvc contracts.IPermisosService, accion string) gin.HandlerFunc {
	return func(c *gin.Context) {
		u, ok := UsuarioEn(c)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "inicie sesión"})
			return
		}
		if !permisosSvc.Permite(u.Rol, accion) {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": domainErrors.ErrSinPermiso.Error()})
			return
		}
		c.Next()
	}
}

// ExigePermiso is an alias for RequierePermiso for backward compatibility.
func ExigePermiso(permisosSvc contracts.IPermisosService, accion string) gin.HandlerFunc {
	return RequierePermiso(permisosSvc, accion)
}
