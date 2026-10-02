package controller

import (
	"errors"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"

	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/contracts"
)

// Route documenta una ruta pública.
type Route struct {
	Metodo      string `json:"metodo"`
	Ruta        string `json:"ruta"`
	Descripcion string `json:"descripcion"`
}

// RutasV1 es el índice que también resume el README y el stub OpenAPI.
var RutasV1 = []Route{
	{Metodo: "GET", Ruta: "/health", Descripcion: "Estado del proceso, Postgres y PostGIS"},
	{Metodo: "GET", Ruta: "/api/v1", Descripcion: "Índice de la API"},
	{Metodo: "GET", Ruta: "/api/v1/openapi.yaml", Descripcion: "Contrato OpenAPI"},
	{Metodo: "GET", Ruta: "/api/v1/geo/resumen", Descripcion: "Conteos de catastro"},
	{Metodo: "GET", Ruta: "/api/v1/geo/areas", Descripcion: "Áreas verdes como FeatureCollection EPSG:4326"},
	{Metodo: "GET", Ruta: "/api/v1/geo/zonas", Descripcion: "Zonas (sin PII) como FeatureCollection EPSG:4326"},
	{Metodo: "GET", Ruta: "/api/v1/geo/capas", Descripcion: "Catálogo de capas auxiliares"},
	{Metodo: "GET", Ruta: "/api/v1/geo/capas/:capa", Descripcion: "jardines_reserva o xerofitica"},
	{Metodo: "GET", Ruta: "/api/v1/geo/edificios", Descripcion: "Huellas OSM del campus para extrusión ligera"},
	{Metodo: "GET", Ruta: "/api/v1/geo/inventario", Descripcion: "Catálogo de overlays de inventario"},
	{Metodo: "GET", Ruta: "/api/v1/geo/inventario/:capa", Descripcion: "Capa opcional: bebederos, fauna, playas, puertas, vereda, flora, cafetos, tachos"},
	{Metodo: "GET", Ruta: "/api/v1/geo/reservas-mock", Descripcion: "Agenda ficticia, sin hojas de cálculo"},
	{Metodo: "GET", Ruta: "/api/v1/operacion/capataces", Descripcion: "Equipos de campo ficticios"},
	{Metodo: "GET", Ruta: "/api/v1/operacion/actividades", Descripcion: "Labores abiertas como FeatureCollection"},
	{Metodo: "POST", Ruta: "/api/v1/operacion/actividades", Descripcion: "Alta idempotente desde un pin"},
	{Metodo: "PATCH", Ruta: "/api/v1/operacion/actividades/:id/asignacion", Descripcion: "Asignar o reasignar equipo"},
	{Metodo: "PATCH", Ruta: "/api/v1/operacion/actividades/:id/estado", Descripcion: "Cambiar estado"},
	{Metodo: "POST", Ruta: "/api/v1/operacion/actividades/:id/archivar", Descripcion: "Baja lógica"},
	{Metodo: "GET", Ruta: "/api/v1/operacion/actividades/:id/timeline", Descripcion: "Bitácora de la labor"},
	{Metodo: "POST", Ruta: "/api/v1/sesion", Descripcion: "Ingreso con cuenta local. Deja una cookie HttpOnly"},
	{Metodo: "GET", Ruta: "/api/v1/sesion", Descripcion: "Usuario de la sesión"},
	{Metodo: "DELETE", Ruta: "/api/v1/sesion", Descripcion: "Cerrar sesión"},
	{Metodo: "GET", Ruta: "/api/v1/accesos/usuarios", Descripcion: "Cuentas y permisos semilla (solo admin)"},
	{Metodo: "GET", Ruta: "/api/v1/catalogos", Descripcion: "Catálogos configurables"},
	{Metodo: "POST", Ruta: "/api/v1/catalogos", Descripcion: "Alta o reactivación de un ítem"},
	{Metodo: "PATCH", Ruta: "/api/v1/catalogos/:id", Descripcion: "Corregir el nombre de un ítem"},
	{Metodo: "POST", Ruta: "/api/v1/catalogos/:id/desactivar", Descripcion: "Baja lógica de un ítem"},
	{Metodo: "GET", Ruta: "/api/v1/catastro/areas", Descripcion: "Fichas de áreas para edición de metadatos"},
	{Metodo: "POST", Ruta: "/api/v1/catastro/areas", Descripcion: "Área sin geometría"},
	{Metodo: "PATCH", Ruta: "/api/v1/catastro/areas/:id", Descripcion: "Editar metadatos de un área"},
	{Metodo: "GET", Ruta: "/api/v1/solicitudes", Descripcion: "Solicitudes e incidencias con código externo"},
	{Metodo: "POST", Ruta: "/api/v1/solicitudes", Descripcion: "Alta manual de solicitud"},
	{Metodo: "GET", Ruta: "/api/v1/ordenes", Descripcion: "Órdenes de servicio tercerizado"},
	{Metodo: "POST", Ruta: "/api/v1/ordenes", Descripcion: "Orden ligada a una labor tercerizada"},
	{Metodo: "GET", Ruta: "/api/v1/riego", Descripcion: "Riego por sector y turno"},
	{Metodo: "POST", Ruta: "/api/v1/riego", Descripcion: "Registrar un turno de riego"},
	{Metodo: "GET", Ruta: "/api/v1/evidencias", Descripcion: "Metadatos de evidencias"},
	{Metodo: "POST", Ruta: "/api/v1/evidencias", Descripcion: "Adjuntar archivo a una labor"},
	{Metodo: "GET", Ruta: "/api/v1/reportes/labores", Descripcion: "Reporte básico. Filtros zona, cuadrilla, origen, desde y hasta. formato=csv o formato=xls"},
	{Metodo: "POST", Ruta: "/api/v1/ia/sugerir-tipo", Descripcion: "Sugerencia local de tipo a partir del título"},
}

