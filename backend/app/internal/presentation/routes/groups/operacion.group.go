// Package groups registers route handlers into Gin router groups.
package groups

import (
	"github.com/gin-gonic/gin"

	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/contracts"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/presentation/controller"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/presentation/middleware"
)

// OperacionGroup handles operational activity routes (/operacion/*).
type OperacionGroup struct {
	controller controller.IIntervencionController
	permisos   contracts.IPermisosService
}

// NewOperacionGroup creates a new OperacionGroup instance.
func NewOperacionGroup(
	controller controller.IIntervencionController,
	permisos contracts.IPermisosService,
) *OperacionGroup {
	return &OperacionGroup{
		controller: controller,
		permisos:   permisos,
	}
}

// Register registers operational activity routes on the router.
func (g *OperacionGroup) Register(router gin.IRouter) {
	op := router.Group("/operacion")
	op.GET("/capataces", middleware.RequierePermiso(g.permisos, "consultar"), g.controller.Capataces)
	op.GET("/actividades", middleware.RequierePermiso(g.permisos, "consultar"), g.controller.List)
	op.POST("/actividades", middleware.RequierePermiso(g.permisos, "validar"), g.controller.Create)
	op.PATCH("/actividades/:id/asignacion", middleware.RequierePermiso(g.permisos, "validar"), g.controller.Assign)
	op.PATCH("/actividades/:id/estado", middleware.RequiereAlgunPermiso(g.permisos, "registrar", "validar"), g.controller.Estado)
	op.POST("/actividades/:id/archivar", middleware.RequierePermiso(g.permisos, "validar"), g.controller.Archive)
	op.GET("/actividades/:id/timeline", middleware.RequierePermiso(g.permisos, "consultar"), g.controller.Timeline)
	op.PATCH("/actividades/:id/ficha", middleware.RequierePermiso(g.permisos, "registrar"), g.controller.Ficha)
	op.POST("/actividades/:id/avances", middleware.RequierePermiso(g.permisos, "registrar"), g.controller.CrearAvance)
}
