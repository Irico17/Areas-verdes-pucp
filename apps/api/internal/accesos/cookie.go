package accesos

import "net/http"

// OpcionesCookie sale del entorno. HttpOnly va siempre.
// Secure solo si hay TLS; en HTTP el navegador no guarda la cookie. SameSite no se deja en None salvo que se pida.
type OpcionesCookie struct {
	Secure   bool
	SameSite http.SameSite
}

// EscribirCookie deja cv_sesion. El valor vacío y maxAge negativo cierran la sesión.
func EscribirCookie(w http.ResponseWriter, opts OpcionesCookie, token string, maxAge int) {
	same := opts.SameSite
	if same == 0 {
		same = http.SameSiteLaxMode
	}
	http.SetCookie(w, &http.Cookie{
		Name:     Cookie,
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   opts.Secure,
		SameSite: same,
		MaxAge:   maxAge,
	})
}
