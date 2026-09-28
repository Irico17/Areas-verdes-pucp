// Package application contains use cases, services, contracts and DTOs.
package application

import (
	"go.uber.org/dig"

	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/usecases"
)

// RegisterContainer registers application-layer dependencies.
func RegisterContainer(container *dig.Container) error {
	return container.Provide(usecases.NewSaludUseCase)
}
