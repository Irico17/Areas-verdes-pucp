package controller

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// HealthResponse is returned by the health endpoint.
type HealthResponse struct {
	Status    string `json:"status"`
	Timestamp string `json:"timestamp"`
}

// IHealthController defines health-check operations.
type IHealthController interface {
	Health(*gin.Context)
}

type healthController struct{}

// NewHealthController creates a health controller.
func NewHealthController() IHealthController { return &healthController{} }

// Health godoc
// @Summary Health check
// @Description Verify the health state of the API service
// @Tags health
// @Produce json
// @Success 200 {object} HealthResponse
// @Router /v1/health [get]
func (*healthController) Health(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, HealthResponse{
		Status: "healthy", Timestamp: time.Now().UTC().Format(time.RFC3339),
	})
}
