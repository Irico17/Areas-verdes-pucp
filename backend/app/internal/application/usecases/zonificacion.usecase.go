package usecases

import (
	"context"
	"encoding/json"
	"regexp"
	"sort"
	"strings"
	"unicode/utf8"

	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/contracts"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/dto"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/entities"
	apperrors "github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/errors"
)

var (
	codigoSectorRe = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)
	colorSectorRe  = regexp.MustCompile(`^#[0-9a-f]{6}$`)
)

type zonificacionUseCase struct {
	repo     contracts.IZonificacionRepository
	archivos contracts.IArchivoEstaticoAdapter
}

// NewZonificacionUseCase creates the zoning use case.
func NewZonificacionUseCase(
	repo contracts.IZonificacionRepository,
	archivos contracts.IArchivoEstaticoAdapter,
) contracts.IZonificacionUseCase {
	return &zonificacionUseCase{repo: repo, archivos: archivos}
}

func (uc *zonificacionUseCase) ListarSectores(ctx context.Context, soloActivos bool) (*dto.SectorListDTO, error) {
	rows, err := uc.repo.ListarSectores(ctx, soloActivos)
	if err != nil {
		return nil, err
	}
	if rows == nil {
		rows = []dto.SectorCapatazDTO{}
	}
	return &dto.SectorListDTO{Sectores: rows}, nil
}

func (uc *zonificacionUseCase) CrearSector(ctx context.Context, in dto.CrearSectorDTO) (dto.SectorCapatazDTO, error) {
	limpio, err := normalizarSector(in.Codigo, in.Nombre, in.Color)
	if err != nil {
		return dto.SectorCapatazDTO{}, err
	}
	in.Codigo, in.Nombre, in.Color = limpio.codigo, limpio.nombre, limpio.color
	return uc.repo.CrearSector(ctx, in)
}

func (uc *zonificacionUseCase) ActualizarSector(ctx context.Context, in dto.ActualizarSectorDTO) (dto.SectorCapatazDTO, error) {
	limpio, err := normalizarSector(in.Codigo, in.Nombre, in.Color)
	if err != nil {
		return dto.SectorCapatazDTO{}, err
	}
	in.Codigo, in.Nombre, in.Color = limpio.codigo, limpio.nombre, limpio.color
	return uc.repo.ActualizarSector(ctx, in)
}

func (uc *zonificacionUseCase) DesactivarSector(ctx context.Context, codigo string, usuarioID int64) error {
	codigo = strings.TrimSpace(strings.ToLower(codigo))
	if !codigoSectorRe.MatchString(codigo) || len(codigo) > 40 {
		return apperrors.ErrCodigoSector
	}
	return uc.repo.DesactivarSector(ctx, codigo, usuarioID)
}

func (uc *zonificacionUseCase) ImportarSectores(ctx context.Context, filas []dto.CrearSectorDTO, usuarioID int64) (dto.ImportacionSectorDTO, error) {
	if len(filas) == 0 {
		return dto.ImportacionSectorDTO{}, apperrors.ErrArchivoSinFilas
	}
	vistos := map[string]dto.CrearSectorDTO{}
	orden := make([]string, 0, len(filas))
	for _, fila := range filas {
		limpio, err := normalizarSector(fila.Codigo, fila.Nombre, fila.Color)
		if err != nil {
			return dto.ImportacionSectorDTO{}, err
		}
		if _, ok := vistos[limpio.codigo]; !ok {
			orden = append(orden, limpio.codigo)
		}
		vistos[limpio.codigo] = dto.CrearSectorDTO{
			Codigo: limpio.codigo,
			Nombre: limpio.nombre,
			Color:  limpio.color,
		}
	}
	unicos := make([]dto.CrearSectorDTO, 0, len(orden))
	for _, codigo := range orden {
		unicos = append(unicos, vistos[codigo])
	}
	return uc.repo.ImportarSectores(ctx, unicos, usuarioID)
}

func (uc *zonificacionUseCase) ListarLugares(ctx context.Context) ([]dto.LugarCatalogoDTO, error) {
	rows, err := uc.repo.ListarLugares(ctx)
	if err != nil {
		return nil, err
	}
	if rows == nil {
		rows = []dto.LugarCatalogoDTO{}
	}
	return rows, nil
}

