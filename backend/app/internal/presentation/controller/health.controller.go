package controller

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/usecases"
)

// HealthResponse is returned by the health endpoint.
type HealthResponse struct {
	Status    string `json:"status"`
	Timestamp string `json:"timestamp"`
	Database  string `json:"database,omitempty"`
	PostGIS   string `json:"postgis,omitempty"`
}

// IHealthController defines health-check operations.
type IHealthController interface {
	Health(*gin.Context)
	LegacyHealth(*gin.Context)
}

type healthController struct {
	saludUseCase usecases.ISaludUseCase
}

// NewHealthController creates a health controller.
func NewHealthController(saludUseCase usecases.ISaludUseCase) IHealthController {
	return &healthController{saludUseCase: saludUseCase}
}

// Health godoc
// @Summary Health check
// @Description Verify the health state of the API service
// @Tags health
// @Produce json
// @Success 200 {object} HealthResponse
// @Router /v1/health [get]
func (c *healthController) Health(ctx *gin.Context) {
	estado, err := c.saludUseCase.VerificarSalud(ctx.Request.Context())
	timestamp := time.Now().UTC().Format(time.RFC3339)

	if err != nil || estado == nil || estado.Database != "up" {
		ctx.JSON(http.StatusServiceUnavailable, HealthResponse{
			Status:    "unhealthy",
			Timestamp: timestamp,
			Database:  "down",
		})
		return
	}

	ctx.JSON(http.StatusOK, HealthResponse{
		Status:    "healthy",
		Timestamp: timestamp,
		Database:  estado.Database,
		PostGIS:   estado.PostGIS,
	})
}

// LegacyHealth handles GET /health on the root router, matching the old API response exactly.
func (c *healthController) LegacyHealth(ctx *gin.Context) {
	body := gin.H{
		"status":  "ok",
		"service": "campus-verde-api",
	}

	estado, err := c.saludUseCase.VerificarSalud(ctx.Request.Context())
	if err != nil || estado == nil || estado.Database != "up" {
		body["status"] = "error"
		body["database"] = "down"
		ctx.JSON(http.StatusServiceUnavailable, body)
		return
	}

	body["database"] = "up"
	body["postgis"] = estado.PostGIS
	ctx.JSON(http.StatusOK, body)
}
