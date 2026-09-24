package catalogos

import (
	"context"
	"regexp"
	"strings"
	"unicode/utf8"

	"gorm.io/gorm"
)

// Clases que la pantalla de catálogos puede administrar.
var Clases = []string{
	"tipo_actividad", "estado", "prioridad", "lugar", "especie",
	"motivo_archivo", "turno", "fuente",
}

var codigoRe = regexp.MustCompile(`^[a-z0-9_]{2,32}$`)

// Item es una fila de catálogo.
type Item struct {
	ID     int64  `json:"id"`
	Clase  string `json:"clase"`
	Codigo string `json:"codigo"`
	Nombre string `json:"nombre"`
	Activo bool   `json:"activo"`
	Orden  int    `json:"orden"`
}

// Store lee y escribe catálogos. No borra filas.
type Store struct{ db *gorm.DB }

func NewStore(db *gorm.DB) *Store { return &Store{db: db} }

func ClaseOK(clase string) bool {
	for _, item := range Clases {
		if item == clase {
			return true
		}
	}
	return false
}

func (s *Store) List(ctx context.Context, clase string, soloActivos bool) ([]Item, error) {
	q := `SELECT id, clase, codigo, nombre, activo, orden FROM catalogos WHERE 1=1`
	args := []any{}
	if clase != "" {
		q += ` AND clase = $1`
		args = append(args, clase)
	}
	if soloActivos {
		if clase != "" {
			q += ` AND activo`
		} else {
			q += ` AND activo`
		}
	}
	q += ` ORDER BY clase, orden, codigo`
	var out []Item
	err := s.db.WithContext(ctx).Raw(q, args...).Scan(&out).Error
	if out == nil {
		out = []Item{}
	}
	return out, err
}

// Activo dice si el código sigue vigente en esa clase.
func (s *Store) Activo(ctx context.Context, clase, codigo string) (bool, error) {
	var n int
	err := s.db.WithContext(ctx).Raw(`
		SELECT count(*) FROM catalogos WHERE clase = $1 AND codigo = $2 AND activo`,
		clase, codigo).Scan(&n).Error
	return n == 1, err
}

func (s *Store) Create(ctx context.Context, clase, codigo, nombre string) (Item, error) {
	var zero Item
	clase = strings.TrimSpace(clase)
	codigo = strings.TrimSpace(codigo)
	nombre = strings.TrimSpace(nombre)
	if !ClaseOK(clase) {
		return zero, errInput("clase de catálogo no reconocida")
	}
	if !codigoRe.MatchString(codigo) {
		return zero, errInput("el código usa minúsculas, números y guion bajo")
	}
	if nombre == "" || utf8.RuneCountInString(nombre) > 80 {
		return zero, errInput("el nombre es obligatorio y de hasta 80 caracteres")
	}
	var item Item
	err := s.db.WithContext(ctx).Raw(`
		INSERT INTO catalogos (clase, codigo, nombre)
		VALUES ($1, $2, $3)
		ON CONFLICT (clase, codigo) DO UPDATE SET nombre = EXCLUDED.nombre, activo = TRUE
		RETURNING id, clase, codigo, nombre, activo, orden`,
		clase, codigo, nombre).Row().Scan(&item.ID, &item.Clase, &item.Codigo, &item.Nombre, &item.Activo, &item.Orden)
	return item, err
}

// Deactivate es la baja lógica.
func (s *Store) Deactivate(ctx context.Context, id int64) error {
	res := s.db.WithContext(ctx).Exec(`UPDATE catalogos SET activo = FALSE WHERE id = $1`, id)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return errInput("no existe ese ítem")
	}
	return nil
}

type inputError string

func (e inputError) Error() string { return string(e) }

func errInput(msg string) error { return inputError(msg) }

// EsEntrada distingue un 400 de un fallo de base.
func EsEntrada(err error) (string, bool) {
	if err == nil {
		return "", false
	}
	if e, ok := err.(inputError); ok {
		return string(e), true
	}
	return "", false
}
