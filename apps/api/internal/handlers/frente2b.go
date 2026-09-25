package handlers

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"strings"

	"campusverde/api/internal/capas"
	"campusverde/api/internal/etl"

	"github.com/gin-gonic/gin"
)

// Frente2B publica tachos, bebederos, puntos, reservas ficticias y capas editables.
type Frente2B struct {
	Store *capas.Store
}

// RegistrarFrente2B es la única línea de registro del frente 2B.
func RegistrarFrente2B(r *gin.Engine, deps Deps) {
	h := Frente2B{}
	if deps.DB != nil {
		h.Store = capas.NewStore(deps.DB)
	}
	g := r.Group("/api/v1/inventario")
	g.GET("/tachos", h.listarTachos)
	g.POST("/tachos", h.guardarTacho)
	g.PATCH("/tachos/:id", h.patchTacho)
	g.GET("/tachos.csv", h.csvTachos)
	g.DELETE("/tachos/:id", h.baja("tachos"))
	g.GET("/bebederos", h.listarBebederos)
	g.POST("/bebederos", h.guardarBebedero)
	g.PATCH("/bebederos/:id", h.patchBebedero)
	g.DELETE("/bebederos/:id", h.baja("bebederos"))
	g.GET("/puntos", h.listarPuntos)
	g.POST("/puntos", h.guardarPunto)
	g.PATCH("/puntos/:id", h.patchPunto)
	g.DELETE("/puntos/:id", h.baja("puntos_pucp"))
	g.POST("/formato/puntos", h.formatoPuntos)
	g.GET("/reservas", h.listarReservas)
	g.POST("/reservas", h.guardarReserva)
	g.PATCH("/reservas/:id", h.patchReserva)
	g.DELETE("/reservas/:id", h.baja("reservas_jardin"))
	g.GET("/capas/:capa", h.listarCapa)
	g.POST("/capas/:capa", h.guardarCapa)
	g.PATCH("/capas/:capa/:id", h.patchCapa)
	g.GET("/export/:capa", h.csvCapa)
	g.DELETE("/capas/:capa/:id", h.bajaCapa)
}

func (h Frente2B) listo(c *gin.Context, accion string) bool {
	if _, ok := exige(c, accion); !ok {
		return false
	}
	if h.Store == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "base de datos no disponible"})
		return false
	}
	return true
}

func (h Frente2B) listarTachos(c *gin.Context) {
	if !h.listo(c, "consultar") {
		return
	}
	rows, err := h.Store.ListarTachos(c.Request.Context())
	if err != nil {
		c.JSON(500, gin.H{"error": "no se pudo leer los tachos"})
		return
	}
	c.JSON(200, gin.H{"tachos": rows})
}

func (h Frente2B) guardarTacho(c *gin.Context) {
	if !h.listo(c, "registrar") {
		return
	}
	var body capas.Tacho
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(400, gin.H{"error": "JSON inválido"})
		return
	}
	item, err := h.Store.GuardarTacho(c.Request.Context(), body)
	if err != nil {
		c.JSON(400, gin.H{"error": "no se pudo guardar el tacho"})
		return
	}
	c.JSON(201, item)
}

func (h Frente2B) patchTacho(c *gin.Context) {
	if !h.listo(c, "registrar") {
		return
	}
	id, ok := idParam(c)
	if !ok {
		return
	}
	var body capas.Tacho
	raw, okBody := cuerpoJSON(c, &body)
	if !okBody {
		return
	}
	item, err := h.Store.ActualizarTacho(c.Request.Context(), id, body, raw)
	if err != nil {
		c.JSON(400, gin.H{"error": "no se pudo actualizar el tacho"})
		return
	}
	c.JSON(200, item)
}

func (h Frente2B) csvTachos(c *gin.Context) {
	if !h.listo(c, "consultar") {
		return
	}
	rows, err := h.Store.ListarTachos(c.Request.Context())
	if err != nil {
		c.JSON(500, gin.H{"error": "no se pudo exportar"})
		return
	}
	c.Header("Content-Type", "text/csv; charset=utf-8")
	c.String(200, capas.CSVTachos(rows))
}

func (h Frente2B) listarBebederos(c *gin.Context) {
	if !h.listo(c, "consultar") {
		return
	}
	rows, err := h.Store.ListarBebederos(c.Request.Context())
	if err != nil {
		c.JSON(500, gin.H{"error": "no se pudo leer los bebederos"})
		return
	}
	c.JSON(200, gin.H{"bebederos": rows})
}

