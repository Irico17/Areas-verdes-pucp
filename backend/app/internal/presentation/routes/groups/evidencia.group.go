// Package groups registers route handlers into Gin router groups.
package groups

import (
	"github.com/gin-gonic/gin"

	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/contracts"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/presentation/controller"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/presentation/middleware"
)

// EvidenciaGroup handles routes for evidencias.
type EvidenciaGroup struct {
	ctrl     controller.IEvidenciaController
	permisos contracts.IPermisosService
}

// NewEvidenciaGroup creates a new EvidenciaGroup instance.
func NewEvidenciaGroup(
	ctrl controller.IEvidenciaController,
	permisos contracts.IPermisosService,
) *EvidenciaGroup {
	return &EvidenciaGroup{
		ctrl:     ctrl,
		permisos: permisos,
	}
}

// Register registers evidencia routes on the router.
func (g *EvidenciaGroup) Register(router gin.IRouter) {
	evi := router.Group("/evidencias")
	evi.GET("", middleware.RequierePermiso(g.permisos, "consultar"), g.ctrl.List)
	evi.POST("", middleware.RequiereAlgunPermiso(g.permisos, "registrar", "evidencias"), g.ctrl.Upload)
	evi.GET("/:id/archivo", middleware.RequierePermiso(g.permisos, "consultar"), g.ctrl.File)
}
