package server

import (
	"net/http"

	"campusverde/api/internal/catastro"
	"campusverde/api/internal/handlers"
	"campusverde/api/internal/operacion"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// Deps son las dependencias del HTTP server.
type Deps struct {
	DB            *gorm.DB
	OpenAPIPath   string
	EdificiosPath string
}

// New arma el router Gin de la API.
func New(deps Deps) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	_ = r.SetTrustedProxies([]string{})
	r.Use(gin.Logger(), gin.Recovery(), cors)

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
	if deps.DB != nil {
		op.Store = operacion.NewStore(deps.DB)
	}

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

	lab := r.Group("/api/v1/operacion")
	lab.GET("/capataces", op.Capataces)
	lab.GET("/actividades", op.List)
	lab.POST("/actividades", op.Create)
	lab.PATCH("/actividades/:id/asignacion", op.Assign)
	lab.PATCH("/actividades/:id/estado", op.Estado)
	lab.POST("/actividades/:id/archivar", op.Archive)
	lab.GET("/actividades/:id/timeline", op.Timeline)

	r.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, gin.H{"error": "ruta no encontrada"})
	})
	return r
}

func cors(c *gin.Context) {
	c.Header("Access-Control-Allow-Origin", "*")
	c.Header("Access-Control-Allow-Methods", "GET, POST, PATCH, OPTIONS")
	c.Header("Access-Control-Allow-Headers", "Content-Type")
	if c.Request.Method == http.MethodOptions {
		c.AbortWithStatus(http.StatusNoContent)
		return
	}
	c.Next()
}
