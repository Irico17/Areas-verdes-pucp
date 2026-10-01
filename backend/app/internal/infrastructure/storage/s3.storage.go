// Package storage implements storage adapters for blob files.
package storage

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"

	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/contracts"
)

// s3Storage implements contracts.IAlmacenArchivos using AWS S3.
type s3Storage struct {
	bucket string
	client *s3.Client
}

// NewS3Storage creates an S3 blob storage client for the given bucket.
func NewS3Storage(ctx context.Context, bucket string) (contracts.IAlmacenArchivos, error) {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, err
	}
	return &s3Storage{
		bucket: bucket,
		client: s3.NewFromConfig(cfg),
	}, nil
}

// Put uploads an object to S3 and returns the canonical reference s3://bucket/key.
func (s *s3Storage) Put(ctx context.Context, name string, r io.Reader, mime string) (string, error) {
	key := strings.TrimPrefix(filepathBase(name), "/")
	_, err := s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(s.bucket),
		Key:         aws.String(key),
		Body:        r,
		ContentType: aws.String(mime),
	})
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("s3://%s/%s", s.bucket, key), nil
}

// Open retrieves an object from S3.
func (s *s3Storage) Open(ctx context.Context, ref string) (io.ReadCloser, error) {
	bucket, key, ok := parseS3(ref)
	if !ok || bucket != s.bucket {
		return nil, fmt.Errorf("referencia de evidencia inválida")
	}
	out, err := s.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, err
	}
	return out.Body, nil
}

func parseS3(ref string) (string, string, bool) {
	const prefix = "s3://"
	if !strings.HasPrefix(ref, prefix) {
		return "", "", false
	}
	rest := strings.TrimPrefix(ref, prefix)
	bucket, key, ok := strings.Cut(rest, "/")
	if !ok || bucket == "" || key == "" {
		return "", "", false
	}
	return bucket, key, true
}

func filepathBase(name string) string {
	name = strings.ReplaceAll(name, "\\", "/")
	if i := strings.LastIndex(name, "/"); i >= 0 {
		return name[i+1:]
	}
	return name
}
