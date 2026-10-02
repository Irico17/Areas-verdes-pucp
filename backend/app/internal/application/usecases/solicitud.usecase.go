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
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/entities"
	domainErrors "github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/errors"
)

type solicitudUseCase struct {
	repo      contracts.ISolicitudRepository
	catalogos contracts.ICatalogoRepository
}

// NewSolicitudUseCase creates a new ISolicitudUseCase.
func NewSolicitudUseCase(repo contracts.ISolicitudRepository, catalogos contracts.ICatalogoRepository) contracts.ISolicitudUseCase {
	return &solicitudUseCase{repo: repo, catalogos: catalogos}
}

func (u *solicitudUseCase) Listar(ctx context.Context) (*dto.SolicitudesResponseDTO, error) {
	items, err := u.repo.Listar(ctx)
	if err != nil {
		return nil, err
	}
	res := make([]dto.SolicitudDTO, 0, len(items))
	for _, it := range items {
		res = append(res, solicitudADTO(it))
	}
	return &dto.SolicitudesResponseDTO{Solicitudes: res}, nil
}

func (u *solicitudUseCase) Crear(ctx context.Context, in dto.CrearSolicitudDTO) (*dto.SolicitudDTO, error) {
	in.Titulo = strings.TrimSpace(in.Titulo)
	in.CodigoExterno = strings.TrimSpace(in.CodigoExterno)
	in.Fuente = strings.TrimSpace(in.Fuente)
	in.Prioridad = strings.TrimSpace(in.Prioridad)
	in.Estado = strings.TrimSpace(in.Estado)
	if in.Prioridad == "" {
		in.Prioridad = "media"
	}
	if in.Titulo == "" || utf8.RuneCountInString(in.Titulo) > 160 {
		return nil, domainErrors.InputError{Reason: "el título de la solicitud es obligatorio"}
	}
	if err := u.fuenteDeCatalogo(ctx, in.Fuente); err != nil {
		return nil, err
	}
	if err := u.estadoDeCatalogo(ctx, in.Estado); err != nil {
		return nil, err
	}
	switch in.Prioridad {
	case "baja", "media", "alta":
	default:
		return nil, domainErrors.InputError{Reason: "prioridad no reconocida"}
	}
	if err := validarCantidadesSolicitud(in.Cantidad, in.CantidadSolicitada, in.CantidadEjecutada); err != nil {
		return nil, err
	}
	if err := validarPuntoSolicitud(in.Lat, in.Lon); err != nil {
		return nil, err
	}
	if in.LugarID != nil && *in.LugarID <= 0 {
		return nil, domainErrors.InputError{Reason: "el lugar no está en el catálogo"}
	}
	if strings.TrimSpace(in.ActividadID) != "" && !uuidRe.MatchString(in.ActividadID) {
		return nil, domainErrors.InputError{Reason: "actividad_id debe ser un UUID"}
	}
	if !uuidRe.MatchString(in.ID) {
		return nil, domainErrors.InputError{Reason: "id debe ser un UUID"}
	}
	copiarCantidadSolicitada(&in.Cantidad, &in.CantidadSolicitada)

	entity, err := u.repo.Crear(ctx, entities.NuevaSolicitud{
		ID:                 in.ID,
		CodigoExterno:      in.CodigoExterno,
		Fuente:             in.Fuente,
		Titulo:             in.Titulo,
		Detalle:            in.Detalle,
		Prioridad:          in.Prioridad,
		Estado:             in.Estado,
		Lugar:              in.Lugar,
		LugarID:            in.LugarID,
		Lat:                in.Lat,
		Lon:                in.Lon,
		Cantidad:           in.Cantidad,
		CantidadSolicitada: in.CantidadSolicitada,
		CantidadEjecutada:  in.CantidadEjecutada,
		ActividadID:        in.ActividadID,
	})
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key") || strings.Contains(err.Error(), "unique") {
			return nil, domainErrors.InputError{Reason: "ese código externo ya está registrado"}
		}
		return nil, err
	}
	out := solicitudADTO(entity)
	return &out, nil
}

