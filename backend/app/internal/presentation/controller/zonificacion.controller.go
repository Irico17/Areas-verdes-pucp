package controller

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"

	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/contracts"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/dto"
	apperrors "github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/errors"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/presentation/middleware"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/presentation/requests"
)

// IZonificacionController defines HTTP handlers for capataz sectors, places, roads and quarters.
type IZonificacionController interface {
	ListarSectores(c *gin.Context)
	CrearSector(c *gin.Context)
	ActualizarSector(c *gin.Context)
	DesactivarSector(c *gin.Context)
	ImportarSectores(c *gin.Context)
	ListarLugares(c *gin.Context)
	ResolverLugar(c *gin.Context)
	Vias(c *gin.Context)
	ImportarVias(c *gin.Context)
	Cuarteles(c *gin.Context)
	Edificios(c *gin.Context)
	CrearReferente(c *gin.Context)
}

type zonificacionController struct {
	uc     contracts.IZonificacionUseCase
	logger zerolog.Logger
}

// NewZonificacionController creates the zoning controller.
func NewZonificacionController(uc contracts.IZonificacionUseCase, logger zerolog.Logger) IZonificacionController {
	return &zonificacionController{uc: uc, logger: logger}
}

// ListarSectores godoc
// @Summary Listar sectores de capataz
// @Description El color del mapa sale de este catálogo
// @Tags zonificacion
// @Produce json
// @Param activos query string false "1 para solo activos"
// @Success 200 {object} dto.SectorListDTO
// @Failure 401 {object} map[string]string
// @Router /v1/zonificacion/sectores [get]
func (ctrl *zonificacionController) ListarSectores(c *gin.Context) {
	res, err := ctrl.uc.ListarSectores(c.Request.Context(), c.Query("activos") == "1")
	if err != nil {
		ctrl.fallo(c, err, "no se pudo leer los sectores de capataz")
		return
	}
	c.JSON(http.StatusOK, res)
}

// CrearSector godoc
// @Summary Alta de un sector de capataz
// @Description Crea el sector. Un código repetido responde 409 y no duplica la fila.
// @Tags zonificacion
// @Accept json
// @Produce json
// @Param body body requests.CrearSectorRequest true "Sector"
// @Success 201 {object} dto.SectorCapatazDTO
// @Failure 400 {object} map[string]string
// @Failure 409 {object} map[string]string
// @Router /v1/zonificacion/sectores [post]
func (ctrl *zonificacionController) CrearSector(c *gin.Context) {
	var req requests.CrearSectorRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "JSON inválido"})
		return
	}
	item, err := ctrl.uc.CrearSector(c.Request.Context(), dto.CrearSectorDTO{
		Codigo:    req.Codigo,
		Nombre:    req.Nombre,
		Color:     req.Color,
		UsuarioID: usuarioID(c),
	})
	if err != nil {
		ctrl.fallo(c, err, "no se pudo guardar el sector de capataz")
		return
	}
	c.JSON(http.StatusCreated, item)
}

// ActualizarSector godoc
// @Summary Corregir nombre, color o vigencia de un sector de capataz
// @Tags zonificacion
// @Accept json
// @Produce json
// @Param codigo path string true "Código"
// @Param body body requests.ActualizarSectorRequest true "Cambios"
// @Success 200 {object} dto.SectorCapatazDTO
// @Failure 404 {object} map[string]string
// @Router /v1/zonificacion/sectores/{codigo} [patch]
func (ctrl *zonificacionController) ActualizarSector(c *gin.Context) {
	var req requests.ActualizarSectorRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "JSON inválido"})
		return
	}
	item, err := ctrl.uc.ActualizarSector(c.Request.Context(), dto.ActualizarSectorDTO{
		Codigo:    c.Param("codigo"),
		Nombre:    req.Nombre,
		Color:     req.Color,
		Activo:    req.Activo,
		UsuarioID: usuarioID(c),
	})
	if err != nil {
		ctrl.fallo(c, err, "no se pudo guardar el sector de capataz")
		return
	}
	c.JSON(http.StatusOK, item)
}

