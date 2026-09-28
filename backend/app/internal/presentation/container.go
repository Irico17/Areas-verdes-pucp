// Package presentation contains the HTTP delivery layer.
package presentation

import (
	"go.uber.org/dig"

	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/presentation/controller"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/presentation/routes"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/presentation/routes/groups"
)

// RegisterContainer registers presentation-layer dependencies.
func RegisterContainer(container *dig.Container) error {
	// Router
	if err := container.Provide(routes.NewRouter); err != nil {
		return err
	}

	// Groups
	for _, constructor := range []any{
		groups.NewHealthGroup,
		groups.NewSwaggerGroup,
		groups.NewMetaGroup,
		groups.NewLegadoGroup,
		groups.NewSesionGroup,
		groups.NewAccesosGroup,
	} {
		if err := container.Provide(constructor); err != nil {
			return err
		}
	}

	// Controllers
	for _, constructor := range []any{
		controller.NewHealthController,
		controller.NewMetaController,
		controller.NewSesionController,
		controller.NewUsuarioController,
	} {
		if err := container.Provide(constructor); err != nil {
			return err
		}
	}

	return nil
}
