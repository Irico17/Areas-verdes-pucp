// Package database provides GORM database connections.
package database

import (
	"fmt"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/shared/config"
)

// NewConnection opens a PostgreSQL connection through GORM.
func NewConnection(cfg *config.Config) (*gorm.DB, error) {
	database := cfg.Database
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s search_path=%s sslmode=%s",
		database.Host, database.Port, database.User, database.Password,
		database.Name, database.Schema, database.SSLMode,
	)
	return gorm.Open(postgres.Open(dsn), &gorm.Config{})
}
