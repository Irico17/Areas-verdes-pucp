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
	catastroCtrl  controller.ICatastroController
	permisos      contracts.IPermisosService
}

// NewCatastroGroup creates a new CatastroGroup.
func NewCatastroGroup(
	areaVerdeCtrl controller.IAreaVerdeController,
	catastroCtrl controller.ICatastroController,
	permisos contracts.IPermisosService,
) *CatastroGroup {
	return &CatastroGroup{
		areaVerdeCtrl: areaVerdeCtrl,
		catastroCtrl:  catastroCtrl,
		permisos:      permisos,
	}
}

// Register registers catastro routes on the router.
func (g *CatastroGroup) Register(router gin.IRouter) {
	// Green areas fichas (frente 1A)
	router.GET("/catastro/areas", middleware.RequierePermiso(g.permisos, "consultar"), g.areaVerdeCtrl.Listar)
	router.POST("/catastro/areas", middleware.RequierePermiso(g.permisos, "registrar"), g.areaVerdeCtrl.Crear)
	router.PATCH("/catastro/areas/:id", middleware.RequierePermiso(g.permisos, "registrar"), g.areaVerdeCtrl.Actualizar)
	router.POST("/catastro/areas/:id/baja", middleware.RequierePermiso(g.permisos, "registrar"), g.areaVerdeCtrl.Baja)

	// Catastro maestro (lote 9)
	router.GET("/catastro/zonas-supervision", middleware.RequierePermiso(g.permisos, "consultar"), g.catastroCtrl.Zonas)
	router.POST("/catastro/zonas-supervision", middleware.RequierePermiso(g.permisos, "registrar"), g.catastroCtrl.CrearZona)
	router.PATCH("/catastro/zonas-supervision/:codigo", middleware.RequierePermiso(g.permisos, "registrar"), g.catastroCtrl.ActualizarZona)
	router.POST("/catastro/zonas-supervision/:codigo/baja", middleware.RequierePermiso(g.permisos, "registrar"), g.catastroCtrl.BajaZona)

	router.GET("/catastro/poligonos", middleware.RequierePermiso(g.permisos, "consultar"), g.catastroCtrl.Poligonos)

	router.GET("/catastro/cuadrillas", middleware.RequierePermiso(g.permisos, "consultar"), g.catastroCtrl.Cuadrillas)
	router.POST("/catastro/cuadrillas", middleware.RequierePermiso(g.permisos, "registrar"), g.catastroCtrl.CrearCuadrilla)

	router.GET("/catastro/lugares", middleware.RequierePermiso(g.permisos, "consultar"), g.catastroCtrl.Lugares)
	router.POST("/catastro/lugares", middleware.RequierePermiso(g.permisos, "registrar"), g.catastroCtrl.CrearLugar)

	router.GET("/catastro/especies", middleware.RequierePermiso(g.permisos, "consultar"), g.catastroCtrl.Especies)
	// El alta de especies usa catalogos, no registrar: el capataz sigue registrando en campo.
	router.POST("/catastro/especies", middleware.RequierePermiso(g.permisos, "catalogos"), g.catastroCtrl.CrearEspecie)

	router.GET("/catastro/ejemplares", middleware.RequierePermiso(g.permisos, "consultar"), g.catastroCtrl.Ejemplares)
	router.POST("/catastro/ejemplares", middleware.RequierePermiso(g.permisos, "registrar"), g.catastroCtrl.CrearEjemplar)
	router.PATCH("/catastro/ejemplares/:id", middleware.RequierePermiso(g.permisos, "registrar"), g.catastroCtrl.ActualizarEjemplar)
	router.GET("/catastro/ejemplares/:id/codigos", middleware.RequierePermiso(g.permisos, "consultar"), g.catastroCtrl.Codigos)
	router.POST("/catastro/ejemplares/:id/codigos", middleware.RequierePermiso(g.permisos, "registrar"), g.catastroCtrl.Recodificar)

	// Reference layers
	router.GET("/catastro/fauna", middleware.RequierePermiso(g.permisos, "consultar"), g.catastroCtrl.Fauna)
	router.GET("/catastro/puertas", middleware.RequierePermiso(g.permisos, "consultar"), g.catastroCtrl.Puertas)
	router.GET("/catastro/playas", middleware.RequierePermiso(g.permisos, "consultar"), g.catastroCtrl.Playas)
	router.GET("/catastro/veredas", middleware.RequierePermiso(g.permisos, "consultar"), g.catastroCtrl.Veredas)
	router.GET("/catastro/xerofiticas", middleware.RequierePermiso(g.permisos, "consultar"), g.catastroCtrl.Xerofiticas)
	router.GET("/catastro/jardines-reserva", middleware.RequierePermiso(g.permisos, "consultar"), g.catastroCtrl.JardinesReserva)
}
