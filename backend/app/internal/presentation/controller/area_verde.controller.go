package controller

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"

	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/contracts"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/dto"
	domainErrors "github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/errors"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/presentation/middleware"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/presentation/requests"
)

// IAreaVerdeController defines HTTP handlers for area verdes and fichas.
type IAreaVerdeController interface {
	Listar(*gin.Context)
	Actualizar(*gin.Context)
	Crear(*gin.Context)
	Baja(*gin.Context)
}

type areaVerdeController struct {
	uc     contracts.IAreaVerdeUseCase
	logger zerolog.Logger
}

// NewAreaVerdeController creates a new AreaVerdeController instance.
func NewAreaVerdeController(uc contracts.IAreaVerdeUseCase, logger zerolog.Logger) IAreaVerdeController {
	return &areaVerdeController{uc: uc, logger: logger}
}

// Listar godoc
// @Summary List green area metadata fichas
// @Tags catastro
// @Produce json
// @Param q query string false "Search query"
// @Success 200 {object} map[string][]dto.FichaDTO
// @Router /v1/catastro/areas [get]
func (ctrl *areaVerdeController) Listar(c *gin.Context) {
	rows, err := ctrl.uc.Fichas(c.Request.Context(), c.Query("q"))
	if err != nil {
		ctrl.logger.Error().Err(err).Msg("area_verde: error al listar fichas")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "no se pudo leer el catastro"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"areas": rows})
}

// Actualizar godoc
// @Summary Update green area metadata ficha
// @Tags catastro
// @Accept json
// @Produce json
// @Param id path string true "Feature ID"
// @Param body body requests.ActualizarFichaRequest true "Updated ficha"
// @Success 200 {object} dto.FichaDTO
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /v1/catastro/areas/{id} [patch]
func (ctrl *areaVerdeController) Actualizar(c *gin.Context) {
	var body requests.ActualizarFichaRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "JSON inválido"})
		return
	}

	var userID *int64
	if u, ok := middleware.UsuarioEn(c); ok && u.ID > 0 {
		userID = &u.ID
	}

	item, err := ctrl.uc.ActualizarFicha(c.Request.Context(), c.Param("id"), dto.ActualizarFichaDTO{
		Nombre:     body.Nombre,
		Uso:        body.Uso,
		RiegoAct:   body.RiegoAct,
		Referencia: body.Referencia,
	}, userID)
	if err != nil {
		if errors.Is(err, domainErrors.ErrFichaNoEncontrada) {
			c.JSON(http.StatusNotFound, gin.H{"error": "área no encontrada"})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": "no se pudo guardar la ficha"})
		return
	}
	c.JSON(http.StatusOK, item)
}

// Crear godoc
// @Summary Create green area without GPS geometry
// @Tags catastro
// @Accept json
// @Produce json
// @Param body body requests.CrearAreaSinGeomRequest true "New area"
// @Success 201 {object} dto.FichaDTO
// @Failure 400 {object} map[string]string
// @Router /v1/catastro/areas [post]
func (ctrl *areaVerdeController) Crear(c *gin.Context) {
	var body requests.CrearAreaSinGeomRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "JSON inválido"})
		return
	}

	var userID *int64
	if u, ok := middleware.UsuarioEn(c); ok && u.ID > 0 {
		userID = &u.ID
	}

	item, err := ctrl.uc.CrearSinGeom(c.Request.Context(), dto.CrearAreaSinGeomDTO{
		FeatureID: body.FeatureID,
		Nombre:    body.Nombre,
		Uso:       body.Uso,
	}, userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no se pudo crear el área sin geometría"})
		return
	}
	c.JSON(http.StatusCreated, item)
}

// Baja godoc
// @Summary Logical deactivation of a green area
// @Tags catastro
// @Produce json
// @Param id path string true "Feature ID"
// @Success 200 {object} map[string]any
// @Failure 404 {object} map[string]string
// @Router /v1/catastro/areas/{id}/baja [post]
func (ctrl *areaVerdeController) Baja(c *gin.Context) {
	var userID *int64
	if u, ok := middleware.UsuarioEn(c); ok && u.ID > 0 {
		userID = &u.ID
	}
	featureID := c.Param("id")
	if err := ctrl.uc.Baja(c.Request.Context(), featureID, userID); err != nil {
		if errors.Is(err, domainErrors.ErrFichaNoEncontrada) {
			c.JSON(http.StatusNotFound, gin.H{"error": "área no encontrada"})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": "no se pudo dar de baja el área"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"feature_id": featureID, "activo": false})
}
