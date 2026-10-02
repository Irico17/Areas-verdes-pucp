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
// If openapiEnabled is false, /openapi.yaml is omitted so requests receive 404.
func (group *MetaGroup) Register(router gin.IRouter, openapiEnabled ...bool) {
	router.GET("", group.controller.Index)
	enabled := true
	if len(openapiEnabled) > 0 {
		enabled = openapiEnabled[0]
	}
	if enabled {
		router.GET("/openapi.yaml", group.controller.OpenAPI)
	}
}
