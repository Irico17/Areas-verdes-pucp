package atencion

import "campusverde/api/internal/operacion"

// ValidarRiego exige la zona por identificador de catálogo, no por texto libre.
func ValidarRiego(zonaID, turno string, superficie float64) error {
	switch zonaID {
	case "Z1", "Z2", "Z3", "Z4":
	default:
		return operacion.InputError{Reason: "el riego guarda la zona por identificador Z1–Z4"}
	}
	if turno != "manana" && turno != "tarde" {
		return operacion.InputError{Reason: "el turno es mañana o tarde"}
	}
	if superficie < 0 {
		return operacion.InputError{Reason: "la superficie no puede ser negativa"}
	}
	return nil
}
