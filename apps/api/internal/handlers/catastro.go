package handlers

import (
	"errors"
	"strconv"

	"campusverde/api/internal/catastro"

	"github.com/gin-gonic/gin"
)

// Modelo expone el catastro nuevo: zonas de supervisión, ejemplares y capas.
type Modelo struct {
	Store *catastro.Store
}

// RegistrarEjemplares publica el modelo del frente 1A.
func RegistrarEjemplares(r *gin.Engine, deps Deps) {
	var store *catastro.Store
	if deps.DB != nil {
		store = catastro.NewStore(deps.DB)
	}
	h := Modelo{Store: store}
	g := r.Group("/api/v1/catastro")
	g.GET("/zonas-supervision", h.Zonas)
	g.POST("/zonas-supervision", h.CrearZona)
	g.GET("/poligonos", h.Poligonos)
	g.GET("/cuadrillas", h.Cuadrillas)
	g.POST("/cuadrillas", h.CrearCuadrilla)
	g.GET("/lugares", h.Lugares)
	g.POST("/lugares", h.CrearLugar)
	g.GET("/especies", h.Especies)
	g.POST("/especies", h.CrearEspecie)
	g.GET("/ejemplares", h.Ejemplares)
	g.POST("/ejemplares", h.CrearEjemplar)
	g.GET("/ejemplares/:id/codigos", h.Codigos)
	g.POST("/ejemplares/:id/codigos", h.Recodificar)
	g.GET("/fauna", h.capa("fauna"))
	g.GET("/puertas", h.capa("puertas"))
	g.GET("/playas", h.capa("playas_estacionamiento"))
	g.GET("/veredas", h.capa("veredas_riesgo"))
	g.GET("/xerofiticas", h.capa("xerofiticas"))
	g.GET("/jardines-reserva", h.capa("jardines_reserva"))
}

func (h Modelo) listo(c *gin.Context, accion string) bool {
	if _, ok := exige(c, accion); !ok {
		return false
	}
	if h.Store == nil {
		c.JSON(503, gin.H{"error": "base de datos no disponible"})
		return false
	}
	return true
}

func (h Modelo) Zonas(c *gin.Context) {
	if !h.listo(c, "consultar") {
		return
	}
	rows, err := h.Store.ListarZonasSupervision(c.Request.Context())
	if err != nil {
		c.JSON(500, gin.H{"error": "no se pudo leer las zonas de supervisión"})
		return
	}
	c.JSON(200, gin.H{"zonas_supervision": rows})
}

func (h Modelo) CrearZona(c *gin.Context) {
	if !h.listo(c, "registrar") {
		return
	}
	var body struct {
		Codigo  string   `json:"codigo"`
		Nombre  string   `json:"nombre"`
		AreaM2  *float64 `json:"area_m2"`
		GeoJSON string   `json:"geojson"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(400, gin.H{"error": "JSON inválido"})
		return
	}
	item, err := h.Store.CrearZonaSupervision(c.Request.Context(), body.Codigo, body.Nombre, body.GeoJSON, body.AreaM2)
	if err != nil {
		c.JSON(400, gin.H{"error": "no se pudo crear la zona de supervisión"})
		return
	}
	c.JSON(201, item)
}

func (h Modelo) Poligonos(c *gin.Context) {
	if !h.listo(c, "consultar") {
		return
	}
	rows, err := h.Store.ListarPoligonos(c.Request.Context())
	if err != nil {
		c.JSON(500, gin.H{"error": "no se pudo leer los polígonos"})
		return
	}
	c.JSON(200, gin.H{"poligonos": rows})
}

func (h Modelo) Cuadrillas(c *gin.Context) {
	if !h.listo(c, "consultar") {
		return
	}
	rows, err := h.Store.ListarCuadrillas(c.Request.Context())
	if err != nil {
		c.JSON(500, gin.H{"error": "no se pudo leer las cuadrillas"})
		return
	}
	c.JSON(200, gin.H{"cuadrillas": rows})
}

func (h Modelo) CrearCuadrilla(c *gin.Context) {
	if !h.listo(c, "registrar") {
		return
	}
	var body struct {
		ID     string `json:"id"`
		Nombre string `json:"nombre_ficticio"`
		Turno  string `json:"turno"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(400, gin.H{"error": "JSON inválido"})
		return
	}
	item, err := h.Store.CrearCuadrilla(c.Request.Context(), body.ID, body.Nombre, body.Turno)
	if err != nil {
		c.JSON(400, gin.H{"error": "no se pudo crear la cuadrilla"})
		return
	}
	c.JSON(201, item)
}

