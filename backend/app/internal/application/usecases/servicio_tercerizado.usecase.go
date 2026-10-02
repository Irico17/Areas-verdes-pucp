// Package usecases contains application business logic.
package usecases

import (
	"context"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/contracts"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/dto"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/entities"
	domainErrors "github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/errors"
)

type servicioTercerizadoUseCase struct {
	repo      contracts.IServicioTercerizadoRepository
	catalogos contracts.ICatalogoRepository
}

// NewServicioTercerizadoUseCase creates a new IServicioTercerizadoUseCase.
func NewServicioTercerizadoUseCase(repo contracts.IServicioTercerizadoRepository, catalogos contracts.ICatalogoRepository) contracts.IServicioTercerizadoUseCase {
	return &servicioTercerizadoUseCase{repo: repo, catalogos: catalogos}
}

func (u *servicioTercerizadoUseCase) Listar(ctx context.Context) (*dto.OrdenesResponseDTO, error) {
	items, err := u.repo.Listar(ctx)
	if err != nil {
		return nil, err
	}
	res := make([]dto.OrdenDTO, 0, len(items))
	for _, it := range items {
		res = append(res, ordenADTO(it))
	}
	return &dto.OrdenesResponseDTO{
		Aviso:   "La orden no cierra la solicitud ni la labor. Una labor tercerizada solo se cierra si además hay ejecución registrada.",
		Ordenes: res,
	}, nil
}

func (u *servicioTercerizadoUseCase) Crear(ctx context.Context, in dto.CrearOrdenDTO, actorRol, capatazID string) (*dto.OrdenDTO, error) {
	in.Referencia = strings.TrimSpace(in.Referencia)
	in.Conformidad = strings.TrimSpace(in.Conformidad)
	if in.Referencia == "" {
		return nil, domainErrors.InputError{Reason: "la referencia de contratación es obligatoria"}
	}
	if utf8.RuneCountInString(in.Conformidad) > 2000 {
		return nil, domainErrors.InputError{Reason: "la conformidad admite hasta 2000 caracteres"}
	}
	if !uuidRe.MatchString(in.ID) || !uuidRe.MatchString(in.ActividadID) {
		return nil, domainErrors.InputError{Reason: "id y actividad_id deben ser UUID"}
	}
	empresa, err := buscarCatalogo(ctx, u.catalogos, "empresa", in.EmpresaID, in.Empresa, "la empresa se elige del catálogo")
	if err != nil {
		return nil, err
	}
	frecuencia, err := buscarCatalogo(ctx, u.catalogos, "frecuencia", in.FrecuenciaID, in.Frecuencia, "la frecuencia se elige del catálogo")
	if err != nil {
		return nil, err
	}

	entity, err := u.repo.Crear(ctx, entities.NuevaOrdenServicio{
		ID:           in.ID,
		ActividadID:  in.ActividadID,
		Empresa:      empresa.Nombre,
		EmpresaID:    &empresa.ID,
		Referencia:   in.Referencia,
		Frecuencia:   frecuencia.Nombre,
		FrecuenciaID: &frecuencia.ID,
		Conformidad:  in.Conformidad,
	}, actorRol, capatazID)
	if err != nil {
		return nil, err
	}
	out := ordenADTO(entity)
	return &out, nil
}

func (u *servicioTercerizadoUseCase) Evidencias(ctx context.Context, ordenID string) (*dto.EvidenciasOrdenDTO, error) {
	if !uuidRe.MatchString(ordenID) {
		return nil, domainErrors.InputError{Reason: "id debe ser un UUID"}
	}
	entity, err := u.repo.ObtenerPorID(ctx, ordenID)
	if err != nil {
		return nil, err
	}
	orden := ordenADTO(entity)
	return &dto.EvidenciasOrdenDTO{OrdenID: orden.ID, Evidencias: orden.Evidencias}, nil
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
		Orden: ordenADTO(entity),
		Aviso: "La conformidad no cierra la solicitud.",
	}, nil
}

func buscarCatalogo(ctx context.Context, repo contracts.ICatalogoRepository, clase string, id int64, texto, fallo string) (entities.CatalogoItem, error) {
	items, err := repo.List(ctx, clase, true)
	if err != nil {
		return entities.CatalogoItem{}, err
	}
	texto = strings.TrimSpace(texto)
	if id > 0 {
		for _, it := range items {
			if it.ID == id {
				return it, nil
			}
		}
		return entities.CatalogoItem{}, domainErrors.InputError{Reason: fallo}
	}
	if texto == "" {
		return entities.CatalogoItem{}, domainErrors.InputError{Reason: fallo}
	}
	for _, it := range items {
		if it.Codigo == texto || it.Nombre == texto {
			return it, nil
		}
	}
	return entities.CatalogoItem{}, domainErrors.InputError{Reason: fallo}
}

func ordenADTO(it *entities.ServicioTercerizado) dto.OrdenDTO {
	evs := make([]dto.EvidenciaOrdenDTO, 0, len(it.Evidencias))
	for _, ev := range it.Evidencias {
		evs = append(evs, dto.EvidenciaOrdenDTO{
			ID:        ev.ID,
			Nombre:    ev.Nombre,
			Mime:      ev.Mime,
			Bytes:     ev.Bytes,
			Nota:      ev.Nota,
			CreatedAt: ev.CreatedAt.UTC().Format(time.RFC3339),
		})
	}
	return dto.OrdenDTO{
		ID:                 it.ID,
		ActividadID:        it.ActividadID,
		Empresa:            it.Empresa,
		EmpresaID:          it.EmpresaID,
		EmpresaCatalogo:    it.EmpresaCatalogo,
		EmpresaCodigo:      it.EmpresaCodigo,
		Referencia:         it.Referencia,
		Frecuencia:         it.Frecuencia,
		FrecuenciaID:       it.FrecuenciaID,
		FrecuenciaCatalogo: it.FrecuenciaCatalogo,
		FrecuenciaCodigo:   it.FrecuenciaCodigo,
		Estado:             it.Estado,
		Conformidad:        it.Conformidad,
		CreatedAt:          it.CreatedAt.UTC().Format(time.RFC3339),
		Evidencias:         evs,
	}
}