// ResolverLugar matches a catalog id or an existing name. An unknown name is
// rejected and is never inserted. lugar_libre is echoed and not written.
func (uc *zonificacionUseCase) ResolverLugar(ctx context.Context, in dto.ResolverLugarDTO) (dto.LugarResueltoDTO, error) {
	out := dto.LugarResueltoDTO{LugarLibre: strings.TrimSpace(in.LugarLibre)}
	if in.LugarID != nil && *in.LugarID > 0 {
		lugar, ok, err := uc.repo.LugarPorID(ctx, *in.LugarID)
		if err != nil {
			return dto.LugarResueltoDTO{}, err
		}
		if !ok {
			return dto.LugarResueltoDTO{}, apperrors.ErrLugarDesconocido
		}
		out.LugarID = &lugar.ID
		out.Nombre = lugar.Nombre
		return out, nil
	}
	nombre := strings.TrimSpace(in.Nombre)
	if nombre == "" {
		return out, nil
	}
	if strings.ContainsAny(nombre, "<>") || utf8.RuneCountInString(nombre) > 120 {
		return dto.LugarResueltoDTO{}, apperrors.ErrLugarDesconocido
	}
	lugar, ok, err := uc.repo.LugarPorNorm(ctx, entities.NormalizarNombre(nombre))
	if err != nil {
		return dto.LugarResueltoDTO{}, err
	}
	if !ok {
		return dto.LugarResueltoDTO{}, apperrors.ErrLugarDesconocido
	}
	out.LugarID = &lugar.ID
	out.Nombre = lugar.Nombre
	return out, nil
}

func (uc *zonificacionUseCase) Vias(ctx context.Context) ([]byte, error) {
	body, err := uc.repo.ViasGeoJSON(ctx)
	if err != nil {
		return nil, err
	}
	if len(body) == 0 {
		return []byte(`{"type":"FeatureCollection","name":"vias","features":[]}`), nil
	}
	return body, nil
}

func (uc *zonificacionUseCase) ImportarVias(ctx context.Context, body []byte, usuarioID int64) (dto.ImportacionViaDTO, error) {
	filas, err := leerViasGeoJSON(body)
	if err != nil {
		return dto.ImportacionViaDTO{}, err
	}
	if len(filas) == 0 {
		return dto.ImportacionViaDTO{}, apperrors.ErrViaVacia
	}
	return uc.repo.ImportarVias(ctx, filas, usuarioID)
}

func (uc *zonificacionUseCase) Cuarteles(ctx context.Context) ([]byte, error) {
	body, err := uc.repo.CuartelesGeoJSON(ctx)
	if err != nil {
		return nil, err
	}
	if len(body) == 0 {
		return []byte(`{"type":"FeatureCollection","name":"cuarteles","aviso":"sin archivo de cuarteles","features":[]}`), nil
	}
	return body, nil
}

func (uc *zonificacionUseCase) Edificios(ctx context.Context) ([]dto.EdificioRefDTO, error) {
	ids, err := uc.idsEdificio(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]dto.EdificioRefDTO, 0, len(ids))
	for id, nombre := range ids {
		out = append(out, dto.EdificioRefDTO{ID: id, Nombre: nombre})
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Nombre == out[j].Nombre {
			return out[i].ID < out[j].ID
		}
		if out[i].Nombre == "" {
			return false
		}
		if out[j].Nombre == "" {
			return true
		}
		return out[i].Nombre < out[j].Nombre
	})
	return out, nil
}

func (uc *zonificacionUseCase) CrearReferente(ctx context.Context, in dto.CrearReferenteDTO) (dto.ReferenteDTO, error) {
	in.EdificioID = strings.TrimSpace(in.EdificioID)
	if in.LugarID <= 0 || in.EdificioID == "" || strings.ContainsAny(in.EdificioID, " <>") {
		return dto.ReferenteDTO{}, apperrors.ErrEdificioDesconocido
	}
	if _, ok, err := uc.repo.LugarPorID(ctx, in.LugarID); err != nil {
		return dto.ReferenteDTO{}, err
	} else if !ok {
		return dto.ReferenteDTO{}, apperrors.ErrLugarDesconocido
	}
	ids, err := uc.idsEdificio(ctx)
	if err != nil {
		return dto.ReferenteDTO{}, err
	}
	if _, ok := ids[in.EdificioID]; !ok {
		return dto.ReferenteDTO{}, apperrors.ErrEdificioDesconocido
	}
	return uc.repo.CrearReferente(ctx, in)
}

type sectorLimpio struct {
	codigo, nombre, color string
}

func normalizarSector(codigo, nombre, color string) (sectorLimpio, error) {
	codigo = strings.TrimSpace(strings.ToLower(codigo))
	nombre = strings.TrimSpace(nombre)
	color = strings.ToLower(strings.TrimSpace(color))
	if !codigoSectorRe.MatchString(codigo) || len(codigo) < 2 || len(codigo) > 40 {
		return sectorLimpio{}, apperrors.ErrCodigoSector
	}
	if nombre == "" || utf8.RuneCountInString(nombre) > 80 || strings.ContainsAny(nombre, "<>") {
		return sectorLimpio{}, apperrors.ErrNombreObligatorio
	}
	if !colorSectorRe.MatchString(color) {
		return sectorLimpio{}, apperrors.ErrColorSector
	}
	return sectorLimpio{codigo: codigo, nombre: nombre, color: color}, nil
}

