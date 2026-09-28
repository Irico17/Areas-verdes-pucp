package groups

import (
	"github.com/gin-gonic/gin"

	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/contracts"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/presentation/controller"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/presentation/middleware"
)

// CatastroGroup handles cadastral management routes (/catastro/*).
type CatastroGroup struct {
	areaVerdeCtrl controller.IAreaVerdeController
	permisos      contracts.IPermisosService
}

// NewCatastroGroup creates a new CatastroGroup.
func NewCatastroGroup(areaVerdeCtrl controller.IAreaVerdeController, permisos contracts.IPermisosService) *CatastroGroup {
	return &CatastroGroup{
		areaVerdeCtrl: areaVerdeCtrl,
		permisos:      permisos,
	}
}

// Register registers catastro routes on the router.
func (g *CatastroGroup) Register(router gin.IRouter) {
	router.GET("/catastro/areas", middleware.RequierePermiso(g.permisos, "consultar"), g.areaVerdeCtrl.Listar)
	router.POST("/catastro/areas", middleware.RequierePermiso(g.permisos, "registrar"), g.areaVerdeCtrl.Crear)
	router.PATCH("/catastro/areas/:id", middleware.RequierePermiso(g.permisos, "registrar"), g.areaVerdeCtrl.Actualizar)
}
