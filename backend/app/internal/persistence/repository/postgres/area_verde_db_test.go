package postgres_test

import (
	"context"
	"testing"
	"time"

	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/entities"
	domainErrors "github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/errors"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/persistence/repository/postgres"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/shared/testutil"
)

func TestAreaVerdeRepository_CRUD(t *testing.T) {
	_, gdb := testutil.MigrarDBTemporal(t, "area_verde_repo")
	ctx := context.Background()
	repo := postgres.NewAreaVerdeRepository(gdb)
	cambioRepo := postgres.NewCambioRepository(gdb)

	// 1. Crear sin geom
	creada, err := repo.CrearSinGeom(ctx, "AV-TEST-01", "Área de Prueba", "jardín")
	if err != nil {
		t.Fatalf("error creando área: %v", err)
	}
	if creada.FeatureID != "AV-TEST-01" || creada.Nombre != "Área de Prueba" || creada.Uso != "jardín" || creada.ConGeom {
		t.Fatalf("datos incorrectos en creada: %+v", creada)
	}

	// 2. Obtener por feature_id
	obtenida, err := repo.ObtenerFichaPorFeatureID(ctx, "AV-TEST-01")
	if err != nil {
		t.Fatalf("error obteniendo ficha: %v", err)
	}
	if obtenida.Nombre != "Área de Prueba" {
		t.Fatalf("nombre inesperado: %s", obtenida.Nombre)
	}

	// 3. Fichas (búsqueda)
	lista, err := repo.Fichas(ctx, "TEST")
	if err != nil {
		t.Fatalf("error listando fichas: %v", err)
	}
	if len(lista) != 1 || lista[0].FeatureID != "AV-TEST-01" {
		t.Fatalf("resultado de búsqueda inesperado: %+v", lista)
	}

	// 4. Actualizar ficha
	actualizada, err := repo.ActualizarFicha(ctx, "AV-TEST-01", "Área Modificada", "recreativo", "aspersión", "cerca al pabellón Z")
	if err != nil {
		t.Fatalf("error actualizando ficha: %v", err)
	}
	if actualizada.Nombre != "Área Modificada" || actualizada.Uso != "recreativo" || actualizada.RiegoAct != "aspersión" || actualizada.Referencia != "cerca al pabellón Z" {
		t.Fatalf("datos actualizados incorrectos: %+v", actualizada)
	}

	// 5. Actualizar ficha inexistente -> ErrFichaNoEncontrada
	_, err = repo.ActualizarFicha(ctx, "AV-NO-EXISTE", "Nombre", "", "", "")
	if err != domainErrors.ErrFichaNoEncontrada {
		t.Fatalf("se esperaba ErrFichaNoEncontrada, obtenido: %v", err)
	}

	// 6. Test CambioRepository
	antes := `{"nombre":"Área de Prueba"}`
	despues := `{"nombre":"Área Modificada"}`
	err = cambioRepo.Crear(ctx, &entities.Cambio{
		Entidad:   "areas_verdes",
		EntidadID: "AV-TEST-01",
		Accion:    "edicion",
		Antes:     &antes,
		Despues:   &despues,
		CreatedAt: time.Now(),
	})
	if err != nil {
		t.Fatalf("error creando registro de cambio: %v", err)
	}

	var totalCambios int64
	if err := gdb.Raw("SELECT count(*) FROM cambios WHERE entidad = 'areas_verdes' AND entidad_id = 'AV-TEST-01'").Scan(&totalCambios).Error; err != nil {
		t.Fatalf("error contando cambios: %v", err)
	}
	if totalCambios != 1 {
		t.Fatalf("se esperaba 1 cambio en la BD, obtenido: %d", totalCambios)
	}
}

