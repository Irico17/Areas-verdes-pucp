// Package groups registers route handlers into Gin router groups.
package groups

import (
	"github.com/gin-gonic/gin"

	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/contracts"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/presentation/controller"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/presentation/middleware"
)

// GeoGroup handles geo read endpoints.
type GeoGroup struct {
	controller controller.IGeoController
	permisos   contracts.IPermisosService
}

// NewGeoGroup creates a new GeoGroup.
func NewGeoGroup(controller controller.IGeoController, permisos contracts.IPermisosService) *GeoGroup {
	return &GeoGroup{controller: controller, permisos: permisos}
}

// Register registers geo routes on the router.
func (g *GeoGroup) Register(router gin.IRouter) {
	router.GET("/geo/resumen", middleware.RequierePermiso(g.permisos, "consultar"), g.controller.Resumen)
	router.GET("/geo/areas", middleware.RequierePermiso(g.permisos, "consultar"), g.controller.Areas)
	router.GET("/geo/zonas", middleware.RequierePermiso(g.permisos, "consultar"), g.controller.Zonas)
	router.GET("/geo/capas", middleware.RequierePermiso(g.permisos, "consultar"), g.controller.Capas)
	router.GET("/geo/capas/:capa", middleware.RequierePermiso(g.permisos, "consultar"), g.controller.Capa)
	router.GET("/geo/edificios", middleware.RequierePermiso(g.permisos, "consultar"), g.controller.Edificios)
}
