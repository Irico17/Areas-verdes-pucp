package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/contracts"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/constants/enums"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/presentation/middleware"
)

// IUsuarioController defines HTTP handlers for user accounts.
type IUsuarioController interface {
	Listar(c *gin.Context)
}

type usuarioController struct {
	usuarioUC contracts.IUsuarioUseCase
}

// NewUsuarioController creates a new user controller.
func NewUsuarioController(usuarioUC contracts.IUsuarioUseCase) IUsuarioController {
	return &usuarioController{usuarioUC: usuarioUC}
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
		c.JSON(http.StatusForbidden, gin.H{"error": "solo administración ve las cuentas"})
		return
	}

	if ctrl.usuarioUC == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "base de datos no disponible"})
		return
	}

	res, err := ctrl.usuarioUC.ListarUsuarios(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "no se pudieron leer las cuentas"})
		return
	}

	c.JSON(http.StatusOK, res)
}