// RutasAreasVerdesV1 es el índice para el prefijo canónico /areas-verdes/v1.
var RutasAreasVerdesV1 = func() []Route {
	rutas := make([]Route, len(RutasV1))
	for i, r := range RutasV1 {
		nuevaRuta := r.Ruta
		if len(nuevaRuta) >= 7 && nuevaRuta[:7] == "/api/v1" {
			nuevaRuta = "/areas-verdes/v1" + nuevaRuta[7:]
		}
		rutas[i] = Route{
			Metodo:      r.Metodo,
			Ruta:        nuevaRuta,
			Descripcion: r.Descripcion,
		}
	}
	return rutas
}()

// IMetaController defines operations for API metadata and contract.
type IMetaController interface {
	Index(*gin.Context)
	OpenAPI(*gin.Context)
}

type metaController struct {
	contrato contracts.IContratoOpenAPI
	logger   zerolog.Logger
}

// NewMetaController creates a meta controller.
func NewMetaController(contrato contracts.IContratoOpenAPI, logger zerolog.Logger) IMetaController {
	return &metaController{contrato: contrato, logger: logger}
}

// IndexResponse represents the API index response.
type IndexResponse struct {
	Servicio string  `json:"servicio"`
	Version  string  `json:"version"`
	CRS      string  `json:"crs"`
	Rutas    []Route `json:"rutas"`
}

// Index serves the API index with routes and service details.
// @Summary API index
// @Description Returns the service information and public routes index
// @Tags meta
// @Produce json
// @Success 200 {object} IndexResponse
// @Router /v1 [get]
func (c *metaController) Index(ctx *gin.Context) {
	rutas := RutasV1
	if len(ctx.Request.URL.Path) >= 13 && ctx.Request.URL.Path[:13] == "/areas-verdes" {
		rutas = RutasAreasVerdesV1
	}
	ctx.JSON(http.StatusOK, IndexResponse{
		Servicio: "campus-verde-api",
		Version:  "v1",
		CRS:      "EPSG:4326",
		Rutas:    rutas,
	})
}

// OpenAPI serves the merged openapi.yaml contract.
// @Summary Legacy OpenAPI YAML contract
// @Description Returns the merged OpenAPI YAML specification
// @Tags meta
// @Produce application/yaml
// @Success 200 {string} string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /v1/openapi.yaml [get]
func (c *metaController) OpenAPI(ctx *gin.Context) {
	body, err := c.contrato.ObtenerContrato()
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "openapi.yaml no disponible"})
			return
		}
		c.logger.Error().Err(err).Msg("meta: no se pudo armar el contrato openapi")
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "no se pudo armar el contrato"})
		return
	}
	ctx.Data(http.StatusOK, "application/yaml; charset=utf-8", body)
}
