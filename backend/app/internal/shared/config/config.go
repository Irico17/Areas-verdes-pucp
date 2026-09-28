// Package config loads application configuration from environment variables.
package config

import (
	"os"

	"github.com/joho/godotenv"
)

// Config contains all application configuration.
type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
}

// ServerConfig contains HTTP server settings.
type ServerConfig struct{ Port, GinMode string }

// DatabaseConfig contains PostgreSQL connection settings.
type DatabaseConfig struct{ Host, Port, User, Password, Name, Schema, SSLMode string }

// New builds configuration from the process environment with local defaults.
func New() *Config {
	_ = godotenv.Load()
	return &Config{
		Server: ServerConfig{
			Port:    valueOrDefault("SERVER_PORT", "8080"),
			GinMode: valueOrDefault("SERVER_GIN_MODE", "debug"),
		},
		Database: DatabaseConfig{
			Host:     valueOrDefault("DATABASE_HOST", "localhost"),
			Port:     valueOrDefault("DATABASE_PORT", "5432"),
			User:     valueOrDefault("DATABASE_USER", "areasverdes"),
			Password: valueOrDefault("DATABASE_PASSWORD", "areasverdes"),
			Name:     valueOrDefault("DATABASE_NAME", "areasverdes"),
			Schema:   valueOrDefault("DATABASE_SCHEMA", "public"),
			SSLMode:  valueOrDefault("DATABASE_SSL_MODE", "disable"),
		},
	}
}

func valueOrDefault(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
