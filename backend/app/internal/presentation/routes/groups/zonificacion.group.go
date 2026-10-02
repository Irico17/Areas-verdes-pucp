package groups

import (
	"github.com/gin-gonic/gin"

	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/contracts"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/presentation/controller"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/presentation/middleware"
)

// ZonificacionGroup registers capataz sectors, places, roads and historical quarters.
type ZonificacionGroup struct {
	controller controller.IZonificacionController
	permisos   contracts.IPermisosService
}

// NewZonificacionGroup creates the zoning route group.
func NewZonificacionGroup(controller controller.IZonificacionController, permisos contracts.IPermisosService) *ZonificacionGroup {
	return &ZonificacionGroup{controller: controller, permisos: permisos}
}

// Register mounts the zoning routes. There is no DELETE: a sector is deactivated.
func (g *ZonificacionGroup) Register(router gin.IRouter) {
	consultar := middleware.RequierePermiso(g.permisos, "consultar")
	editar := middleware.RequierePermiso(g.permisos, "catalogos")

	router.GET("/zonificacion/sectores", consultar, g.controller.ListarSectores)
	router.POST("/zonificacion/sectores/importar", editar, g.controller.ImportarSectores)
	router.POST("/zonificacion/sectores", editar, g.controller.CrearSector)
	router.PATCH("/zonificacion/sectores/:codigo", editar, g.controller.ActualizarSector)
	router.POST("/zonificacion/sectores/:codigo/desactivar", editar, g.controller.DesactivarSector)

	router.GET("/zonificacion/lugares", consultar, g.controller.ListarLugares)
	router.POST("/zonificacion/lugares/resolver", consultar, g.controller.ResolverLugar)

	router.GET("/zonificacion/vias", consultar, g.controller.Vias)
	router.POST("/zonificacion/vias/importar", editar, g.controller.ImportarVias)

	router.GET("/zonificacion/cuarteles", consultar, g.controller.Cuarteles)

	router.GET("/zonificacion/edificios", consultar, g.controller.Edificios)
	router.POST("/zonificacion/referentes", editar, g.controller.CrearReferente)
}
