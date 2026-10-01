// Package groups registers route handlers into Gin router groups.
package groups

import (
	"github.com/gin-gonic/gin"

	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/contracts"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/presentation/controller"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/presentation/middleware"
)

// IAGroup registers routes for /ia.
type IAGroup struct {
	ctrl     controller.IIAController
	permisos contracts.IPermisosService
}

// NewIAGroup creates a new IAGroup instance.
func NewIAGroup(
	ctrl controller.IIAController,
	permisos contracts.IPermisosService,
) *IAGroup {
	return &IAGroup{
		ctrl:     ctrl,
		permisos: permisos,
	}
}

// Register registers IA routes on the router.
func (g *IAGroup) Register(router gin.IRouter) {
	ia := router.Group("/ia")
	ia.POST("/sugerir-tipo", middleware.RequierePermiso(g.permisos, "consultar"), g.ctrl.Sugerir)
}
