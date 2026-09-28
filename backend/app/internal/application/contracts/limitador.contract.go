// Package contracts defines application layer contracts and interfaces.
package contracts

// ILimitador defines the rate-limiter interface.
type ILimitador interface {
	Permitir(clave string) bool
}
