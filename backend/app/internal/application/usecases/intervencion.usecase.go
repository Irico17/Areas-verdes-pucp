// Package usecases implements application business logic.
package usecases

import (
	"context"
	"math"
	"regexp"
	"sort"
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

func origenDe(v string) string {
	v = strings.TrimSpace(v)
	if v == "" {
		return "interna"
	}
	return v
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

func fechaFiltro(v string) bool {
	if strings.TrimSpace(v) == "" {
		return true
	}
	_, err := time.Parse("2006-01-02", v)
	return err == nil
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
	Origen            string
	CodigoExterno     string
	UnidadSolicitante string
	NivelRiesgo       string
	FechaProgramada   string
	Cantidad          *float64
	Subtipo           string
	Clase             string
	Personal          []string
	LugarID           string
	ZonaSupervisionID string
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
		math.Abs(saved.Lat-in.Lat) < 1e-5 &&
		origenDe(saved.Origen) == origenDe(in.Origen) &&
		saved.CodigoExterno == strings.TrimSpace(in.CodigoExterno) &&
		saved.UnidadSolicitante == strings.TrimSpace(in.UnidadSolicitante) &&
		saved.NivelRiesgo == strings.TrimSpace(in.NivelRiesgo) &&
		saved.FechaProgramada == strings.TrimSpace(in.FechaProgramada) &&
		mismaCantidad(saved.Cantidad, in.Cantidad) &&
		saved.Subtipo == strings.TrimSpace(in.Subtipo) &&
		saved.Clase == strings.TrimSpace(in.Clase) &&
		saved.LugarID == strings.TrimSpace(in.LugarID) &&
		saved.ZonaSupervisionID == strings.TrimSpace(in.ZonaSupervisionID) &&
		mismosNombres(saved.Personal, in.Personal)
}

func mismaCantidad(a, b *float64) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return math.Abs(*a-*b) < 1e-6
}

func mismosNombres(a, b []string) bool {
	if len(a) == 0 && len(b) == 0 {
		return true
	}
	if len(a) != len(b) {
		return false
	}
	aa := append([]string(nil), a...)
	bb := append([]string(nil), b...)
	sort.Strings(aa)
	sort.Strings(bb)
	for i := range aa {
		if aa[i] != bb[i] {
			return false
		}
	}
	return true
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
	if !refCatalogo(f.ZonaSupervisionID) || !refCatalogo(f.CuadrillaID) || !refCatalogo(f.Origen) || !refCatalogo(f.Sector) {
		return entities.Collection("actividades"), domainErrors.InputError{Reason: "filtro no reconocido"}
	}
	if f.NivelRiesgo != "" && !slugRe.MatchString(f.NivelRiesgo) {
		return entities.Collection("actividades"), domainErrors.InputError{Reason: "nivel_riesgo no reconocido"}
	}
	if f.Ejecutor != "" && f.Ejecutor != string(enums.EjecutorPropia) && f.Ejecutor != string(enums.EjecutorTercerizada) {
		return entities.Collection("actividades"), domainErrors.InputError{Reason: "ejecutor no reconocido"}
	}
	if !fechaFiltro(f.Desde) || !fechaFiltro(f.Hasta) {
		return entities.Collection("actividades"), domainErrors.InputError{Reason: "la fecha del filtro no es válida"}
	}
	if f.Desde != "" && f.Hasta != "" && f.Hasta < f.Desde {
		return entities.Collection("actividades"), domainErrors.InputError{Reason: "la fecha hasta es anterior a desde"}
	}

	filter := entities.FiltroIntervenciones{
		Rol:               f.Rol,
		CapatazID:         f.CapatazID,
		Estado:            f.Estado,
		Tipo:              f.Tipo,
		ZonaSupervisionID: f.ZonaSupervisionID,
		CuadrillaID:       f.CuadrillaID,
		Origen:            f.Origen,
		Sector:            f.Sector,
		Ejecutor:          f.Ejecutor,
		NivelRiesgo:       f.NivelRiesgo,
		Desde:             f.Desde,
		Hasta:             f.Hasta,
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
	if err := validarAlta(ctx, u.catalogos, in); err != nil {
		return zero, err
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
		LugarID:           strings.TrimSpace(in.LugarID),
		ZonaSupervisionID: strings.TrimSpace(in.ZonaSupervisionID),
		Origen:            strings.TrimSpace(in.Origen),
		CodigoExterno:     strings.TrimSpace(in.CodigoExterno),
		UnidadSolicitante: strings.TrimSpace(in.UnidadSolicitante),
		NivelRiesgo:       strings.TrimSpace(in.NivelRiesgo),
		FechaProgramada:   strings.TrimSpace(in.FechaProgramada),
		Cantidad:          in.Cantidad,
		Subtipo:           strings.TrimSpace(in.Subtipo),
		Clase:             strings.TrimSpace(in.Clase),
		Personal:          nombresPersonal(in.Personal),
		LugarLibre:        strings.TrimSpace(in.LugarLibre),
		LugarTexto:        strings.TrimSpace(in.LugarTexto),
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

var subtipoRe = regexp.MustCompile(`^[a-z0-9_]{2,40}$`)

func validarAlta(ctx context.Context, catalogos contracts.ICatalogoRepository, in dto.CrearIntervencionDTO) error {
	if strings.TrimSpace(in.LugarLibre) != "" || strings.TrimSpace(in.LugarTexto) != "" {
		return domainErrors.InputError{Reason: "el lugar no se escribe a mano; elija uno del catálogo"}
	}
	if v := strings.TrimSpace(in.Origen); v != "" {
		if !slugRe.MatchString(v) {
			return domainErrors.InputError{Reason: "origen no reconocido"}
		}
		ok, err := catalogos.Activo(ctx, "fuente", v)
		if err != nil {
			return err
		}
		if !ok {
			return domainErrors.InputError{Reason: "origen no está en el catálogo activo"}
		}
	}
	if v := strings.TrimSpace(in.NivelRiesgo); v != "" {
		if !slugRe.MatchString(v) {
			return domainErrors.InputError{Reason: "nivel_riesgo no reconocido"}
		}
		ok, err := catalogos.Activo(ctx, "nivel_riesgo", v)
		if err != nil {
			return err
		}
		if !ok {
			return domainErrors.InputError{Reason: "nivel_riesgo no está en el catálogo activo"}
		}
	}
	clase := strings.TrimSpace(in.Clase)
	subtipo := strings.TrimSpace(in.Subtipo)
	if subtipo != "" && clase == "" {
		return domainErrors.InputError{Reason: "el tipo de actividad exige su clase"}
	}
	if clase != "" {
		if !slugRe.MatchString(clase) {
			return domainErrors.InputError{Reason: "clase no reconocida"}
		}
		ok, err := catalogos.Activo(ctx, "clase_actividad", clase)
		if err != nil {
			return err
		}
		if !ok {
			return domainErrors.InputError{Reason: "clase no está en el catálogo activo"}
		}
	}
	if subtipo != "" && !subtipoRe.MatchString(subtipo) {
		return domainErrors.InputError{Reason: "tipo de actividad no reconocido"}
	}
	if subtipo != "" {
		ok, err := catalogos.Activo(ctx, "subtipo_actividad", subtipo)
		if err != nil {
			return err
		}
		if !ok {
			return domainErrors.InputError{Reason: "el tipo de actividad no está en el catálogo activo"}
		}
	}
	if fecha := strings.TrimSpace(in.FechaProgramada); fecha != "" {
		if _, err := time.Parse("2006-01-02", fecha); err != nil {
			return domainErrors.InputError{Reason: "fecha_programada debe ser aaaa-mm-dd"}
		}
	}
	if in.Cantidad != nil && (*in.Cantidad < 0 || *in.Cantidad > 1000000) {
		return domainErrors.InputError{Reason: "cantidad debe estar entre 0 y 1000000"}
	}
	if utf8.RuneCountInString(strings.TrimSpace(in.CodigoExterno)) > 80 {
		return domainErrors.InputError{Reason: "codigo_externo admite hasta 80 caracteres"}
	}
	if utf8.RuneCountInString(strings.TrimSpace(in.UnidadSolicitante)) > 160 {
		return domainErrors.InputError{Reason: "unidad_solicitante admite hasta 160 caracteres"}
	}
	for _, nombre := range in.Personal {
		n := utf8.RuneCountInString(strings.TrimSpace(nombre))
		if n < 2 || n > 80 {
			return domainErrors.InputError{Reason: "el personal debe ser un nombre ficticio del catálogo"}
		}
	}
	return nil
}

func nombresPersonal(nombres []string) []string {
	if len(nombres) == 0 {
		return nil
	}
	out := make([]string, 0, len(nombres))
	vistos := map[string]struct{}{}
	for _, nombre := range nombres {
		limpio := strings.TrimSpace(nombre)
		clave := strings.ToLower(limpio)
		if limpio == "" {
			continue
		}
		if _, ok := vistos[clave]; ok {
			continue
		}
		vistos[clave] = struct{}{}
		out = append(out, limpio)
	}
	return out
}

func (u *intervencionUseCase) Taxonomia(ctx context.Context) (dto.TaxonomiaActividadDTO, error) {
	row, err := u.repo.Taxonomia(ctx)
	if err != nil {
		return dto.TaxonomiaActividadDTO{}, err
	}
	out := dto.TaxonomiaActividadDTO{
		Clases:   make([]dto.ClaseActividadDTO, 0, len(row.Clases)),
		Riesgos:  make([]dto.OpcionCatalogoDTO, 0, len(row.Riesgos)),
		Origenes: make([]dto.OpcionCatalogoDTO, 0, len(row.Origenes)),
		Personal: make([]dto.PersonalFicticioDTO, 0, len(row.Personal)),
	}
	for _, clase := range row.Clases {
		item := dto.ClaseActividadDTO{
			Codigo: clase.Codigo,
			Nombre: clase.Nombre,
			Tipos:  make([]dto.OpcionCatalogoDTO, 0, len(clase.Tipos)),
		}
		for _, tipo := range clase.Tipos {
			item.Tipos = append(item.Tipos, dto.OpcionCatalogoDTO{Codigo: tipo.Codigo, Nombre: tipo.Nombre})
		}
		out.Clases = append(out.Clases, item)
	}
	for _, riesgo := range row.Riesgos {
		out.Riesgos = append(out.Riesgos, dto.OpcionCatalogoDTO{Codigo: riesgo.Codigo, Nombre: riesgo.Nombre})
	}
	for _, origen := range row.Origenes {
		out.Origenes = append(out.Origenes, dto.OpcionCatalogoDTO{Codigo: origen.Codigo, Nombre: origen.Nombre})
	}
	for _, persona := range row.Personal {
		out.Personal = append(out.Personal, dto.PersonalFicticioDTO{ID: persona.ID, NombreFicticio: persona.NombreFicticio})
	}
	return out, nil
}

func (u *intervencionUseCase) ListarPersonal(ctx context.Context, actividadID string) (dto.PersonalLaborResponseDTO, error) {
	if !uuidRe.MatchString(actividadID) {
		return dto.PersonalLaborResponseDTO{}, domainErrors.InputError{Reason: "id debe ser un UUID"}
	}
	rows, err := u.repo.ListarPersonal(ctx, actividadID)
	if err != nil {
		return dto.PersonalLaborResponseDTO{}, err
	}
	return personalDTO(rows), nil
}

func (u *intervencionUseCase) RegistrarPersonal(ctx context.Context, in dto.RegistrarPersonalDTO) (dto.PersonalLaborResponseDTO, error) {
	if !puedeAsignar(in.ActorRol) {
		if in.ActorRol == RolCapataz {
			return dto.PersonalLaborResponseDTO{}, domainErrors.ErrOperacionProhibido
		}
		return dto.PersonalLaborResponseDTO{}, domainErrors.InputError{Reason: "actor_rol debe ser jefatura, coordinacion o admin"}
	}
	if !uuidRe.MatchString(in.ActividadID) {
		return dto.PersonalLaborResponseDTO{}, domainErrors.InputError{Reason: "id debe ser un UUID"}
	}
	nombres := nombresPersonal(in.Nombres)
	if len(nombres) == 0 {
		return dto.PersonalLaborResponseDTO{}, domainErrors.InputError{Reason: "indique al menos un nombre ficticio"}
	}
	for _, nombre := range nombres {
		n := utf8.RuneCountInString(nombre)
		if n < 2 || n > 80 {
			return dto.PersonalLaborResponseDTO{}, domainErrors.InputError{Reason: "el personal debe ser un nombre ficticio del catálogo"}
		}
	}
	rows, err := u.repo.RegistrarPersonal(ctx, in.ActividadID, nombres)
	if err != nil {
		return dto.PersonalLaborResponseDTO{}, err
	}
	return personalDTO(rows), nil
}

func personalDTO(rows []entities.PersonalLabor) dto.PersonalLaborResponseDTO {
	out := dto.PersonalLaborResponseDTO{Personal: make([]dto.PersonalLaborDTO, 0, len(rows))}
	for _, row := range rows {
		out.Personal = append(out.Personal, dto.PersonalLaborDTO{
			ID:             row.ID,
			NombreFicticio: row.NombreFicticio,
			RolCampo:       row.RolCampo,
		})
	}
	return out
}
