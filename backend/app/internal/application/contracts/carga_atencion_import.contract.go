// Package contracts defines interfaces for repositories, services, and adapters.
package contracts

import (
	"context"
)

// ICargaAtencionImportRepository defines database operations for places mapping and attention records.
type ICargaAtencionImportRepository interface {
	// MapaLugares returns a map of normalized place name to place ID for active places.
	MapaLugares(ctx context.Context) (map[string]string, error)
}
