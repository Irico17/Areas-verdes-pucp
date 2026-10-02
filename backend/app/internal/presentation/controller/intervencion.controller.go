// Package controller implements HTTP request handlers for the presentation layer.
package controller

import (
	"errors"
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"

	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/contracts"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/dto"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/usecases"
	domainErrors "github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/errors"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/presentation/middleware"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/presentation/requests"
)

// IIntervencionController defines HTTP endpoints for activities, capataces, lifecycle, and progress.
type IIntervencionController interface {
	Capataces(*gin.Context)
	List(*gin.Context)
	Create(*gin.Context)
	Assign(*gin.Context)
	Estado(*gin.Context)
	Archive(*gin.Context)
	Timeline(*gin.Context)
	Ficha(*gin.Context)
	CrearAvance(*gin.Context)
	Taxonomia(*gin.Context)
	ListarPersonal(*gin.Context)
	RegistrarPersonal(*gin.Context)
}

type intervencionController struct {
	uc     contracts.IIntervencionUseCase
	logger zerolog.Logger
}

// NewIntervencionController creates a new IIntervencionController instance.
func NewIntervencionController(
	uc contracts.IIntervencionUseCase,
	logger zerolog.Logger,
) IIntervencionController {
	return &intervencionController{
		uc:     uc,
		logger: logger,
	}
}

func actorDeSesion(c *gin.Context, capatazCuerpo string) (dto.UsuarioSesionDTO, string, bool) {
	u, ok := middleware.UsuarioEn(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "inicie sesión"})
		return dto.UsuarioSesionDTO{}, "", false
	}
	if u.Rol == usecases.RolCapataz {
		if strings.TrimSpace(u.CapatazID) == "" {
			c.JSON(http.StatusForbidden, gin.H{"error": "capataz sin identificador asignado"})
			return u, "", false
		}
		return u, u.CapatazID, true
	}
	return u, capatazCuerpo, true
}

func writeOperacionErr(c *gin.Context, logger *zerolog.Logger, err error) {
	var input domainErrors.InputError
	var forb domainErrors.ForbiddenError
	switch {
	case errors.As(err, &forb):
		c.JSON(http.StatusForbidden, gin.H{"error": forb.Reason})
	case errors.Is(err, domainErrors.ErrOperacionProhibido):
		c.JSON(http.StatusForbidden, gin.H{"error": "este rol no puede hacer esa acción"})
	case errors.Is(err, domainErrors.ErrLaborNoEncontrada):
		c.JSON(http.StatusNotFound, gin.H{"error": "labor no encontrada"})
	case errors.Is(err, domainErrors.ErrLaborConflicto):
		c.JSON(http.StatusConflict, gin.H{"error": "ese id ya existe con otro contenido"})
	case errors.As(err, &input):
		c.JSON(http.StatusBadRequest, gin.H{"error": input.Reason})
	default:
		if logger != nil {
			logger.Error().Err(err).Msg("operacion")
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "no se pudo completar la operación"})
	}
}

// Capataces godoc
// @Summary List active capataces
// @Tags operacion
// @Produce json
// @Success 200 {object} dto.CapatacesResponseDTO
// @Router /v1/operacion/capataces [get]
func (ctrl *intervencionController) Capataces(c *gin.Context) {
	if _, ok := middleware.UsuarioEn(c); !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "inicie sesión"})
		return
	}
	rows, err := ctrl.uc.ListarCapataces(c.Request.Context())
	if err != nil {
		ctrl.logger.Error().Err(err).Msg("capataces")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "no se pudo leer los equipos"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"capataces": rows})
}

// List godoc
// @Summary List activities as GeoJSON FeatureCollection
// @Tags operacion
// @Produce json
// @Success 200 {object} entities.FeatureCollection
// @Router /v1/operacion/actividades [get]
func (ctrl *intervencionController) List(c *gin.Context) {
	u, ok := middleware.UsuarioEn(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "inicie sesión"})
		return
	}

	q := dto.FiltroIntervencionesDTO{
		Estado:            c.Query("estado"),
		Tipo:              c.Query("tipo"),
		ZonaSupervisionID: c.Query("zona_supervision_id"),
		CuadrillaID:       c.Query("cuadrilla_id"),
		Origen:            c.Query("origen"),
		SoloAbiertas:      c.DefaultQuery("abiertas", "1") != "0",
	}

	if u.Rol == usecases.RolCapataz {
		q.Rol = usecases.RolCapataz
		q.CapatazID = u.CapatazID
	} else {
		q.Rol = c.Query("rol")
		q.CapatazID = c.Query("capataz_id")
	}

	fc, err := ctrl.uc.ListarActividades(c.Request.Context(), q)
	if err != nil {
		writeOperacionErr(c, &ctrl.logger, err)
		return
	}
	c.Header("Content-Type", "application/geo+json; charset=utf-8")
	c.JSON(http.StatusOK, fc)
}

