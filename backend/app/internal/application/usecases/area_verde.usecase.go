package usecases

import (
	"context"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/contracts"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/dto"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/entities"
	domainErrors "github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/errors"
)

// areaVerdeUseCase implements the area-card use cases. Like the old API
// (catastro.Store.ActualizarFicha/CrearSinGeom) it does not write `cambios`.
type areaVerdeUseCase struct {
	repo contracts.IAreaVerdeRepository
}

// NewAreaVerdeUseCase creates a new AreaVerdeUseCase instance.
func NewAreaVerdeUseCase(repo contracts.IAreaVerdeRepository) contracts.IAreaVerdeUseCase {
	return &areaVerdeUseCase{repo: repo}
}

func (u *areaVerdeUseCase) Fichas(ctx context.Context, q string) ([]dto.FichaDTO, error) {
	rows, err := u.repo.Fichas(ctx, q)
	if err != nil {
		return nil, err
	}
	out := make([]dto.FichaDTO, len(rows))
	for i, r := range rows {
		out[i] = *fichaEntityToDTO(&r)
	}
	return out, nil
}

func (u *areaVerdeUseCase) ActualizarFicha(ctx context.Context, featureID string, req dto.ActualizarFichaDTO, usuarioID *int64) (*dto.FichaDTO, error) {
	featureID = strings.TrimSpace(featureID)
	nombre := strings.TrimSpace(req.Nombre)
	uso := strings.TrimSpace(req.Uso)
	riego := strings.TrimSpace(req.RiegoAct)
	referencia := strings.TrimSpace(req.Referencia)

	if featureID == "" || utf8.RuneCountInString(nombre) > 160 {
		return nil, domainErrors.ErrEntrada
	}

	item, err := u.repo.ActualizarFicha(ctx, featureID, nombre, uso, riego, referencia)
	if err != nil {
		return nil, err
	}

	return fichaEntityToDTO(item), nil
}

func (u *areaVerdeUseCase) CrearSinGeom(ctx context.Context, req dto.CrearAreaSinGeomDTO, usuarioID *int64) (*dto.FichaDTO, error) {
	featureID := strings.TrimSpace(req.FeatureID)
	nombre := strings.TrimSpace(req.Nombre)
	uso := strings.TrimSpace(req.Uso)

	// The old API inserted the row and then failed the 160-rune check with 400;
	// the check now runs before the insert (same status, no orphan row).
	if nombre == "" || utf8.RuneCountInString(nombre) > 160 {
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

	return fichaEntityToDTO(item), nil
}

func (u *areaVerdeUseCase) Baja(ctx context.Context, featureID string, usuarioID *int64) error {
	featureID = strings.TrimSpace(featureID)
	if featureID == "" {
		return domainErrors.ErrEntrada
	}
	return u.repo.Baja(ctx, featureID, usuarioID)
}

func fichaEntityToDTO(e *entities.AreaVerdeFicha) *dto.FichaDTO {
	if e == nil {
		return nil
	}
	return &dto.FichaDTO{
		FeatureID:  e.FeatureID,
		Nombre:     e.Nombre,
		Uso:        e.Uso,
		RiegoAct:   e.RiegoAct,
		Referencia: e.Referencia,
		AreaM2:     e.AreaM2,
		ConGeom:    e.ConGeom,
		Activo:     e.Activo,
	}
}
