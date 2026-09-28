package controller

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"

	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/contracts"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/dto"
	apperrors "github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/errors"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/presentation/requests"
)

// ICatalogoController defines HTTP handlers for catalog management.
type ICatalogoController interface {
	Listar(c *gin.Context)
	Crear(c *gin.Context)
	Desactivar(c *gin.Context)
}

type catalogoController struct {
	catalogoUC contracts.ICatalogoUseCase
	logger     zerolog.Logger
}

// NewCatalogoController creates a new catalog controller.
func NewCatalogoController(catalogoUC contracts.ICatalogoUseCase, logger zerolog.Logger) ICatalogoController {
	return &catalogoController{catalogoUC: catalogoUC, logger: logger}
}

// Listar handles GET /catalogos.
// @Summary Listar ítems de catálogo
// @Description Devuelve ítems de catálogo con filtro opcional por clase y solo activos
// @Tags Catálogos
// @Produce json
// @Param clase query string false "Clase de catálogo"
// @Param activos query string false "Solo activos si es 1"
// @Success 200 {object} dto.CatalogoListResponseDTO
// @Failure 401 {object} map[string]string "inicie sesión"
// @Failure 403 {object} map[string]string "su rol no tiene ese permiso"
// @Failure 500 {object} map[string]string "no se pudo leer el catálogo"
// @Failure 503 {object} map[string]string "base de datos no disponible"
// @Router /v1/catalogos [get]
func (ctrl *catalogoController) Listar(c *gin.Context) {
	filtro := dto.FiltroCatalogoDTO{
		Clase:       c.Query("clase"),
		SoloActivos: c.Query("activos") == "1",
	}

	res, err := ctrl.catalogoUC.Listar(c.Request.Context(), filtro)
	if err != nil {
		ctrl.logger.Error().Err(err).Msg("catalogo: error al listar catalogo")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "no se pudo leer el catálogo"})
		return
	}

	c.JSON(http.StatusOK, res)
}

// Crear handles POST /catalogos.
// @Summary Crear o reactivar un ítem de catálogo
// @Description Crea un ítem nuevo o actualiza/reactiva uno existente (requiere permiso catalogos)
// @Tags Catálogos
// @Accept json
// @Produce json
// @Param body body requests.CrearCatalogoRequest true "Datos del ítem"
// @Success 201 {object} dto.CatalogoItemDTO
// @Failure 400 {object} map[string]string "error de validación o JSON inválido"
// @Failure 401 {object} map[string]string "inicie sesión"
// @Failure 403 {object} map[string]string "su rol no tiene ese permiso"
// @Failure 500 {object} map[string]string "no se pudo guardar el ítem"
// @Failure 503 {object} map[string]string "base de datos no disponible"
// @Router /v1/catalogos [post]
func (ctrl *catalogoController) Crear(c *gin.Context) {
	var req requests.CrearCatalogoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "JSON inválido"})
		return
	}

	item, err := ctrl.catalogoUC.Crear(c.Request.Context(), dto.CrearCatalogoDTO{
		Clase:  req.Clase,
		Codigo: req.Codigo,
		Nombre: req.Nombre,
	})
	if err != nil {
		if errors.Is(err, apperrors.ErrClaseNoReconocida) ||
			errors.Is(err, apperrors.ErrCodigoInvalido) ||
			errors.Is(err, apperrors.ErrNombreObligatorio) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		ctrl.logger.Error().Err(err).Msg("catalogo: error al crear item")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "no se pudo guardar el ítem"})
		return
	}

	c.JSON(http.StatusCreated, item)
}

// Desactivar handles POST /catalogos/:id/desactivar.
// @Summary Desactivar un ítem de catálogo (baja lógica)
// @Description Realiza una baja lógica cambiando activo=false sin borrar la fila
// @Tags Catálogos
// @Produce json
// @Param id path int true "ID del ítem"
// @Success 200 {object} dto.DesactivarCatalogoResponseDTO
// @Failure 400 {object} map[string]string "id inválido"
// @Failure 401 {object} map[string]string "inicie sesión"
// @Failure 403 {object} map[string]string "su rol no tiene ese permiso"
// @Failure 404 {object} map[string]string "no existe ese ítem"
// @Failure 500 {object} map[string]string "no se pudo desactivar"
// @Failure 503 {object} map[string]string "base de datos no disponible"
// @Router /v1/catalogos/{id}/desactivar [post]
func (ctrl *catalogoController) Desactivar(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id inválido"})
		return
	}

	res, err := ctrl.catalogoUC.Desactivar(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, apperrors.ErrItemNoExiste) {
			c.JSON(http.StatusNotFound, gin.H{"error": "no existe ese ítem"})
			return
		}
		ctrl.logger.Error().Err(err).Msg("catalogo: error al desactivar item")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "no se pudo desactivar"})
		return
	}

	c.JSON(http.StatusOK, res)
}
