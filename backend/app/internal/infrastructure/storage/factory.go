// Package storage implements storage adapters for blob files.
package storage

import (
	"context"
	"io"
	"strings"
	"sync"

	"github.com/rs/zerolog"

	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/contracts"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/shared/config"
)

// avisoDisco registra una sola vez que no hay cubo y se usa disco.
var avisoDisco sync.Once

// NewAlmacenArchivos creates blob storage from configuration.
// A non-empty EVIDENCIAS_BUCKET selects the private bucket. Otherwise files
// stay on local disk and that choice is logged once.
func NewAlmacenArchivos(cfg *config.Config, logger zerolog.Logger) contracts.IAlmacenArchivos {
	return newAlmacen(cfg.Evidencias.Bucket, cfg.Evidencias.Dir, logger, NewS3Storage)
}

func newAlmacen(
	bucket, dir string,
	logger zerolog.Logger,
	abrirS3 func(context.Context, string) (contracts.IAlmacenArchivos, error),
) contracts.IAlmacenArchivos {
	bucket = strings.TrimSpace(bucket)
	if bucket == "" {
		avisoDisco.Do(func() {
			logger.Info().Str("dir", dir).Msg("evidencias: EVIDENCIAS_BUCKET no está definido; los archivos quedan en disco")
		})
		return NewDiscoStorage(dir)
	}

	s3Store, err := abrirS3(context.Background(), bucket)
	if err != nil {
		logger.Error().Err(err).Msg("evidencias: no se pudo abrir el cubo privado")
		return almacenNoDisponible{err: err}
	}
	logger.Info().Str("cubo", bucket).Msg("evidencias: cubo privado")
	return s3Store
}

// almacenNoDisponible refuses writes when a bucket was configured but the
// client could not be opened. It does not fall back to disk: the API would
// otherwise report s3 while the bytes stayed local.
type almacenNoDisponible struct {
	err error
}

func (a almacenNoDisponible) Put(context.Context, string, io.Reader, string) (string, error) {
	return "", a.err
}

func (a almacenNoDisponible) Open(context.Context, string) (io.ReadCloser, error) {
	return nil, a.err
}