func (u *solicitudUseCase) Editar(ctx context.Context, in dto.EditarSolicitudDTO) (*dto.SolicitudDTO, error) {
	if !uuidRe.MatchString(in.ID) {
		return nil, domainErrors.InputError{Reason: "id debe ser un UUID"}
	}
	in.Titulo = strings.TrimSpace(in.Titulo)
	in.Estado = strings.TrimSpace(in.Estado)
	if in.Titulo == "" {
		return nil, domainErrors.InputError{Reason: "el título de la solicitud es obligatorio"}
	}
	if err := u.estadoDeCatalogo(ctx, in.Estado); err != nil {
		return nil, err
	}
	if err := validarCantidadesSolicitud(0, in.CantidadSolicitada, in.CantidadEjecutada); err != nil {
		return nil, err
	}
	if err := validarPuntoSolicitud(in.Lat, in.Lon); err != nil {
		return nil, err
	}
	if in.LugarID != nil && *in.LugarID <= 0 {
		return nil, domainErrors.InputError{Reason: "el lugar no está en el catálogo"}
	}
	in.CodigoExterno = ConservarCodigoSolicitud(in.CodigoExterno)

	entity, err := u.repo.Editar(ctx, entities.EditarSolicitud{
		ID:                 in.ID,
		CodigoExterno:      in.CodigoExterno,
		Titulo:             in.Titulo,
		Detalle:            in.Detalle,
		Prioridad:          in.Prioridad,
		Estado:             in.Estado,
		Lugar:              in.Lugar,
		LugarID:            in.LugarID,
		Lat:                in.Lat,
		Lon:                in.Lon,
		CantidadSolicitada: in.CantidadSolicitada,
		CantidadEjecutada:  in.CantidadEjecutada,
	})
	if err != nil {
		return nil, err
	}
	out := solicitudADTO(entity)
	return &out, nil
}

func (u *solicitudUseCase) fuenteDeCatalogo(ctx context.Context, fuente string) error {
	if !slugRe.MatchString(fuente) {
		return domainErrors.InputError{Reason: "la fuente no está en el catálogo activo"}
	}
	ok, err := u.catalogos.Activo(ctx, "fuente", fuente)
	if err != nil {
		return err
	}
	if !ok {
		return domainErrors.InputError{Reason: "la fuente no está en el catálogo activo"}
	}
	return nil
}

func (u *solicitudUseCase) estadoDeCatalogo(ctx context.Context, estado string) error {
	if estado == "" {
		return nil
	}
	if !slugRe.MatchString(estado) {
		return domainErrors.InputError{Reason: "el estado de la solicitud no está en el catálogo activo"}
	}
	ok, err := u.catalogos.Activo(ctx, "estado_solicitud", estado)
	if err != nil {
		return err
	}
	if !ok {
		return domainErrors.InputError{Reason: "el estado de la solicitud no está en el catálogo activo"}
	}
	return nil
}

func validarCantidadesSolicitud(cantidad int, solicitada, ejecutada *int) error {
	if cantidad < 0 || cantidad > 1000000 {
		return domainErrors.InputError{Reason: "la cantidad debe estar entre 0 y 1000000"}
	}
	for _, n := range []*int{solicitada, ejecutada} {
		if n != nil && (*n < 0 || *n > 1000000) {
			return domainErrors.InputError{Reason: "la cantidad debe estar entre 0 y 1000000"}
		}
	}
	return nil
}

func validarPuntoSolicitud(lat, lon *float64) error {
	if (lat == nil) != (lon == nil) {
		return domainErrors.InputError{Reason: "lat y lon van juntos"}
	}
	if lat == nil {
		return nil
	}
	if *lat < MinLat || *lat > MaxLat || *lon < MinLon || *lon > MaxLon {
		return domainErrors.InputError{Reason: "el punto queda fuera del campus"}
	}
	return nil
}

// copiarCantidadSolicitada llena la pedida desde cantidad, y no borra cantidad si ya vino.
func copiarCantidadSolicitada(cantidad *int, solicitada **int) {
	if *solicitada == nil && *cantidad != 0 {
		n := *cantidad
		*solicitada = &n
	}
	if *cantidad == 0 && *solicitada != nil {
		*cantidad = **solicitada
	}
}

func solicitudADTO(it *entities.Solicitud) dto.SolicitudDTO {
	return dto.SolicitudDTO{
		ID:                 it.ID,
		CodigoExterno:      it.CodigoExterno,
		Fuente:             it.Fuente,
		Titulo:             it.Titulo,
		Detalle:            it.Detalle,
		Prioridad:          it.Prioridad,
		Estado:             it.Estado,
		Lugar:              it.Lugar,
		LugarID:            it.LugarID,
		LugarNombre:        it.LugarNombre,
		Lat:                it.Lat,
		Lon:                it.Lon,
		Cantidad:           it.Cantidad,
		CantidadSolicitada: it.CantidadSolicitada,
		CantidadEjecutada:  it.CantidadEjecutada,
		ActividadID:        it.ActividadID,
		CreatedAt:          it.CreatedAt.UTC().Format(time.RFC3339),
	}
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
