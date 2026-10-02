package storage

import (
	"bytes"
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/rs/zerolog"

	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/contracts"
)

type dobleS3 struct {
	puts []*s3.PutObjectInput
	gets []*s3.GetObjectInput
	body []byte
}

func (d *dobleS3) PutObject(_ context.Context, params *s3.PutObjectInput, _ ...func(*s3.Options)) (*s3.PutObjectOutput, error) {
	if params.Body != nil {
		b, err := io.ReadAll(params.Body)
		if err != nil {
			return nil, err
		}
		d.body = b
	}
	d.puts = append(d.puts, params)
	return &s3.PutObjectOutput{}, nil
}

func (d *dobleS3) GetObject(_ context.Context, params *s3.GetObjectInput, _ ...func(*s3.Options)) (*s3.GetObjectOutput, error) {
	d.gets = append(d.gets, params)
	return &s3.GetObjectOutput{Body: io.NopCloser(bytes.NewReader(d.body))}, nil
}

func TestConBucketLlamaPutObject(t *testing.T) {
	doble := &dobleS3{}
	var cuboAbierto string
	store := newAlmacen(" cubo-de-prueba ", t.TempDir(), zerolog.Nop(), func(_ context.Context, bucket string) (contracts.IAlmacenArchivos, error) {
		cuboAbierto = bucket
		return nuevoS3(bucket, doble), nil
	})
	if cuboAbierto != "cubo-de-prueba" {
		t.Fatalf("cubo abierto %q", cuboAbierto)
	}

	const cuerpo = "bytes-de-la-evidencia"
	ref, err := store.Put(context.Background(), "dir/foto.jpg", bytes.NewBufferString(cuerpo), "image/jpeg")
	if err != nil {
		t.Fatal(err)
	}
	if len(doble.puts) != 1 {
		t.Fatalf("PutObject se llamó %d veces", len(doble.puts))
	}
	in := doble.puts[0]
	if aws.ToString(in.Bucket) != "cubo-de-prueba" || aws.ToString(in.Key) != "foto.jpg" {
		t.Fatalf("PutObject bucket=%q key=%q", aws.ToString(in.Bucket), aws.ToString(in.Key))
	}
	if aws.ToString(in.ContentType) != "image/jpeg" {
		t.Fatalf("content-type %q", aws.ToString(in.ContentType))
	}
	if in.ACL != "" {
		t.Fatalf("el cubo es privado; no debe pedirse ACL %q", in.ACL)
	}
	if string(doble.body) != cuerpo {
		t.Fatalf("cuerpo enviado %q", doble.body)
	}
	if ref != "foto.jpg" || strings.Contains(ref, "://") || strings.Contains(ref, "amazonaws") {
		t.Fatalf("la base debe guardar la clave, no una URL: %q", ref)
	}

	rc, err := store.Open(context.Background(), ref)
	if err != nil {
		t.Fatal(err)
	}
	defer rc.Close()
	got, _ := io.ReadAll(rc)
	if string(got) != cuerpo {
		t.Fatalf("lectura %q", got)
	}
	if len(doble.gets) != 1 || aws.ToString(doble.gets[0].Key) != "foto.jpg" {
		t.Fatalf("GetObject %+v", doble.gets)
	}

	legacy, err := store.Open(context.Background(), "s3://cubo-de-prueba/foto.jpg")
	if err != nil {
		t.Fatal(err)
	}
	_ = legacy.Close()

	if _, err := store.Open(context.Background(), "https://cubo-de-prueba.s3.amazonaws.com/foto.jpg"); err == nil {
		t.Fatal("una URL pública no es una clave")
	}
	if _, err := store.Open(context.Background(), "s3://otro-cubo/foto.jpg"); err == nil {
		t.Fatal("un cubo distinto no se abre")
	}
}

func TestSinBucketEscribeEnDirectorioTemporal(t *testing.T) {
	avisoDisco = sync.Once{}
	dir := t.TempDir()
	var buf bytes.Buffer
	logger := zerolog.New(&buf)
	abrioS3 := false
	store := newAlmacen("   ", dir, logger, func(context.Context, string) (contracts.IAlmacenArchivos, error) {
		abrioS3 = true
		return nil, nil
	})
	_ = newAlmacen("", dir, logger, func(context.Context, string) (contracts.IAlmacenArchivos, error) {
		abrioS3 = true
		return nil, nil
	})
	if abrioS3 {
		t.Fatal("sin EVIDENCIAS_BUCKET no se abre el cliente S3")
	}

	const cuerpo = "en-disco"
	ref, err := store.Put(context.Background(), "subdir/nota.pdf", bytes.NewBufferString(cuerpo), "application/pdf")
	if err != nil {
		t.Fatal(err)
	}
	esperado := filepath.Join(dir, "nota.pdf")
	if ref != esperado {
		t.Fatalf("ruta %q, esperaba %q", ref, esperado)
	}
	data, err := os.ReadFile(esperado)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != cuerpo {
		t.Fatalf("archivo %q", data)
	}
	if !strings.HasPrefix(ref, dir) {
		t.Fatalf("el archivo salió del directorio temporal: %s", ref)
	}

	log := buf.String()
	if strings.Count(log, "EVIDENCIAS_BUCKET no está definido") != 1 {
		t.Fatalf("el aviso de disco debe salir una vez, log: %s", log)
	}
}

func TestConBucketSinClienteNoEscribeEnDisco(t *testing.T) {
	dir := t.TempDir()
	store := newAlmacen("cubo-de-prueba", dir, zerolog.Nop(), func(context.Context, string) (contracts.IAlmacenArchivos, error) {
		return nil, errors.New("sin credenciales")
	})
	_, err := store.Put(context.Background(), "foto.jpg", bytes.NewBufferString("no"), "image/jpeg")
	if err == nil {
		t.Fatal("un cubo que no abre no debe guardar")
	}
	entries, readErr := os.ReadDir(dir)
	if readErr != nil {
		t.Fatal(readErr)
	}
	if len(entries) != 0 {
		t.Fatalf("no debe haber archivos en disco: %d", len(entries))
	}
}
