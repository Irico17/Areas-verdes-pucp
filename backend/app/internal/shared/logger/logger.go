// Package logger initializes the service logger.
package logger

import (
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// InitLogger initializes the global logger based on the configured mode.
func InitLogger(mode string) zerolog.Logger {
	if mode == "debug" {
		log.Logger = zerolog.New(zerolog.ConsoleWriter{Out: os.Stdout}).
			With().
			Timestamp().
			Caller().
			Logger()
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
		return log.Logger
	}

	log.Logger = zerolog.New(os.Stdout).
		With().
		Timestamp().
		Caller().
		Logger()
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	return log.Logger
}
