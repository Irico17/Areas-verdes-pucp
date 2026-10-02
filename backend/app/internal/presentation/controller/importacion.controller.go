// Package controller provides HTTP handlers for presentation endpoints.
package controller

import (
	"database/sql"
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"

	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/contracts"
	apperrors "github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/errors"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/presentation/middleware"
)

// IImportacionController defines HTTP handlers for file import endpoints.
type IImportacionController interface {
	Entidades(c *gin.Context)
	Previsualizar(c *gin.Context)
	Confirmar(c *gin.Context)
}

type importacionController struct {
	importacionUC contracts.IImportacionUseCase
	logger        zerolog.Logger
}

// NewImportacionController creates a new IImportacionController instance.
func NewImportacionController(
	importacionUC contracts.IImportacionUseCase,
	logger zerolog.Logger,
) IImportacionController {
	return &importacionController{
		importacionUC: importacionUC,
		logger:        logger,
	}
}

// Entidades handles GET /importaciones/entidades.
// @Summary Entidades importables
// @Description Devuelve la lista de entidades importables del mapa de datos
// @Tags importaciones
// @Produce json
// @Success 200 {object} dto.EntidadesImportablesResponseDTO
// @Failure 401 {object} map[string]string "inicie sesión"
// @Failure 403 {object} map[string]string "su rol no tiene ese permiso"
// @Router /v1/importaciones/entidades [get]
func (ctrl *importacionController) Entidades(c *gin.Context) {
	entidades := ctrl.importacionUC.Entidades(c.Request.Context())
	c.JSON(http.StatusOK, gin.H{"entidades": entidades})
}

// Previsualizar handles POST /importaciones.
// @Summary Vista previa de importación
// @Description Sube un archivo CSV/GeoJSON/JSON y genera una vista previa persistida en lote
// @Tags importaciones
// @Accept multipart/form-data
// @Produce json
// @Param entidad query string true "Nombre de la entidad"
// @Param archivo formData file true "Archivo a importar"
// @Success 200 {object} dto.VistaPreviaResponseDTO
// @Failure 400 {object} map[string]string "falta el archivo o entidad no importable"
// @Failure 401 {object} map[string]string "inicie sesión"
// @Failure 403 {object} map[string]string "su rol no tiene ese permiso"
// @Failure 500 {object} map[string]string "no se pudo guardar la vista previa"
// @Router /v1/importaciones [post]
func (ctrl *importacionController) Previsualizar(c *gin.Context) {
	u, ok := middleware.UsuarioEn(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "inicie sesión"})
		return
	}

	entidad := c.Query("entidad")
	archivo, err := c.FormFile("archivo")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "falta el archivo"})
		return
	}

	f, err := archivo.Open()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no se pudo leer el archivo"})
		return
	}
	defer f.Close()

	body, err := io.ReadAll(io.LimitReader(f, 32<<20))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no se pudo leer el archivo"})
		return
	}

	res, err := ctrl.importacionUC.Previsualizar(c.Request.Context(), entidad, archivo.Filename, body, u.ID)
	if err != nil {
		ctrl.writeImportErr(c, err)
		return
	}

	c.JSON(http.StatusOK, res)
}

// Confirmar handles POST /importaciones/:id/confirmar.
// @Summary Confirmar importación
// @Description Confirma un lote de importación en estado vista previa y escribe las filas
// @Tags importaciones
// @Produce json
// @Param id path int true "ID del lote"
// @Success 201 {object} dto.ConfirmarImportacionResponseDTO
// @Failure 400 {object} map[string]string "lote inválido"
// @Failure 401 {object} map[string]string "inicie sesión"
// @Failure 403 {object} map[string]string "su rol no tiene ese permiso"
// @Failure 404 {object} map[string]string "lote no encontrado"
// @Failure 409 {object} map[string]string "el lote no está en vista previa"
// @Failure 422 {object} map[string]any "ninguna fila válida"
// @Failure 500 {object} map[string]string "no se pudo importar"
// @Router /v1/importaciones/{id}/confirmar [post]
func (ctrl *importacionController) Confirmar(c *gin.Context) {
	u, ok := middleware.UsuarioEn(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "inicie sesión"})
		return
	}

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id < 1 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "lote inválido"})
		return
	}

	res, err := ctrl.importacionUC.Confirmar(c.Request.Context(), id, u.ID)
	if err != nil {
		ctrl.writeImportErr(c, err)
		return
	}

	c.JSON(http.StatusCreated, res)
}

func (ctrl *importacionController) writeImportErr(c *gin.Context, err error) {
	switch {
	case errors.Is(err, apperrors.ErrGuardarVistaPrevia):
		ctrl.logger.Error().Err(err).Msg("importaciones")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "no se pudo guardar la vista previa"})
	case errors.Is(err, apperrors.ErrEntidad) || strings.Contains(err.Error(), "entidad no importable"):
		c.JSON(http.StatusBadRequest, gin.H{"error": "entidad no importable"})
	case errors.Is(err, apperrors.ErrSinValidas) || strings.Contains(err.Error(), "ninguna fila válida"):
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "ninguna fila válida: no se escribió", "escrito": false})
	case errors.Is(err, apperrors.ErrLote) || errors.Is(err, apperrors.ErrLoteNoConfirmable) || strings.Contains(err.Error(), "lote no está en vista previa") || strings.Contains(err.Error(), "lote no confirmable"):
		c.JSON(http.StatusConflict, gin.H{"error": "el lote no está en vista previa"})
	case errors.Is(err, sql.ErrNoRows):
		c.JSON(http.StatusNotFound, gin.H{"error": "lote no encontrado"})
	default:
		msg := err.Error()
		if strings.Contains(msg, "SQLSTATE") || strings.Contains(msg, "pq:") {
			ctrl.logger.Error().Err(err).Msg("importaciones: error de base de datos")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "no se pudo importar"})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": msg})
	}
}
