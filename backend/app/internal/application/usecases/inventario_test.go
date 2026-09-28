package usecases_test

import (
	"context"
	"errors"
	"testing"

	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/usecases"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/entities"
	domainErrors "github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/errors"
)

type mockInventarioRepo struct {
	indexFunc func(ctx context.Context) (entities.IndiceInventario, error)
	capaFunc  func(ctx context.Context, capa string) (entities.FeatureCollection, error)
}

func (m *mockInventarioRepo) Index(ctx context.Context) (entities.IndiceInventario, error) {
	if m.indexFunc != nil {
		return m.indexFunc(ctx)
	}
	return entities.IndiceInventario{}, nil
}

func (m *mockInventarioRepo) Capa(ctx context.Context, capa string) (entities.FeatureCollection, error) {
	if m.capaFunc != nil {
		return m.capaFunc(ctx, capa)
	}
	return entities.Collection(capa), nil
}

type mockFotoDiscoAdapter struct {
	rutaFunc func(name string) (string, error)
}

func (m *mockFotoDiscoAdapter) RutaFoto(name string) (string, error) {
	if m.rutaFunc != nil {
		return m.rutaFunc(name)
	}
	return "/path/to/" + name, nil
}

func TestInventarioUseCase_Index(t *testing.T) {
	repo := &mockInventarioRepo{
		indexFunc: func(_ context.Context) (entities.IndiceInventario, error) {
			return entities.IndiceInventario{
				Capas: entities.CapasConocidas,
				Cargadas: []entities.CapaResumen{
					{Capa: "bebederos", Features: 10},
				},
			}, nil
		},
	}
	uc := usecases.NewInventarioUseCase(repo, &mockFotoDiscoAdapter{})

	idx, err := uc.Index(context.Background())
	if err != nil {
		t.Fatalf("error inesperado: %v", err)
	}
	if len(idx.Capas) != 11 || len(idx.Cargadas) != 1 {
		t.Fatalf("resultado inesperado en Index: %+v", idx)
	}
}

func TestInventarioUseCase_Capa(t *testing.T) {
	repo := &mockInventarioRepo{
		capaFunc: func(_ context.Context, capa string) (entities.FeatureCollection, error) {
			if capa == "desconocida" {
				return entities.Collection(capa), domainErrors.ErrCapaInventarioDesconocida
			}
			fc := entities.Collection(capa)
			fc.Features = append(fc.Features, entities.Feature{
				Type: "Feature",
				ID:   "BB-1",
			})
			return fc, nil
		},
	}
	uc := usecases.NewInventarioUseCase(repo, &mockFotoDiscoAdapter{})

	fc, err := uc.Capa(context.Background(), "bebederos")
	if err != nil {
		t.Fatalf("error inesperado: %v", err)
	}
	if len(fc.Features) != 1 || fc.Features[0].ID != "BB-1" {
		t.Fatalf("features inesperadas: %+v", fc)
	}

	_, errDesc := uc.Capa(context.Background(), "desconocida")
	if !errors.Is(errDesc, domainErrors.ErrCapaInventarioDesconocida) {
		t.Fatalf("se esperaba ErrCapaInventarioDesconocida, obtenido %v", errDesc)
	}
}

func TestInventarioUseCase_Foto(t *testing.T) {
	fotoAdapter := &mockFotoDiscoAdapter{
		rutaFunc: func(name string) (string, error) {
			if name == "invalido.png" {
				return "", domainErrors.ErrFotografiaNoDisponible
			}
			return "/data/raw/drive_fotos/" + name, nil
		},
	}
	uc := usecases.NewInventarioUseCase(&mockInventarioRepo{}, fotoAdapter)

	path, err := uc.Foto(context.Background(), "foto.jpg")
	if err != nil {
		t.Fatalf("error inesperado: %v", err)
	}
	if path != "/data/raw/drive_fotos/foto.jpg" {
		t.Fatalf("ruta inesperada: %s", path)
	}

	_, errInv := uc.Foto(context.Background(), "invalido.png")
	if !errors.Is(errInv, domainErrors.ErrFotografiaNoDisponible) {
		t.Fatalf("se esperaba ErrFotografiaNoDisponible, obtenido %v", errInv)
	}
}
