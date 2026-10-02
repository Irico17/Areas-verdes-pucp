package mapper

import (
	"time"

	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/entities"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/persistence/models"
)

// CuadrillaModelToEntity maps CuadrillaModel to domain Cuadrilla.
func CuadrillaModelToEntity(m *models.CuadrillaModel) *entities.Cuadrilla {
	if m == nil {
		return nil
	}
	return &entities.Cuadrilla{
		ID:             m.ID,
		NombreFicticio: m.NombreFicticio,
		Turno:          m.Turno,
		Activo:         m.Activo,
	}
}

// ZonaSupervisionModelToEntity maps ZonaSupervisionModel to domain ZonaSupervision.
func ZonaSupervisionModelToEntity(m *models.ZonaSupervisionModel, conGeom bool) *entities.ZonaSupervision {
	if m == nil {
		return nil
	}
	return &entities.ZonaSupervision{
		ID:      m.ID,
		Codigo:  m.Codigo,
		Nombre:  m.Nombre,
		AreaM2:  m.AreaM2,
		ConGeom: conGeom,
		Activo:  m.Activo,
	}
}

// LugarModelToEntity maps LugarModel to domain Lugar.
func LugarModelToEntity(m *models.LugarModel) *entities.Lugar {
	if m == nil {
		return nil
	}
	return &entities.Lugar{
		ID:                m.ID,
		Nombre:            m.Nombre,
		NombreNorm:        m.NombreNorm,
		Lat:               m.Lat,
		Lon:               m.Lon,
		ZonaSupervisionID: m.ZonaSupervisionID,
		Activo:            m.Activo,
	}
}

// EspecieModelToEntity maps EspecieModel to domain Especie.
func EspecieModelToEntity(m *models.EspecieModel) *entities.Especie {
	if m == nil {
		return nil
	}
	comun := ""
	if m.NombreComun != nil {
		comun = *m.NombreComun
	}
	return &entities.Especie{
		ID:               m.ID,
		NombreCientifico: m.NombreCientifico,
		NombreComun:      comun,
		Activo:           m.Activo,
	}
}

// EjemplarModelToEntity maps EjemplarModel to domain Ejemplar.
func EjemplarModelToEntity(m *models.EjemplarModel) *entities.Ejemplar {
	if m == nil {
		return nil
	}
	var codigo, comun, tipoVeg, ref, obs string
	if m.Codigo != nil {
		codigo = *m.Codigo
	}
	if m.NombreComun != nil {
		comun = *m.NombreComun
	}
	if m.TipoVegetacion != nil {
		tipoVeg = *m.TipoVegetacion
	}
	if m.Referencia != nil {
		ref = *m.Referencia
	}
	if m.ObservacionFen2026 != nil {
		obs = *m.ObservacionFen2026
	}
	var sectorNombre, sectorClase, cientifico, lugar string
	if m.SectorCuartelNombre != nil {
		sectorNombre = *m.SectorCuartelNombre
	}
	if m.SectorCuartelClase != nil {
		sectorClase = *m.SectorCuartelClase
	}
	if m.EspecieCientifico != nil {
		cientifico = *m.EspecieCientifico
	}
	if m.LugarNombre != nil {
		lugar = *m.LugarNombre
	}
	return &entities.Ejemplar{
		ID:                  m.ID,
		NumeroOrigen:        m.NumeroOrigen,
		Codigo:              codigo,
		EspecieID:           m.EspecieID,
		NombreComun:         comun,
		TipoVegetacion:      tipoVeg,
		Cantidad:            m.Cantidad,
		UbicacionLugarID:    m.UbicacionLugarID,
		Referencia:          ref,
		Lat:                 m.Lat,
		Lon:                 m.Lon,
		ObservacionFen2026:  obs,
		Salud:               m.Salud,
		SectorCuartelID:     m.SectorCuartelID,
		SectorCuartelNombre: sectorNombre,
		SectorCuartelClase:  sectorClase,
		EspecieCientifico:   cientifico,
		LugarNombre:         lugar,
		Activo:              m.Activo,
	}
}

// CodigoHistoricoModelToEntity maps CodigoHistoricoModel to domain CodigoHistorico.
func CodigoHistoricoModelToEntity(m *models.CodigoHistoricoModel) *entities.CodigoHistorico {
	if m == nil {
		return nil
	}
	nuevo := ""
	if m.CodigoNuevo != nil {
		nuevo = *m.CodigoNuevo
	}
	return &entities.CodigoHistorico{
		ID:             m.ID,
		EjemplarID:     m.EjemplarID,
		CodigoAnterior: m.CodigoAnterior,
		CodigoNuevo:    nuevo,
	}
}

// PoligonoCuadrillaModelToEntity maps PoligonoCuadrillaModel to domain PoligonoCuadrilla.
func PoligonoCuadrillaModelToEntity(m *models.PoligonoCuadrillaModel, conGeom bool) *entities.PoligonoCuadrilla {
	if m == nil {
		return nil
	}
	var codigo, nombre, cuadrillaID string
	if m.Codigo != nil {
		codigo = *m.Codigo
	}
	if m.Nombre != nil {
		nombre = *m.Nombre
	}
	if m.CuadrillaID != nil {
		cuadrillaID = *m.CuadrillaID
	}
	return &entities.PoligonoCuadrilla{
		ID:                m.ID,
		FeatureID:         m.FeatureID,
		Codigo:            codigo,
		Nombre:            nombre,
		CuadrillaID:       cuadrillaID,
		ZonaSupervisionID: m.ZonaSupervisionID,
		ConGeom:           conGeom,
		Activo:            m.Activo,
	}
}

// CapaAuxiliarModelToEntity maps CapaAuxiliarModel to domain CapaFicha.
func CapaAuxiliarModelToEntity(m *models.CapaAuxiliarModel) *entities.CapaFicha {
	if m == nil {
		return nil
	}
	var codigo, nombre, clase, riego, pertenecen string
	if m.Codigo != nil {
		codigo = *m.Codigo
	}
	if m.Nombre != nil {
		nombre = *m.Nombre
	}
	if m.Clase != nil {
		clase = *m.Clase
	}
	if m.RiegoAct != nil {
		riego = *m.RiegoAct
	}
	if m.Pertenecen != nil {
		pertenecen = *m.Pertenecen
	}
	return &entities.CapaFicha{
		ID:         m.ID,
		FeatureID:  m.FeatureID,
		Nombre:     nombre,
		Codigo:     codigo,
		Clase:      clase,
		Riego:      riego,
		Pertenecen: pertenecen,
		Activo:     true,
	}
}

// CambioEntityToModel maps domain Cambio to CambioModel.
func CambioEntityToModel(c *entities.Cambio) *models.CambioModel {
	if c == nil {
		return nil
	}
	createdAt := c.CreatedAt
	if createdAt.IsZero() {
		createdAt = time.Now()
	}
	return &models.CambioModel{
		ID:        c.ID,
		Entidad:   c.Entidad,
		EntidadID: c.EntidadID,
		Accion:    c.Accion,
		Antes:     c.Antes,
		Despues:   c.Despues,
		UsuarioID: c.UsuarioID,
		LoteID:    c.LoteID,
		CreatedAt: createdAt,
	}
}
