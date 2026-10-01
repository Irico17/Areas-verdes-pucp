// Package controller implements HTTP request handlers for the presentation layer.
package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"

	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/contracts"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/presentation/requests"
)

// IViveroController defines HTTP endpoints for nursery activity records.
type IViveroController interface {
	List(*gin.Context)
	Create(*gin.Context)
	Edit(*gin.Context)
	Archive(*gin.Context)
}

type viveroController struct {
	uc     contracts.IViveroUseCase
	logger zerolog.Logger
}

// NewViveroController creates a new IViveroController instance.
func NewViveroController(
	uc contracts.IViveroUseCase,
	logger zerolog.Logger,
) IViveroController {
	return &viveroController{
		uc:     uc,
		logger: logger,
	}
}

// List godoc
// @Summary List nursery activity records
// @Tags vivero
// @Produce json
// @Param mes query string false "Filter by month (YYYY-MM)"
// @Success 200 {object} dto.ViveroResponseDTO
// @Router /v1/vivero [get]
func (ctrl *viveroController) List(c *gin.Context) {
	mes := c.Query("mes")
	res, err := ctrl.uc.Listar(c.Request.Context(), mes)
	if err != nil {
		writeOperacionErr(c, &ctrl.logger, err)
		return
	}
	c.JSON(http.StatusOK, res)
}

// Create godoc
// @Summary Create a nursery activity record
// @Tags vivero
// @Accept json
// @Produce json
// @Param request body requests.GuardarViveroRequest true "Vivero payload"
// @Success 201 {object} dto.ViveroDTO
// @Router /v1/vivero [post]
func (ctrl *viveroController) Create(c *gin.Context) {
	var req requests.GuardarViveroRequest
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
// @Summary Edit a nursery activity record
// @Tags vivero
// @Accept json
// @Produce json
// @Param id path string true "Vivero UUID"
// @Param request body requests.GuardarViveroRequest true "Vivero edit payload"
// @Success 200 {object} dto.ViveroDTO
// @Router /v1/vivero/{id} [patch]
func (ctrl *viveroController) Edit(c *gin.Context) {
	var req requests.GuardarViveroRequest
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
// @Summary Archive a nursery activity record
// @Tags vivero
// @Produce json
// @Param id path string true "Vivero UUID"
// @Success 200 {object} map[string]bool
// @Router /v1/vivero/{id}/archivar [post]
func (ctrl *viveroController) Archive(c *gin.Context) {
	id := c.Param("id")
	if err := ctrl.uc.Archivar(c.Request.Context(), id); err != nil {
		writeOperacionErr(c, &ctrl.logger, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"ok": true})
}
