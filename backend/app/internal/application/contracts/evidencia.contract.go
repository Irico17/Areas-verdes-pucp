// Package contracts defines interfaces implemented across layers.
package contracts

import (
	"context"
	"io"

	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/dto"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/entities"
)

// IEvidenciaRepository defines persistence operations for evidencias.
type IEvidenciaRepository interface {
	Listar(ctx context.Context, actividadID string) ([]entities.Evidencia, error)
	ObtenerRutaYMime(ctx context.Context, id string) (ruta string, mime string, err error)
	Guardar(ctx context.Context, in entities.GuardarEvidencia, storage IAlmacenArchivos) (*entities.ResultadoEvidencia, error)
}

// IEvidenciaUseCase defines usecase operations for evidencias.
type IEvidenciaUseCase interface {
	Listar(ctx context.Context, actividadID string) (*dto.ListarEvidenciasResponseDTO, error)
	Subir(ctx context.Context, in dto.SubirEvidenciaDTO) (*dto.SubirEvidenciaResponseDTO, error)
	Abrir(ctx context.Context, id string) (io.ReadCloser, string, error)
}
