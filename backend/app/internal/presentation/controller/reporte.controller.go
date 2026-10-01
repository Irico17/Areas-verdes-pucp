// Package controller implements HTTP request handlers for the presentation layer.
package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"

	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/contracts"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/dto"
)

// IReporteController defines HTTP handlers for reporting endpoints.
type IReporteController interface {
	ReporteLabores(*gin.Context)
}

type reporteController struct {
	uc     contracts.IReporteUseCase
	logger zerolog.Logger
}

// NewReporteController creates a new IReporteController instance.
func NewReporteController(
	uc contracts.IReporteUseCase,
	logger zerolog.Logger,
) IReporteController {
	return &reporteController{
		uc:     uc,
		logger: logger,
	}
}

// ReporteLabores godoc
// @Summary Generate or export activity reports
// @Tags reportes
// @Produce json
// @Produce text/csv
// @Produce application/vnd.ms-excel
// @Param estado query string false "Filter by status"
// @Param desde query string false "Start date (YYYY-MM-DD)"
// @Param hasta query string false "End date (YYYY-MM-DD)"
// @Param zona query string false "Filter by zone"
// @Param cuadrilla query string false "Filter by cuadrilla"
// @Param origen query string false "Filter by origin"
// @Param formato query string false "Export format (csv or xls)"
// @Success 200 {object} dto.ReporteResponseDTO
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Router /v1/reportes/labores [get]
func (ctrl *reporteController) ReporteLabores(c *gin.Context) {
	filtro := dto.FiltroReporteDTO{
		Estado:    c.Query("estado"),
		Desde:     c.Query("desde"),
		Hasta:     c.Query("hasta"),
		Zona:      c.Query("zona"),
		Cuadrilla: c.Query("cuadrilla"),
		Origen:    c.Query("origen"),
	}

	switch c.Query("formato") {
	case "csv":
		c.Header("Content-Type", "text/csv; charset=utf-8")
		c.Header("Content-Disposition", `attachment; filename="labores.csv"`)
		if err := ctrl.uc.Exportar(c.Request.Context(), filtro, "csv", c.Writer); err != nil {
			writeOperacionErr(c, &ctrl.logger, err)
		}
	case "xls":
		c.Header("Content-Type", "application/vnd.ms-excel")
		c.Header("Content-Disposition", `attachment; filename="labores.xls"`)
		if err := ctrl.uc.Exportar(c.Request.Context(), filtro, "xls", c.Writer); err != nil {
			writeOperacionErr(c, &ctrl.logger, err)
		}
	default:
		rep, err := ctrl.uc.ObtenerReporte(c.Request.Context(), filtro)
		if err != nil {
			writeOperacionErr(c, &ctrl.logger, err)
			return
		}
		c.JSON(http.StatusOK, rep)
	}
}
