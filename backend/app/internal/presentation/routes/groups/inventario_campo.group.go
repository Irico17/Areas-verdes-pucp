// Package groups registers route handlers into Gin router groups.
package groups

import (
	"github.com/gin-gonic/gin"

	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/contracts"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/presentation/controller"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/presentation/middleware"
)

// InventarioCampoGroup handles field inventory routes (/inventario/*).
type InventarioCampoGroup struct {
	controller controller.IInventarioCampoController
	permisos   contracts.IPermisosService
}

// NewInventarioCampoGroup creates a new InventarioCampoGroup instance.
func NewInventarioCampoGroup(
	controller controller.IInventarioCampoController,
	permisos contracts.IPermisosService,
) *InventarioCampoGroup {
	return &InventarioCampoGroup{
		controller: controller,
		permisos:   permisos,
	}
}

// Register registers field inventory routes on the router.
func (g *InventarioCampoGroup) Register(router gin.IRouter) {
	// Tachos (5 routes)
	router.GET("/inventario/tachos", middleware.RequierePermiso(g.permisos, "consultar"), g.controller.ListarTachos)
	router.POST("/inventario/tachos", middleware.RequierePermiso(g.permisos, "registrar"), g.controller.GuardarTacho)
	router.PATCH("/inventario/tachos/:id", middleware.RequierePermiso(g.permisos, "registrar"), g.controller.PatchTacho)
	router.GET("/inventario/tachos.csv", middleware.RequierePermiso(g.permisos, "consultar"), g.controller.CSVTachos)
	router.DELETE("/inventario/tachos/:id", middleware.RequierePermiso(g.permisos, "registrar"), g.controller.BajaTacho)

	// Bebederos (4 routes)
	router.GET("/inventario/bebederos", middleware.RequierePermiso(g.permisos, "consultar"), g.controller.ListarBebederos)
	router.POST("/inventario/bebederos", middleware.RequierePermiso(g.permisos, "registrar"), g.controller.GuardarBebedero)
	router.PATCH("/inventario/bebederos/:id", middleware.RequierePermiso(g.permisos, "registrar"), g.controller.PatchBebedero)
	router.DELETE("/inventario/bebederos/:id", middleware.RequierePermiso(g.permisos, "registrar"), g.controller.BajaBebedero)

	// Puntos PUCP & formato (5 routes)
	router.GET("/inventario/puntos", middleware.RequierePermiso(g.permisos, "consultar"), g.controller.ListarPuntos)
	router.POST("/inventario/puntos", middleware.RequierePermiso(g.permisos, "registrar"), g.controller.GuardarPunto)
	router.PATCH("/inventario/puntos/:id", middleware.RequierePermiso(g.permisos, "registrar"), g.controller.PatchPunto)
	router.DELETE("/inventario/puntos/:id", middleware.RequierePermiso(g.permisos, "registrar"), g.controller.BajaPunto)
	router.POST("/inventario/formato/puntos", middleware.RequierePermiso(g.permisos, "consultar"), g.controller.FormatoPuntos)

	// Reservas jardín (4 routes)
	router.GET("/inventario/reservas", middleware.RequierePermiso(g.permisos, "consultar"), g.controller.ListarReservas)
	router.POST("/inventario/reservas", middleware.RequierePermiso(g.permisos, "registrar"), g.controller.GuardarReserva)
	router.PATCH("/inventario/reservas/:id", middleware.RequierePermiso(g.permisos, "registrar"), g.controller.PatchReserva)
	router.DELETE("/inventario/reservas/:id", middleware.RequierePermiso(g.permisos, "registrar"), g.controller.BajaReserva)

	// Capas editables & export (5 routes)
	router.GET("/inventario/capas/:capa", middleware.RequierePermiso(g.permisos, "consultar"), g.controller.ListarCapa)
	router.POST("/inventario/capas/:capa", middleware.RequierePermiso(g.permisos, "registrar"), g.controller.GuardarCapa)
	router.PATCH("/inventario/capas/:capa/:id", middleware.RequierePermiso(g.permisos, "registrar"), g.controller.PatchCapa)
	router.GET("/inventario/export/:capa", middleware.RequierePermiso(g.permisos, "consultar"), g.controller.CSVCapa)
	router.DELETE("/inventario/capas/:capa/:id", middleware.RequierePermiso(g.permisos, "registrar"), g.controller.BajaCapa)
}
