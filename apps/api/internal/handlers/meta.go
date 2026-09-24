package handlers

import (
	"os"

	"github.com/gin-gonic/gin"
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
}

// Meta es el índice JSON de /api/v1.
type Meta struct {
	OpenAPIPath string
}

func (m Meta) Index(c *gin.Context) {
	c.JSON(200, gin.H{
		"servicio": "campus-verde-api",
		"version":  "v1",
		"crs":      "EPSG:4326",
		"rutas":    RutasV1,
	})
}

func (m Meta) OpenAPI(c *gin.Context) {
	body, err := os.ReadFile(m.OpenAPIPath)
	if err != nil {
		c.JSON(404, gin.H{"error": "openapi.yaml no disponible"})
		return
	}
	c.Data(200, "application/yaml; charset=utf-8", body)
}
