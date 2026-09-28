package groups

import (
	"github.com/gin-gonic/gin"

	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/presentation/controller"
)

// HealthGroup groups health-check routes.
type HealthGroup struct{ controller controller.IHealthController }

// NewHealthGroup creates the health route group.
func NewHealthGroup(controller controller.IHealthController) *HealthGroup {
	return &HealthGroup{controller: controller}
}

// Register adds the health endpoint to the router.
func (group *HealthGroup) Register(router gin.IRouter) {
	router.GET("/health", group.controller.Health)
}
