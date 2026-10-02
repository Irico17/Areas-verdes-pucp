package postgres_test

import (
	"context"
	"testing"

	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/entities"
	domainErrors "github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/errors"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/persistence/repository/postgres"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/shared/testutil"
)

func TestCatastroRepositories_DB(t *testing.T) {
	_, gdb := testutil.MigrarDBTemporal(t, "catastro_repo")
	ctx := context.Background()

	// 1. Zona de supervisión
	zonaRepo := postgres.NewZonaSupervisionRepository(gdb)
	area := 5000.0
	geoJSONPoly := `{"type":"MultiPolygon","coordinates":[[[[-77.085,-12.072],[-77.080,-12.072],[-77.080,-12.068],[-77.085,-12.068],[-77.085,-12.072]]]]}`
	nuevaZona, err := zonaRepo.Crear(ctx, "Z1", "Zona 1 Test", geoJSONPoly, &area)
	if err != nil {
		t.Fatalf("error creando zona: %v", err)
	}
	if nuevaZona.ID == 0 || nuevaZona.Codigo != "Z1" {
		t.Fatalf("zona creada inesperada: %+v", nuevaZona)
	}
	zonas, err := zonaRepo.Listar(ctx)
	if err != nil {
		t.Fatalf("error listando zonas: %v", err)
	}
	if len(zonas) != 1 {
		t.Fatalf("se esperaba 1 zona, obtenidas: %d", len(zonas))
	}

	// 2. Cuadrillas
	cuadRepo := postgres.NewCuadrillaRepository(gdb)
	nuevaCuad, err := cuadRepo.Crear(ctx, "cua-test", "Cuadrilla Test", "manana")
	if err != nil || nuevaCuad.ID != "cua-test" {
		t.Fatalf("error creando cuadrilla: %v", err)
	}
	cuads, err := cuadRepo.Listar(ctx)
	if err != nil {
		t.Fatalf("error listando cuadrillas: %v", err)
	}
	if len(cuads) == 0 {
		t.Fatal("se esperaban cuadrillas cargadas")
	}

	// 3. Lugares
	lugarRepo := postgres.NewLugarRepository(gdb)
	nuevoLugar, err := lugarRepo.Crear(ctx, "Jardín de Prueba", -12.07, -77.08, &nuevaZona.ID)
	if err != nil || nuevoLugar.ID == 0 {
		t.Fatalf("error creando lugar: %v", err)
	}
	lugares, err := lugarRepo.Listar(ctx)
	if err != nil {
		t.Fatalf("error listando lugares: %v", err)
	}
	if len(lugares) == 0 {
		t.Fatal("se esperaba al menos 1 lugar")
	}

	// 4. Especies
	espRepo := postgres.NewEspecieRepository(gdb)
	nuevaEsp, err := espRepo.Crear(ctx, "Testus botanicus", "Planta Test")
	if err != nil || nuevaEsp.ID == 0 {
		t.Fatalf("error creando especie: %v", err)
	}
	especies, err := espRepo.Listar(ctx)
	if err != nil {
		t.Fatalf("error listando especies: %v", err)
	}
	if len(especies) == 0 {
		t.Fatal("se esperaba al menos 1 especie")
	}

	// 5. Ejemplares y Recodificar (Port de TestRecodificarConCodigoPrevioVacio)
	ejRepo := postgres.NewEjemplarRepository(gdb)
	ej, err := ejRepo.Crear(ctx, entities.Ejemplar{Cantidad: 1, NombreComun: "Tipuana"})
	if err != nil {
		t.Fatalf("error creando ejemplar: %v", err)
	}
	hist, err := ejRepo.Recodificar(ctx, ej.ID, "AV-100", nil)
	if err != nil {
		t.Fatalf("error recodificando: %v", err)
	}
	if hist.CodigoAnterior != "" || hist.CodigoNuevo != "AV-100" || hist.ID == 0 {
		t.Fatalf("historial = %+v", hist)
	}
	var codigo string
	if err := gdb.Raw(`SELECT codigo FROM ejemplares WHERE id = ?`, ej.ID).Scan(&codigo).Error; err != nil {
		t.Fatalf("error verificando codigo: %v", err)
	}
	if codigo != "AV-100" {
		t.Fatalf("codigo = %s, esperado AV-100", codigo)
	}

	// Listar ejemplares
	ejemplares, total, err := ejRepo.Listar(ctx, 10, 0, "")
	if err != nil || total == 0 || len(ejemplares) == 0 {
		t.Fatalf("error listando ejemplares: %v (total=%d)", err, total)
	}

	// Listar codigos
	cods, err := ejRepo.ListarCodigos(ctx, ej.ID)
	if err != nil || len(cods) == 0 {
		t.Fatalf("error listando codigos: %v", err)
	}

	// 6. Polígonos y Capas de referencia
	refRepo := postgres.NewCatastroReferenciaRepository(gdb)
	poligonos, err := refRepo.ListarPoligonos(ctx)
	if err != nil {
		t.Fatalf("error listando poligonos: %v", err)
	}
	_ = poligonos

	for _, capa := range []string{"fauna", "puertas", "playas_estacionamiento", "veredas_riesgo", "xerofiticas", "jardines_reserva"} {
		filas, err := refRepo.ListarCapa(ctx, capa)
		if err != nil {
			t.Fatalf("error listando capa %s: %v", capa, err)
		}
		_ = filas
	}
}

