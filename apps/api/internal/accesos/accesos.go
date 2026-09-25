package accesos

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"net/http"
	"strings"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

const Cookie = "cv_sesion"

// Usuario es una cuenta local. CapatazID vacía si el rol no es de campo.
type Usuario struct {
	ID        int64  `json:"id"`
	Usuario   string `json:"usuario"`
	Nombre    string `json:"nombre"`
	Rol       string `json:"rol"`
	CapatazID string `json:"capataz_id,omitempty"`
}

type seedUser struct {
	usuario, nombre, rol, capataz string
}

var semillas = []seedUser{
	{"norte", "Equipo Norte", "capataz", "cap-norte"},
	{"sur", "Equipo Sur", "capataz", "cap-sur"},
	{"riego", "Equipo Riego", "capataz", "cap-riego"},
	{"coordinacion", "Coordinación", "coordinacion", ""},
	{"jefatura", "Jefatura", "jefatura", ""},
	{"admin", "Administración", "admin", ""},
}

// Matriz es la semilla de permisos. No es el SSO ni un editor de políticas.
var Matriz = map[string][]string{
	"capataz":      {"consultar", "registrar"},
	"coordinacion": {"consultar", "registrar", "validar", "solicitudes", "reportes"},
	"jefatura":     {"consultar", "validar", "reportes", "solicitudes"},
	"admin":        {"consultar", "registrar", "validar", "reportes", "catalogos", "solicitudes"},
}

// Permite dice si el rol tiene la acción semilla.
func Permite(rol, accion string) bool {
	for _, item := range Matriz[rol] {
		if item == accion {
			return true
		}
	}
	return false
}

// PermiteAlguno basta con una de las acciones.
func PermiteAlguno(rol string, acciones ...string) bool {
	for _, accion := range acciones {
		if Permite(rol, accion) {
			return true
		}
	}
	return false
}

// Ensure crea las cuentas locales si faltan y reescribe la matriz de permisos.
// La clave es la de desarrollo; no se registra en el log.
func Ensure(db *gorm.DB, password string) error {
	if db == nil {
		return errors.New("sin base")
	}
	password = strings.TrimSpace(password)
	if password == "" {
		return errors.New("falta CAMPUS_DEV_PASSWORD")
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	return db.Transaction(func(tx *gorm.DB) error {
		for _, item := range semillas {
			var n int
			if err := tx.Raw(`SELECT count(*) FROM usuarios WHERE usuario = $1`, item.usuario).Scan(&n).Error; err != nil {
				return err
			}
			if n > 0 {
				continue
			}
			if err := tx.Exec(`
				INSERT INTO usuarios (usuario, nombre, rol, capataz_id, password_hash)
				VALUES ($1, $2, $3, NULLIF($4, ''), $5)`,
				item.usuario, item.nombre, item.rol, item.capataz, string(hash),
			).Error; err != nil {
				return err
			}
		}
		if err := tx.Exec(`DELETE FROM permisos`).Error; err != nil {
			return err
		}
		for rol, acciones := range Matriz {
			for _, accion := range acciones {
				if err := tx.Exec(`INSERT INTO permisos (rol, accion) VALUES ($1, $2)`, rol, accion).Error; err != nil {
					return err
				}
			}
		}
		return nil
	})
}

// Store resuelve sesión contra Postgres.
type Store struct {
	db *gorm.DB
}

func NewStore(db *gorm.DB) *Store { return &Store{db: db} }

// Login comprueba la clave y deja un token opaco. El hash de la clave no sale.
func (s *Store) Login(ctx context.Context, usuario, clave string) (string, Usuario, error) {
	var row struct {
		ID        int64
		Usuario   string
		Nombre    string
		Rol       string
		CapatazID *string
		Hash      string
		Activo    bool
	}
	err := s.db.WithContext(ctx).Raw(`
		SELECT id, usuario, nombre, rol, capataz_id, password_hash, activo
		FROM usuarios WHERE usuario = $1`, strings.TrimSpace(usuario)).Row().Scan(
		&row.ID, &row.Usuario, &row.Nombre, &row.Rol, &row.CapatazID, &row.Hash, &row.Activo,
	)
	if err != nil || !row.Activo || bcrypt.CompareHashAndPassword([]byte(row.Hash), []byte(clave)) != nil {
		return "", Usuario{}, errors.New("credenciales")
	}
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return "", Usuario{}, err
	}
	token := hex.EncodeToString(raw)
	sum := sha256.Sum256([]byte(token))
	if err := s.db.WithContext(ctx).Exec(`
		INSERT INTO sesiones (token_hash, usuario_id, expires_at)
		VALUES ($1, $2, now() + interval '12 hours')`,
		hex.EncodeToString(sum[:]), row.ID,
	).Error; err != nil {
		return "", Usuario{}, err
	}
	u := Usuario{ID: row.ID, Usuario: row.Usuario, Nombre: row.Nombre, Rol: row.Rol}
	if row.CapatazID != nil {
		u.CapatazID = *row.CapatazID
	}
	return token, u, nil
}

// FromToken resuelve la cookie. El token en claro no se guarda.
func (s *Store) FromToken(ctx context.Context, token string) (Usuario, error) {
	sum := sha256.Sum256([]byte(token))
	var u Usuario
	var cap *string
	err := s.db.WithContext(ctx).Raw(`
		SELECT u.id, u.usuario, u.nombre, u.rol, u.capataz_id
		FROM sesiones s
		JOIN usuarios u ON u.id = s.usuario_id
		WHERE s.token_hash = $1 AND s.expires_at > now() AND u.activo`,
		hex.EncodeToString(sum[:]),
	).Row().Scan(&u.ID, &u.Usuario, &u.Nombre, &u.Rol, &cap)
	if err != nil {
		return Usuario{}, err
	}
	if cap != nil {
		u.CapatazID = *cap
	}
	return u, nil
}

// Logout borra la sesión de esa cookie.
func (s *Store) Logout(ctx context.Context, token string) {
	if token == "" {
		return
	}
	sum := sha256.Sum256([]byte(token))
	_ = s.db.WithContext(ctx).Exec(`DELETE FROM sesiones WHERE token_hash = $1`, hex.EncodeToString(sum[:])).Error
}

// Listado de cuentas, sin hash.
func (s *Store) Usuarios(ctx context.Context) ([]Usuario, error) {
	rows, err := s.db.WithContext(ctx).Raw(`
		SELECT id, usuario, nombre, rol, COALESCE(capataz_id, '')
		FROM usuarios ORDER BY rol, usuario`).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []Usuario{}
	for rows.Next() {
		var u Usuario
		if err := rows.Scan(&u.ID, &u.Usuario, &u.Nombre, &u.Rol, &u.CapatazID); err != nil {
			return nil, err
		}
		out = append(out, u)
	}
	return out, rows.Err()
}

type Permiso struct {
	Rol    string `json:"rol"`
	Accion string `json:"accion"`
}

func (s *Store) Permisos(ctx context.Context) ([]Permiso, error) {
	var out []Permiso
	err := s.db.WithContext(ctx).Raw(`SELECT rol, accion FROM permisos ORDER BY rol, accion`).Scan(&out).Error
	if out == nil {
		out = []Permiso{}
	}
	return out, err
}

// Token de la cookie, si vino.
func Token(r *http.Request) string {
	c, err := r.Cookie(Cookie)
	if err != nil {
		return ""
	}
	return c.Value
}
