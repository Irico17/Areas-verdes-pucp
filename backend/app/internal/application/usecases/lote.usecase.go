// Package usecases contains application business logic and orchestration.
package usecases

import (
	"context"
	"encoding/json"
	"time"

	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/contracts"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/dto"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/entities"
	apperrors "github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/errors"
)

type loteUseCase struct {
	repo contracts.ILoteRepository
}

// NewLoteUseCase creates a new ILoteUseCase instance.
func NewLoteUseCase(repo contracts.ILoteRepository) contracts.ILoteUseCase {
	return &loteUseCase{repo: repo}
}

func (uc *loteUseCase) Importar(ctx context.Context, usuarioID int64, req dto.ImportarLoteDTO) (*dto.ImportarLoteResponseDTO, error) {
	filas := make([]entities.FilaLote, len(req.Filas))
	for i, f := range req.Filas {
		filas[i] = entities.FilaLote{
			EntidadID: f.EntidadID,
			Accion:    f.Accion,
			Antes:     f.Antes,
			Despues:   f.Despues,
		}
	}
	id, err := uc.repo.Importar(ctx, usuarioID, req.Entidad, filas)
	if err != nil {
		return nil, err
	}
	return &dto.ImportarLoteResponseDTO{
		LoteID: id,
		Filas:  len(req.Filas),
	}, nil
}

func (uc *loteUseCase) Revertir(ctx context.Context, loteID, usuarioID int64, confirmar bool) (*dto.ReporteReversionDTO, error) {
	rep, err := uc.repo.Revertir(ctx, loteID, usuarioID, confirmar)
	excluidas := make([]dto.ExcluidaDTO, len(rep.Excluidas))
	for i, e := range rep.Excluidas {
		excluidas[i] = dto.ExcluidaDTO{
			EntidadID: e.EntidadID,
			Motivo:    e.Motivo,
		}
	}
	revertidas := rep.Revertidas
	if revertidas == nil {
		revertidas = []string{}
	}
	dtoRep := &dto.ReporteReversionDTO{
		LoteID:     rep.LoteID,
		Revertidas: revertidas,
		Excluidas:  excluidas,
	}
	return dtoRep, err
}

func (uc *loteUseCase) Editar(ctx context.Context, usuarioID int64, req dto.EditarAuditoriaDTO) (*dto.EditarAuditoriaResponseDTO, error) {
	raw, err := json.Marshal(req.Despues)
	if err != nil {
		return nil, apperrors.InputError{Reason: "JSON inválido"}
	}
	if err := uc.repo.Editar(ctx, usuarioID, req.Entidad, req.EntidadID, raw); err != nil {
		return nil, err
	}
	return &dto.EditarAuditoriaResponseDTO{
		Editada:   true,
		EntidadID: req.EntidadID,
	}, nil
}

func (uc *loteUseCase) Timeline(ctx context.Context, f dto.FiltroAuditoriaDTO) ([]dto.EventoAuditoriaDTO, error) {
	filtro := entities.FiltroAuditoria{
		Entidad:   f.Entidad,
		EntidadID: f.EntidadID,
		Zona:      f.Zona,
		Origen:    f.Origen,
		Desde:     f.Desde,
		Hasta:     f.Hasta,
	}
	ev, err := uc.repo.Timeline(ctx, filtro)
	if err != nil {
		return nil, err
	}
	return mapearEventosADTO(ev), nil
}

func (uc *loteUseCase) Historial(ctx context.Context, f dto.FiltroAuditoriaDTO) ([]dto.EventoAuditoriaDTO, error) {
	filtro := entities.FiltroAuditoria{
		Entidad:   f.Entidad,
		EntidadID: f.EntidadID,
		Zona:      f.Zona,
		Origen:    f.Origen,
		Desde:     f.Desde,
		Hasta:     f.Hasta,
	}
	ev, err := uc.repo.Historial(ctx, filtro)
	if err != nil {
		return nil, err
	}
	return mapearEventosADTO(ev), nil
}

func mapearEventosADTO(ev []entities.EventoAuditoria) []dto.EventoAuditoriaDTO {
	out := make([]dto.EventoAuditoriaDTO, len(ev))
	for i, e := range ev {
		out[i] = dto.EventoAuditoriaDTO{
			ID:        e.ID,
			Entidad:   e.Entidad,
			EntidadID: e.EntidadID,
			Accion:    e.Accion,
			UsuarioID: e.UsuarioID,
			Usuario:   e.Usuario,
			Nombre:    e.Nombre,
			LoteID:    e.LoteID,
			CreatedAt: e.CreatedAt.UTC().Format(time.RFC3339),
		}
	}
	return out
}
