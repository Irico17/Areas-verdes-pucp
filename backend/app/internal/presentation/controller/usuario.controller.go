package controller

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"

	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/contracts"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/dto"
	domainErrors "github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/errors"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/presentation/middleware"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/presentation/requests"
)

// IUsuarioController defines HTTP handlers for user accounts.
type IUsuarioController interface {
	Listar(c *gin.Context)
	Crear(c *gin.Context)
	Actualizar(c *gin.Context)
	CambiarClavePropia(c *gin.Context)
	ActualizarPermiso(c *gin.Context)
	CrearRol(c *gin.Context)
	ActualizarRol(c *gin.Context)
}

type usuarioController struct {
	usuarioUC contracts.IUsuarioUseCase
	permisos  contracts.IPermisosService
	logger    zerolog.Logger
}

// NewUsuarioController creates a new user controller.
func NewUsuarioController(
	usuarioUC contracts.IUsuarioUseCase,
	permisos contracts.IPermisosService,
	logger zerolog.Logger,
) IUsuarioController {
	return &usuarioController{usuarioUC: usuarioUC, permisos: permisos, logger: logger}
}

func (ctrl *usuarioController) exigeUsuarios(c *gin.Context) (dto.UsuarioSesionDTO, bool) {
	u, ok := middleware.UsuarioEn(c)
	if !ok || ctrl.permisos == nil || !ctrl.permisos.Permite(u.Rol, "usuarios") {
		c.JSON(http.StatusForbidden, gin.H{"error": domainErrors.ErrSinPermiso.Error()})
		return dto.UsuarioSesionDTO{}, false
	}
	return u, true
}

// Listar handles GET /accesos/usuarios.
// @Summary Listar cuentas y permisos
// @Description Devuelve cuentas, roles y la matriz. Jefatura de sección y administrador del sistema.
// @Tags Accesos
// @Produce json
// @Success 200 {object} dto.UsuariosResponseDTO
// @Failure 403 {object} map[string]string "su rol no tiene ese permiso"
// @Failure 500 {object} map[string]string "no se pudieron leer las cuentas"
// @Router /v1/accesos/usuarios [get]
func (ctrl *usuarioController) Listar(c *gin.Context) {
	if _, ok := ctrl.exigeUsuarios(c); !ok {
		return
	}
	res, err := ctrl.usuarioUC.ListarUsuarios(c.Request.Context())
	if err != nil {
		ctrl.logger.Error().Err(err).Msg("usuario: error al listar cuentas")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "no se pudieron leer las cuentas"})
		return
	}
	c.JSON(http.StatusOK, res)
}

// Crear handles POST /accesos/usuarios.
// @Summary Alta de cuenta
// @Description Crea una cuenta con clave propia y debe_cambiar_password
// @Tags Accesos
// @Accept json
// @Produce json
// @Param cuerpo body requests.CrearUsuarioRequest true "Cuenta nueva"
// @Success 201 {object} map[string]any "cuenta creada"
// @Failure 400 {object} map[string]string "datos inválidos"
// @Failure 403 {object} map[string]string "su rol no tiene ese permiso"
// @Failure 409 {object} map[string]string "usuario ya existe"
// @Router /v1/accesos/usuarios [post]
func (ctrl *usuarioController) Crear(c *gin.Context) {
	actor, ok := ctrl.exigeUsuarios(c)
	if !ok {
		return
	}
	var req requests.CrearUsuarioRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "JSON inválido"})
		return
	}
	cuenta, err := ctrl.usuarioUC.Crear(c.Request.Context(), actor.ID, dto.CrearCuentaDTO{
		Usuario:   req.Usuario,
		Nombre:    req.Nombre,
		Rol:       req.Rol,
		Clave:     req.Clave,
		CapatazID: req.CapatazID,
	})
	if err != nil {
		ctrl.responder(c, err, "no se pudo crear la cuenta")
		return
	}
	c.JSON(http.StatusCreated, gin.H{"cuenta": cuenta})
}

// Actualizar handles PATCH /accesos/usuarios/:usuario.
// @Summary Actualizar cuenta
// @Description Cambia nombre, rol, activo o clave. La baja es lógica.
// @Tags Accesos
// @Accept json
// @Produce json
// @Param usuario path string true "Usuario"
// @Param cuerpo body requests.ActualizarUsuarioRequest true "Campos a cambiar"
// @Success 200 {object} map[string]any "cuenta actualizada"
// @Failure 400 {object} map[string]string "datos inválidos"
// @Failure 403 {object} map[string]string "su rol no tiene ese permiso"
// @Failure 404 {object} map[string]string "no existe esa cuenta"
// @Router /v1/accesos/usuarios/{usuario} [patch]
func (ctrl *usuarioController) Actualizar(c *gin.Context) {
	actor, ok := ctrl.exigeUsuarios(c)
	if !ok {
		return
	}
	var req requests.ActualizarUsuarioRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "JSON inválido"})
		return
	}
	cuenta, err := ctrl.usuarioUC.Actualizar(c.Request.Context(), actor.ID, actor.Usuario, c.Param("usuario"), dto.ActualizarCuentaDTO{
		Nombre:    req.Nombre,
		Rol:       req.Rol,
		Activo:    req.Activo,
		Clave:     req.Clave,
		CapatazID: req.CapatazID,
	})
	if err != nil {
		ctrl.responder(c, err, "no se pudo actualizar la cuenta")
		return
	}
	c.JSON(http.StatusOK, gin.H{"cuenta": cuenta})
}

