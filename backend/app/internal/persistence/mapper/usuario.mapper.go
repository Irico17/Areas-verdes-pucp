// Package mapper provides domain-model conversions.
package mapper

import (
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/entities"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/persistence/models"
)

// UsuarioToEntity maps UsuarioModel to domain Usuario entity.
func UsuarioToEntity(m *models.UsuarioModel) *entities.Usuario {
	if m == nil {
		return nil
	}
	capataz := ""
	if m.CapatazID != nil {
		capataz = *m.CapatazID
	}
	return &entities.Usuario{
		ID:           m.ID,
		Usuario:      m.Usuario,
		Nombre:       m.Nombre,
		Rol:          m.Rol,
		CapatazID:    capataz,
		PasswordHash: m.PasswordHash,
		Activo:       m.Activo,
	}
}

// UsuarioToModel maps domain Usuario entity to UsuarioModel.
func UsuarioToModel(e *entities.Usuario) *models.UsuarioModel {
	if e == nil {
		return nil
	}
	var capataz *string
	if e.CapatazID != "" {
		capataz = &e.CapatazID
	}
	return &models.UsuarioModel{
		ID:           e.ID,
		Usuario:      e.Usuario,
		Nombre:       e.Nombre,
		Rol:          e.Rol,
		CapatazID:    capataz,
		PasswordHash: e.PasswordHash,
		Activo:       e.Activo,
	}
}
