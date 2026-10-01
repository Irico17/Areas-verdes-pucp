package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"

	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/contracts"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/constants/enums"
	domainErrors "github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/errors"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/presentation/middleware"
)

// IUsuarioController defines HTTP handlers for user accounts.
type IUsuarioController interface {
	Listar(c *gin.Context)
}

type usuarioController struct {
	usuarioUC contracts.IUsuarioUseCase
	logger    zerolog.Logger
}

// NewUsuarioController creates a new user controller.
func NewUsuarioController(usuarioUC contracts.IUsuarioUseCase, logger zerolog.Logger) IUsuarioController {
	return &usuarioController{usuarioUC: usuarioUC, logger: logger}
}

// Listar handles GET /accesos/usuarios.
// @Summary Listar cuentas y permisos
// @Description Devuelve la lista de cuentas locales y permisos (solo admin)
// @Tags Accesos
// @Produce json
// @Success 200 {object} dto.UsuariosResponseDTO
// @Failure 403 {object} map[string]string "solo administración ve las cuentas"
// @Failure 500 {object} map[string]string "no se pudieron leer las cuentas"
// @Failure 503 {object} map[string]string "base de datos no disponible"
// @Router /v1/accesos/usuarios [get]
func (ctrl *usuarioController) Listar(c *gin.Context) {
	u, ok := middleware.UsuarioEn(c)
	if !ok || u.Rol != enums.RolAdmin.String() {
		c.JSON(http.StatusForbidden, gin.H{"error": domainErrors.ErrSoloAdmin.Error()})
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