// CambiarClavePropia handles POST /sesion/clave for the signed-in person.
// @Summary Cambiar la clave propia
// @Description La persona reemplaza su clave y limpia debe_cambiar_password
// @Tags Accesos
// @Accept json
// @Produce json
// @Param cuerpo body requests.CambiarClaveRequest true "Clave actual y nueva"
// @Success 200 {object} map[string]any "clave actualizada"
// @Failure 400 {object} map[string]string "clave inválida"
// @Failure 401 {object} map[string]string "sin sesión o clave incorrecta"
// @Router /v1/sesion/clave [post]
func (ctrl *usuarioController) CambiarClavePropia(c *gin.Context) {
	u, ok := middleware.UsuarioEn(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": domainErrors.ErrSinSesion.Error()})
		return
	}
	var req requests.CambiarClaveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "JSON inválido"})
		return
	}
	usuario, err := ctrl.usuarioUC.CambiarClavePropia(c.Request.Context(), u.Usuario, req.ClaveActual, req.ClaveNueva)
	if err != nil {
		ctrl.responder(c, err, "no se pudo cambiar la clave")
		return
	}
	c.JSON(http.StatusOK, gin.H{"usuario": usuario})
}

// ActualizarPermiso handles PATCH /accesos/permisos.
// @Summary Actualizar un permiso
// @Description Concede o retira una acción. Permite lo lee sin reiniciar.
// @Tags Accesos
// @Accept json
// @Produce json
// @Param cuerpo body requests.ActualizarPermisoRequest true "Celda de la matriz"
// @Success 200 {object} map[string]bool "ok"
// @Failure 400 {object} map[string]string "acción inválida"
// @Failure 403 {object} map[string]string "su rol no tiene ese permiso"
// @Router /v1/accesos/permisos [patch]
func (ctrl *usuarioController) ActualizarPermiso(c *gin.Context) {
	actor, ok := ctrl.exigeUsuarios(c)
	if !ok {
		return
	}
	var req requests.ActualizarPermisoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "JSON inválido"})
		return
	}
	if err := ctrl.usuarioUC.ActualizarPermiso(c.Request.Context(), actor.ID, req.Rol, req.Accion, req.Concedido); err != nil {
		ctrl.responder(c, err, "no se pudo actualizar el permiso")
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// CrearRol handles POST /accesos/roles.
// @Summary Alta de rol
// @Description Crea un rol sin renombrar los códigos existentes
// @Tags Accesos
// @Accept json
// @Produce json
// @Param cuerpo body requests.CrearRolRequest true "Código y nombre"
// @Success 201 {object} map[string]any "rol creado"
// @Failure 400 {object} map[string]string "código inválido"
// @Failure 403 {object} map[string]string "su rol no tiene ese permiso"
// @Failure 409 {object} map[string]string "el rol ya existe"
// @Router /v1/accesos/roles [post]
func (ctrl *usuarioController) CrearRol(c *gin.Context) {
	actor, ok := ctrl.exigeUsuarios(c)
	if !ok {
		return
	}
	var req requests.CrearRolRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "JSON inválido"})
		return
	}
	rol, err := ctrl.usuarioUC.CrearRol(c.Request.Context(), actor.ID, req.Codigo, req.Nombre)
	if err != nil {
		ctrl.responder(c, err, "no se pudo crear el rol")
		return
	}
	c.JSON(http.StatusCreated, gin.H{"rol": rol})
}

// ActualizarRol handles PATCH /accesos/roles/:codigo.
// @Summary Activar o desactivar un rol
// @Description Un rol inactivo no inicia sesión. No renombra el código.
// @Tags Accesos
// @Accept json
// @Produce json
// @Param codigo path string true "Código del rol"
// @Param cuerpo body requests.ActualizarRolRequest true "Activo"
// @Success 200 {object} map[string]bool "ok"
// @Failure 400 {object} map[string]string "JSON inválido"
// @Failure 403 {object} map[string]string "su rol no tiene ese permiso"
// @Router /v1/accesos/roles/{codigo} [patch]
func (ctrl *usuarioController) ActualizarRol(c *gin.Context) {
	actor, ok := ctrl.exigeUsuarios(c)
	if !ok {
		return
	}
	var req requests.ActualizarRolRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.Activo == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "JSON inválido"})
		return
	}
	if err := ctrl.usuarioUC.ActualizarRol(c.Request.Context(), actor.ID, c.Param("codigo"), *req.Activo); err != nil {
		ctrl.responder(c, err, "no se pudo actualizar el rol")
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func (ctrl *usuarioController) responder(c *gin.Context, err error, generico string) {
	switch {
	case errors.Is(err, domainErrors.ErrUsuarioYaExiste), errors.Is(err, domainErrors.ErrRolYaExiste):
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
	case errors.Is(err, domainErrors.ErrCuentaNoEncontrada):
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
	case errors.Is(err, domainErrors.ErrCredencialesInvalidas):
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
	case errors.Is(err, domainErrors.ErrClaveInvalida),
		errors.Is(err, domainErrors.ErrNombreInvalido),
		errors.Is(err, domainErrors.ErrUsuarioInvalido),
		errors.Is(err, domainErrors.ErrRolNoDisponible),
		errors.Is(err, domainErrors.ErrNoBajaPropia),
		errors.Is(err, domainErrors.ErrCodigoRolInvalido),
		errors.Is(err, domainErrors.ErrAccionInvalida),
		errors.Is(err, domainErrors.ErrCuadrillaNoExiste):
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	default:
		ctrl.logger.Error().Err(err).Msg("usuario: error de cuentas")
		c.JSON(http.StatusInternalServerError, gin.H{"error": generico})
	}
}
