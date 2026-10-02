package database

import "database/sql"

// ConteosVisibles returns green areas with geometry and crew polygons that
// already have an operational sector. Both are what the map draws.
func ConteosVisibles(sqlDB *sql.DB) (areasConGeom, zonasConSector int, err error) {
	err = sqlDB.QueryRow(`
		SELECT
		  (SELECT count(*) FROM areas_verdes
		    WHERE geom IS NOT NULL AND NOT ST_IsEmpty(geom)),
		  (SELECT count(*) FROM poligonos_cuadrilla
		    WHERE geom IS NOT NULL AND NOT ST_IsEmpty(geom) AND sector IS NOT NULL)`).
		Scan(&areasConGeom, &zonasConSector)
	return areasConGeom, zonasConSector, err
}

// CatastroIncompleto is true when the loaded cadastre is below the published
// source counts (521 areas and 534 sectors). A single fictional polygon does
// not count as the campus map.
func CatastroIncompleto(areasConGeom, zonasConSector, minAreas, minZonas int) bool {
	return areasConGeom < minAreas || zonasConSector < minZonas
}
