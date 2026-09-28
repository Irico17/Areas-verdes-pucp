package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/dto"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/services"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/presentation/middleware"
)

func TestExigePermisoMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	permSvc := services.NewPermisosService()

	setupApp := func() *gin.Engine {
		r := gin.New()
		r.GET("/protegido-catalogos", middleware.ExigePermiso(permSvc, "catalogos"), func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"ok": true})
		})
		r.GET("/protegido-registrar", middleware.ExigePermiso(permSvc, "registrar"), func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"ok": true})
		})
		return r
	}

	app := setupApp()

	// 1. Sin sesión -> 401
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/protegido-catalogos", nil)
	app.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("esperado 401, obtuve %d", w.Code)
	}
	if w.Body.String() != `{"error":"inicie sesión"}` {
		t.Fatalf("cuerpo 401 inesperado: %s", w.Body.String())
	}

	// 2. Rol capataz intentando catalogos -> 403
	appWithCapataz := gin.New()
	appWithCapataz.Use(func(c *gin.Context) {
		c.Set("usuario", dto.UsuarioSesionDTO{
			ID:      1,
			Usuario: "norte",
			Rol:     "capataz",
		})
		c.Next()
	})
	appWithCapataz.GET("/protegido-catalogos", middleware.ExigePermiso(permSvc, "catalogos"), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})
	appWithCapataz.GET("/protegido-registrar", middleware.ExigePermiso(permSvc, "registrar"), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/protegido-catalogos", nil)
	appWithCapataz.ServeHTTP(w, req)
	if w.Code != http.StatusForbidden {
		t.Fatalf("esperado 403, obtuve %d", w.Code)
	}
	if w.Body.String() != `{"error":"su rol no tiene ese permiso"}` {
		t.Fatalf("cuerpo 403 inesperado: %s", w.Body.String())
	}

	// 3. Rol capataz con registrar -> 200
	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/protegido-registrar", nil)
	appWithCapataz.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("esperado 200, obtuve %d", w.Code)
	}
	if w.Body.String() != `{"ok":true}` {
		t.Fatalf("cuerpo 200 inesperado: %s", w.Body.String())
	}
}
