package groups

import (
	"net/http"

	"github.com/gin-gonic/gin"
	swaggoFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// SwaggerGroup groups Swagger routes.
type SwaggerGroup struct{}

// NewSwaggerGroup creates a Swagger route group.
func NewSwaggerGroup() *SwaggerGroup {
	return &SwaggerGroup{}
}

// Register adds Swagger UI routes to router.
func (*SwaggerGroup) Register(router *gin.RouterGroup) {
	router.GET("/swagger", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "/areas-verdes/v1/swagger/index.html")
	})
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggoFiles.Handler))
}
