// Package usecases contains application use case orchestrators.
package usecases

import (
	"context"
	"encoding/csv"
	"strconv"
	"strings"

	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/contracts"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/dto"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/entities"
	domainErrors "github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/errors"
)

type inventarioCampoUseCase struct {
	repo   contracts.IInventarioCampoRepository
	parser contracts.IPuntosParser
}

// NewInventarioCampoUseCase creates a new IInventarioCampoUseCase instance.
func NewInventarioCampoUseCase(
	repo contracts.IInventarioCampoRepository,
	parser contracts.IPuntosParser,
) contracts.IInventarioCampoUseCase {
	return &inventarioCampoUseCase{
		repo:   repo,
		parser: parser,
	}
}

// CuerpoConContacto checks whether the JSON body mentions phone, placeId, or website.
func CuerpoConContacto(raw string) bool {
	low := strings.ToLower(raw)
	return strings.Contains(low, "phone") || strings.Contains(raw, "placeId") || strings.Contains(low, "place_id") || strings.Contains(low, "website")
}

// UrlLimpia strips placeId or phone references from point URLs.
func UrlLimpia(raw string) string {
	raw = strings.TrimSpace(raw)
	if strings.Contains(raw, "placeId") || strings.Contains(raw, "place_id") || strings.Contains(strings.ToLower(raw), "phone") {
		return ""
	}
	return raw
}

func conteosNoNegativos(t dto.TachoDTO) bool {
	for _, n := range []int{
		t.NoAprovechables, t.PapelCarton, t.Plastico, t.Vidrio,
		t.Pilas, t.Peligrosos, t.RAEE, t.Metales, t.Aniquem,
		t.IntermediosPlastico, t.IntermediosMetal,
	} {
		if n < 0 {
			return false
		}
	}
	return strings.HasPrefix(strings.TrimSpace(t.Codigo), "PT")
}

func tachoEntityToDTO(t *entities.Tacho) *dto.TachoDTO {
	if t == nil {
		return nil
	}
	return &dto.TachoDTO{
		ID:                  t.ID,
		Codigo:              t.Codigo,
		Lat:                 t.Lat,
		Lon:                 t.Lon,
		Nota:                t.Nota,
		Lugar:               t.Lugar,
		Espacios:            t.Espacios,
		Accion:              t.Accion,
		TachoActual:         t.TachoActual,
		TachoNuevo:          t.TachoNuevo,
		Recomendaciones:     t.Recomendaciones,
		NoAprovechables:     t.NoAprovechables,
		PapelCarton:         t.PapelCarton,
		Plastico:            t.Plastico,
		Vidrio:              t.Vidrio,
		Pilas:               t.Pilas,
		Peligrosos:          t.Peligrosos,
		RAEE:                t.RAEE,
		Metales:             t.Metales,
		Aniquem:             t.Aniquem,
		IntermediosPlastico: t.IntermediosPlastico,
		IntermediosMetal:    t.IntermediosMetal,
		Activo:              t.Activo,
	}
}

func tachoDTOToEntity(d dto.TachoDTO) entities.Tacho {
	return entities.Tacho{
		ID:                  d.ID,
		Codigo:              d.Codigo,
		Lat:                 d.Lat,
		Lon:                 d.Lon,
		Nota:                d.Nota,
		Lugar:               d.Lugar,
		Espacios:            d.Espacios,
		Accion:              d.Accion,
		TachoActual:         d.TachoActual,
		TachoNuevo:          d.TachoNuevo,
		Recomendaciones:     d.Recomendaciones,
		NoAprovechables:     d.NoAprovechables,
		PapelCarton:         d.PapelCarton,
		Plastico:            d.Plastico,
		Vidrio:              d.Vidrio,
		Pilas:               d.Pilas,
		Peligrosos:          d.Peligrosos,
		RAEE:                d.RAEE,
		Metales:             d.Metales,
		Aniquem:             d.Aniquem,
		IntermediosPlastico: d.IntermediosPlastico,
		IntermediosMetal:    d.IntermediosMetal,
		Activo:              d.Activo,
	}
}

