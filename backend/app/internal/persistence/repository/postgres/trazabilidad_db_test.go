package postgres

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"strings"
	"testing"

	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/entities"
	domainErrors "github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/errors"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/infrastructure/storage"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/shared/testutil"
)

func TestTrazabilidadReasignaAvanceYEvidencia(t *testing.T) {
	_, gdb := testutil.MigrarDBTemporal(t, "traza")
	ctx := context.Background()
	repo := NewIntervencionRepository(gdb)
	laborID := "12121212-1111-4111-8111-121212121212"

	var uid int64
	if err := gdb.Raw(`
		INSERT INTO usuarios (usuario, nombre, rol, password_hash)
		VALUES ('coord-traza', 'Inés Calderón', 'coordinacion', 'hash')
		RETURNING id`).Row().Scan(&uid); err != nil {
		t.Fatal(err)
	}

	if _, _, err := repo.Create(ctx, entities.NuevaIntervencion{
		ID:                laborID,
		Tipo:              "riego",
		Titulo:            "Riego del eje de prueba",
		Detalle:           "Aspersores",
		Lon:               -77.08,
		Lat:               -12.07,
		AssignedCapatazID: "cap-norte",
		ActorRol:          "coordinacion",
		Ejecutor:          "propia",
		UsuarioID:         uid,
	}); err != nil {
		t.Fatal(err)
	}

	if _, err := repo.Assign(ctx, entities.AsignarIntervencion{
		ID:        laborID,
		CapatazID: "cap-sur",
		ActorRol:  "coordinacion",
		UsuarioID: uid,
	}); err != nil {
		t.Fatal(err)
	}

	var anterior, nuevo string
	if err := gdb.Raw(`
		SELECT capataz_anterior, capataz_id
		FROM actividad_eventos
		WHERE actividad_id = $1 AND tipo = 'reasignada'`, laborID).Row().Scan(&anterior, &nuevo); err != nil {
		t.Fatal(err)
	}
	if anterior != "cap-norte" || nuevo != "cap-sur" {
		t.Fatalf("reasignación = anterior %s nuevo %s", anterior, nuevo)
	}

	hitoID, err := repo.RegistrarHito(ctx, entities.NuevoHito{
		ActividadID: laborID,
		Tipo:        "inicio",
		Texto:       "Se abre el turno de riego",
		ActorRol:    "coordinacion",
		UsuarioID:   uid,
	})
	if err != nil || hitoID == 0 {
		t.Fatalf("hito inicio: id=%d err=%v", hitoID, err)
	}

	err = gdb.Exec(`
		INSERT INTO actividad_eventos (actividad_id, tipo, actor_rol, nota, usuario_id)
		VALUES ($1, 'inventado', 'coordinacion', 'no debe entrar', $2)`, laborID, uid).Error
	if err == nil || !strings.Contains(err.Error(), "actividad_eventos_tipo_chk") {
		t.Fatalf("el CHECK debía rechazar el tipo inventado, obtuvo %v", err)
	}

	if err := repo.CrearAvance(ctx, entities.NuevoAvance{
		ActividadID:   laborID,
		ID:            "13131313-1111-4111-8111-131313131313",
		Fecha:         "2026-10-02",
		Nota:          "Primer tramo regado",
		AreaFeatureID: "AV-0001",
		ActorRol:      "coordinacion",
		UsuarioID:     uid,
	}); err != nil {
		t.Fatal(err)
	}

	var avances int
	if err := gdb.Raw(`
		SELECT count(*) FROM actividad_eventos
		WHERE actividad_id = $1 AND tipo = 'avance' AND nota = 'Primer tramo regado' AND usuario_id = $2`,
		laborID, uid).Scan(&avances).Error; err != nil {
		t.Fatal(err)
	}
	if avances != 1 {
		t.Fatalf("eventos de avance = %d", avances)
	}

	var eventoReasig int64
	if err := gdb.Raw(`
		SELECT id FROM actividad_eventos
		WHERE actividad_id = $1 AND tipo = 'reasignada'`, laborID).Row().Scan(&eventoReasig); err != nil {
		t.Fatal(err)
	}

	foto := []byte{0xFF, 0xD8, 0xFF, 0xD9}
	sum := sha256.Sum256(foto)
	store := storage.NewDiscoStorage(t.TempDir())
	evidencias := NewEvidenciaRepository(gdb)
	guardada, err := evidencias.Guardar(ctx, entities.GuardarEvidencia{
		ID:          "14141414-1111-4111-8111-141414141414",
		ActividadID: laborID,
		SolicitudID: "77777777-7777-4777-8777-777777777777",
		Nombre:      "aspersor.jpg",
		Mime:        "image/jpeg",
		Ext:         ".jpg",
		Bytes:       len(foto),
		Contenido:   foto,
		Nota:        "Foto del tramo",
		Hash:        hex.EncodeToString(sum[:]),
		Rol:         "coordinacion",
		UsuarioID:   uid,
		EventoID:    eventoReasig,
	}, store)
	if err != nil {
		t.Fatal(err)
	}
	if guardada.Idempotente {
		t.Fatal("el alta no es un reintento")
	}

	var solicitud string
	var eventoGuardado int64
	if err := gdb.Raw(`
		SELECT COALESCE(solicitud_id::text, ''), COALESCE(evento_id, 0)
		FROM evidencias WHERE id = $1`, guardada.ID).Row().Scan(&solicitud, &eventoGuardado); err != nil {
		t.Fatal(err)
	}
	if solicitud != "77777777-7777-4777-8777-777777777777" || eventoGuardado != eventoReasig {
		t.Fatalf("evidencia solicitud=%s evento=%d", solicitud, eventoGuardado)
	}

	eventos, err := repo.Timeline(ctx, laborID)
	if err != nil {
		t.Fatal(err)
	}
	var visto bool
	for _, ev := range eventos {
		if ev.ID != eventoReasig {
			continue
		}
		visto = true
		if ev.CapatazAnterior == nil || *ev.CapatazAnterior != "cap-norte" {
			t.Fatalf("timeline sin cuadrilla anterior: %+v", ev)
		}
		if ev.CuadrillaAnterior == nil || *ev.CuadrillaAnterior == "" {
			t.Fatalf("timeline sin nombre anterior: %+v", ev)
		}
		if ev.Equipo == nil || *ev.Equipo == "" {
			t.Fatalf("timeline sin cuadrilla nueva: %+v", ev)
		}
		if len(ev.Evidencias) != 1 || ev.Evidencias[0].Nombre != "aspersor.jpg" {
			t.Fatalf("la evidencia no cuelga del evento: %+v", ev.Evidencias)
		}
		if ev.UsuarioID == nil || *ev.UsuarioID != uid {
			t.Fatalf("evento sin usuario: %+v", ev)
		}
	}
	if !visto {
		t.Fatal("el timeline no trajo la reasignación")
	}

	_, err = evidencias.Guardar(ctx, entities.GuardarEvidencia{
		ID:          "15151515-1111-4111-8111-151515151515",
		ActividadID: laborID,
		Nombre:      "ajena.jpg",
		Mime:        "image/jpeg",
		Ext:         ".jpg",
		Bytes:       len(foto),
		Contenido:   foto,
		Hash:        hex.EncodeToString(sum[:]),
		Rol:         "coordinacion",
		UsuarioID:   uid,
		EventoID:    999999,
	}, store)
	if _, ok := err.(domainErrors.InputError); !ok {
		t.Fatalf("evento ajeno: %v", err)
	}
}
