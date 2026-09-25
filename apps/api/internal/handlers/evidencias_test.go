package handlers

import (
	"bytes"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"campusverde/api/internal/accesos"
	"campusverde/api/internal/atencion"
	"campusverde/api/internal/blobs"
	"campusverde/api/internal/db"
	"campusverde/api/internal/migrate"

	"github.com/gin-gonic/gin"
	_ "github.com/jackc/pgx/v5/stdlib"
)

func TestSubirEvidencia409YIdempotente(t *testing.T) {
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
	name := "campus_verde_evidencia"
	if _, err := admin.Exec(`SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE datname = $1 AND pid <> pg_backend_pid()`, name); err != nil {
		t.Fatal(err)
	}
	if _, err := admin.Exec("DROP DATABASE IF EXISTS " + name); err != nil {
		t.Fatal(err)
	}
	if _, err := admin.Exec("CREATE DATABASE " + name); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_, _ = admin.Exec(`SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE datname = $1 AND pid <> pg_backend_pid()`, name)
		_, _ = admin.Exec("DROP DATABASE IF EXISTS " + name)
	})

	url := fmt.Sprintf("postgres://campus:campus@127.0.0.1:5432/%s?sslmode=disable", name)
	gdb, err := db.Open(url)
	if err != nil {
		t.Fatal(err)
	}
	sqlDB, err := gdb.DB()
	if err != nil {
		t.Fatal(err)
	}
	defer sqlDB.Close()
	dir := filepath.Join("..", "..", "migrations")
	if err := migrate.Apply(gdb, dir); err != nil {
		t.Fatal(err)
	}

	gin.SetMode(gin.TestMode)
	h := Atencion{Store: atencion.NewStore(gdb), Files: blobs.Disk{Dir: t.TempDir()}}
	labor := "11111111-1111-4111-8111-111111111111"
	id := "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa"
	foto := []byte{0xFF, 0xD8, 0xFF, 0xD9}
	otra := []byte{0xFF, 0xD8, 0xFF, 0x00, 0xD9}

	primero := postEvidencia(t, h, id, labor, foto, "capataz", "cap-norte")
	if primero.Code != 201 {
		t.Fatalf("alta: %d %s", primero.Code, primero.Body.String())
	}
	var n int
	if err := gdb.Raw(`SELECT count(*) FROM actividad_eventos WHERE actividad_id = $1 AND tipo = 'evidencia'`, labor).Scan(&n).Error; err != nil {
		t.Fatal(err)
	}
	if n != 1 {
		t.Fatalf("eventos evidencia = %d", n)
	}

	igual := postEvidencia(t, h, id, labor, foto, "capataz", "cap-norte")
	if igual.Code != 200 {
		t.Fatalf("idempotente: %d %s", igual.Code, igual.Body.String())
	}
	var body map[string]any
	if err := json.Unmarshal(igual.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body["idempotente"] != true {
		t.Fatalf("respuesta: %s", igual.Body.String())
	}
	if err := gdb.Raw(`SELECT count(*) FROM actividad_eventos WHERE actividad_id = $1 AND tipo = 'evidencia'`, labor).Scan(&n).Error; err != nil {
		t.Fatal(err)
	}
	if n != 1 {
		t.Fatalf("el reintento duplicó la bitácora: %d", n)
	}

	conflicto := postEvidencia(t, h, id, labor, otra, "capataz", "cap-norte")
	if conflicto.Code != 409 {
		t.Fatalf("409: %d %s", conflicto.Code, conflicto.Body.String())
	}

	ajena := postEvidencia(t, h, "bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbbb", labor, foto, "capataz", "cap-sur")
	if ajena.Code != 403 {
		t.Fatalf("cuadrilla: %d %s", ajena.Code, ajena.Body.String())
	}
}

func postEvidencia(t *testing.T, h Atencion, id, labor string, foto []byte, rol, capataz string) *httptest.ResponseRecorder {
	t.Helper()
	sum := sha256.Sum256(foto)
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	_ = w.WriteField("id", id)
	_ = w.WriteField("actividad_id", labor)
	_ = w.WriteField("sha256", hex.EncodeToString(sum[:]))
	_ = w.WriteField("lat", "-12.06955")
	_ = w.WriteField("lon", "-77.07955")
	_ = w.WriteField("exif", `{"origen":"exif"}`)
	part, err := w.CreateFormFile("archivo", "foto.jpg")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := part.Write(foto); err != nil {
		t.Fatal(err)
	}
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest(http.MethodPost, "/api/v1/evidencias", &buf)
	req.Header.Set("Content-Type", w.FormDataContentType())
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = req
	c.Set("usuario", accesos.Usuario{Rol: rol, CapatazID: capataz, Nombre: "Equipo de prueba"})
	h.SubirEvidencia(c)
	return rec
}
