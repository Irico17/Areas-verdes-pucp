// Package routes configures the HTTP router.
package routes

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/dig"

	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/presentation/routes/groups"
)

const basePath = "/areas-verdes/v1"

// Router is the main HTTP router.
type Router struct {
	engine       *gin.Engine
	healthGroup  *groups.HealthGroup
	swaggerGroup *groups.SwaggerGroup
}

// RouterParams contains injected router dependencies.
type RouterParams struct {
	dig.In

	Engine       *gin.Engine
	HealthGroup  *groups.HealthGroup
	SwaggerGroup *groups.SwaggerGroup
}

// NewRouter creates the main router.
func NewRouter(p RouterParams) *Router {
	return &Router{
		engine:       p.Engine,
		healthGroup:  p.HealthGroup,
		swaggerGroup: p.SwaggerGroup,
	}
}

// Setup registers middleware and service routes.
func (r *Router) Setup() {
	r.engine.Use(gin.Recovery())

	servicePath := r.engine.Group(basePath)
	r.healthGroup.Register(servicePath)
	r.swaggerGroup.Register(servicePath)
}
