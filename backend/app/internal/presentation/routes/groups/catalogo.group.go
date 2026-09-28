package groups

import (
	"github.com/gin-gonic/gin"

	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/contracts"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/presentation/controller"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/presentation/middleware"
)

// CatalogoGroup handles catalog routes.
type CatalogoGroup struct {
	controller controller.ICatalogoController
	permisos   contracts.IPermisosService
}

// NewCatalogoGroup creates a new catalog group.
func NewCatalogoGroup(controller controller.ICatalogoController, permisos contracts.IPermisosService) *CatalogoGroup {
	return &CatalogoGroup{controller: controller, permisos: permisos}
}

// Register registers catalog routes on a router.
func (g *CatalogoGroup) Register(router gin.IRouter) {
	router.GET("/catalogos", middleware.RequierePermiso(g.permisos, "consultar"), g.controller.Listar)
	router.POST("/catalogos", middleware.RequierePermiso(g.permisos, "catalogos"), g.controller.Crear)
	router.POST("/catalogos/:id/desactivar", middleware.RequierePermiso(g.permisos, "catalogos"), g.controller.Desactivar)
}
