// Package groups registers route handlers into Gin router groups.
package groups

import (
	"github.com/gin-gonic/gin"

	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/contracts"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/presentation/controller"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/presentation/middleware"
)

// InventarioGroup handles legacy inventory overlay and photo routes.
type InventarioGroup struct {
	controller controller.IInventarioController
	permisos   contracts.IPermisosService
}

// NewInventarioGroup creates a new InventarioGroup.
func NewInventarioGroup(controller controller.IInventarioController, permisos contracts.IPermisosService) *InventarioGroup {
	return &InventarioGroup{
		controller: controller,
		permisos:   permisos,
	}
}

// Register registers inventory routes on the router.
// Note: /geo/inventario/fotos/:name MUST be registered before /geo/inventario/:capa
// to ensure Gin routing precedence matches the legacy API behavior.
func (g *InventarioGroup) Register(router gin.IRouter) {
	router.GET("/geo/inventario", middleware.RequierePermiso(g.permisos, "consultar"), g.controller.Index)
	router.GET("/geo/inventario/fotos/:name", middleware.RequierePermiso(g.permisos, "consultar"), g.controller.Foto)
	router.GET("/geo/inventario/:capa", middleware.RequierePermiso(g.permisos, "consultar"), g.controller.Capa)
}
