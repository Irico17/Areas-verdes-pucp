package usecases

import (
	"context"
	"errors"
	"regexp"
	"strings"

	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/contracts"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/dto"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/entities"
	domainErrors "github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/errors"
)

const avisoCuentas = "La jefatura de sección administra las cuentas."

const largoMinimoClave = 10

var (
	reUsuario = regexp.MustCompile(`^[a-z][a-z0-9._-]{2,39}$`)
	reCodigo  = regexp.MustCompile(`^[a-z][a-z0-9_]{1,31}$`)
	reAccion  = regexp.MustCompile(`^[a-z][a-z0-9_]{1,40}$`)
)

type usuarioUseCase struct {
	usuarioRepo contracts.IUsuarioRepository
	permisoRepo contracts.IPermisoRepository
	hasher      contracts.IHasher
	auditoria   contracts.IAuditoriaService
	tx          contracts.ITransaccion
}

// NewUsuarioUseCase creates a new user usecase implementation.
func NewUsuarioUseCase(
	usuarioRepo contracts.IUsuarioRepository,
	permisoRepo contracts.IPermisoRepository,
	hasher contracts.IHasher,
	auditoria contracts.IAuditoriaService,
	tx contracts.ITransaccion,
) contracts.IUsuarioUseCase {
	return &usuarioUseCase{
		usuarioRepo: usuarioRepo,
		permisoRepo: permisoRepo,
		hasher:      hasher,
		auditoria:   auditoria,
		tx:          tx,
	}
}

func (uc *usuarioUseCase) ListarUsuarios(ctx context.Context) (*dto.UsuariosResponseDTO, error) {
	usuarios, err := uc.usuarioRepo.Listar(ctx)
	if err != nil {
		return nil, err
	}
	permisos, err := uc.permisoRepo.Listar(ctx)
	if err != nil {
		return nil, err
	}
	roles, err := uc.permisoRepo.ListarRoles(ctx)
	if err != nil {
		return nil, err
	}

	usuariosDTO := make([]dto.CuentaDTO, 0, len(usuarios))
	for i := range usuarios {
		usuariosDTO = append(usuariosDTO, cuentaDe(&usuarios[i]))
	}
	permisosDTO := make([]dto.PermisoDTO, 0, len(permisos))
	for _, p := range permisos {
		permisosDTO = append(permisosDTO, dto.PermisoDTO{Rol: p.Rol, Accion: p.Accion})
	}
	rolesDTO := make([]dto.RolDTO, 0, len(roles))
	for _, rol := range roles {
		rolesDTO = append(rolesDTO, dto.RolDTO{Codigo: rol.Codigo, Nombre: rol.Nombre, Activo: rol.Activo})
	}
	return &dto.UsuariosResponseDTO{
		Usuarios: usuariosDTO,
		Permisos: permisosDTO,
		Roles:    rolesDTO,
		Aviso:    avisoCuentas,
	}, nil
}

