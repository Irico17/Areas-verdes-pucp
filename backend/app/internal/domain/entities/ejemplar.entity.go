package entities

import (
	"strings"

	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/constants/enums"
	domainErrors "github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/errors"
)

// Ejemplar represents an individual flora specimen on campus.
type Ejemplar struct {
	ID                  int64    `json:"id"`
	NumeroOrigen        *int     `json:"numero_origen,omitempty"`
	Codigo              string   `json:"codigo"`
	EspecieID           *int64   `json:"especie_id,omitempty"`
	NombreComun         string   `json:"nombre_comun"`
	TipoVegetacion      string   `json:"tipo_vegetacion"`
	Cantidad            int      `json:"cantidad"`
	UbicacionLugarID    *int64   `json:"ubicacion_lugar_id,omitempty"`
	Referencia          string   `json:"referencia"`
	Lat                 *float64 `json:"lat,omitempty"`
	Lon                 *float64 `json:"lon,omitempty"`
	ObservacionFen2026  string   `json:"observacion_fen_2026"`
	Salud               *string  `json:"salud"`
	SectorCuartelID     *int64   `json:"sector_cuartel_id,omitempty"`
	SectorCuartelNombre string   `json:"sector_cuartel_nombre,omitempty"`
	SectorCuartelClase  string   `json:"sector_cuartel_clase,omitempty"`
	EspecieCientifico   string   `json:"especie_cientifico,omitempty"`
	LugarNombre         string   `json:"lugar_nombre,omitempty"`
	Activo              bool     `json:"activo"`
}

// ParcheEjemplar is a partial edit. A Tiene* flag means the JSON key was present.
// A nil pointer with the flag set clears that column.
type ParcheEjemplar struct {
	Codigo           *string
	TieneCodigo      bool
	EspecieID        *int64
	TieneEspecie     bool
	Salud            *string
	TieneSalud       bool
	Lat              *float64
	Lon              *float64
	TieneLat         bool
	TieneLon         bool
	UbicacionLugarID *int64
	TieneLugar       bool
	SectorCuartelID  *int64
	TieneSector      bool
}

// ValidarSaludEjemplar accepts the closed set used by the ficha, or empty (sin dato).
func ValidarSaludEjemplar(salud string) error {
	switch strings.TrimSpace(salud) {
	case "", "bueno", "regular", "deteriorado":
		return nil
	default:
		return domainErrors.ErrEntrada
	}
}

// Validar checks if the ejemplar fields are valid for creation.
func (e *Ejemplar) Validar() error {
	if e.Cantidad == 0 {
		e.Cantidad = 1
	}
	if err := ValidarCantidad(e.Cantidad); err != nil {
		return err
	}
	if err := enums.ValidarTipoVegetacion(e.TipoVegetacion); err != nil {
		return err
	}
	if err := ValidarReferencia(e.Referencia); err != nil {
		return err
	}
	if e.Lat != nil || e.Lon != nil {
		if e.Lat == nil || e.Lon == nil {
			return domainErrors.ErrEntrada
		}
		if err := ValidarPuntoCampus(*e.Lat, *e.Lon); err != nil {
			return err
		}
	}
	return nil
}
