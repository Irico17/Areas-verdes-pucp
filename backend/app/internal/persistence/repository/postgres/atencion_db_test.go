package postgres

import (
	"context"
	"testing"

	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/dto"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/shared/testutil"
)

func TestAtencion_SolicitudRepository(t *testing.T) {
	_, gdb := testutil.MigrarDBTemporal(t, "solicitud_repo")
	repo := NewSolicitudRepository(gdb)
	ctx := context.Background()

	solID := "11111111-2222-4333-8444-555555555555"
	crearDTO := dto.CrearSolicitudDTO{
		ID:            solID,
		CodigoExterno: "EXT-001",
		Fuente:        "centuria",
		Titulo:        "Poda urgente",
		Detalle:       "Ramas bajas cerca del pabellón Z",
		Prioridad:     "alta",
		Lugar:         "Frente a Pabellón Z",
		Cantidad:      2,
	}

	// 1. Crear
	creada, err := repo.Crear(ctx, crearDTO)
	if err != nil {
		t.Fatalf("error al crear solicitud: %v", err)
	}
	if creada.ID != solID || creada.CodigoExterno == nil || *creada.CodigoExterno != "EXT-001" {
		t.Fatalf("solicitud creada inesperada: %+v", creada)
	}

	// 2. ObtenerPorID
	recuperada, err := repo.ObtenerPorID(ctx, solID)
	if err != nil {
		t.Fatalf("error al obtener solicitud por ID: %v", err)
	}
	if recuperada.Titulo != crearDTO.Titulo {
		t.Fatalf("titulo esperado %q, obtenido %q", crearDTO.Titulo, recuperada.Titulo)
	}

	// 3. Editar
	editarDTO := dto.EditarSolicitudDTO{
		ID:            solID,
		CodigoExterno: "EXT-001-MOD",
		Titulo:        "Poda urgente modificada",
		Detalle:       "Ramas muy bajas",
		Prioridad:     "media",
		Lugar:         "Pabellón Z",
	}
	actualizada, err := repo.Editar(ctx, editarDTO)
	if err != nil {
		t.Fatalf("error al editar solicitud: %v", err)
	}
	if actualizada.Titulo != "Poda urgente modificada" || actualizada.Prioridad != "media" {
		t.Fatalf("actualización incorrecta: %+v", actualizada)
	}

	// 4. Listar
	lista, err := repo.Listar(ctx)
	if err != nil {
		t.Fatalf("error al listar solicitudes: %v", err)
	}
	if len(lista) == 0 {
		t.Fatal("se esperaba al menos 1 solicitud en la lista")
	}
}

func TestAtencion_ServicioTercerizadoRepository(t *testing.T) {
	_, gdb := testutil.MigrarDBTemporal(t, "orden_repo")
	repo := NewServicioTercerizadoRepository(gdb)
	ctx := context.Background()

	actividadID := "aaaaaaaa-1111-4aaa-8aaa-aaaaaaaaaaa1"

	// Insertar actividad con ejecutor='tercerizada' asignada a cap-norte
	if err := gdb.Exec(`
		INSERT INTO actividades (id, tipo, estado, titulo, detalle, assigned_capataz_id, geom, ejecutor)
		VALUES (?, 'poda', 'pendiente', 'Poda mayor', 'Servicio externo', 'cap-norte', ST_SetSRID(ST_MakePoint(-77.08, -12.07), 4326), 'tercerizada')
	`, actividadID).Error; err != nil {
		t.Fatalf("error insertando actividad de prueba: %v", err)
	}

	ordenID := "bbbbbbbb-2222-4bbb-8bbb-bbbbbbbbbbb2"
	crearDTO := dto.CrearOrdenDTO{
		ID:          ordenID,
		ActividadID: actividadID,
		Empresa:     "Arboristas SAC",
		Referencia:  "OC-2026-999",
		Frecuencia:  "mensual",
	}

	// 1. Crear como capataz asignado
	creada, err := repo.Crear(ctx, crearDTO, "capataz", "cap-norte")
	if err != nil {
		t.Fatalf("error creando orden de servicio: %v", err)
	}
	if creada.ID != ordenID || creada.Empresa != "Arboristas SAC" {
		t.Fatalf("orden creada inesperada: %+v", creada)
	}

	// 2. ObtenerPorID
	recuperada, err := repo.ObtenerPorID(ctx, ordenID)
	if err != nil {
		t.Fatalf("error al obtener orden por ID: %v", err)
	}
	if recuperada.Empresa != "Arboristas SAC" {
		t.Fatalf("empresa esperada %q, obtenida %q", "Arboristas SAC", recuperada.Empresa)
	}

	// 3. Editar como capataz asignado
	editarDTO := dto.EditarOrdenDTO{
		ID:          ordenID,
		Estado:      "ejecutada",
		Conformidad: "Trabajo concluido según especificaciones",
	}
	actualizada, err := repo.Editar(ctx, editarDTO, "capataz", "cap-norte")
	if err != nil {
		t.Fatalf("error al actualizar orden: %v", err)
	}
	if actualizada.Estado != "ejecutada" || actualizada.Conformidad != "Trabajo concluido según especificaciones" {
		t.Fatalf("actualización de orden incorrecta: %+v", actualizada)
	}

	// 4. Intentar editar como capataz no asignado (debe fallar)
	_, errNoAsig := repo.Editar(ctx, editarDTO, "capataz", "cap-sur")
	if errNoAsig == nil {
		t.Fatal("se esperaba error al editar orden asignada a otro capataz")
	}

	// 5. Listar
	lista, err := repo.Listar(ctx)
	if err != nil {
		t.Fatalf("error al listar órdenes: %v", err)
	}
	if len(lista) == 0 {
		t.Fatal("se esperaba al menos 1 orden en la lista")
	}
}