func (uc *usuarioUseCase) Crear(ctx context.Context, actorID int64, in dto.CrearCuentaDTO) (*dto.CuentaDTO, error) {
	usuario := strings.TrimSpace(strings.ToLower(in.Usuario))
	nombre := strings.TrimSpace(in.Nombre)
	rol := strings.TrimSpace(in.Rol)
	if !reUsuario.MatchString(usuario) {
		return nil, domainErrors.ErrUsuarioInvalido
	}
	if nombre == "" || len(nombre) > 80 {
		return nil, domainErrors.ErrNombreInvalido
	}
	if len(in.Clave) < largoMinimoClave {
		return nil, domainErrors.ErrClaveInvalida
	}
	activo, err := uc.permisoRepo.RolActivo(ctx, rol)
	if err != nil {
		return nil, err
	}
	if !activo {
		return nil, domainErrors.ErrRolNoDisponible
	}
	existe, err := uc.usuarioRepo.ExistePorUsuario(ctx, usuario)
	if err != nil {
		return nil, err
	}
	if existe {
		return nil, domainErrors.ErrUsuarioYaExiste
	}
	hash, err := uc.hasher.Hash(in.Clave)
	if err != nil {
		return nil, err
	}
	cuenta := &entities.Usuario{
		Usuario:             usuario,
		Nombre:              nombre,
		Rol:                 rol,
		CapatazID:           strings.TrimSpace(in.CapatazID),
		PasswordHash:        hash,
		Activo:              true,
		DebeCambiarPassword: true,
		RolActivo:           true,
	}
	uid := actorID
	err = uc.tx.Ejecutar(ctx, func(txCtx context.Context) error {
		if err := uc.usuarioRepo.Crear(txCtx, cuenta); err != nil {
			return traducirErrorCuenta(err)
		}
		return uc.auditoria.RegistrarCambio(txCtx, dto.RegistrarCambioDTO{
			Entidad:   "usuarios",
			EntidadID: usuario,
			Accion:    "alta",
			Despues:   vistaAuditoria(cuenta, false),
			UsuarioID: &uid,
		})
	})
	if err != nil {
		return nil, err
	}
	guardada, err := uc.usuarioRepo.ObtenerPorUsuario(ctx, usuario)
	if err != nil {
		return nil, err
	}
	out := cuentaDe(guardada)
	return &out, nil
}

func (uc *usuarioUseCase) Actualizar(ctx context.Context, actorID int64, actorUsuario, usuario string, in dto.ActualizarCuentaDTO) (*dto.CuentaDTO, error) {
	usuario = strings.TrimSpace(usuario)
	actual, err := uc.usuarioRepo.ObtenerPorUsuario(ctx, usuario)
	if err != nil || actual == nil {
		return nil, domainErrors.ErrCuentaNoEncontrada
	}
	antes := vistaAuditoria(actual, false)
	if in.Nombre != nil {
		nombre := strings.TrimSpace(*in.Nombre)
		if nombre == "" || len(nombre) > 80 {
			return nil, domainErrors.ErrNombreInvalido
		}
		actual.Nombre = nombre
	}
	if in.Rol != nil {
		rol := strings.TrimSpace(*in.Rol)
		activo, err := uc.permisoRepo.RolActivo(ctx, rol)
		if err != nil {
			return nil, err
		}
		if !activo {
			return nil, domainErrors.ErrRolNoDisponible
		}
		actual.Rol = rol
	}
	if in.CapatazID != nil {
		actual.CapatazID = strings.TrimSpace(*in.CapatazID)
	}
	if in.Activo != nil {
		if !*in.Activo && usuario == strings.TrimSpace(actorUsuario) {
			return nil, domainErrors.ErrNoBajaPropia
		}
		actual.Activo = *in.Activo
	}
	claveCambiada := false
	if in.Clave != nil {
		if len(*in.Clave) < largoMinimoClave {
			return nil, domainErrors.ErrClaveInvalida
		}
		hash, err := uc.hasher.Hash(*in.Clave)
		if err != nil {
			return nil, err
		}
		actual.PasswordHash = hash
		actual.DebeCambiarPassword = true
		claveCambiada = true
	}
	uid := actorID
	accion := "edicion"
	if in.Activo != nil && !*in.Activo {
		accion = "baja"
	}
	err = uc.tx.Ejecutar(ctx, func(txCtx context.Context) error {
		if err := uc.usuarioRepo.Actualizar(txCtx, actual); err != nil {
			return traducirErrorCuenta(err)
		}
		return uc.auditoria.RegistrarCambio(txCtx, dto.RegistrarCambioDTO{
			Entidad:   "usuarios",
			EntidadID: usuario,
			Accion:    accion,
			Antes:     antes,
			Despues:   vistaAuditoria(actual, claveCambiada),
			UsuarioID: &uid,
		})
	})
	if err != nil {
		return nil, err
	}
	out := cuentaDe(actual)
	return &out, nil
}

