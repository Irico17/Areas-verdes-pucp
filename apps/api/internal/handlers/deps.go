package handlers

import "gorm.io/gorm"

// Deps es el cableado que recibe cada registrar(r, deps).
// Un frente nuevo no añade campos: arma sus stores con DB y los directorios ya definidos.
type Deps struct {
	DB               *gorm.DB
	OpenAPIPath      string
	EdificiosPath    string
	ReservasPath     string
	FotosDir         string
	EvidenciasDir    string
	EvidenciasBucket string
}
