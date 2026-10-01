// Package usecases contains application workflow logic.
package usecases

import (
	"context"

	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/contracts"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/dto"
)

type iaUseCase struct {
	sugeridor contracts.ISugeridorTipo
}

// NewIAUseCase creates a new IIAUseCase instance.
func NewIAUseCase(sugeridor contracts.ISugeridorTipo) contracts.IIAUseCase {
	return &iaUseCase{sugeridor: sugeridor}
}

// Sugerir suggests an activity type based on title keywords.
func (uc *iaUseCase) Sugerir(_ context.Context, titulo string) dto.SugerenciaIADTO {
	s := uc.sugeridor.SugerirTipo(titulo)
	return dto.SugerenciaIADTO{
		Codigo:         s.Codigo,
		Etiqueta:       s.Etiqueta,
		Explicacion:    s.Explicacion,
		Confianza:      s.Confianza,
		RequiereHumano: s.RequiereHumano,
	}
}