// Create godoc
// @Summary Create an activity from map pin or reference
// @Tags operacion
// @Accept json
// @Produce json
// @Success 201 {object} dto.CrearIntervencionResponseDTO
// @Router /v1/operacion/actividades [post]
func (ctrl *intervencionController) Create(c *gin.Context) {
	var body requests.CrearIntervencionRequest
	if err := c.ShouldBindJSON(&body); err != nil && !errors.Is(err, io.EOF) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "JSON inválido"})
		return
	}

	u, capataz, ok := actorDeSesion(c, body.AssignedCapatazID)
	if !ok {
		return
	}

	in := dto.CrearIntervencionDTO{
		ID:                body.ID,
		Tipo:              body.Tipo,
		Titulo:            body.Titulo,
		Detalle:           body.Detalle,
		Lon:               body.Lon,
		Lat:               body.Lat,
		AreaFeatureID:     body.AreaFeatureID,
		ZonaFeatureID:     body.ZonaFeatureID,
		AssignedCapatazID: capataz,
		ActorRol:          u.Rol,
		Ejecutor:          body.Ejecutor,
		UsuarioID:         u.ID,
		LugarID:           body.LugarID,
		LugarLibre:        body.LugarLibre,
		LugarTexto:        body.Lugar,
		ZonaSupervisionID: body.ZonaSupervisionID,
		Origen:            body.Origen,
		CodigoExterno:     body.CodigoExterno,
		UnidadSolicitante: body.UnidadSolicitante,
		NivelRiesgo:       body.NivelRiesgo,
		FechaProgramada:   body.FechaProgramada,
		Cantidad:          body.Cantidad,
		Subtipo:           body.Subtipo,
		Clase:             body.Clase,
		Personal:          body.Personal,
	}

	res, err := ctrl.uc.CrearActividad(c.Request.Context(), in)
	if err != nil {
		writeOperacionErr(c, &ctrl.logger, err)
		return
	}

	status := http.StatusOK
	if res.Creada {
		status = http.StatusCreated
	}
	c.JSON(status, gin.H{"creada": res.Creada, "feature": res.Feature})
}

// Assign godoc
// @Summary Assign an activity to a capataz team
// @Tags operacion
// @Accept json
// @Produce json
// @Success 200 {object} map[string]any
// @Router /v1/operacion/actividades/{id}/asignacion [patch]
func (ctrl *intervencionController) Assign(c *gin.Context) {
	var body requests.AsignarIntervencionRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "JSON inválido"})
		return
	}

	u, capataz, ok := actorDeSesion(c, body.CapatazID)
	if !ok {
		return
	}

	in := dto.AsignarIntervencionDTO{
		ID:        c.Param("id"),
		CapatazID: capataz,
		ActorRol:  u.Rol,
		UsuarioID: u.ID,
	}

	feature, err := ctrl.uc.AsignarActividad(c.Request.Context(), in)
	if err != nil {
		writeOperacionErr(c, &ctrl.logger, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"feature": feature})
}

// Estado godoc
// @Summary Update the state of an activity
// @Tags operacion
// @Accept json
// @Produce json
// @Success 200 {object} map[string]any
// @Router /v1/operacion/actividades/{id}/estado [patch]
func (ctrl *intervencionController) Estado(c *gin.Context) {
	var body requests.CambiarEstadoRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "JSON inválido"})
		return
	}

	u, capataz, ok := actorDeSesion(c, body.CapatazID)
	if !ok {
		return
	}

	in := dto.CambiarEstadoDTO{
		ID:        c.Param("id"),
		Estado:    body.Estado,
		ActorRol:  u.Rol,
		CapatazID: capataz,
		UsuarioID: u.ID,
	}

	feature, err := ctrl.uc.CambiarEstado(c.Request.Context(), in)
	if err != nil {
		writeOperacionErr(c, &ctrl.logger, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"feature": feature})
}

// Archive godoc
// @Summary Soft-delete/archive an activity with reason
// @Tags operacion
// @Accept json
// @Produce json
// @Success 200 {object} dto.ArchivarIntervencionResponseDTO
// @Router /v1/operacion/actividades/{id}/archivar [post]
func (ctrl *intervencionController) Archive(c *gin.Context) {
	var body requests.ArchivarIntervencionRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "JSON inválido"})
		return
	}

	u, _, ok := actorDeSesion(c, "")
	if !ok {
		return
	}

	in := dto.ArchivarIntervencionDTO{
		ID:        c.Param("id"),
		ActorRol:  u.Rol,
		Motivo:    body.Motivo,
		UsuarioID: u.ID,
	}

	res, err := ctrl.uc.ArchivarActividad(c.Request.Context(), in)
	if err != nil {
		writeOperacionErr(c, &ctrl.logger, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"archivada": res.Archivada, "id": res.ID})
}