func (h Frente2B) guardarBebedero(c *gin.Context) {
	if !h.listo(c, "registrar") {
		return
	}
	var body capas.Bebedero
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(400, gin.H{"error": "JSON inválido"})
		return
	}
	item, err := h.Store.GuardarBebedero(c.Request.Context(), body)
	if err != nil {
		c.JSON(400, gin.H{"error": "no se pudo guardar el bebedero"})
		return
	}
	c.JSON(201, item)
}

func (h Frente2B) patchBebedero(c *gin.Context) {
	if !h.listo(c, "registrar") {
		return
	}
	id, ok := idParam(c)
	if !ok {
		return
	}
	var body capas.Bebedero
	raw, okBody := cuerpoJSON(c, &body)
	if !okBody {
		return
	}
	item, err := h.Store.ActualizarBebedero(c.Request.Context(), id, body, raw)
	if err != nil {
		c.JSON(400, gin.H{"error": "no se pudo actualizar el bebedero"})
		return
	}
	c.JSON(200, item)
}

func (h Frente2B) listarPuntos(c *gin.Context) {
	if !h.listo(c, "consultar") {
		return
	}
	rows, err := h.Store.ListarPuntos(c.Request.Context(), c.Query("q"))
	if err != nil {
		c.JSON(500, gin.H{"error": "no se pudo leer los puntos"})
		return
	}
	c.JSON(200, gin.H{"puntos": rows})
}

func (h Frente2B) guardarPunto(c *gin.Context) {
	if !h.listo(c, "registrar") {
		return
	}
	raw, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(400, gin.H{"error": "JSON inválido"})
		return
	}
	if capas.CuerpoConContacto(string(raw)) {
		c.JSON(400, gin.H{"error": "no se guardan teléfono, placeId ni website", "columnas_omitidas": []string{"phone", "placeId", "website"}})
		return
	}
	c.Request.Body = io.NopCloser(strings.NewReader(string(raw)))
	var body capas.Punto
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(400, gin.H{"error": "JSON inválido"})
		return
	}
	item, err := h.Store.GuardarPunto(c.Request.Context(), body)
	if err != nil {
		c.JSON(400, gin.H{"error": "no se pudo guardar el punto"})
		return
	}
	c.JSON(201, item)
}

func (h Frente2B) patchPunto(c *gin.Context) {
	if !h.listo(c, "registrar") {
		return
	}
	id, ok := idParam(c)
	if !ok {
		return
	}
	raw, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(400, gin.H{"error": "JSON inválido"})
		return
	}
	if capas.CuerpoConContacto(string(raw)) {
		c.JSON(400, gin.H{"error": "no se guardan teléfono, placeId ni website", "columnas_omitidas": []string{"phone", "placeId", "website"}})
		return
	}
	var body capas.Punto
	if err := json.Unmarshal(raw, &body); err != nil {
		c.JSON(400, gin.H{"error": "JSON inválido"})
		return
	}
	item, err := h.Store.ActualizarPunto(c.Request.Context(), id, body, raw)
	if err != nil {
		c.JSON(400, gin.H{"error": "no se pudo actualizar el punto"})
		return
	}
	c.JSON(200, item)
}

func (h Frente2B) formatoPuntos(c *gin.Context) {
	if !h.listo(c, "consultar") {
		return
	}
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(400, gin.H{"error": "CSV inválido"})
		return
	}
	rows, rech, omitidas, err := etl.LeerPuntosPUCP(body)
	if err != nil {
		c.JSON(400, gin.H{"error": "CSV inválido"})
		return
	}
	c.JSON(200, gin.H{
		"filas":             len(rows),
		"rechazados":        rech,
		"columnas_omitidas": omitidas,
		"nota":              "columnas omitidas por datos personales",
	})
}

func (h Frente2B) listarReservas(c *gin.Context) {
	if !h.listo(c, "consultar") {
		return
	}
	rows, err := h.Store.ListarReservas(c.Request.Context(), c.Query("desde"), c.Query("hasta"))
	if err != nil {
		c.JSON(500, gin.H{"error": "no se pudo leer las reservas"})
		return
	}
	c.JSON(200, gin.H{"origen": "ficticio", "aviso": "Agenda ficticia. La hoja de reservas responde 401 y no se abre.", "reservas": rows})
}

func (h Frente2B) guardarReserva(c *gin.Context) {
	if !h.listo(c, "registrar") {
		return
	}
	var body capas.Reserva
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(400, gin.H{"error": "JSON inválido"})
		return
	}
	if body.Origen != "" && body.Origen != "ficticio" {
		c.JSON(400, gin.H{"error": "mientras la hoja responda 401 solo se acepta origen ficticio"})
		return
	}
	item, err := h.Store.GuardarReserva(c.Request.Context(), body)
	if err != nil {
		c.JSON(400, gin.H{"error": "no se pudo guardar la reserva"})
		return
	}
	c.JSON(201, item)
}

