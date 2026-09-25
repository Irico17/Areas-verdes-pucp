package handlers

import (
	"database/sql"
	"errors"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"

	"campusverde/api/internal/etl"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// Importaciones es el importador único CSV/XLSX/GeoJSON.
type Importaciones struct {
	DB *gorm.DB
}

// RegistrarImportaciones publica la vista previa y la confirmación.
func RegistrarImportaciones(r *gin.Engine, deps Deps) {
	h := Importaciones{}
	if deps.DB != nil {
		h.DB = deps.DB
	}
	r.GET("/api/v1/importaciones/entidades", h.Entidades)
	r.POST("/api/v1/importaciones", h.Previsualizar)
	r.POST("/api/v1/importaciones/:id/confirmar", h.Confirmar)
}

func (h Importaciones) Entidades(c *gin.Context) {
	if _, ok := exige(c, "consultar"); !ok {
		return
	}
	c.JSON(http.StatusOK, gin.H{"entidades": etl.EntidadesImportables})
}

func (h Importaciones) Previsualizar(c *gin.Context) {
	u, ok := exige(c, "validar")
	if !ok {
		return
	}
	if h.DB == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "base de datos no disponible"})
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
	vista, err := etl.Previsualizar(entidad, archivo.Filename, body)
	if err != nil {
		writeImportErr(c, err)
		return
	}
	var id int64
	err = h.DB.Raw(`
		INSERT INTO lotes_importacion (entidad, estado, usuario_id, filas, contenido, nombre_archivo)
		VALUES ($1, 'vista_previa', $2, $3, $4, $5)
		RETURNING id`, entidad, u.ID, vista.Validas, body, archivo.Filename).Row().Scan(&id)
	if err != nil {
		log.Printf("importaciones: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "no se pudo guardar la vista previa"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"id":                id,
		"lote_id":           id,
		"entidad":           vista.Entidad,
		"formato":           vista.Formato,
		"validas":           vista.Validas,
		"errores":           vista.Errores,
		"filas":             vista.Filas,
		"columnas_omitidas": vista.ColumnasOmitidas,
		"aviso_omitidas":    vista.AvisoOmitidas,
		"avisos":            vista.Avisos,
		"escrito":           false,
	})
}

func (h Importaciones) Confirmar(c *gin.Context) {
	u, ok := exige(c, "validar")
	if !ok {
		return
	}
	if h.DB == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "base de datos no disponible"})
		return
	}
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id < 1 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "lote inválido"})
		return
	}
	loteID, n, err := etl.Confirmar(h.DB, id, u.ID)
	if err != nil {
		writeImportErr(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{
		"lote_id": loteID,
		"validas": n,
		"escrito": true,
	})
}

func writeImportErr(c *gin.Context, err error) {
	switch {
	case errors.Is(err, etl.ErrEntidad):
		c.JSON(http.StatusBadRequest, gin.H{"error": "entidad no importable"})
	case errors.Is(err, etl.ErrSinValidas):
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "ninguna fila válida: no se escribió", "escrito": false})
	case errors.Is(err, etl.ErrLote):
		c.JSON(http.StatusConflict, gin.H{"error": "el lote no está en vista previa"})
	case errors.Is(err, sql.ErrNoRows):
		c.JSON(http.StatusNotFound, gin.H{"error": "lote no encontrado"})
	default:
		msg := err.Error()
		if strings.Contains(msg, "SQLSTATE") || strings.Contains(msg, "pq:") {
			log.Printf("importaciones: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "no se pudo importar"})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": msg})
	}
}
