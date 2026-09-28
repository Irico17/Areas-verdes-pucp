package postgres_test

import (
	"context"
	"testing"

	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/entities"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/persistence/repository/postgres"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/shared/testutil"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func TestInventarioCampoRepository_Database(t *testing.T) {
	rawDB, gdb := testutil.MigrarDBTemporal(t, "inv_campo")
	defer rawDB.Close()

	ctx := context.Background()
	repo := postgres.NewInventarioCampoRepository(gdb)

	// 1. TACHOS
	lat := -12.071
	lon := -77.081
	tachoInput := entities.Tacho{
		Codigo:          "PT-001",
		Lat:             &lat,
		Lon:             &lon,
		Nota:            "Tacho cerca a biblioteca",
		Lugar:           "Biblioteca Central",
		NoAprovechables: 10,
		PapelCarton:     5,
		Plastico:        3,
	}
	savedTacho, err := repo.GuardarTacho(ctx, tachoInput)
	if err != nil {
		t.Fatalf("GuardarTacho fallo: %v", err)
	}
	if savedTacho.ID == 0 || !savedTacho.Activo {
		t.Fatalf("ID o activo incorrecto en savedTacho: %+v", savedTacho)
	}

	tachos, err := repo.ListarTachos(ctx)
	if err != nil {
		t.Fatalf("ListarTachos fallo: %v", err)
	}
	if len(tachos) != 1 || tachos[0].Codigo != "PT-001" {
		t.Fatalf("ListarTachos inesperado: %+v", tachos)
	}

	// ActualizarTacho parcial con ClavesPatch
	patchDoc := []byte(`{"nota":"Nota modificada","no_aprovechables":25}`)
	updatedTacho, err := repo.ActualizarTacho(ctx, savedTacho.ID, entities.Tacho{
		Nota:            "Nota modificada",
		NoAprovechables: 25,
	}, patchDoc)
	if err != nil {
		t.Fatalf("ActualizarTacho fallo: %v", err)
	}
	if updatedTacho.ID != savedTacho.ID {
		t.Fatalf("ID modificado en ActualizarTacho: %d", updatedTacho.ID)
	}

	// 2. BEBEDEROS
	bebederoInput := entities.Bebedero{
		Codigo:  "PT_BEB_01",
		Subtipo: "fuente",
		Estado:  "operativo",
		Sede:    "Pando",
		Lat:     &lat,
		Lon:     &lon,
	}
	savedBeb, err := repo.GuardarBebedero(ctx, bebederoInput)
	if err != nil {
		t.Fatalf("GuardarBebedero fallo: %v", err)
	}
	if savedBeb.ID < 1 {
		t.Fatalf("ID inválido: %d", savedBeb.ID)
	}
	bebederos, err := repo.ListarBebederos(ctx)
	if err != nil {
		t.Fatalf("ListarBebederos fallo: %v", err)
	}
	if len(bebederos) != 1 || bebederos[0].Codigo != "PT_BEB_01" {
		t.Fatalf("ListarBebederos inesperado: %+v", bebederos)
	}

	// 3. PUNTOS PUCP
	puntoInput := entities.PuntoPUCP{
		Titulo: "Comedor Central",
		Lat:    lat,
		Lon:    lon,
		URL:    "https://example.com/pucp",
	}
	savedPunto, err := repo.GuardarPunto(ctx, puntoInput)
	if err != nil {
		t.Fatalf("GuardarPunto fallo: %v", err)
	}
	if savedPunto.ID < 1 {
		t.Fatalf("ID inválido: %d", savedPunto.ID)
	}
	puntos, err := repo.ListarPuntos(ctx, "Comedor")
	if err != nil {
		t.Fatalf("ListarPuntos fallo: %v", err)
	}
	if len(puntos) != 1 || puntos[0].Titulo != "Comedor Central" {
		t.Fatalf("ListarPuntos con filtro 'Comedor' fallo: %+v", puntos)
	}

	// 4. RESERVAS JARDIN
	reservaInput := entities.ReservaJardin{
		Fecha:      "2026-10-15",
		HoraInicio: "10:00",
		HoraFin:    "12:00",
		Estado:     "reservado",
		Evento:     "Feria Botanica",
		Unidad:     "DAES",
	}
	savedReserva, err := repo.GuardarReserva(ctx, reservaInput)
	if err != nil {
		t.Fatalf("GuardarReserva fallo: %v", err)
	}
	if savedReserva.ID < 1 {
		t.Fatalf("ID inválido: %d", savedReserva.ID)
	}
	reservas, err := repo.ListarReservas(ctx, "2026-10-01", "2026-10-31")
	if err != nil {
		t.Fatalf("ListarReservas fallo: %v", err)
	}
	if len(reservas) != 1 || reservas[0].Evento != "Feria Botanica" {
		t.Fatalf("ListarReservas inesperado: %+v", reservas)
	}

	// 5. FICHAS CAPA (fauna, puertas, etc.)
	area := 150.5
	perimetro := 50.2
	fichaInput := entities.FichaCapa{
		FeatureID:  "XER-001",
		Clase:      "xerofitica",
		Riego:      "goteo",
		AreaM2:     &area,
		PerimetroM: &perimetro,
	}
	savedFicha, err := repo.GuardarFicha(ctx, "xerofiticas", fichaInput)
	if err != nil {
		t.Fatalf("GuardarFicha xerofiticas fallo: %v", err)
	}
	if savedFicha.ID < 1 {
		t.Fatalf("ID inválido: %d", savedFicha.ID)
	}
	fichas, err := repo.ListarFichas(ctx, "xerofiticas")
	if err != nil {
		t.Fatalf("ListarFichas xerofiticas fallo: %v", err)
	}
	if len(fichas) != 1 || fichas[0].FeatureID != "XER-001" {
		t.Fatalf("ListarFichas inesperado: %+v", fichas)
	}

	// 6. BAJA LOGICA (SOFT DELETE)
	// Verificar conteo total de filas antes de la baja
	var countBefore int64
	gdb.Raw("SELECT count(*) FROM tachos").Scan(&countBefore)
	if countBefore != 1 {
		t.Fatalf("esperado 1 fila en tachos, obtenido %d", countBefore)
	}

	if err := repo.Baja(ctx, "tachos", savedTacho.ID); err != nil {
		t.Fatalf("Baja tacho fallo: %v", err)
	}

	// Tras la baja, ListarTachos no debe devolver el tacho inactivo
	tachosActivos, err := repo.ListarTachos(ctx)
	if err != nil {
		t.Fatalf("ListarTachos tras baja fallo: %v", err)
	}
	if len(tachosActivos) != 0 {
		t.Fatalf("se esperaban 0 tachos activos tras baja, obtenidos %d", len(tachosActivos))
	}

	// El conteo total de filas NO debe bajar (baja lógica)
	var countAfter int64
	gdb.Raw("SELECT count(*) FROM tachos").Scan(&countAfter)
	if countAfter != countBefore {
		t.Fatalf("el conteo total bajo tras DELETE: antes %d, despues %d", countBefore, countAfter)
	}

	// Segunda baja al mismo id debe dar error (no encontrado / ya inactivo)
	if err := repo.Baja(ctx, "tachos", savedTacho.ID); err == nil {
		t.Fatalf("segunda baja debia fallar con registro no encontrado")
	}
}
