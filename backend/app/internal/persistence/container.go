// Package persistence wires database and repository dependencies.
package persistence

import (
	"go.uber.org/dig"

	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/persistence/database"
)

// RegisterContainer registers lazy persistence-layer providers.
func RegisterContainer(container *dig.Container) error {
	return container.Provide(database.NewConnection)
}
