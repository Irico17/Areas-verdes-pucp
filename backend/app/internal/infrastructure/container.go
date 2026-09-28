// Package infrastructure contains adapters for external services.
package infrastructure

import (
	"time"

	"go.uber.org/dig"

	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/contracts"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/infrastructure/ratelimit"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/shared/config"
)

// RegisterContainer registers infrastructure-layer dependencies.
func RegisterContainer(container *dig.Container) error {
	return container.Provide(func(cfg *config.Config) contracts.ILimitador {
		max := cfg.Seguridad.LoginMax
		if max < 1 {
			max = 8
		}
		ventana := cfg.Seguridad.LoginVentana
		if ventana <= 0 {
			ventana = time.Minute
		}
		return ratelimit.NewMemoriaLimitador(max, ventana)
	})
}