func bebederoEntityToDTO(b *entities.Bebedero) *dto.BebederoDTO {
	if b == nil {
		return nil
	}
	return &dto.BebederoDTO{
		ID:      b.ID,
		Codigo:  b.Codigo,
		Subtipo: b.Subtipo,
		Estado:  b.Estado,
		Sede:    b.Sede,
		Lat:     b.Lat,
		Lon:     b.Lon,
		Activo:  b.Activo,
	}
}

func bebederoDTOToEntity(d dto.BebederoDTO) entities.Bebedero {
	return entities.Bebedero{
		ID:      d.ID,
		Codigo:  d.Codigo,
		Subtipo: d.Subtipo,
		Estado:  d.Estado,
		Sede:    d.Sede,
		Lat:     d.Lat,
		Lon:     d.Lon,
		Activo:  d.Activo,
	}
}

func puntoEntityToDTO(p *entities.PuntoPUCP) *dto.PuntoDTO {
	if p == nil {
		return nil
	}
	return &dto.PuntoDTO{
		ID:     p.ID,
		Titulo: p.Titulo,
		Lat:    p.Lat,
		Lon:    p.Lon,
		URL:    p.URL,
		Activo: p.Activo,
	}
}

func puntoDTOToEntity(d dto.PuntoDTO) entities.PuntoPUCP {
	return entities.PuntoPUCP{
		ID:     d.ID,
		Titulo: d.Titulo,
		Lat:    d.Lat,
		Lon:    d.Lon,
		URL:    d.URL,
		Activo: d.Activo,
	}
}

func reservaEntityToDTO(r *entities.ReservaJardin) *dto.ReservaDTO {
	if r == nil {
		return nil
	}
	return &dto.ReservaDTO{
		ID:         r.ID,
		JardinID:   r.JardinID,
		Fecha:      r.Fecha,
		HoraInicio: r.HoraInicio,
		HoraFin:    r.HoraFin,
		Estado:     r.Estado,
		Evento:     r.Evento,
		Unidad:     r.Unidad,
		Origen:     r.Origen,
		Activo:     r.Activo,
	}
}

func reservaDTOToEntity(d dto.ReservaDTO) entities.ReservaJardin {
	return entities.ReservaJardin{
		ID:         d.ID,
		JardinID:   d.JardinID,
		Fecha:      d.Fecha,
		HoraInicio: d.HoraInicio,
		HoraFin:    d.HoraFin,
		Estado:     d.Estado,
		Evento:     d.Evento,
		Unidad:     d.Unidad,
		Origen:     d.Origen,
		Activo:     d.Activo,
	}
}

func fichaCapaEntityToDTO(f *entities.FichaCapa) *dto.FichaCapaDTO {
	if f == nil {
		return nil
	}
	return &dto.FichaCapaDTO{
		ID:         f.ID,
		FeatureID:  f.FeatureID,
		Nombre:     f.Nombre,
		Codigo:     f.Codigo,
		Nota:       f.Nota,
		Clase:      f.Clase,
		Riego:      f.Riego,
		AreaM2:     f.AreaM2,
		PerimetroM: f.PerimetroM,
		Pertenecen: f.Pertenecen,
		Uso:        f.Uso,
		GeoJSON:    f.GeoJSON,
		Activo:     f.Activo,
	}
}

func fichaCapaDTOToEntity(d dto.FichaCapaDTO) entities.FichaCapa {
	return entities.FichaCapa{
		ID:         d.ID,
		FeatureID:  d.FeatureID,
		Nombre:     d.Nombre,
		Codigo:     d.Codigo,
		Nota:       d.Nota,
		Clase:      d.Clase,
		Riego:      d.Riego,
		AreaM2:     d.AreaM2,
		PerimetroM: d.PerimetroM,
		Pertenecen: d.Pertenecen,
		Uso:        d.Uso,
		GeoJSON:    d.GeoJSON,
		Activo:     d.Activo,
	}
}

// -------------------------------------------------------------
// TACHOS
// -------------------------------------------------------------

