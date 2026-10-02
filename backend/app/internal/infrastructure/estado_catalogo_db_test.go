package infrastructure_test

import (
	"context"
	"errors"
	"testing"

	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/dto"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/usecases"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/entities"
	domainErrors "github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/errors"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/persistence/repository/postgres"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/shared/testutil"
)

func TestSetEstadoExigeCatalogoYTransicion(t *testing.T) {
	_, gdb := testutil.MigrarDBTemporal(t, "est_cat")
	ctx := context.Background()
	uc := usecases.NewIntervencionUseCase(
		postgres.NewIntervencionRepository(gdb),
		postgres.NewCatalogoRepository(gdb),
	)
	porIniciar := "11111111-1111-4111-8111-111111111111"
	enProceso := "22222222-2222-4222-8222-222222222222"

	var ejecutadoID int64
	if err := gdb.Raw(`SELECT id FROM catalogos WHERE clase = 'estado' AND codigo = 'ejecutado'`).Scan(&ejecutadoID).Error; err != nil {
		t.Fatal(err)
	}
	if ejecutadoID == 0 {
		t.Fatal("la migración no cargó ejecutado")
	}
	if err := postgres.NewCatalogoRepository(gdb).Deactivate(ctx, ejecutadoID, 0); err != nil {
		t.Fatal(err)
	}

	_, err := uc.CambiarEstado(ctx, dto.CambiarEstadoDTO{
		ID: enProceso, Estado: "ejecutado", ActorRol: "coordinacion",
	})
	var input domainErrors.InputError
	if !errors.As(err, &input) || input.Reason != "el estado no está activo en el catálogo" {
		t.Fatalf("sin ítem activo se esperaba rechazo de catálogo, obtuve %v", err)
	}

	if _, err := postgres.NewCatalogoRepository(gdb).Create(ctx, "estado", "ejecutado", "Ejecutado"); err != nil {
		t.Fatal(err)
	}

	_, err = uc.CambiarEstado(ctx, dto.CambiarEstadoDTO{
		ID: porIniciar, Estado: "cerrada", ActorRol: "coordinacion",
	})
	if !errors.As(err, &input) || input.Reason != "esa transición de estado no está permitida" {
		t.Fatalf("cerrar desde por iniciar debía fallar, obtuve %v", err)
	}

	feat, err := uc.CambiarEstado(ctx, dto.CambiarEstadoDTO{
		ID: enProceso, Estado: "ejecutado", ActorRol: "coordinacion",
	})
	if err != nil {
		t.Fatal(err)
	}
	props, ok := feat.Properties.(entities.ActividadProperties)
	if !ok || props.Estado != "ejecutado" || props.EstadoEtiqueta != "Ejecutado" {
		t.Fatalf("respuesta inesperada: %+v", feat.Properties)
	}
}
