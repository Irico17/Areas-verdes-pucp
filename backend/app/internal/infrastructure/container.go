// Package infrastructure contains adapters for external services.
package infrastructure

import "go.uber.org/dig"

// RegisterContainer registers infrastructure-layer dependencies.
func RegisterContainer(_ *dig.Container) error { return nil }
