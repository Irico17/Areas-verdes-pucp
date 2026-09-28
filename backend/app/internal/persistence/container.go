// Package persistence wires database and repository dependencies.
package persistence

import (
	"go.uber.org/dig"

	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/persistence/database"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/persistence/repository/postgres"
)

// RegisterContainer registers lazy persistence-layer providers.
func RegisterContainer(container *dig.Container) error {
	providers := []any{
		database.NewConnection,
		postgres.NewSaludRepository,
		postgres.NewUsuarioRepository,
		postgres.NewSesionRepository,
		postgres.NewPermisoRepository,
	}

	for _, provider := range providers {
		if err := container.Provide(provider); err != nil {
			return err
		}
	}
	return nil
}
