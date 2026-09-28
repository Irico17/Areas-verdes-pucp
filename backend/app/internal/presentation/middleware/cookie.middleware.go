package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

const (
	// CookieSesion is the session cookie name.
	CookieSesion = "cv_sesion"
	ctxCookie    = "cookie_opts"
)

// OpcionesCookie defines cookie attributes.
type OpcionesCookie struct {
	Secure   bool
	SameSite http.SameSite
}

// ConCookie attaches cookie policy options to the gin context.
func ConCookie(opts OpcionesCookie) gin.HandlerFunc {
	if opts.SameSite == 0 {
		opts.SameSite = http.SameSiteLaxMode
	}
	return func(c *gin.Context) {
		c.Set(ctxCookie, opts)
		c.Next()
	}
}

// ObtenerOpcionesCookie retrieves the cookie configuration from context.
func ObtenerOpcionesCookie(c *gin.Context) OpcionesCookie {
	v, ok := c.Get(ctxCookie)
	if !ok {
		return OpcionesCookie{SameSite: http.SameSiteLaxMode}
	}
	opts, ok := v.(OpcionesCookie)
	if !ok {
		return OpcionesCookie{SameSite: http.SameSiteLaxMode}
	}
	return opts
}

// EscribirCookie writes cv_sesion cookie to the HTTP response.
func EscribirCookie(w http.ResponseWriter, opts OpcionesCookie, token string, maxAge int) {
	same := opts.SameSite
	if same == 0 {
		same = http.SameSiteLaxMode
	}
	http.SetCookie(w, &http.Cookie{
		Name:     CookieSesion,
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   opts.Secure,
		SameSite: same,
		MaxAge:   maxAge,
	})
}

// ParseSameSite converts string to http.SameSite.
func ParseSameSite(s string) http.SameSite {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "strict":
		return http.SameSiteStrictMode
	case "none":
		return http.SameSiteNoneMode
	default:
		return http.SameSiteLaxMode
	}
}
