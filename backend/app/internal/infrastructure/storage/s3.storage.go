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

// clienteS3 es el subconjunto del cliente de AWS que usa el adaptador.
// Producción pasa *s3.Client; los tests pasan un doble.
type clienteS3 interface {
	PutObject(ctx context.Context, params *s3.PutObjectInput, optFns ...func(*s3.Options)) (*s3.PutObjectOutput, error)
	GetObject(ctx context.Context, params *s3.GetObjectInput, optFns ...func(*s3.Options)) (*s3.GetObjectOutput, error)
}

// s3Storage implements contracts.IAlmacenArchivos using a private S3 bucket.
type s3Storage struct {
	bucket string
	client clienteS3
}

// NewS3Storage creates an S3 blob storage client for the given private bucket.
// Credentials come from the SDK default chain (instance role, or keys injected
// outside this repository). This function does not read secrets from the repo.
func NewS3Storage(ctx context.Context, bucket string) (contracts.IAlmacenArchivos, error) {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, err
	}
	return nuevoS3(bucket, s3.NewFromConfig(cfg)), nil
}

func nuevoS3(bucket string, client clienteS3) contracts.IAlmacenArchivos {
	return &s3Storage{
		bucket: strings.TrimSpace(bucket),
		client: client,
	}
}

// Put uploads the object and returns the object key. The database stores that
// key, not a public URL. The bucket stays private: no public ACL is set.
func (s *s3Storage) Put(ctx context.Context, name string, r io.Reader, mime string) (string, error) {
	key, err := claveObjeto(name)
	if err != nil {
		return "", err
	}
	_, err = s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(s.bucket),
		Key:         aws.String(key),
		Body:        r,
		ContentType: aws.String(mime),
	})
	if err != nil {
		return "", err
	}
	return key, nil
}

// Open reads an object by the stored key. A legacy s3://bucket/key reference
// still opens when the bucket matches. Public URLs are rejected.
func (s *s3Storage) Open(ctx context.Context, ref string) (io.ReadCloser, error) {
	key, err := claveDesdeReferencia(s.bucket, ref)
	if err != nil {
		return nil, err
	}
	out, err := s.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, err
	}
	return out.Body, nil
}

func claveObjeto(name string) (string, error) {
	key := strings.TrimPrefix(filepathBase(name), "/")
	if key == "" || key == "." || key == ".." || strings.Contains(key, "..") {
		return "", fmt.Errorf("nombre de evidencia inválido")
	}
	return key, nil
}

func claveDesdeReferencia(bucket, ref string) (string, error) {
	ref = strings.TrimSpace(ref)
	if ref == "" {
		return "", fmt.Errorf("referencia de evidencia inválida")
	}
	if strings.Contains(ref, "://") {
		if !strings.HasPrefix(ref, "s3://") {
			return "", fmt.Errorf("referencia de evidencia inválida")
		}
		gotBucket, key, ok := parseS3(ref)
		if !ok || gotBucket != bucket || strings.Contains(key, "..") {
			return "", fmt.Errorf("referencia de evidencia inválida")
		}
		return key, nil
	}
	if strings.HasPrefix(ref, "/") || strings.Contains(ref, `\`) || strings.Contains(ref, "..") {
		return "", fmt.Errorf("referencia de evidencia inválida")
	}
	return ref, nil
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
	name = strings.ReplaceAll(name, `\`, "/")
	if i := strings.LastIndex(name, "/"); i >= 0 {
		return name[i+1:]
	}
	return name
}
