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
