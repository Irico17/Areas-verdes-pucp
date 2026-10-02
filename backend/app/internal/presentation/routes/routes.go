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
	engine               *gin.Engine
	cfg                  *config.Config
	logger               zerolog.Logger
	limitador            contracts.ILimitador
	sesionUC             contracts.ISesionUseCase
	healthGroup          *groups.HealthGroup
	metaGroup            *groups.MetaGroup
	legadoGroup          *groups.LegadoGroup
	swaggerGroup         *groups.SwaggerGroup
	sesionGroup          *groups.SesionGroup
	accesosGroup         *groups.AccesosGroup
	catalogoGroup        *groups.CatalogoGroup
	geoGroup             *groups.GeoGroup
	catastroGroup        *groups.CatastroGroup
	inventarioGroup      *groups.InventarioGroup
	reservasMockGroup    *groups.ReservasMockGroup
	inventarioCampoGroup *groups.InventarioCampoGroup
	operacionGroup       *groups.OperacionGroup
	atencionGroup        *groups.AtencionGroup
	podaGroup            *groups.PodaGroup
	viveroGroup          *groups.ViveroGroup
	evidenciaGroup       *groups.EvidenciaGroup
	reporteGroup         *groups.ReporteGroup
	iaGroup              *groups.IAGroup
	auditoriaGroup       *groups.AuditoriaGroup
	importacionGroup     *groups.ImportacionGroup
	zonificacionGroup    *groups.ZonificacionGroup
}

// RouterParams contains injected router dependencies.
type RouterParams struct {
	dig.In

	Engine               *gin.Engine
	Config               *config.Config
	Logger               zerolog.Logger
	Limitador            contracts.ILimitador
	SesionUC             contracts.ISesionUseCase
	HealthGroup          *groups.HealthGroup
	MetaGroup            *groups.MetaGroup
	LegadoGroup          *groups.LegadoGroup
	SwaggerGroup         *groups.SwaggerGroup
	SesionGroup          *groups.SesionGroup
	AccesosGroup         *groups.AccesosGroup
	CatalogoGroup        *groups.CatalogoGroup
	GeoGroup             *groups.GeoGroup
	CatastroGroup        *groups.CatastroGroup
	InventarioGroup      *groups.InventarioGroup
	ReservasMockGroup    *groups.ReservasMockGroup
	InventarioCampoGroup *groups.InventarioCampoGroup
	OperacionGroup       *groups.OperacionGroup
	AtencionGroup        *groups.AtencionGroup
	PodaGroup            *groups.PodaGroup
	ViveroGroup          *groups.ViveroGroup
	EvidenciaGroup       *groups.EvidenciaGroup
	ReporteGroup         *groups.ReporteGroup
	IAGroup              *groups.IAGroup
	AuditoriaGroup       *groups.AuditoriaGroup
	ImportacionGroup     *groups.ImportacionGroup
	ZonificacionGroup    *groups.ZonificacionGroup
}

// NewRouter creates the main router.
func NewRouter(p RouterParams) *Router {
	return &Router{
		engine:               p.Engine,
		cfg:                  p.Config,
		logger:               p.Logger,
		limitador:            p.Limitador,
		sesionUC:             p.SesionUC,
		healthGroup:          p.HealthGroup,
		metaGroup:            p.MetaGroup,
		legadoGroup:          p.LegadoGroup,
		swaggerGroup:         p.SwaggerGroup,
		sesionGroup:          p.SesionGroup,
		accesosGroup:         p.AccesosGroup,
		catalogoGroup:        p.CatalogoGroup,
		geoGroup:             p.GeoGroup,
		catastroGroup:        p.CatastroGroup,
		inventarioGroup:      p.InventarioGroup,
		reservasMockGroup:    p.ReservasMockGroup,
		inventarioCampoGroup: p.InventarioCampoGroup,
		operacionGroup:       p.OperacionGroup,
		atencionGroup:        p.AtencionGroup,
		podaGroup:            p.PodaGroup,
		viveroGroup:          p.ViveroGroup,
		evidenciaGroup:       p.EvidenciaGroup,
		reporteGroup:         p.ReporteGroup,
		iaGroup:              p.IAGroup,
		auditoriaGroup:       p.AuditoriaGroup,
		importacionGroup:     p.ImportacionGroup,
		zonificacionGroup:    p.ZonificacionGroup,
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

	openapiEnabled := r.cfg == nil || r.cfg.AppEnv != "produccion"
	for _, prefix := range []gin.IRouter{servicePath, legacyPath} {
		r.metaGroup.Register(prefix, openapiEnabled)
		r.sesionGroup.Register(prefix)
		r.accesosGroup.Register(prefix)
		r.catalogoGroup.Register(prefix)
		r.geoGroup.Register(prefix)
		r.catastroGroup.Register(prefix)
		r.inventarioGroup.Register(prefix)
		r.reservasMockGroup.Register(prefix)
		r.inventarioCampoGroup.Register(prefix)
		r.operacionGroup.Register(prefix)
		r.atencionGroup.Register(prefix)
		r.podaGroup.Register(prefix)
		r.viveroGroup.Register(prefix)
		r.evidenciaGroup.Register(prefix)
		r.reporteGroup.Register(prefix)
		r.iaGroup.Register(prefix)
		r.auditoriaGroup.Register(prefix)
		r.importacionGroup.Register(prefix)
		r.zonificacionGroup.Register(prefix)
	}

	if r.cfg.Swagger.Enabled {
		r.swaggerGroup.Register(servicePath)
	}

	r.legadoGroup.Register(r.engine)

	r.engine.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, gin.H{"error": domainErrors.ErrRutaNoEncontrada.Error()})
	})
}
