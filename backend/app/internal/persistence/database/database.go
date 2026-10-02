// Package database provides GORM database connections.
package database

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
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

	host, port, dbname, schema, ssl := connectionTarget(database)
	log.Info().
		Str("host", host).
		Str("port", port).
		Str("dbname", dbname).
		Str("schema", schema).
		Str("ssl", ssl).
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

// connectionTarget describes where the pool will connect, without user or password.
// When DATABASE_URL is set it wins over the granular defaults (localhost / areasverdes).
func connectionTarget(database config.DatabaseConfig) (host, port, dbname, schema, ssl string) {
	schema = database.Schema
	ssl = database.SSLMode
	raw := strings.TrimSpace(database.URL)
	if raw == "" {
		return database.Host, database.Port, database.Name, schema, ssl
	}
	if strings.Contains(raw, "://") {
		if u, err := url.Parse(raw); err == nil && u.Hostname() != "" {
			host = u.Hostname()
			port = u.Port()
			if port == "" {
				port = "5432"
			}
			dbname = strings.TrimPrefix(u.Path, "/")
			if q := u.Query().Get("sslmode"); q != "" {
				ssl = q
			}
			if q := u.Query().Get("search_path"); q != "" {
				schema = q
			}
			return host, port, dbname, schema, ssl
		}
	}
	fields := parseKeywordDSN(raw)
	host = firstNonEmpty(fields["host"], "desconocido")
	port = firstNonEmpty(fields["port"], "5432")
	dbname = firstNonEmpty(fields["dbname"], fields["database"])
	if q := fields["sslmode"]; q != "" {
		ssl = q
	}
	if q := fields["search_path"]; q != "" {
		schema = q
	}
	return host, port, dbname, schema, ssl
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}

// parseKeywordDSN reads a libpq keyword DSN. The password value is parsed and discarded by the caller.
func parseKeywordDSN(dsn string) map[string]string {
	out := map[string]string{}
	i := 0
	for i < len(dsn) {
		for i < len(dsn) && (dsn[i] == ' ' || dsn[i] == '\t') {
			i++
		}
		if i >= len(dsn) {
			break
		}
		eq := strings.IndexByte(dsn[i:], '=')
		if eq <= 0 {
			break
		}
		key := strings.ToLower(strings.TrimSpace(dsn[i : i+eq]))
		i += eq + 1
		if i < len(dsn) && dsn[i] == '\'' {
			i++
			var b strings.Builder
			for i < len(dsn) {
				if dsn[i] == '\\' && i+1 < len(dsn) {
					b.WriteByte(dsn[i+1])
					i += 2
					continue
				}
				if dsn[i] == '\'' {
					i++
					break
				}
				b.WriteByte(dsn[i])
				i++
			}
			out[key] = b.String()
			continue
		}
		end := i
		for end < len(dsn) && dsn[end] != ' ' && dsn[end] != '\t' {
			end++
		}
		out[key] = dsn[i:end]
		i = end
	}
	return out
}
