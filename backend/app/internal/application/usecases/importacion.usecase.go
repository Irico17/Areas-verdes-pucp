// Package usecases contains application business logic orchestrators.
package usecases

import (
	"context"
	"fmt"

	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/contracts"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/dto"
	apperrors "github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/errors"
)

type importacionUseCase struct {
	lector contracts.ILectorArchivo
	write  contracts.ICargaWriteRepository
}

// NewImportacionUseCase creates a new IImportacionUseCase instance.
func NewImportacionUseCase(
	lector contracts.ILectorArchivo,
	write contracts.ICargaWriteRepository,
) contracts.IImportacionUseCase {
	return &importacionUseCase{
		lector: lector,
		write:  write,
	}
}

// Entidades returns the list of importable entity names.
func (u *importacionUseCase) Entidades(ctx context.Context) []string {
	if u.lector != nil {
		return u.lector.EntidadesImportables()
	}
	return nil
}

// Previsualizar parses and validates an uploaded file, saving a batch preview.
func (u *importacionUseCase) Previsualizar(ctx context.Context, entidad, nombre string, body []byte, usuarioID int64) (*dto.VistaPreviaResponseDTO, error) {
	vista, err := u.lector.Previsualizar(entidad, nombre, body)
	if err != nil {
		return nil, err
	}

	id, err := u.write.GuardarVistaPrevia(ctx, entidad, usuarioID, vista.Validas, body, nombre)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", apperrors.ErrGuardarVistaPrevia, err)
	}

	return &dto.VistaPreviaResponseDTO{
		ID:               id,
		LoteID:           id,
		Entidad:          vista.Entidad,
		Formato:          vista.Formato,
		Validas:          vista.Validas,
		Errores:          vista.Errores,
		Filas:            vista.Filas,
		ColumnasOmitidas: vista.ColumnasOmitidas,
		AvisoOmitidas:    vista.AvisoOmitidas,
		Avisos:           vista.Avisos,
		Escrito:          false,
	}, nil
}

// Confirmar confirms and writes an import batch to the database.
func (u *importacionUseCase) Confirmar(ctx context.Context, loteID, usuarioID int64) (*dto.ConfirmarImportacionResponseDTO, error) {
	confirmadoID, validas, err := u.write.Confirmar(ctx, loteID, usuarioID)
	if err != nil {
		return nil, err
	}

	return &dto.ConfirmarImportacionResponseDTO{
		LoteID:  confirmadoID,
		Validas: validas,
		Escrito: true,
	}, nil
}
