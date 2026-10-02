// Package controller implements HTTP request handlers for the presentation layer.
package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"

	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/contracts"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/presentation/middleware"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/presentation/requests"
)

// IServicioTercerizadoController defines HTTP endpoints for work orders (ordenes_servicio).
type IServicioTercerizadoController interface {
	List(*gin.Context)
	Create(*gin.Context)
	Edit(*gin.Context)
	Evidencias(*gin.Context)
}

type servicioTercerizadoController struct {
	uc     contracts.IServicioTercerizadoUseCase
	logger zerolog.Logger
}

// NewServicioTercerizadoController creates a new IServicioTercerizadoController instance.
func NewServicioTercerizadoController(
	uc contracts.IServicioTercerizadoUseCase,
	logger zerolog.Logger,
) IServicioTercerizadoController {
	return &servicioTercerizadoController{
		uc:     uc,
		logger: logger,
	}
}

// List godoc
// @Summary List work orders
// @Tags atencion
// @Produce json
// @Success 200 {object} dto.OrdenesResponseDTO
// @Router /v1/ordenes [get]
func (ctrl *servicioTercerizadoController) List(c *gin.Context) {
	res, err := ctrl.uc.Listar(c.Request.Context())
	if err != nil {
		ctrl.logger.Error().Err(err).Msg("ordenes")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "no se pudieron leer las órdenes"})
		return
	}
	c.JSON(http.StatusOK, res)
}

// Create godoc
// @Summary Create a work order
// @Tags atencion
// @Accept json
// @Produce json
// @Param request body requests.CrearOrdenRequest true "Work order payload"
// @Success 201 {object} dto.OrdenDTO
// @Router /v1/ordenes [post]
func (ctrl *servicioTercerizadoController) Create(c *gin.Context) {
	u, ok := middleware.UsuarioEn(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "inicie sesión"})
		return
	}

	var req requests.CrearOrdenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "JSON inválido"})
		return
	}

	res, err := ctrl.uc.Crear(c.Request.Context(), req.ToDTO(), u.Rol, u.CapatazID)
	if err != nil {
		writeOperacionErr(c, &ctrl.logger, err)
		return
	}

	c.JSON(http.StatusCreated, res)
}

// Evidencias godoc
// @Summary Evidence already linked to a work order
// @Tags atencion
// @Produce json
// @Param id path string true "Work order UUID"
// @Success 200 {object} dto.EvidenciasOrdenDTO
// @Router /v1/ordenes/{id}/evidencias [get]
func (ctrl *servicioTercerizadoController) Evidencias(c *gin.Context) {
	res, err := ctrl.uc.Evidencias(c.Request.Context(), c.Param("id"))
	if err != nil {
		writeOperacionErr(c, &ctrl.logger, err)
		return
	}
	c.JSON(http.StatusOK, res)
}

// Edit godoc
// @Summary Edit a work order
// @Tags atencion
// @Accept json
// @Produce json
// @Param id path string true "Work order UUID"
// @Param request body requests.EditarOrdenRequest true "Work order edit payload"
// @Success 200 {object} dto.EditarOrdenResponseDTO
// @Router /v1/ordenes/{id} [patch]
func (ctrl *servicioTercerizadoController) Edit(c *gin.Context) {
	u, ok := middleware.UsuarioEn(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "inicie sesión"})
		return
	}

	var req requests.EditarOrdenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "JSON inválido"})
		return
	}

	id := c.Param("id")
	res, err := ctrl.uc.Editar(c.Request.Context(), req.ToDTO(id), u.Rol, u.CapatazID)
	if err != nil {
		writeOperacionErr(c, &ctrl.logger, err)
		return
	}

	c.JSON(http.StatusOK, res)
}
