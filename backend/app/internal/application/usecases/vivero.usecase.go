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

type viveroUseCase struct {
	repo contracts.IViveroRepository
}

// NewViveroUseCase creates a new IViveroUseCase instance.
func NewViveroUseCase(repo contracts.IViveroRepository) contracts.IViveroUseCase {
	return &viveroUseCase{repo: repo}
}

func (u *viveroUseCase) Listar(ctx context.Context, mes string) (*dto.ViveroResponseDTO, error) {
	mes = strings.TrimSpace(mes)
	if mes != "" && (len(mes) != 7 || mes[4] != '-') {
		return nil, domainErrors.InputError{Reason: "el mes usa AAAA-MM"}
	}
	items, err := u.repo.Listar(ctx, mes)
	if err != nil {
		return nil, err
	}
	out := make([]dto.ViveroDTO, 0, len(items))
	for _, it := range items {
		v := dto.ViveroDTO{
			ID:            it.ID,
			Area:          it.Area,
			Subproceso:    it.Subproceso,
			Etapa:         it.Etapa,
			Descripcion:   it.Descripcion,
			Observaciones: it.Observaciones,
			Responsables:  it.Responsables,
			LugarLibre:    it.LugarLibre,
		}
		if it.Fecha != nil {
			v.Fecha = *it.Fecha
		}
		if it.LugarID != nil {
			v.LugarID = *it.LugarID
		}
		out = append(out, v)
	}
	return &dto.ViveroResponseDTO{Registros: out}, nil
}

func validarVivero(in dto.GuardarViveroDTO) (string, error) {
	if !uuidRe.MatchString(in.ID) {
		return "", domainErrors.InputError{Reason: "id debe ser un UUID"}
	}
	area := strings.TrimSpace(in.Area)
	switch area {
	case "Fauna", "Flora", "Ambiental", "Otros", "":
	default:
		return "", domainErrors.InputError{Reason: "el área debe estar en el catálogo"}
	}
	return area, nil
}

func (u *viveroUseCase) Crear(ctx context.Context, in dto.GuardarViveroDTO) (*dto.ViveroDTO, error) {
	area, err := validarVivero(in)
	if err != nil {
		return nil, err
	}
	in.Area = area
	entity, err := u.repo.Guardar(ctx, entities.GuardarVivero{
		ID:            in.ID,
		Fecha:         in.Fecha,
		Area:          in.Area,
		Subproceso:    in.Subproceso,
		Etapa:         in.Etapa,
		Descripcion:   in.Descripcion,
		Observaciones: in.Observaciones,
		Responsables:  in.Responsables,
		LugarID:       in.LugarID,
		LugarLibre:    in.LugarLibre,
	})
	if err != nil {
		return nil, err
	}
	fecha := ""
	if entity.Fecha != nil {
		fecha = *entity.Fecha
	}
	lugarID := ""
	if entity.LugarID != nil {
		lugarID = *entity.LugarID
	}
	return &dto.ViveroDTO{
		ID:           entity.ID,
		Fecha:        fecha,
		Area:         entity.Area,
		Subproceso:   entity.Subproceso,
		Etapa:        entity.Etapa,
		Descripcion:  entity.Descripcion,
		Responsables: entity.Responsables,
		LugarID:      lugarID,
		LugarLibre:   entity.LugarLibre,
	}, nil
}

func (u *viveroUseCase) Editar(ctx context.Context, in dto.GuardarViveroDTO) (*dto.ViveroDTO, error) {
	area, err := validarVivero(in)
	if err != nil {
		return nil, err
	}
	in.Area = area
	entity, err := u.repo.Guardar(ctx, entities.GuardarVivero{
		ID:            in.ID,
		Fecha:         in.Fecha,
		Area:          in.Area,
		Subproceso:    in.Subproceso,
		Etapa:         in.Etapa,
		Descripcion:   in.Descripcion,
		Observaciones: in.Observaciones,
		Responsables:  in.Responsables,
		LugarID:       in.LugarID,
		LugarLibre:    in.LugarLibre,
	})
	if err != nil {
		return nil, err
	}
	fecha := ""
	if entity.Fecha != nil {
		fecha = *entity.Fecha
	}
	lugarID := ""
	if entity.LugarID != nil {
		lugarID = *entity.LugarID
	}
	return &dto.ViveroDTO{
		ID:           entity.ID,
		Fecha:        fecha,
		Area:         entity.Area,
		Subproceso:   entity.Subproceso,
		Etapa:        entity.Etapa,
		Descripcion:  entity.Descripcion,
		Responsables: entity.Responsables,
		LugarID:      lugarID,
		LugarLibre:   entity.LugarLibre,
	}, nil
}

func (u *viveroUseCase) Archivar(ctx context.Context, id string) error {
	if !uuidRe.MatchString(id) {
		return domainErrors.InputError{Reason: "id debe ser un UUID"}
	}
	return u.repo.Archivar(ctx, id)
}
