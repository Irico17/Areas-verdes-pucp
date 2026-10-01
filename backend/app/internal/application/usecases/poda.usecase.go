// Package usecases contains application business logic workflows.
package usecases

import (
	"context"
	"strings"

	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/contracts"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/dto"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/entities"
	domainErrors "github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/errors"
)

type podaUseCase struct {
	repo contracts.IPodaRepository
}

// NewPodaUseCase creates a new IPodaUseCase instance.
func NewPodaUseCase(repo contracts.IPodaRepository) contracts.IPodaUseCase {
	return &podaUseCase{repo: repo}
}

func validarPoda(in dto.GuardarPodaDTO) error {
	codigo := strings.TrimSpace(in.Codigo)
	if len(codigo) < 4 || !strings.HasPrefix(codigo, "PO-") {
		return domainErrors.InputError{Reason: "el código de poda es PO-n"}
	}
	if !uuidRe.MatchString(in.ID) {
		return domainErrors.InputError{Reason: "id debe ser un UUID"}
	}
	switch strings.TrimSpace(in.Prioridad) {
	case "baja", "media", "alta":
	default:
		return domainErrors.InputError{Reason: "prioridad no reconocida"}
	}
	if in.CantidadPedida < 0 || in.CantidadEjecutada < 0 {
		return domainErrors.InputError{Reason: "las cantidades son cero o más"}
	}
	return nil
}

func (u *podaUseCase) Listar(ctx context.Context) (*dto.PodasResponseDTO, error) {
	items, err := u.repo.Listar(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]dto.PodaDTO, 0, len(items))
	for _, it := range items {
		p := dto.PodaDTO{
			ID:                it.ID,
			Codigo:            it.Codigo,
			Tipo:              it.Tipo,
			TipoActividad:     it.TipoActividad,
			Personal:          it.Personal,
			Ubicacion:         it.Ubicacion,
			Unidad:            it.Unidad,
			CantidadPedida:    it.CantidadPedida,
			CantidadEjecutada: it.CantidadEjecutada,
			Prioridad:         it.Prioridad,
			Comentario:        it.Comentario,
			NombreComun:       it.NombreComun,
			NombreCientifico:  it.NombreCientifico,
		}
		if it.CodigoExterno != nil {
			p.CodigoExterno = *it.CodigoExterno
		}
		if it.FechaReporte != nil {
			p.FechaReporte = *it.FechaReporte
		}
		if it.FechaEjecucion != nil {
			p.FechaEjecucion = *it.FechaEjecucion
		}
		out = append(out, p)
	}
	return &dto.PodasResponseDTO{Podas: out}, nil
}

func (u *podaUseCase) Crear(ctx context.Context, in dto.GuardarPodaDTO) (*dto.PodaDTO, error) {
	in.CodigoExterno = CodigoExterno(in.CodigoExterno)
	if err := validarPoda(in); err != nil {
		return nil, err
	}
	entity, err := u.repo.Guardar(ctx, entities.GuardarPoda{
		ID:                in.ID,
		Codigo:            in.Codigo,
		CodigoExterno:     in.CodigoExterno,
		Tipo:              in.Tipo,
		TipoActividad:     in.TipoActividad,
		FechaReporte:      in.FechaReporte,
		FechaEjecucion:    in.FechaEjecucion,
		Personal:          in.Personal,
		Ubicacion:         in.Ubicacion,
		Unidad:            in.Unidad,
		CantidadPedida:    in.CantidadPedida,
		CantidadEjecutada: in.CantidadEjecutada,
		Prioridad:         in.Prioridad,
		Comentario:        in.Comentario,
		NombreComun:       in.NombreComun,
		NombreCientifico:  in.NombreCientifico,
	})
	if err != nil {
		return nil, err
	}
	codExt := ""
	if entity.CodigoExterno != nil {
		codExt = *entity.CodigoExterno
	}
	return &dto.PodaDTO{
		ID:                entity.ID,
		Codigo:            strings.TrimSpace(entity.Codigo),
		CodigoExterno:     codExt,
		Tipo:              entity.Tipo,
		TipoActividad:     entity.TipoActividad,
		Personal:          entity.Personal,
		Ubicacion:         entity.Ubicacion,
		Unidad:            entity.Unidad,
		CantidadPedida:    entity.CantidadPedida,
		CantidadEjecutada: entity.CantidadEjecutada,
		Prioridad:         strings.TrimSpace(entity.Prioridad),
		Comentario:        entity.Comentario,
	}, nil
}

func (u *podaUseCase) Editar(ctx context.Context, in dto.GuardarPodaDTO) (*dto.PodaDTO, error) {
	in.CodigoExterno = CodigoExterno(in.CodigoExterno)
	if err := validarPoda(in); err != nil {
		return nil, err
	}
	entity, err := u.repo.Guardar(ctx, entities.GuardarPoda{
		ID:                in.ID,
		Codigo:            in.Codigo,
		CodigoExterno:     in.CodigoExterno,
		Tipo:              in.Tipo,
		TipoActividad:     in.TipoActividad,
		FechaReporte:      in.FechaReporte,
		FechaEjecucion:    in.FechaEjecucion,
		Personal:          in.Personal,
		Ubicacion:         in.Ubicacion,
		Unidad:            in.Unidad,
		CantidadPedida:    in.CantidadPedida,
		CantidadEjecutada: in.CantidadEjecutada,
		Prioridad:         in.Prioridad,
		Comentario:        in.Comentario,
		NombreComun:       in.NombreComun,
		NombreCientifico:  in.NombreCientifico,
	})
	if err != nil {
		return nil, err
	}
	codExt := ""
	if entity.CodigoExterno != nil {
		codExt = *entity.CodigoExterno
	}
	return &dto.PodaDTO{
		ID:                entity.ID,
		Codigo:            strings.TrimSpace(entity.Codigo),
		CodigoExterno:     codExt,
		Tipo:              entity.Tipo,
		TipoActividad:     entity.TipoActividad,
		Personal:          entity.Personal,
		Ubicacion:         entity.Ubicacion,
		Unidad:            entity.Unidad,
		CantidadPedida:    entity.CantidadPedida,
		CantidadEjecutada: entity.CantidadEjecutada,
		Prioridad:         strings.TrimSpace(entity.Prioridad),
		Comentario:        entity.Comentario,
	}, nil
}

func (u *podaUseCase) Archivar(ctx context.Context, id string) error {
	if !uuidRe.MatchString(id) {
		return domainErrors.InputError{Reason: "id debe ser un UUID"}
	}
	return u.repo.Archivar(ctx, id)
}
