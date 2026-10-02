package postgres_test

import (
	"context"
	"testing"

	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/services"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/usecases"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/entities"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/infrastructure/seguridad"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/persistence/database"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/persistence/repository/postgres"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/shared/testutil"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func TestEnsureConservaFilasExtraYEsIdempotente(t *testing.T) {
	sqlDB, gdb := testutil.MigrarDBTemporal(t, "accesos_ensure")
	defer sqlDB.Close()

	usuarioRepo := postgres.NewUsuarioRepository(gdb)
	permisoRepo := postgres.NewPermisoRepository(gdb)
	sesionRepo := postgres.NewSesionRepository(gdb)
	hasher := seguridad.NewBcryptHasher()
	tx := database.NewTransaccion(gdb)
	permisosSvc := services.NewPermisosService()

	semillaUC := usecases.NewSemillaAccesosUseCase(tx, usuarioRepo, permisoRepo, permisosSvc, hasher)
	sesionUC := usecases.NewSesionUseCase(sesionRepo, usuarioRepo, hasher)
	usuarioUC := usecases.NewUsuarioUseCase(usuarioRepo, permisoRepo)
	ctx := context.Background()

	// 1. Insertamos un permiso extra
	if err := gdb.Exec(`INSERT INTO permisos (rol, accion) VALUES ('admin', 'accion_extra') ON CONFLICT DO NOTHING`).Error; err != nil {
		t.Fatal(err)
	}

	// 2. Ejecutamos Ensure
	if err := semillaUC.Ensure(ctx, "pando-local"); err != nil {
		t.Fatalf("primer Ensure falló: %v", err)
	}

	// Verificamos que el permiso extra sobrevive
	var n int
	if err := gdb.Raw(`SELECT count(*) FROM permisos WHERE rol = 'admin' AND accion = 'accion_extra'`).Scan(&n).Error; err != nil {
		t.Fatal(err)
	}
	if n != 1 {
		t.Fatalf("el permiso extra no sobrevivió a Ensure: n=%d", n)
	}

	// Verificamos que jefatura tiene 'evidencias'
	var nJef int
	if err := gdb.Raw(`SELECT count(*) FROM permisos WHERE rol = 'jefatura' AND accion = 'evidencias'`).Scan(&nJef).Error; err != nil {
		t.Fatal(err)
	}
	if nJef != 1 {
		t.Fatalf("jefatura no tiene permiso de evidencias: n=%d", nJef)
	}

	var count1 int
	if err := gdb.Raw(`SELECT count(*) FROM permisos`).Scan(&count1).Error; err != nil {
		t.Fatal(err)
	}

	// 3. Ejecutamos Ensure por segunda vez (idempotencia)
	if err := semillaUC.Ensure(ctx, "pando-local"); err != nil {
		t.Fatalf("segundo Ensure falló: %v", err)
	}

	var count2 int
	if err := gdb.Raw(`SELECT count(*) FROM permisos`).Scan(&count2).Error; err != nil {
		t.Fatal(err)
	}
	if count1 != count2 {
		t.Fatalf("Ensure no fue idempotente: count1=%d, count2=%d", count1, count2)
	}

	// 4. Verificamos que Login y Resolver devuelven rol_nombre
	token, u, err := sesionUC.Login(ctx, "jefatura", "pando-local")
	if err != nil {
		t.Fatalf("Login falló: %v", err)
	}
	if u.RolNombre != "Jefatura de sección" {
		t.Fatalf("Login rol_nombre esperado 'Jefatura de sección', obtuve %q", u.RolNombre)
	}

	uTok, err := sesionUC.Resolver(ctx, token)
	if err != nil {
		t.Fatalf("Resolver falló: %v", err)
	}
	if uTok.RolNombre != "Jefatura de sección" {
		t.Fatalf("Resolver rol_nombre esperado 'Jefatura de sección', obtuve %q", uTok.RolNombre)
	}

	// Verificamos ListarUsuarios
	res, err := usuarioUC.ListarUsuarios(ctx)
	if err != nil {
		t.Fatalf("ListarUsuarios falló: %v", err)
	}
	encontrado := false
	for _, usr := range res.Usuarios {
		if usr.Usuario == "coordinacion" {
			encontrado = true
			if usr.RolNombre != "Ingeniería / Coordinación" {
				t.Fatalf("Usuarios rol_nombre esperado 'Ingeniería / Coordinación', obtuve %q", usr.RolNombre)
			}
		}
	}
	if !encontrado {
		t.Fatal("usuario coordinacion no encontrado en ListarUsuarios()")
	}
}

