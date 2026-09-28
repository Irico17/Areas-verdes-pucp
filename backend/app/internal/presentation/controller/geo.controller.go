// Package controller implements HTTP handlers.
package controller

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/contracts"
	domainEntities "github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/entities"
	domainErrors "github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/errors"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/presentation/requests"
)

// IGeoController defines HTTP endpoints for reading cadastral geometries.
type IGeoController interface {
	Areas(*gin.Context)
	Zonas(*gin.Context)
	Capa(*gin.Context)
	Capas(*gin.Context)
	Resumen(*gin.Context)
	Edificios(*gin.Context)
}

type geoController struct {
	uc contracts.IGeoUseCase
}

// NewGeoController creates a new GeoController.
func NewGeoController(uc contracts.IGeoUseCase) IGeoController {
	return &geoController{uc: uc}
}

// Areas godoc
// @Summary Cadastral green area polygons
// @Tags geo
// @Produce application/geo+json
// @Success 200 {object} entities.FeatureCollection
// @Router /geo/areas [get]
func (ctrl *geoController) Areas(c *gin.Context) {
	f, err := requests.ParseFiltroGeo(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	fc, err := ctrl.uc.Areas(c.Request.Context(), f)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "no se pudo leer el catastro"})
		return
	}
	writeFC(c, fc)
}

// Zonas godoc
// @Summary Cadastral zones polygons
// @Tags geo
// @Produce application/geo+json
// @Success 200 {object} entities.FeatureCollection
// @Router /geo/zonas [get]
func (ctrl *geoController) Zonas(c *gin.Context) {
	f, err := requests.ParseFiltroGeo(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	fc, err := ctrl.uc.Zonas(c.Request.Context(), f)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "no se pudo leer el catastro"})
		return
	}
	writeFC(c, fc)
}

// Capa godoc
// @Summary Auxiliary reference layer
// @Tags geo
// @Produce application/geo+json
// @Param capa path string true "Layer name"
// @Success 200 {object} entities.FeatureCollection
// @Failure 404 {object} map[string]any
// @Router /geo/capas/{capa} [get]
func (ctrl *geoController) Capa(c *gin.Context) {
	f, err := requests.ParseFiltroGeo(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	capa := c.Param("capa")
	fc, err := ctrl.uc.Capa(c.Request.Context(), capa, f)
	if errors.Is(err, domainErrors.ErrCapaDesconocida) {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "capa desconocida",
			"capas": []string{"jardines_reserva", "xerofitica"},
		})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "no se pudo leer el catastro"})
		return
	}
	writeFC(c, fc)
}

// Capas godoc
// @Summary List available auxiliary layers
// @Tags geo
// @Produce json
// @Success 200 {object} dto.CapasIndexDTO
// @Router /geo/capas [get]
func (ctrl *geoController) Capas(c *gin.Context) {
	idx, err := ctrl.uc.Capas(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "no se pudo leer el catastro"})
		return
	}
	c.JSON(http.StatusOK, idx)
}

// Resumen godoc
// @Summary Summary counts of cadastral data
// @Tags geo
// @Produce json
// @Success 200 {object} dto.ResumenDTO
// @Router /geo/resumen [get]
func (ctrl *geoController) Resumen(c *gin.Context) {
	res, err := ctrl.uc.Resumen(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "no se pudo leer el catastro"})
		return
	}
	c.JSON(http.StatusOK, res)
}

// Edificios godoc
// @Summary Campus buildings extract
// @Tags geo
// @Produce application/geo+json
// @Success 200 {object} entities.FeatureCollection
// @Router /geo/edificios [get]
func (ctrl *geoController) Edificios(c *gin.Context) {
	body, err := ctrl.uc.Edificios(c.Request.Context())
	if err != nil {
		writeFC(c, domainEntities.Collection("edificios"))
		return
	}
	c.Data(http.StatusOK, "application/geo+json; charset=utf-8", body)
}

func writeFC(c *gin.Context, fc domainEntities.FeatureCollection) {
	c.Header("Content-Type", "application/geo+json; charset=utf-8")
	c.JSON(http.StatusOK, fc)
}
