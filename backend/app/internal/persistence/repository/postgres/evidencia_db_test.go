package postgres

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"strings"
	"testing"

	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/entities"
	domainErrors "github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/errors"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/infrastructure/storage"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/shared/testutil"
)

func TestSubirEvidencia409YIdempotente(t *testing.T) {
	_, gdb := testutil.MigrarDBTemporal(t, "evid_idemp")

	filesDir := t.TempDir()
	store := storage.NewDiscoStorage(filesDir)
	repo := NewEvidenciaRepository(gdb)

	labor := "11111111-1111-4111-8111-111111111111"
	id := "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa"
	orden := "cccccccc-cccc-4ccc-8ccc-cccccccccccc"
	foto := []byte{0xFF, 0xD8, 0xFF, 0xD9}
	otra := []byte{0xFF, 0xD8, 0xFF, 0x00, 0xD9}

	var uid int64
	if err := gdb.Raw(`
		INSERT INTO usuarios (usuario, nombre, rol, capataz_id, password_hash)
		VALUES ('prueba-campo', 'Equipo de prueba', 'capataz', 'cap-norte', 'x')
		RETURNING id`).Row().Scan(&uid); err != nil {
		t.Fatal(err)
	}
	if err := gdb.Exec(`
		INSERT INTO ordenes_servicio (id, actividad_id, empresa, referencia)
		VALUES ($1, $2, 'Cuadrilla de prueba', 'OS-1')`, orden, labor).Error; err != nil {
		t.Fatal(err)
	}

	sum := sha256.Sum256(foto)
	hash := hex.EncodeToString(sum[:])
	lat := -12.06955
	lon := -77.07955
	exifStr := `{"fecha":"2026:09:25 10:00:00","lat":-12.07}`

	// 1. Alta inicial
	primero, err := repo.Guardar(context.Background(), entities.GuardarEvidencia{
		ID:          id,
		ActividadID: labor,
		OrdenID:     orden,
		Nombre:      "foto.jpg",
		Mime:        "image/jpeg",
		Ext:         ".jpg",
		Bytes:       len(foto),
		Contenido:   foto,
		Nota:        "Primera subida",
		Hash:        hash,
		Lat:         &lat,
		Lon:         &lon,
		Exif:        exifStr,
		Rol:         "capataz",
		CapatazID:   "cap-norte",
		UsuarioID:   uid,
	}, store)
	if err != nil {
		t.Fatalf("alta falló: %v", err)
	}
	if primero.ID != id || primero.Idempotente {
		t.Fatalf("alta inesperada: %+v", primero)
	}

	var n int
	if err := gdb.Raw(`SELECT count(*) FROM actividad_eventos WHERE actividad_id = $1 AND tipo = 'evidencia'`, labor).Scan(&n).Error; err != nil {
		t.Fatal(err)
	}
	if n != 1 {
		t.Fatalf("eventos evidencia = %d", n)
	}
	var gotUser int64
	var ruta, suma, exif string
	var ordenGuardada string
	if err := gdb.Raw(`
		SELECT usuario_id FROM actividad_eventos
		WHERE actividad_id = $1 AND tipo = 'evidencia'`, labor).Row().Scan(&gotUser); err != nil {
		t.Fatal(err)
	}
	if gotUser != uid {
		t.Fatalf("usuario_id = %d", gotUser)
	}
	if err := gdb.Raw(`
		SELECT ruta, sha256, COALESCE(exif::text, ''), COALESCE(orden_id::text, '')
		FROM evidencias WHERE id = $1`, id).Row().Scan(&ruta, &suma, &exif, &ordenGuardada); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(ruta, suma) {
		t.Fatalf("la clave no deriva del hash: %s", ruta)
	}
	if strings.Contains(exif, "modelo") || !strings.Contains(exif, "fecha") {
		t.Fatalf("exif filtrado: %s", exif)
	}
	if ordenGuardada != orden {
		t.Fatalf("orden_id = %s", ordenGuardada)
	}

	// 2. Reintento idéntico -> idempotente
	igual, err := repo.Guardar(context.Background(), entities.GuardarEvidencia{
		ID:          id,
		ActividadID: labor,
		Nombre:      "foto.jpg",
		Mime:        "image/jpeg",
		Ext:         ".jpg",
		Bytes:       len(foto),
		Contenido:   foto,
		Hash:        hash,
		Lat:         &lat,
		Lon:         &lon,
		Exif:        exifStr,
		Rol:         "capataz",
		CapatazID:   "cap-norte",
		UsuarioID:   uid,
	}, store)
	if err != nil {
		t.Fatalf("reintento falló: %v", err)
	}
	if !igual.Idempotente {
		t.Fatalf("esperaba idempotente=true, obtenido %+v", igual)
	}
	if err := gdb.Raw(`SELECT count(*) FROM actividad_eventos WHERE actividad_id = $1 AND tipo = 'evidencia'`, labor).Scan(&n).Error; err != nil {
		t.Fatal(err)
	}
	if n != 1 {
		t.Fatalf("el reintento duplicó la bitácora: %d", n)
	}

	// 3. Capataz ajeno con mismo ID -> 403 (permiso antes de 409)
	sumOtra := sha256.Sum256(otra)
	hashOtra := hex.EncodeToString(sumOtra[:])
	_, err = repo.Guardar(context.Background(), entities.GuardarEvidencia{
		ID:          id,
		ActividadID: labor,
		Nombre:      "otra.jpg",
		Mime:        "image/jpeg",
		Ext:         ".jpg",
		Bytes:       len(otra),
		Contenido:   otra,
		Hash:        hashOtra,
		Rol:         "capataz",
		CapatazID:   "cap-sur",
		UsuarioID:   uid,
	}, store)
	if err != domainErrors.ErrOperacionProhibido {
		t.Fatalf("esperaba ErrOperacionProhibido, obtenido %v", err)
	}

	// 4. Mismo ID pero contenido distinto -> 409 conflicto
	_, err = repo.Guardar(context.Background(), entities.GuardarEvidencia{
		ID:          id,
		ActividadID: labor,
		Nombre:      "otra.jpg",
		Mime:        "image/jpeg",
		Ext:         ".jpg",
		Bytes:       len(otra),
		Contenido:   otra,
		Hash:        hashOtra,
		Rol:         "capataz",
		CapatazID:   "cap-norte",
		UsuarioID:   uid,
	}, store)
	if err != domainErrors.ErrLaborConflicto {
		t.Fatalf("esperaba ErrLaborConflicto, obtenido %v", err)
	}

	// 5. Capataz ajeno con nuevo ID -> 403
	_, err = repo.Guardar(context.Background(), entities.GuardarEvidencia{
		ID:          "bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbbb",
		ActividadID: labor,
		Nombre:      "foto.jpg",
		Mime:        "image/jpeg",
		Ext:         ".jpg",
		Bytes:       len(foto),
		Contenido:   foto,
		Hash:        hash,
		Rol:         "capataz",
		CapatazID:   "cap-sur",
		UsuarioID:   uid,
	}, store)
	if err != domainErrors.ErrOperacionProhibido {
		t.Fatalf("esperaba ErrOperacionProhibido, obtenido %v", err)
	}
}