func (h Modelo) Lugares(c *gin.Context) {
	if !h.listo(c, "consultar") {
		return
	}
	rows, err := h.Store.ListarLugares(c.Request.Context())
	if err != nil {
		c.JSON(500, gin.H{"error": "no se pudo leer los lugares"})
		return
	}
	c.JSON(200, gin.H{"lugares": rows})
}

func (h Modelo) CrearLugar(c *gin.Context) {
	if !h.listo(c, "registrar") {
		return
	}
	var body struct {
		Nombre string  `json:"nombre"`
		Lat    float64 `json:"lat"`
		Lon    float64 `json:"lon"`
		ZonaID *int64  `json:"zona_supervision_id"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(400, gin.H{"error": "JSON inválido"})
		return
	}
	item, err := h.Store.CrearLugar(c.Request.Context(), body.Nombre, body.Lat, body.Lon, body.ZonaID)
	if err != nil {
		c.JSON(400, gin.H{"error": "no se pudo crear el lugar"})
		return
	}
	c.JSON(201, item)
}

func (h Modelo) Especies(c *gin.Context) {
	if !h.listo(c, "consultar") {
		return
	}
	rows, err := h.Store.ListarEspecies(c.Request.Context())
	if err != nil {
		c.JSON(500, gin.H{"error": "no se pudo leer las especies"})
		return
	}
	c.JSON(200, gin.H{"especies": rows})
}

func (h Modelo) CrearEspecie(c *gin.Context) {
	if !h.listo(c, "registrar") {
		return
	}
	var body struct {
		Cientifico string `json:"nombre_cientifico"`
		Comun      string `json:"nombre_comun"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(400, gin.H{"error": "JSON inválido"})
		return
	}
	item, err := h.Store.CrearEspecie(c.Request.Context(), body.Cientifico, body.Comun)
	if err != nil {
		c.JSON(400, gin.H{"error": "no se pudo crear la especie"})
		return
	}
	c.JSON(201, item)
}

func (h Modelo) Ejemplares(c *gin.Context) {
	if !h.listo(c, "consultar") {
		return
	}
	rows, err := h.Store.ListarEjemplares(c.Request.Context())
	if err != nil {
		c.JSON(500, gin.H{"error": "no se pudo leer los ejemplares"})
		return
	}
	c.JSON(200, gin.H{"ejemplares": rows})
}

func (h Modelo) CrearEjemplar(c *gin.Context) {
	if !h.listo(c, "registrar") {
		return
	}
	var body catastro.Ejemplar
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(400, gin.H{"error": "JSON inválido"})
		return
	}
	item, err := h.Store.CrearEjemplar(c.Request.Context(), body)
	if err != nil {
		c.JSON(400, gin.H{"error": "no se pudo crear el ejemplar"})
		return
	}
	c.JSON(201, item)
}

func (h Modelo) Codigos(c *gin.Context) {
	if !h.listo(c, "consultar") {
		return
	}
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(400, gin.H{"error": "id inválido"})
		return
	}
	rows, err := h.Store.ListarCodigos(c.Request.Context(), id)
	if err != nil {
		c.JSON(500, gin.H{"error": "no se pudo leer el historial"})
		return
	}
	c.JSON(200, gin.H{"codigos": rows})
}

func (h Modelo) Recodificar(c *gin.Context) {
	if !h.listo(c, "registrar") {
		return
	}
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(400, gin.H{"error": "id inválido"})
		return
	}
	var body struct {
		Codigo string `json:"codigo_nuevo"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(400, gin.H{"error": "JSON inválido"})
		return
	}
	item, err := h.Store.Recodificar(c.Request.Context(), id, body.Codigo)
	if errors.Is(err, catastro.ErrNoEncontrado) {
		c.JSON(404, gin.H{"error": "ejemplar sin código anterior"})
		return
	}
	if err != nil {
		c.JSON(400, gin.H{"error": "no se pudo recodificar"})
		return
	}
	c.JSON(201, item)
}

func (h Modelo) capa(tabla string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !h.listo(c, "consultar") {
			return
		}
		rows, err := h.Store.ListarCapa(c.Request.Context(), tabla)
		if err != nil {
			c.JSON(500, gin.H{"error": "no se pudo leer la capa"})
			return
		}
		c.JSON(200, gin.H{"filas": rows})
	}
}
