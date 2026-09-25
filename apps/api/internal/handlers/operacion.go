package handlers

import (
	"errors"
	"io"
	"log"

	"campusverde/api/internal/accesos"
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
	Ejecutor          string  `json:"ejecutor"`
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
	Motivo   string `json:"motivo"`
}

// RegistrarOperacion publica labores. registrar(r, deps) del frente operación.
// 1B puede tocar el middleware de server.go; 2A no reescribe estas rutas.
func RegistrarOperacion(r *gin.Engine, deps Deps) {
	var op Operacion
	if deps.DB != nil {
		op.Store = operacion.NewStore(deps.DB)
	}
	lab := r.Group("/api/v1/operacion")
	lab.GET("/capataces", op.Capataces)
	lab.GET("/actividades", op.List)
	lab.POST("/actividades", op.Create)
	lab.PATCH("/actividades/:id/asignacion", op.Assign)
	lab.PATCH("/actividades/:id/estado", op.Estado)
	lab.POST("/actividades/:id/archivar", op.Archive)
	lab.GET("/actividades/:id/timeline", op.Timeline)
	lab.PATCH("/actividades/:id/ficha", op.Ficha)
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
		Estado:            c.Query("estado"),
		Tipo:              c.Query("tipo"),
		ZonaSupervisionID: c.Query("zona_supervision_id"),
		CuadrillaID:       c.Query("cuadrilla_id"),
		Origen:            c.Query("origen"),
		SoloAbiertas:      c.DefaultQuery("abiertas", "1") != "0",
	}
	if u, ok := usuarioEn(c); ok {
		if u.Rol == "capataz" {
			q.Rol = operacion.RolCapataz
			q.CapatazID = u.CapatazID
		} else {
			q.Rol = c.Query("rol")
			q.CapatazID = c.Query("capataz_id")
		}
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
	if _, ok := exige(c, "validar"); !ok {
		return
	}
	var body createBody
	if err := c.ShouldBindJSON(&body); err != nil && !errors.Is(err, io.EOF) {
		c.JSON(400, gin.H{"error": "JSON inválido"})
		return
	}
	u, capataz, ok := actorDeSesion(c, body.AssignedCapatazID)
	if !ok {
		c.JSON(401, gin.H{"error": "inicie sesión"})
		return
	}
	in := operacion.CreateInput{
		ID: body.ID, Tipo: body.Tipo, Titulo: body.Titulo, Detalle: body.Detalle,
		Lon: body.Lon, Lat: body.Lat, AreaFeatureID: body.AreaFeatureID,
		ZonaFeatureID: body.ZonaFeatureID, AssignedCapatazID: capataz,
		ActorRol: u.Rol, Ejecutor: body.Ejecutor, UsuarioID: u.ID,
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
	if _, ok := exige(c, "validar"); !ok {
		return
	}
	var body assignBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(400, gin.H{"error": "JSON inválido"})
		return
	}
	u, capataz, ok := actorDeSesion(c, body.CapatazID)
	if !ok {
		c.JSON(401, gin.H{"error": "inicie sesión"})
		return
	}
	rol := u.Rol
	if err := operacion.ValidateAsignacion(rol, capataz); err != nil {
		writeOperacionErr(c, err)
		return
	}
	if h.Store == nil {
		c.JSON(503, gin.H{"error": "base de datos no disponible"})
		return
	}
	feature, err := h.Store.Assign(c.Request.Context(), c.Param("id"), capataz, rol, u.ID)
	if err != nil {
		writeOperacionErr(c, err)
		return
	}
	c.JSON(200, gin.H{"feature": feature})
}

func (h Operacion) Estado(c *gin.Context) {
	u, ok := usuarioEn(c)
	if !ok {
		c.JSON(401, gin.H{"error": "inicie sesión"})
		return
	}
	if !accesos.PermiteAlguno(u.Rol, "registrar", "validar") {
		c.JSON(403, gin.H{"error": "su rol no tiene ese permiso"})
		return
	}
	var body estadoBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(400, gin.H{"error": "JSON inválido"})
		return
	}
	_, capataz, ok := actorDeSesion(c, body.CapatazID)
	if !ok {
		c.JSON(401, gin.H{"error": "inicie sesión"})
		return
	}
	rol := u.Rol
	if err := operacion.ValidateEstado(body.Estado, rol, capataz); err != nil {
		writeOperacionErr(c, err)
		return
	}
	if h.Store == nil {
		c.JSON(503, gin.H{"error": "base de datos no disponible"})
		return
	}
	feature, err := h.Store.SetEstado(c.Request.Context(), c.Param("id"), body.Estado, rol, capataz, u.ID)
	if err != nil {
		writeOperacionErr(c, err)
		return
	}
	c.JSON(200, gin.H{"feature": feature})
}

func (h Operacion) Archive(c *gin.Context) {
	if _, ok := exige(c, "validar"); !ok {
		return
	}
	var body archiveBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(400, gin.H{"error": "JSON inválido"})
		return
	}
	u, _, ok := actorDeSesion(c, "")
	if !ok {
		c.JSON(401, gin.H{"error": "inicie sesión"})
		return
	}
	rol := u.Rol
	if err := operacion.ValidateArchivo(rol); err != nil {
		writeOperacionErr(c, err)
		return
	}
	if h.Store == nil {
		c.JSON(503, gin.H{"error": "base de datos no disponible"})
		return
	}
	if err := h.Store.Archive(c.Request.Context(), c.Param("id"), rol, body.Motivo, u.ID); err != nil {
		writeOperacionErr(c, err)
		return
	}
	c.JSON(200, gin.H{"archivada": true, "id": c.Param("id")})
}

func (h Operacion) Ficha(c *gin.Context) {
	if _, ok := exige(c, "registrar"); !ok {
		return
	}
	if h.Store == nil {
		c.JSON(503, gin.H{"error": "base de datos no disponible"})
		return
	}
	var body operacion.FichaInput
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(400, gin.H{"error": "JSON inválido"})
		return
	}
	body.ID = c.Param("id")
	if err := h.Store.GuardarFicha(c.Request.Context(), body); err != nil {
		writeOperacionErr(c, err)
		return
	}
	c.JSON(200, gin.H{"id": body.ID})
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
