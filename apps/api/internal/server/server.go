package server

import (
	"net/http"

	"campusverde/api/internal/accesos"
	"campusverde/api/internal/atencion"
	"campusverde/api/internal/blobs"
	"campusverde/api/internal/catalogos"
	"campusverde/api/internal/catastro"
	"campusverde/api/internal/handlers"
	"campusverde/api/internal/inventario"
	"campusverde/api/internal/operacion"

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
}

// New arma el router Gin de la API.
func New(deps Deps) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	_ = r.SetTrustedProxies([]string{})
	r.Use(gin.Logger(), gin.Recovery(), cors)

	var acc *accesos.Store
	var cats *catalogos.Store
	var aten *atencion.Store
	if deps.DB != nil {
		acc = accesos.NewStore(deps.DB)
		cats = catalogos.NewStore(deps.DB)
		aten = atencion.NewStore(deps.DB)
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

	var store *catastro.Store
	if deps.DB != nil {
		store = catastro.NewStore(deps.DB)
	}

	health := handlers.Health{DB: deps.DB, Store: store}
	geo := handlers.Geo{EdificiosPath: deps.EdificiosPath}
	if store != nil {
		geo.Source = store
	}
	meta := handlers.Meta{OpenAPIPath: deps.OpenAPIPath}

	var op handlers.Operacion
	var inv handlers.Inventario
	inv.FotosDir = deps.FotosDir
	if deps.DB != nil {
		op.Store = operacion.NewStore(deps.DB)
		inv.Store = inventario.NewStore(deps.DB)
	}
	reservas := handlers.Reservas{Path: deps.ReservasPath}

	r.GET("/health", health.Get)
	r.GET("/api/v1", meta.Index)
	r.GET("/api/v1/openapi.yaml", meta.OpenAPI)

	v1 := r.Group("/api/v1/geo")
	v1.GET("/resumen", geo.Resumen)
	v1.GET("/areas", geo.Areas)
	v1.GET("/zonas", geo.Zonas)
	v1.GET("/capas", geo.Capas)
	v1.GET("/capas/:capa", geo.Capa)
	v1.GET("/edificios", geo.Edificios)
	v1.GET("/inventario", inv.Index)
	v1.GET("/inventario/fotos/:name", inv.Foto)
	v1.GET("/inventario/:capa", inv.Capa)
	v1.GET("/reservas-mock", reservas.Get)

	lab := r.Group("/api/v1/operacion")
	lab.GET("/capataces", op.Capataces)
	lab.GET("/actividades", op.List)
	lab.POST("/actividades", op.Create)
	lab.PATCH("/actividades/:id/asignacion", op.Assign)
	lab.PATCH("/actividades/:id/estado", op.Estado)
	lab.POST("/actividades/:id/archivar", op.Archive)
	lab.GET("/actividades/:id/timeline", op.Timeline)

	ses := handlers.Sesion{Store: acc}
	r.POST("/api/v1/sesion", ses.Entrar)
	r.GET("/api/v1/sesion", ses.Actual)
	r.DELETE("/api/v1/sesion", ses.Salir)
	r.GET("/api/v1/accesos/usuarios", ses.Usuarios)

	cat := handlers.Catalogo{Store: cats}
	r.GET("/api/v1/catalogos", cat.List)
	r.POST("/api/v1/catalogos", cat.Create)
	r.POST("/api/v1/catalogos/:id/desactivar", cat.Off)

	fichas := handlers.Fichas{Store: store}
	r.GET("/api/v1/catastro/areas", fichas.List)
	r.POST("/api/v1/catastro/areas", fichas.Create)
	r.PATCH("/api/v1/catastro/areas/:id", fichas.Patch)

	at := handlers.Atencion{Store: aten, Dir: deps.EvidenciasDir, Files: blobs.Open(deps.EvidenciasDir, deps.EvidenciasBucket)}
	r.GET("/api/v1/solicitudes", at.Solicitudes)
	r.POST("/api/v1/solicitudes", at.CrearSolicitud)
	r.GET("/api/v1/ordenes", at.Ordenes)
	r.POST("/api/v1/ordenes", at.CrearOrden)
	r.GET("/api/v1/riego", at.Riego)
	r.POST("/api/v1/riego", at.CrearRiego)
	r.GET("/api/v1/evidencias", at.Evidencias)
	r.POST("/api/v1/evidencias", at.SubirEvidencia)
	r.GET("/api/v1/evidencias/:id/archivo", at.Archivo)
	r.GET("/api/v1/reportes/labores", at.Reporte)
	r.POST("/api/v1/ia/sugerir-tipo", at.Sugerir)

	r.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, gin.H{"error": "ruta no encontrada"})
	})
	return r
}

func cors(c *gin.Context) {
	c.Header("Access-Control-Allow-Origin", "*")
	c.Header("Access-Control-Allow-Methods", "GET, POST, PATCH, DELETE, OPTIONS")
	c.Header("Access-Control-Allow-Headers", "Content-Type")
	if c.Request.Method == http.MethodOptions {
		c.AbortWithStatus(http.StatusNoContent)
		return
	}
	c.Next()
}
