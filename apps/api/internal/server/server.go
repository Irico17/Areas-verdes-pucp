package server

import (
	"net/http"

	"campusverde/api/internal/catastro"
	"campusverde/api/internal/handlers"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// Deps son las dependencias del HTTP server.
type Deps struct {
	DB          *gorm.DB
	OpenAPIPath string
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
	geo := handlers.Geo{}
	if store != nil {
		geo.Source = store
	}
	meta := handlers.Meta{OpenAPIPath: deps.OpenAPIPath}

	r.GET("/health", health.Get)
	r.GET("/api/v1", meta.Index)
	r.GET("/api/v1/openapi.yaml", meta.OpenAPI)

	v1 := r.Group("/api/v1/geo")
	v1.GET("/resumen", geo.Resumen)
	v1.GET("/areas", geo.Areas)
	v1.GET("/zonas", geo.Zonas)
	v1.GET("/capas", geo.Capas)
	v1.GET("/capas/:capa", geo.Capa)

	r.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, gin.H{"error": "ruta no encontrada"})
	})
	return r
}

func cors(c *gin.Context) {
	c.Header("Access-Control-Allow-Origin", "*")
	c.Header("Access-Control-Allow-Methods", "GET, OPTIONS")
	c.Header("Access-Control-Allow-Headers", "Content-Type")
	if c.Request.Method == http.MethodOptions {
		c.AbortWithStatus(http.StatusNoContent)
		return
	}
	c.Next()
}
