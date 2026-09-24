package handlers

import (
	"context"
	"time"

	"campusverde/api/internal/catastro"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// Health informa si el proceso, Postgres y PostGIS responden.
type Health struct {
	DB    *gorm.DB
	Store *catastro.Store
}

func (h Health) Get(c *gin.Context) {
	body := gin.H{
		"status":  "ok",
		"service": "campus-verde-api",
	}
	if h.DB == nil || h.Store == nil {
		body["status"] = "error"
		body["database"] = "down"
		c.JSON(503, body)
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 3*time.Second)
	defer cancel()

	version, err := h.Store.PostGISVersion(ctx)
	if err != nil {
		body["status"] = "error"
		body["database"] = "down"
		c.JSON(503, body)
		return
	}
	body["database"] = "up"
	body["postgis"] = version
	c.JSON(200, body)
}
