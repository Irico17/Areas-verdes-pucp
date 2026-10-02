// Package usecases implements application business logic.
package usecases

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/contracts"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/dto"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/entities"
	domainErrors "github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/errors"
)

const topeBytes = 8 << 20

func uuidOK(v string) bool {
	return uuidRe.MatchString(v)
}

type evidenciaUseCase struct {
	repo    contracts.IEvidenciaRepository
	storage contracts.IAlmacenArchivos
	exifSvc contracts.IExifService
}

// NewEvidenciaUseCase creates a new instance of IEvidenciaUseCase.
func NewEvidenciaUseCase(
	repo contracts.IEvidenciaRepository,
	storage contracts.IAlmacenArchivos,
	exifSvc contracts.IExifService,
) contracts.IEvidenciaUseCase {
	return &evidenciaUseCase{
		repo:    repo,
		storage: storage,
		exifSvc: exifSvc,
	}
}

// Listar retrieves evidence attachments optionally filtered by activity ID.
func (u *evidenciaUseCase) Listar(ctx context.Context, actividadID string) (*dto.ListarEvidenciasResponseDTO, error) {
	items, err := u.repo.Listar(ctx, actividadID)
	if err != nil {
		return nil, err
	}
	out := make([]dto.EvidenciaDTO, len(items))
	for i, item := range items {
		out[i] = dto.EvidenciaDTO{
			ID:          item.ID,
			ActividadID: item.ActividadID,
			Nombre:      item.Nombre,
			Mime:        item.Mime,
			Bytes:       item.Bytes,
			Nota:        item.Nota,
			EventoID:    item.EventoID,
			CreatedAt:   item.CreatedAt.UTC().Format(time.RFC3339),
		}
	}
	return &dto.ListarEvidenciasResponseDTO{Evidencias: out}, nil
}

// Subir validates inputs, checks idempotency and coordinates storing evidence metadata and blob.
func (u *evidenciaUseCase) Subir(ctx context.Context, in dto.SubirEvidenciaDTO) (*dto.SubirEvidenciaResponseDTO, error) {
	if !uuidOK(in.ID) || !uuidOK(in.ActividadID) {
		return nil, domainErrors.InputError{Reason: "id y actividad_id deben ser UUID"}
	}
	if len(in.Contenido) == 0 {
		return nil, domainErrors.InputError{Reason: "falta el archivo"}
	}
	if len(in.Contenido) > topeBytes {
		return nil, domainErrors.InputError{Reason: "el archivo supera 8 MB"}
	}
	mime, ext, ok := u.exifSvc.MimeReal(in.Contenido)
	if !ok {
		return nil, domainErrors.InputError{Reason: "se admite jpg, png, webp o pdf"}
	}
	sum := sha256.Sum256(in.Contenido)
	hash := hex.EncodeToString(sum[:])
	pedido := strings.ToLower(strings.TrimSpace(in.SHA256))
	if pedido != "" && pedido != hash {
		return nil, domainErrors.InputError{Reason: "el hash no coincide con el archivo"}
	}
	if (in.Lat == nil) != (in.Lon == nil) {
		return nil, domainErrors.InputError{Reason: "lat y lon van juntos"}
	}
	if in.Lat != nil && (*in.Lat < -90 || *in.Lat > 90 || *in.Lon < -180 || *in.Lon > 180) {
		return nil, domainErrors.InputError{Reason: "la ubicación está fuera de rango"}
	}
	nombre := strings.TrimSpace(in.Nombre)
	if nombre == "" {
		nombre = "evidencia" + ext
	}
	if utf8.RuneCountInString(nombre) > 180 {
		return nil, domainErrors.InputError{Reason: "el nombre es demasiado largo"}
	}
	nota := strings.TrimSpace(in.Nota)
	if utf8.RuneCountInString(nota) > 500 {
		return nil, domainErrors.InputError{Reason: "la nota es demasiado larga"}
	}
	exifData, err := u.exifSvc.FiltrarExif(in.Exif)
	if err != nil {
		return nil, err
	}
	ordenID := strings.TrimSpace(in.OrdenID)
	if ordenID != "" && !uuidOK(ordenID) {
		return nil, domainErrors.InputError{Reason: "orden_id debe ser un UUID"}
	}
	solicitudID := strings.TrimSpace(in.SolicitudID)
	if solicitudID != "" && !uuidOK(solicitudID) {
		return nil, domainErrors.InputError{Reason: "solicitud_id debe ser un UUID"}
	}
	if in.EventoID < 0 {
		return nil, domainErrors.InputError{Reason: "evento_id no es válido"}
	}

	res, err := u.repo.Guardar(ctx, entities.GuardarEvidencia{
		ID:          in.ID,
		ActividadID: in.ActividadID,
		OrdenID:     ordenID,
		Nombre:      nombre,
		Mime:        mime,
		Ext:         ext,
		Bytes:       len(in.Contenido),
		Contenido:   in.Contenido,
		Nota:        nota,
		Hash:        hash,
		Lat:         in.Lat,
		Lon:         in.Lon,
		Exif:        exifData,
		Rol:         in.Rol,
		CapatazID:   in.CapatazID,
		UsuarioID:   in.UsuarioID,
		EventoID:    in.EventoID,
		SolicitudID: solicitudID,
	}, u.storage)
	if err != nil {
		return nil, err
	}

	return &dto.SubirEvidenciaResponseDTO{
		ID:          res.ID,
		Idempotente: res.Idempotente,
	}, nil
}

// Abrir retrieves the file stream and MIME type for downloading an evidence attachment.
func (u *evidenciaUseCase) Abrir(ctx context.Context, id string) (io.ReadCloser, string, error) {
	ruta, mime, err := u.repo.ObtenerRutaYMime(ctx, id)
	if err != nil {
		return nil, "", err
	}
	body, err := u.storage.Open(ctx, ruta)
	if err != nil {
		return nil, "", domainErrors.ErrArchivoNoDisponible
	}
	return body, mime, nil
}
