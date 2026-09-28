// Package ioc builds the dependency injection container.
package ioc

import (
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"go.uber.org/dig"

	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/infrastructure"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/persistence"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/presentation"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/shared/config"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/shared/logger"
)

// BuildContainer registers all application layers in a single container.
func BuildContainer() (*dig.Container, error) {
	container := dig.New()

	// Config
	if err := container.Provide(config.New); err != nil {
		return nil, err
	}

	// Logger
	if err := container.Provide(func(cfg *config.Config) zerolog.Logger {
		return logger.InitLogger(cfg.Server.GinMode)
	}); err != nil {
		return nil, err
	}

	// Gin
	if err := container.Provide(func(cfg *config.Config) *gin.Engine {
		gin.SetMode(cfg.Server.GinMode)
		return gin.New()
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
