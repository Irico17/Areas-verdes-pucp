package postgres_test

import (
	"context"
	"encoding/json"
	"errors"
	"strconv"
	"testing"

	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/entities"
	apperrors "github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/errors"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/persistence/repository/postgres"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/shared/testutil"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func TestZonificacionRepositorio_SectorLugarViaCuartel(t *testing.T) {
	_, gdb := testutil.MigrarDBTemporal(t, "zonif")
	ctx := context.Background()
	repo := postgres.NewZonificacionRepository(gdb)

	item, err := repo.CrearSector(ctx, entities.NuevoSectorCapataz{
		Codigo: "sector-lago", Nombre: "Sector lago (ficticio)", Color: "#3f73b0", UsuarioID: 0,
	})
	if err != nil {
		t.Fatalf("alta: %v", err)
	}
	if item.ID == 0 || !item.Activo {
		t.Fatalf("sector %+v", item)
	}
	_, err = repo.CrearSector(ctx, entities.NuevoSectorCapataz{
		Codigo: "sector-lago", Nombre: "Otro", Color: "#3f73b0",
	})
	if !errors.Is(err, apperrors.ErrSectorDuplicado) {
		t.Fatalf("duplicado: %v", err)
	}
	var filas int64
	if err := gdb.Raw(`SELECT count(*) FROM sectores_capataz WHERE codigo = 'sector-lago'`).Scan(&filas).Error; err != nil {
		t.Fatal(err)
	}
	if filas != 1 {
		t.Fatalf("filas del código %d", filas)
	}
	var cambios int64
	if err := gdb.Raw(`SELECT count(*) FROM cambios WHERE entidad = 'sectores_capataz' AND entidad_id = ? AND accion = 'alta'`, strconv.FormatInt(item.ID, 10)).Scan(&cambios).Error; err != nil {
		t.Fatal(err)
	}
	if cambios != 1 {
		t.Fatalf("cambios de alta %d", cambios)
	}

	imp, err := repo.ImportarSectores(ctx, []entities.NuevoSectorCapataz{{
		Codigo: "sector-lago", Nombre: "Sector lago corregido", Color: "#6b5596",
	}}, 0)
	if err != nil {
		t.Fatal(err)
	}
	if imp.Creados != 0 || imp.Actualizados != 1 {
		t.Fatalf("importación %+v", imp)
	}
	if err := gdb.Raw(`SELECT count(*) FROM sectores_capataz WHERE codigo = 'sector-lago'`).Scan(&filas).Error; err != nil {
		t.Fatal(err)
	}
	if filas != 1 {
		t.Fatalf("la importación duplicó el sector: %d", filas)
	}

	antes, err := repo.ContarLugares(ctx)
	if err != nil {
		t.Fatal(err)
	}
	_, ok, err := repo.LugarPorNorm(ctx, "sitio que no existe")
	if err != nil || ok {
		t.Fatalf("lugar desconocido ok=%v err=%v", ok, err)
	}
	despues, err := repo.ContarLugares(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if antes != despues {
		t.Fatalf("se insertó un lugar: %d -> %d", antes, despues)
	}

	vias, err := repo.ViasGeoJSON(ctx)
	if err != nil {
		t.Fatal(err)
	}
	var capa struct {
		Features []json.RawMessage `json:"features"`
	}
	if err := json.Unmarshal(vias, &capa); err != nil {
		t.Fatal(err)
	}
	if len(capa.Features) != 0 {
		t.Fatalf("vías inventadas: %s", vias)
	}
	linea := `{"type":"LineString","coordinates":[[-77.0800,-12.0700],[-77.0790,-12.0695]]}`
	escrito, err := repo.ImportarVias(ctx, []entities.ViaAlta{{
		FeatureID: "via-prueba", Nombre: "Sendero de prueba", GeoJSON: linea,
	}}, 0)
	if err != nil {
		t.Fatal(err)
	}
	if escrito.Creadas != 1 {
		t.Fatalf("vías %+v", escrito)
	}
	otra, err := repo.ImportarVias(ctx, []entities.ViaAlta{{
		FeatureID: "via-prueba", Nombre: "Sendero de prueba", GeoJSON: linea,
	}}, 0)
	if err != nil {
		t.Fatal(err)
	}
	if otra.Creadas != 0 || otra.Actualizadas != 1 {
		t.Fatalf("segunda importación %+v", otra)
	}
	if err := gdb.Raw(`SELECT count(*) FROM vias WHERE feature_id = 'via-prueba'`).Scan(&filas).Error; err != nil {
		t.Fatal(err)
	}
	if filas != 1 {
		t.Fatalf("vías duplicadas %d", filas)
	}

	cuarteles, err := repo.CuartelesGeoJSON(ctx)
	if err != nil {
		t.Fatal(err)
	}
	var histo struct {
		Aviso    string            `json:"aviso"`
		Features []json.RawMessage `json:"features"`
	}
	if err := json.Unmarshal(cuarteles, &histo); err != nil {
		t.Fatal(err)
	}
	if histo.Aviso != "sin archivo de cuarteles" || len(histo.Features) != 0 {
		t.Fatalf("cuarteles %s", cuarteles)
	}
	var nCuarteles int64
	if err := gdb.Raw(`SELECT count(*) FROM cuarteles_historico`).Scan(&nCuarteles).Error; err != nil {
		t.Fatal(err)
	}
	if nCuarteles != 0 {
		t.Fatalf("se inventaron cuarteles: %d", nCuarteles)
	}
}