func TestAreaVerdeRepository_Baja(t *testing.T) {
	_, gdb := testutil.MigrarDBTemporal(t, "area_verde_baja")
	ctx := context.Background()
	repo := postgres.NewAreaVerdeRepository(gdb)
	geoRepo := postgres.NewGeoRepository(gdb)

	polyGeom := `ST_SetSRID(ST_Multi(ST_GeomFromGeoJSON('{"type":"Polygon","coordinates":[[[-77.085,-12.072],[-77.080,-12.072],[-77.080,-12.068],[-77.085,-12.068],[-77.085,-12.072]]]}')), 4326)`
	if err := gdb.Exec(`
		INSERT INTO areas_verdes (feature_id, source_index, codigo, nombre, uso, geom, activo)
		VALUES ('AV-ACTIVA-01', 100, 'AV-ACTIVA-01', 'Área Activa', 'jardín', ` + polyGeom + `, true)`).Error; err != nil {
		t.Fatalf("insertar area activa: %v", err)
	}

	if err := gdb.Exec(`
		INSERT INTO areas_verdes (feature_id, source_index, codigo, nombre, uso, geom, activo)
		VALUES ('AV-BAJA-01', 101, 'AV-BAJA-01', 'Área a dar de baja', 'jardín', ` + polyGeom + `, true)`).Error; err != nil {
		t.Fatalf("insertar area baja: %v", err)
	}

	var uid int64
	if err := gdb.Raw(`
		INSERT INTO usuarios (usuario, nombre, rol, password_hash)
		VALUES ('baja.area', 'Usuario baja área', 'coordinacion', 'no-es-clave')
		RETURNING id`).Row().Scan(&uid); err != nil {
		t.Fatalf("insertar usuario: %v", err)
	}
	if err := repo.Baja(ctx, "AV-BAJA-01", &uid); err != nil {
		t.Fatalf("baja: %v", err)
	}

	var activo bool
	if err := gdb.Raw(`SELECT activo FROM areas_verdes WHERE feature_id = 'AV-BAJA-01'`).Scan(&activo).Error; err != nil {
		t.Fatalf("leer activo: %v", err)
	}
	if activo {
		t.Fatal("la fila debe seguir existiendo con activo=false")
	}

	var total int64
	if err := gdb.Raw(`SELECT count(*) FROM cambios WHERE entidad = 'areas_verdes' AND entidad_id = 'AV-BAJA-01' AND accion = 'baja'`).Scan(&total).Error; err != nil {
		t.Fatalf("contar cambios: %v", err)
	}
	if total != 1 {
		t.Fatalf("cambios de baja = %d, esperado 1", total)
	}

	if err := repo.Baja(ctx, "AV-BAJA-01", &uid); err != nil {
		t.Fatalf("segunda baja: %v", err)
	}
	if err := gdb.Raw(`SELECT count(*) FROM cambios WHERE entidad = 'areas_verdes' AND entidad_id = 'AV-BAJA-01' AND accion = 'baja'`).Scan(&total).Error; err != nil {
		t.Fatalf("recontar cambios: %v", err)
	}
	if total != 1 {
		t.Fatalf("la segunda baja no debe duplicar el historial, obtenido %d", total)
	}

	if err := repo.Baja(ctx, "AV-NO-EXISTE", nil); err != domainErrors.ErrFichaNoEncontrada {
		t.Fatalf("esperado ErrFichaNoEncontrada, obtenido %v", err)
	}

	// 1. Fichas: área activa aparece, área de baja no aparece
	listaTodas, err := repo.Fichas(ctx, "")
	if err != nil {
		t.Fatalf("fichas todas tras baja: %v", err)
	}
	if len(listaTodas) != 1 || listaTodas[0].FeatureID != "AV-ACTIVA-01" {
		t.Fatalf("esperada solo 1 area activa (AV-ACTIVA-01), obtenido: %+v", listaTodas)
	}
	listaBaja, err := repo.Fichas(ctx, "AV-BAJA-01")
	if err != nil {
		t.Fatalf("fichas tras baja: %v", err)
	}
	if len(listaBaja) != 0 {
		t.Fatalf("listado debe ocultar áreas inactivas, obtenido %+v", listaBaja)
	}

	// 2. geo/areas: área activa aparece, área de baja no aparece
	fc, err := geoRepo.Areas(ctx, entities.FiltroGeo{})
	if err != nil {
		t.Fatalf("geo areas: %v", err)
	}
	if len(fc.Features) != 1 || fc.Features[0].ID != "AV-ACTIVA-01" {
		t.Fatalf("geo/areas debe incluir solo activa, obtenido %d features: %+v", len(fc.Features), fc.Features)
	}

	// 3. Resumen: cuenta solo activas
	res, err := geoRepo.Resumen(ctx)
	if err != nil {
		t.Fatalf("geo resumen: %v", err)
	}
	if res.Areas != 1 || res.AreasConGeometria != 1 {
		t.Fatalf("resumen debe contar solo activas (1/1), obtenido areas=%d con_geom=%d", res.Areas, res.AreasConGeometria)
	}

	// 4. ObtenerFichaPorFeatureID: la ficha de la baja sí se obtiene y expone activo=false
	fichaBaja, err := repo.ObtenerFichaPorFeatureID(ctx, "AV-BAJA-01")
	if err != nil {
		t.Fatalf("ficha tras baja: %v", err)
	}
	if fichaBaja.Activo {
		t.Fatal("la ficha de baja debe exponer activo=false")
	}

	fichaActiva, err := repo.ObtenerFichaPorFeatureID(ctx, "AV-ACTIVA-01")
	if err != nil {
		t.Fatalf("ficha activa: %v", err)
	}
	if !fichaActiva.Activo {
		t.Fatal("la ficha activa debe exponer activo=true")
	}
}
