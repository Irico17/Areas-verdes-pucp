// Package groups registers route handlers into Gin router groups.
package groups

import (
	"github.com/gin-gonic/gin"

	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/contracts"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/presentation/controller"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/presentation/middleware"
)

// ReporteGroup registers routes for /reportes.
type ReporteGroup struct {
	ctrl     controller.IReporteController
	permisos contracts.IPermisosService
}

// NewReporteGroup creates a new ReporteGroup instance.
func NewReporteGroup(
	ctrl controller.IReporteController,
	permisos contracts.IPermisosService,
) *ReporteGroup {
	return &ReporteGroup{
		ctrl:     ctrl,
		permisos: permisos,
	}
}

// Register registers report routes on the router.
func (g *ReporteGroup) Register(router gin.IRouter) {
	rep := router.Group("/reportes")
	rep.GET("/labores", middleware.RequierePermiso(g.permisos, "reportes"), g.ctrl.ReporteLabores)
}