func (uc *usuarioUseCase) CambiarClavePropia(ctx context.Context, usuario, actualClave, nueva string) (*dto.UsuarioSesionDTO, error) {
	cuenta, err := uc.usuarioRepo.ObtenerPorUsuario(ctx, strings.TrimSpace(usuario))
	if err != nil || cuenta == nil || !cuenta.Activo || !cuenta.RolActivo {
		return nil, domainErrors.ErrCredencialesInvalidas
	}
	if err := uc.hasher.Compare(cuenta.PasswordHash, actualClave); err != nil {
		return nil, domainErrors.ErrCredencialesInvalidas
	}
	if len(nueva) < largoMinimoClave {
		return nil, domainErrors.ErrClaveInvalida
	}
	hash, err := uc.hasher.Hash(nueva)
	if err != nil {
		return nil, err
	}
	antesFlag := cuenta.DebeCambiarPassword
	cuenta.PasswordHash = hash
	cuenta.DebeCambiarPassword = false
	uid := cuenta.ID
	err = uc.tx.Ejecutar(ctx, func(txCtx context.Context) error {
		if err := uc.usuarioRepo.Actualizar(txCtx, cuenta); err != nil {
			return err
		}
		return uc.auditoria.RegistrarCambio(txCtx, dto.RegistrarCambioDTO{
			Entidad:   "usuarios",
			EntidadID: cuenta.Usuario,
			Accion:    "edicion",
			Antes:     map[string]any{"debe_cambiar_password": antesFlag},
			Despues:   map[string]any{"debe_cambiar_password": false, "clave_actualizada": true},
			UsuarioID: &uid,
		})
	})
	if err != nil {
		return nil, err
	}
	return sesionDe(cuenta), nil
}

func (uc *usuarioUseCase) ActualizarPermiso(ctx context.Context, actorID int64, rol, accion string, concedido bool) error {
	rol = strings.TrimSpace(rol)
	accion = strings.TrimSpace(accion)
	if !reAccion.MatchString(accion) {
		return domainErrors.ErrAccionInvalida
	}
	roles, err := uc.permisoRepo.ListarRoles(ctx)
	if err != nil {
		return err
	}
	if buscarRol(roles, rol) == nil {
		return domainErrors.ErrRolNoDisponible
	}
	ya, err := uc.permisoRepo.Concedido(ctx, rol, accion)
	if err != nil {
		return err
	}
	if ya == concedido {
		return nil
	}
	uid := actorID
	return uc.tx.Ejecutar(ctx, func(txCtx context.Context) error {
		if err := uc.permisoRepo.Establecer(txCtx, rol, accion, concedido); err != nil {
			return traducirErrorCuenta(err)
		}
		return uc.auditoria.RegistrarCambio(txCtx, dto.RegistrarCambioDTO{
			Entidad:   "permisos",
			EntidadID: rol + ":" + accion,
			Accion:    "edicion",
			Antes:     map[string]any{"rol": rol, "accion": accion, "concedido": ya},
			Despues:   map[string]any{"rol": rol, "accion": accion, "concedido": concedido},
			UsuarioID: &uid,
		})
	})
}

func (uc *usuarioUseCase) CrearRol(ctx context.Context, actorID int64, codigo, nombre string) (*dto.RolDTO, error) {
	codigo = strings.TrimSpace(strings.ToLower(codigo))
	nombre = strings.TrimSpace(nombre)
	if !reCodigo.MatchString(codigo) {
		return nil, domainErrors.ErrCodigoRolInvalido
	}
	if nombre == "" || len(nombre) > 80 {
		return nil, domainErrors.ErrNombreInvalido
	}
	roles, err := uc.permisoRepo.ListarRoles(ctx)
	if err != nil {
		return nil, err
	}
	if buscarRol(roles, codigo) != nil {
		return nil, domainErrors.ErrRolYaExiste
	}
	rol := entities.RolCatalogo{Codigo: codigo, Nombre: nombre, Activo: true}
	uid := actorID
	err = uc.tx.Ejecutar(ctx, func(txCtx context.Context) error {
		if err := uc.permisoRepo.CrearRol(txCtx, rol); err != nil {
			if errors.Is(err, domainErrors.ErrRolYaExiste) {
				return err
			}
			return traducirErrorCuenta(err)
		}
		return uc.auditoria.RegistrarCambio(txCtx, dto.RegistrarCambioDTO{
			Entidad:   "roles",
			EntidadID: codigo,
			Accion:    "alta",
			Despues:   map[string]any{"codigo": codigo, "nombre": nombre, "activo": true},
			UsuarioID: &uid,
		})
	})
	if err != nil {
		return nil, err
	}
	return &dto.RolDTO{Codigo: codigo, Nombre: nombre, Activo: true}, nil
}

