package handlers

import (
	"context"
	"errors"
	"log"
	"os"
	"strconv"

	"campusverde/api/internal/catastro"
	"campusverde/api/internal/geojson"

	"github.com/gin-gonic/gin"
)

// GeoSource lee el catastro. *catastro.Store la cumple.
type GeoSource interface {
	Areas(ctx context.Context, f catastro.Filter) (geojson.FeatureCollection, error)
	Zonas(ctx context.Context, f catastro.Filter) (geojson.FeatureCollection, error)
	Capa(ctx context.Context, capa string, f catastro.Filter) (geojson.FeatureCollection, error)
	Capas(ctx context.Context) (catastro.CapasIndex, error)
	Resumen(ctx context.Context) (catastro.Resumen, error)
}

// Geo expone FeatureCollections de catastro.
type Geo struct {
	Source        GeoSource
	EdificiosPath string
}

func (h Geo) Areas(c *gin.Context) {
	h.collection(c, func(ctx context.Context, f catastro.Filter) (geojson.FeatureCollection, error) {
		return h.Source.Areas(ctx, f)
	})
}

func (h Geo) Zonas(c *gin.Context) {
	h.collection(c, func(ctx context.Context, f catastro.Filter) (geojson.FeatureCollection, error) {
		return h.Source.Zonas(ctx, f)
	})
}

func (h Geo) Capa(c *gin.Context) {
	if h.Source == nil {
		c.JSON(503, gin.H{"error": "base de datos no disponible"})
		return
	}
	f, err := parseFilter(c)
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	fc, err := h.Source.Capa(c.Request.Context(), c.Param("capa"), f)
	if errors.Is(err, catastro.ErrCapaDesconocida) {
		c.JSON(404, gin.H{"error": "capa desconocida", "capas": catastro.CapasConocidas})
		return
	}
	if err != nil {
		log.Printf("geo capa: %v", err)
		c.JSON(500, gin.H{"error": "no se pudo leer el catastro"})
		return
	}
	writeFC(c, fc)
}

func (h Geo) Capas(c *gin.Context) {
	if h.Source == nil {
		c.JSON(503, gin.H{"error": "base de datos no disponible"})
		return
	}
	idx, err := h.Source.Capas(c.Request.Context())
	if err != nil {
		log.Printf("geo capas: %v", err)
		c.JSON(500, gin.H{"error": "no se pudo leer el catastro"})
		return
	}
	c.JSON(200, idx)
}

// Edificios sirve el extracto OSM commiteado. Si falta el archivo, responde colección vacía.
func (h Geo) Edificios(c *gin.Context) {
	body, err := os.ReadFile(h.EdificiosPath)
	if err != nil {
		writeFC(c, geojson.Collection("edificios"))
		return
	}
	c.Data(200, "application/geo+json; charset=utf-8", body)
}

func (h Geo) Resumen(c *gin.Context) {
	if h.Source == nil {
		c.JSON(503, gin.H{"error": "base de datos no disponible"})
		return
	}
	res, err := h.Source.Resumen(c.Request.Context())
	if err != nil {
		log.Printf("geo resumen: %v", err)
		c.JSON(500, gin.H{"error": "no se pudo leer el catastro"})
		return
	}
	c.JSON(200, res)
}

func (h Geo) collection(c *gin.Context, load func(context.Context, catastro.Filter) (geojson.FeatureCollection, error)) {
	if h.Source == nil {
		c.JSON(503, gin.H{"error": "base de datos no disponible"})
		return
	}
	f, err := parseFilter(c)
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	fc, err := load(c.Request.Context(), f)
	if err != nil {
		log.Printf("geo: %v", err)
		c.JSON(500, gin.H{"error": "no se pudo leer el catastro"})
		return
	}
	writeFC(c, fc)
}

func writeFC(c *gin.Context, fc geojson.FeatureCollection) {
	c.Header("Content-Type", "application/geo+json; charset=utf-8")
	c.JSON(200, fc)
}

func parseFilter(c *gin.Context) (catastro.Filter, error) {
	bbox, err := geojson.ParseBBox(c.Query("bbox"))
	if err != nil {
		return catastro.Filter{}, err
	}
	limit, err := parseLimit(c.Query("limit"))
	if err != nil {
		return catastro.Filter{}, err
	}
	return catastro.Filter{BBox: bbox, Limit: limit}, nil
}

// RegistrarGeo publica el catastro de lectura. registrar(r, deps) del frente geo.
// 1A añade su línea en server.go; no reescribe este bloque salvo paths nuevos de ejemplar.
func RegistrarGeo(r *gin.Engine, deps Deps) {
	geo := Geo{EdificiosPath: deps.EdificiosPath}
	if deps.DB != nil {
		geo.Source = catastro.NewStore(deps.DB)
	}
	v1 := r.Group("/api/v1/geo")
	v1.GET("/resumen", geo.Resumen)
	v1.GET("/areas", geo.Areas)
	v1.GET("/zonas", geo.Zonas)
	v1.GET("/capas", geo.Capas)
	v1.GET("/capas/:capa", geo.Capa)
	v1.GET("/edificios", geo.Edificios)
}

func parseLimit(raw string) (int, error) {
	if raw == "" {
		return 0, nil
	}
	n, err := strconv.Atoi(raw)
	if err != nil || n < 1 || n > 10000 {
		return 0, errLimit
	}
	return n, nil
}
