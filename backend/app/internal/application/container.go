// Package application contains use cases, services, contracts and DTOs.
package application

import (
	"go.uber.org/dig"

	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/services"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/usecases"
)

// RegisterContainer registers application-layer dependencies.
func RegisterContainer(container *dig.Container) error {
	providers := []any{
		usecases.NewSaludUseCase,
		services.NewPermisosService,
		usecases.NewSesionUseCase,
		usecases.NewUsuarioUseCase,
		usecases.NewSemillaAccesosUseCase,
		usecases.NewCatalogoUseCase,
		services.NewAuditoriaService,
		usecases.NewGeoUseCase,
		usecases.NewAreaVerdeUseCase,
		usecases.NewZonaSupervisionUseCase,
		usecases.NewCuadrillaUseCase,
		usecases.NewLugarUseCase,
		usecases.NewEspecieUseCase,
		usecases.NewEjemplarUseCase,
		usecases.NewCatastroReferenciaUseCase,
		usecases.NewInventarioUseCase,
		usecases.NewReservasMockUseCase,
		usecases.NewInventarioCampoUseCase,
		usecases.NewIntervencionUseCase,
		usecases.NewSolicitudUseCase,
		usecases.NewServicioTercerizadoUseCase,
		usecases.NewRiegoUseCase,
		usecases.NewPodaUseCase,
		usecases.NewViveroUseCase,
		services.NewExifService,
		usecases.NewEvidenciaUseCase,
		services.NewSugeridorTipoService,
		usecases.NewReporteUseCase,
		usecases.NewIAUseCase,
		usecases.NewLoteUseCase,
		usecases.NewImportacionUseCase,
		usecases.NewCargaInicialUseCase,
		usecases.NewCargaLoteUseCase,
		usecases.NewSectoresUseCase,
	}

	for _, provider := range providers {
		if err := container.Provide(provider); err != nil {
			return err
		}
	}
	return nil
}
