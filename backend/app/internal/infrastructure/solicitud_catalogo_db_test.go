package infrastructure_test

import (
	"context"
	"errors"
	"testing"

	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/dto"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/usecases"
	domainErrors "github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/errors"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/persistence/repository/postgres"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/shared/testutil"
)

func TestSolicitudCantidadDistintaYFuenteFueraDeCatalogo(t *testing.T) {
	_, gdb := testutil.MigrarDBTemporal(t, "sol_cat")
	uc := usecases.NewSolicitudUseCase(postgres.NewSolicitudRepository(gdb), postgres.NewCatalogoRepository(gdb))
	ctx := context.Background()

	pedida := 4
	ejecutada := 2
	lat := -12.0696
	lon := -77.0796
	creada, err := uc.Crear(ctx, dto.CrearSolicitudDTO{
		ID:                 "11111111-1111-4111-8111-111111111111",
		Fuente:             "osg",
		Titulo:             "Poda de demostración",
		Prioridad:          "media",
		CantidadSolicitada: &pedida,
		CantidadEjecutada:  &ejecutada,
		Lat:                &lat,
		Lon:                &lon,
	})
	if err != nil {
		t.Fatalf("alta con cantidades distintas: %v", err)
	}
	if creada.CantidadSolicitada == nil || *creada.CantidadSolicitada != 4 {
		t.Fatalf("cantidad solicitada = %v", creada.CantidadSolicitada)
	}
	if creada.CantidadEjecutada == nil || *creada.CantidadEjecutada != 2 {
		t.Fatalf("cantidad ejecutada = %v", creada.CantidadEjecutada)
	}
	if creada.Cantidad == nil || *creada.Cantidad != 4 {
		t.Fatalf("cantidad histórica = %v; no debe perderse", creada.Cantidad)
	}
	if creada.Lat == nil || creada.Lon == nil {
		t.Fatal("el punto dentro del campus no se guardó")
	}

	fueraLat := -12.20
	_, err = uc.Crear(ctx, dto.CrearSolicitudDTO{
		ID:     "22222222-2222-4222-8222-222222222222",
		Fuente: "osg",
		Titulo: "Punto fuera",
		Lat:    &fueraLat,
		Lon:    &lon,
	})
	var fuera domainErrors.InputError
	if !errors.As(err, &fuera) || fuera.Reason != "el punto queda fuera del campus" {
		t.Fatalf("punto fuera: %v", err)
	}

	_, err = uc.Crear(ctx, dto.CrearSolicitudDTO{
		ID:     "33333333-3333-4333-8333-333333333333",
		Fuente: "whatsapp",
		Titulo: "Fuente ajena",
		Lat:    &lat,
		Lon:    &lon,
	})
	var input domainErrors.InputError
	if !errors.As(err, &input) {
		t.Fatalf("fuente fuera de catálogo debía ser 400, fue %v", err)
	}
}
