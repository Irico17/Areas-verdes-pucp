// Package groups defines HTTP route registration groups.
package groups

import (
	"github.com/gin-gonic/gin"

	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/contracts"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/presentation/controller"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/presentation/middleware"
)

// AuditoriaGroup manages route registration for audit and batch operations.
type AuditoriaGroup struct {
	ctrl     controller.IAuditoriaController
	permisos contracts.IPermisosService
}

// NewAuditoriaGroup creates a new AuditoriaGroup instance.
func NewAuditoriaGroup(
	ctrl controller.IAuditoriaController,
	permisos contracts.IPermisosService,
) *AuditoriaGroup {
	return &AuditoriaGroup{
		ctrl:     ctrl,
		permisos: permisos,
	}
}

// Register registers audit and lotes routes into the given router.
func (g *AuditoriaGroup) Register(router gin.IRouter) {
	router.POST("/lotes", middleware.RequierePermiso(g.permisos, "validar"), g.ctrl.Importar)
	router.POST("/lotes/:id/revertir", middleware.RequierePermiso(g.permisos, "validar"), g.ctrl.Revertir)
	router.POST("/auditoria/ediciones", middleware.RequierePermiso(g.permisos, "validar"), g.ctrl.Editar)
	router.GET("/auditoria/cambios", middleware.RequierePermiso(g.permisos, "consultar"), g.ctrl.Historial)
	router.GET("/auditoria/timeline", middleware.RequierePermiso(g.permisos, "consultar"), g.ctrl.Timeline)
}
