package operacion

import (
	"errors"
	"math"
	"regexp"
	"strings"
	"unicode/utf8"
)

var (
	// ErrProhibido: el rol no puede ejecutar la acción.
	ErrProhibido = errors.New("prohibido")
	// ErrNoEncontrada: no hay labor con ese id.
	ErrNoEncontrada = errors.New("no encontrada")
	// ErrConflicto: el UUID ya existe con otro contenido.
	ErrConflicto = errors.New("conflicto")
	// ErrValidacion marca un InputError.
	ErrValidacion = errors.New("validacion")
)

// InputError es un 400 con mensaje para la UI.
type InputError struct{ Reason string }

func (e InputError) Error() string { return e.Reason }
func (e InputError) Unwrap() error { return ErrValidacion }

var uuidRe = regexp.MustCompile(`(?i)^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)

func known(list []string, value string) bool {
	for _, item := range list {
		if item == value {
			return true
		}
	}
	return false
}

func rolConocido(rol string) bool {
	return rol == RolJefatura || rol == RolCoordinacion || rol == RolCapataz
}

func puedeAsignar(rol string) bool {
	return rol == RolJefatura || rol == RolCoordinacion
}

// ValidateCreate rechaza altas incompletas o de capataz antes de tocar la base.
func ValidateCreate(in CreateInput) error {
	if !puedeAsignar(in.ActorRol) {
		if in.ActorRol == RolCapataz {
			return ErrProhibido
		}
		return InputError{Reason: "actor_rol debe ser jefatura o coordinacion"}
	}
	if !uuidRe.MatchString(in.ID) {
		return InputError{Reason: "id debe ser un UUID"}
	}
	if !known(Tipos, in.Tipo) {
		return InputError{Reason: "tipo no reconocido"}
	}
	titulo := strings.TrimSpace(in.Titulo)
	if titulo == "" || utf8.RuneCountInString(titulo) > 160 {
		return InputError{Reason: "titulo es obligatorio y de hasta 160 caracteres"}
	}
	if utf8.RuneCountInString(in.Detalle) > 2000 {
		return InputError{Reason: "detalle admite hasta 2000 caracteres"}
	}
	if in.Lon < MinLon || in.Lon > MaxLon || in.Lat < MinLat || in.Lat > MaxLat {
		return InputError{Reason: "el punto queda fuera del campus"}
	}
	return nil
}

// ValidateQuery exige equipo cuando el rol es capataz.
func ValidateQuery(q Query) error {
	if q.Rol != "" && !rolConocido(q.Rol) {
		return InputError{Reason: "rol no reconocido"}
	}
	if q.Rol == RolCapataz && strings.TrimSpace(q.CapatazID) == "" {
		return InputError{Reason: "capataz_id es obligatorio para el rol capataz"}
	}
	if q.Estado != "" && !known(Estados, q.Estado) {
		return InputError{Reason: "estado no reconocido"}
	}
	if q.Tipo != "" && !known(Tipos, q.Tipo) {
		return InputError{Reason: "tipo no reconocido"}
	}
	return nil
}

// ValidateEstado comprueba el nuevo estado y el rol.
func ValidateEstado(estado, actorRol, capatazID string) error {
	if !rolConocido(actorRol) {
		return InputError{Reason: "actor_rol no reconocido"}
	}
	if !known(Estados, estado) {
		return InputError{Reason: "estado no reconocido"}
	}
	if actorRol == RolCapataz && strings.TrimSpace(capatazID) == "" {
		return InputError{Reason: "capataz_id es obligatorio para cambiar el estado"}
	}
	return nil
}

// ValidateAsignacion impide que un capataz reasigne.
func ValidateAsignacion(actorRol, capatazID string) error {
	if !puedeAsignar(actorRol) {
		if actorRol == RolCapataz {
			return ErrProhibido
		}
		return InputError{Reason: "actor_rol debe ser jefatura o coordinacion"}
	}
	if strings.TrimSpace(capatazID) == "" {
		return InputError{Reason: "capataz_id es obligatorio"}
	}
	return nil
}

// ValidateArchivo solo lo hace jefatura o coordinación.
func ValidateArchivo(actorRol string) error {
	if puedeAsignar(actorRol) {
		return nil
	}
	if actorRol == RolCapataz {
		return ErrProhibido
	}
	return InputError{Reason: "actor_rol debe ser jefatura o coordinacion"}
}

// SamePayload compara un reintento con la fila ya guardada.
func SamePayload(saved Saved, in CreateInput) bool {
	return saved.Tipo == in.Tipo &&
		saved.Titulo == strings.TrimSpace(in.Titulo) &&
		saved.Detalle == strings.TrimSpace(in.Detalle) &&
		saved.AssignedCapatazID == strings.TrimSpace(in.AssignedCapatazID) &&
		saved.AreaFeatureID == strings.TrimSpace(in.AreaFeatureID) &&
		saved.ZonaFeatureID == strings.TrimSpace(in.ZonaFeatureID) &&
		math.Abs(saved.Lon-in.Lon) < 1e-5 &&
		math.Abs(saved.Lat-in.Lat) < 1e-5
}

func eventoEstado(estado string) string {
	if estado == "cancelada" {
		return "cancelada"
	}
	return "estado"
}

func blankPtr(v string) *string {
	v = strings.TrimSpace(v)
	if v == "" {
		return nil
	}
	return &v
}
