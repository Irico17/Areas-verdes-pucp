package contracts

import (
	"context"

	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/dto"
)

// IZonificacionRepository persists capataz sectors, roads, quarters and building references.
type IZonificacionRepository interface {
	ListarSectores(ctx context.Context, soloActivos bool) ([]dto.SectorCapatazDTO, error)
	CrearSector(ctx context.Context, in dto.CrearSectorDTO) (dto.SectorCapatazDTO, error)
	ActualizarSector(ctx context.Context, in dto.ActualizarSectorDTO) (dto.SectorCapatazDTO, error)
	DesactivarSector(ctx context.Context, codigo string, usuarioID int64) error
	ImportarSectores(ctx context.Context, filas []dto.CrearSectorDTO, usuarioID int64) (dto.ImportacionSectorDTO, error)
	ListarLugares(ctx context.Context) ([]dto.LugarCatalogoDTO, error)
	LugarPorID(ctx context.Context, id int64) (dto.LugarCatalogoDTO, bool, error)
	LugarPorNorm(ctx context.Context, nombreNorm string) (dto.LugarCatalogoDTO, bool, error)
	ContarLugares(ctx context.Context) (int64, error)
	ViasGeoJSON(ctx context.Context) ([]byte, error)
	ImportarVias(ctx context.Context, filas []dto.ViaAltaDTO, usuarioID int64) (dto.ImportacionViaDTO, error)
	CuartelesGeoJSON(ctx context.Context) ([]byte, error)
	CrearReferente(ctx context.Context, in dto.CrearReferenteDTO) (dto.ReferenteDTO, error)
}

// IZonificacionUseCase is the capataz-sector, place, road and quarter flow.
type IZonificacionUseCase interface {
	ListarSectores(ctx context.Context, soloActivos bool) (*dto.SectorListDTO, error)
	CrearSector(ctx context.Context, in dto.CrearSectorDTO) (dto.SectorCapatazDTO, error)
	ActualizarSector(ctx context.Context, in dto.ActualizarSectorDTO) (dto.SectorCapatazDTO, error)
	DesactivarSector(ctx context.Context, codigo string, usuarioID int64) error
	ImportarSectores(ctx context.Context, filas []dto.CrearSectorDTO, usuarioID int64) (dto.ImportacionSectorDTO, error)
	ListarLugares(ctx context.Context) ([]dto.LugarCatalogoDTO, error)
	ResolverLugar(ctx context.Context, in dto.ResolverLugarDTO) (dto.LugarResueltoDTO, error)
	Vias(ctx context.Context) ([]byte, error)
	ImportarVias(ctx context.Context, body []byte, usuarioID int64) (dto.ImportacionViaDTO, error)
	Cuarteles(ctx context.Context) ([]byte, error)
	Edificios(ctx context.Context) ([]dto.EdificioRefDTO, error)
	CrearReferente(ctx context.Context, in dto.CrearReferenteDTO) (dto.ReferenteDTO, error)
}
