// Package contracts defines application interfaces.
package contracts

import "context"

// ISaludRepository provides health check operations against PostgreSQL/PostGIS.
type ISaludRepository interface {
	PostGISVersion(ctx context.Context) (string, error)
}
