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
	if g.permisos != nil {
		router.GET("/catalogos", middleware.ExigePermiso(g.permisos, "consultar"), g.controller.Listar)
		router.POST("/catalogos", middleware.ExigePermiso(g.permisos, "catalogos"), g.controller.Crear)
		router.POST("/catalogos/:id/desactivar", middleware.ExigePermiso(g.permisos, "catalogos"), g.controller.Desactivar)
	} else {
		router.GET("/catalogos", g.controller.Listar)
		router.POST("/catalogos", g.controller.Crear)
		router.POST("/catalogos/:id/desactivar", g.controller.Desactivar)
	}
}
