// Package models defines GORM database models mapping to PostgreSQL tables.
package models

// AllModels returns an instance of each defined model struct for schema inspection and drift detection.
func AllModels() []any {
	return []any{
		&ActividadModel{},
		&ActividadAvanceModel{},
		&ActividadEventoModel{},
		&AreaVerdeModel{},
		&BebederoModel{},
		&CambioModel{},
		&CapaAuxiliarModel{},
		&CapatazModel{},
		&CatalogoModel{},
		&CodigoHistoricoModel{},
		&CuadrillaModel{},
		&EjemplarModel{},
		&EspecieModel{},
		&EvidenciaModel{},
		&InventarioModel{},
		&LugarModel{},
		&OrdenServicioModel{},
		&PermisoModel{},
		&PersonalLaborModel{},
		&PodaModel{},
		&PoligonoCuadrillaModel{},
		&PuntoPUCPModel{},
		&ReservaJardinModel{},
		&RiegoRegistroModel{},
		&SesionModel{},
		&SolicitudModel{},
		&TachoModel{},
		&UsuarioModel{},
		&ViveroCatalogoModel{},
		&ViveroRegistroModel{},
		&ZonaSupervisionModel{},
	}
}
