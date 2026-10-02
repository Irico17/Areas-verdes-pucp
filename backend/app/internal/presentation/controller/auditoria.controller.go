// Package controller provides HTTP handlers for presentation endpoints.
package controller

import (
	"errors"
	"io"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"

	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/contracts"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/dto"
	apperrors "github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/errors"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/presentation/middleware"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/presentation/requests"
)

// IAuditoriaController defines HTTP handlers for audit and batch operations.
type IAuditoriaController interface {
	Importar(c *gin.Context)
	Revertir(c *gin.Context)
	Editar(c *gin.Context)
	Historial(c *gin.Context)
	Timeline(c *gin.Context)
}

type auditoriaController struct {
	loteUC contracts.ILoteUseCase
	logger zerolog.Logger
}

// NewAuditoriaController creates a new IAuditoriaController instance.
func NewAuditoriaController(loteUC contracts.ILoteUseCase, logger zerolog.Logger) IAuditoriaController {
	return &auditoriaController{
		loteUC: loteUC,
		logger: logger,
	}
}

// Importar handles POST /lotes.
// @Summary Importar lote reversible
// @Description Importa filas en un lote reversible con usuario de sesión
// @Tags auditoria
// @Accept json
// @Produce json
// @Param body body requests.LoteRequest true "Lote a importar"
// @Success 201 {object} dto.ImportarLoteResponseDTO
// @Failure 400 {object} map[string]string "JSON inválido o validación de filas"
// @Failure 401 {object} map[string]string "inicie sesión"
// @Failure 403 {object} map[string]string "su rol no tiene ese permiso"
// @Failure 500 {object} map[string]string "no se pudo completar la auditoría"
// @Router /v1/lotes [post]
func (ctrl *auditoriaController) Importar(c *gin.Context) {
	u, ok := middleware.UsuarioEn(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "inicie sesión"})
		return
	}
	var req requests.LoteRequest
	if err := c.ShouldBindJSON(&req); err != nil && !errors.Is(err, io.EOF) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "JSON inválido"})
		return
	}
	filasDTO := make([]dto.FilaLoteDTO, len(req.Filas))
	for i, f := range req.Filas {
		filasDTO[i] = dto.FilaLoteDTO{
			EntidadID: f.EntidadID,
			Accion:    f.Accion,
			Antes:     f.Antes,
			Despues:   f.Despues,
		}
	}
	res, err := ctrl.loteUC.Importar(c.Request.Context(), u.ID, dto.ImportarLoteDTO{
		Entidad: req.Entidad,
		Filas:   filasDTO,
	})
	if err != nil {
		ctrl.writeAuditoriaErr(c, err)
		return
	}
	c.JSON(http.StatusCreated, res)
}

// Revertir handles POST /lotes/:id/revertir.
// @Summary Revertir lote
// @Description Revierte un lote de importación si no hay filas editadas después o si se confirma
// @Tags auditoria
// @Accept json
// @Produce json
// @Param id path int true "ID del lote"
// @Param body body requests.RevertirLoteRequest false "Confirmar si hay filas posteriores"
// @Success 200 {object} dto.ReporteReversionDTO
// @Failure 400 {object} map[string]string "lote inválido o ya revertido"
// @Failure 401 {object} map[string]string "inicie sesión"
// @Failure 403 {object} map[string]string "su rol no tiene ese permiso"
// @Failure 404 {object} map[string]string "lote no encontrado"
// @Failure 409 {object} map[string]any "hay filas editadas después del lote"
// @Failure 500 {object} map[string]string "no se pudo completar la auditoría"
// @Router /v1/lotes/{id}/revertir [post]
func (ctrl *auditoriaController) Revertir(c *gin.Context) {
	u, ok := middleware.UsuarioEn(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "inicie sesión"})
		return
	}
	loteID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || loteID < 1 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "lote inválido"})
		return
	}
	var req requests.RevertirLoteRequest
	if err := c.ShouldBindJSON(&req); err != nil && !errors.Is(err, io.EOF) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "JSON inválido"})
		return
	}
	rep, err := ctrl.loteUC.Revertir(c.Request.Context(), loteID, u.ID, req.Confirmar)
	if err != nil {
		if errors.Is(err, apperrors.ErrConfirmacion) {
			excluidas := []dto.ExcluidaDTO{}
			if rep != nil && rep.Excluidas != nil {
				excluidas = rep.Excluidas
			}
			c.JSON(http.StatusConflict, gin.H{
				"error":     "hay filas editadas después del lote",
				"excluidas": excluidas,
			})
			return
		}
		ctrl.writeAuditoriaErr(c, err)
		return
	}
	c.JSON(http.StatusOK, rep)
}