// DesactivarSector godoc
// @Summary Desactivar un sector de capataz
// @Description Baja lógica. No existe una ruta que borre la fila.
// @Tags zonificacion
// @Produce json
// @Param codigo path string true "Código"
// @Success 200 {object} map[string]any
// @Failure 404 {object} map[string]string
// @Router /v1/zonificacion/sectores/{codigo}/desactivar [post]
func (ctrl *zonificacionController) DesactivarSector(c *gin.Context) {
	codigo := c.Param("codigo")
	if err := ctrl.uc.DesactivarSector(c.Request.Context(), codigo, usuarioID(c)); err != nil {
		ctrl.fallo(c, err, "no se pudo desactivar el sector de capataz")
		return
	}
	c.JSON(http.StatusOK, gin.H{"codigo": codigo, "activo": false})
}

// ImportarSectores godoc
// @Summary Importar sectores de capataz
// @Description Un código que ya existe se actualiza. No se inserta una segunda fila.
// @Tags zonificacion
// @Accept json
// @Produce json
// @Param body body requests.ImportarSectoresRequest true "Filas"
// @Success 200 {object} dto.ImportacionSectorDTO
// @Router /v1/zonificacion/sectores/importar [post]
func (ctrl *zonificacionController) ImportarSectores(c *gin.Context) {
	var req requests.ImportarSectoresRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "JSON inválido"})
		return
	}
	filas := make([]dto.CrearSectorDTO, 0, len(req.Sectores))
	for _, fila := range req.Sectores {
		filas = append(filas, dto.CrearSectorDTO{Codigo: fila.Codigo, Nombre: fila.Nombre, Color: fila.Color})
	}
	res, err := ctrl.uc.ImportarSectores(c.Request.Context(), filas, usuarioID(c))
	if err != nil {
		ctrl.fallo(c, err, "no se pudo importar los sectores de capataz")
		return
	}
	c.JSON(http.StatusOK, res)
}

// ListarLugares godoc
// @Summary Lugares del catálogo
// @Tags zonificacion
// @Produce json
// @Success 200 {object} map[string]any
// @Router /v1/zonificacion/lugares [get]
func (ctrl *zonificacionController) ListarLugares(c *gin.Context) {
	rows, err := ctrl.uc.ListarLugares(c.Request.Context())
	if err != nil {
		ctrl.fallo(c, err, "no se pudo leer los lugares")
		return
	}
	c.JSON(http.StatusOK, gin.H{"lugares": rows})
}

// ResolverLugar godoc
// @Summary Resolver un lugar del catálogo
// @Description Un nombre que no existe no se inserta. lugar_libre solo se lee.
// @Tags zonificacion
// @Accept json
// @Produce json
// @Param body body requests.ResolverLugarRequest true "Lugar"
// @Success 200 {object} dto.LugarResueltoDTO
// @Failure 400 {object} map[string]string
// @Router /v1/zonificacion/lugares/resolver [post]
func (ctrl *zonificacionController) ResolverLugar(c *gin.Context) {
	var req requests.ResolverLugarRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "JSON inválido"})
		return
	}
	res, err := ctrl.uc.ResolverLugar(c.Request.Context(), dto.ResolverLugarDTO{
		LugarID:    req.LugarID,
		Nombre:     req.Nombre,
		LugarLibre: req.LugarLibre,
	})
	if err != nil {
		ctrl.fallo(c, err, "no se pudo resolver el lugar")
		return
	}
	c.JSON(http.StatusOK, res)
}

// Vias godoc
// @Summary Capa de vías
// @Description Vacía hasta que se importe un GeoJSON. No inventa filas.
// @Tags zonificacion
// @Produce json
// @Success 200 {object} map[string]any
// @Router /v1/zonificacion/vias [get]
func (ctrl *zonificacionController) Vias(c *gin.Context) {
	body, err := ctrl.uc.Vias(c.Request.Context())
	if err != nil {
		ctrl.fallo(c, err, "no se pudo leer las vías")
		return
	}
	c.Data(http.StatusOK, "application/geo+json; charset=utf-8", body)
}

