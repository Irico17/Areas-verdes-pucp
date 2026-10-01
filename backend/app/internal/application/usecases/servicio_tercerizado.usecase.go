// Package usecases contains application business logic.
package usecases

import (
	"context"
	"strings"
	"time"

	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/contracts"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/dto"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/entities"
	domainErrors "github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/errors"
)

type servicioTercerizadoUseCase struct {
	repo contracts.IServicioTercerizadoRepository
}

// NewServicioTercerizadoUseCase creates a new IServicioTercerizadoUseCase.
func NewServicioTercerizadoUseCase(repo contracts.IServicioTercerizadoRepository) contracts.IServicioTercerizadoUseCase {
	return &servicioTercerizadoUseCase{repo: repo}
}

func (u *servicioTercerizadoUseCase) Listar(ctx context.Context) (*dto.OrdenesResponseDTO, error) {
	items, err := u.repo.Listar(ctx)
	if err != nil {
		return nil, err
	}
	res := make([]dto.OrdenDTO, 0, len(items))
	for _, it := range items {
		o := dto.OrdenDTO{
			ID:          it.ID,
			ActividadID: it.ActividadID,
			Empresa:     it.Empresa,
			Referencia:  it.Referencia,
			Frecuencia:  it.Frecuencia,
			Estado:      it.Estado,
			Conformidad: it.Conformidad,
			CreatedAt:   it.CreatedAt.UTC().Format(time.RFC3339),
		}
		res = append(res, o)
	}
	return &dto.OrdenesResponseDTO{
		Aviso:   "La orden no cierra la solicitud ni la labor. Una labor tercerizada solo se cierra si además hay ejecución registrada.",
		Ordenes: res,
	}, nil
}

func (u *servicioTercerizadoUseCase) Crear(ctx context.Context, in dto.CrearOrdenDTO, actorRol, capatazID string) (*dto.OrdenDTO, error) {
	in.Empresa = strings.TrimSpace(in.Empresa)
	in.Referencia = strings.TrimSpace(in.Referencia)
	if in.Empresa == "" || in.Referencia == "" {
		return nil, domainErrors.InputError{Reason: "la empresa y la referencia son obligatorias"}
	}
	if !uuidRe.MatchString(in.ID) || !uuidRe.MatchString(in.ActividadID) {
		return nil, domainErrors.InputError{Reason: "id y actividad_id deben ser UUID"}
	}

	entity, err := u.repo.Crear(ctx, entities.NuevaOrdenServicio{
		ID:          in.ID,
		ActividadID: in.ActividadID,
		Empresa:     in.Empresa,
		Referencia:  in.Referencia,
		Frecuencia:  in.Frecuencia,
	}, actorRol, capatazID)
	if err != nil {
		return nil, err
	}

	return &dto.OrdenDTO{
		ID:          entity.ID,
		ActividadID: entity.ActividadID,
		Empresa:     entity.Empresa,
		Referencia:  entity.Referencia,
		Frecuencia:  entity.Frecuencia,
		Estado:      entity.Estado,
		Conformidad: entity.Conformidad,
		CreatedAt:   entity.CreatedAt.UTC().Format(time.RFC3339),
	}, nil
}

func (u *servicioTercerizadoUseCase) Editar(ctx context.Context, in dto.EditarOrdenDTO, actorRol, capatazID string) (*dto.EditarOrdenResponseDTO, error) {
	if !uuidRe.MatchString(in.ID) {
		return nil, domainErrors.InputError{Reason: "id debe ser un UUID"}
	}
	estado := strings.TrimSpace(in.Estado)
	if estado == "" {
		estado = "en_proceso"
	}
	switch estado {
	case "en_proceso", "ejecutada", "conforme":
	default:
		return nil, domainErrors.InputError{Reason: "estado de orden no reconocido"}
	}
	in.Estado = estado

	entity, err := u.repo.Editar(ctx, entities.EditarOrdenServicio{
		ID:               in.ID,
		Conformidad:      in.Conformidad,
		PeriodoInicio:    in.PeriodoInicio,
		PeriodoFin:       in.PeriodoFin,
		ReporteProveedor: in.ReporteProveedor,
		Estado:           in.Estado,
	}, actorRol, capatazID)
	if err != nil {
		return nil, err
	}

	return &dto.EditarOrdenResponseDTO{
		Orden: dto.OrdenDTO{
			ID:          entity.ID,
			ActividadID: entity.ActividadID,
			Empresa:     entity.Empresa,
			Referencia:  entity.Referencia,
			Frecuencia:  entity.Frecuencia,
			Estado:      entity.Estado,
			Conformidad: entity.Conformidad,
			CreatedAt:   entity.CreatedAt.UTC().Format(time.RFC3339),
		},
		Aviso: "La conformidad no cierra la solicitud.",
	}, nil
}
