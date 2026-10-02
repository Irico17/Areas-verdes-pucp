package enums

// ClaseCatalogo identifies recognized catalog classes.
type ClaseCatalogo string

const (
	// ClaseTipoActividad represents the 'tipo_actividad' catalog class.
	ClaseTipoActividad ClaseCatalogo = "tipo_actividad"
	// ClaseEstado represents the 'estado' catalog class.
	ClaseEstado ClaseCatalogo = "estado"
	// ClasePrioridad represents the 'prioridad' catalog class.
	ClasePrioridad ClaseCatalogo = "prioridad"
	// ClaseLugar represents the 'lugar' catalog class.
	ClaseLugar ClaseCatalogo = "lugar"
	// ClaseEspecie represents the 'especie' catalog class.
	ClaseEspecie ClaseCatalogo = "especie"
	// ClaseMotivoArchivo represents the 'motivo_archivo' catalog class.
	ClaseMotivoArchivo ClaseCatalogo = "motivo_archivo"
	// ClaseTurno represents the 'turno' catalog class.
	ClaseTurno ClaseCatalogo = "turno"
	// ClaseFuente represents the 'fuente' catalog class.
	ClaseFuente ClaseCatalogo = "fuente"
	// ClaseClaseActividad represents the 'clase_actividad' catalog class.
	ClaseClaseActividad ClaseCatalogo = "clase_actividad"
	// ClasePlaga represents the 'plaga' catalog class.
	ClasePlaga ClaseCatalogo = "plaga"
	// ClaseProductoFitosanitario represents the 'producto_fitosanitario' catalog class.
	ClaseProductoFitosanitario ClaseCatalogo = "producto_fitosanitario"
	// ClaseFrecuencia represents the 'frecuencia' catalog class.
	ClaseFrecuencia ClaseCatalogo = "frecuencia"
	// ClaseSede represents the 'sede' catalog class.
	ClaseSede ClaseCatalogo = "sede"
	// ClaseCuartel represents the 'cuartel' catalog class.
	ClaseCuartel ClaseCatalogo = "cuartel"
	// ClaseSectorCapataz represents the 'sector_capataz' catalog class.
	ClaseSectorCapataz ClaseCatalogo = "sector_capataz"
	// ClaseNivelRiesgo represents the 'nivel_riesgo' catalog class.
	// Medio is not seeded. A later catalog row does not need a CHECK migration.
	ClaseNivelRiesgo ClaseCatalogo = "nivel_riesgo"
	// ClaseSubtipoActividad is the second level of an activity class.
	ClaseSubtipoActividad ClaseCatalogo = "subtipo_actividad"
)

// String returns the string representation of the catalog class.
func (c ClaseCatalogo) String() string {
	return string(c)
}

// ClasesCatalogoValidas returns the list of all supported catalog classes in canonical order.
func ClasesCatalogoValidas() []string {
	return []string{
		string(ClaseTipoActividad),
		string(ClaseEstado),
		string(ClasePrioridad),
		string(ClaseLugar),
		string(ClaseEspecie),
		string(ClaseMotivoArchivo),
		string(ClaseTurno),
		string(ClaseFuente),
		string(ClaseClaseActividad),
		string(ClasePlaga),
		string(ClaseProductoFitosanitario),
		string(ClaseFrecuencia),
		string(ClaseSede),
		string(ClaseCuartel),
		string(ClaseSectorCapataz),
		string(ClaseNivelRiesgo),
		string(ClaseSubtipoActividad),
	}
}

// EsValida checks if the catalog class is recognized.
func (c ClaseCatalogo) EsValida() bool {
	for _, valid := range ClasesCatalogoValidas() {
		if string(c) == valid {
			return true
		}
	}
	return false
}
