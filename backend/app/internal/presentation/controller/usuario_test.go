package controller_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/dto"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/constants/enums"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/presentation/controller"
)

type mockUsuarioControllerUseCase struct {
	listarFn func(ctx context.Context) (*dto.UsuariosResponseDTO, error)
}

func (m *mockUsuarioControllerUseCase) ListarUsuarios(ctx context.Context) (*dto.UsuariosResponseDTO, error) {
	if m.listarFn != nil {
		return m.listarFn(ctx)
	}
	return &dto.UsuariosResponseDTO{
		Usuarios: []dto.UsuarioSesionDTO{},
		Permisos: []dto.PermisoDTO{},
		Aviso:    "aviso de prueba",
	}, nil
}

func TestUsuarioController_Listar(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockUC := &mockUsuarioControllerUseCase{
		listarFn: func(ctx context.Context) (*dto.UsuariosResponseDTO, error) {
			return &dto.UsuariosResponseDTO{
				Usuarios: []dto.UsuarioSesionDTO{
					{ID: 1, Usuario: "admin", Rol: enums.RolAdmin.String(), RolNombre: "Administrador"},
				},
				Permisos: []dto.PermisoDTO{
					{Rol: enums.RolAdmin.String(), Accion: "consultar"},
				},
				Aviso: "Cuentas locales de desarrollo. El SSO de la PUCP sigue pendiente de validación.",
			}, nil
		},
	}

	ctrl := controller.NewUsuarioController(mockUC)

	// 1. Sin sesión -> 403
	rNoAuth := gin.New()
	rNoAuth.GET("/api/v1/accesos/usuarios", ctrl.Listar)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/accesos/usuarios", nil)
	rNoAuth.ServeHTTP(w, req)
	if w.Code != http.StatusForbidden {
		t.Fatalf("esperado 403, obtuve %d", w.Code)
	}
	if w.Body.String() != `{"error":"solo administración ve las cuentas"}` {
		t.Fatalf("cuerpo 403 inesperado: %s", w.Body.String())
	}

	// 2. Rol capataz -> 403
	rCapataz := gin.New()
	rCapataz.Use(func(c *gin.Context) {
		c.Set("usuario", dto.UsuarioSesionDTO{
			ID:      2,
			Usuario: "norte",
			Rol:     enums.RolCapataz.String(),
		})
		c.Next()
	})
	rCapataz.GET("/api/v1/accesos/usuarios", ctrl.Listar)

	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/api/v1/accesos/usuarios", nil)
	rCapataz.ServeHTTP(w, req)
	if w.Code != http.StatusForbidden {
		t.Fatalf("esperado 403, obtuve %d", w.Code)
	}
	if w.Body.String() != `{"error":"solo administración ve las cuentas"}` {
		t.Fatalf("cuerpo 403 inesperado: %s", w.Body.String())
	}

	// 3. Rol admin -> 200
	rAdmin := gin.New()
	rAdmin.Use(func(c *gin.Context) {
		c.Set("usuario", dto.UsuarioSesionDTO{
			ID:      1,
			Usuario: "admin",
			Rol:     enums.RolAdmin.String(),
		})
		c.Next()
	})
	rAdmin.GET("/api/v1/accesos/usuarios", ctrl.Listar)

	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/api/v1/accesos/usuarios", nil)
	rAdmin.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("esperado 200, obtuve %d", w.Code)
	}
	var res dto.UsuariosResponseDTO
	if err := json.Unmarshal(w.Body.Bytes(), &res); err != nil {
		t.Fatal(err)
	}
	if len(res.Usuarios) != 1 || res.Usuarios[0].Usuario != "admin" {
		t.Fatalf("usuarios devueltos inesperados: %+v", res.Usuarios)
	}
	if len(res.Permisos) != 1 || res.Permisos[0].Accion != "consultar" {
		t.Fatalf("permisos devueltos inesperados: %+v", res.Permisos)
	}
}
