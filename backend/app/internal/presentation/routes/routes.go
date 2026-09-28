// Package routes configures the HTTP router.
package routes

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"go.uber.org/dig"

	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/contracts"
	domainErrors "github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/errors"
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
	engine        *gin.Engine
	cfg           *config.Config
	logger        zerolog.Logger
	limitador     contracts.ILimitador
	sesionUC      contracts.ISesionUseCase
	healthGroup   *groups.HealthGroup
	metaGroup     *groups.MetaGroup
	legadoGroup   *groups.LegadoGroup
	swaggerGroup  *groups.SwaggerGroup
	sesionGroup   *groups.SesionGroup
	accesosGroup  *groups.AccesosGroup
	catalogoGroup *groups.CatalogoGroup
	geoGroup      *groups.GeoGroup
	catastroGroup *groups.CatastroGroup
}

// RouterParams contains injected router dependencies.
type RouterParams struct {
	dig.In

	Engine        *gin.Engine
	Config        *config.Config
	Logger        zerolog.Logger
	Limitador     contracts.ILimitador
	SesionUC      contracts.ISesionUseCase
	HealthGroup   *groups.HealthGroup
	MetaGroup     *groups.MetaGroup
	LegadoGroup   *groups.LegadoGroup
	SwaggerGroup  *groups.SwaggerGroup
	SesionGroup   *groups.SesionGroup
	AccesosGroup  *groups.AccesosGroup
	CatalogoGroup *groups.CatalogoGroup
	GeoGroup      *groups.GeoGroup
	CatastroGroup *groups.CatastroGroup
}

// NewRouter creates the main router.
func NewRouter(p RouterParams) *Router {
	return &Router{
		engine:        p.Engine,
		cfg:           p.Config,
		logger:        p.Logger,
		limitador:     p.Limitador,
		sesionUC:      p.SesionUC,
		healthGroup:   p.HealthGroup,
		metaGroup:     p.MetaGroup,
		legadoGroup:   p.LegadoGroup,
		swaggerGroup:  p.SwaggerGroup,
		sesionGroup:   p.SesionGroup,
		accesosGroup:  p.AccesosGroup,
		catalogoGroup: p.CatalogoGroup,
		geoGroup:      p.GeoGroup,
		catastroGroup: p.CatastroGroup,
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
		middleware.Errores(r.logger),
		middleware.LimiteLogin(r.limitador),
		middleware.Auth(r.sesionUC),
	)

	servicePath := r.engine.Group(basePath)
	legacyPath := r.engine.Group(legacyBasePath)

	// Health group is mounted only on /areas-verdes/v1 per decision 17
	r.healthGroup.Register(servicePath)

	for _, prefix := range []gin.IRouter{servicePath, legacyPath} {
		r.metaGroup.Register(prefix)
		r.sesionGroup.Register(prefix)
		r.accesosGroup.Register(prefix)
		r.catalogoGroup.Register(prefix)
		r.geoGroup.Register(prefix)
		r.catastroGroup.Register(prefix)
	}

	if r.cfg.Swagger.Enabled {
		r.swaggerGroup.Register(servicePath)
	}

	r.legadoGroup.Register(r.engine)

	r.engine.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, gin.H{"error": domainErrors.ErrRutaNoEncontrada.Error()})
	})
}
