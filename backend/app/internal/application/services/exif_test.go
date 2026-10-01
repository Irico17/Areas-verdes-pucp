package services

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"

	domainErrors "github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/errors"
)

func TestFiltrarExifDejaSoloLaListaBlanca(t *testing.T) {
	svc := NewExifService()
	raw := json.RawMessage(`{"fecha":"2026:09:25 10:00:00","lat":-12.07,"lon":-77.08,"orientación":1,"modelo":"no","gps":{"a":1}}`)
	got, err := svc.FiltrarExif(raw)
	if err != nil {
		t.Fatal(err)
	}
	s, _ := got.(string)
	if strings.Contains(s, "modelo") || strings.Contains(s, "gps") {
		t.Fatalf("se coló un campo: %s", s)
	}
	for _, clave := range []string{"fecha", "lat", "lon", "orientacion"} {
		if !strings.Contains(s, clave) {
			t.Fatalf("falta %s en %s", clave, s)
		}
	}
}

func TestFiltrarExifRechazaElTope(t *testing.T) {
	svc := NewExifService()
	raw := json.RawMessage(`{"fecha":"` + strings.Repeat("a", topeExif) + `"}`)
	_, err := svc.FiltrarExif(raw)
	var input domainErrors.InputError
	if !errors.As(err, &input) || !strings.Contains(input.Reason, "tamaño") {
		t.Fatalf("esperaba tope, obtuvo %v", err)
	}
}

func TestFiltrarExifRechazaJSONInvalido(t *testing.T) {
	svc := NewExifService()
	_, err := svc.FiltrarExif(json.RawMessage(`esto no es json`))
	var input domainErrors.InputError
	if !errors.As(err, &input) || !strings.Contains(input.Reason, "objeto JSON") {
		t.Fatalf("esperaba error json inválido, obtuvo %v", err)
	}
}

func TestFiltrarExifNuloOVacioRetornaNil(t *testing.T) {
	svc := NewExifService()
	got, err := svc.FiltrarExif(nil)
	if err != nil || got != nil {
		t.Fatalf("esperaba nil, obtuvo %v, %v", got, err)
	}
	got, err = svc.FiltrarExif(json.RawMessage(`null`))
	if err != nil || got != nil {
		t.Fatalf("esperaba nil, obtuvo %v, %v", got, err)
	}
	got, err = svc.FiltrarExif(json.RawMessage(`   `))
	if err != nil || got != nil {
		t.Fatalf("esperaba nil, obtuvo %v, %v", got, err)
	}
}

func TestMimeRealFormatosValidos(t *testing.T) {
	svc := NewExifService()

	// JPEG
	jpeg := []byte{0xFF, 0xD8, 0xFF, 0xD9}
	mime, ext, ok := svc.MimeReal(jpeg)
	if !ok || mime != "image/jpeg" || ext != ".jpg" {
		t.Fatalf("JPEG: esperado image/jpeg, .jpg, true; obtenido %s, %s, %v", mime, ext, ok)
	}

	// PNG
	png := []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}
	mime, ext, ok = svc.MimeReal(png)
	if !ok || mime != "image/png" || ext != ".png" {
		t.Fatalf("PNG: esperado image/png, .png, true; obtenido %s, %s, %v", mime, ext, ok)
	}

	// PDF
	pdf := []byte("%PDF-1.4\n")
	mime, ext, ok = svc.MimeReal(pdf)
	if !ok || mime != "application/pdf" || ext != ".pdf" {
		t.Fatalf("PDF: esperado application/pdf, .pdf, true; obtenido %s, %s, %v", mime, ext, ok)
	}

	// WebP (RIFF....WEBP)
	webp := []byte("RIFF1234WEBPVP8 ")
	mime, ext, ok = svc.MimeReal(webp)
	if !ok || mime != "image/webp" || ext != ".webp" {
		t.Fatalf("WebP: esperado image/webp, .webp, true; obtenido %s, %s, %v", mime, ext, ok)
	}

	// Invalido
	invalido := []byte("texto plano no soportado")
	_, _, ok = svc.MimeReal(invalido)
	if ok {
		t.Fatal("esperaba ok=false para archivo no reconocido")
	}
}
