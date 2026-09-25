package handlers

import (
	"encoding/json"
	"strconv"
	"strings"

	"campusverde/api/internal/accesos"
	"campusverde/api/internal/blobs"
	"campusverde/api/internal/evidencias"
	"campusverde/api/internal/operacion"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func guardarEvidenciaForm(c *gin.Context, db *gorm.DB, files blobs.Store, u accesos.Usuario) (evidencias.Resultado, error) {
	file, err := c.FormFile("archivo")
	if err != nil {
		return evidencias.Resultado{}, operacion.InputError{Reason: "falta el archivo"}
	}
	src, err := file.Open()
	if err != nil {
		return evidencias.Resultado{}, operacion.InputError{Reason: "no se pudo leer el archivo"}
	}
	defer src.Close()
	body, err := evidencias.Leer(src)
	if err != nil {
		return evidencias.Resultado{}, err
	}
	lat, lon, err := puntoForm(c.PostForm("lat"), c.PostForm("lon"))
	if err != nil {
		return evidencias.Resultado{}, err
	}
	var exif json.RawMessage
	if raw := strings.TrimSpace(c.PostForm("exif")); raw != "" {
		exif = json.RawMessage(raw)
	}
	return evidencias.Guardar(c.Request.Context(), db, files, evidencias.Entrada{
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
}

func puntoForm(latRaw, lonRaw string) (*float64, *float64, error) {
	latRaw = strings.TrimSpace(latRaw)
	lonRaw = strings.TrimSpace(lonRaw)
	if latRaw == "" && lonRaw == "" {
		return nil, nil, nil
	}
	if latRaw == "" || lonRaw == "" {
		return nil, nil, operacion.InputError{Reason: "lat y lon van juntos"}
	}
	lat, err := strconv.ParseFloat(latRaw, 64)
	if err != nil {
		return nil, nil, operacion.InputError{Reason: "lat no es un número"}
	}
	lon, err := strconv.ParseFloat(lonRaw, 64)
	if err != nil {
		return nil, nil, operacion.InputError{Reason: "lon no es un número"}
	}
	return &lat, &lon, nil
}
