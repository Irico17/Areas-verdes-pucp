package server

import (
	"net/http"
	"time"

	"campusverde/api/internal/accesos"
	"campusverde/api/internal/handlers"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// Deps son las dependencias del HTTP server.
type Deps struct {
	DB               *gorm.DB
	OpenAPIPath      string
	EdificiosPath    string
	ReservasPath     string
	FotosDir         string
	EvidenciasDir    string
	EvidenciasBucket string
	Seguridad        handlers.OpcionesDeSeguridad
}

// New arma el router Gin de la API.
func New(deps Deps) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	_ = r.SetTrustedProxies([]string{})
	seg := deps.Seguridad
	if seg.LoginCada < 1 {
		seg.LoginCada = 8
	}
	if seg.LoginVentana <= 0 {
		seg.LoginVentana = time.Minute
	}
	r.Use(Bitacora(), gin.Recovery(), handlers.CORS(seg.CORSOrigins), handlers.ConCookie(seg.Cookie), handlers.LimiteLogin(seg.LoginCada, seg.LoginVentana))

	var acc *accesos.Store
	if deps.DB != nil {
		acc = accesos.NewStore(deps.DB)
	}
	r.Use(func(c *gin.Context) {
		if acc != nil {
			if token := accesos.Token(c.Request); token != "" {
				if u, err := acc.FromToken(c.Request.Context(), token); err == nil {
					c.Set("usuario", u)
				}
			}
		}
		c.Next()
	})

	hdeps := handlers.Deps{
		DB:               deps.DB,
		OpenAPIPath:      deps.OpenAPIPath,
		EdificiosPath:    deps.EdificiosPath,
		ReservasPath:     deps.ReservasPath,
		FotosDir:         deps.FotosDir,
		EvidenciasDir:    deps.EvidenciasDir,
		EvidenciasBucket: deps.EvidenciasBucket,
	}

	// Registro de rutas. Cada frente añade una sola línea registrar(r, deps)
	// y su archivo handlers/<frente>.go o openapi/<tag>.yaml.
	// No reescribas este bloque.
	handlers.RegistrarSistema(r, hdeps)
	handlers.RegistrarGeo(r, hdeps)
	handlers.RegistrarInventario(r, hdeps)
	handlers.RegistrarReservas(r, hdeps)
	handlers.RegistrarOperacion(r, hdeps)
	handlers.RegistrarAccesos(r, hdeps)
	handlers.RegistrarCatalogos(r, hdeps)
	handlers.RegistrarCatastro(r, hdeps)
	handlers.RegistrarAtencion(r, hdeps)
	handlers.RegistrarEjemplares(r, hdeps)
	// 1B seguridad:         middleware de arriba; sin línea nueva de ruta
	// 1D evidencias:        ya está en RegistrarAtencion; un path nuevo es handlers.RegistrarEvidencias(r, hdeps)
	handlers.RegistrarPoda(r, hdeps)
	handlers.RegistrarFrente2B(r, hdeps)
	handlers.RegistrarAuditoria(r, hdeps)
	// 2D reportes:          ampliar RegistrarAtencion
	// 3A importaciones:     handlers.RegistrarImportaciones(r, hdeps)

	r.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, gin.H{"error": "ruta no encontrada"})
	})
	return r
}