func (uc *usuarioUseCase) ActualizarRol(ctx context.Context, actorID int64, codigo string, activo bool) error {
	codigo = strings.TrimSpace(codigo)
	roles, err := uc.permisoRepo.ListarRoles(ctx)
	if err != nil {
		return err
	}
	actual := buscarRol(roles, codigo)
	if actual == nil {
		return domainErrors.ErrRolNoDisponible
	}
	if actual.Activo == activo {
		return nil
	}
	uid := actorID
	accion := "edicion"
	if !activo {
		accion = "baja"
	}
	return uc.tx.Ejecutar(ctx, func(txCtx context.Context) error {
		if err := uc.permisoRepo.ActualizarRol(txCtx, codigo, activo); err != nil {
			return err
		}
		return uc.auditoria.RegistrarCambio(txCtx, dto.RegistrarCambioDTO{
			Entidad:   "roles",
			EntidadID: codigo,
			Accion:    accion,
			Antes:     map[string]any{"codigo": actual.Codigo, "nombre": actual.Nombre, "activo": actual.Activo},
			Despues:   map[string]any{"codigo": actual.Codigo, "nombre": actual.Nombre, "activo": activo},
			UsuarioID: &uid,
		})
	})
}

func cuentaDe(u *entities.Usuario) dto.CuentaDTO {
	if u == nil {
		return dto.CuentaDTO{}
	}
	return dto.CuentaDTO{
		ID:                  u.ID,
		Usuario:             u.Usuario,
		Nombre:              u.Nombre,
		Rol:                 u.Rol,
		RolNombre:           u.RolNombre,
		CapatazID:           u.CapatazID,
		Activo:              u.Activo,
		DebeCambiarPassword: u.DebeCambiarPassword,
	}
}

func sesionDe(u *entities.Usuario) *dto.UsuarioSesionDTO {
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

func vistaAuditoria(u *entities.Usuario, claveActualizada bool) map[string]any {
	vista := map[string]any{
		"usuario":               u.Usuario,
		"nombre":                u.Nombre,
		"rol":                   u.Rol,
		"capataz_id":            u.CapatazID,
		"activo":                u.Activo,
		"debe_cambiar_password": u.DebeCambiarPassword,
	}
	if claveActualizada {
		vista["clave_actualizada"] = true
	}
	return vista
}

func buscarRol(roles []entities.RolCatalogo, codigo string) *entities.RolCatalogo {
	for i := range roles {
		if roles[i].Codigo == codigo {
			return &roles[i]
		}
	}
	return nil
}

func traducirErrorCuenta(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, domainErrors.ErrRolYaExiste) || errors.Is(err, domainErrors.ErrCuadrillaNoExiste) {
		return err
	}
	texto := err.Error()
	switch {
	case strings.Contains(texto, "usuarios_usuario_key") || strings.Contains(texto, "duplicate key") && strings.Contains(texto, "usuarios"):
		return domainErrors.ErrUsuarioYaExiste
	case strings.Contains(texto, "roles_pkey") || strings.Contains(texto, "duplicate key") && strings.Contains(texto, "roles"):
		return domainErrors.ErrRolYaExiste
	case strings.Contains(texto, "usuarios_capataz_id_fkey") || strings.Contains(texto, "capataces"):
		return domainErrors.ErrCuadrillaNoExiste
	case strings.Contains(texto, "usuarios_rol_fkey") || strings.Contains(texto, "permisos_rol_fkey"):
		return domainErrors.ErrRolNoDisponible
	default:
		return err
	}
}
