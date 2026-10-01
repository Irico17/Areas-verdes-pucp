// Package enums defines domain enumeration types.
package enums

// TurnoRiego represents the irrigation shift.
type TurnoRiego string

const (
	// TurnoRiegoManana is the morning shift.
	TurnoRiegoManana TurnoRiego = "manana"
	// TurnoRiegoTarde is the afternoon shift.
	TurnoRiegoTarde TurnoRiego = "tarde"
)

// String returns the string representation.
func (t TurnoRiego) String() string {
	return string(t)
}

// EsTurnoRiegoValido checks if the shift is recognized.
func EsTurnoRiegoValido(v string) bool {
	switch TurnoRiego(v) {
	case TurnoRiegoManana, TurnoRiegoTarde:
		return true
	default:
		return false
	}
}
