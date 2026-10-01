// Package controller implements HTTP request handlers for the presentation layer.
package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"

	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/contracts"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/presentation/requests"
)

// IPodaController defines HTTP endpoints for poda records.
type IPodaController interface {
	List(*gin.Context)
	Create(*gin.Context)
	Edit(*gin.Context)
	Archive(*gin.Context)
}

type podaController struct {
	uc     contracts.IPodaUseCase
	logger zerolog.Logger
}

// NewPodaController creates a new IPodaController instance.
func NewPodaController(
	uc contracts.IPodaUseCase,
	logger zerolog.Logger,
) IPodaController {
	return &podaController{
		uc:     uc,
		logger: logger,
	}
}

// List godoc
// @Summary List poda records
// @Tags poda
// @Produce json
// @Success 200 {object} dto.PodasResponseDTO
// @Router /v1/podas [get]
func (ctrl *podaController) List(c *gin.Context) {
	res, err := ctrl.uc.Listar(c.Request.Context())
	if err != nil {
		ctrl.logger.Error().Err(err).Msg("podas")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "no se pudieron leer las podas"})
		return
	}
	c.JSON(http.StatusOK, res)
}

// Create godoc
// @Summary Create a poda record
// @Tags poda
// @Accept json
// @Produce json
// @Param request body requests.GuardarPodaRequest true "Poda payload"
// @Success 201 {object} dto.PodaDTO
// @Router /v1/podas [post]
func (ctrl *podaController) Create(c *gin.Context) {
	var req requests.GuardarPodaRequest
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
// @Summary Edit a poda record
// @Tags poda
// @Accept json
// @Produce json
// @Param id path string true "Poda UUID"
// @Param request body requests.GuardarPodaRequest true "Poda edit payload"
// @Success 200 {object} dto.PodaDTO
// @Router /v1/podas/{id} [patch]
func (ctrl *podaController) Edit(c *gin.Context) {
	var req requests.GuardarPodaRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "JSON inválido"})
		return
	}

	id := c.Param("id")
	res, err := ctrl.uc.Editar(c.Request.Context(), req.ToDTOWithID(id))
	if err != nil {
		writeOperacionErr(c, &ctrl.logger, err)
		return
	}

	c.JSON(http.StatusOK, res)
}

// Archive godoc
// @Summary Archive a poda record
// @Tags poda
// @Produce json
// @Param id path string true "Poda UUID"
// @Success 200 {object} map[string]bool
// @Router /v1/podas/{id}/archivar [post]
func (ctrl *podaController) Archive(c *gin.Context) {
	id := c.Param("id")
	if err := ctrl.uc.Archivar(c.Request.Context(), id); err != nil {
		writeOperacionErr(c, &ctrl.logger, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"ok": true})
}
