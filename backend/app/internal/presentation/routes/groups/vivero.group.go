// Package groups registers route handlers into Gin router groups.
package groups

import (
	"github.com/gin-gonic/gin"

	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/contracts"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/presentation/controller"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/presentation/middleware"
)

// ViveroGroup handles routes for nursery activities.
type ViveroGroup struct {
	ctrl     controller.IViveroController
	permisos contracts.IPermisosService
}

// NewViveroGroup creates a new ViveroGroup instance.
func NewViveroGroup(
	ctrl controller.IViveroController,
	permisos contracts.IPermisosService,
) *ViveroGroup {
	return &ViveroGroup{
		ctrl:     ctrl,
		permisos: permisos,
	}
}

// Register registers vivero routes on the router.
func (g *ViveroGroup) Register(router gin.IRouter) {
	viv := router.Group("/vivero")
	viv.GET("", middleware.RequierePermiso(g.permisos, "consultar"), g.ctrl.List)
	viv.POST("", middleware.RequierePermiso(g.permisos, "registrar"), g.ctrl.Create)
	viv.PATCH("/:id", middleware.RequierePermiso(g.permisos, "registrar"), g.ctrl.Edit)
	viv.POST("/:id/archivar", middleware.RequierePermiso(g.permisos, "validar"), g.ctrl.Archive)
}
