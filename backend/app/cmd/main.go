// Package main starts the Areas Verdes HTTP API.
package main

import (
	"fmt"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"

	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/cmd/ioc"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/docs"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/presentation/routes"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/shared/config"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/shared/logger"
)

// @title           Areas Verdes API $SWAGGER_ENV
// @version         1.0.0
// @description     API for the Areas Verdes service
// @termsOfService  http://swagger.io/terms/
// @contact.name    Group_12
// @contact.email   support@example.com (update)
// @license.name    Proprietary
// @license.url     https://example.com
// @BasePath        /areas-verdes
func main() {
	cfg := config.GetConfig()
	if err := cfg.Validar(); err != nil {
		fmt.Fprintf(os.Stderr, "ERROR de configuración: %v\n", err)
		os.Exit(1)
	}
	logger.InitLogger(cfg.Server.LogLevel, cfg.Server.LogFormat)
	log.Info().
		Str("app_env", cfg.AppEnv).
		Str("gin_mode", cfg.Server.GinMode).
		Str("log_level", cfg.Server.LogLevel).
		Bool("swagger", cfg.Swagger.Enabled).
		Msg("configuración cargada")

	if cfg.Swagger.Host != "" {
		docs.SwaggerInfo.Host = cfg.Swagger.Host
	}

	container, err := ioc.BuildContainer()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to build dependency container")
	}

	err = container.Invoke(func(router *routes.Router, engine *gin.Engine, cfg *config.Config) error {
		router.Setup()
		address := fmt.Sprintf(":%s", cfg.Server.Port)
		log.Info().Str("port", address).Msg("Areas Verdes server listening")
		return engine.Run(address)
	})
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to start application")
	}
}
