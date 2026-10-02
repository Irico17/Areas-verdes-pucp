// Package ioc builds the dependency injection container.
package ioc

import (
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"go.uber.org/dig"

	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/infrastructure"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/persistence"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/presentation"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/shared/config"
)

// BuildContainer registers all application layers in a single container.
func BuildContainer() (*dig.Container, error) {
	container := dig.New()

	// Config
	if err := container.Provide(func() (*config.Config, error) {
		cfg := config.GetConfig()
		if err := cfg.Validar(); err != nil {
			return nil, err
		}
		return cfg, nil
	}); err != nil {
		return nil, err
	}

	// Logger: provide zerolog.Logger from the global logger
	if err := container.Provide(func() zerolog.Logger {
		return log.Logger
	}); err != nil {
		return nil, err
	}

	// Gin
	if err := container.Provide(func(cfg *config.Config) (*gin.Engine, error) {
		gin.SetMode(cfg.Server.GinMode)
		engine := gin.New()
		if err := engine.SetTrustedProxies(cfg.Server.TrustedProxies); err != nil {
			return nil, err
		}
		return engine, nil
	}); err != nil {
		return nil, err
	}

	// Layers
	registrars := []func(*dig.Container) error{
		application.RegisterContainer,
		infrastructure.RegisterContainer,
		persistence.RegisterContainer,
		presentation.RegisterContainer,
	}
	for _, register := range registrars {
		if err := register(container); err != nil {
			return nil, err
		}
	}

	return container, nil
}
