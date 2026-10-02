package contracts

import (
	"context"

	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/dto"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/entities"
)

// ICambioRepository defines persistence operations on the cambios table.
type ICambioRepository interface {
	Crear(ctx context.Context, cambio *entities.Cambio) error
}

// IAuditoriaService defines high-level change recording services.
type IAuditoriaService interface {
	RegistrarCambio(ctx context.Context, req dto.RegistrarCambioDTO) error
}

// ILoteRepository defines persistence operations for audit logs and batch import/reversion.
type ILoteRepository interface {
	Importar(ctx context.Context, usuarioID int64, entidad string, filas []entities.FilaLote) (int64, error)
	Editar(ctx context.Context, usuarioID int64, entidad, entidadID string, despues []byte) error
	Revertir(ctx context.Context, loteID, usuarioID int64, confirmar bool) (*entities.ReporteReversion, error)
	Timeline(ctx context.Context, f entities.FiltroAuditoria) ([]entities.EventoAuditoria, error)
	Historial(ctx context.Context, f entities.FiltroAuditoria) ([]entities.EventoAuditoria, error)
}

// ILoteUseCase defines application use cases for audit logs and batch operations.
type ILoteUseCase interface {
	Importar(ctx context.Context, usuarioID int64, req dto.ImportarLoteDTO) (*dto.ImportarLoteResponseDTO, error)
	Revertir(ctx context.Context, loteID, usuarioID int64, confirmar bool) (*dto.ReporteReversionDTO, error)
	Editar(ctx context.Context, usuarioID int64, req dto.EditarAuditoriaDTO) (*dto.EditarAuditoriaResponseDTO, error)
	Timeline(ctx context.Context, f dto.FiltroAuditoriaDTO) ([]dto.EventoAuditoriaDTO, error)
	Historial(ctx context.Context, f dto.FiltroAuditoriaDTO) ([]dto.EventoAuditoriaDTO, error)
}