// Timeline godoc
// @Summary Retrieve the full event timeline of an activity
// @Tags operacion
// @Produce json
// @Success 200 {object} dto.TimelineResponseDTO
// @Router /v1/operacion/actividades/{id}/timeline [get]
func (ctrl *intervencionController) Timeline(c *gin.Context) {
	if _, ok := middleware.UsuarioEn(c); !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "inicie sesión"})
		return
	}
	tl, err := ctrl.uc.Timeline(c.Request.Context(), c.Param("id"))
	if err != nil {
		writeOperacionErr(c, &ctrl.logger, err)
		return
	}
	c.JSON(http.StatusOK, tl)
}

// Ficha godoc
// @Summary Update activity metadata from desktop form
// @Tags operacion
// @Accept json
// @Produce json
// @Success 200 {object} map[string]any
// @Router /v1/operacion/actividades/{id}/ficha [patch]
func (ctrl *intervencionController) Ficha(c *gin.Context) {
	u, ok := middleware.UsuarioEn(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "inicie sesión"})
		return
	}

	var body requests.FichaIntervencionRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "JSON inválido"})
		return
	}

	in := dto.FichaIntervencionDTO{
		ID:             c.Param("id"),
		Clase:          body.Clase,
		FechaSolicitud: body.FechaSolicitud,
		FechaAtencion:  body.FechaAtencion,
		Lugar:          body.Lugar,
		Comentario:     body.Comentario,
		ActorRol:       u.Rol,
		CapatazID:      u.CapatazID,
	}

	if err := ctrl.uc.GuardarFicha(c.Request.Context(), in); err != nil {
		writeOperacionErr(c, &ctrl.logger, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"id": in.ID})
}

// CrearAvance godoc
// @Summary Record progress log entry for an activity
// @Tags operacion
// @Accept json
// @Produce json
// @Success 201 {object} map[string]any
// @Router /v1/operacion/actividades/{id}/avances [post]
func (ctrl *intervencionController) CrearAvance(c *gin.Context) {
	u, ok := middleware.UsuarioEn(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "inicie sesión"})
		return
	}

	var body requests.CrearAvanceRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "JSON inválido"})
		return
	}

	in := dto.CrearAvanceDTO{
		ActividadID:   c.Param("id"),
		ID:            body.ID,
		Fecha:         body.Fecha,
		Nota:          body.Nota,
		AreaFeatureID: body.AreaFeatureID,
		EjemplarRef:   body.EjemplarRef,
		ActorRol:      u.Rol,
		CapatazID:     u.CapatazID,
	}

	if err := ctrl.uc.CrearAvance(c.Request.Context(), in); err != nil {
		writeOperacionErr(c, &ctrl.logger, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"id": body.ID})
}

// Taxonomia godoc
// @Summary Activity class, type, risk and fictional staff catalogs
// @Tags operacion
// @Produce json
// @Success 200 {object} dto.TaxonomiaActividadDTO
// @Router /v1/operacion/taxonomia-actividad [get]
func (ctrl *intervencionController) Taxonomia(c *gin.Context) {
	if _, ok := middleware.UsuarioEn(c); !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "inicie sesión"})
		return
	}
	out, err := ctrl.uc.Taxonomia(c.Request.Context())
	if err != nil {
		writeOperacionErr(c, &ctrl.logger, err)
		return
	}
	c.JSON(http.StatusOK, out)
}

// ListarPersonal godoc
// @Summary List fictional staff assigned to an activity
// @Tags operacion
// @Produce json
// @Success 200 {object} dto.PersonalLaborResponseDTO
// @Router /v1/operacion/actividades/{id}/personal [get]
func (ctrl *intervencionController) ListarPersonal(c *gin.Context) {
	if _, ok := middleware.UsuarioEn(c); !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "inicie sesión"})
		return
	}
	out, err := ctrl.uc.ListarPersonal(c.Request.Context(), c.Param("id"))
	if err != nil {
		writeOperacionErr(c, &ctrl.logger, err)
		return
	}
	c.JSON(http.StatusOK, out)
}

// RegistrarPersonal godoc
// @Summary Assign fictional staff to an activity without creating accounts
// @Tags operacion
// @Accept json
// @Produce json
// @Success 201 {object} dto.PersonalLaborResponseDTO
// @Router /v1/operacion/actividades/{id}/personal [post]
func (ctrl *intervencionController) RegistrarPersonal(c *gin.Context) {
	u, ok := middleware.UsuarioEn(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "inicie sesión"})
		return
	}
	var body requests.RegistrarPersonalRequest
	if err := c.ShouldBindJSON(&body); err != nil && !errors.Is(err, io.EOF) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "JSON inválido"})
		return
	}
	nombres := body.Nombres
	if strings.TrimSpace(body.NombreFicticio) != "" {
		nombres = append(nombres, body.NombreFicticio)
	}
	out, err := ctrl.uc.RegistrarPersonal(c.Request.Context(), dto.RegistrarPersonalDTO{
		ActividadID: c.Param("id"),
		Nombres:     nombres,
		ActorRol:    u.Rol,
	})
	if err != nil {
		writeOperacionErr(c, &ctrl.logger, err)
		return
	}
	c.JSON(http.StatusCreated, out)
}