func TestUsuariosRolForeignKeyYCatalogoRoles(t *testing.T) {
	sqlDB, gdb := testutil.MigrarDBTemporal(t, "usuarios_rol_fk")
	defer sqlDB.Close()

	ctx := context.Background()
	usuarioRepo := postgres.NewUsuarioRepository(gdb)
	permisoRepo := postgres.NewPermisoRepository(gdb)
	hasher := seguridad.NewBcryptHasher()

	semillaUC := usecases.NewSemillaAccesosUseCase(database.NewTransaccion(gdb), usuarioRepo, permisoRepo, services.NewPermisosService(), hasher)
	if err := semillaUC.Ensure(ctx, "pando-local"); err != nil {
		t.Fatalf("Ensure falló: %v", err)
	}

	// 1. Verificar que usuarios_rol_fkey existe y usuarios_rol_chk fue eliminado
	var fkCount int
	if err := gdb.Raw(`SELECT count(*) FROM pg_constraint WHERE conname = 'usuarios_rol_fkey'`).Scan(&fkCount).Error; err != nil {
		t.Fatal(err)
	}
	if fkCount != 1 {
		t.Fatalf("falta restricción usuarios_rol_fkey en usuarios (count=%d)", fkCount)
	}

	var chkCount int
	if err := gdb.Raw(`SELECT count(*) FROM pg_constraint WHERE conname = 'usuarios_rol_chk'`).Scan(&chkCount).Error; err != nil {
		t.Fatal(err)
	}
	if chkCount != 0 {
		t.Fatalf("usuarios_rol_chk no debe existir tras migración 046 (count=%d)", chkCount)
	}

	// 2. Verificar que insertar un rol inexistente falla por FK
	err := usuarioRepo.Crear(ctx, &entities.Usuario{
		Usuario:      "invalido",
		Nombre:       "Rol Inválido",
		Rol:          "rol_inexistente",
		PasswordHash: "$2a$10$abcdefghijklmnopqrstuu",
	})
	if err == nil {
		t.Fatal("se esperaba error al insertar usuario con rol inexistente en catálogo roles")
	}

	// 3. Demostrar roles configurables (RNF-05): agregar nuevo rol en catálogo roles
	if err := gdb.Exec(`
		INSERT INTO roles (codigo, nombre, orden, descripcion, activo)
		VALUES ('supervisor_nuevo', 'Supervisor Nuevo', 5, 'Rol de prueba configurable', true)
		ON CONFLICT (codigo) DO NOTHING
	`).Error; err != nil {
		t.Fatalf("no se pudo insertar rol configurable: %v", err)
	}

	// 4. Insertar usuario con el nuevo rol configurable debe tener éxito
	err = usuarioRepo.Crear(ctx, &entities.Usuario{
		Usuario:      "nuevo_supervisor",
		Nombre:       "Nuevo Supervisor",
		Rol:          "supervisor_nuevo",
		PasswordHash: "$2a$10$abcdefghijklmnopqrstuu",
	})
	if err != nil {
		t.Fatalf("falló inserción de usuario con rol configurable: %v", err)
	}

	// 5. Verificar que ObtenerPorUsuario resuelve rol_nombre del nuevo rol
	uNuevo, err := usuarioRepo.ObtenerPorUsuario(ctx, "nuevo_supervisor")
	if err != nil {
		t.Fatalf("ObtenerPorUsuario falló: %v", err)
	}
	if uNuevo.Rol != "supervisor_nuevo" {
		t.Errorf("uNuevo.Rol = %q, esperado 'supervisor_nuevo'", uNuevo.Rol)
	}
	if uNuevo.RolNombre != "Supervisor Nuevo" {
		t.Errorf("uNuevo.RolNombre = %q, esperado 'Supervisor Nuevo'", uNuevo.RolNombre)
	}

	// 6. Idempotencia: volver a aplicar las migraciones no causa error
	if err := database.Apply(gdb, testutil.FindMigrationsDir()); err != nil {
		t.Fatalf("re-aplicar migraciones falló: %v", err)
	}
}
