// Package controller contains HTTP handlers for presenting domain data.
package controller

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"

	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/contracts"
	domainErrors "github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/errors"
)

// IReservasMockController defines HTTP handlers for the mock reservations agenda.
type IReservasMockController interface {
	Get(*gin.Context)
}

type reservasMockController struct {
	uc     contracts.IReservasMockUseCase
	logger zerolog.Logger
}

// NewReservasMockController creates a new ReservasMockController.
func NewReservasMockController(uc contracts.IReservasMockUseCase, logger zerolog.Logger) IReservasMockController {
	return &reservasMockController{uc: uc, logger: logger}
}

// Get godoc
// @Summary Mock reservations schedule without external spreadsheet references
// @Tags geo
// @Produce json
// @Success 200 {object} dto.ReservasMockResponseDTO
// @Failure 500 {object} map[string]string
// @Router /v1/geo/reservas-mock [get]
func (ctrl *reservasMockController) Get(c *gin.Context) {
	_, encoded, err := ctrl.uc.ObtenerAgenda(c.Request.Context())
	if errors.Is(err, domainErrors.ErrAgendaFicticiaReferenciaExterna) {
		ctrl.logger.Error().Msg("agenda ficticia contiene referencia externa a Google Sheets")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "la agenda ficticia todavía arrastra una referencia externa"})
		return
	}
	if err != nil {
		ctrl.logger.Error().Err(err).Msg("error leyendo agenda ficticia")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "no se pudo leer la agenda ficticia"})
		return
	}
	c.Data(http.StatusOK, "application/json; charset=utf-8", encoded)
}
