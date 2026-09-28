package handlers

import (
	"database/sql"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"campusverde/api/internal/accesos"
	"campusverde/api/internal/atencion"
	"campusverde/api/internal/db"
	"campusverde/api/internal/migrate"
	"campusverde/api/internal/operacion"

	"github.com/gin-gonic/gin"
)

func TestCreateSinCookieEs401(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := Operacion{}
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	body := `{"id":"11111111-1111-4111-8111-111111111111","tipo":"riego","titulo":"Riego","lon":-77.08,"lat":-12.07,"actor_rol":"jefatura"}`
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/operacion/actividades", strings.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")
	h.Create(c)
	if w.Code != 401 {
		t.Fatalf("código %d cuerpo %s", w.Code, w.Body.String())
	}
}

func TestCreateCapatazConSesionEs403(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := Operacion{}
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("usuario", accesos.Usuario{Rol: "capataz", CapatazID: "cap-norte", Usuario: "norte"})
	body := `{"id":"11111111-1111-4111-8111-111111111111","tipo":"riego","titulo":"Riego","lon":-77.08,"lat":-12.07,"actor_rol":"jefatura"}`
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/operacion/actividades", strings.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")
	h.Create(c)
	if w.Code != 403 {
		t.Fatalf("código %d cuerpo %s", w.Code, w.Body.String())
	}
}

func TestEstadoCapatazSinCapatazIDRetorna403(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := Operacion{}
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("usuario", accesos.Usuario{Rol: "capataz", CapatazID: "", Usuario: "sin-id"})
	body := `{"estado":"en_proceso","capataz_id":"cap-norte"}`
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/operacion/actividades/11111111-1111-4111-8111-111111111111/estado", strings.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")
	h.Estado(c)
	if w.Code != 403 {
		t.Fatalf("código %d cuerpo %s", w.Code, w.Body.String())
	}
	if !strings.Contains(w.Body.String(), "sin identificador asignado") {
		t.Fatalf("mensaje inesperado: %s", w.Body.String())
	}
}

func TestListSinSesionEs401(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := Operacion{}
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/operacion/actividades", nil)
	h.List(c)
	if w.Code != 401 {
		t.Fatalf("código %d cuerpo %s", w.Code, w.Body.String())
	}
}

func TestTimelineSinSesionEs401(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := Operacion{}
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/operacion/actividades/1/timeline", nil)
	h.Timeline(c)
	if w.Code != 401 {
		t.Fatalf("código %d cuerpo %s", w.Code, w.Body.String())
	}
}

func TestCapatacesSinSesionEs401(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := Operacion{}
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/operacion/capataces", nil)
	h.Capataces(c)
	if w.Code != 401 {
		t.Fatalf("código %d cuerpo %s", w.Code, w.Body.String())
	}
}

func TestListCapatazSinEquipo(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := Operacion{}
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("usuario", accesos.Usuario{Rol: "coordinacion", Usuario: "coord"})
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/operacion/actividades?rol=capataz", nil)
	h.List(c)
	if w.Code != 400 {
		t.Fatalf("sin capataz_id código esperado 400, obtuve %d: %s", w.Code, w.Body.String())
	}

	w = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(w)
	c.Set("usuario", accesos.Usuario{Rol: "coordinacion", Usuario: "coord"})
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/operacion/actividades?rol=capataz&capataz_id=cap-norte", nil)
	h.List(c)
	if w.Code != 503 {
		t.Fatalf("con capataz_id sin base disponible: código %d cuerpo %s", w.Code, w.Body.String())
	}
}

func TestCapatazNoPuedeCerrarNiCancelar(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := Operacion{}

	for _, estado := range []string{"cerrada", "cancelada"} {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Set("usuario", accesos.Usuario{Rol: "capataz", CapatazID: "cap-norte", Usuario: "norte"})
		c.Params = gin.Params{{Key: "id", Value: "11111111-1111-4111-8111-111111111111"}}
		c.Request = httptest.NewRequest(http.MethodPatch, "/api/v1/operacion/actividades/1/estado",
			strings.NewReader(`{"estado":"`+estado+`","capataz_id":"cap-norte"}`))
		c.Request.Header.Set("Content-Type", "application/json")
		h.Estado(c)
		if w.Code != 403 {
			t.Fatalf("estado %s debía ser 403, fue %d: %s", estado, w.Code, w.Body.String())
		}
		if !strings.Contains(w.Body.String(), "el capataz no puede cerrar ni cancelar una labor") {
			t.Fatalf("mensaje esperado no encontrado: %s", w.Body.String())
		}
	}
}

