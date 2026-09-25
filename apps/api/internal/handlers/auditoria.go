package handlers

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"strconv"

	"campusverde/api/internal/auditoria"

	"github.com/gin-gonic/gin"
)

// Auditoria publica lotes reversibles y el historial con usuario de sesión.
type Auditoria struct {
	Store *auditoria.Store
}

type loteBody struct {
	Entidad string           `json:"entidad"`
	Filas   []auditoria.Fila `json:"filas"`
}

type revertirBody struct {
	Confirmar bool `json:"confirmar"`
}

// RegistrarAuditoria publica el historial y la reversión de lotes.
func RegistrarAuditoria(r *gin.Engine, deps Deps) {
	var h Auditoria
	if deps.DB != nil {
		h.Store = auditoria.NewStore(deps.DB)
	}
	r.POST("/api/v1/lotes", h.Importar)
	r.POST("/api/v1/lotes/:id/revertir", h.Revertir)
	r.POST("/api/v1/auditoria/ediciones", h.Editar)
	r.GET("/api/v1/auditoria/cambios", h.Historial)
	r.GET("/api/v1/auditoria/timeline", h.Timeline)
}

func (h Auditoria) Importar(c *gin.Context) {
	u, ok := exige(c, "validar")
	if !ok {
		return
	}
	if h.Store == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "base de datos no disponible"})
		return
	}
	var body loteBody
	if err := c.ShouldBindJSON(&body); err != nil && !errors.Is(err, io.EOF) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "JSON inválido"})
		return
	}
	id, err := h.Store.Importar(c.Request.Context(), u.ID, body.Entidad, body.Filas)
	if err != nil {
		writeAuditoriaErr(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"lote_id": id, "filas": len(body.Filas)})
}

func (h Auditoria) Revertir(c *gin.Context) {
	u, ok := exige(c, "validar")
	if !ok {
		return
	}
	if h.Store == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "base de datos no disponible"})
		return
	}
	loteID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || loteID < 1 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "lote inválido"})
		return
	}
	var body revertirBody
	if err := c.ShouldBindJSON(&body); err != nil && !errors.Is(err, io.EOF) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "JSON inválido"})
		return
	}
	rep, err := h.Store.Revertir(c.Request.Context(), loteID, u.ID, body.Confirmar)
	if err != nil {
		if errors.Is(err, auditoria.ErrConfirmacion) {
			c.JSON(http.StatusConflict, gin.H{
				"error":     "hay filas editadas después del lote",
				"excluidas": rep.Excluidas,
			})
			return
		}
		writeAuditoriaErr(c, err)
		return
	}
	c.JSON(http.StatusOK, rep)
}

func (h Auditoria) Editar(c *gin.Context) {
	u, ok := exige(c, "validar")
	if !ok {
		return
	}
	if h.Store == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "base de datos no disponible"})
		return
	}
	var body struct {
		Entidad   string         `json:"entidad"`
		EntidadID string         `json:"entidad_id"`
		Despues   map[string]any `json:"despues"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "JSON inválido"})
		return
	}
	raw, err := json.Marshal(body.Despues)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "JSON inválido"})
		return
	}
	if err := h.Store.Editar(c.Request.Context(), u.ID, body.Entidad, body.EntidadID, raw); err != nil {
		writeAuditoriaErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"editada": true, "entidad_id": body.EntidadID})
}

func (h Auditoria) Historial(c *gin.Context) {
	if _, ok := exige(c, "consultar"); !ok {
		return
	}
	h.listar(c, false)
}

func (h Auditoria) Timeline(c *gin.Context) {
	if _, ok := exige(c, "consultar"); !ok {
		return
	}
	h.listar(c, true)
}

func (h Auditoria) listar(c *gin.Context, timeline bool) {
	if h.Store == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "base de datos no disponible"})
		return
	}
	f := auditoria.Filtro{
		Entidad:   c.Query("entidad"),
		EntidadID: c.Query("entidad_id"),
		Zona:      c.Query("zona"),
		Origen:    c.Query("origen"),
		Desde:     c.Query("desde"),
		Hasta:     c.Query("hasta"),
	}
	var (
		ev  []auditoria.Evento
		err error
	)
	if timeline {
		ev, err = h.Store.Timeline(c.Request.Context(), f)
	} else {
		ev, err = h.Store.Historial(c.Request.Context(), f)
	}
	if err != nil {
		writeAuditoriaErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"eventos": ev})
}

func writeAuditoriaErr(c *gin.Context, err error) {
	var input auditoria.InputError
	switch {
	case errors.Is(err, auditoria.ErrNoEncontrado):
		c.JSON(http.StatusNotFound, gin.H{"error": "lote no encontrado"})
	case errors.As(err, &input):
		c.JSON(http.StatusBadRequest, gin.H{"error": input.Reason})
	default:
		log.Printf("auditoria: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "no se pudo completar la auditoría"})
	}
}
