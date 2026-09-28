package usecases

import (
	"context"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/contracts"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/dto"
	domainErrors "github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/errors"
)

type areaVerdeUseCase struct {
	repo      contracts.IAreaVerdeRepository
	auditoria contracts.IAuditoriaService
}

// NewAreaVerdeUseCase creates a new AreaVerdeUseCase instance.
func NewAreaVerdeUseCase(repo contracts.IAreaVerdeRepository, auditoria contracts.IAuditoriaService) contracts.IAreaVerdeUseCase {
	return &areaVerdeUseCase{
		repo:      repo,
		auditoria: auditoria,
	}
}

func (u *areaVerdeUseCase) Fichas(ctx context.Context, q string) ([]dto.FichaDTO, error) {
	rows, err := u.repo.Fichas(ctx, q)
	if err != nil {
		return nil, err
	}
	if rows == nil {
		return []dto.FichaDTO{}, nil
	}
	return rows, nil
}

func (u *areaVerdeUseCase) ActualizarFicha(ctx context.Context, featureID string, req dto.ActualizarFichaDTO, usuarioID *int64) (*dto.FichaDTO, error) {
	featureID = strings.TrimSpace(featureID)
	nombre := strings.TrimSpace(req.Nombre)
	uso := strings.TrimSpace(req.Uso)
	riego := strings.TrimSpace(req.RiegoAct)
	referencia := strings.TrimSpace(req.Referencia)

	if featureID == "" || utf8.RuneCountInString(nombre) > 160 || utf8.RuneCountInString(referencia) > 500 {
		return nil, domainErrors.ErrEntrada
	}

	antes, err := u.repo.ObtenerFichaPorFeatureID(ctx, featureID)
	if err != nil {
		return nil, err
	}
	if antes == nil {
		return nil, domainErrors.ErrFichaNoEncontrada
	}

	item, err := u.repo.ActualizarFicha(ctx, featureID, nombre, uso, riego, referencia)
	if err != nil {
		return nil, err
	}

	if u.auditoria != nil {
		_ = u.auditoria.RegistrarCambio(ctx, dto.RegistrarCambioDTO{
			Entidad:   "areas_verdes",
			EntidadID: featureID,
			Accion:    "edicion",
			Antes: map[string]any{
				"nombre":     antes.Nombre,
				"uso":        antes.Uso,
				"riego_act":  antes.RiegoAct,
				"referencia": antes.Referencia,
			},
			Despues: map[string]any{
				"nombre":     item.Nombre,
				"uso":        item.Uso,
				"riego_act":  item.RiegoAct,
				"referencia": item.Referencia,
			},
			UsuarioID: usuarioID,
		})
	}

	return item, nil
}

func (u *areaVerdeUseCase) CrearSinGeom(ctx context.Context, req dto.CrearAreaSinGeomDTO, usuarioID *int64) (*dto.FichaDTO, error) {
	featureID := strings.TrimSpace(req.FeatureID)
	nombre := strings.TrimSpace(req.Nombre)
	uso := strings.TrimSpace(req.Uso)

	if nombre == "" {
		return nil, domainErrors.ErrEntrada
	}
	if featureID == "" {
		featureID = fmt.Sprintf("AV-P%d", time.Now().Unix()%100000000)
	}
	if len(featureID) > 40 || strings.ContainsAny(featureID, " \t") {
		return nil, domainErrors.ErrEntrada
	}

	item, err := u.repo.CrearSinGeom(ctx, featureID, nombre, uso)
	if err != nil {
		return nil, err
	}

	if u.auditoria != nil {
		_ = u.auditoria.RegistrarCambio(ctx, dto.RegistrarCambioDTO{
			Entidad:   "areas_verdes",
			EntidadID: item.FeatureID,
			Accion:    "alta",
			Antes:     nil,
			Despues: map[string]any{
				"feature_id": item.FeatureID,
				"nombre":     item.Nombre,
				"uso":        item.Uso,
			},
			UsuarioID: usuarioID,
		})
	}

	return item, nil
}
