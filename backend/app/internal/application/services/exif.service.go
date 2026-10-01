// Package services contains cross-cutting application services.
package services

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strings"
	"unicode/utf8"

	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/contracts"
	domainErrors "github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/errors"
)

const topeExif = 1024

type exifService struct{}

// NewExifService creates a new instance of IExifService.
func NewExifService() contracts.IExifService {
	return &exifService{}
}

// FiltrarExif validates and sanitizes an EXIF JSON object, keeping only allowed fields
// (fecha, lat, lon, orientacion) within size limits.
func (s *exifService) FiltrarExif(raw json.RawMessage) (any, error) {
	trimmed := bytes.TrimSpace(raw)
	if len(trimmed) == 0 || string(trimmed) == "null" {
		return nil, nil
	}
	if len(trimmed) > topeExif {
		return nil, domainErrors.InputError{Reason: "exif supera el tamaño permitido"}
	}
	var obj map[string]json.RawMessage
	if err := json.Unmarshal(trimmed, &obj); err != nil {
		return nil, domainErrors.InputError{Reason: "exif debe ser un objeto JSON"}
	}
	out := map[string]any{}
	if v, ok := obj["fecha"]; ok {
		var strVal string
		if json.Unmarshal(v, &strVal) == nil && strVal != "" && utf8.RuneCountInString(strVal) <= 40 {
			out["fecha"] = strVal
		}
	}
	if n, ok := numeroExif(obj["lat"]); ok && n >= -90 && n <= 90 {
		out["lat"] = n
	}
	if n, ok := numeroExif(obj["lon"]); ok && n >= -180 && n <= 180 {
		out["lon"] = n
	}
	for _, key := range []string{"orientacion", "orientación"} {
		if v, ok := obj[key]; !ok {
			continue
		} else if n, ok := numeroExif(v); ok {
			out["orientacion"] = n
			break
		} else {
			var strVal string
			if json.Unmarshal(v, &strVal) == nil && strVal != "" && utf8.RuneCountInString(strVal) <= 16 {
				out["orientacion"] = strVal
				break
			}
		}
	}
	if len(out) == 0 {
		return nil, nil
	}
	b, err := json.Marshal(out)
	if err != nil {
		return nil, err
	}
	return string(b), nil
}

func numeroExif(raw json.RawMessage) (float64, bool) {
	if len(raw) == 0 {
		return 0, false
	}
	var n float64
	if json.Unmarshal(raw, &n) != nil {
		return 0, false
	}
	return n, true
}

// MimeReal sniffs the actual MIME type of the file bytes and returns (mime, extension, ok).
// Supported formats: JPEG (.jpg), PNG (.png), WebP (.webp), PDF (.pdf).
func (s *exifService) MimeReal(b []byte) (string, string, bool) {
	if len(b) >= 12 && bytes.Equal(b[0:4], []byte("RIFF")) && bytes.Equal(b[8:12], []byte("WEBP")) {
		return "image/webp", ".webp", true
	}
	ct := http.DetectContentType(b)
	switch {
	case strings.HasPrefix(ct, "image/jpeg"):
		return "image/jpeg", ".jpg", true
	case strings.HasPrefix(ct, "image/png"):
		return "image/png", ".png", true
	case strings.HasPrefix(ct, "application/pdf"):
		return "application/pdf", ".pdf", true
	default:
		return "", "", false
	}
}
