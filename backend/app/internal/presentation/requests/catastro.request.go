package requests

import "encoding/json"

// CrearZonaSupervisionRequest binds input JSON for creating a supervision zone.
type CrearZonaSupervisionRequest struct {
	Codigo  string   `json:"codigo"`
	Nombre  string   `json:"nombre"`
	AreaM2  *float64 `json:"area_m2"`
	GeoJSON string   `json:"geojson"`
}

// ActualizarZonaSupervisionRequest binds the PWA payload (geom object) and the geojson string used by POST.
type ActualizarZonaSupervisionRequest struct {
	Codigo  string          `json:"codigo"`
	Nombre  string          `json:"nombre"`
	AreaM2  *float64        `json:"area_m2"`
	GeoJSON string          `json:"geojson"`
	Geom    json.RawMessage `json:"geom"`
}

// CrearCuadrillaRequest binds input JSON for creating a work team.
type CrearCuadrillaRequest struct {
	ID     string `json:"id"`
	Nombre string `json:"nombre_ficticio"`
	Turno  string `json:"turno"`
}

// CrearLugarRequest binds input JSON for creating a place.
type CrearLugarRequest struct {
	Nombre string  `json:"nombre"`
	Lat    float64 `json:"lat"`
	Lon    float64 `json:"lon"`
	ZonaID *int64  `json:"zona_supervision_id"`
}

// CrearEspecieRequest binds input JSON for creating a botanical species.
type CrearEspecieRequest struct {
	Cientifico string `json:"nombre_cientifico"`
	Comun      string `json:"nombre_comun"`
}

// CrearEjemplarRequest binds input JSON for creating a flora specimen.
type CrearEjemplarRequest struct {
	NumeroOrigen       *int     `json:"numero_origen,omitempty"`
	Codigo             string   `json:"codigo"`
	EspecieID          *int64   `json:"especie_id,omitempty"`
	NombreComun        string   `json:"nombre_comun"`
	TipoVegetacion     string   `json:"tipo_vegetacion"`
	Cantidad           int      `json:"cantidad"`
	UbicacionLugarID   *int64   `json:"ubicacion_lugar_id,omitempty"`
	Referencia         string   `json:"referencia"`
	Lat                *float64 `json:"lat,omitempty"`
	Lon                *float64 `json:"lon,omitempty"`
	ObservacionFen2026 string   `json:"observacion_fen_2026"`
}

// RecodificarRequest binds input JSON for recodifying a specimen.
type RecodificarRequest struct {
	Codigo string `json:"codigo_nuevo"`
}

// ActualizarEjemplarRequest binds a partial edit of a flora specimen.
// A key present as null clears that field. An absent key is left unchanged.
type ActualizarEjemplarRequest struct {
	Codigo           *string  `json:"codigo,omitempty"`
	EspecieID        *int64   `json:"especie_id,omitempty"`
	Salud            *string  `json:"salud,omitempty"`
	Lat              *float64 `json:"lat,omitempty"`
	Lon              *float64 `json:"lon,omitempty"`
	UbicacionLugarID *int64   `json:"ubicacion_lugar_id,omitempty"`
	SectorCuartelID  *int64   `json:"sector_cuartel_id,omitempty"`
	presentes        map[string]json.RawMessage
}

// UnmarshalJSON records which keys arrived so null can clear a column.
func (r *ActualizarEjemplarRequest) UnmarshalJSON(data []byte) error {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	r.presentes = raw
	if err := asignarCampo(raw, "codigo", &r.Codigo); err != nil {
		return err
	}
	if err := asignarCampo(raw, "especie_id", &r.EspecieID); err != nil {
		return err
	}
	if err := asignarCampo(raw, "salud", &r.Salud); err != nil {
		return err
	}
	if err := asignarCampo(raw, "lat", &r.Lat); err != nil {
		return err
	}
	if err := asignarCampo(raw, "lon", &r.Lon); err != nil {
		return err
	}
	if err := asignarCampo(raw, "ubicacion_lugar_id", &r.UbicacionLugarID); err != nil {
		return err
	}
	return asignarCampo(raw, "sector_cuartel_id", &r.SectorCuartelID)
}

func asignarCampo[T any](raw map[string]json.RawMessage, clave string, dest **T) error {
	v, ok := raw[clave]
	if !ok || string(v) == "null" {
		return nil
	}
	var valor T
	if err := json.Unmarshal(v, &valor); err != nil {
		return err
	}
	*dest = &valor
	return nil
}

// Presente reports whether the JSON object included the key.
func (r ActualizarEjemplarRequest) Presente(clave string) bool {
	if r.presentes == nil {
		return false
	}
	_, ok := r.presentes[clave]
	return ok
}
