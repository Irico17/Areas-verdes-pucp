package blobs

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// S3 guarda evidencias en un cubo. La referencia es s3://cubo/clave.
type S3 struct {
	Bucket string
	Client *s3.Client
}

func NewS3(ctx context.Context, bucket string) (S3, error) {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return S3{}, err
	}
	return S3{Bucket: bucket, Client: s3.NewFromConfig(cfg)}, nil
}

func (s S3) Put(ctx context.Context, name string, r io.Reader, mime string) (string, error) {
	key := strings.TrimPrefix(filepathBase(name), "/")
	_, err := s.Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(s.Bucket),
		Key:         aws.String(key),
		Body:        r,
		ContentType: aws.String(mime),
	})
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("s3://%s/%s", s.Bucket, key), nil
}

func (s S3) Open(ctx context.Context, ref string) (io.ReadCloser, error) {
	bucket, key, ok := parseS3(ref)
	if !ok || bucket != s.Bucket {
		return nil, fmt.Errorf("referencia de evidencia inválida")
	}
	out, err := s.Client.GetObject(ctx, &s3.GetObjectInput{
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
