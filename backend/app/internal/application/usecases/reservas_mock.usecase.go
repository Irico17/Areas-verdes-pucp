// Package usecases contains application use case implementations.
package usecases

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/contracts"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/dto"
	domainErrors "github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/errors"
)

const avisoAgendaFicticia = "Agenda ficticia de demostración. No está conectada a una hoja de cálculo ni a una fuente institucional."

type reservasMockUseCase struct {
	adapter contracts.IReservasMockAdapter
}

// NewReservasMockUseCase creates a new instance of IReservasMockUseCase.
func NewReservasMockUseCase(adapter contracts.IReservasMockAdapter) contracts.IReservasMockUseCase {
	return &reservasMockUseCase{
		adapter: adapter,
	}
}

// ObtenerAgenda retrieves and sanitizes the mock reservations schedule.
func (u *reservasMockUseCase) ObtenerAgenda(ctx context.Context) (dto.ReservasMockResponseDTO, []byte, error) {
	body, err := u.adapter.LeerReservas(ctx)
	if err != nil {
		resp := dto.ReservasMockResponseDTO{
			Fake:     true,
			Aviso:    avisoAgendaFicticia,
			Total:    0,
			Reservas: []dto.ReservaItemDTO{},
		}
		encoded, _ := json.Marshal(resp)
		return resp, encoded, nil
	}

	var raw struct {
		Reservas []dto.ReservaItemDTO `json:"reservas"`
	}
	if err := json.Unmarshal(body, &raw); err != nil {
		return dto.ReservasMockResponseDTO{}, nil, domainErrors.ErrLeerAgendaFicticia
	}

	for i := range raw.Reservas {
		raw.Reservas[i].Notas = "Dato ficticio de demostración."
		raw.Reservas[i].Fake = true
	}
	if raw.Reservas == nil {
		raw.Reservas = []dto.ReservaItemDTO{}
	}

	resp := dto.ReservasMockResponseDTO{
		Fake:     true,
		Aviso:    avisoAgendaFicticia,
		Total:    len(raw.Reservas),
		Reservas: raw.Reservas,
	}

	encoded, err := json.Marshal(resp)
	if err != nil {
		return dto.ReservasMockResponseDTO{}, nil, domainErrors.ErrLeerAgendaFicticia
	}

	if strings.Contains(string(encoded), "SHEET") || strings.Contains(string(encoded), "docs.google.com") {
		return dto.ReservasMockResponseDTO{}, nil, domainErrors.ErrAgendaFicticiaReferenciaExterna
	}

	return resp, encoded, nil
}
