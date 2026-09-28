// Package application contains use cases, services, contracts and DTOs.
package application

import "go.uber.org/dig"

// RegisterContainer registers application-layer dependencies.
func RegisterContainer(_ *dig.Container) error { return nil }
