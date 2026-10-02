// Package contracts defines interfaces for repositories, services, and adapters.
package contracts

import (
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/dto"
)

// ILectorArchivo defines file reading, parsing, and preview operations.
type ILectorArchivo interface {
	// EntidadesImportables returns the list of entities that can be imported.
	EntidadesImportables() []string

	// Previsualizar inspects and validates an uploaded file for a given entity without database modifications.
	Previsualizar(entidad, nombre string, body []byte) (dto.VistaPreviaDTO, error)
}
