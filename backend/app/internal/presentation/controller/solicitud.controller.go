// Package controller implements HTTP request handlers for the presentation layer.
package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"

	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/contracts"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/presentation/requests"
)

// ISolicitudController defines HTTP endpoints for solicitudes.
type ISolicitudController interface {
	List(*gin.Context)
	Create(*gin.Context)
	Edit(*gin.Context)
}

type solicitudController struct {
	uc     contracts.ISolicitudUseCase
	logger zerolog.Logger
}

// NewSolicitudController creates a new ISolicitudController instance.
func NewSolicitudController(
	uc contracts.ISolicitudUseCase,
	logger zerolog.Logger,
) ISolicitudController {
	return &solicitudController{
		uc:     uc,
		logger: logger,
	}
}

// List godoc
// @Summary List service requests
// @Tags atencion
// @Produce json
// @Success 200 {object} dto.SolicitudesResponseDTO
// @Router /v1/solicitudes [get]
func (ctrl *solicitudController) List(c *gin.Context) {
	res, err := ctrl.uc.Listar(c.Request.Context())
	if err != nil {
		ctrl.logger.Error().Err(err).Msg("solicitudes")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "no se pudieron leer las solicitudes"})
		return
	}
	c.JSON(http.StatusOK, res)
}

// Create godoc
// @Summary Create a service request
// @Tags atencion
// @Accept json
// @Produce json
// @Param request body requests.CrearSolicitudRequest true "Solicitud payload"
// @Success 201 {object} dto.SolicitudDTO
// @Router /v1/solicitudes [post]
func (ctrl *solicitudController) Create(c *gin.Context) {
	var req requests.CrearSolicitudRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "JSON inválido"})
		return
	}

	res, err := ctrl.uc.Crear(c.Request.Context(), req.ToDTO())
	if err != nil {
		writeOperacionErr(c, &ctrl.logger, err)
		return
	}

	c.JSON(http.StatusCreated, res)
}

// Edit godoc
// @Summary Edit a service request
// @Tags atencion
// @Accept json
// @Produce json
// @Param id path string true "Solicitud UUID"
// @Param request body requests.EditarSolicitudRequest true "Solicitud edit payload"
// @Success 200 {object} dto.SolicitudDTO
// @Router /v1/solicitudes/{id} [patch]
func (ctrl *solicitudController) Edit(c *gin.Context) {
	var req requests.EditarSolicitudRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "JSON inválido"})
		return
	}

	id := c.Param("id")
	res, err := ctrl.uc.Editar(c.Request.Context(), req.ToDTO(id))
	if err != nil {
		writeOperacionErr(c, &ctrl.logger, err)
		return
	}

	c.JSON(http.StatusOK, res)
}
