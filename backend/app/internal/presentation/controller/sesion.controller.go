package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/contracts"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/presentation/middleware"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/presentation/requests"
)

// ISesionController defines HTTP handlers for user authentication and session management.
type ISesionController interface {
	Entrar(c *gin.Context)
	Actual(c *gin.Context)
	Salir(c *gin.Context)
}

type sesionController struct {
	sesionUC contracts.ISesionUseCase
}

// NewSesionController creates a new session controller.
func NewSesionController(sesionUC contracts.ISesionUseCase) ISesionController {
	return &sesionController{sesionUC: sesionUC}
}

// Entrar handles POST /sesion.
// @Summary Iniciar sesión
// @Description Autentica usuario y establece cookie de sesión cv_sesion
// @Tags Sesion
// @Accept json
// @Produce json
// @Param credenciales body requests.LoginRequest true "Credenciales de usuario"
// @Success 200 {object} map[string]any "usuario autenticado"
// @Failure 400 {object} map[string]string "JSON inválido"
// @Failure 401 {object} map[string]string "usuario o clave incorrectos"
// @Failure 503 {object} map[string]string "base de datos no disponible"
// @Router /v1/sesion [post]
func (ctrl *sesionController) Entrar(c *gin.Context) {
	if ctrl.sesionUC == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "base de datos no disponible"})
		return
	}

	var req requests.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "JSON inválido"})
		return
	}

	token, user, err := ctrl.sesionUC.Login(c.Request.Context(), req.Usuario, req.Clave)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "usuario o clave incorrectos"})
		return
	}

	opts := middleware.ObtenerOpcionesCookie(c)
	middleware.EscribirCookie(c.Writer, opts, token, 12*60*60)
	c.JSON(http.StatusOK, gin.H{"usuario": user})
}

// Actual handles GET /sesion.
// @Summary Obtener sesión actual
// @Description Devuelve los datos del usuario asociado a la sesión activa
// @Tags Sesion
// @Produce json
// @Success 200 {object} map[string]any "usuario de la sesión"
// @Failure 401 {object} map[string]string "sin sesión"
// @Router /v1/sesion [get]
func (ctrl *sesionController) Actual(c *gin.Context) {
	u, ok := middleware.UsuarioEn(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "sin sesión"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"usuario": u})
}

// Salir handles DELETE /sesion.
// @Summary Cerrar sesión
// @Description Invalida el token en el servidor y limpia la cookie de sesión
// @Tags Sesion
// @Produce json
// @Success 200 {object} map[string]bool "ok: true"
// @Router /v1/sesion [delete]
func (ctrl *sesionController) Salir(c *gin.Context) {
	if ctrl.sesionUC != nil {
		if cookie, err := c.Request.Cookie(middleware.CookieSesion); err == nil {
			ctrl.sesionUC.Logout(c.Request.Context(), cookie.Value)
		}
	}
	opts := middleware.ObtenerOpcionesCookie(c)
	middleware.EscribirCookie(c.Writer, opts, "", -1)
	c.JSON(http.StatusOK, gin.H{"ok": true})
}
