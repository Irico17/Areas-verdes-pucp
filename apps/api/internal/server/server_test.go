package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"campusverde/api/internal/handlers"
)

func TestHealthYIndiceSinDB(t *testing.T) {
	r := New(Deps{OpenAPIPath: "no-existe.yaml"})

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/health", nil))
	if w.Code != 503 {
		t.Fatalf("health %d %s", w.Code, w.Body.String())
	}

	w = httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/v1", nil))
	if w.Code != 200 {
		t.Fatalf("indice %d", w.Code)
	}
	var body struct {
		CRS   string `json:"crs"`
		Rutas []any  `json:"rutas"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body.CRS != "EPSG:4326" || len(body.Rutas) < 4 {
		t.Fatalf("indice %#v", body)
	}

	w = httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/v1/geo/areas", nil))
	if w.Code != 503 {
		t.Fatalf("areas sin db %d", w.Code)
	}
}

// rutasCongeladas es el contrato HTTP de la ola 0. Un frente nuevo suma una ruta;
// no cambia el código de las que ya estaban.
var rutasCongeladas = []struct {
	metodo string
	ruta   string
	codigo int
}{
	{http.MethodGet, "/health", 503},
	{http.MethodGet, "/api/v1", 200},
	{http.MethodGet, "/api/v1/openapi.yaml", 404},
	{http.MethodGet, "/api/v1/geo/resumen", 503},
	{http.MethodGet, "/api/v1/geo/areas", 503},
	{http.MethodGet, "/api/v1/geo/zonas", 503},
	{http.MethodGet, "/api/v1/geo/capas", 503},
	{http.MethodGet, "/api/v1/geo/capas/jardines_reserva", 503},
	{http.MethodGet, "/api/v1/geo/edificios", 200},
	{http.MethodGet, "/api/v1/geo/inventario", 503},
	{http.MethodGet, "/api/v1/geo/inventario/fotos/x.jpg", 404},
	{http.MethodGet, "/api/v1/geo/inventario/bebederos", 503},
	{http.MethodGet, "/api/v1/geo/reservas-mock", 200},
	{http.MethodGet, "/api/v1/operacion/capataces", 503},
	{http.MethodGet, "/api/v1/operacion/actividades", 503},
	{http.MethodPost, "/api/v1/operacion/actividades", 401},
	{http.MethodPatch, "/api/v1/operacion/actividades/1/asignacion", 401},
	{http.MethodPatch, "/api/v1/operacion/actividades/1/estado", 401},
	{http.MethodPost, "/api/v1/operacion/actividades/1/archivar", 401},
	{http.MethodGet, "/api/v1/operacion/actividades/1/timeline", 503},
	{http.MethodPost, "/api/v1/sesion", 503},
	{http.MethodGet, "/api/v1/sesion", 401},
	{http.MethodDelete, "/api/v1/sesion", 200},
	{http.MethodGet, "/api/v1/accesos/usuarios", 403},
	{http.MethodGet, "/api/v1/catalogos", 401},
	{http.MethodPost, "/api/v1/catalogos", 401},
	{http.MethodPost, "/api/v1/catalogos/1/desactivar", 401},
	{http.MethodGet, "/api/v1/catastro/areas", 401},
	{http.MethodPost, "/api/v1/catastro/areas", 401},
	{http.MethodPatch, "/api/v1/catastro/areas/AV-1", 401},
	{http.MethodGet, "/api/v1/solicitudes", 401},
	{http.MethodPost, "/api/v1/solicitudes", 401},
	{http.MethodGet, "/api/v1/ordenes", 401},
	{http.MethodPost, "/api/v1/ordenes", 401},
	{http.MethodGet, "/api/v1/riego", 401},
	{http.MethodPost, "/api/v1/riego", 401},
	{http.MethodGet, "/api/v1/evidencias", 401},
	{http.MethodPost, "/api/v1/evidencias", 401},
	{http.MethodGet, "/api/v1/evidencias/1/archivo", 401},
	{http.MethodGet, "/api/v1/reportes/labores", 401},
	{http.MethodPost, "/api/v1/ia/sugerir-tipo", 401},
	{http.MethodGet, "/api/v1/no-existe", 404},
}

func TestRutasActualesRespondenIgual(t *testing.T) {
	r := New(Deps{OpenAPIPath: "no-existe.yaml"})

	vistas := map[string]bool{}
	for _, rt := range r.Routes() {
		vistas[rt.Method+" "+rt.Path] = true
	}
	for _, want := range rutasCongeladas {
		if want.codigo == 404 && want.ruta == "/api/v1/no-existe" {
			continue
		}
		key := want.metodo + " " + plantilla(want.ruta)
		if !vistas[key] {
			t.Errorf("falta la ruta registrada %s", key)
		}
	}
	if len(vistas) != 60 {
		t.Fatalf("rutas registradas = %d, se esperaban 60: %v", len(vistas), vistas)
	}

	for _, want := range rutasCongeladas {
		w := httptest.NewRecorder()
		r.ServeHTTP(w, httptest.NewRequest(want.metodo, want.ruta, nil))
		if w.Code != want.codigo {
			t.Errorf("%s %s -> %d, se esperaba %d (%s)", want.metodo, want.ruta, w.Code, want.codigo, w.Body.String())
		}
	}

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodOptions, "/health", nil))
	if w.Code != http.StatusNoContent {
		t.Fatalf("options %d", w.Code)
	}
}

func TestCORSSoloReflejaElOrigenConfigurado(t *testing.T) {
	r := New(Deps{
		OpenAPIPath: "no-existe.yaml",
		Seguridad:   handlers.OpcionesDeSeguridad{CORSOrigins: []string{"http://localhost:5173"}},
	})
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1", nil)
	req.Header.Set("Origin", "http://localhost:5173")
	r.ServeHTTP(w, req)
	if got := w.Header().Get("Access-Control-Allow-Origin"); got != "http://localhost:5173" {
		t.Fatalf("origen %q", got)
	}
	if strings.Contains(w.Header().Get("Access-Control-Allow-Origin"), "*") {
		t.Fatal("no debe reflejar asterisco")
	}

	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/api/v1", nil)
	req.Header.Set("Origin", "https://otro.example")
	r.ServeHTTP(w, req)
	if got := w.Header().Get("Access-Control-Allow-Origin"); got != "" {
		t.Fatalf("origen ajeno %q", got)
	}
}

func TestLoginTieneRateLimit(t *testing.T) {
	r := New(Deps{
		OpenAPIPath: "no-existe.yaml",
		Seguridad:   handlers.OpcionesDeSeguridad{LoginCada: 2, LoginVentana: time.Minute},
	})
	for i := 0; i < 2; i++ {
		w := httptest.NewRecorder()
		r.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/api/v1/sesion", nil))
		if w.Code == http.StatusTooManyRequests {
			t.Fatalf("intento %d cortado antes de tiempo", i)
		}
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/api/v1/sesion", nil))
	if w.Code != http.StatusTooManyRequests {
		t.Fatalf("rate limit %d %s", w.Code, w.Body.String())
	}
}

func plantilla(ruta string) string {
	switch ruta {
	case "/api/v1/geo/capas/jardines_reserva":
		return "/api/v1/geo/capas/:capa"
	case "/api/v1/geo/inventario/fotos/x.jpg":
		return "/api/v1/geo/inventario/fotos/:name"
	case "/api/v1/geo/inventario/bebederos":
		return "/api/v1/geo/inventario/:capa"
	case "/api/v1/operacion/actividades/1/asignacion":
		return "/api/v1/operacion/actividades/:id/asignacion"
	case "/api/v1/operacion/actividades/1/estado":
		return "/api/v1/operacion/actividades/:id/estado"
	case "/api/v1/operacion/actividades/1/archivar":
		return "/api/v1/operacion/actividades/:id/archivar"
	case "/api/v1/operacion/actividades/1/timeline":
		return "/api/v1/operacion/actividades/:id/timeline"
	case "/api/v1/catalogos/1/desactivar":
		return "/api/v1/catalogos/:id/desactivar"
	case "/api/v1/catastro/areas/AV-1":
		return "/api/v1/catastro/areas/:id"
	case "/api/v1/evidencias/1/archivo":
		return "/api/v1/evidencias/:id/archivo"
	default:
		return ruta
	}
}
