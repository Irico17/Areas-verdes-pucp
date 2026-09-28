package controller

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"

	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/contracts"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/dto"
	domainErrors "github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/errors"
)

// IInventarioCampoController defines HTTP handler methods for frente 2B field inventory.
type IInventarioCampoController interface {
	ListarTachos(*gin.Context)
	GuardarTacho(*gin.Context)
	PatchTacho(*gin.Context)
	CSVTachos(*gin.Context)
	BajaTacho(*gin.Context)

	ListarBebederos(*gin.Context)
	GuardarBebedero(*gin.Context)
	PatchBebedero(*gin.Context)
	BajaBebedero(*gin.Context)

	ListarPuntos(*gin.Context)
	GuardarPunto(*gin.Context)
	PatchPunto(*gin.Context)
	BajaPunto(*gin.Context)
	FormatoPuntos(*gin.Context)

	ListarReservas(*gin.Context)
	GuardarReserva(*gin.Context)
	PatchReserva(*gin.Context)
	BajaReserva(*gin.Context)

	ListarCapa(*gin.Context)
	GuardarCapa(*gin.Context)
	PatchCapa(*gin.Context)
	CSVCapa(*gin.Context)
	BajaCapa(*gin.Context)
}

type inventarioCampoController struct {
	uc     contracts.IInventarioCampoUseCase
	logger zerolog.Logger
}

// NewInventarioCampoController creates a new IInventarioCampoController instance.
func NewInventarioCampoController(
	uc contracts.IInventarioCampoUseCase,
	logger zerolog.Logger,
) IInventarioCampoController {
	return &inventarioCampoController{
		uc:     uc,
		logger: logger,
	}
}

func idParam(c *gin.Context) (int64, bool) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id < 1 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id inválido"})
		return 0, false
	}
	return id, true
}

func cuerpoJSON(c *gin.Context, dest any) ([]byte, bool) {
	raw, err := io.ReadAll(c.Request.Body)
	if err != nil || json.Unmarshal(raw, dest) != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "JSON inválido"})
		return nil, false
	}
	return raw, true
}

// ListarTachos godoc
// @Summary List waste bins
// @Tags inventario
// @Produce json
// @Success 200 {object} dto.ListarTachosResponseDTO
// @Router /v1/inventario/tachos [get]
func (ctrl *inventarioCampoController) ListarTachos(c *gin.Context) {
	res, err := ctrl.uc.ListarTachos(c.Request.Context())
	if err != nil {
		ctrl.logger.Error().Err(err).Msg("inventario: error al leer tachos")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "no se pudo leer los tachos"})
		return
	}
	c.JSON(http.StatusOK, res)
}

// GuardarTacho godoc
// @Summary Create or upsert a waste bin
// @Tags inventario
// @Accept json
// @Produce json
// @Success 201 {object} dto.TachoDTO
// @Router /v1/inventario/tachos [post]
func (ctrl *inventarioCampoController) GuardarTacho(c *gin.Context) {
	var body dto.TachoDTO
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "JSON inválido"})
		return
	}
	item, err := ctrl.uc.GuardarTacho(c.Request.Context(), body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no se pudo guardar el tacho"})
		return
	}
	c.JSON(http.StatusCreated, item)
}

// PatchTacho godoc
// @Summary Partially update a waste bin
// @Tags inventario
// @Accept json
// @Produce json
// @Success 200 {object} dto.TachoDTO
// @Router /v1/inventario/tachos/{id} [patch]
func (ctrl *inventarioCampoController) PatchTacho(c *gin.Context) {
	id, ok := idParam(c)
	if !ok {
		return
	}
	var body dto.TachoDTO
	raw, okBody := cuerpoJSON(c, &body)
	if !okBody {
		return
	}
	item, err := ctrl.uc.ActualizarTacho(c.Request.Context(), id, body, raw)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no se pudo actualizar el tacho"})
		return
	}
	c.JSON(http.StatusOK, item)
}

// CSVTachos godoc
// @Summary Export waste bins as CSV
// @Tags inventario
// @Produce text/csv
// @Success 200 {string} string "CSV"
// @Router /v1/inventario/tachos.csv [get]
func (ctrl *inventarioCampoController) CSVTachos(c *gin.Context) {
	csvData, err := ctrl.uc.CSVTachos(c.Request.Context())
	if err != nil {
		ctrl.logger.Error().Err(err).Msg("inventario: error al exportar tachos.csv")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "no se pudo exportar"})
		return
	}
	c.Header("Content-Type", "text/csv; charset=utf-8")
	c.String(http.StatusOK, csvData)
}

