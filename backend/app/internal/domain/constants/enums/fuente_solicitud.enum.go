// Package enums defines domain enumeration types.
package enums

// FuenteSolicitud represents the channel/source of a service request.
type FuenteSolicitud string

const (
	// FuenteCenturia is for Centuria requests.
	FuenteCenturia FuenteSolicitud = "centuria"
	// FuenteOSG is for OSG requests.
	FuenteOSG FuenteSolicitud = "osg"
	// FuenteCorreo is for email requests.
	FuenteCorreo FuenteSolicitud = "correo"
	// FuenteInterna is for internal requests.
	FuenteInterna FuenteSolicitud = "interna"
)

// String returns the string representation.
func (f FuenteSolicitud) String() string {
	return string(f)
}

// EsFuenteSolicitudValida checks if the value is a recognized source.
func EsFuenteSolicitudValida(v string) bool {
	switch FuenteSolicitud(v) {
	case FuenteCenturia, FuenteOSG, FuenteCorreo, FuenteInterna:
		return true
	default:
		return false
	}
}
