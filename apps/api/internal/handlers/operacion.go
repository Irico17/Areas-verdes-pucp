package handlers

import (
	"errors"
	"io"
	"log"

	"campusverde/api/internal/operacion"

	"github.com/gin-gonic/gin"
)

// Operacion expone labores, asignación y bitácora.
type Operacion struct {
	Store *operacion.Store
}

type createBody struct {
	ID                string  `json:"id"`
	Tipo              string  `json:"tipo"`
	Titulo            string  `json:"titulo"`
	Detalle           string  `json:"detalle"`
	Lon               float64 `json:"lon"`
	Lat               float64 `json:"lat"`
	AreaFeatureID     string  `json:"area_feature_id"`
	ZonaFeatureID     string  `json:"zona_feature_id"`
	AssignedCapatazID string  `json:"assigned_capataz_id"`
	ActorRol          string  `json:"actor_rol"`
}

type assignBody struct {
	CapatazID string `json:"capataz_id"`
	ActorRol  string `json:"actor_rol"`
}

type estadoBody struct {
	Estado    string `json:"estado"`
	ActorRol  string `json:"actor_rol"`
	CapatazID string `json:"capataz_id"`
}

type archiveBody struct {
	ActorRol string `json:"actor_rol"`
}

func (h Operacion) Capataces(c *gin.Context) {
	if h.Store == nil {
		c.JSON(503, gin.H{"error": "base de datos no disponible"})
		return
	}
	rows, err := h.Store.Capataces(c.Request.Context())
	if err != nil {
		log.Printf("capataces: %v", err)
		c.JSON(500, gin.H{"error": "no se pudo leer los equipos"})
		return
	}
	c.JSON(200, gin.H{"capataces": rows})
}

func (h Operacion) List(c *gin.Context) {
	q := operacion.Query{
		Rol:          c.Query("rol"),
		CapatazID:    c.Query("capataz_id"),
		Estado:       c.Query("estado"),
		Tipo:         c.Query("tipo"),
		SoloAbiertas: c.DefaultQuery("abiertas", "1") != "0",
	}
	if err := operacion.ValidateQuery(q); err != nil {
		writeOperacionErr(c, err)
		return
	}
	if h.Store == nil {
		c.JSON(503, gin.H{"error": "base de datos no disponible"})
		return
	}
	fc, err := h.Store.List(c.Request.Context(), q)
	if err != nil {
		writeOperacionErr(c, err)
		return
	}
	c.Header("Content-Type", "application/geo+json; charset=utf-8")
	c.JSON(200, fc)
}

func (h Operacion) Create(c *gin.Context) {
	var body createBody
	if err := c.ShouldBindJSON(&body); err != nil && !errors.Is(err, io.EOF) {
		c.JSON(400, gin.H{"error": "JSON inválido"})
		return
	}
	in := operacion.CreateInput{
		ID: body.ID, Tipo: body.Tipo, Titulo: body.Titulo, Detalle: body.Detalle,
		Lon: body.Lon, Lat: body.Lat, AreaFeatureID: body.AreaFeatureID,
		ZonaFeatureID: body.ZonaFeatureID, AssignedCapatazID: body.AssignedCapatazID,
		ActorRol: body.ActorRol,
	}
	if err := operacion.ValidateCreate(in); err != nil {
		writeOperacionErr(c, err)
		return
	}
	if h.Store == nil {
		c.JSON(503, gin.H{"error": "base de datos no disponible"})
		return
	}
	feature, created, err := h.Store.Create(c.Request.Context(), in)
	if err != nil {
		writeOperacionErr(c, err)
		return
	}
	status := 200
	if created {
		status = 201
	}
	c.JSON(status, gin.H{"creada": created, "feature": feature})
}

func (h Operacion) Assign(c *gin.Context) {
	var body assignBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(400, gin.H{"error": "JSON inválido"})
		return
	}
	if err := operacion.ValidateAsignacion(body.ActorRol, body.CapatazID); err != nil {
		writeOperacionErr(c, err)
		return
	}
	if h.Store == nil {
		c.JSON(503, gin.H{"error": "base de datos no disponible"})
		return
	}
	feature, err := h.Store.Assign(c.Request.Context(), c.Param("id"), body.CapatazID, body.ActorRol)
	if err != nil {
		writeOperacionErr(c, err)
		return
	}
	c.JSON(200, gin.H{"feature": feature})
}

func (h Operacion) Estado(c *gin.Context) {
	var body estadoBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(400, gin.H{"error": "JSON inválido"})
		return
	}
	if err := operacion.ValidateEstado(body.Estado, body.ActorRol, body.CapatazID); err != nil {
		writeOperacionErr(c, err)
		return
	}
	if h.Store == nil {
		c.JSON(503, gin.H{"error": "base de datos no disponible"})
		return
	}
	feature, err := h.Store.SetEstado(c.Request.Context(), c.Param("id"), body.Estado, body.ActorRol, body.CapatazID)
	if err != nil {
		writeOperacionErr(c, err)
		return
	}
	c.JSON(200, gin.H{"feature": feature})
}

func (h Operacion) Archive(c *gin.Context) {
	var body archiveBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(400, gin.H{"error": "JSON inválido"})
		return
	}
	if err := operacion.ValidateArchivo(body.ActorRol); err != nil {
		writeOperacionErr(c, err)
		return
	}
	if h.Store == nil {
		c.JSON(503, gin.H{"error": "base de datos no disponible"})
		return
	}
	if err := h.Store.Archive(c.Request.Context(), c.Param("id"), body.ActorRol); err != nil {
		writeOperacionErr(c, err)
		return
	}
	c.JSON(200, gin.H{"archivada": true, "id": c.Param("id")})
}

func (h Operacion) Timeline(c *gin.Context) {
	if h.Store == nil {
		c.JSON(503, gin.H{"error": "base de datos no disponible"})
		return
	}
	tl, err := h.Store.Timeline(c.Request.Context(), c.Param("id"))
	if err != nil {
		writeOperacionErr(c, err)
		return
	}
	c.JSON(200, tl)
}

func writeOperacionErr(c *gin.Context, err error) {
	var input operacion.InputError
	switch {
	case errors.Is(err, operacion.ErrProhibido):
		c.JSON(403, gin.H{"error": "este rol no puede hacer esa acción"})
	case errors.Is(err, operacion.ErrNoEncontrada):
		c.JSON(404, gin.H{"error": "labor no encontrada"})
	case errors.Is(err, operacion.ErrConflicto):
		c.JSON(409, gin.H{"error": "ese id ya existe con otro contenido"})
	case errors.As(err, &input):
		c.JSON(400, gin.H{"error": input.Reason})
	default:
		log.Printf("operacion: %v", err)
		c.JSON(500, gin.H{"error": "no se pudo completar la operación"})
	}
}