// Editar handles POST /auditoria/ediciones.
// @Summary Registrar edición auditada
// @Description Registra un cambio manual posterior al lote
// @Tags auditoria
// @Accept json
// @Produce json
// @Param body body requests.EditarAuditoriaRequest true "Datos de la edición"
// @Success 200 {object} dto.EditarAuditoriaResponseDTO
// @Failure 400 {object} map[string]string "JSON inválido o validación"
// @Failure 401 {object} map[string]string "inicie sesión"
// @Failure 403 {object} map[string]string "su rol no tiene ese permiso"
// @Failure 500 {object} map[string]string "no se pudo completar la auditoría"
// @Router /v1/auditoria/ediciones [post]
func (ctrl *auditoriaController) Editar(c *gin.Context) {
	u, ok := middleware.UsuarioEn(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "inicie sesión"})
		return
	}
	var req requests.EditarAuditoriaRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "JSON inválido"})
		return
	}
	res, err := ctrl.loteUC.Editar(c.Request.Context(), u.ID, dto.EditarAuditoriaDTO{
		Entidad:   req.Entidad,
		EntidadID: req.EntidadID,
		Despues:   req.Despues,
	})
	if err != nil {
		ctrl.writeAuditoriaErr(c, err)
		return
	}
	c.JSON(http.StatusOK, res)
}

// Historial handles GET /auditoria/cambios.
// @Summary Historial filtrable de cambios
// @Description Devuelve el listado de cambios filtrable por zona, origen y fechas
// @Tags auditoria
// @Produce json
// @Param entidad query string false "Entidad"
// @Param entidad_id query string false "ID de entidad"
// @Param zona query string false "Zona"
// @Param origen query string false "Origen"
// @Param desde query string false "Fecha desde"
// @Param hasta query string false "Fecha hasta"
// @Success 200 {object} dto.HistorialAuditoriaResponseDTO
// @Failure 400 {object} map[string]string "error de fecha"
// @Failure 401 {object} map[string]string "inicie sesión"
// @Failure 403 {object} map[string]string "su rol no tiene ese permiso"
// @Failure 500 {object} map[string]string "no se pudo completar la auditoría"
// @Router /v1/auditoria/cambios [get]
func (ctrl *auditoriaController) Historial(c *gin.Context) {
	ctrl.listar(c, false)
}

// Timeline handles GET /auditoria/timeline.
// @Summary Timeline de cambios de una fila
// @Description Devuelve los cambios de una fila con usuario de sesión
// @Tags auditoria
// @Produce json
// @Param entidad query string true "Entidad"
// @Param entidad_id query string true "ID de entidad"
// @Success 200 {object} dto.HistorialAuditoriaResponseDTO
// @Failure 400 {object} map[string]string "entidad y entidad_id son obligatorios"
// @Failure 401 {object} map[string]string "inicie sesión"
// @Failure 403 {object} map[string]string "su rol no tiene ese permiso"
// @Failure 500 {object} map[string]string "no se pudo completar la auditoría"
// @Router /v1/auditoria/timeline [get]
func (ctrl *auditoriaController) Timeline(c *gin.Context) {
	ctrl.listar(c, true)
}

func (ctrl *auditoriaController) listar(c *gin.Context, timeline bool) {
	f := dto.FiltroAuditoriaDTO{
		Entidad:   c.Query("entidad"),
		EntidadID: c.Query("entidad_id"),
		Zona:      c.Query("zona"),
		Origen:    c.Query("origen"),
		Desde:     c.Query("desde"),
		Hasta:     c.Query("hasta"),
	}
	var (
		ev  []dto.EventoAuditoriaDTO
		err error
	)
	if timeline {
		ev, err = ctrl.loteUC.Timeline(c.Request.Context(), f)
	} else {
		ev, err = ctrl.loteUC.Historial(c.Request.Context(), f)
	}
	if err != nil {
		ctrl.writeAuditoriaErr(c, err)
		return
	}
	if ev == nil {
		ev = []dto.EventoAuditoriaDTO{}
	}
	c.JSON(http.StatusOK, gin.H{"eventos": ev})
}

func (ctrl *auditoriaController) writeAuditoriaErr(c *gin.Context, err error) {
	var input apperrors.InputError
	switch {
	case errors.Is(err, apperrors.ErrLoteNoEncontrado):
		c.JSON(http.StatusNotFound, gin.H{"error": "lote no encontrado"})
	case errors.As(err, &input):
		c.JSON(http.StatusBadRequest, gin.H{"error": input.Reason})
	default:
		ctrl.logger.Error().Err(err).Msg("auditoria: error interno")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "no se pudo completar la auditoría"})
	}
}
