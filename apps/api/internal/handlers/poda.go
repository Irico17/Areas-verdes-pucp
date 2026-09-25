package handlers

import (
	"net/http"

	"campusverde/api/internal/atencion"
	"campusverde/api/internal/etl"
	"campusverde/api/internal/operacion"

	"github.com/gin-gonic/gin"
)

// RegistrarPoda publica poda, vivero y la edición de solicitudes del frente 2A.
func RegistrarPoda(r *gin.Engine, deps Deps) {
	var store *atencion.Store
	if deps.DB != nil {
		store = atencion.NewStore(deps.DB)
	}
	h := PodaVivero{Store: store}
	r.GET("/api/v1/podas", h.ListarPodas)
	r.POST("/api/v1/podas", h.CrearPoda)
	r.PATCH("/api/v1/podas/:id", h.EditarPoda)
	r.POST("/api/v1/podas/:id/archivar", h.ArchivarPoda)
	r.GET("/api/v1/vivero", h.ListarVivero)
	r.POST("/api/v1/vivero", h.CrearVivero)
	r.PATCH("/api/v1/vivero/:id", h.EditarVivero)
	r.POST("/api/v1/vivero/:id/archivar", h.ArchivarVivero)
	r.PATCH("/api/v1/solicitudes/:id", h.EditarSolicitud)
	r.PATCH("/api/v1/ordenes/:id", h.EditarOrden)
	r.POST("/api/v1/operacion/actividades/:id/avances", h.CrearAvance)
}

// PodaVivero atiende los formularios de escritorio de poda y vivero.
type PodaVivero struct {
	Store *atencion.Store
}

func (h PodaVivero) ready(c *gin.Context) bool {
	if h.Store == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "base de datos no disponible"})
		return false
	}
	return true
}

func (h PodaVivero) ListarPodas(c *gin.Context) {
	if _, ok := exige(c, "consultar"); !ok {
		return
	}
	if !h.ready(c) {
		return
	}
	rows, err := h.Store.ListarPodas(c.Request.Context())
	if err != nil {
		c.JSON(500, gin.H{"error": "no se pudieron leer las podas"})
		return
	}
	c.JSON(200, gin.H{"podas": rows})
}

func (h PodaVivero) CrearPoda(c *gin.Context) {
	if _, ok := exige(c, "registrar"); !ok {
		return
	}
	if !h.ready(c) {
		return
	}
	var body atencion.PodaInput
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(400, gin.H{"error": "JSON inválido"})
		return
	}
	body.CodigoExterno = etl.CodigoExterno(body.CodigoExterno)
	item, err := h.Store.GuardarPoda(c.Request.Context(), body)
	if err != nil {
		writeAtencion(c, err)
		return
	}
	c.JSON(201, item)
}

func (h PodaVivero) EditarPoda(c *gin.Context) {
	if _, ok := exige(c, "registrar"); !ok {
		return
	}
	if !h.ready(c) {
		return
	}
	var body atencion.PodaInput
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(400, gin.H{"error": "JSON inválido"})
		return
	}
	body.ID = c.Param("id")
	body.CodigoExterno = etl.CodigoExterno(body.CodigoExterno)
	item, err := h.Store.GuardarPoda(c.Request.Context(), body)
	if err != nil {
		writeAtencion(c, err)
		return
	}
	c.JSON(200, item)
}

func (h PodaVivero) ArchivarPoda(c *gin.Context) {
	if _, ok := exige(c, "validar"); !ok {
		return
	}
	if !h.ready(c) {
		return
	}
	if err := h.Store.ArchivarPoda(c.Request.Context(), c.Param("id")); err != nil {
		writeAtencion(c, err)
		return
	}
	c.JSON(200, gin.H{"ok": true})
}

func (h PodaVivero) ListarVivero(c *gin.Context) {
	if _, ok := exige(c, "consultar"); !ok {
		return
	}
	if !h.ready(c) {
		return
	}
	rows, err := h.Store.ListarVivero(c.Request.Context(), c.Query("mes"))
	if err != nil {
		writeAtencion(c, err)
		return
	}
	c.JSON(200, gin.H{"registros": rows})
}

func (h PodaVivero) CrearVivero(c *gin.Context) {
	if _, ok := exige(c, "registrar"); !ok {
		return
	}
	if !h.ready(c) {
		return
	}
	var body atencion.ViveroInput
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(400, gin.H{"error": "JSON inválido"})
		return
	}
	item, err := h.Store.GuardarVivero(c.Request.Context(), body)
	if err != nil {
		writeAtencion(c, err)
		return
	}
	c.JSON(201, item)
}

func (h PodaVivero) EditarVivero(c *gin.Context) {
	if _, ok := exige(c, "registrar"); !ok {
		return
	}
	if !h.ready(c) {
		return
	}
	var body atencion.ViveroInput
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(400, gin.H{"error": "JSON inválido"})
		return
	}
	body.ID = c.Param("id")
	item, err := h.Store.GuardarVivero(c.Request.Context(), body)
	if err != nil {
		writeAtencion(c, err)
		return
	}
	c.JSON(200, item)
}

func (h PodaVivero) ArchivarVivero(c *gin.Context) {
	if _, ok := exige(c, "validar"); !ok {
		return
	}
	if !h.ready(c) {
		return
	}
	if err := h.Store.ArchivarVivero(c.Request.Context(), c.Param("id")); err != nil {
		writeAtencion(c, err)
		return
	}
	c.JSON(200, gin.H{"ok": true})
}

func (h PodaVivero) EditarSolicitud(c *gin.Context) {
	if _, ok := exige(c, "solicitudes"); !ok {
		return
	}
	if !h.ready(c) {
		return
	}
	var body atencion.SolicitudInput
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(400, gin.H{"error": "JSON inválido"})
		return
	}
	body.ID = c.Param("id")
	body.CodigoExterno = etl.CodigoExterno(body.CodigoExterno)
	item, err := h.Store.EditarSolicitud(c.Request.Context(), body)
	if err != nil {
		writeAtencion(c, err)
		return
	}
	c.JSON(200, item)
}

func (h PodaVivero) EditarOrden(c *gin.Context) {
	if _, ok := exige(c, "registrar"); !ok {
		return
	}
	if !h.ready(c) {
		return
	}
	var body atencion.OrdenInput
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(400, gin.H{"error": "JSON inválido"})
		return
	}
	body.ID = c.Param("id")
	item, err := h.Store.EditarOrden(c.Request.Context(), body)
	if err != nil {
		writeAtencion(c, err)
		return
	}
	c.JSON(200, gin.H{"orden": item, "aviso": "La conformidad no cierra la solicitud."})
}

func (h PodaVivero) CrearAvance(c *gin.Context) {
	if _, ok := exige(c, "registrar"); !ok {
		return
	}
	if !h.ready(c) {
		return
	}
	var body struct {
		ID       string `json:"id"`
		Fecha    string `json:"fecha"`
		Nota     string `json:"nota"`
		Area     string `json:"area_feature_id"`
		Ejemplar string `json:"ejemplar_ref"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(400, gin.H{"error": "JSON inválido"})
		return
	}
	if err := h.Store.CrearAvance(c.Request.Context(), c.Param("id"), body.ID, body.Fecha, body.Nota, body.Area, body.Ejemplar); err != nil {
		writeAtencion(c, err)
		return
	}
	c.JSON(201, gin.H{"id": body.ID})
}

// El código externo de una solicitud editada no se inventa: lo filtra CodigoExterno.
var _ = operacion.ErrNoEncontrada
