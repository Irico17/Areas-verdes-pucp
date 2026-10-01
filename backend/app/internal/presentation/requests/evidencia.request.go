// Package requests defines HTTP request payloads and query bindings.
package requests

import (
	"io"
	"strconv"
	"strings"

	domainErrors "github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/errors"
)

const topeBytes = 8 << 20

// PuntoForm parses and validates optional latitude and longitude strings.
// Both must be provided together or both omitted.
func PuntoForm(latRaw, lonRaw string) (*float64, *float64, error) {
	latRaw = strings.TrimSpace(latRaw)
	lonRaw = strings.TrimSpace(lonRaw)
	if latRaw == "" && lonRaw == "" {
		return nil, nil, nil
	}
	if latRaw == "" || lonRaw == "" {
		return nil, nil, domainErrors.InputError{Reason: "lat y lon van juntos"}
	}
	lat, err := strconv.ParseFloat(latRaw, 64)
	if err != nil {
		return nil, nil, domainErrors.InputError{Reason: "lat no es un número"}
	}
	lon, err := strconv.ParseFloat(lonRaw, 64)
	if err != nil {
		return nil, nil, domainErrors.InputError{Reason: "lon no es un número"}
	}
	return &lat, &lon, nil
}

// LeerArchivo reads up to 8 MB + 1 byte from r, returning an error if it exceeds the limit.
func LeerArchivo(r io.Reader) ([]byte, error) {
	b, err := io.ReadAll(io.LimitReader(r, topeBytes+1))
	if err != nil {
		return nil, err
	}
	if len(b) > topeBytes {
		return nil, domainErrors.InputError{Reason: "el archivo supera 8 MB"}
	}
	return b, nil
}
