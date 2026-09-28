package usecases_test

import (
	"context"
	"errors"
	"testing"

	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/usecases"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/entities"
	domainErrors "github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/errors"
)

type fakeUsuarioRepo struct {
	usuarios map[string]*entities.Usuario
}

func (f *fakeUsuarioRepo) ObtenerPorUsuario(_ context.Context, u string) (*entities.Usuario, error) {
	usr, ok := f.usuarios[u]
	if !ok {
		return nil, errors.New("not found")
	}
	return usr, nil
}

func (f *fakeUsuarioRepo) Listar(_ context.Context) ([]entities.Usuario, error) {
	out := make([]entities.Usuario, 0, len(f.usuarios))
	for _, u := range f.usuarios {
		out = append(out, *u)
	}
	return out, nil
}

func (f *fakeUsuarioRepo) ExistePorUsuario(_ context.Context, u string) (bool, error) {
	_, ok := f.usuarios[u]
	return ok, nil
}

func (f *fakeUsuarioRepo) Crear(_ context.Context, u *entities.Usuario) error {
	f.usuarios[u.Usuario] = u
	return nil
}

type fakeSesionRepo struct {
	sesiones      map[string]*entities.Sesion
	usuariosByTok map[string]*entities.Usuario
	deletedTokens []string
}

func (f *fakeSesionRepo) Crear(_ context.Context, s *entities.Sesion) error {
	f.sesiones[s.TokenHash] = s
	return nil
}

func (f *fakeSesionRepo) ObtenerPorTokenHash(_ context.Context, tokenHash string) (*entities.Usuario, error) {
	u, ok := f.usuariosByTok[tokenHash]
	if !ok {
		return nil, errors.New("sesion no encontrada")
	}
	return u, nil
}

func (f *fakeSesionRepo) EliminarPorTokenHash(_ context.Context, tokenHash string) error {
	f.deletedTokens = append(f.deletedTokens, tokenHash)
	delete(f.sesiones, tokenHash)
	delete(f.usuariosByTok, tokenHash)
	return nil
}

type fakeHasher struct {
	compareErr error
	compared   []string
}

func (f *fakeHasher) Hash(password string) (string, error) {
	return "hash_" + password, nil
}

func (f *fakeHasher) Compare(hash, password string) error {
	f.compared = append(f.compared, hash+":"+password)
	if f.compareErr != nil {
		return f.compareErr
	}
	if hash != "hash_"+password {
		return errors.New("mismatch")
	}
	return nil
}

