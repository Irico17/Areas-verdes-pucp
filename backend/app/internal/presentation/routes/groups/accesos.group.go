package groups

import (
	"github.com/gin-gonic/gin"

	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/presentation/controller"
)

// AccesosGroup handles access management routes.
type AccesosGroup struct {
	controller controller.IUsuarioController
}

// NewAccesosGroup creates a new accesos group.
func NewAccesosGroup(controller controller.IUsuarioController) *AccesosGroup {
	return &AccesosGroup{controller: controller}
}

// Register registers accesos routes on a router.
func (g *AccesosGroup) Register(router gin.IRouter) {
	router.GET("/accesos/usuarios", g.controller.Listar)
	router.POST("/accesos/usuarios", g.controller.Crear)
	router.PATCH("/accesos/usuarios/:usuario", g.controller.Actualizar)
	router.PATCH("/accesos/permisos", g.controller.ActualizarPermiso)
	router.POST("/accesos/roles", g.controller.CrearRol)
	router.PATCH("/accesos/roles/:codigo", g.controller.ActualizarRol)
	router.POST("/sesion/clave", g.controller.CambiarClavePropia)
}
