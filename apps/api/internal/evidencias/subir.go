package evidencias

import (
	"bytes"
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"regexp"
	"strings"
	"unicode/utf8"

	"campusverde/api/internal/blobs"
	"campusverde/api/internal/operacion"

	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"
)

const topeBytes = 8 << 20

// Entrada es el POST multipart ya leído. El hash lo calcula el servidor.
type Entrada struct {
	ID          string
	ActividadID string
	Nombre      string
	Nota        string
	SHA256      string
	Lat         *float64
	Lon         *float64
	Exif        json.RawMessage
	Rol         string
	CapatazID   string
	Contenido   []byte
}

// Resultado de una subida. Idempotente es cierto si el UUID ya estaba con el mismo hash.
type Resultado struct {
	ID          string
	Idempotente bool
}

// Guardar valida MIME real, cuadrilla y hash. El mismo UUID con otro contenido es 409.
func Guardar(ctx context.Context, db *gorm.DB, files blobs.Store, in Entrada) (Resultado, error) {
	if db == nil || files == nil {
		return Resultado{}, errors.New("almacén no disponible")
	}
	if !uuidOK(in.ID) || !uuidOK(in.ActividadID) {
		return Resultado{}, operacion.InputError{Reason: "id y actividad_id deben ser UUID"}
	}
	if len(in.Contenido) == 0 {
		return Resultado{}, operacion.InputError{Reason: "falta el archivo"}
	}
	if len(in.Contenido) > topeBytes {
		return Resultado{}, operacion.InputError{Reason: "el archivo supera 8 MB"}
	}
	mime, ext, ok := mimeReal(in.Contenido)
	if !ok {
		return Resultado{}, operacion.InputError{Reason: "se admite jpg, png, webp o pdf"}
	}
	sum := sha256.Sum256(in.Contenido)
	hash := hex.EncodeToString(sum[:])
	pedido := strings.ToLower(strings.TrimSpace(in.SHA256))
	if pedido != "" && pedido != hash {
		return Resultado{}, operacion.InputError{Reason: "el hash no coincide con el archivo"}
	}
	if (in.Lat == nil) != (in.Lon == nil) {
		return Resultado{}, operacion.InputError{Reason: "lat y lon van juntos"}
	}
	if in.Lat != nil && (*in.Lat < -90 || *in.Lat > 90 || *in.Lon < -180 || *in.Lon > 180) {
		return Resultado{}, operacion.InputError{Reason: "la ubicación está fuera de rango"}
	}
	nombre := strings.TrimSpace(in.Nombre)
	if nombre == "" {
		nombre = "evidencia" + ext
	}
	if utf8.RuneCountInString(nombre) > 180 {
		return Resultado{}, operacion.InputError{Reason: "el nombre es demasiado largo"}
	}
	nota := strings.TrimSpace(in.Nota)
	if utf8.RuneCountInString(nota) > 500 {
		return Resultado{}, operacion.InputError{Reason: "la nota es demasiado larga"}
	}
	var exif any
	if len(bytes.TrimSpace(in.Exif)) > 0 && string(bytes.TrimSpace(in.Exif)) != "null" {
		if !json.Valid(in.Exif) || in.Exif[0] != '{' {
			return Resultado{}, operacion.InputError{Reason: "exif debe ser un objeto JSON"}
		}
		exif = string(in.Exif)
	}

	var out Resultado
	err := db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var previo sql.NullString
		row := tx.Raw(`SELECT sha256 FROM evidencias WHERE id = $1`, in.ID).Row()
		scanErr := row.Scan(&previo)
		if scanErr == nil {
			if previo.Valid && strings.EqualFold(previo.String, hash) {
				out = Resultado{ID: in.ID, Idempotente: true}
				return nil
			}
			return operacion.ErrConflicto
		}
		if !errors.Is(scanErr, sql.ErrNoRows) {
			return scanErr
		}

		var asignado string
		asig := tx.Raw(`SELECT COALESCE(assigned_capataz_id, '') FROM actividades WHERE id = $1`, in.ActividadID).Row()
		if err := asig.Scan(&asignado); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return operacion.ErrNoEncontrada
			}
			return err
		}
		if in.Rol == "capataz" && asignado != strings.TrimSpace(in.CapatazID) {
			return operacion.ErrProhibido
		}

		ref, err := files.Put(ctx, in.ID+ext, bytes.NewReader(in.Contenido), mime)
		if err != nil {
			return err
		}
		res := tx.Exec(`
			INSERT INTO evidencias (id, actividad_id, nombre, mime, bytes, ruta, nota, sha256, lat, lon, exif)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11::jsonb)`,
			in.ID, in.ActividadID, nombre, mime, len(in.Contenido), ref, nota, hash, in.Lat, in.Lon, exif,
		)
		if res.Error != nil {
			if esDuplicado(res.Error) {
				var ya sql.NullString
				if err := tx.Raw(`SELECT sha256 FROM evidencias WHERE id = $1`, in.ID).Row().Scan(&ya); err != nil {
					return err
				}
				if ya.Valid && strings.EqualFold(ya.String, hash) {
					out = Resultado{ID: in.ID, Idempotente: true}
					return nil
				}
				return operacion.ErrConflicto
			}
			return res.Error
		}
		if err := tx.Exec(`
			INSERT INTO actividad_eventos (actividad_id, tipo, actor_rol, nota)
			VALUES ($1, 'evidencia', $2, $3)`,
			in.ActividadID, rolEvento(in.Rol), "Evidencia: "+nombre,
		).Error; err != nil {
			return err
		}
		out = Resultado{ID: in.ID}
		return nil
	})
	return out, err
}

func rolEvento(rol string) string {
	switch rol {
	case "capataz", "coordinacion", "admin", "jefatura":
		return rol
	default:
		return "sesion"
	}
}

var uuidRe = regexp.MustCompile(`(?i)^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)

func uuidOK(v string) bool {
	return uuidRe.MatchString(v)
}

func mimeReal(b []byte) (string, string, bool) {
	if len(b) >= 12 && bytes.Equal(b[0:4], []byte("RIFF")) && bytes.Equal(b[8:12], []byte("WEBP")) {
		return "image/webp", ".webp", true
	}
	ct := http.DetectContentType(b)
	switch {
	case strings.HasPrefix(ct, "image/jpeg"):
		return "image/jpeg", ".jpg", true
	case strings.HasPrefix(ct, "image/png"):
		return "image/png", ".png", true
	case strings.HasPrefix(ct, "application/pdf"):
		return "application/pdf", ".pdf", true
	default:
		return "", "", false
	}
}

func esDuplicado(err error) bool {
	var pg *pgconn.PgError
	if errors.As(err, &pg) && pg.Code == "23505" {
		return true
	}
	return strings.Contains(err.Error(), "duplicate key")
}

// Leer limita el cuerpo a 8 MB + 1 para rechazar lo que se pasa.
func Leer(r io.Reader) ([]byte, error) {
	b, err := io.ReadAll(io.LimitReader(r, topeBytes+1))
	if err != nil {
		return nil, err
	}
	if len(b) > topeBytes {
		return nil, operacion.InputError{Reason: "el archivo supera 8 MB"}
	}
	return b, nil
}
