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
		database.NewTransaccion,
		postgres.NewSaludRepository,
		postgres.NewUsuarioRepository,
		postgres.NewSesionRepository,
		postgres.NewPermisoRepository,
		postgres.NewCatalogoRepository,
		postgres.NewCambioRepository,
		postgres.NewGeoRepository,
		postgres.NewAreaVerdeRepository,
		postgres.NewZonaSupervisionRepository,
		postgres.NewCuadrillaRepository,
		postgres.NewLugarRepository,
		postgres.NewEspecieRepository,
		postgres.NewEjemplarRepository,
		postgres.NewCatastroReferenciaRepository,
		postgres.NewInventarioRepository,
	}

	for _, provider := range providers {
		if err := container.Provide(provider); err != nil {
			return err
		}
	}
	return nil
}
