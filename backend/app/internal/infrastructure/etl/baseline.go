package etl

// Conteos del baseline recuperado (data/raw, 2026-09-24).
// El modo estricto falla si el archivo presente no coincide: evita cargar un parseo truncado.
const (
	ExpectedAreas      = 521
	ExpectedZonas      = 534
	ExpectedJardines   = 21
	ExpectedXerofitica = 10
)

// PIIFragments son valores del campo "jefes" que no deben salir en data/v1 ni en la API.
var PIIFragments = []string{
	"jefes",
	"responsable",
	"responsable",
	"responsable",
	"responsable",
	"responsable",
	"Bosque húme",
	"campo depo",
}