func TestAtencion_RiegoRepository(t *testing.T) {
	_, gdb := testutil.MigrarDBTemporal(t, "riego_repo")
	repo := NewRiegoRepository(gdb)
	ctx := context.Background()

	// Insertar zona de supervisión Z1
	if err := gdb.Exec(`
		INSERT INTO zonas_supervision (codigo, nombre, geom)
		VALUES ('Z1', 'Zona 1', ST_Multi(ST_GeomFromText('POLYGON((-77.08 -12.07, -77.07 -12.07, -77.07 -12.06, -77.08 -12.06, -77.08 -12.07))', 4326)))
		ON CONFLICT (codigo) DO NOTHING
	`).Error; err != nil {
		t.Fatalf("error insertando zona de supervisión: %v", err)
	}

	riegoID := "cccccccc-3333-4ccc-8ccc-cccccccccccc"
	crearDTO := dto.CrearRiegoDTO{
		ID:        riegoID,
		ZonaID:    "Z1",
		Sector:    "Sector Jardines Centrales",
		Turno:     "manana",
		CapatazID: "cap-norte",
		Fecha:     "2026-10-01",
		Nota:      "Riego por aspersión concluido",
	}

	// Contar registros iniciales para cap-norte
	iniNorte, err := repo.Listar(ctx, "cap-norte")
	if err != nil {
		t.Fatalf("error listando riego inicial cap-norte: %v", err)
	}

	// 1. Crear
	if err := repo.Crear(ctx, crearDTO); err != nil {
		t.Fatalf("error al crear registro de riego: %v", err)
	}

	// 2. Listar sin filtro (para coordinación/admin)
	listaTodos, err := repo.Listar(ctx, "")
	if err != nil {
		t.Fatalf("error listando riego sin filtro: %v", err)
	}
	if len(listaTodos) == 0 {
		t.Fatal("se esperaba al menos 1 registro de riego")
	}

	// 3. Listar filtrado por capataz asignado
	listaNorte, err := repo.Listar(ctx, "cap-norte")
	if err != nil {
		t.Fatalf("error listando riego como cap-norte: %v", err)
	}
	if len(listaNorte) != len(iniNorte)+1 {
		t.Fatalf("se esperaba %d registros para cap-norte, obtenido %d", len(iniNorte)+1, len(listaNorte))
	}
	var found bool
	for _, it := range listaNorte {
		if it.ID == riegoID {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("el registro creado %s no se encontró en el listado de cap-norte", riegoID)
	}

	// 4. Listar filtrado por otro capataz (no debe incluir el nuevo riegoID)
	listaSur, err := repo.Listar(ctx, "cap-sur")
	if err != nil {
		t.Fatalf("error listando riego como cap-sur: %v", err)
	}
	for _, it := range listaSur {
		if it.ID == riegoID {
			t.Fatalf("capataz sur no debería ver el riego %s de cap-norte", riegoID)
		}
	}
}