func TestSesionUseCase_Login(t *testing.T) {
	uRepo := &fakeUsuarioRepo{
		usuarios: map[string]*entities.Usuario{
			"admin": {
				ID:           1,
				Usuario:      "admin",
				Nombre:       "Administrador",
				Rol:          "admin",
				RolNombre:    "Administrador",
				PasswordHash: "hash_pando-local",
				Activo:       true,
			},
			"inactivo": {
				ID:           2,
				Usuario:      "inactivo",
				Nombre:       "Inactivo",
				Rol:          "capataz",
				PasswordHash: "hash_pando-local",
				Activo:       false,
			},
		},
	}
	sRepo := &fakeSesionRepo{
		sesiones:      make(map[string]*entities.Sesion),
		usuariosByTok: make(map[string]*entities.Usuario),
	}
	hasher := &fakeHasher{}

	uc := usecases.NewSesionUseCase(sRepo, uRepo, hasher)
	ctx := context.Background()

	// 1. Clave vacía o usuario vacío corre comparación contra dummyHash y retorna error
	_, _, err := uc.Login(ctx, "", "clave")
	if !errors.Is(err, domainErrors.ErrCredencialesInvalidas) {
		t.Fatalf("esperado ErrCredencialesInvalidas, obtenido %v", err)
	}
	if len(hasher.compared) == 0 {
		t.Fatal("login con usuario vacío debió ejecutar compare contra dummy hash")
	}

	hasher.compared = nil
	_, _, err = uc.Login(ctx, "admin", "")
	if !errors.Is(err, domainErrors.ErrCredencialesInvalidas) {
		t.Fatalf("esperado ErrCredencialesInvalidas, obtenido %v", err)
	}
	if len(hasher.compared) == 0 {
		t.Fatal("login con clave vacía debió ejecutar compare contra dummy hash")
	}

	// 2. Usuario inexistente
	hasher.compared = nil
	_, _, err = uc.Login(ctx, "no-existe", "clave")
	if !errors.Is(err, domainErrors.ErrCredencialesInvalidas) {
		t.Fatalf("esperado ErrCredencialesInvalidas, obtenido %v", err)
	}
	if len(hasher.compared) == 0 {
		t.Fatal("login con usuario inexistente debió ejecutar compare contra dummy hash")
	}

	// 3. Usuario inactivo
	hasher.compared = nil
	_, _, err = uc.Login(ctx, "inactivo", "pando-local")
	if !errors.Is(err, domainErrors.ErrCredencialesInvalidas) {
		t.Fatalf("esperado ErrCredencialesInvalidas, obtenido %v", err)
	}
	if len(hasher.compared) == 0 {
		t.Fatal("login con usuario inactivo debió ejecutar compare contra dummy hash")
	}

	// 4. Clave incorrecta
	hasher.compared = nil
	_, _, err = uc.Login(ctx, "admin", "clave-erronea")
	if !errors.Is(err, domainErrors.ErrCredencialesInvalidas) {
		t.Fatalf("esperado ErrCredencialesInvalidas, obtenido %v", err)
	}

	// 5. Login exitoso
	hasher.compared = nil
	token, dto, err := uc.Login(ctx, "admin", "pando-local")
	if err != nil {
		t.Fatalf("login exitoso falló: %v", err)
	}
	if token == "" {
		t.Fatal("se esperaba token no vacío")
	}
	if dto == nil || dto.Usuario != "admin" || dto.Rol != "admin" {
		t.Fatalf("dto de usuario inesperado: %+v", dto)
	}
	if len(sRepo.sesiones) != 1 {
		t.Fatalf("se esperaba 1 sesión creada, hay %d", len(sRepo.sesiones))
	}
}

func TestSesionUseCase_Resolver(t *testing.T) {
	sRepo := &fakeSesionRepo{
		sesiones:      make(map[string]*entities.Sesion),
		usuariosByTok: make(map[string]*entities.Usuario),
	}
	uRepo := &fakeUsuarioRepo{}
	hasher := &fakeHasher{}

	uc := usecases.NewSesionUseCase(sRepo, uRepo, hasher)
	ctx := context.Background()

	// 1. Token vacío
	_, err := uc.Resolver(ctx, "")
	if !errors.Is(err, domainErrors.ErrSinSesion) {
		t.Fatalf("token vacío: esperado ErrSinSesion, obtenido %v", err)
	}

	// 2. Token desconocido
	_, err = uc.Resolver(ctx, "token-desconocido")
	if !errors.Is(err, domainErrors.ErrSinSesion) {
		t.Fatalf("token desconocido: esperado ErrSinSesion, obtenido %v", err)
	}
}

func TestSesionUseCase_Logout(t *testing.T) {
	sRepo := &fakeSesionRepo{
		sesiones:      make(map[string]*entities.Sesion),
		usuariosByTok: make(map[string]*entities.Usuario),
	}
	uRepo := &fakeUsuarioRepo{}
	hasher := &fakeHasher{}

	uc := usecases.NewSesionUseCase(sRepo, uRepo, hasher)
	ctx := context.Background()

	// 1. Logout con token vacío
	uc.Logout(ctx, "")
	if len(sRepo.deletedTokens) != 0 {
		t.Fatal("logout vacío no debe eliminar tokens")
	}

	// 2. Logout con token válido
	uc.Logout(ctx, "token-123")
	if len(sRepo.deletedTokens) != 1 {
		t.Fatalf("se esperaba 1 token eliminado, hay %d", len(sRepo.deletedTokens))
	}
}
