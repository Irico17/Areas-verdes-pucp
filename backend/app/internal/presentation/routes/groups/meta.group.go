package groups

import (
	"github.com/gin-gonic/gin"

	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/presentation/controller"
)

// MetaGroup groups metadata and contract routes.
type MetaGroup struct {
	controller controller.IMetaController
}

// NewMetaGroup creates a new metadata route group.
func NewMetaGroup(controller controller.IMetaController) *MetaGroup {
	return &MetaGroup{controller: controller}
}

// Register registers the index and contract routes under the given router.
func (group *MetaGroup) Register(router gin.IRouter) {
	router.GET("", group.controller.Index)
	router.GET("/openapi.yaml", group.controller.OpenAPI)
}
