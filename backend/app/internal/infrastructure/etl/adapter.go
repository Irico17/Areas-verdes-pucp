// Package etl implements data extraction, transformation, normalization, and file parsing.
package etl

import (
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/contracts"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/dto"
)

type lectorArchivoAdapter struct{}

// NewLectorArchivoAdapter returns a new contracts.ILectorArchivo implementation.
func NewLectorArchivoAdapter() contracts.ILectorArchivo {
	return &lectorArchivoAdapter{}
}

func (a *lectorArchivoAdapter) EntidadesImportables() []string {
	return EntidadesImportables
}

func (a *lectorArchivoAdapter) Previsualizar(entidad, nombre string, body []byte) (dto.VistaPreviaDTO, error) {
	vista, err := Previsualizar(entidad, nombre, body)
	if err != nil {
		return dto.VistaPreviaDTO{}, err
	}
	erroresDTO := make([]dto.ErrorFilaDTO, len(vista.Errores))
	for i, e := range vista.Errores {
		erroresDTO[i] = dto.ErrorFilaDTO{
			Fila:   e.Fila,
			Campo:  e.Campo,
			Motivo: e.Motivo,
		}
	}
	return dto.VistaPreviaDTO{
		Entidad:          vista.Entidad,
		Formato:          vista.Formato,
		Validas:          vista.Validas,
		Errores:          erroresDTO,
		Filas:            vista.Filas,
		ColumnasOmitidas: vista.ColumnasOmitidas,
		AvisoOmitidas:    vista.AvisoOmitidas,
		Avisos:           vista.Avisos,
		Escrito:          vista.Escrito,
	}, nil
}
