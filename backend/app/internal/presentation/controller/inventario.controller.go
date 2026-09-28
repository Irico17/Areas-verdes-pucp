// Package controller contains HTTP handlers for presenting domain data.
package controller

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"

	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/contracts"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/entities"
	domainErrors "github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/errors"
)

// IInventarioController defines HTTP handlers for legacy inventory overlays and photos.
type IInventarioController interface {
	Index(*gin.Context)
	Capa(*gin.Context)
	Foto(*gin.Context)
}

type inventarioController struct {
	uc     contracts.IInventarioUseCase
	logger zerolog.Logger
}

// NewInventarioController creates a new InventarioController.
func NewInventarioController(uc contracts.IInventarioUseCase, logger zerolog.Logger) IInventarioController {
	return &inventarioController{uc: uc, logger: logger}
}

// Index godoc
// @Summary Inventory overlay layers catalogue and loaded counts
// @Tags geo
// @Produce json
// @Success 200 {object} dto.IndiceInventarioDTO
// @Failure 500 {object} map[string]string
// @Router /geo/inventario [get]
func (ctrl *inventarioController) Index(c *gin.Context) {
	idx, err := ctrl.uc.Index(c.Request.Context())
	if err != nil {
		ctrl.logger.Error().Err(err).Msg("inventario: error leyendo indice")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "no se pudo leer el inventario"})
		return
	}
	c.JSON(http.StatusOK, idx)
}

// Capa godoc
// @Summary Inventory overlay layer features
// @Tags geo
// @Produce application/geo+json
// @Param capa path string true "Layer name"
// @Success 200 {object} entities.FeatureCollection
// @Failure 404 {object} dto.CapaDesconocidaErrorDTO
// @Failure 500 {object} map[string]string
// @Router /geo/inventario/{capa} [get]
func (ctrl *inventarioController) Capa(c *gin.Context) {
	capa := c.Param("capa")
	fc, err := ctrl.uc.Capa(c.Request.Context(), capa)
	if errors.Is(err, domainErrors.ErrCapaInventarioDesconocida) {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "capa de inventario desconocida",
			"capas": entities.CapasConocidas,
		})
		return
	}
	if err != nil {
		ctrl.logger.Error().Err(err).Str("capa", capa).Msg("inventario capa: error leyendo capa")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "no se pudo leer el inventario"})
		return
	}
	c.Header("Content-Type", "application/geo+json; charset=utf-8")
	c.JSON(http.StatusOK, fc)
}

// Foto godoc
// @Summary Serve local inventory JPEG photograph
// @Tags geo
// @Produce image/jpeg
// @Param name path string true "Photo file name"
// @Success 200
// @Failure 404 {object} map[string]string
// @Router /geo/inventario/fotos/{name} [get]
func (ctrl *inventarioController) Foto(c *gin.Context) {
	path, err := ctrl.uc.Foto(c.Request.Context(), c.Param("name"))
	if errors.Is(err, domainErrors.ErrFotografiaNoDisponible) {
		c.JSON(http.StatusNotFound, gin.H{"error": "fotografía no disponible"})
		return
	}
	if errors.Is(err, domainErrors.ErrFotografiaNoRecuperada) {
		c.JSON(http.StatusNotFound, gin.H{"error": "fotografía no recuperada"})
		return
	}
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "fotografía no recuperada"})
		return
	}
	c.File(path)
}