func (uc *inventarioCampoUseCase) ListarTachos(ctx context.Context) (dto.ListarTachosResponseDTO, error) {
	items, err := uc.repo.ListarTachos(ctx)
	if err != nil {
		return dto.ListarTachosResponseDTO{}, err
	}
	out := make([]dto.TachoDTO, 0, len(items))
	for i := range items {
		out = append(out, *tachoEntityToDTO(&items[i]))
	}
	return dto.ListarTachosResponseDTO{Tachos: out}, nil
}

func (uc *inventarioCampoUseCase) GuardarTacho(ctx context.Context, req dto.TachoDTO) (*dto.TachoDTO, error) {
	if !conteosNoNegativos(req) {
		return nil, domainErrors.ErrTachoInvalido
	}
	entity := tachoDTOToEntity(req)
	saved, err := uc.repo.GuardarTacho(ctx, entity)
	if err != nil {
		return nil, err
	}
	return tachoEntityToDTO(saved), nil
}

func (uc *inventarioCampoUseCase) ActualizarTacho(ctx context.Context, id int64, req dto.TachoDTO, rawJSON []byte) (*dto.TachoDTO, error) {
	entity := tachoDTOToEntity(req)
	updated, err := uc.repo.ActualizarTacho(ctx, id, entity, rawJSON)
	if err != nil {
		return nil, err
	}
	return tachoEntityToDTO(updated), nil
}

func (uc *inventarioCampoUseCase) CSVTachos(ctx context.Context) (string, error) {
	rows, err := uc.repo.ListarTachos(ctx)
	if err != nil {
		return "", err
	}
	var b strings.Builder
	w := csv.NewWriter(&b)
	_ = w.Write([]string{"codigo", "lugar", "recomendaciones", "no_aprovechables", "papel_carton", "plastico", "vidrio", "pilas", "peligrosos", "raee", "metales", "aniquem", "intermedios_plastico", "intermedios_metal"})
	for _, t := range rows {
		_ = w.Write([]string{
			t.Codigo, t.Lugar, t.Recomendaciones,
			strconv.Itoa(t.NoAprovechables), strconv.Itoa(t.PapelCarton),
			strconv.Itoa(t.Plastico), strconv.Itoa(t.Vidrio),
			strconv.Itoa(t.Pilas), strconv.Itoa(t.Peligrosos),
			strconv.Itoa(t.RAEE), strconv.Itoa(t.Metales),
			strconv.Itoa(t.Aniquem), strconv.Itoa(t.IntermediosPlastico),
			strconv.Itoa(t.IntermediosMetal),
		})
	}
	w.Flush()
	return b.String(), nil
}

func (uc *inventarioCampoUseCase) BajaTacho(ctx context.Context, id int64) error {
	return uc.repo.Baja(ctx, "tachos", id)
}

// -------------------------------------------------------------
// BEBEDEROS
// -------------------------------------------------------------

func (uc *inventarioCampoUseCase) ListarBebederos(ctx context.Context) (dto.ListarBebederosResponseDTO, error) {
	items, err := uc.repo.ListarBebederos(ctx)
	if err != nil {
		return dto.ListarBebederosResponseDTO{}, err
	}
	out := make([]dto.BebederoDTO, 0, len(items))
	for i := range items {
		out = append(out, *bebederoEntityToDTO(&items[i]))
	}
	return dto.ListarBebederosResponseDTO{Bebederos: out}, nil
}

func (uc *inventarioCampoUseCase) GuardarBebedero(ctx context.Context, req dto.BebederoDTO) (*dto.BebederoDTO, error) {
	switch req.Subtipo {
	case "fuente", "llenador", "nuevo", "deterioro", "baja":
	default:
		return nil, domainErrors.ErrBebederoInvalido
	}
	if req.Estado == "" || !strings.HasPrefix(req.Codigo, "PT_") {
		return nil, domainErrors.ErrBebederoInvalido
	}
	entity := bebederoDTOToEntity(req)
	saved, err := uc.repo.GuardarBebedero(ctx, entity)
	if err != nil {
		return nil, err
	}
	return bebederoEntityToDTO(saved), nil
}

func (uc *inventarioCampoUseCase) ActualizarBebedero(ctx context.Context, id int64, req dto.BebederoDTO, rawJSON []byte) (*dto.BebederoDTO, error) {
	entity := bebederoDTOToEntity(req)
	updated, err := uc.repo.ActualizarBebedero(ctx, id, entity, rawJSON)
	if err != nil {
		return nil, err
	}
	return bebederoEntityToDTO(updated), nil
}

