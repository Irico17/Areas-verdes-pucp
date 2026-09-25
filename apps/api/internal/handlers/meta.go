package handlers

import (
	"os"

	"campusverde/api/internal/catastro"

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
	{Metodo: "POST", Ruta: "/api/v1/sesion", Descripcion: "Ingreso con cuenta local. Deja una cookie HttpOnly"},
	{Metodo: "GET", Ruta: "/api/v1/sesion", Descripcion: "Usuario de la sesión"},
	{Metodo: "DELETE", Ruta: "/api/v1/sesion", Descripcion: "Cerrar sesión"},
	{Metodo: "GET", Ruta: "/api/v1/accesos/usuarios", Descripcion: "Cuentas y permisos semilla (solo admin)"},
	{Metodo: "GET", Ruta: "/api/v1/catalogos", Descripcion: "Catálogos configurables"},
	{Metodo: "POST", Ruta: "/api/v1/catalogos", Descripcion: "Alta o reactivación de un ítem"},
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
	if _, err := os.Stat(m.OpenAPIPath); err != nil {
		c.JSON(404, gin.H{"error": "openapi.yaml no disponible"})
		return
	}
	body, err := UnirContrato(m.OpenAPIPath)
	if err != nil {
		c.JSON(500, gin.H{"error": "no se pudo armar el contrato"})
		return
	}
	c.Data(200, "application/yaml; charset=utf-8", body)
}

// RegistrarSistema publica salud, índice y el contrato. registrar(r, deps) del frente sistema.
func RegistrarSistema(r *gin.Engine, deps Deps) {
	health := Health{DB: deps.DB}
	if deps.DB != nil {
		health.Store = catastro.NewStore(deps.DB)
	}
	meta := Meta{OpenAPIPath: deps.OpenAPIPath}
	r.GET("/health", health.Get)
	r.GET("/api/v1", meta.Index)
	r.GET("/api/v1/openapi.yaml", meta.OpenAPI)
}
