package groups

import (
	"github.com/gin-gonic/gin"

	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/presentation/controller"
)

// LegadoGroup groups legacy root routes.
type LegadoGroup struct {
	healthController controller.IHealthController
	metaController   controller.IMetaController
}

// NewLegadoGroup creates a new legacy route group.
func NewLegadoGroup(health controller.IHealthController, meta controller.IMetaController) *LegadoGroup {
	return &LegadoGroup{
		healthController: health,
		metaController:   meta,
	}
}

// Register registers legacy root endpoints on the router.
func (group *LegadoGroup) Register(router gin.IRouter) {
	router.GET("/health", group.healthController.LegacyHealth)
}
