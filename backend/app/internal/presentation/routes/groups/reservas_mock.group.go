// Package groups registers route handlers into Gin router groups.
package groups

import (
	"github.com/gin-gonic/gin"

	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/contracts"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/presentation/controller"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/presentation/middleware"
)

// ReservasMockGroup handles mock reservations agenda routes (/geo/reservas-mock).
type ReservasMockGroup struct {
	controller controller.IReservasMockController
	permisos   contracts.IPermisosService
}

// NewReservasMockGroup creates a new ReservasMockGroup.
func NewReservasMockGroup(controller controller.IReservasMockController, permisos contracts.IPermisosService) *ReservasMockGroup {
	return &ReservasMockGroup{
		controller: controller,
		permisos:   permisos,
	}
}

// Register registers mock reservations routes on the router.
func (g *ReservasMockGroup) Register(router gin.IRouter) {
	router.GET("/geo/reservas-mock", middleware.RequierePermiso(g.permisos, "consultar"), g.controller.Get)
}
