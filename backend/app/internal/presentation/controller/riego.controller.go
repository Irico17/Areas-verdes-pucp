// Package controller implements HTTP request handlers for the presentation layer.
package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"

	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/contracts"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/constants/enums"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/presentation/middleware"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/presentation/requests"
)

// IRiegoController defines HTTP endpoints for irrigation logs.
type IRiegoController interface {
	List(*gin.Context)
	Create(*gin.Context)
}

type riegoController struct {
	uc     contracts.IRiegoUseCase
	logger zerolog.Logger
}

// NewRiegoController creates a new IRiegoController instance.
func NewRiegoController(
	uc contracts.IRiegoUseCase,
	logger zerolog.Logger,
) IRiegoController {
	return &riegoController{
		uc:     uc,
		logger: logger,
	}
}

// List godoc
// @Summary List irrigation shift logs
// @Tags atencion
// @Produce json
// @Success 200 {object} dto.RiegoResponseDTO
// @Router /v1/riego [get]
func (ctrl *riegoController) List(c *gin.Context) {
	u, ok := middleware.UsuarioEn(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "inicie sesión"})
		return
	}

	scope := ""
	if u.Rol == string(enums.RolCapataz) {
		scope = u.CapatazID
	}

	res, err := ctrl.uc.Listar(c.Request.Context(), scope)
	if err != nil {
		ctrl.logger.Error().Err(err).Msg("riego")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "no se pudo leer el riego"})
		return
	}

	c.JSON(http.StatusOK, res)
}

// Create godoc
// @Summary Create an irrigation log entry
// @Tags atencion
// @Accept json
// @Produce json
// @Param request body requests.CrearRiegoRequest true "Irrigation payload"
// @Success 201 {object} dto.CrearRiegoResponseDTO
// @Router /v1/riego [post]
func (ctrl *riegoController) Create(c *gin.Context) {
	u, ok := middleware.UsuarioEn(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "inicie sesión"})
		return
	}

	var req requests.CrearRiegoRequest
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
