// Package services implements cross-cutting domain services.
package services

import (
	"context"
	"encoding/json"
	"reflect"

	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/contracts"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/dto"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/entities"
)

type auditoriaService struct {
	repo contracts.ICambioRepository
}

// NewAuditoriaService creates a new auditoria service instance.
func NewAuditoriaService(repo contracts.ICambioRepository) contracts.IAuditoriaService {
	return &auditoriaService{repo: repo}
}

// RegistrarCambio records an audit entry in the cambios table, skipping no-op edits.
func (s *auditoriaService) RegistrarCambio(ctx context.Context, req dto.RegistrarCambioDTO) error {
	var antesStr *string
	if req.Antes != nil {
		if raw, ok := req.Antes.(string); ok {
			antesStr = &raw
		} else if rawBytes, ok := req.Antes.([]byte); ok {
			str := string(rawBytes)
			antesStr = &str
		} else {
			b, err := json.Marshal(req.Antes)
			if err != nil {
				return err
			}
			str := string(b)
			antesStr = &str
		}
	}

	var despuesStr *string
	if req.Despues != nil {
		if raw, ok := req.Despues.(string); ok {
			despuesStr = &raw
		} else if rawBytes, ok := req.Despues.([]byte); ok {
			str := string(rawBytes)
			despuesStr = &str
		} else {
			b, err := json.Marshal(req.Despues)
			if err != nil {
				return err
			}
			str := string(b)
			despuesStr = &str
		}
	}

	// Skip no-op edits where antes == despues
	if antesStr != nil && despuesStr != nil {
		var a, d any
		if errA := json.Unmarshal([]byte(*antesStr), &a); errA == nil {
			if errD := json.Unmarshal([]byte(*despuesStr), &d); errD == nil {
				if reflect.DeepEqual(a, d) {
					return nil
				}
			}
		}
	}

	cambio := &entities.Cambio{
		Entidad:   req.Entidad,
		EntidadID: req.EntidadID,
		Accion:    req.Accion,
		Antes:     antesStr,
		Despues:   despuesStr,
		UsuarioID: req.UsuarioID,
		LoteID:    req.LoteID,
	}

	return s.repo.Crear(ctx, cambio)
}