func TestArchivoCabecerasCache(t *testing.T) {
	_, gdb := testutil.MigrarDBTemporal(t, "evid_cache")

	filesDir := t.TempDir()
	store := storage.NewDiscoStorage(filesDir)
	repo := NewEvidenciaRepository(gdb)

	labor := "11111111-1111-4111-8111-111111111111"
	id := "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa"
	foto := []byte{0xFF, 0xD8, 0xFF, 0xD9}
	var uid int64
	if err := gdb.Raw(`
		INSERT INTO usuarios (usuario, nombre, rol, capataz_id, password_hash)
		VALUES ('prueba-campo', 'Equipo de prueba', 'capataz', 'cap-norte', 'x')
		RETURNING id`).Row().Scan(&uid); err != nil {
		t.Fatal(err)
	}

	sum := sha256.Sum256(foto)
	hash := hex.EncodeToString(sum[:])

	_, err := repo.Guardar(context.Background(), entities.GuardarEvidencia{
		ID:          id,
		ActividadID: labor,
		Nombre:      "foto.jpg",
		Mime:        "image/jpeg",
		Ext:         ".jpg",
		Bytes:       len(foto),
		Contenido:   foto,
		Hash:        hash,
		Rol:         "capataz",
		CapatazID:   "cap-norte",
		UsuarioID:   uid,
	}, store)
	if err != nil {
		t.Fatalf("guardar falló: %v", err)
	}

	ruta, mime, err := repo.ObtenerRutaYMime(context.Background(), id)
	if err != nil {
		t.Fatalf("ObtenerRutaYMime falló: %v", err)
	}
	if mime != "image/jpeg" {
		t.Fatalf("mime esperado image/jpeg, obtenido %s", mime)
	}

	rc, err := store.Open(context.Background(), ruta)
	if err != nil {
		t.Fatalf("storage.Open falló: %v", err)
	}
	defer rc.Close()

	recuperado, err := io.ReadAll(rc)
	if err != nil {
		t.Fatalf("lectura falló: %v", err)
	}
	if string(recuperado) != string(foto) {
		t.Fatalf("contenido no coincide")
	}
}

