package usecases_test

import (
	"context"
	"testing"

	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/dto"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/usecases"
	domainEntities "github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/entities"
	domainErrors "github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/errors"
)

type mockGeoRepo struct {
	areas   domainEntities.FeatureCollection
	zonas   domainEntities.FeatureCollection
	resumen domainEntities.ResumenCatastro
	capas   domainEntities.CapasIndex
}

func (m *mockGeoRepo) Areas(_ context.Context, _ domainEntities.FiltroGeo) (domainEntities.FeatureCollection, error) {
	return m.areas, nil
}

func (m *mockGeoRepo) Zonas(_ context.Context, _ domainEntities.FiltroGeo) (domainEntities.FeatureCollection, error) {
	return m.zonas, nil
}

func (m *mockGeoRepo) Capa(_ context.Context, capa string, _ domainEntities.FiltroGeo) (domainEntities.FeatureCollection, error) {
	if capa != "xerofitica" && capa != "jardines_reserva" {
		return domainEntities.FeatureCollection{}, domainErrors.ErrCapaDesconocida
	}
	return domainEntities.Collection(capa), nil
}

func (m *mockGeoRepo) Capas(context.Context) (domainEntities.CapasIndex, error) {
	return m.capas, nil
}

func (m *mockGeoRepo) Resumen(context.Context) (domainEntities.ResumenCatastro, error) {
	return m.resumen, nil
}

type mockArchivoEstaticoAdapter struct {
	edificios    []byte
	edificiosErr error
}

func (m *mockArchivoEstaticoAdapter) LeerEdificios(context.Context) ([]byte, error) {
	if m.edificiosErr != nil {
		return nil, m.edificiosErr
	}
	return m.edificios, nil
}

func TestGeoUseCase_CapaDesconocida(t *testing.T) {
	repo := &mockGeoRepo{}
	fileAdapter := &mockArchivoEstaticoAdapter{}
	uc := usecases.NewGeoUseCase(repo, fileAdapter)

	_, err := uc.Capa(context.Background(), "no-existe", dto.FiltroGeoDTO{})
	if err != domainErrors.ErrCapaDesconocida {
		t.Fatalf("se esperaba ErrCapaDesconocida, obtenido: %v", err)
	}

	fc, err := uc.Capa(context.Background(), "xerofitica", dto.FiltroGeoDTO{})
	if err != nil {
		t.Fatalf("error inesperado en capa xerofitica: %v", err)
	}
	if fc.Name != "xerofitica" {
		t.Fatalf("nombre esperado 'xerofitica', obtenido: %s", fc.Name)
	}
}

func TestGeoUseCase_EdificiosRetornaColeccionVaciaSiError(t *testing.T) {
	repo := &mockGeoRepo{}
	fileAdapter := &mockArchivoEstaticoAdapter{edificiosErr: domainErrors.ErrEntrada}
	uc := usecases.NewGeoUseCase(repo, fileAdapter)

	bytes, err := uc.Edificios(context.Background())
	if err != nil {
		t.Fatalf("no debía fallar, debía retornar colección vacía: %v", err)
	}
	if string(bytes) != `{"type":"FeatureCollection","name":"edificios","features":[]}` {
		t.Fatalf("contenido inesperado: %s", string(bytes))
	}
}
