package blobs

import (
	"context"
	"io"
	"log"
)

// Store guarda y abre un archivo de evidencia.
type Store interface {
	Put(ctx context.Context, name string, r io.Reader, mime string) (string, error)
	Open(ctx context.Context, ref string) (io.ReadCloser, error)
}

// Open elige S3 si hay cubo y las credenciales cargan; si no, disco local.
func Open(dir, bucket string) Store {
	if bucket != "" {
		client, err := NewS3(context.Background(), bucket)
		if err != nil {
			log.Printf("evidencias: S3 no disponible (%v); se usa disco", err)
			return Disk{Dir: dir}
		}
		log.Printf("evidencias: cubo %s", bucket)
		return client
	}
	return Disk{Dir: dir}
}
