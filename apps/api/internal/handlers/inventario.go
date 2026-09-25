package handlers

import (
	"errors"
	"log"
	"os"
	"path/filepath"
	"strings"

	"campusverde/api/internal/inventario"

	"github.com/gin-gonic/gin"
)

// Inventario sirve overlays legacy y, si existen, JPEG locales.
type Inventario struct {
	Store    *inventario.Store
	FotosDir string
}

func (h Inventario) Index(c *gin.Context) {
	if h.Store == nil {
		c.JSON(503, gin.H{"error": "base de datos no disponible"})
		return
	}
	idx, err := h.Store.Index(c.Request.Context())
	if err != nil {
		log.Printf("inventario: %v", err)
		c.JSON(500, gin.H{"error": "no se pudo leer el inventario"})
		return
	}
	c.JSON(200, idx)
}

func (h Inventario) Capa(c *gin.Context) {
	if h.Store == nil {
		c.JSON(503, gin.H{"error": "base de datos no disponible"})
		return
	}
	fc, err := h.Store.Capa(c.Request.Context(), c.Param("capa"))
	if errors.Is(err, inventario.ErrCapa) {
		c.JSON(404, gin.H{"error": "capa de inventario desconocida", "capas": inventario.Capas})
		return
	}
	if err != nil {
		log.Printf("inventario capa: %v", err)
		c.JSON(500, gin.H{"error": "no se pudo leer el inventario"})
		return
	}
	c.Header("Content-Type", "application/geo+json; charset=utf-8")
	c.JSON(200, fc)
}

// RegistrarInventario publica overlays. registrar(r, deps) del frente inventario.
// 2B amplía este archivo. El orden fotos antes de :capa se conserva.
func RegistrarInventario(r *gin.Engine, deps Deps) {
	inv := Inventario{FotosDir: deps.FotosDir}
	if deps.DB != nil {
		inv.Store = inventario.NewStore(deps.DB)
	}
	g := r.Group("/api/v1/geo")
	g.GET("/inventario", inv.Index)
	g.GET("/inventario/fotos/:name", inv.Foto)
	g.GET("/inventario/:capa", inv.Capa)
}

// Foto entrega un JPEG recuperado. No publica el índice ni ids de Drive.
func (h Inventario) Foto(c *gin.Context) {
	name := filepath.Base(c.Param("name"))
	lower := strings.ToLower(name)
	if name == "." || name == "/" || (!strings.HasSuffix(lower, ".jpg") && !strings.HasSuffix(lower, ".jpeg")) {
		c.JSON(404, gin.H{"error": "fotografía no disponible"})
		return
	}
	if h.FotosDir == "" {
		c.JSON(404, gin.H{"error": "fotografía no disponible"})
		return
	}
	path := filepath.Join(h.FotosDir, name)
	if _, err := os.Stat(path); err != nil {
		c.JSON(404, gin.H{"error": "fotografía no recuperada"})
		return
	}
	c.File(path)
}