func TestCapatazFichaYAvancesPermisos(t *testing.T) {
	base := os.Getenv("MIGRATE_TEST_URL")
	if base == "" {
		base = "postgres://campus:campus@127.0.0.1:5432/postgres?sslmode=disable"
	}
	admin, err := sql.Open("pgx", base)
	if err != nil {
		t.Fatal(err)
	}
	defer admin.Close()
	if err := admin.Ping(); err != nil {
		t.Skipf("sin postgres de prueba: %v", err)
	}
	name := "campus_verde_test_op_perm"
	if _, err := admin.Exec(`SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE datname = $1 AND pid <> pg_backend_pid()`, name); err != nil {
		t.Fatal(err)
	}
	if _, err := admin.Exec("DROP DATABASE IF EXISTS " + name); err != nil {
		t.Fatal(err)
	}
	if _, err := admin.Exec("CREATE DATABASE " + name); err != nil {
		t.Fatal(err)
	}
	defer func() {
		_, _ = admin.Exec(`SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE datname = $1 AND pid <> pg_backend_pid()`, name)
		_, _ = admin.Exec("DROP DATABASE IF EXISTS " + name)
	}()

	gdb, err := db.Open(fmt.Sprintf("postgres://campus:campus@127.0.0.1:5432/%s?sslmode=disable", name))
	if err != nil {
		t.Fatal(err)
	}
	dir := filepath.Join("..", "..", "migrations")
	if err := migrate.Apply(gdb, dir); err != nil {
		t.Fatal(err)
	}

	gin.SetMode(gin.TestMode)
	opHandler := Operacion{Store: operacion.NewStore(gdb)}
	podaHandler := PodaVivero{Store: atencion.NewStore(gdb)}
	laborNorte := "11111111-1111-4111-8111-111111111111" // asignada a cap-norte en 003

	// 1. PATCH /ficha por cap-sur (capataz ajeno) -> 403
	{
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Set("usuario", accesos.Usuario{Rol: "capataz", CapatazID: "cap-sur", Usuario: "sur"})
		c.Params = gin.Params{{Key: "id", Value: laborNorte}}
		c.Request = httptest.NewRequest(http.MethodPatch, "/api/v1/operacion/actividades/"+laborNorte+"/ficha",
			strings.NewReader(`{"comentario":"intento sur"}`))
		c.Request.Header.Set("Content-Type", "application/json")
		opHandler.Ficha(c)
		if w.Code != 403 {
			t.Fatalf("capataz ajeno en ficha status esperado 403, obtuve %d: %s", w.Code, w.Body.String())
		}
	}

	// 2. PATCH /ficha por cap-norte (asignado) -> 200
	{
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Set("usuario", accesos.Usuario{Rol: "capataz", CapatazID: "cap-norte", Usuario: "norte"})
		c.Params = gin.Params{{Key: "id", Value: laborNorte}}
		c.Request = httptest.NewRequest(http.MethodPatch, "/api/v1/operacion/actividades/"+laborNorte+"/ficha",
			strings.NewReader(`{"comentario":"nota de capataz asignado"}`))
		c.Request.Header.Set("Content-Type", "application/json")
		opHandler.Ficha(c)
		if w.Code != 200 {
			t.Fatalf("capataz asignado en ficha status esperado 200, obtuve %d: %s", w.Code, w.Body.String())
		}
	}

	// 3. PATCH /ficha por coordinacion (oficina) -> 200
	{
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Set("usuario", accesos.Usuario{Rol: "coordinacion", Usuario: "coord"})
		c.Params = gin.Params{{Key: "id", Value: laborNorte}}
		c.Request = httptest.NewRequest(http.MethodPatch, "/api/v1/operacion/actividades/"+laborNorte+"/ficha",
			strings.NewReader(`{"comentario":"nota de coordinacion"}`))
		c.Request.Header.Set("Content-Type", "application/json")
		opHandler.Ficha(c)
		if w.Code != 200 {
			t.Fatalf("coordinacion en ficha status esperado 200, obtuve %d: %s", w.Code, w.Body.String())
		}
	}

	// 4. POST /avances por cap-sur (capataz ajeno) -> 403
	{
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Set("usuario", accesos.Usuario{Rol: "capataz", CapatazID: "cap-sur", Usuario: "sur"})
		c.Params = gin.Params{{Key: "id", Value: laborNorte}}
		body := `{"id":"22222222-2222-4222-8222-222222222221","fecha":"2026-09-28","nota":"intento avance sur","area_feature_id":"AV-0001"}`
		c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/operacion/actividades/"+laborNorte+"/avances",
			strings.NewReader(body))
		c.Request.Header.Set("Content-Type", "application/json")
		podaHandler.CrearAvance(c)
		if w.Code != 403 {
			t.Fatalf("capataz ajeno en avance status esperado 403, obtuve %d: %s", w.Code, w.Body.String())
		}
	}

	// 5. POST /avances por cap-norte (asignado) -> 201
	{
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Set("usuario", accesos.Usuario{Rol: "capataz", CapatazID: "cap-norte", Usuario: "norte"})
		c.Params = gin.Params{{Key: "id", Value: laborNorte}}
		body := `{"id":"22222222-2222-4222-8222-222222222222","fecha":"2026-09-28","nota":"avance norte","area_feature_id":"AV-0001"}`
		c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/operacion/actividades/"+laborNorte+"/avances",
			strings.NewReader(body))
		c.Request.Header.Set("Content-Type", "application/json")
		podaHandler.CrearAvance(c)
		if w.Code != 201 {
			t.Fatalf("capataz asignado en avance status esperado 201, obtuve %d: %s", w.Code, w.Body.String())
		}
	}

	// 6. POST /avances por coordinacion (oficina) -> 201
	{
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Set("usuario", accesos.Usuario{Rol: "coordinacion", Usuario: "coord"})
		c.Params = gin.Params{{Key: "id", Value: laborNorte}}
		body := `{"id":"22222222-2222-4222-8222-222222222223","fecha":"2026-09-28","nota":"avance coord","area_feature_id":"AV-0001"}`
		c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/operacion/actividades/"+laborNorte+"/avances",
			strings.NewReader(body))
		c.Request.Header.Set("Content-Type", "application/json")
		podaHandler.CrearAvance(c)
		if w.Code != 201 {
			t.Fatalf("coordinacion en avance status esperado 201, obtuve %d: %s", w.Code, w.Body.String())
		}
	}

	// 7. Configurar laborNorte como tercerizada para probar ordenes
	if err := gdb.Exec(`UPDATE actividades SET ejecutor = 'tercerizada' WHERE id = $1`, laborNorte).Error; err != nil {
		t.Fatal(err)
	}
	atHandler := Atencion{Store: atencion.NewStore(gdb)}
	ordenID := "33333333-3333-4333-8333-333333333331"

	// 8. POST /ordenes por cap-sur (capataz ajeno) -> 403
	{
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Set("usuario", accesos.Usuario{Rol: "capataz", CapatazID: "cap-sur", Usuario: "sur"})
		body := fmt.Sprintf(`{"id":"%s","actividad_id":"%s","empresa":"Eulen","referencia":"ORD-01"}`, ordenID, laborNorte)
		c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/ordenes", strings.NewReader(body))
		c.Request.Header.Set("Content-Type", "application/json")
		atHandler.CrearOrden(c)
		if w.Code != 403 {
			t.Fatalf("capataz ajeno en crear orden esperado 403, obtuve %d: %s", w.Code, w.Body.String())
		}
	}

	// 9. POST /ordenes por cap-norte (asignado) -> 201
	{
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Set("usuario", accesos.Usuario{Rol: "capataz", CapatazID: "cap-norte", Usuario: "norte"})
		body := fmt.Sprintf(`{"id":"%s","actividad_id":"%s","empresa":"Eulen","referencia":"ORD-01"}`, ordenID, laborNorte)
		c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/ordenes", strings.NewReader(body))
		c.Request.Header.Set("Content-Type", "application/json")
		atHandler.CrearOrden(c)
		if w.Code != 201 {
			t.Fatalf("capataz asignado en crear orden esperado 201, obtuve %d: %s", w.Code, w.Body.String())
		}
	}

	// 10. PATCH /ordenes/:id por cap-sur (capataz ajeno) -> 403
	{
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Set("usuario", accesos.Usuario{Rol: "capataz", CapatazID: "cap-sur", Usuario: "sur"})
		c.Params = gin.Params{{Key: "id", Value: ordenID}}
		c.Request = httptest.NewRequest(http.MethodPatch, "/api/v1/ordenes/"+ordenID,
			strings.NewReader(`{"conformidad":"conforme","estado":"conforme"}`))
		c.Request.Header.Set("Content-Type", "application/json")
		podaHandler.EditarOrden(c)
		if w.Code != 403 {
			t.Fatalf("capataz ajeno en editar orden esperado 403, obtuve %d: %s", w.Code, w.Body.String())
		}
	}

	// 11. PATCH /ordenes/:id por cap-norte (asignado) -> 200
	{
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Set("usuario", accesos.Usuario{Rol: "capataz", CapatazID: "cap-norte", Usuario: "norte"})
		c.Params = gin.Params{{Key: "id", Value: ordenID}}
		c.Request = httptest.NewRequest(http.MethodPatch, "/api/v1/ordenes/"+ordenID,
			strings.NewReader(`{"conformidad":"conforme","estado":"conforme"}`))
		c.Request.Header.Set("Content-Type", "application/json")
		podaHandler.EditarOrden(c)
		if w.Code != 200 {
			t.Fatalf("capataz asignado en editar orden esperado 200, obtuve %d: %s", w.Code, w.Body.String())
		}
	}

	// 12. Coordinación cierra la labor -> 200
	{
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Set("usuario", accesos.Usuario{Rol: "coordinacion", Usuario: "coord"})
		c.Params = gin.Params{{Key: "id", Value: laborNorte}}
		c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/operacion/actividades/"+laborNorte+"/estado",
			strings.NewReader(`{"estado":"cerrada"}`))
		c.Request.Header.Set("Content-Type", "application/json")
		opHandler.Estado(c)
		if w.Code != 200 {
			t.Fatalf("coordinacion cerrar labor esperado 200, obtuve %d: %s", w.Code, w.Body.String())
		}
	}

	// 13. Capataz asignado (cap-norte) intenta reabrir labor cerrada a pendiente -> 403
	{
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Set("usuario", accesos.Usuario{Rol: "capataz", CapatazID: "cap-norte", Usuario: "norte"})
		c.Params = gin.Params{{Key: "id", Value: laborNorte}}
		c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/operacion/actividades/"+laborNorte+"/estado",
			strings.NewReader(`{"estado":"pendiente"}`))
		c.Request.Header.Set("Content-Type", "application/json")
		opHandler.Estado(c)
		if w.Code != 403 {
			t.Fatalf("capataz reabriendo labor cerrada esperado 403, obtuve %d: %s", w.Code, w.Body.String())
		}
	}

	// 14. Coordinación sí puede reabrir labor cerrada a pendiente -> 200
	{
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Set("usuario", accesos.Usuario{Rol: "coordinacion", Usuario: "coord"})
		c.Params = gin.Params{{Key: "id", Value: laborNorte}}
		c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/operacion/actividades/"+laborNorte+"/estado",
			strings.NewReader(`{"estado":"pendiente"}`))
		c.Request.Header.Set("Content-Type", "application/json")
		opHandler.Estado(c)
		if w.Code != 200 {
			t.Fatalf("coordinacion reabriendo labor cerrada esperado 200, obtuve %d: %s", w.Code, w.Body.String())
		}
	}
}
