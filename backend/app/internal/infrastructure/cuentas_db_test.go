package infrastructure_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"

	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/services"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/usecases"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/infrastructure/seguridad"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/persistence/database"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/persistence/repository/postgres"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/presentation/controller"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/presentation/middleware"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/presentation/routes/groups"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/shared/testutil"
)

func TestGestionCuentasJefatura(t *testing.T) {
	sqlDB, gdb := testutil.MigrarDBTemporal(t, "cuentas")
	defer sqlDB.Close()

	ctx := context.Background()
	usuarioRepo := postgres.NewUsuarioRepository(gdb)
	permisoRepo := postgres.NewPermisoRepository(gdb)
	sesionRepo := postgres.NewSesionRepository(gdb)
	cambioRepo := postgres.NewCambioRepository(gdb)
	hasher := seguridad.NewBcryptHasher()
	tx := database.NewTransaccion(gdb)
	permisosSvc := services.NewPermisosService(permisoRepo)
	auditoria := services.NewAuditoriaService(cambioRepo)
	semilla := usecases.NewSemillaAccesosUseCase(tx, usuarioRepo, permisoRepo, permisosSvc, hasher)
	if err := semilla.Ensure(ctx, "pando-local"); err != nil {
		t.Fatalf("Ensure: %v", err)
	}

	usuarioUC := usecases.NewUsuarioUseCase(usuarioRepo, permisoRepo, hasher, auditoria, tx)
	sesionUC := usecases.NewSesionUseCase(sesionRepo, usuarioRepo, hasher)
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	engine.Use(middleware.Auth(sesionUC))
	api := engine.Group("/areas-verdes/v1")
	groups.NewSesionGroup(controller.NewSesionController(sesionUC)).Register(api)
	groups.NewAccesosGroup(controller.NewUsuarioController(usuarioUC, permisosSvc, zerolog.Nop())).Register(api)

	// El login de coordinación que ya existía sigue abriendo sesión.
	coord := pedir(engine, http.MethodPost, "/areas-verdes/v1/sesion", "", `{"usuario":"coordinacion","clave":"pando-local"}`)
	if coord.Code != http.StatusOK {
		t.Fatalf("login coordinacion: %d %s", coord.Code, coord.Body.String())
	}

	jef := pedir(engine, http.MethodPost, "/areas-verdes/v1/sesion", "", `{"usuario":"jefatura","clave":"pando-local"}`)
	if jef.Code != http.StatusOK {
		t.Fatalf("login jefatura: %d %s", jef.Code, jef.Body.String())
	}
	cookieJef := cookieDe(t, jef)

	// POST que antes era 404 ahora existe. Sin sesión responde 403, no 404.
	sinSesion := pedir(engine, http.MethodPost, "/areas-verdes/v1/accesos/usuarios", "", `{"usuario":"camila.herrera","nombre":"Camila Herrera","rol":"capataz","clave":"clave-camila-01"}`)
	if sinSesion.Code != http.StatusForbidden {
		t.Fatalf("alta sin sesión: %d %s", sinSesion.Code, sinSesion.Body.String())
	}

	cap := pedir(engine, http.MethodPost, "/areas-verdes/v1/sesion", "", `{"usuario":"norte","clave":"pando-local"}`)
	if cap.Code != http.StatusOK {
		t.Fatalf("login capataz: %d %s", cap.Code, cap.Body.String())
	}
	altaCapataz := pedir(engine, http.MethodPost, "/areas-verdes/v1/accesos/usuarios", cookieDe(t, cap), `{"usuario":"camila.herrera","nombre":"Camila Herrera","rol":"capataz","clave":"clave-camila-01"}`)
	if altaCapataz.Code != http.StatusForbidden {
		t.Fatalf("alta por capataz: %d %s", altaCapataz.Code, altaCapataz.Body.String())
	}

	alta := pedir(engine, http.MethodPost, "/areas-verdes/v1/accesos/usuarios", cookieJef, `{"usuario":"camila.herrera","nombre":"Camila Herrera","rol":"capataz","clave":"clave-camila-01","capataz_id":"cap-norte"}`)
	if alta.Code != http.StatusCreated {
		t.Fatalf("alta por jefatura: %d %s", alta.Code, alta.Body.String())
	}

	nueva := pedir(engine, http.MethodPost, "/areas-verdes/v1/sesion", "", `{"usuario":"camila.herrera","clave":"clave-camila-01"}`)
	if nueva.Code != http.StatusOK {
		t.Fatalf("login cuenta nueva: %d %s", nueva.Code, nueva.Body.String())
	}
	var cuerpo map[string]any
	if err := json.Unmarshal(nueva.Body.Bytes(), &cuerpo); err != nil {
		t.Fatal(err)
	}
	usuario, _ := cuerpo["usuario"].(map[string]any)
	if usuario["debe_cambiar_password"] != true {
		t.Fatalf("la cuenta nueva debe cambiar la clave: %+v", usuario)
	}

	lista := pedir(engine, http.MethodGet, "/areas-verdes/v1/accesos/usuarios", cookieJef, "")
	if lista.Code != http.StatusOK {
		t.Fatalf("listar: %d %s", lista.Code, lista.Body.String())
	}
	if strings.Contains(lista.Body.String(), "SSO") {
		t.Fatalf("el listado no debe mencionar SSO: %s", lista.Body.String())
	}

	baja := pedir(engine, http.MethodPatch, "/areas-verdes/v1/accesos/usuarios/camila.herrera", cookieJef, `{"activo":false}`)
	if baja.Code != http.StatusOK {
		t.Fatalf("baja lógica: %d %s", baja.Code, baja.Body.String())
	}
	inactiva := pedir(engine, http.MethodPost, "/areas-verdes/v1/sesion", "", `{"usuario":"camila.herrera","clave":"clave-camila-01"}`)
	if inactiva.Code != http.StatusUnauthorized {
		t.Fatalf("usuario inactivo no abre sesión: %d %s", inactiva.Code, inactiva.Body.String())
	}

	var cambios int
	if err := gdb.Raw(`SELECT count(*) FROM cambios WHERE entidad = 'usuarios' AND entidad_id = 'camila.herrera'`).Scan(&cambios).Error; err != nil {
		t.Fatal(err)
	}
	if cambios < 2 {
		t.Fatalf("se esperaban alta y baja en cambios, hay %d", cambios)
	}

	// Permite lee la fila vigente: quitar registrar al capataz se ve sin reiniciar.
	if !permisosSvc.Permite("capataz", "registrar") {
		t.Fatal("la semilla concede registrar al capataz")
	}
	if err := gdb.Exec(`DELETE FROM permisos WHERE rol = 'capataz' AND accion = 'registrar'`).Error; err != nil {
		t.Fatal(err)
	}
	if permisosSvc.Permite("capataz", "registrar") {
		t.Fatal("Permite debe recargar la fila sin otro proceso")
	}

	// Un rol inactivo no entra. La cuenta de coordinación sigue en la tabla.
	if err := gdb.Exec(`UPDATE roles SET activo = false WHERE codigo = 'coordinacion'`).Error; err != nil {
		t.Fatal(err)
	}
	bloqueado := pedir(engine, http.MethodPost, "/areas-verdes/v1/sesion", "", `{"usuario":"coordinacion","clave":"pando-local"}`)
	if bloqueado.Code != http.StatusUnauthorized {
		t.Fatalf("rol inactivo: %d %s", bloqueado.Code, bloqueado.Body.String())
	}
	var sigue int
	if err := gdb.Raw(`SELECT count(*) FROM usuarios WHERE usuario = 'coordinacion'`).Scan(&sigue).Error; err != nil {
		t.Fatal(err)
	}
	if sigue != 1 {
		t.Fatal("desactivar el rol no debe borrar la cuenta")
	}
}

func pedir(engine *gin.Engine, metodo, ruta, cookie, body string) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	var req *http.Request
	if body == "" {
		req = httptest.NewRequest(metodo, ruta, nil)
	} else {
		req = httptest.NewRequest(metodo, ruta, bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
	}
	if cookie != "" {
		req.AddCookie(&http.Cookie{Name: middleware.CookieSesion, Value: cookie})
	}
	engine.ServeHTTP(w, req)
	return w
}

func cookieDe(t *testing.T, w *httptest.ResponseRecorder) string {
	t.Helper()
	res := w.Result()
	defer res.Body.Close()
	for _, c := range res.Cookies() {
		if c.Name == middleware.CookieSesion && c.HttpOnly {
			return c.Value
		}
	}
	t.Fatal("la respuesta no trajo la cookie HttpOnly de sesión")
	return ""
}