// BajaTacho godoc
// @Summary Soft delete a waste bin
// @Tags inventario
// @Produce json
// @Success 200 {object} dto.BajaResponseDTO
// @Router /v1/inventario/tachos/{id} [delete]
func (ctrl *inventarioCampoController) BajaTacho(c *gin.Context) {
	id, ok := idParam(c)
	if !ok {
		return
	}
	if err := ctrl.uc.BajaTacho(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "no se encontró el registro"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"activo": false})
}

// ListarBebederos godoc
// @Summary List drinking fountains
// @Tags inventario
// @Produce json
// @Success 200 {object} dto.ListarBebederosResponseDTO
// @Router /v1/inventario/bebederos [get]
func (ctrl *inventarioCampoController) ListarBebederos(c *gin.Context) {
	res, err := ctrl.uc.ListarBebederos(c.Request.Context())
	if err != nil {
		ctrl.logger.Error().Err(err).Msg("inventario: error al leer bebederos")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "no se pudo leer los bebederos"})
		return
	}
	c.JSON(http.StatusOK, res)
}

// GuardarBebedero godoc
// @Summary Create or upsert a drinking fountain
// @Tags inventario
// @Accept json
// @Produce json
// @Success 201 {object} dto.BebederoDTO
// @Router /v1/inventario/bebederos [post]
func (ctrl *inventarioCampoController) GuardarBebedero(c *gin.Context) {
	var body dto.BebederoDTO
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "JSON inválido"})
		return
	}
	item, err := ctrl.uc.GuardarBebedero(c.Request.Context(), body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no se pudo guardar el bebedero"})
		return
	}
	c.JSON(http.StatusCreated, item)
}

// PatchBebedero godoc
// @Summary Partially update a drinking fountain
// @Tags inventario
// @Accept json
// @Produce json
// @Success 200 {object} dto.BebederoDTO
// @Router /v1/inventario/bebederos/{id} [patch]
func (ctrl *inventarioCampoController) PatchBebedero(c *gin.Context) {
	id, ok := idParam(c)
	if !ok {
		return
	}
	var body dto.BebederoDTO
	raw, okBody := cuerpoJSON(c, &body)
	if !okBody {
		return
	}
	item, err := ctrl.uc.ActualizarBebedero(c.Request.Context(), id, body, raw)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no se pudo actualizar el bebedero"})
		return
	}
	c.JSON(http.StatusOK, item)
}

// BajaBebedero godoc
// @Summary Soft delete a drinking fountain
// @Tags inventario
// @Produce json
// @Success 200 {object} dto.BajaResponseDTO
// @Router /v1/inventario/bebederos/{id} [delete]
func (ctrl *inventarioCampoController) BajaBebedero(c *gin.Context) {
	id, ok := idParam(c)
	if !ok {
		return
	}
	if err := ctrl.uc.BajaBebedero(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "no se encontró el registro"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"activo": false})
}

// ListarPuntos godoc
// @Summary List campus points of interest
// @Tags inventario
// @Produce json
// @Param q query string false "Search query"
// @Success 200 {object} dto.ListarPuntosResponseDTO
// @Router /v1/inventario/puntos [get]
func (ctrl *inventarioCampoController) ListarPuntos(c *gin.Context) {
	res, err := ctrl.uc.ListarPuntos(c.Request.Context(), c.Query("q"))
	if err != nil {
		ctrl.logger.Error().Err(err).Msg("inventario: error al leer puntos")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "no se pudo leer los puntos"})
		return
	}
	c.JSON(http.StatusOK, res)
}

// GuardarPunto godoc
// @Summary Create or upsert a campus point of interest
// @Tags inventario
// @Accept json
// @Produce json
// @Success 201 {object} dto.PuntoDTO
// @Router /v1/inventario/puntos [post]
func (ctrl *inventarioCampoController) GuardarPunto(c *gin.Context) {
	raw, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "JSON inválido"})
		return
	}
	var body dto.PuntoDTO
	if err := json.Unmarshal(raw, &body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "JSON inválido"})
		return
	}
	item, err := ctrl.uc.GuardarPunto(c.Request.Context(), body, raw)
	if err != nil {
		if errors.Is(err, domainErrors.ErrPuntoContacto) {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":             "no se guardan teléfono, placeId ni website",
				"columnas_omitidas": []string{"phone", "placeId", "website"},
			})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": "no se pudo guardar el punto"})
		return
	}
	c.JSON(http.StatusCreated, item)
}

