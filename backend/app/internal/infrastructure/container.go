// Package infrastructure contains adapters for external services.
package infrastructure

import (
	"go.uber.org/dig"

	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/infrastructure/archivos"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/infrastructure/ratelimit"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/infrastructure/seguridad"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/infrastructure/storage"
)

// RegisterContainer registers infrastructure-layer dependencies.
func RegisterContainer(container *dig.Container) error {
	for _, provider := range []any{
		ratelimit.NewLimitadorFromConfig,
		seguridad.NewBcryptHasher,
		archivos.NewContratoOpenAPIAdapter,
		archivos.NewGeoJSONEstaticoAdapter,
		archivos.NewFotoDiscoAdapter,
		archivos.NewReservasMockAdapter,
		archivos.NewPuntosParser,
		storage.NewAlmacenArchivos,
	} {
		if err := container.Provide(provider); err != nil {
			return err
		}
	}
	return nil
}
