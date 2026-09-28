// Package contracts defines interfaces implemented across layers.
package contracts

import (
	"context"

	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/dto"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/entities"
)

// IPuntosParser parses puntos PUCP CSV content and detects excluded contact columns.
type IPuntosParser interface {
	LeerPuntosPUCP(body []byte) ([]entities.PuntoCarga, []entities.RechazoFormato, []string, error)
}

// IInventarioCampoRepository defines persistence operations for frente 2B field inventory.
type IInventarioCampoRepository interface {
	ListarTachos(ctx context.Context) ([]entities.Tacho, error)
	GuardarTacho(ctx context.Context, t entities.Tacho) (*entities.Tacho, error)
	ActualizarTacho(ctx context.Context, id int64, t entities.Tacho, rawJSON []byte) (*entities.Tacho, error)

	ListarBebederos(ctx context.Context) ([]entities.Bebedero, error)
	GuardarBebedero(ctx context.Context, b entities.Bebedero) (*entities.Bebedero, error)
	ActualizarBebedero(ctx context.Context, id int64, b entities.Bebedero, rawJSON []byte) (*entities.Bebedero, error)

	ListarPuntos(ctx context.Context, q string) ([]entities.PuntoPUCP, error)
	GuardarPunto(ctx context.Context, p entities.PuntoPUCP) (*entities.PuntoPUCP, error)
	ActualizarPunto(ctx context.Context, id int64, p entities.PuntoPUCP, rawJSON []byte) (*entities.PuntoPUCP, error)

	ListarReservas(ctx context.Context, desde, hasta string) ([]entities.ReservaJardin, error)
	GuardarReserva(ctx context.Context, r entities.ReservaJardin) (*entities.ReservaJardin, error)
	ActualizarReserva(ctx context.Context, id int64, r entities.ReservaJardin, rawJSON []byte) (*entities.ReservaJardin, error)

	ListarFichas(ctx context.Context, capa string) ([]entities.FichaCapa, error)
	GuardarFicha(ctx context.Context, capa string, f entities.FichaCapa) (*entities.FichaCapa, error)
	ActualizarFicha(ctx context.Context, capa string, id int64, f entities.FichaCapa, rawJSON []byte) (*entities.FichaCapa, error)

	Baja(ctx context.Context, tabla string, id int64) error
}

// IInventarioCampoUseCase defines application logic for frente 2B field inventory.
type IInventarioCampoUseCase interface {
	ListarTachos(ctx context.Context) (dto.ListarTachosResponseDTO, error)
	GuardarTacho(ctx context.Context, req dto.TachoDTO) (*dto.TachoDTO, error)
	ActualizarTacho(ctx context.Context, id int64, req dto.TachoDTO, rawJSON []byte) (*dto.TachoDTO, error)
	CSVTachos(ctx context.Context) (string, error)
	BajaTacho(ctx context.Context, id int64) error

	ListarBebederos(ctx context.Context) (dto.ListarBebederosResponseDTO, error)
	GuardarBebedero(ctx context.Context, req dto.BebederoDTO) (*dto.BebederoDTO, error)
	ActualizarBebedero(ctx context.Context, id int64, req dto.BebederoDTO, rawJSON []byte) (*dto.BebederoDTO, error)
	BajaBebedero(ctx context.Context, id int64) error

	ListarPuntos(ctx context.Context, q string) (dto.ListarPuntosResponseDTO, error)
	GuardarPunto(ctx context.Context, req dto.PuntoDTO, rawBody []byte) (*dto.PuntoDTO, error)
	ActualizarPunto(ctx context.Context, id int64, req dto.PuntoDTO, rawJSON []byte) (*dto.PuntoDTO, error)
	BajaPunto(ctx context.Context, id int64) error
	FormatoPuntos(ctx context.Context, body []byte) (dto.FormatoPuntosResponseDTO, error)

	ListarReservas(ctx context.Context, desde, hasta string) (dto.ListarReservasResponseDTO, error)
	GuardarReserva(ctx context.Context, req dto.ReservaDTO) (*dto.ReservaDTO, error)
	ActualizarReserva(ctx context.Context, id int64, req dto.ReservaDTO, rawJSON []byte) (*dto.ReservaDTO, error)
	BajaReserva(ctx context.Context, id int64) error

	ListarCapa(ctx context.Context, capa string) (dto.ListarFichasCapaResponseDTO, error)
	GuardarCapa(ctx context.Context, capa string, req dto.FichaCapaDTO) (*dto.FichaCapaDTO, error)
	ActualizarCapa(ctx context.Context, capa string, id int64, req dto.FichaCapaDTO, rawJSON []byte) (*dto.FichaCapaDTO, error)
	CSVCapa(ctx context.Context, capa string) (string, error)
	BajaCapa(ctx context.Context, capa string, id int64) error
}
