package middleware_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/dto"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/constants/enums"
	domainErrors "github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/errors"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/presentation/middleware"
)

type mockSesionUseCase struct {
	usuarios map[string]*dto.UsuarioSesionDTO
}

func (m *mockSesionUseCase) Login(ctx context.Context, usuario, clave string) (string, *dto.UsuarioSesionDTO, error) {
	return "", nil, nil
}

func (m *mockSesionUseCase) Actual(ctx context.Context, token string) (*dto.UsuarioSesionDTO, error) {
	return m.Resolver(ctx, token)
}

func (m *mockSesionUseCase) Logout(ctx context.Context, token string) {}

func (m *mockSesionUseCase) Resolver(ctx context.Context, token string) (*dto.UsuarioSesionDTO, error) {
	u, ok := m.usuarios[token]
	if !ok {
		return nil, domainErrors.ErrSinSesion
	}
	return u, nil
}

func TestAuthMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockUC := &mockSesionUseCase{
		usuarios: map[string]*dto.UsuarioSesionDTO{
			"valid-token": {
				ID:        1,
				Usuario:   "admin",
				Nombre:    "Administrador",
				Rol:       enums.RolAdmin.String(),
				RolNombre: "Administrador del sistema",
			},
		},
	}

	router := gin.New()
	router.Use(middleware.Auth(mockUC))
	router.GET("/test-auth", func(c *gin.Context) {
		u, ok := middleware.UsuarioEn(c)
		if !ok {
			c.JSON(http.StatusOK, gin.H{"autenticado": false})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"autenticado": true,
			"usuario":     u.Usuario,
			"rol":         u.Rol,
		})
	})

	// 1. Sin cookie
	req := httptest.NewRequest(http.MethodGet, "/test-auth", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("status %d", w.Code)
	}
	if w.Body.String() != `{"autenticado":false}` {
		t.Fatalf("esperado no autenticado, obtuve: %s", w.Body.String())
	}

	// 2. Con token inválido
	req = httptest.NewRequest(http.MethodGet, "/test-auth", nil)
	req.AddCookie(&http.Cookie{Name: "cv_sesion", Value: "token-invalido"})
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("status %d", w.Code)
	}
	if w.Body.String() != `{"autenticado":false}` {
		t.Fatalf("esperado no autenticado, obtuve: %s", w.Body.String())
	}

	// 3. Con token válido
	req = httptest.NewRequest(http.MethodGet, "/test-auth", nil)
	req.AddCookie(&http.Cookie{Name: "cv_sesion", Value: "valid-token"})
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("status %d", w.Code)
	}
	expected := `{"autenticado":true,"rol":"admin","usuario":"admin"}`
	if w.Body.String() != expected {
		t.Fatalf("esperado %s, obtuve: %s", expected, w.Body.String())
	}
}
