package controller

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"

	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/contracts"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/dto"
	domainErrors "github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/errors"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/presentation/middleware"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/presentation/requests"
)

// ICatastroController defines HTTP handlers for master cadastral resources.
type ICatastroController interface {
	Zonas(*gin.Context)
	CrearZona(*gin.Context)
	ActualizarZona(*gin.Context)
	BajaZona(*gin.Context)
	Poligonos(*gin.Context)
	Cuadrillas(*gin.Context)
	CrearCuadrilla(*gin.Context)
	Lugares(*gin.Context)
	CrearLugar(*gin.Context)
	Especies(*gin.Context)
	CrearEspecie(*gin.Context)
	Ejemplares(*gin.Context)
	CrearEjemplar(*gin.Context)
	Codigos(*gin.Context)
	Recodificar(*gin.Context)
	Fauna(*gin.Context)
	Puertas(*gin.Context)
	Playas(*gin.Context)
	Veredas(*gin.Context)
	Xerofiticas(*gin.Context)
	JardinesReserva(*gin.Context)
}

type catastroController struct {
	zonaUC      contracts.IZonaSupervisionUseCase
	cuadrillaUC contracts.ICuadrillaUseCase
	lugarUC     contracts.ILugarUseCase
	especieUC   contracts.IEspecieUseCase
	ejemplarUC  contracts.IEjemplarUseCase
	refUC       contracts.ICatastroReferenciaUseCase
	logger      zerolog.Logger
}

// NewCatastroController creates a new ICatastroController instance.
func NewCatastroController(
	zonaUC contracts.IZonaSupervisionUseCase,
	cuadrillaUC contracts.ICuadrillaUseCase,
	lugarUC contracts.ILugarUseCase,
	especieUC contracts.IEspecieUseCase,
	ejemplarUC contracts.IEjemplarUseCase,
	refUC contracts.ICatastroReferenciaUseCase,
	logger zerolog.Logger,
) ICatastroController {
	return &catastroController{
		zonaUC:      zonaUC,
		cuadrillaUC: cuadrillaUC,
		lugarUC:     lugarUC,
		especieUC:   especieUC,
		ejemplarUC:  ejemplarUC,
		refUC:       refUC,
		logger:      logger,
	}
}

func pagina(limitRaw, offsetRaw string) (int, int) {
	limit, _ := strconv.Atoi(limitRaw)
	offset, _ := strconv.Atoi(offsetRaw)
	if limitRaw == "" {
		return 0, 0
	}
	if limit < 1 {
		limit = 1
	}
	if limit > 2000 {
		limit = 2000
	}
	if offset < 0 {
		offset = 0
	}
	return limit, offset
}

// Zonas godoc
// @Summary List supervision zones
// @Tags catastro
// @Produce json
// @Success 200 {object} map[string][]dto.ZonaSupervisionDTO
// @Router /v1/catastro/zonas-supervision [get]
func (ctrl *catastroController) Zonas(c *gin.Context) {
	rows, err := ctrl.zonaUC.Listar(c.Request.Context())
	if err != nil {
		ctrl.logger.Error().Err(err).Msg("catastro: error al leer zonas")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "no se pudo leer las zonas de supervisión"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"zonas_supervision": rows})
}

// CrearZona godoc
// @Summary Create supervision zone
// @Tags catastro
// @Accept json
// @Produce json
// @Param body body requests.CrearZonaSupervisionRequest true "Supervision zone data"
// @Success 201 {object} dto.ZonaSupervisionDTO
// @Failure 400 {object} map[string]string
// @Router /v1/catastro/zonas-supervision [post]
func (ctrl *catastroController) CrearZona(c *gin.Context) {
	var body requests.CrearZonaSupervisionRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "JSON inválido"})
		return
	}
	item, err := ctrl.zonaUC.Crear(c.Request.Context(), dto.CrearZonaSupervisionDTO{
		Codigo:  body.Codigo,
		Nombre:  body.Nombre,
		AreaM2:  body.AreaM2,
		GeoJSON: body.GeoJSON,
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no se pudo crear la zona de supervisión"})
		return
	}
	c.JSON(http.StatusCreated, item)
}

