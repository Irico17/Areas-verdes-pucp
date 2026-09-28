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

const (
	basePath       = "/areas-verdes/v1"
	legacyBasePath = "/api/v1"
)

// Router is the main HTTP router.
type Router struct {
	engine       *gin.Engine
	cfg          *config.Config
	logger       zerolog.Logger
	limitador    contracts.ILimitador
	healthGroup  *groups.HealthGroup
	metaGroup    *groups.MetaGroup
	legadoGroup  *groups.LegadoGroup
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
	MetaGroup    *groups.MetaGroup
	LegadoGroup  *groups.LegadoGroup
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
		metaGroup:    p.MetaGroup,
		legadoGroup:  p.LegadoGroup,
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
	legacyPath := r.engine.Group(legacyBasePath)

	for _, prefix := range []gin.IRouter{servicePath, legacyPath} {
		if r.healthGroup != nil {
			r.healthGroup.Register(prefix)
		}
		if r.metaGroup != nil {
			r.metaGroup.Register(prefix)
		}
	}

	if r.cfg.Swagger.Enabled && r.swaggerGroup != nil {
		r.swaggerGroup.Register(servicePath)
	}

	if r.legadoGroup != nil {
		r.legadoGroup.Register(r.engine)
	}

	r.engine.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, gin.H{"error": "ruta no encontrada"})
	})
}
