// Package usecases implements application business logic.
package usecases

import (
	"context"
	"math"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/contracts"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/dto"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/constants/enums"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/entities"
	domainErrors "github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/errors"
)

// Roles de supervisión.
const (
	RolJefatura     = "jefatura"
	RolCoordinacion = "coordinacion"
	RolCapataz      = "capataz"
	RolAdmin        = "admin"
)

// Bounding box aproximado del campus Pando (EPSG:4326), con margen corto.
const (
	MinLon = -77.090
	MaxLon = -77.070
	MinLat = -12.080
	MaxLat = -12.060
)

var (
	uuidRe = regexp.MustCompile(`(?i)^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)
	slugRe = regexp.MustCompile(`^[a-z0-9_]{2,32}$`)
	refRe  = regexp.MustCompile(`^[A-Za-z0-9_-]{0,40}$`)
)

func rolConocido(rol string) bool {
	return rol == RolJefatura || rol == RolCoordinacion || rol == RolCapataz || rol == RolAdmin
}

func puedeAsignar(rol string) bool {
	return rol == RolJefatura || rol == RolCoordinacion || rol == RolAdmin
}

func ejecutorDe(v string) string {
	if strings.TrimSpace(v) == string(enums.EjecutorTercerizada) {
		return string(enums.EjecutorTercerizada)
	}
	return string(enums.EjecutorPropia)
}

func refCatalogo(v string) bool {
	return refRe.MatchString(v)
}

// SavedPayload represents a persisted labor used for idempotent duplicate payload comparison.
type SavedPayload struct {
	Tipo              string
	Titulo            string
	Detalle           string
	AssignedCapatazID string
	AreaFeatureID     string
	ZonaFeatureID     string
	Lon               float64
	Lat               float64
	Ejecutor          string
	CreatedAt         time.Time
}

// SamePayload compares an incoming creation payload with the existing saved row for idempotency.
func SamePayload(saved SavedPayload, in entities.NuevaIntervencion) bool {
	return saved.Tipo == in.Tipo &&
		saved.Titulo == strings.TrimSpace(in.Titulo) &&
		saved.Detalle == strings.TrimSpace(in.Detalle) &&
		saved.AssignedCapatazID == strings.TrimSpace(in.AssignedCapatazID) &&
		saved.AreaFeatureID == strings.TrimSpace(in.AreaFeatureID) &&
		saved.ZonaFeatureID == strings.TrimSpace(in.ZonaFeatureID) &&
		ejecutorDe(saved.Ejecutor) == ejecutorDe(in.Ejecutor) &&
		math.Abs(saved.Lon-in.Lon) < 1e-5 &&
		math.Abs(saved.Lat-in.Lat) < 1e-5
}

// PuedeCerrar validates requirements for closing a labor.
func PuedeCerrar(ejecutor string, tieneOrden, tieneEjecucion bool) error {
	if ejecutorDe(ejecutor) == string(enums.EjecutorTercerizada) && !tieneOrden {
		return domainErrors.InputError{Reason: "una labor tercerizada no se cierra sin una orden de servicio"}
	}
	if !tieneEjecucion {
		return domainErrors.InputError{Reason: "no se cierra la labor sin una ejecución registrada"}
	}
	return nil
}

type intervencionUseCase struct {
	repo      contracts.IIntervencionRepository
	catalogos contracts.ICatalogoRepository
}

// NewIntervencionUseCase creates a new IIntervencionUseCase.
func NewIntervencionUseCase(repo contracts.IIntervencionRepository, catalogos contracts.ICatalogoRepository) contracts.IIntervencionUseCase {
	return &intervencionUseCase{repo: repo, catalogos: catalogos}
}

func (u *intervencionUseCase) ListarCapataces(ctx context.Context) ([]dto.CapatazDTO, error) {
	rows, err := u.repo.Capataces(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]dto.CapatazDTO, len(rows))
	for i, r := range rows {
		out[i] = dto.CapatazDTO{
			ID:     r.ID,
			Equipo: r.Equipo,
			Turno:  r.Turno,
		}
	}
	return out, nil
}

func (u *intervencionUseCase) ListarActividades(ctx context.Context, f dto.FiltroIntervencionesDTO) (entities.FeatureCollection, error) {
	if f.Rol != "" && !rolConocido(f.Rol) {
		return entities.Collection("actividades"), domainErrors.InputError{Reason: "rol no reconocido"}
	}
	if f.Rol == RolCapataz && strings.TrimSpace(f.CapatazID) == "" {
		return entities.Collection("actividades"), domainErrors.InputError{Reason: "capataz_id es obligatorio para el rol capataz"}
	}
	if f.Estado != "" && !slugRe.MatchString(f.Estado) {
		return entities.Collection("actividades"), domainErrors.InputError{Reason: "estado no reconocido"}
	}
	if f.Tipo != "" && !slugRe.MatchString(f.Tipo) {
		return entities.Collection("actividades"), domainErrors.InputError{Reason: "tipo no reconocido"}
	}
	if !refCatalogo(f.ZonaSupervisionID) || !refCatalogo(f.CuadrillaID) || !refCatalogo(f.Origen) {
		return entities.Collection("actividades"), domainErrors.InputError{Reason: "filtro no reconocido"}
	}

	filter := entities.FiltroIntervenciones{
		Rol:               f.Rol,
		CapatazID:         f.CapatazID,
		Estado:            f.Estado,
		Tipo:              f.Tipo,
		ZonaSupervisionID: f.ZonaSupervisionID,
		CuadrillaID:       f.CuadrillaID,
		Origen:            f.Origen,
		SoloAbiertas:      f.SoloAbiertas,
	}
	return u.repo.List(ctx, filter)
}

func (u *intervencionUseCase) CrearActividad(ctx context.Context, in dto.CrearIntervencionDTO) (dto.CrearIntervencionResponseDTO, error) {
	var zero dto.CrearIntervencionResponseDTO
	if !puedeAsignar(in.ActorRol) {
		if in.ActorRol == RolCapataz {
			return zero, domainErrors.ErrOperacionProhibido
		}
		return zero, domainErrors.InputError{Reason: "actor_rol debe ser jefatura, coordinacion o admin"}
	}
	if !uuidRe.MatchString(in.ID) {
		return zero, domainErrors.InputError{Reason: "id debe ser un UUID"}
	}
	if !slugRe.MatchString(in.Tipo) {
		return zero, domainErrors.InputError{Reason: "tipo no reconocido"}
	}
	titulo := strings.TrimSpace(in.Titulo)
	if titulo == "" || utf8.RuneCountInString(titulo) > 160 {
		return zero, domainErrors.InputError{Reason: "titulo es obligatorio y de hasta 160 caracteres"}
	}
	if utf8.RuneCountInString(in.Detalle) > 2000 {
		return zero, domainErrors.InputError{Reason: "detalle admite hasta 2000 caracteres"}
	}
	sinPunto := strings.TrimSpace(in.LugarID) != "" || strings.TrimSpace(in.ZonaSupervisionID) != ""
	if !sinPunto && (in.Lon < MinLon || in.Lon > MaxLon || in.Lat < MinLat || in.Lat > MaxLat) {
		return zero, domainErrors.InputError{Reason: "el punto queda fuera del campus"}
	}

	cmd := entities.NuevaIntervencion{
		ID:                in.ID,
		Tipo:              in.Tipo,
		Titulo:            in.Titulo,
		Detalle:           in.Detalle,
		Lon:               in.Lon,
		Lat:               in.Lat,
		AreaFeatureID:     in.AreaFeatureID,
		ZonaFeatureID:     in.ZonaFeatureID,
		AssignedCapatazID: in.AssignedCapatazID,
		ActorRol:          in.ActorRol,
		Ejecutor:          in.Ejecutor,
		UsuarioID:         in.UsuarioID,
		LugarID:           in.LugarID,
		ZonaSupervisionID: in.ZonaSupervisionID,
	}
	feat, created, err := u.repo.Create(ctx, cmd)
	if err != nil {
		return zero, err
	}
	return dto.CrearIntervencionResponseDTO{
		Creada:  created,
		Feature: feat,
	}, nil
}

func (u *intervencionUseCase) AsignarActividad(ctx context.Context, in dto.AsignarIntervencionDTO) (entities.Feature, error) {
	var zero entities.Feature
	if !puedeAsignar(in.ActorRol) {
		if in.ActorRol == RolCapataz {
			return zero, domainErrors.ErrOperacionProhibido
		}
		return zero, domainErrors.InputError{Reason: "actor_rol debe ser jefatura, coordinacion o admin"}
	}
	if strings.TrimSpace(in.CapatazID) == "" {
		return zero, domainErrors.InputError{Reason: "capataz_id es obligatorio"}
	}
	cmd := entities.AsignarIntervencion{
		ID:        in.ID,
		CapatazID: in.CapatazID,
		ActorRol:  in.ActorRol,
		UsuarioID: in.UsuarioID,
	}
	return u.repo.Assign(ctx, cmd)
}

func (u *intervencionUseCase) CambiarEstado(ctx context.Context, in dto.CambiarEstadoDTO) (entities.Feature, error) {
	var zero entities.Feature
	if !rolConocido(in.ActorRol) {
		return zero, domainErrors.InputError{Reason: "actor_rol no reconocido"}
	}
	if !slugRe.MatchString(in.Estado) {
		return zero, domainErrors.InputError{Reason: "estado no reconocido"}
	}
	if in.ActorRol == RolCapataz {
		if strings.TrimSpace(in.CapatazID) == "" {
			return zero, domainErrors.InputError{Reason: "capataz_id es obligatorio para cambiar el estado"}
		}
		if in.Estado == string(enums.EstadoCerrada) || in.Estado == string(enums.EstadoCancelada) || in.Estado == "archivada" {
			return zero, domainErrors.ForbiddenError{Reason: "el capataz no puede cerrar ni cancelar una labor"}
		}
	}
	nombre, err := u.etiquetaEstadoActivo(ctx, in.Estado)
	if err != nil {
		return zero, err
	}
	actual, err := u.repo.EstadoActual(ctx, in.ID)
	if err != nil {
		return zero, err
	}
	if !enums.TransicionEstadoPermitida(actual, in.Estado) {
		return zero, domainErrors.InputError{Reason: "esa transición de estado no está permitida"}
	}
	cmd := entities.CambiarEstadoIntervencion{
		ID:        in.ID,
		Estado:    in.Estado,
		ActorRol:  in.ActorRol,
		CapatazID: in.CapatazID,
		UsuarioID: in.UsuarioID,
	}
	feat, err := u.repo.SetEstado(ctx, cmd)
	if err != nil {
		return zero, err
	}
	return conEtiquetaEstado(feat, nombre), nil
}

func (u *intervencionUseCase) etiquetaEstadoActivo(ctx context.Context, codigo string) (string, error) {
	items, err := u.catalogos.List(ctx, string(enums.ClaseEstado), false)
	if err != nil {
		return "", err
	}
	for _, item := range items {
		if item.Codigo == codigo && item.Activo {
			return item.Nombre, nil
		}
	}
	return "", domainErrors.InputError{Reason: "el estado no está activo en el catálogo"}
}

func conEtiquetaEstado(feat entities.Feature, nombre string) entities.Feature {
	switch props := feat.Properties.(type) {
	case entities.ActividadProperties:
		props.EstadoEtiqueta = nombre
		feat.Properties = props
	case *entities.ActividadProperties:
		if props != nil {
			copia := *props
			copia.EstadoEtiqueta = nombre
			feat.Properties = copia
		}
	}
	return feat
}

func (u *intervencionUseCase) ArchivarActividad(ctx context.Context, in dto.ArchivarIntervencionDTO) (dto.ArchivarIntervencionResponseDTO, error) {
	var zero dto.ArchivarIntervencionResponseDTO
	if !puedeAsignar(in.ActorRol) {
		if in.ActorRol == RolCapataz {
			return zero, domainErrors.ErrOperacionProhibido
		}
		return zero, domainErrors.InputError{Reason: "actor_rol debe ser jefatura, coordinacion o admin"}
	}
	cmd := entities.ArchivarIntervencion{
		ID:        in.ID,
		ActorRol:  in.ActorRol,
		Motivo:    in.Motivo,
		UsuarioID: in.UsuarioID,
	}
	if err := u.repo.Archive(ctx, cmd); err != nil {
		return zero, err
	}
	return dto.ArchivarIntervencionResponseDTO{
		Archivada: true,
		ID:        in.ID,
	}, nil
}

func (u *intervencionUseCase) Timeline(ctx context.Context, id string) (dto.TimelineResponseDTO, error) {
	out := dto.TimelineResponseDTO{ActividadID: id, Eventos: []dto.EventoTimelineDTO{}}
	if !uuidRe.MatchString(id) {
		return out, domainErrors.InputError{Reason: "id debe ser un UUID"}
	}
	events, err := u.repo.Timeline(ctx, id)
	if err != nil {
		return out, err
	}
	out.Eventos = make([]dto.EventoTimelineDTO, len(events))
	for i, ev := range events {
		out.Eventos[i] = dto.EventoTimelineDTO{
			ID:          ev.ID,
			Tipo:        ev.Tipo,
			Estado:      ev.Estado,
			CapatazID:   ev.CapatazID,
			Equipo:      ev.Equipo,
			ActorRol:    ev.ActorRol,
			UsuarioID:   ev.UsuarioID,
			Usuario:     ev.Usuario,
			Nombre:      ev.Nombre,
			Nota:        ev.Nota,
			UUIDCliente: ev.UUIDCliente,
			CreatedAt:   ev.CreatedAt.UTC().Format(time.RFC3339),
		}
	}
	return out, nil
}

func (u *intervencionUseCase) GuardarFicha(ctx context.Context, in dto.FichaIntervencionDTO) error {
	if !uuidRe.MatchString(in.ID) {
		return domainErrors.InputError{Reason: "id debe ser un UUID"}
	}
	if in.FechaSolicitud != "" && in.FechaAtencion != "" && in.FechaAtencion < in.FechaSolicitud {
		return domainErrors.InputError{Reason: "la atención no puede ser anterior a la solicitud"}
	}
	cmd := entities.FichaIntervencion{
		ID:             in.ID,
		Clase:          in.Clase,
		FechaSolicitud: in.FechaSolicitud,
		FechaAtencion:  in.FechaAtencion,
		Lugar:          in.Lugar,
		Comentario:     in.Comentario,
		ActorRol:       in.ActorRol,
		CapatazID:      in.CapatazID,
	}
	return u.repo.GuardarFicha(ctx, cmd)
}

func (u *intervencionUseCase) CrearAvance(ctx context.Context, in dto.CrearAvanceDTO) error {
	if !uuidRe.MatchString(in.ActividadID) || !uuidRe.MatchString(in.ID) {
		return domainErrors.InputError{Reason: "id debe ser un UUID"}
	}
	if _, err := time.Parse("2006-01-02", in.Fecha); err != nil {
		return domainErrors.InputError{Reason: "la fecha usa el formato AAAA-MM-DD"}
	}
	if strings.TrimSpace(in.AreaFeatureID) == "" && strings.TrimSpace(in.EjemplarRef) == "" {
		return domainErrors.InputError{Reason: "el avance se liga a un área o a un ejemplar"}
	}
	cmd := entities.NuevoAvance{
		ActividadID:   in.ActividadID,
		ID:            in.ID,
		Fecha:         in.Fecha,
		Nota:          in.Nota,
		AreaFeatureID: in.AreaFeatureID,
		EjemplarRef:   in.EjemplarRef,
		ActorRol:      in.ActorRol,
		CapatazID:     in.CapatazID,
	}
	return u.repo.CrearAvance(ctx, cmd)
}