// ActualizarZona godoc
// @Summary Update a supervision zone
// @Tags catastro
// @Accept json
// @Produce json
// @Param codigo path string true "Zone code"
// @Param body body requests.ActualizarZonaSupervisionRequest true "Supervision zone data"
// @Success 200 {object} dto.ZonaSupervisionDTO
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /v1/catastro/zonas-supervision/{codigo} [patch]
func (ctrl *catastroController) ActualizarZona(c *gin.Context) {
	var body requests.ActualizarZonaSupervisionRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "JSON inválido"})
		return
	}
	codigo := strings.TrimSpace(c.Param("codigo"))
	if body.Codigo != "" && strings.TrimSpace(body.Codigo) != codigo {
		c.JSON(http.StatusBadRequest, gin.H{"error": "el código de la ruta no coincide con el cuerpo"})
		return
	}
	geojson, err := geometriaZona(body.Geom, body.GeoJSON)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "JSON inválido"})
		return
	}
	var userID *int64
	if u, ok := middleware.UsuarioEn(c); ok && u.ID > 0 {
		userID = &u.ID
	}
	item, err := ctrl.zonaUC.Actualizar(c.Request.Context(), codigo, dto.ActualizarZonaSupervisionDTO{
		Nombre:  body.Nombre,
		AreaM2:  body.AreaM2,
		GeoJSON: geojson,
	}, userID)
	if err != nil {
		if errors.Is(err, domainErrors.ErrNoEncontrado) {
			c.JSON(http.StatusNotFound, gin.H{"error": "zona de supervisión no encontrada"})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": "no se pudo actualizar la zona de supervisión"})
		return
	}
	c.JSON(http.StatusOK, item)
}

// BajaZona godoc
// @Summary Logical deactivation of a supervision zone
// @Tags catastro
// @Produce json
// @Param codigo path string true "Zone code"
// @Success 200 {object} dto.ZonaSupervisionDTO
// @Failure 404 {object} map[string]string
// @Router /v1/catastro/zonas-supervision/{codigo}/baja [post]
func (ctrl *catastroController) BajaZona(c *gin.Context) {
	var userID *int64
	if u, ok := middleware.UsuarioEn(c); ok && u.ID > 0 {
		userID = &u.ID
	}
	item, err := ctrl.zonaUC.Baja(c.Request.Context(), c.Param("codigo"), userID)
	if err != nil {
		if errors.Is(err, domainErrors.ErrNoEncontrado) {
			c.JSON(http.StatusNotFound, gin.H{"error": "zona de supervisión no encontrada"})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": "no se pudo dar de baja la zona de supervisión"})
		return
	}
	c.JSON(http.StatusOK, item)
}

func geometriaZona(geom json.RawMessage, geojson string) (string, error) {
	geom = bytes.TrimSpace(geom)
	if len(geom) > 0 && string(geom) != "null" {
		if !json.Valid(geom) {
			return "", domainErrors.ErrEntrada
		}
		return string(geom), nil
	}
	return strings.TrimSpace(geojson), nil
}

// Poligonos godoc
// @Summary List work team sector polygons
// @Tags catastro
// @Produce json
// @Success 200 {object} map[string][]dto.PoligonoCuadrillaDTO
// @Router /v1/catastro/poligonos [get]
func (ctrl *catastroController) Poligonos(c *gin.Context) {
	rows, err := ctrl.refUC.ListarPoligonos(c.Request.Context())
	if err != nil {
		ctrl.logger.Error().Err(err).Msg("catastro: error al leer poligonos")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "no se pudo leer los polígonos"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"poligonos": rows})
}

// Cuadrillas godoc
// @Summary List work teams
// @Tags catastro
// @Produce json
// @Success 200 {object} map[string][]dto.CuadrillaDTO
// @Router /v1/catastro/cuadrillas [get]
func (ctrl *catastroController) Cuadrillas(c *gin.Context) {
	rows, err := ctrl.cuadrillaUC.Listar(c.Request.Context())
	if err != nil {
		ctrl.logger.Error().Err(err).Msg("catastro: error al leer cuadrillas")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "no se pudo leer las cuadrillas"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"cuadrillas": rows})
}

// CrearCuadrilla godoc
// @Summary Create work team
// @Tags catastro
// @Accept json
// @Produce json
// @Param body body requests.CrearCuadrillaRequest true "Work team data"
// @Success 201 {object} dto.CuadrillaDTO
// @Failure 400 {object} map[string]string
// @Router /v1/catastro/cuadrillas [post]
func (ctrl *catastroController) CrearCuadrilla(c *gin.Context) {
	var body requests.CrearCuadrillaRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "JSON inválido"})
		return
	}
	item, err := ctrl.cuadrillaUC.Crear(c.Request.Context(), dto.CrearCuadrillaDTO{
		ID:     body.ID,
		Nombre: body.Nombre,
		Turno:  body.Turno,
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no se pudo crear la cuadrilla"})
		return
	}
	c.JSON(http.StatusCreated, item)
}

