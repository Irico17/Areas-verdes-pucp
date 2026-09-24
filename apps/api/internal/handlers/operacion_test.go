package handlers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestCreateCapatazSinBase(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := Operacion{}
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	body := `{"id":"11111111-1111-4111-8111-111111111111","tipo":"riego","titulo":"Riego","lon":-77.08,"lat":-12.07,"actor_rol":"capataz"}`
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/operacion/actividades", strings.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")
	h.Create(c)
	if w.Code != 403 {
		t.Fatalf("código %d cuerpo %s", w.Code, w.Body.String())
	}
}

func TestListCapatazSinEquipo(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := Operacion{}
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/operacion/actividades?rol=capataz", nil)
	h.List(c)
	if w.Code != 400 {
		t.Fatalf("código %d cuerpo %s", w.Code, w.Body.String())
	}
}