func TestZonaSupervisionRepository_ActualizarYBaja(t *testing.T) {
	_, gdb := testutil.MigrarDBTemporal(t, "zona_edicion_baja")
	ctx := context.Background()
	zonaRepo := postgres.NewZonaSupervisionRepository(gdb)
	area := 5000.0
	geoJSONPoly := `{"type":"MultiPolygon","coordinates":[[[[-77.085,-12.072],[-77.080,-12.072],[-77.080,-12.068],[-77.085,-12.068],[-77.085,-12.072]]]]}`
	if _, err := zonaRepo.Crear(ctx, "Z2", "Zona original", geoJSONPoly, &area); err != nil {
		t.Fatalf("crear: %v", err)
	}

	nuevaArea := 6100.0
	var uid int64
	if err := gdb.Raw(`
		INSERT INTO usuarios (usuario, nombre, rol, password_hash)
		VALUES ('baja.zona', 'Usuario baja zona', 'coordinacion', 'no-es-clave')
		RETURNING id`).Row().Scan(&uid); err != nil {
		t.Fatalf("insertar usuario: %v", err)
	}
	editada, err := zonaRepo.Actualizar(ctx, "Z2", "Zona editada", "", &nuevaArea, &uid)
	if err != nil {
		t.Fatalf("actualizar: %v", err)
	}
	if editada.Nombre != "Zona editada" || editada.AreaM2 == nil || *editada.AreaM2 != nuevaArea || !editada.Activo {
		t.Fatalf("zona editada inesperada: %+v", editada)
	}

	var ediciones int64
	if err := gdb.Raw(`SELECT count(*) FROM cambios WHERE entidad = 'zonas_supervision' AND entidad_id = 'Z2' AND accion = 'edicion'`).Scan(&ediciones).Error; err != nil {
		t.Fatalf("contar ediciones: %v", err)
	}
	if ediciones != 1 {
		t.Fatalf("ediciones = %d, esperado 1", ediciones)
	}

	baja, err := zonaRepo.Baja(ctx, "Z2", &uid)
	if err != nil {
		t.Fatalf("baja: %v", err)
	}
	if baja.Activo {
		t.Fatal("la zona debe quedar inactiva y la fila debe seguir existiendo")
	}
	var activas int64
	if err := gdb.Raw(`SELECT count(*) FROM zonas_supervision WHERE codigo = 'Z2'`).Scan(&activas).Error; err != nil {
		t.Fatalf("contar filas: %v", err)
	}
	if activas != 1 {
		t.Fatalf("la fila debe permanecer, count=%d", activas)
	}

	if _, err := zonaRepo.Baja(ctx, "Z2", &uid); err != nil {
		t.Fatalf("segunda baja: %v", err)
	}
	var bajas int64
	if err := gdb.Raw(`SELECT count(*) FROM cambios WHERE entidad = 'zonas_supervision' AND entidad_id = 'Z2' AND accion = 'baja'`).Scan(&bajas).Error; err != nil {
		t.Fatalf("contar bajas: %v", err)
	}
	if bajas != 1 {
		t.Fatalf("la segunda baja no debe duplicar el historial, obtenido %d", bajas)
	}

	if _, err := zonaRepo.Actualizar(ctx, "Z2", "No debe", "", nil, nil); err != domainErrors.ErrNoEncontrado {
		t.Fatalf("editar una zona inactiva debe dar ErrNoEncontrado, obtenido %v", err)
	}
	if _, err := zonaRepo.Baja(ctx, "Z9", nil); err != domainErrors.ErrNoEncontrado {
		t.Fatalf("baja inexistente: %v", err)
	}
}
