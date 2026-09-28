// Package requests defines request binding and parsing structs and helpers.
package requests

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/dto"
	domainErrors "github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/errors"
)

// ParseFiltroGeo extracts and validates bbox and limit query parameters.
func ParseFiltroGeo(c *gin.Context) (dto.FiltroGeoDTO, error) {
	bbox, err := dto.ParseBBox(c.Query("bbox"))
	if err != nil {
		return dto.FiltroGeoDTO{}, err
	}
	limit, err := parseLimit(c.Query("limit"))
	if err != nil {
		return dto.FiltroGeoDTO{}, err
	}
	return dto.FiltroGeoDTO{BBox: bbox, Limit: limit}, nil
}

func parseLimit(raw string) (int, error) {
	if raw == "" {
		return 0, nil
	}
	n, err := strconv.Atoi(raw)
	if err != nil || n < 1 || n > 10000 {
		return 0, domainErrors.ErrLimit
	}
	return n, nil
}
