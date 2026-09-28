// Package database provides GORM database connections.
package database

import (
	"fmt"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/shared/config"
)

// NewConnection opens a PostgreSQL connection through GORM with pool settings.
func NewConnection(cfg *config.Config) (*gorm.DB, error) {
	database := cfg.Database
	dsn := database.URL
	if dsn == "" {
		dsn = fmt.Sprintf(
			"host=%s port=%s user=%s password=%s dbname=%s search_path=%s sslmode=%s",
			database.Host, database.Port, database.User, database.Password,
			database.Name, database.Schema, database.SSLMode,
		)
	}

	gdb, err := gorm.Open(postgres.New(postgres.Config{
		DSN:                  dsn,
		PreferSimpleProtocol: true,
	}), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn),
	})
	if err != nil {
		return nil, err
	}

	sqlDB, err := gdb.DB()
	if err != nil {
		return nil, err
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

	return gdb, nil
}
