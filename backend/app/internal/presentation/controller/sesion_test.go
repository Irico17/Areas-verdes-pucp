package controller_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/dto"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/constants/enums"
	domainErrors "github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/errors"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/presentation/controller"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/presentation/middleware"
)

type mockSesionControllerUseCase struct {
	loginFn    func(ctx context.Context, usuario, clave string) (string, *dto.UsuarioSesionDTO, error)
	resolverFn func(ctx context.Context, token string) (*dto.UsuarioSesionDTO, error)
	logoutFn   func(ctx context.Context, token string)
}

func (m *mockSesionControllerUseCase) Login(ctx context.Context, usuario, clave string) (string, *dto.UsuarioSesionDTO, error) {
	if m.loginFn != nil {
		return m.loginFn(ctx, usuario, clave)
	}
	return "", nil, domainErrors.ErrCredencialesInvalidas
}

func (m *mockSesionControllerUseCase) Actual(ctx context.Context, token string) (*dto.UsuarioSesionDTO, error) {
	return m.Resolver(ctx, token)
}

func (m *mockSesionControllerUseCase) Resolver(ctx context.Context, token string) (*dto.UsuarioSesionDTO, error) {
	if m.resolverFn != nil {
		return m.resolverFn(ctx, token)
	}
	return nil, domainErrors.ErrSinSesion
}

func (m *mockSesionControllerUseCase) Logout(ctx context.Context, token string) {
	if m.logoutFn != nil {
		m.logoutFn(ctx, token)
	}
}

func setupSesionRouter(uc *mockSesionControllerUseCase) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(middleware.ConCookie(middleware.OpcionesCookie{
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
	}))
	ctrl := controller.NewSesionController(uc)
	r.POST("/api/v1/sesion", ctrl.Entrar)
	r.GET("/api/v1/sesion", ctrl.Actual)
	r.DELETE("/api/v1/sesion", ctrl.Salir)
	return r
}

func TestSesionController_Entrar(t *testing.T) {
	uc := &mockSesionControllerUseCase{
		loginFn: func(ctx context.Context, usuario, clave string) (string, *dto.UsuarioSesionDTO, error) {
			if usuario == "admin" && clave == "pando-local" {
				return "test-token-123", &dto.UsuarioSesionDTO{
					ID:        6,
					Usuario:   "admin",
					Nombre:    "Administración",
					Rol:       enums.RolAdmin.String(),
					RolNombre: "Administrador del sistema",
				}, nil
			}
			return "", nil, domainErrors.ErrCredencialesInvalidas
		},
	}
	r := setupSesionRouter(uc)

	// 1. JSON inválido -> 400
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/sesion", bytes.NewBufferString("not-json"))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("esperado 400, obtuve %d", w.Code)
	}
	if w.Body.String() != `{"error":"JSON inválido"}` {
		t.Fatalf("cuerpo 400 inesperado: %s", w.Body.String())
	}

	// 2. Credenciales incorrectas -> 401
	w = httptest.NewRecorder()
	body, _ := json.Marshal(map[string]string{"usuario": "admin", "clave": "erronea"})
	req = httptest.NewRequest(http.MethodPost, "/api/v1/sesion", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("esperado 401, obtuve %d", w.Code)
	}
	if w.Body.String() != `{"error":"usuario o clave incorrectos"}` {
		t.Fatalf("cuerpo 401 inesperado: %s", w.Body.String())
	}

	// 3. Login exitoso -> 200 + cookie
	w = httptest.NewRecorder()
	body, _ = json.Marshal(map[string]string{"usuario": "admin", "clave": "pando-local"})
	req = httptest.NewRequest(http.MethodPost, "/api/v1/sesion", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("esperado 200, obtuve %d", w.Code)
	}
	cookies := w.Result().Cookies()
	var foundCookie *http.Cookie
	for _, c := range cookies {
		if c.Name == "cv_sesion" {
			foundCookie = c
			break
		}
	}
	if foundCookie == nil || foundCookie.Value != "test-token-123" {
		t.Fatalf("cookie no establecida correctamente: %v", cookies)
	}
	var res struct {
		Usuario dto.UsuarioSesionDTO `json:"usuario"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &res); err != nil {
		t.Fatal(err)
	}
	if res.Usuario.Usuario != "admin" || res.Usuario.RolNombre != "Administrador del sistema" {
		t.Fatalf("usuario retornado inesperado: %+v", res.Usuario)
	}
}

func TestSesionController_Actual(t *testing.T) {
	uc := &mockSesionControllerUseCase{}
	r := setupSesionRouter(uc)

	// 1. Sin sesión -> 401
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/sesion", nil)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("esperado 401, obtuve %d", w.Code)
	}
	if w.Body.String() != `{"error":"sin sesión"}` {
		t.Fatalf("cuerpo 401 inesperado: %s", w.Body.String())
	}

	// 2. Con sesión seteada en contexto -> 200
	rWithUser := gin.New()
	rWithUser.Use(func(c *gin.Context) {
		c.Set("usuario", dto.UsuarioSesionDTO{
			ID:        4,
			Usuario:   "coordinacion",
			Nombre:    "Coordinación",
			Rol:       enums.RolCoordinacion.String(),
			RolNombre: "Ingeniería/Coordinación",
		})
		c.Next()
	})
	ctrl := controller.NewSesionController(uc)
	rWithUser.GET("/api/v1/sesion", ctrl.Actual)

	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/api/v1/sesion", nil)
	rWithUser.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("esperado 200, obtuve %d", w.Code)
	}
	var res struct {
		Usuario dto.UsuarioSesionDTO `json:"usuario"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &res); err != nil {
		t.Fatal(err)
	}
	if res.Usuario.RolNombre != "Ingeniería/Coordinación" {
		t.Fatalf("rol_nombre inesperado: %+v", res.Usuario)
	}
}

func TestSesionController_Salir(t *testing.T) {
	loggedOutToken := ""
	uc := &mockSesionControllerUseCase{
		logoutFn: func(ctx context.Context, token string) {
			loggedOutToken = token
		},
	}
	r := setupSesionRouter(uc)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodDelete, "/api/v1/sesion", nil)
	req.AddCookie(&http.Cookie{Name: "cv_sesion", Value: "token-para-borrar"})
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("esperado 200, obtuve %d", w.Code)
	}
	if w.Body.String() != `{"ok":true}` {
		t.Fatalf("esperado ok:true, obtuve %s", w.Body.String())
	}
	if loggedOutToken != "token-para-borrar" {
		t.Fatalf("token deslogueado esperado 'token-para-borrar', obtuve %q", loggedOutToken)
	}
	cookies := w.Result().Cookies()
	var foundCookie *http.Cookie
	for _, c := range cookies {
		if c.Name == "cv_sesion" {
			foundCookie = c
			break
		}
	}
	if foundCookie == nil || foundCookie.MaxAge > 0 {
		t.Fatalf("cookie cv_sesion debería tener MaxAge <= 0: %v", foundCookie)
	}
}
