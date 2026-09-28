package entities

// CodigoHistorico preserves the previous and new codes of a specimen when recodified.
type CodigoHistorico struct {
	ID             int64  `json:"id"`
	EjemplarID     int64  `json:"ejemplar_id"`
	CodigoAnterior string `json:"codigo_anterior"`
	CodigoNuevo    string `json:"codigo_nuevo"`
}
