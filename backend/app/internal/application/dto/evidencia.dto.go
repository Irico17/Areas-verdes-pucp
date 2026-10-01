// Package dto defines data transfer objects for the application layer.
package dto

// EvidenciaDTO represents an evidence record in HTTP responses.
type EvidenciaDTO struct {
	ID          string `json:"id"`
	ActividadID string `json:"actividad_id,omitempty"`
	Nombre      string `json:"nombre"`
	Mime        string `json:"mime"`
	Bytes       int    `json:"bytes"`
	Nota        string `json:"nota,omitempty"`
	CreatedAt   string `json:"created_at"`
}

// ListarEvidenciasResponseDTO is the JSON response for GET /evidencias.
type ListarEvidenciasResponseDTO struct {
	Evidencias []EvidenciaDTO `json:"evidencias"`
}

// SubirEvidenciaResponseDTO is the JSON response for POST /evidencias.
type SubirEvidenciaResponseDTO struct {
	ID          string `json:"id"`
	Idempotente bool   `json:"idempotente"`
}

// SubirEvidenciaDTO represents the parsed application input for evidence creation.
type SubirEvidenciaDTO struct {
	ID          string
	ActividadID string
	Nombre      string
	Nota        string
	SHA256      string
	Lat         *float64
	Lon         *float64
	Exif        []byte
	Rol         string
	CapatazID   string
	UsuarioID   int64
	OrdenID     string
	Contenido   []byte
}
