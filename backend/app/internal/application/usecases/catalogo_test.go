package usecases_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/dto"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/usecases"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/constants/enums"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/entities"
	apperrors "github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/errors"
)

type mockCatalogoRepository struct {
	listFn       func(ctx context.Context, clase string, soloActivos bool) ([]entities.CatalogoItem, error)
	activoFn     func(ctx context.Context, clase, codigo string) (bool, error)
	createFn     func(ctx context.Context, clase, codigo, nombre string) (*entities.CatalogoItem, error)
	deactivateFn func(ctx context.Context, id int64, usuarioID int64) error
	renombrarFn  func(ctx context.Context, id int64, nombre string, usuarioID int64) (*entities.CatalogoItem, error)
}

func (m *mockCatalogoRepository) List(ctx context.Context, clase string, soloActivos bool) ([]entities.CatalogoItem, error) {
	if m.listFn != nil {
		return m.listFn(ctx, clase, soloActivos)
	}
	return []entities.CatalogoItem{}, nil
}

func (m *mockCatalogoRepository) Activo(ctx context.Context, clase, codigo string) (bool, error) {
	if m.activoFn != nil {
		return m.activoFn(ctx, clase, codigo)
	}
	return true, nil
}

func (m *mockCatalogoRepository) Create(ctx context.Context, clase, codigo, nombre string) (*entities.CatalogoItem, error) {
	if m.createFn != nil {
		return m.createFn(ctx, clase, codigo, nombre)
	}
	return &entities.CatalogoItem{
		ID:     1,
		Clase:  clase,
		Codigo: codigo,
		Nombre: nombre,
		Activo: true,
		Orden:  0,
	}, nil
}

func (m *mockCatalogoRepository) Deactivate(ctx context.Context, id int64, usuarioID int64) error {
	if m.deactivateFn != nil {
		return m.deactivateFn(ctx, id, usuarioID)
	}
	return nil
}

func (m *mockCatalogoRepository) Renombrar(ctx context.Context, id int64, nombre string, usuarioID int64) (*entities.CatalogoItem, error) {
	if m.renombrarFn != nil {
		return m.renombrarFn(ctx, id, nombre, usuarioID)
	}
	return &entities.CatalogoItem{ID: id, Nombre: nombre, Activo: true}, nil
}

func TestCatalogoUseCase_Crear_Validaciones(t *testing.T) {
	uc := usecases.NewCatalogoUseCase(&mockCatalogoRepository{})

	tests := []struct {
		name    string
		in      dto.CrearCatalogoDTO
		wantErr error
	}{
		{
			name: "clase no reconocida",
			in: dto.CrearCatalogoDTO{
				Clase:  "invalida",
				Codigo: "codigo_valido",
				Nombre: "Nombre Valido",
			},
			wantErr: apperrors.ErrClaseNoReconocida,
		},
		{
			name: "clase vacia",
			in: dto.CrearCatalogoDTO{
				Clase:  " ",
				Codigo: "codigo_valido",
				Nombre: "Nombre Valido",
			},
			wantErr: apperrors.ErrClaseNoReconocida,
		},
		{
			name: "codigo con mayusculas",
			in: dto.CrearCatalogoDTO{
				Clase:  "estado",
				Codigo: "CodigoInvalido",
				Nombre: "Nombre Valido",
			},
			wantErr: apperrors.ErrCodigoInvalido,
		},
		{
			name: "codigo demasiado corto",
			in: dto.CrearCatalogoDTO{
				Clase:  "estado",
				Codigo: "a",
				Nombre: "Nombre Valido",
			},
			wantErr: apperrors.ErrCodigoInvalido,
		},
		{
			name: "codigo demasiado largo (mas de 32 chars)",
			in: dto.CrearCatalogoDTO{
				Clase:  "estado",
				Codigo: strings.Repeat("a", 33),
				Nombre: "Nombre Valido",
			},
			wantErr: apperrors.ErrCodigoInvalido,
		},
		{
			name: "codigo con caracteres invalidos",
			in: dto.CrearCatalogoDTO{
				Clase:  "estado",
				Codigo: "codigo-invalido!",
				Nombre: "Nombre Valido",
			},
			wantErr: apperrors.ErrCodigoInvalido,
		},
		{
			name: "nombre vacio",
			in: dto.CrearCatalogoDTO{
				Clase:  "estado",
				Codigo: "codigo_valido",
				Nombre: "   ",
			},
			wantErr: apperrors.ErrNombreObligatorio,
		},
		{
			name: "nombre mas de 80 caracteres",
			in: dto.CrearCatalogoDTO{
				Clase:  "estado",
				Codigo: "codigo_valido",
				Nombre: strings.Repeat("ñ", 81),
			},
			wantErr: apperrors.ErrNombreObligatorio,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := uc.Crear(context.Background(), tc.in)
			if !errors.Is(err, tc.wantErr) {
				t.Fatalf("se esperaba error %v, se obtuvo %v", tc.wantErr, err)
			}
		})
	}
}

