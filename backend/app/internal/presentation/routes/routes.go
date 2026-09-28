// Package routes configures the HTTP router.
package routes

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"go.uber.org/dig"

	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/contracts"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/presentation/middleware"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/presentation/routes/groups"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/shared/config"
)

const basePath = "/areas-verdes/v1"

// Router is the main HTTP router.
type Router struct {
	engine       *gin.Engine
	cfg          *config.Config
	logger       zerolog.Logger
	limitador    contracts.ILimitador
	healthGroup  *groups.HealthGroup
	swaggerGroup *groups.SwaggerGroup
}

// RouterParams contains injected router dependencies.
type RouterParams struct {
	dig.In

	Engine       *gin.Engine
	Config       *config.Config
	Logger       zerolog.Logger
	Limitador    contracts.ILimitador `optional:"true"`
	HealthGroup  *groups.HealthGroup
	SwaggerGroup *groups.SwaggerGroup
}

// NewRouter creates the main router.
func NewRouter(p RouterParams) *Router {
	cfg := p.Config
	if cfg == nil {
		cfg = config.New()
	}
	return &Router{
		engine:       p.Engine,
		cfg:          cfg,
		logger:       p.Logger,
		limitador:    p.Limitador,
		healthGroup:  p.HealthGroup,
		swaggerGroup: p.SwaggerGroup,
	}
}

// Setup registers middleware and service routes.
func (r *Router) Setup() {
	r.engine.Use(
		middleware.Bitacora(r.logger),
		gin.Recovery(),
		middleware.CORS(r.cfg.Seguridad.CORSOrigins),
		middleware.ConCookie(middleware.OpcionesCookie{
			Secure:   r.cfg.Seguridad.CookieSecure,
			SameSite: middleware.ParseSameSite(r.cfg.Seguridad.CookieSameSite),
		}),
		middleware.Errores(),
	)

	if r.limitador != nil {
		r.engine.Use(middleware.LimiteLogin(r.limitador))
	}

	servicePath := r.engine.Group(basePath)
	r.healthGroup.Register(servicePath)

	if r.cfg.Swagger.Enabled {
		r.swaggerGroup.Register(servicePath)
	}

	r.engine.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, gin.H{"error": "ruta no encontrada"})
	})
}
