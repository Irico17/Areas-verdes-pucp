// Package controller implements HTTP request handlers for the presentation layer.
package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/contracts"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/presentation/requests"
)

// IIAController defines HTTP handlers for IA-related endpoints.
type IIAController interface {
	Sugerir(*gin.Context)
}

type iaController struct {
	uc contracts.IIAUseCase
}

// NewIAController creates a new IIAController instance.
func NewIAController(uc contracts.IIAUseCase) IIAController {
	return &iaController{uc: uc}
}

// Sugerir godoc
// @Summary Suggest activity type based on title keywords
// @Tags ia
// @Accept json
// @Produce json
// @Param request body requests.SugerirTipoRequest true "Title payload"
// @Success 200 {object} dto.SugerenciaIADTO
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Router /v1/ia/sugerir-tipo [post]
func (ctrl *iaController) Sugerir(c *gin.Context) {
	var body requests.SugerirTipoRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "JSON inválido"})
		return
	}

	res := ctrl.uc.Sugerir(c.Request.Context(), body.Titulo)
	c.JSON(http.StatusOK, res)
}