// Lugares godoc
// @Summary List campus locations
// @Tags catastro
// @Produce json
// @Success 200 {object} map[string][]dto.LugarDTO
// @Router /v1/catastro/lugares [get]
func (ctrl *catastroController) Lugares(c *gin.Context) {
	rows, err := ctrl.lugarUC.Listar(c.Request.Context())
	if err != nil {
		ctrl.logger.Error().Err(err).Msg("catastro: error al leer lugares")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "no se pudo leer los lugares"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"lugares": rows})
}

// CrearLugar godoc
// @Summary Create campus location
// @Tags catastro
// @Accept json
// @Produce json
// @Param body body requests.CrearLugarRequest true "Place data"
// @Success 201 {object} dto.LugarDTO
// @Failure 400 {object} map[string]string
// @Router /v1/catastro/lugares [post]
func (ctrl *catastroController) CrearLugar(c *gin.Context) {
	var body requests.CrearLugarRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "JSON inválido"})
		return
	}
	item, err := ctrl.lugarUC.Crear(c.Request.Context(), dto.CrearLugarDTO{
		Nombre: body.Nombre,
		Lat:    body.Lat,
		Lon:    body.Lon,
		ZonaID: body.ZonaID,
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no se pudo crear el lugar"})
		return
	}
	c.JSON(http.StatusCreated, item)
}

// Especies godoc
// @Summary List botanical species
// @Tags catastro
// @Produce json
// @Success 200 {object} map[string][]dto.EspecieDTO
// @Router /v1/catastro/especies [get]
func (ctrl *catastroController) Especies(c *gin.Context) {
	rows, err := ctrl.especieUC.Listar(c.Request.Context())
	if err != nil {
		ctrl.logger.Error().Err(err).Msg("catastro: error al leer especies")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "no se pudo leer las especies"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"especies": rows})
}

// CrearEspecie godoc
// @Summary Create botanical species
// @Tags catastro
// @Accept json
// @Produce json
// @Param body body requests.CrearEspecieRequest true "Species data"
// @Success 201 {object} dto.EspecieDTO
// @Failure 400 {object} map[string]string
// @Router /v1/catastro/especies [post]
func (ctrl *catastroController) CrearEspecie(c *gin.Context) {
	var body requests.CrearEspecieRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "JSON inválido"})
		return
	}
	item, err := ctrl.especieUC.Crear(c.Request.Context(), dto.CrearEspecieDTO{
		Cientifico: body.Cientifico,
		Comun:      body.Comun,
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no se pudo crear la especie"})
		return
	}
	c.JSON(http.StatusCreated, item)
}

// Ejemplares godoc
// @Summary List flora specimens with pagination
// @Tags catastro
// @Produce json
// @Param limit query int false "Pagination limit"
// @Param offset query int false "Pagination offset"
// @Success 200 {object} dto.EjemplaresPaginadosDTO
// @Router /v1/catastro/ejemplares [get]
func (ctrl *catastroController) Ejemplares(c *gin.Context) {
	limit, offset := pagina(c.Query("limit"), c.Query("offset"))
	res, err := ctrl.ejemplarUC.Listar(c.Request.Context(), limit, offset)
	if err != nil {
		ctrl.logger.Error().Err(err).Msg("catastro: error al leer ejemplares")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "no se pudo leer los ejemplares"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"ejemplares": res.Ejemplares,
		"total":      res.Total,
		"limit":      res.Limit,
		"offset":     res.Offset,
	})
}

// CrearEjemplar godoc
// @Summary Create flora specimen
// @Tags catastro
// @Accept json
// @Produce json
// @Param body body requests.CrearEjemplarRequest true "Specimen data"
// @Success 201 {object} dto.EjemplarDTO
// @Failure 400 {object} map[string]string
// @Router /v1/catastro/ejemplares [post]
func (ctrl *catastroController) CrearEjemplar(c *gin.Context) {
	var body requests.CrearEjemplarRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "JSON inválido"})
		return
	}
	item, err := ctrl.ejemplarUC.Crear(c.Request.Context(), dto.EjemplarDTO{
		NumeroOrigen:       body.NumeroOrigen,
		Codigo:             body.Codigo,
		EspecieID:          body.EspecieID,
		NombreComun:        body.NombreComun,
		TipoVegetacion:     body.TipoVegetacion,
		Cantidad:           body.Cantidad,
		UbicacionLugarID:   body.UbicacionLugarID,
		Referencia:         body.Referencia,
		Lat:                body.Lat,
		Lon:                body.Lon,
		ObservacionFen2026: body.ObservacionFen2026,
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no se pudo crear el ejemplar"})
		return
	}
	c.JSON(http.StatusCreated, item)
}