// PatchPunto godoc
// @Summary Partially update a campus point of interest
// @Tags inventario
// @Accept json
// @Produce json
// @Success 200 {object} dto.PuntoDTO
// @Router /v1/inventario/puntos/{id} [patch]
func (ctrl *inventarioCampoController) PatchPunto(c *gin.Context) {
	id, ok := idParam(c)
	if !ok {
		return
	}
	raw, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "JSON inválido"})
		return
	}
	var body dto.PuntoDTO
	if err := json.Unmarshal(raw, &body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "JSON inválido"})
		return
	}
	item, err := ctrl.uc.ActualizarPunto(c.Request.Context(), id, body, raw)
	if err != nil {
		if errors.Is(err, domainErrors.ErrPuntoContacto) {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":             "no se guardan teléfono, placeId ni website",
				"columnas_omitidas": []string{"phone", "placeId", "website"},
			})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": "no se pudo actualizar el punto"})
		return
	}
	c.JSON(http.StatusOK, item)
}

// BajaPunto godoc
// @Summary Soft delete a campus point of interest
// @Tags inventario
// @Produce json
// @Success 200 {object} dto.BajaResponseDTO
// @Router /v1/inventario/puntos/{id} [delete]
func (ctrl *inventarioCampoController) BajaPunto(c *gin.Context) {
	id, ok := idParam(c)
	if !ok {
		return
	}
	if err := ctrl.uc.BajaPunto(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "no se encontró el registro"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"activo": false})
}

// FormatoPuntos godoc
// @Summary Parse and validate puntos PUCP CSV file
// @Tags inventario
// @Produce json
// @Success 200 {object} dto.FormatoPuntosResponseDTO
// @Router /v1/inventario/formato/puntos [post]
func (ctrl *inventarioCampoController) FormatoPuntos(c *gin.Context) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "CSV inválido"})
		return
	}
	res, err := ctrl.uc.FormatoPuntos(c.Request.Context(), body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "CSV inválido"})
		return
	}
	c.JSON(http.StatusOK, res)
}

// ListarReservas godoc
// @Summary List garden reservations
// @Tags inventario
// @Produce json
// @Param desde query string false "Start date"
// @Param hasta query string false "End date"
// @Success 200 {object} dto.ListarReservasResponseDTO
// @Router /v1/inventario/reservas [get]
func (ctrl *inventarioCampoController) ListarReservas(c *gin.Context) {
	res, err := ctrl.uc.ListarReservas(c.Request.Context(), c.Query("desde"), c.Query("hasta"))
	if err != nil {
		ctrl.logger.Error().Err(err).Msg("inventario: error al leer reservas")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "no se pudo leer las reservas"})
		return
	}
	c.JSON(http.StatusOK, res)
}

// GuardarReserva godoc
// @Summary Create a garden reservation
// @Tags inventario
// @Accept json
// @Produce json
// @Success 201 {object} dto.ReservaDTO
// @Router /v1/inventario/reservas [post]
func (ctrl *inventarioCampoController) GuardarReserva(c *gin.Context) {
	var body dto.ReservaDTO
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "JSON inválido"})
		return
	}
	if body.Origen != "" && body.Origen != "ficticio" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "mientras la hoja responda 401 solo se acepta origen ficticio"})
		return
	}
	item, err := ctrl.uc.GuardarReserva(c.Request.Context(), body)
	if err != nil {
		if errors.Is(err, domainErrors.ErrReservaOrigen) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "mientras la hoja responda 401 solo se acepta origen ficticio"})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": "no se pudo guardar la reserva"})
		return
	}
	c.JSON(http.StatusCreated, item)
}

