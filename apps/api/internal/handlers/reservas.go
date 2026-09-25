package handlers

import (
	"encoding/json"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
)

// Reservas publica solo la agenda ficticia, sin identificadores de hojas de cálculo.
type Reservas struct {
	Path string
}

type reservaItem struct {
	ID                string `json:"id"`
	Jardin            string `json:"jardin"`
	JardinCodigo      string `json:"jardin_codigo,omitempty"`
	Fecha             string `json:"fecha"`
	Hora              string `json:"hora"`
	Evento            string `json:"evento"`
	UnidadResponsable string `json:"unidad_responsable,omitempty"`
	Estado            string `json:"estado"`
	Notas             string `json:"notas"`
	Fake              bool   `json:"fake"`
}

// RegistrarReservas publica la agenda ficticia. registrar(r, deps) del frente reservas.
func RegistrarReservas(r *gin.Engine, deps Deps) {
	reservas := Reservas{Path: deps.ReservasPath}
	r.GET("/api/v1/geo/reservas-mock", reservas.Get)
}

func (h Reservas) Get(c *gin.Context) {
	aviso := "Agenda ficticia de demostración. No está conectada a una hoja de cálculo ni a una fuente institucional."
	body, err := os.ReadFile(h.Path)
	if err != nil {
		c.JSON(200, gin.H{"fake": true, "aviso": aviso, "total": 0, "reservas": []reservaItem{}})
		return
	}
	var raw struct {
		Reservas []reservaItem `json:"reservas"`
	}
	if err := json.Unmarshal(body, &raw); err != nil {
		c.JSON(500, gin.H{"error": "no se pudo leer la agenda ficticia"})
		return
	}
	for i := range raw.Reservas {
		raw.Reservas[i].Notas = "Dato ficticio de demostración."
		raw.Reservas[i].Fake = true
	}
	if raw.Reservas == nil {
		raw.Reservas = []reservaItem{}
	}
	payload := gin.H{
		"fake":     true,
		"aviso":    aviso,
		"total":    len(raw.Reservas),
		"reservas": raw.Reservas,
	}
	encoded, err := json.Marshal(payload)
	if err != nil {
		c.JSON(500, gin.H{"error": "no se pudo leer la agenda ficticia"})
		return
	}
	if strings.Contains(string(encoded), "SHEET") || strings.Contains(string(encoded), "docs.google.com") {
		c.JSON(500, gin.H{"error": "la agenda ficticia todavía arrastra una referencia externa"})
		return
	}
	c.Data(200, "application/json; charset=utf-8", encoded)
}
