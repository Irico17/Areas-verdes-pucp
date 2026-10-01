// Package storage implements storage adapters for blob files.
package storage

import (
	"context"

	"github.com/rs/zerolog"

	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/contracts"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/shared/config"
)

// NewAlmacenArchivos creates a blob storage instance based on configuration.
// If Bucket is set, it attempts to initialize S3. If S3 fails or Bucket is empty,
// it falls back to local disk storage.
func NewAlmacenArchivos(cfg *config.Config, logger zerolog.Logger) contracts.IAlmacenArchivos {
	bucket := cfg.Evidencias.Bucket
	dir := cfg.Evidencias.Dir

	if bucket != "" {
		s3Store, err := NewS3Storage(context.Background(), bucket)
		if err != nil {
			logger.Warn().Err(err).Msgf("evidencias: S3 no disponible; se usa disco (%s)", dir)
			return NewDiscoStorage(dir)
		}
		logger.Info().Msgf("evidencias: cubo %s", bucket)
		return s3Store
	}

	return NewDiscoStorage(dir)
}
