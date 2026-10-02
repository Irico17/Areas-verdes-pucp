package services

import (
	"context"
	"sync"

	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/contracts"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/entities"
	domainErrors "github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/errors"
)

type memoriaPermisos struct {
	mu        sync.Mutex
	concedido map[string]bool
	roles     map[string]entities.RolCatalogo
}

// NuevaMemoriaPermisos copies the seed into an editable in-memory store.
func NuevaMemoriaPermisos(matriz map[string][]string) contracts.IPermisoRepository {
	m := &memoriaPermisos{
		concedido: map[string]bool{},
		roles:     map[string]entities.RolCatalogo{},
	}
	orden := 1
	for rol, acciones := range matriz {
		m.roles[rol] = entities.RolCatalogo{Codigo: rol, Nombre: rol, Activo: true, Orden: orden}
		orden++
		for _, accion := range acciones {
			m.concedido[rol+"|"+accion] = true
		}
	}
	return m
}

func (m *memoriaPermisos) Listar(_ context.Context) ([]entities.Permiso, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := make([]entities.Permiso, 0, len(m.concedido))
	for clave, ok := range m.concedido {
		if !ok {
			continue
		}
		rol, accion, encontrado := partirClave(clave)
		if !encontrado {
			continue
		}
		out = append(out, entities.Permiso{Rol: rol, Accion: accion})
	}
	return out, nil
}

func (m *memoriaPermisos) Sembrar(_ context.Context, permisos []entities.Permiso) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, p := range permisos {
		if _, ok := m.roles[p.Rol]; !ok {
			m.roles[p.Rol] = entities.RolCatalogo{Codigo: p.Rol, Nombre: p.Rol, Activo: true}
		}
		m.concedido[p.Rol+"|"+p.Accion] = true
	}
	return nil
}

func (m *memoriaPermisos) Concedido(_ context.Context, rol, accion string) (bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.concedido[rol+"|"+accion], nil
}

func (m *memoriaPermisos) Establecer(_ context.Context, rol, accion string, concedido bool) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	clave := rol + "|" + accion
	if concedido {
		m.concedido[clave] = true
		return nil
	}
	delete(m.concedido, clave)
	return nil
}

func (m *memoriaPermisos) RolActivo(_ context.Context, rol string) (bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	fila, ok := m.roles[rol]
	if !ok {
		return false, nil
	}
	return fila.Activo, nil
}

func (m *memoriaPermisos) ListarRoles(_ context.Context) ([]entities.RolCatalogo, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := make([]entities.RolCatalogo, 0, len(m.roles))
	for _, rol := range m.roles {
		out = append(out, rol)
	}
	return out, nil
}

func (m *memoriaPermisos) CrearRol(_ context.Context, rol entities.RolCatalogo) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.roles[rol.Codigo]; ok {
		return domainErrors.ErrRolYaExiste
	}
	m.roles[rol.Codigo] = rol
	return nil
}

func (m *memoriaPermisos) ActualizarRol(_ context.Context, codigo string, activo bool) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	fila, ok := m.roles[codigo]
	if !ok {
		return domainErrors.ErrRolNoDisponible
	}
	fila.Activo = activo
	m.roles[codigo] = fila
	return nil
}

func partirClave(clave string) (string, string, bool) {
	for i := 0; i < len(clave); i++ {
		if clave[i] == '|' {
			return clave[:i], clave[i+1:], true
		}
	}
	return "", "", false
}
