// Package usecases contains application business logic.
package usecases

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/contracts"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/dto"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/constants/enums"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/entities"
	domainErrors "github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/errors"
)

// ValidarRiego verifies that zone, shift, and area are valid for an irrigation record.
func ValidarRiego(zonaID, turno string, superficie float64) error {
	switch zonaID {
	case "Z1", "Z2", "Z3", "Z4":
	default:
		return domainErrors.InputError{Reason: "el riego guarda la zona por identificador Z1–Z4"}
	}
	if turno != "manana" && turno != "tarde" {
		return domainErrors.InputError{Reason: "el turno es mañana o tarde"}
	}
	if superficie < 0 {
		return domainErrors.InputError{Reason: "la superficie no puede ser negativa"}
	}
	return nil
}

// ConsultaRiego builds the SQL query and arguments for listing irrigation records.
func ConsultaRiego(capatazID string) (string, []any) {
	q := `
		SELECT r.id::text, COALESCE(s.nombre, r.sector), r.turno, COALESCE(r.capataz_id, ''), COALESCE(c.equipo, ''),
		       to_char(r.fecha, 'YYYY-MM-DD'), r.nota, COALESCE(z.codigo, ''), COALESCE(r.ciclo, '')
		FROM riego_registros r
		LEFT JOIN capataces c ON c.id = r.capataz_id
		LEFT JOIN zonas_supervision z ON z.id = r.zona_supervision_id
		LEFT JOIN sectores_capataz s ON s.id = r.sector_id`
	var args []any
	if id := strings.TrimSpace(capatazID); id != "" {
		q += ` WHERE r.capataz_id = $1`
		args = append(args, id)
	}
	q += ` ORDER BY r.fecha DESC, r.created_at DESC LIMIT 100`
	return q, args
}

type riegoUseCase struct {
	repo contracts.IRiegoRepository
}

// NewRiegoUseCase creates a new IRiegoUseCase.
func NewRiegoUseCase(repo contracts.IRiegoRepository) contracts.IRiegoUseCase {
	return &riegoUseCase{repo: repo}
}

// porcentajeCobertura es entero y trunca. Cero sectores activos da 0, no un error.
// No es la fórmula oficial: la jefatura de sección todavía no la valida.
func porcentajeCobertura(regados, activos int) int {
	if activos <= 0 || regados <= 0 {
		return 0
	}
	if regados > activos {
		regados = activos
	}
	return regados * 100 / activos
}

func leyendaCobertura(porcentaje int) string {
	return fmt.Sprintf("Cobertura provisional: %d %%. Pendiente de validar con la jefatura de sección.", porcentaje)
}

func (u *riegoUseCase) Listar(ctx context.Context, capatazID string) (*dto.RiegoResponseDTO, error) {
	items, err := u.repo.Listar(ctx, capatazID)
	if err != nil {
		return nil, err
	}
	regados, activos, err := u.repo.CoberturaMes(ctx)
	if err != nil {
		return nil, err
	}
	cobertura := porcentajeCobertura(regados, activos)
	registros := make([]dto.RiegoDTO, 0, len(items))
	for _, it := range items {
		r := dto.RiegoDTO{
			ID:     it.ID,
			Sector: it.Sector,
			Turno:  it.Turno,
			Fecha:  it.Fecha,
			Nota:   it.Nota,
			Ciclo:  it.Ciclo,
		}
		if it.CapatazID != nil {
			r.CapatazID = *it.CapatazID
		}
		if it.Equipo != nil {
			r.Equipo = *it.Equipo
		}
		if it.ZonaSupervisionCode != nil {
			r.ZonaID = *it.ZonaSupervisionCode
		}
		registros = append(registros, r)
	}
	return &dto.RiegoResponseDTO{
		Aviso:       leyendaCobertura(cobertura),
		Provisional: true,
		Cobertura:   cobertura,
		Registros:   registros,
	}, nil
}

func (u *riegoUseCase) Crear(ctx context.Context, in dto.CrearRiegoDTO, actorRol, capatazID string) (*dto.CrearRiegoResponseDTO, error) {
	if actorRol == string(enums.RolCapataz) {
		in.CapatazID = capatazID
	}
	in.Sector = strings.TrimSpace(in.Sector)
	in.Turno = strings.TrimSpace(in.Turno)
	in.ZonaID = strings.TrimSpace(in.ZonaID)
	if err := ValidarRiego(in.ZonaID, in.Turno, in.Superficie); err != nil {
		return nil, err
	}
	if in.SectorID <= 0 {
		return nil, domainErrors.InputError{Reason: "el sector es el del catálogo de capataz"}
	}
	if _, err := time.Parse("2006-01-02", in.Fecha); err != nil {
		return nil, domainErrors.InputError{Reason: "la fecha usa el formato AAAA-MM-DD"}
	}
	if !uuidRe.MatchString(in.ID) {
		return nil, domainErrors.InputError{Reason: "id debe ser un UUID"}
	}
	if err := u.repo.Crear(ctx, entities.NuevoTurnoRiego{
		ID:         in.ID,
		Sector:     in.Sector,
		SectorID:   in.SectorID,
		Turno:      in.Turno,
		CapatazID:  in.CapatazID,
		Fecha:      in.Fecha,
		Nota:       in.Nota,
		ZonaID:     in.ZonaID,
		Ciclo:      in.Ciclo,
		Superficie: in.Superficie,
	}); err != nil {
		return nil, err
	}
	return &dto.CrearRiegoResponseDTO{ID: in.ID}, nil
}