// Codigos godoc
// @Summary List code history for specimen
// @Tags catastro
// @Produce json
// @Param id path int true "Specimen ID"
// @Success 200 {object} map[string][]dto.CodigoHistoricoDTO
// @Failure 400 {object} map[string]string
// @Router /v1/catastro/ejemplares/{id}/codigos [get]
func (ctrl *catastroController) Codigos(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id inválido"})
		return
	}
	rows, err := ctrl.ejemplarUC.ListarCodigos(c.Request.Context(), id)
	if err != nil {
		ctrl.logger.Error().Err(err).Int64("id", id).Msg("catastro: error al leer codigos")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "no se pudo leer el historial"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"codigos": rows})
}

// Recodificar godoc
// @Summary Recodify specimen
// @Tags catastro
// @Accept json
// @Produce json
// @Param id path int true "Specimen ID"
// @Param body body requests.RecodificarRequest true "New code"
// @Success 201 {object} dto.CodigoHistoricoDTO
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /v1/catastro/ejemplares/{id}/codigos [post]
func (ctrl *catastroController) Recodificar(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id inválido"})
		return
	}
	var body requests.RecodificarRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "JSON inválido"})
		return
	}
	item, err := ctrl.ejemplarUC.Recodificar(c.Request.Context(), id, dto.RecodificarDTO{Codigo: body.Codigo})
	if errors.Is(err, domainErrors.ErrNoEncontrado) {
		c.JSON(http.StatusNotFound, gin.H{"error": "ejemplar sin código anterior"})
		return
	}
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no se pudo recodificar"})
		return
	}
	c.JSON(http.StatusCreated, item)
}

func (ctrl *catastroController) capa(c *gin.Context, tabla string) {
	rows, err := ctrl.refUC.ListarCapa(c.Request.Context(), tabla)
	if err != nil {
		ctrl.logger.Error().Err(err).Str("capa", tabla).Msg("catastro: error al leer capa")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "no se pudo leer la capa"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"filas": rows})
}

// Fauna godoc
// @Summary List fauna reference layer
// @Tags catastro
// @Produce json
// @Success 200 {object} map[string][]dto.CapaFichaDTO
// @Router /v1/catastro/fauna [get]
func (ctrl *catastroController) Fauna(c *gin.Context) {
	ctrl.capa(c, "fauna")
}

// Puertas godoc
// @Summary List puertas reference layer
// @Tags catastro
// @Produce json
// @Success 200 {object} map[string][]dto.CapaFichaDTO
// @Router /v1/catastro/puertas [get]
func (ctrl *catastroController) Puertas(c *gin.Context) {
	ctrl.capa(c, "puertas")
}

// Playas godoc
// @Summary List playas reference layer
// @Tags catastro
// @Produce json
// @Success 200 {object} map[string][]dto.CapaFichaDTO
// @Router /v1/catastro/playas [get]
func (ctrl *catastroController) Playas(c *gin.Context) {
	ctrl.capa(c, "playas_estacionamiento")
}

// Veredas godoc
// @Summary List veredas de riesgo reference layer
// @Tags catastro
// @Produce json
// @Success 200 {object} map[string][]dto.CapaFichaDTO
// @Router /v1/catastro/veredas [get]
func (ctrl *catastroController) Veredas(c *gin.Context) {
	ctrl.capa(c, "veredas_riesgo")
}

// Xerofiticas godoc
// @Summary List xerofiticas reference layer
// @Tags catastro
// @Produce json
// @Success 200 {object} map[string][]dto.CapaFichaDTO
// @Router /v1/catastro/xerofiticas [get]
func (ctrl *catastroController) Xerofiticas(c *gin.Context) {
	ctrl.capa(c, "xerofiticas")
}

// JardinesReserva godoc
// @Summary List jardines de reserva reference layer
// @Tags catastro
// @Produce json
// @Success 200 {object} map[string][]dto.CapaFichaDTO
// @Router /v1/catastro/jardines-reserva [get]
func (ctrl *catastroController) JardinesReserva(c *gin.Context) {
	ctrl.capa(c, "jardines_reserva")
}