type featureVia struct {
	ID         string          `json:"id"`
	Geometry   json.RawMessage `json:"geometry"`
	Properties map[string]any  `json:"properties"`
}

func leerViasGeoJSON(body []byte) ([]dto.ViaAltaDTO, error) {
	var fc struct {
		Type     string       `json:"type"`
		Features []featureVia `json:"features"`
	}
	if err := json.Unmarshal(body, &fc); err != nil || fc.Type != "FeatureCollection" {
		return nil, apperrors.ErrGeomVia
	}
	vistos := map[string]dto.ViaAltaDTO{}
	orden := make([]string, 0, len(fc.Features))
	for _, f := range fc.Features {
		if len(f.Geometry) == 0 || string(f.Geometry) == "null" {
			return nil, apperrors.ErrGeomVia
		}
		if err := lineaEnCampus(f.Geometry); err != nil {
			return nil, err
		}
		id := strings.TrimSpace(f.ID)
		if id == "" {
			id = textoProp(f.Properties, "feature_id")
		}
		if id == "" || strings.ContainsAny(id, " <>") || len(id) > 80 {
			return nil, apperrors.ErrGeomVia
		}
		nombre := textoProp(f.Properties, "nombre")
		if nombre == "" {
			nombre = textoProp(f.Properties, "name")
		}
		if strings.ContainsAny(nombre, "<>") || utf8.RuneCountInString(nombre) > 120 {
			return nil, apperrors.ErrGeomVia
		}
		if _, ok := vistos[id]; !ok {
			orden = append(orden, id)
		}
		vistos[id] = dto.ViaAltaDTO{FeatureID: id, Nombre: nombre, GeoJSON: string(f.Geometry)}
	}
	out := make([]dto.ViaAltaDTO, 0, len(orden))
	for _, id := range orden {
		out = append(out, vistos[id])
	}
	return out, nil
}

func textoProp(props map[string]any, clave string) string {
	if props == nil {
		return ""
	}
	v, ok := props[clave]
	if !ok || v == nil {
		return ""
	}
	s, ok := v.(string)
	if !ok {
		return ""
	}
	return strings.TrimSpace(s)
}

func lineaEnCampus(geom json.RawMessage) error {
	var g struct {
		Type        string          `json:"type"`
		Coordinates json.RawMessage `json:"coordinates"`
	}
	if err := json.Unmarshal(geom, &g); err != nil {
		return apperrors.ErrGeomVia
	}
	var puntos [][2]float64
	switch g.Type {
	case "LineString":
		var linea [][2]float64
		if err := json.Unmarshal(g.Coordinates, &linea); err != nil {
			return apperrors.ErrGeomVia
		}
		puntos = linea
	case "MultiLineString":
		var lineas [][][2]float64
		if err := json.Unmarshal(g.Coordinates, &lineas); err != nil {
			return apperrors.ErrGeomVia
		}
		for _, linea := range lineas {
			puntos = append(puntos, linea...)
		}
	default:
		return apperrors.ErrGeomVia
	}
	if len(puntos) < 2 {
		return apperrors.ErrGeomVia
	}
	for _, p := range puntos {
		if err := entities.ValidarPuntoCampus(p[1], p[0]); err != nil {
			return apperrors.ErrGeomVia
		}
	}
	return nil
}

func (uc *zonificacionUseCase) idsEdificio(ctx context.Context) (map[string]string, error) {
	if uc.archivos == nil {
		return nil, apperrors.ErrEdificioDesconocido
	}
	body, err := uc.archivos.LeerEdificios(ctx)
	if err != nil || len(body) == 0 {
		return nil, apperrors.ErrEdificioDesconocido
	}
	var fc struct {
		Features []struct {
			ID         string `json:"id"`
			Properties struct {
				FeatureID string `json:"feature_id"`
				Nombre    string `json:"nombre"`
			} `json:"properties"`
		} `json:"features"`
	}
	if err := json.Unmarshal(body, &fc); err != nil {
		return nil, apperrors.ErrEdificioDesconocido
	}
	out := map[string]string{}
	for _, f := range fc.Features {
		id := strings.TrimSpace(f.ID)
		if id == "" {
			id = strings.TrimSpace(f.Properties.FeatureID)
		}
		if id == "" {
			continue
		}
		out[id] = strings.TrimSpace(f.Properties.Nombre)
	}
	if len(out) == 0 {
		return nil, apperrors.ErrEdificioDesconocido
	}
	return out, nil
}
