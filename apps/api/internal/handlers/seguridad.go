package handlers

import (
	"net/http"
	"time"

	"campusverde/api/internal/accesos"

	"github.com/gin-gonic/gin"
)

const ctxCookie = "cookie_opts"

// OpcionesDeSeguridad las arma el proceso a partir del entorno.
type OpcionesDeSeguridad struct {
	CORSOrigins  []string
	Cookie       accesos.OpcionesCookie
	LoginCada    int
	LoginVentana time.Duration
}

// CORS refleja solo un origen de la lista. Nunca escribe asterisco.
func CORS(origenes []string) gin.HandlerFunc {
	permitidos := map[string]struct{}{}
	for _, o := range origenes {
		if o == "" || o == "*" {
			continue
		}
		permitidos[o] = struct{}{}
	}
	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		if _, ok := permitidos[origin]; ok {
			c.Header("Access-Control-Allow-Origin", origin)
			c.Header("Vary", "Origin")
			c.Header("Access-Control-Allow-Credentials", "true")
		}
		c.Header("Access-Control-Allow-Methods", "GET, POST, PATCH, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type")
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}

// ConCookie deja la política de cookie en el contexto para el login.
func ConCookie(opts accesos.OpcionesCookie) gin.HandlerFunc {
	if opts.SameSite == 0 {
		opts.SameSite = http.SameSiteLaxMode
	}
	return func(c *gin.Context) {
		c.Set(ctxCookie, opts)
		c.Next()
	}
}

func cookieDe(c *gin.Context) accesos.OpcionesCookie {
	v, ok := c.Get(ctxCookie)
	if !ok {
		return accesos.OpcionesCookie{SameSite: http.SameSiteLaxMode}
	}
	opts, ok := v.(accesos.OpcionesCookie)
	if !ok {
		return accesos.OpcionesCookie{SameSite: http.SameSiteLaxMode}
	}
	return opts
}

// LimiteLogin corta POST /api/v1/sesion cuando la misma IP insiste.
func LimiteLogin(max int, ventana time.Duration) gin.HandlerFunc {
	lim := accesos.NuevoLimite(max, ventana)
	return func(c *gin.Context) {
		if c.Request.Method == http.MethodPost && c.Request.URL.Path == "/api/v1/sesion" {
			if !lim.Permitir(c.ClientIP()) {
				c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": "demasiados intentos de ingreso"})
				return
			}
		}
		c.Next()
	}
}
