package usecases

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"strings"

	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/contracts"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/dto"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/entities"
	domainErrors "github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/errors"
)

// dummyHash is a valid bcrypt hash used to prevent timing attacks when credentials or users are invalid.
const dummyHash = "$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy"

type sesionUseCase struct {
	sesionRepo  contracts.ISesionRepository
	usuarioRepo contracts.IUsuarioRepository
	hasher      contracts.IHasher
}

// NewSesionUseCase creates a new session usecase implementation.
func NewSesionUseCase(
	sesionRepo contracts.ISesionRepository,
	usuarioRepo contracts.IUsuarioRepository,
	hasher contracts.IHasher,
) contracts.ISesionUseCase {
	return &sesionUseCase{
		sesionRepo:  sesionRepo,
		usuarioRepo: usuarioRepo,
		hasher:      hasher,
	}
}

func (uc *sesionUseCase) Login(ctx context.Context, usuario, clave string) (string, *dto.UsuarioSesionDTO, error) {
	usuario = strings.TrimSpace(usuario)
	if usuario == "" || clave == "" {
		_ = uc.hasher.Compare(dummyHash, "dummy-password")
		return "", nil, domainErrors.ErrCredencialesInvalidas
	}

	u, err := uc.usuarioRepo.ObtenerPorUsuario(ctx, usuario)
	if err != nil || u == nil || !u.Activo || !u.RolActivo {
		_ = uc.hasher.Compare(dummyHash, clave)
		return "", nil, domainErrors.ErrCredencialesInvalidas
	}

	if err := uc.hasher.Compare(u.PasswordHash, clave); err != nil {
		return "", nil, domainErrors.ErrCredencialesInvalidas
	}

	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return "", nil, err
	}
	token := hex.EncodeToString(raw)
	sum := sha256.Sum256([]byte(token))
	tokenHash := hex.EncodeToString(sum[:])

	sesion := &entities.Sesion{
		TokenHash: tokenHash,
		UsuarioID: u.ID,
	}
	if err := uc.sesionRepo.Crear(ctx, sesion); err != nil {
		return "", nil, err
	}

	return token, sesionDesdeUsuario(u), nil
}

func (uc *sesionUseCase) Resolver(ctx context.Context, token string) (*dto.UsuarioSesionDTO, error) {
	token = strings.TrimSpace(token)
	if token == "" {
		return nil, domainErrors.ErrSinSesion
	}

	sum := sha256.Sum256([]byte(token))
	tokenHash := hex.EncodeToString(sum[:])

	u, err := uc.sesionRepo.ObtenerPorTokenHash(ctx, tokenHash)
	if err != nil || u == nil {
		return nil, domainErrors.ErrSinSesion
	}

	return sesionDesdeUsuario(u), nil
}

func sesionDesdeUsuario(u *entities.Usuario) *dto.UsuarioSesionDTO {
	return &dto.UsuarioSesionDTO{
		ID:                  u.ID,
		Usuario:             u.Usuario,
		Nombre:              u.Nombre,
		Rol:                 u.Rol,
		RolNombre:           u.RolNombre,
		CapatazID:           u.CapatazID,
		DebeCambiarPassword: u.DebeCambiarPassword,
	}
}

func (uc *sesionUseCase) Logout(ctx context.Context, token string) {
	token = strings.TrimSpace(token)
	if token == "" {
		return
	}
	sum := sha256.Sum256([]byte(token))
	tokenHash := hex.EncodeToString(sum[:])
	_ = uc.sesionRepo.EliminarPorTokenHash(ctx, tokenHash)
}