func (uc *inventarioCampoUseCase) BajaBebedero(ctx context.Context, id int64) error {
	return uc.repo.Baja(ctx, "bebederos", id)
}

// -------------------------------------------------------------
// PUNTOS PUCP
// -------------------------------------------------------------

func (uc *inventarioCampoUseCase) ListarPuntos(ctx context.Context, q string) (dto.ListarPuntosResponseDTO, error) {
	items, err := uc.repo.ListarPuntos(ctx, q)
	if err != nil {
		return dto.ListarPuntosResponseDTO{}, err
	}
	out := make([]dto.PuntoDTO, 0, len(items))
	for i := range items {
		out = append(out, *puntoEntityToDTO(&items[i]))
	}
	return dto.ListarPuntosResponseDTO{Puntos: out}, nil
}

func (uc *inventarioCampoUseCase) GuardarPunto(ctx context.Context, req dto.PuntoDTO, rawBody []byte) (*dto.PuntoDTO, error) {
	if CuerpoConContacto(string(rawBody)) {
		return nil, domainErrors.ErrPuntoContacto
	}
	req.Titulo = strings.TrimSpace(req.Titulo)
	req.URL = UrlLimpia(req.URL)
	if req.Titulo == "" || req.Lat < -12.20 || req.Lat > -11.90 || req.Lon < -77.30 || req.Lon > -76.90 {
		return nil, domainErrors.ErrPuntoInvalido
	}
	entity := puntoDTOToEntity(req)
	saved, err := uc.repo.GuardarPunto(ctx, entity)
	if err != nil {
		return nil, err
	}
	return puntoEntityToDTO(saved), nil
}

func (uc *inventarioCampoUseCase) ActualizarPunto(ctx context.Context, id int64, req dto.PuntoDTO, rawJSON []byte) (*dto.PuntoDTO, error) {
	if CuerpoConContacto(string(rawJSON)) {
		return nil, domainErrors.ErrPuntoContacto
	}
	entity := puntoDTOToEntity(req)
	updated, err := uc.repo.ActualizarPunto(ctx, id, entity, rawJSON)
	if err != nil {
		return nil, err
	}
	return puntoEntityToDTO(updated), nil
}

func (uc *inventarioCampoUseCase) BajaPunto(ctx context.Context, id int64) error {
	return uc.repo.Baja(ctx, "puntos_pucp", id)
}

func (uc *inventarioCampoUseCase) FormatoPuntos(ctx context.Context, body []byte) (dto.FormatoPuntosResponseDTO, error) {
	rows, rech, omitidas, err := uc.parser.LeerPuntosPUCP(body)
	if err != nil {
		return dto.FormatoPuntosResponseDTO{}, domainErrors.ErrCSVInvalido
	}
	rechDTOs := make([]dto.RechazoDTO, 0, len(rech))
	for _, r := range rech {
		rechDTOs = append(rechDTOs, dto.RechazoDTO{
			Fuente: r.Fuente,
			Fila:   r.Fila,
			Campo:  r.Campo,
			Motivo: r.Motivo,
		})
	}
	return dto.FormatoPuntosResponseDTO{
		Filas:            len(rows),
		Rechazados:       rechDTOs,
		ColumnasOmitidas: omitidas,
		Nota:             "columnas omitidas por datos personales",
	}, nil
}

// -------------------------------------------------------------
// RESERVAS JARDIN
// -------------------------------------------------------------

func (uc *inventarioCampoUseCase) ListarReservas(ctx context.Context, desde, hasta string) (dto.ListarReservasResponseDTO, error) {
	items, err := uc.repo.ListarReservas(ctx, desde, hasta)
	if err != nil {
		return dto.ListarReservasResponseDTO{}, err
	}
	out := make([]dto.ReservaDTO, 0, len(items))
	for i := range items {
		out = append(out, *reservaEntityToDTO(&items[i]))
	}
	return dto.ListarReservasResponseDTO{
		Origen:   "ficticio",
		Aviso:    "Agenda ficticia. La hoja de reservas responde 401 y no se abre.",
		Reservas: out,
	}, nil
}