func TestJefaturaPuedeSubirEvidenciaYCapatazLimitadoASuLabor(t *testing.T) {
	_, gdb := testutil.MigrarDBTemporal(t, "evid_perm")

	filesDir := t.TempDir()
	store := storage.NewDiscoStorage(filesDir)
	repo := NewEvidenciaRepository(gdb)

	laborNorte := "11111111-1111-4111-8111-111111111111" // asignada a cap-norte en semilla
	foto := []byte{0xFF, 0xD8, 0xFF, 0xD9}

	var uidJef, uidSur int64
	if err := gdb.Raw(`
		INSERT INTO usuarios (usuario, nombre, rol, password_hash)
		VALUES ('test-jefatura', 'Jefe Test', 'jefatura', 'x')
		RETURNING id`).Row().Scan(&uidJef); err != nil {
		t.Fatal(err)
	}
	if err := gdb.Raw(`
		INSERT INTO usuarios (usuario, nombre, rol, capataz_id, password_hash)
		VALUES ('test-sur', 'Equipo Sur', 'capataz', 'cap-sur', 'x')
		RETURNING id`).Row().Scan(&uidSur); err != nil {
		t.Fatal(err)
	}

	sum := sha256.Sum256(foto)
	hash := hex.EncodeToString(sum[:])

	// 1. Jefatura puede subir evidencia
	idJef := "bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbbb"
	resJef, err := repo.Guardar(context.Background(), entities.GuardarEvidencia{
		ID:          idJef,
		ActividadID: laborNorte,
		Nombre:      "foto.jpg",
		Mime:        "image/jpeg",
		Ext:         ".jpg",
		Bytes:       len(foto),
		Contenido:   foto,
		Hash:        hash,
		Rol:         "jefatura",
		UsuarioID:   uidJef,
	}, store)
	if err != nil {
		t.Fatalf("jefatura subida falló: %v", err)
	}
	if resJef.ID != idJef {
		t.Fatalf("id esperado %s, obtenido %s", idJef, resJef.ID)
	}

	// 2. Capataz ajeno (cap-sur en labor de cap-norte) recibe prohibido
	idSur := "cccccccc-cccc-4ccc-8ccc-cccccccccccc"
	_, err = repo.Guardar(context.Background(), entities.GuardarEvidencia{
		ID:          idSur,
		ActividadID: laborNorte,
		Nombre:      "foto.jpg",
		Mime:        "image/jpeg",
		Ext:         ".jpg",
		Bytes:       len(foto),
		Contenido:   foto,
		Hash:        hash,
		Rol:         "capataz",
		CapatazID:   "cap-sur",
		UsuarioID:   uidSur,
	}, store)
	if err != domainErrors.ErrOperacionProhibido {
		t.Fatalf("capataz ajeno esperado ErrOperacionProhibido, obtenido %v", err)
	}
}
