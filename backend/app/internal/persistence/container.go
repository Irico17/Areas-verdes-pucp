// Package persistence wires database and repository dependencies.
package persistence

import (
	"go.uber.org/dig"

	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/persistence/database"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/persistence/repository/postgres"
)

// RegisterContainer registers lazy persistence-layer providers.
func RegisterContainer(container *dig.Container) error {
	if err := container.Provide(database.NewConnection); err != nil {
		return err
	}
	return container.Provide(postgres.NewSaludRepository)
}