func TestCatalogoUseCase_Crear_Exito(t *testing.T) {
	mockRepo := &mockCatalogoRepository{
		createFn: func(ctx context.Context, clase, codigo, nombre string) (*entities.CatalogoItem, error) {
			return &entities.CatalogoItem{
				ID:     42,
				Clase:  clase,
				Codigo: codigo,
				Nombre: nombre,
				Activo: true,
				Orden:  1,
			}, nil
		},
	}

	uc := usecases.NewCatalogoUseCase(mockRepo)

	res, err := uc.Crear(context.Background(), dto.CrearCatalogoDTO{
		Clase:  "  tipo_actividad  ",
		Codigo: "  nuevo_tipo  ",
		Nombre: "  Nuevo Tipo De Actividad  ",
	})
	if err != nil {
		t.Fatalf("error inesperado: %v", err)
	}
	if res.ID != 42 || res.Clase != "tipo_actividad" || res.Codigo != "nuevo_tipo" || res.Nombre != "Nuevo Tipo De Actividad" {
		t.Fatalf("item creado inesperado: %+v", res)
	}
}

func TestCatalogoUseCase_Listar(t *testing.T) {
	mockRepo := &mockCatalogoRepository{
		listFn: func(ctx context.Context, clase string, soloActivos bool) ([]entities.CatalogoItem, error) {
			return []entities.CatalogoItem{
				{ID: 1, Clase: "estado", Codigo: "activo", Nombre: "Activo", Activo: true, Orden: 1},
				{ID: 2, Clase: "estado", Codigo: "inactivo", Nombre: "Inactivo", Activo: false, Orden: 2},
			}, nil
		},
	}

	uc := usecases.NewCatalogoUseCase(mockRepo)

	res, err := uc.Listar(context.Background(), dto.FiltroCatalogoDTO{Clase: "estado", SoloActivos: false})
	if err != nil {
		t.Fatalf("error inesperado: %v", err)
	}

	if len(res.Items) != 2 {
		t.Fatalf("se esperaban 2 items, se obtuvieron %d", len(res.Items))
	}
	if len(res.Clases) != len(enums.ClasesCatalogoValidas()) {
		t.Fatalf("se esperaban %d clases, se obtuvieron %d", len(enums.ClasesCatalogoValidas()), len(res.Clases))
	}
}

func TestCatalogoUseCase_Renombrar(t *testing.T) {
	mockRepo := &mockCatalogoRepository{
		renombrarFn: func(ctx context.Context, id int64, nombre string, usuarioID int64) (*entities.CatalogoItem, error) {
			if id != 3 || nombre != "Por iniciar" || usuarioID != 9 {
				t.Fatalf("argumentos inesperados: %d %q %d", id, nombre, usuarioID)
			}
			return &entities.CatalogoItem{ID: id, Clase: "estado", Codigo: "pendiente", Nombre: nombre, Activo: true}, nil
		},
	}
	uc := usecases.NewCatalogoUseCase(mockRepo)
	res, err := uc.Renombrar(context.Background(), dto.RenombrarCatalogoDTO{ID: 3, Nombre: "  Por iniciar  ", UsuarioID: 9})
	if err != nil {
		t.Fatal(err)
	}
	if res.Nombre != "Por iniciar" || res.ID != 3 {
		t.Fatalf("respuesta inesperada: %+v", res)
	}

	_, err = uc.Renombrar(context.Background(), dto.RenombrarCatalogoDTO{ID: 3, Nombre: "   "})
	if !errors.Is(err, apperrors.ErrNombreObligatorio) {
		t.Fatalf("se esperaba nombre obligatorio, se obtuvo %v", err)
	}
}

func TestCatalogoUseCase_Desactivar(t *testing.T) {
	t.Run("exito", func(t *testing.T) {
		mockRepo := &mockCatalogoRepository{
			deactivateFn: func(ctx context.Context, id int64, usuarioID int64) error {
				if id != 10 {
					t.Fatalf("id esperado 10, obtenido %d", id)
				}
				if usuarioID != 4 {
					t.Fatalf("usuario esperado 4, obtenido %d", usuarioID)
				}
				return nil
			},
		}
		uc := usecases.NewCatalogoUseCase(mockRepo)
		res, err := uc.Desactivar(context.Background(), 10, 4)
		if err != nil {
			t.Fatalf("error inesperado: %v", err)
		}
		if res.Activo != false || res.ID != 10 {
			t.Fatalf("respuesta inesperada: %+v", res)
		}
	})

	t.Run("no existe", func(t *testing.T) {
		mockRepo := &mockCatalogoRepository{
			deactivateFn: func(ctx context.Context, id int64, usuarioID int64) error {
				return apperrors.ErrItemNoExiste
			},
		}
		uc := usecases.NewCatalogoUseCase(mockRepo)
		_, err := uc.Desactivar(context.Background(), 999, 0)
		if !errors.Is(err, apperrors.ErrItemNoExiste) {
			t.Fatalf("se esperaba ErrItemNoExiste, se obtuvo %v", err)
		}
	})
}