func (uc *inventarioCampoUseCase) GuardarReserva(ctx context.Context, req dto.ReservaDTO) (*dto.ReservaDTO, error) {
	if req.Origen != "" && req.Origen != "ficticio" {
		return nil, domainErrors.ErrReservaOrigen
	}
	switch req.Estado {
	case "reservado", "realizado", "cancelado":
	default:
		return nil, domainErrors.ErrReservaInvalida
	}
	if req.HoraFin <= req.HoraInicio || req.Fecha == "" || strings.TrimSpace(req.Evento) == "" {
		return nil, domainErrors.ErrReservaInvalida
	}
	entity := reservaDTOToEntity(req)
	saved, err := uc.repo.GuardarReserva(ctx, entity)
	if err != nil {
		return nil, err
	}
	return reservaEntityToDTO(saved), nil
}

func (uc *inventarioCampoUseCase) ActualizarReserva(ctx context.Context, id int64, req dto.ReservaDTO, rawJSON []byte) (*dto.ReservaDTO, error) {
	entity := reservaDTOToEntity(req)
	updated, err := uc.repo.ActualizarReserva(ctx, id, entity, rawJSON)
	if err != nil {
		return nil, err
	}
	return reservaEntityToDTO(updated), nil
}

func (uc *inventarioCampoUseCase) BajaReserva(ctx context.Context, id int64) error {
	return uc.repo.Baja(ctx, "reservas_jardin", id)
}

// -------------------------------------------------------------
// CAPAS AUXILIARES (FAUNA, PUERTAS, ETC.)
// -------------------------------------------------------------

func (uc *inventarioCampoUseCase) ListarCapa(ctx context.Context, capa string) (dto.ListarFichasCapaResponseDTO, error) {
	items, err := uc.repo.ListarFichas(ctx, capa)
	if err != nil {
		return dto.ListarFichasCapaResponseDTO{}, err
	}
	out := make([]dto.FichaCapaDTO, 0, len(items))
	for i := range items {
		out = append(out, *fichaCapaEntityToDTO(&items[i]))
	}
	return dto.ListarFichasCapaResponseDTO{Filas: out}, nil
}

func (uc *inventarioCampoUseCase) GuardarCapa(ctx context.Context, capa string, req dto.FichaCapaDTO) (*dto.FichaCapaDTO, error) {
	req.FeatureID = strings.TrimSpace(req.FeatureID)
	if req.FeatureID == "" {
		return nil, domainErrors.ErrFichaInvalida
	}
	entity := fichaCapaDTOToEntity(req)
	saved, err := uc.repo.GuardarFicha(ctx, capa, entity)
	if err != nil {
		return nil, err
	}
	return fichaCapaEntityToDTO(saved), nil
}

func (uc *inventarioCampoUseCase) ActualizarCapa(ctx context.Context, capa string, id int64, req dto.FichaCapaDTO, rawJSON []byte) (*dto.FichaCapaDTO, error) {
	entity := fichaCapaDTOToEntity(req)
	updated, err := uc.repo.ActualizarFicha(ctx, capa, id, entity, rawJSON)
	if err != nil {
		return nil, err
	}
	return fichaCapaEntityToDTO(updated), nil
}

func (uc *inventarioCampoUseCase) CSVCapa(ctx context.Context, capa string) (string, error) {
	rows, err := uc.repo.ListarFichas(ctx, capa)
	if err != nil {
		return "", err
	}
	var b strings.Builder
	b.WriteString("feature_id,nombre,codigo,nota,clase,riego,pertenecen,uso\n")
	for _, f := range rows {
		b.WriteString(strings.Join([]string{f.FeatureID, f.Nombre, f.Codigo, f.Nota, f.Clase, f.Riego, f.Pertenecen, f.Uso}, ",") + "\n")
	}
	return b.String(), nil
}

func (uc *inventarioCampoUseCase) BajaCapa(ctx context.Context, capa string, id int64) error {
	return uc.repo.Baja(ctx, capa, id)
}
