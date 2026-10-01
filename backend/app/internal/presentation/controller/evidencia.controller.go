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
	domainErrors "github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/errors"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/presentation/middleware"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/presentation/requests"
)

// IEvidenciaController defines HTTP endpoints for evidencias.
type IEvidenciaController interface {
	List(*gin.Context)
	Upload(*gin.Context)
	File(*gin.Context)
}

type evidenciaController struct {
	uc     contracts.IEvidenciaUseCase
	logger zerolog.Logger
}

// NewEvidenciaController creates a new instance of IEvidenciaController.
func NewEvidenciaController(
	uc contracts.IEvidenciaUseCase,
	logger zerolog.Logger,
) IEvidenciaController {
	return &evidenciaController{
		uc:     uc,
		logger: logger,
	}
}

// List godoc
// @Summary List evidence attachments
// @Tags evidencias
// @Produce json
// @Param actividad_id query string false "Filter by activity UUID"
// @Success 200 {object} dto.ListarEvidenciasResponseDTO
// @Router /v1/evidencias [get]
func (ctrl *evidenciaController) List(c *gin.Context) {
	res, err := ctrl.uc.Listar(c.Request.Context(), c.Query("actividad_id"))
	if err != nil {
		ctrl.logger.Error().Err(err).Msg("evidencias")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "no se pudieron leer las evidencias"})
		return
	}
	c.JSON(http.StatusOK, res)
}

// Upload godoc
// @Summary Upload an evidence attachment
// @Tags evidencias
// @Accept multipart/form-data
// @Produce json
// @Param id formData string true "Evidence UUID"
// @Param actividad_id formData string true "Activity UUID"
// @Param orden_id formData string false "Order UUID"
// @Param nota formData string false "Note"
// @Param sha256 formData string false "Expected SHA256"
// @Param lat formData string false "Latitude"
// @Param lon formData string false "Longitude"
// @Param exif formData string false "EXIF JSON"
// @Param archivo formData file true "File to upload"
// @Success 201 {object} dto.SubirEvidenciaResponseDTO
// @Success 200 {object} dto.SubirEvidenciaResponseDTO
// @Router /v1/evidencias [post]
func (ctrl *evidenciaController) Upload(c *gin.Context) {
	u, ok := middleware.UsuarioEn(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "inicie sesión"})
		return
	}

	file, err := c.FormFile("archivo")
	if err != nil {
		writeOperacionErr(c, &ctrl.logger, domainErrors.InputError{Reason: "falta el archivo"})
		return
	}
	src, err := file.Open()
	if err != nil {
		writeOperacionErr(c, &ctrl.logger, domainErrors.InputError{Reason: "no se pudo leer el archivo"})
		return
	}
	defer src.Close()

	body, err := requests.LeerArchivo(src)
	if err != nil {
		writeOperacionErr(c, &ctrl.logger, err)
		return
	}

	lat, lon, err := requests.PuntoForm(c.PostForm("lat"), c.PostForm("lon"))
	if err != nil {
		writeOperacionErr(c, &ctrl.logger, err)
		return
	}

	var exif []byte
	if raw := strings.TrimSpace(c.PostForm("exif")); raw != "" {
		exif = []byte(raw)
	}

	res, err := ctrl.uc.Subir(c.Request.Context(), dto.SubirEvidenciaDTO{
		ID:          c.PostForm("id"),
		ActividadID: c.PostForm("actividad_id"),
		Nombre:      file.Filename,
		Nota:        c.PostForm("nota"),
		SHA256:      c.PostForm("sha256"),
		Lat:         lat,
		Lon:         lon,
		Exif:        exif,
		Rol:         u.Rol,
		CapatazID:   u.CapatazID,
		UsuarioID:   u.ID,
		OrdenID:     c.PostForm("orden_id"),
		Contenido:   body,
	})
	if err != nil {
		writeOperacionErr(c, &ctrl.logger, err)
		return
	}

	code := http.StatusCreated
	if res.Idempotente {
		code = http.StatusOK
	}
	c.JSON(code, gin.H{"id": res.ID, "idempotente": res.Idempotente})
}

// File godoc
// @Summary Download an evidence file
// @Tags evidencias
// @Param id path string true "Evidence UUID"
// @Success 200
// @Router /v1/evidencias/{id}/archivo [get]
func (ctrl *evidenciaController) File(c *gin.Context) {
	body, mime, err := ctrl.uc.Abrir(c.Request.Context(), c.Param("id"))
	if err != nil {
		if errors.Is(err, domainErrors.ErrArchivoNoDisponible) {
			c.JSON(http.StatusNotFound, gin.H{"error": "archivo no disponible"})
			return
		}
		writeOperacionErr(c, &ctrl.logger, err)
		return
	}
	defer body.Close()

	c.Header("Content-Type", mime)
	c.Header("Cache-Control", "private, max-age=86400")
	c.Header("X-Content-Type-Options", "nosniff")
	c.Header("Content-Disposition", "inline")
	c.Status(http.StatusOK)
	_, _ = io.Copy(c.Writer, body)
}