// PatchReserva godoc
// @Summary Partially update a garden reservation
// @Tags inventario
// @Accept json
// @Produce json
// @Success 200 {object} dto.ReservaDTO
// @Router /v1/inventario/reservas/{id} [patch]
func (ctrl *inventarioCampoController) PatchReserva(c *gin.Context) {
	id, ok := idParam(c)
	if !ok {
		return
	}
	var body dto.ReservaDTO
	raw, okBody := cuerpoJSON(c, &body)
	if !okBody {
		return
	}
	if body.Origen != "" && body.Origen != "ficticio" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "mientras la hoja responda 401 solo se acepta origen ficticio"})
		return
	}
	item, err := ctrl.uc.ActualizarReserva(c.Request.Context(), id, body, raw)
	if err != nil {
		if errors.Is(err, domainErrors.ErrReservaOrigen) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "mientras la hoja responda 401 solo se acepta origen ficticio"})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": "no se pudo actualizar la reserva"})
		return
	}
	c.JSON(http.StatusOK, item)
}

// BajaReserva godoc
// @Summary Soft delete a garden reservation
// @Tags inventario
// @Produce json
// @Success 200 {object} dto.BajaResponseDTO
// @Router /v1/inventario/reservas/{id} [delete]
func (ctrl *inventarioCampoController) BajaReserva(c *gin.Context) {
	id, ok := idParam(c)
	if !ok {
		return
	}
	if err := ctrl.uc.BajaReserva(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "no se encontró el registro"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"activo": false})
}

// ListarCapa godoc
// @Summary List records from an editable layer
// @Tags inventario
// @Produce json
// @Param capa path string true "Layer name"
// @Success 200 {object} dto.ListarFichasCapaResponseDTO
// @Router /v1/inventario/capas/{capa} [get]
func (ctrl *inventarioCampoController) ListarCapa(c *gin.Context) {
	res, err := ctrl.uc.ListarCapa(c.Request.Context(), c.Param("capa"))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "capa desconocida"})
		return
	}
	c.JSON(http.StatusOK, res)
}

// GuardarCapa godoc
// @Summary Create or upsert a record in an editable layer
// @Tags inventario
// @Accept json
// @Produce json
// @Param capa path string true "Layer name"
// @Success 201 {object} dto.FichaCapaDTO
// @Router /v1/inventario/capas/{capa} [post]
func (ctrl *inventarioCampoController) GuardarCapa(c *gin.Context) {
	var body dto.FichaCapaDTO
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "JSON inválido"})
		return
	}
	item, err := ctrl.uc.GuardarCapa(c.Request.Context(), c.Param("capa"), body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no se pudo guardar la ficha"})
		return
	}
	c.JSON(http.StatusCreated, item)
}

// PatchCapa godoc
// @Summary Partially update a record in an editable layer
// @Tags inventario
// @Accept json
// @Produce json
// @Param capa path string true "Layer name"
// @Success 200 {object} dto.FichaCapaDTO
// @Router /v1/inventario/capas/{capa}/{id} [patch]
func (ctrl *inventarioCampoController) PatchCapa(c *gin.Context) {
	id, ok := idParam(c)
	if !ok {
		return
	}
	var body dto.FichaCapaDTO
	raw, okBody := cuerpoJSON(c, &body)
	if !okBody {
		return
	}
	item, err := ctrl.uc.ActualizarCapa(c.Request.Context(), c.Param("capa"), id, body, raw)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no se pudo actualizar la ficha"})
		return
	}
	c.JSON(http.StatusOK, item)
}

// CSVCapa godoc
// @Summary Export an editable layer as CSV
// @Tags inventario
// @Produce text/csv
// @Param capa path string true "Layer name"
// @Success 200 {string} string "CSV"
// @Router /v1/inventario/export/{capa} [get]
func (ctrl *inventarioCampoController) CSVCapa(c *gin.Context) {
	csvData, err := ctrl.uc.CSVCapa(c.Request.Context(), c.Param("capa"))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "capa desconocida"})
		return
	}
	c.Header("Content-Type", "text/csv; charset=utf-8")
	c.String(http.StatusOK, csvData)
}

// BajaCapa godoc
// @Summary Soft delete a record in an editable layer
// @Tags inventario
// @Produce json
// @Param capa path string true "Layer name"
// @Success 200 {object} dto.BajaResponseDTO
// @Router /v1/inventario/capas/{capa}/{id} [delete]
func (ctrl *inventarioCampoController) BajaCapa(c *gin.Context) {
	id, ok := idParam(c)
	if !ok {
		return
	}
	if err := ctrl.uc.BajaCapa(c.Request.Context(), c.Param("capa"), id); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "no se encontró el registro"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"activo": false})
}
