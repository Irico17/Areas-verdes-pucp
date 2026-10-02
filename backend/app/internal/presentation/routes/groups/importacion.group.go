// Package groups defines HTTP route registration groups.
package groups

import (
	"github.com/gin-gonic/gin"

	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/contracts"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/presentation/controller"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/presentation/middleware"
)

// ImportacionGroup manages route registration for file import endpoints.
type ImportacionGroup struct {
	ctrl     controller.IImportacionController
	permisos contracts.IPermisosService
}

// NewImportacionGroup creates a new ImportacionGroup instance.
func NewImportacionGroup(
	ctrl controller.IImportacionController,
	permisos contracts.IPermisosService,
) *ImportacionGroup {
	return &ImportacionGroup{
		ctrl:     ctrl,
		permisos: permisos,
	}
}

// Register registers import routes into the given router.
func (g *ImportacionGroup) Register(router gin.IRouter) {
	router.GET("/importaciones/entidades", middleware.RequierePermiso(g.permisos, "consultar"), g.ctrl.Entidades)
	router.POST("/importaciones", middleware.RequierePermiso(g.permisos, "validar"), g.ctrl.Previsualizar)
	router.POST("/importaciones/:id/confirmar", middleware.RequierePermiso(g.permisos, "validar"), g.ctrl.Confirmar)
}
