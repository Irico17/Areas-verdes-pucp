// Package groups registers route handlers into Gin router groups.
package groups

import (
	"github.com/gin-gonic/gin"

	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/contracts"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/presentation/controller"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/presentation/middleware"
)

// AtencionGroup handles routes for solicitudes, ordenes, and riego.
type AtencionGroup struct {
	solicitudCtrl controller.ISolicitudController
	ordenCtrl     controller.IServicioTercerizadoController
	riegoCtrl     controller.IRiegoController
	permisos      contracts.IPermisosService
}

// NewAtencionGroup creates a new AtencionGroup instance.
func NewAtencionGroup(
	solicitudCtrl controller.ISolicitudController,
	ordenCtrl controller.IServicioTercerizadoController,
	riegoCtrl controller.IRiegoController,
	permisos contracts.IPermisosService,
) *AtencionGroup {
	return &AtencionGroup{
		solicitudCtrl: solicitudCtrl,
		ordenCtrl:     ordenCtrl,
		riegoCtrl:     riegoCtrl,
		permisos:      permisos,
	}
}

// Register registers solicitudes, ordenes, and riego routes on the router.
func (g *AtencionGroup) Register(router gin.IRouter) {
	// Solicitudes
	sol := router.Group("/solicitudes")
	sol.GET("", middleware.RequierePermiso(g.permisos, "solicitudes"), g.solicitudCtrl.List)
	sol.POST("", middleware.RequierePermiso(g.permisos, "solicitudes"), g.solicitudCtrl.Create)
	sol.PATCH("/:id", middleware.RequierePermiso(g.permisos, "solicitudes"), g.solicitudCtrl.Edit)

	// Ordenes
	ord := router.Group("/ordenes")
	ord.GET("", middleware.RequierePermiso(g.permisos, "consultar"), g.ordenCtrl.List)
	ord.POST("", middleware.RequierePermiso(g.permisos, "registrar"), g.ordenCtrl.Create)
	ord.GET("/:id/evidencias", middleware.RequierePermiso(g.permisos, "consultar"), g.ordenCtrl.Evidencias)
	ord.PATCH("/:id", middleware.RequierePermiso(g.permisos, "registrar"), g.ordenCtrl.Edit)

	// Riego
	rie := router.Group("/riego")
	rie.GET("", middleware.RequierePermiso(g.permisos, "consultar"), g.riegoCtrl.List)
	rie.POST("", middleware.RequierePermiso(g.permisos, "registrar"), g.riegoCtrl.Create)
}
