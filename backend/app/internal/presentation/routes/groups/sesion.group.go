// Package groups groups presentation endpoints.
package groups

import (
	"github.com/gin-gonic/gin"

	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/presentation/controller"
)

// SesionGroup handles authentication and session routes.
type SesionGroup struct {
	controller controller.ISesionController
}

// NewSesionGroup creates a new session group.
func NewSesionGroup(controller controller.ISesionController) *SesionGroup {
	return &SesionGroup{controller: controller}
}

// Register registers session routes on a router.
func (g *SesionGroup) Register(router gin.IRouter) {
	router.POST("/sesion", g.controller.Entrar)
	router.GET("/sesion", g.controller.Actual)
	router.DELETE("/sesion", g.controller.Salir)
}
