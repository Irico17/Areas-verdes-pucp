// Package groups registers route handlers into Gin router groups.
package groups

import (
	"github.com/gin-gonic/gin"

	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/contracts"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/presentation/controller"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/presentation/middleware"
)

// PodaGroup handles routes for podas.
type PodaGroup struct {
	ctrl     controller.IPodaController
	permisos contracts.IPermisosService
}

// NewPodaGroup creates a new PodaGroup instance.
func NewPodaGroup(
	ctrl controller.IPodaController,
	permisos contracts.IPermisosService,
) *PodaGroup {
	return &PodaGroup{
		ctrl:     ctrl,
		permisos: permisos,
	}
}

// Register registers poda routes on the router.
func (g *PodaGroup) Register(router gin.IRouter) {
	pod := router.Group("/podas")
	pod.GET("", middleware.RequierePermiso(g.permisos, "consultar"), g.ctrl.List)
	pod.POST("", middleware.RequierePermiso(g.permisos, "registrar"), g.ctrl.Create)
	pod.PATCH("/:id", middleware.RequierePermiso(g.permisos, "registrar"), g.ctrl.Edit)
	pod.POST("/:id/archivar", middleware.RequierePermiso(g.permisos, "validar"), g.ctrl.Archive)
}