func (h Frente2B) patchReserva(c *gin.Context) {
	if !h.listo(c, "registrar") {
		return
	}
	id, ok := idParam(c)
	if !ok {
		return
	}
	var body capas.Reserva
	raw, okBody := cuerpoJSON(c, &body)
	if !okBody {
		return
	}
	if body.Origen != "" && body.Origen != "ficticio" {
		c.JSON(400, gin.H{"error": "mientras la hoja responda 401 solo se acepta origen ficticio"})
		return
	}
	item, err := h.Store.ActualizarReserva(c.Request.Context(), id, body, raw)
	if err != nil {
		c.JSON(400, gin.H{"error": "no se pudo actualizar la reserva"})
		return
	}
	c.JSON(200, item)
}

func (h Frente2B) listarCapa(c *gin.Context) {
	if !h.listo(c, "consultar") {
		return
	}
	rows, err := h.Store.ListarFichas(c.Request.Context(), c.Param("capa"))
	if err != nil {
		c.JSON(404, gin.H{"error": "capa desconocida"})
		return
	}
	c.JSON(200, gin.H{"filas": rows})
}

func (h Frente2B) guardarCapa(c *gin.Context) {
	if !h.listo(c, "registrar") {
		return
	}
	var body capas.Ficha
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(400, gin.H{"error": "JSON inválido"})
		return
	}
	item, err := h.Store.GuardarFicha(c.Request.Context(), c.Param("capa"), body)
	if err != nil {
		c.JSON(400, gin.H{"error": "no se pudo guardar la ficha"})
		return
	}
	c.JSON(201, item)
}

func (h Frente2B) patchCapa(c *gin.Context) {
	if !h.listo(c, "registrar") {
		return
	}
	id, ok := idParam(c)
	if !ok {
		return
	}
	var body capas.Ficha
	raw, okBody := cuerpoJSON(c, &body)
	if !okBody {
		return
	}
	item, err := h.Store.ActualizarFicha(c.Request.Context(), c.Param("capa"), id, body, raw)
	if err != nil {
		c.JSON(400, gin.H{"error": "no se pudo actualizar la ficha"})
		return
	}
	c.JSON(200, item)
}

func (h Frente2B) csvCapa(c *gin.Context) {
	if !h.listo(c, "consultar") {
		return
	}
	rows, err := h.Store.ListarFichas(c.Request.Context(), c.Param("capa"))
	if err != nil {
		c.JSON(404, gin.H{"error": "capa desconocida"})
		return
	}
	var b strings.Builder
	b.WriteString("feature_id,nombre,codigo,nota,clase,riego,pertenecen,uso\n")
	for _, f := range rows {
		b.WriteString(strings.Join([]string{f.FeatureID, f.Nombre, f.Codigo, f.Nota, f.Clase, f.Riego, f.Pertenecen, f.Uso}, ",") + "\n")
	}
	c.Header("Content-Type", "text/csv; charset=utf-8")
	c.String(200, b.String())
}

func cuerpoJSON(c *gin.Context, dest any) (json.RawMessage, bool) {
	raw, err := io.ReadAll(c.Request.Body)
	if err != nil || json.Unmarshal(raw, dest) != nil {
		c.JSON(400, gin.H{"error": "JSON inválido"})
		return nil, false
	}
	return raw, true
}

func idParam(c *gin.Context) (int64, bool) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id < 1 {
		c.JSON(400, gin.H{"error": "id inválido"})
		return 0, false
	}
	return id, true
}

func (h Frente2B) baja(tabla string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !h.listo(c, "registrar") {
			return
		}
		id, err := strconv.ParseInt(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(400, gin.H{"error": "id inválido"})
			return
		}
		if err := h.Store.Baja(c.Request.Context(), tabla, id); err != nil {
			c.JSON(404, gin.H{"error": "no se encontró el registro"})
			return
		}
		c.JSON(200, gin.H{"activo": false})
	}
}

func (h Frente2B) bajaCapa(c *gin.Context) {
	if !h.listo(c, "registrar") {
		return
	}
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(400, gin.H{"error": "id inválido"})
		return
	}
	if err := h.Store.Baja(c.Request.Context(), c.Param("capa"), id); err != nil {
		c.JSON(404, gin.H{"error": "no se encontró el registro"})
		return
	}
	c.JSON(200, gin.H{"activo": false})
}