// ImportarVias godoc
// @Summary Importar vías desde GeoJSON
// @Tags zonificacion
// @Accept json
// @Produce json
// @Success 200 {object} dto.ImportacionViaDTO
// @Failure 400 {object} map[string]string
// @Router /v1/zonificacion/vias/importar [post]
func (ctrl *zonificacionController) ImportarVias(c *gin.Context) {
	body, err := c.GetRawData()
	if err != nil || len(body) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "JSON inválido"})
		return
	}
	res, err := ctrl.uc.ImportarVias(c.Request.Context(), body, usuarioID(c))
	if err != nil {
		ctrl.fallo(c, err, "no se pudo importar las vías")
		return
	}
	c.JSON(http.StatusOK, res)
}

// Cuarteles godoc
// @Summary Cuarteles históricos
// @Description Solo lectura. Sin shape responde el aviso y ninguna geometría.
// @Tags zonificacion
// @Produce json
// @Success 200 {object} map[string]any
// @Router /v1/zonificacion/cuarteles [get]
func (ctrl *zonificacionController) Cuarteles(c *gin.Context) {
	body, err := ctrl.uc.Cuarteles(c.Request.Context())
	if err != nil {
		ctrl.fallo(c, err, "no se pudo leer los cuarteles")
		return
	}
	c.Data(http.StatusOK, "application/geo+json; charset=utf-8", body)
}

// Edificios godoc
// @Summary Ids de edificio para elegir un referente
// @Tags zonificacion
// @Produce json
// @Success 200 {object} map[string]any
// @Router /v1/zonificacion/edificios [get]
func (ctrl *zonificacionController) Edificios(c *gin.Context) {
	rows, err := ctrl.uc.Edificios(c.Request.Context())
	if err != nil {
		ctrl.fallo(c, err, "no se pudo leer los edificios")
		return
	}
	c.JSON(http.StatusOK, gin.H{"edificios": rows})
}

// CrearReferente godoc
// @Summary Guardar un edificio como referente, por id
// @Tags zonificacion
// @Accept json
// @Produce json
// @Param body body requests.CrearReferenteRequest true "Par lugar y edificio"
// @Success 201 {object} dto.ReferenteDTO
// @Failure 400 {object} map[string]string
// @Router /v1/zonificacion/referentes [post]
func (ctrl *zonificacionController) CrearReferente(c *gin.Context) {
	var req requests.CrearReferenteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "JSON inválido"})
		return
	}
	item, err := ctrl.uc.CrearReferente(c.Request.Context(), dto.CrearReferenteDTO{
		LugarID:    req.LugarID,
		EdificioID: req.EdificioID,
		UsuarioID:  usuarioID(c),
	})
	if err != nil {
		ctrl.fallo(c, err, "no se pudo guardar el referente")
		return
	}
	status := http.StatusOK
	if item.Creado {
		status = http.StatusCreated
	}
	c.JSON(status, item)
}

func (ctrl *zonificacionController) fallo(c *gin.Context, err error, generico string) {
	switch {
	case errors.Is(err, apperrors.ErrSectorDuplicado):
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
	case errors.Is(err, apperrors.ErrSectorNoExiste):
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
	case errors.Is(err, apperrors.ErrCodigoSector),
		errors.Is(err, apperrors.ErrColorSector),
		errors.Is(err, apperrors.ErrNombreObligatorio),
		errors.Is(err, apperrors.ErrLugarDesconocido),
		errors.Is(err, apperrors.ErrEdificioDesconocido),
		errors.Is(err, apperrors.ErrViaVacia),
		errors.Is(err, apperrors.ErrArchivoSinFilas),
		errors.Is(err, apperrors.ErrGeomVia):
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	default:
		ctrl.logger.Error().Err(err).Msg("zonificacion")
		c.JSON(http.StatusInternalServerError, gin.H{"error": generico})
	}
}

func usuarioID(c *gin.Context) int64 {
	u, ok := middleware.UsuarioEn(c)
	if !ok {
		return 0
	}
	return u.ID
}
