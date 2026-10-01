// Package usecases contains application business logic.
package usecases

import (
	"context"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/contracts"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/dto"
	domainErrors "github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/errors"
)

type solicitudUseCase struct {
	repo contracts.ISolicitudRepository
}

// NewSolicitudUseCase creates a new ISolicitudUseCase.
func NewSolicitudUseCase(repo contracts.ISolicitudRepository) contracts.ISolicitudUseCase {
	return &solicitudUseCase{repo: repo}
}

func (u *solicitudUseCase) Listar(ctx context.Context) (*dto.SolicitudesResponseDTO, error) {
	items, err := u.repo.Listar(ctx)
	if err != nil {
		return nil, err
	}
	res := make([]dto.SolicitudDTO, 0, len(items))
	for _, it := range items {
		s := dto.SolicitudDTO{
			ID:            it.ID,
			CodigoExterno: it.CodigoExterno,
			Fuente:        it.Fuente,
			Titulo:        it.Titulo,
			Detalle:       it.Detalle,
			Prioridad:     it.Prioridad,
			Estado:        it.Estado,
			Lugar:         it.Lugar,
			Cantidad:      it.Cantidad,
			ActividadID:   it.ActividadID,
			CreatedAt:     it.CreatedAt.UTC().Format(time.RFC3339),
		}
		res = append(res, s)
	}
	return &dto.SolicitudesResponseDTO{Solicitudes: res}, nil
}

func (u *solicitudUseCase) Crear(ctx context.Context, in dto.CrearSolicitudDTO) (*dto.SolicitudDTO, error) {
	in.Titulo = strings.TrimSpace(in.Titulo)
	in.CodigoExterno = strings.TrimSpace(in.CodigoExterno)
	in.Fuente = strings.TrimSpace(in.Fuente)
	in.Prioridad = strings.TrimSpace(in.Prioridad)
	if in.Prioridad == "" {
		in.Prioridad = "media"
	}
	if in.Titulo == "" || utf8.RuneCountInString(in.Titulo) > 160 {
		return nil, domainErrors.InputError{Reason: "el título de la solicitud es obligatorio"}
	}
	switch in.Fuente {
	case "centuria", "osg", "correo", "interna":
	default:
		return nil, domainErrors.InputError{Reason: "la fuente debe ser centuria, osg, correo o interna"}
	}
	switch in.Prioridad {
	case "baja", "media", "alta":
	default:
		return nil, domainErrors.InputError{Reason: "prioridad no reconocida"}
	}
	if strings.TrimSpace(in.ActividadID) != "" && !uuidRe.MatchString(in.ActividadID) {
		return nil, domainErrors.InputError{Reason: "actividad_id debe ser un UUID"}
	}
	if !uuidRe.MatchString(in.ID) {
		return nil, domainErrors.InputError{Reason: "id debe ser un UUID"}
	}

	entity, err := u.repo.Crear(ctx, in)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key") || strings.Contains(err.Error(), "unique") {
			return nil, domainErrors.InputError{Reason: "ese código externo ya está registrado"}
		}
		return nil, err
	}

	return &dto.SolicitudDTO{
		ID:            entity.ID,
		CodigoExterno: entity.CodigoExterno,
		Fuente:        entity.Fuente,
		Titulo:        entity.Titulo,
		Detalle:       entity.Detalle,
		Prioridad:     entity.Prioridad,
		Estado:        entity.Estado,
		Lugar:         entity.Lugar,
		Cantidad:      entity.Cantidad,
		ActividadID:   entity.ActividadID,
		CreatedAt:     entity.CreatedAt.UTC().Format(time.RFC3339),
	}, nil
}

func (u *solicitudUseCase) Editar(ctx context.Context, in dto.EditarSolicitudDTO) (*dto.SolicitudDTO, error) {
	if !uuidRe.MatchString(in.ID) {
		return nil, domainErrors.InputError{Reason: "id debe ser un UUID"}
	}
	in.Titulo = strings.TrimSpace(in.Titulo)
	if in.Titulo == "" {
		return nil, domainErrors.InputError{Reason: "el título de la solicitud es obligatorio"}
	}
	in.CodigoExterno = ConservarCodigoSolicitud(in.CodigoExterno)

	entity, err := u.repo.Editar(ctx, in)
	if err != nil {
		return nil, err
	}

	return &dto.SolicitudDTO{
		ID:            entity.ID,
		CodigoExterno: entity.CodigoExterno,
		Fuente:        entity.Fuente,
		Titulo:        entity.Titulo,
		Detalle:       entity.Detalle,
		Prioridad:     entity.Prioridad,
		Estado:        entity.Estado,
		Lugar:         entity.Lugar,
		Cantidad:      entity.Cantidad,
		ActividadID:   entity.ActividadID,
		CreatedAt:     entity.CreatedAt.UTC().Format(time.RFC3339),
	}, nil
}

// ConservarCodigoSolicitud normalizes OSG- codes and clears placeholders.
func ConservarCodigoSolicitud(valor string) string {
	if codigo := CodigoExterno(valor); codigo != "" {
		return codigo
	}
	v := strings.TrimSpace(valor)
	if v == "" {
		return ""
	}
	switch strings.ToLower(normalizarPersona(v)) {
	case "aun no tiene codigo", "no aplica", "sin codigo", "s/c":
		return ""
	}
	return v
}

// CodigoExterno preserves uppercase OSG- prefixes.
func CodigoExterno(valor string) string {
	v := strings.TrimSpace(valor)
	if v == "" {
		return ""
	}
	norma := strings.ToLower(normalizarPersona(v))
	switch norma {
	case "aun no tiene codigo", "no aplica", "sin codigo", "s/c":
		return ""
	}
	if strings.HasPrefix(strings.ToUpper(v), "OSG-") {
		return strings.ToUpper(v)
	}
	return ""
}

func normalizarPersona(s string) string {
	s = strings.TrimSpace(strings.ToLower(s))
	var b strings.Builder
	for _, r := range s {
		switch r {
		case 'á', 'à':
			r = 'a'
		case 'é', 'è':
			r = 'e'
		case 'í', 'ì':
			r = 'i'
		case 'ó', 'ò':
			r = 'o'
		case 'ú', 'ù', 'ü':
			r = 'u'
		case 'ñ':
			r = 'n'
		}
		if unicode.Is(unicode.Mn, r) {
			continue
		}
		b.WriteRune(r)
	}
	return strings.Join(strings.Fields(b.String()), " ")
}
