// Package database provides GORM database connections.
package database

import (
	"fmt"
	"net/http"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	appErrors "github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/errors"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/shared/config"
	"github.com/rs/zerolog/log"
)

// NewDatabaseConnection establishes a connection to PostgreSQL through GORM with pool settings.
func NewDatabaseConnection(cfg *config.Config) (*gorm.DB, error) {
	database := cfg.Database
	dsn := database.URL
	if dsn == "" {
		dsn = fmt.Sprintf(
			"host=%s port=%s user=%s password=%s dbname=%s search_path=%s sslmode=%s",
			database.Host, database.Port, database.User, database.Password,
			database.Name, database.Schema, database.SSLMode,
		)
	}

	log.Info().
		Str("host", database.Host).
		Str("port", database.Port).
		Str("dbname", database.Name).
		Str("schema", database.Schema).
		Str("ssl", database.SSLMode).
		Msg("Connecting to database")

	gdb, err := gorm.Open(postgres.New(postgres.Config{
		DSN:                  dsn,
		PreferSimpleProtocol: true,
	}), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn),
	})
	if err != nil {
		return nil, appErrors.NewApplicationError(
			appErrors.DBDatabaseConnection,
			http.StatusServiceUnavailable,
			err,
		)
	}

	sqlDB, err := gdb.DB()
	if err != nil {
		return nil, appErrors.NewApplicationError(
			appErrors.DBDatabaseConnection,
			http.StatusServiceUnavailable,
			err,
		)
	}

	if err = sqlDB.Ping(); err != nil {
		return nil, appErrors.NewApplicationError(
			appErrors.DBDatabaseConnection,
			http.StatusServiceUnavailable,
			err,
		)
	}

	maxOpen := database.MaxOpenConns
	if maxOpen <= 0 {
		maxOpen = 10
	}
	maxIdle := database.MaxIdleConns
	if maxIdle <= 0 {
		maxIdle = 4
	}
	lifetime := database.ConnMaxLifetime
	if lifetime <= 0 {
		lifetime = 30 * time.Minute
	}

	sqlDB.SetMaxOpenConns(maxOpen)
	sqlDB.SetMaxIdleConns(maxIdle)
	sqlDB.SetConnMaxLifetime(lifetime)

	log.Info().Msg("Database connected successfully")

	return gdb, nil
}

// NewConnection is an alias for NewDatabaseConnection for backward compatibility.
func NewConnection(cfg *config.Config) (*gorm.DB, error) {
	return NewDatabaseConnection(cfg)
}
