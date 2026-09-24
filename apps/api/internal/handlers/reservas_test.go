package handlers

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestReservasSinSheet(t *testing.T) {
	gin.SetMode(gin.TestMode)
	dir := t.TempDir()
	path := filepath.Join(dir, "mock.json")
	body := []byte(`{
	  "_meta": {"fake": true, "warning": "SHEET_ID_RESERVAS=1R3Xz8A5xIVm-s0duMQhav2YAJrQdlSAYgKYeoEvoT8s"},
	  "reservas": [{
	    "id": "FAKE-RES-001",
	    "jardin": "Jardín Rosales",
	    "fecha": "2026-09-22",
	    "hora": "15:00-17:00",
	    "evento": "Feria",
	    "estado": "Confirmado",
	    "notas": "viene del Sheet real"
	  }]
	}`)
	if err := os.WriteFile(path, body, 0o644); err != nil {
		t.Fatal(err)
	}
	h := Reservas{Path: path}
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/geo/reservas-mock", nil)
	h.Get(c)
	if w.Code != 200 {
		t.Fatalf("código %d %s", w.Code, w.Body.String())
	}
	text := w.Body.String()
	for _, forbidden := range []string{"SHEET", "1R3Xz", "docs.google.com", "viene del Sheet"} {
		if strings.Contains(text, forbidden) {
			t.Fatalf("se filtró %q en %s", forbidden, text)
		}
	}
	if !strings.Contains(text, "Agenda ficticia") || !strings.Contains(text, "Jardín Rosales") {
		t.Fatalf("respuesta %s", text)
	}
}
