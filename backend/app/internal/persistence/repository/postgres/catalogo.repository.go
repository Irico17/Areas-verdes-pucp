package postgres

import (
	"context"
	"database/sql"
	"errors"
	"strconv"

	"gorm.io/gorm"

	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/contracts"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/entities"
	apperrors "github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/domain/errors"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/persistence/mapper"
	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/persistence/models"
)

type catalogoRepository struct {
	db *gorm.DB
}

// NewCatalogoRepository creates a new postgres repository for catalog items.
func NewCatalogoRepository(db *gorm.DB) contracts.ICatalogoRepository {
	return &catalogoRepository{db: db}
}

func (r *catalogoRepository) List(ctx context.Context, clase string, soloActivos bool) ([]entities.CatalogoItem, error) {
	q := `SELECT id, clase, codigo, nombre, activo, orden, provisional FROM catalogos WHERE 1=1`
	args := []any{}
	if clase != "" {
		q += ` AND clase = $1`
		args = append(args, clase)
	}
	if soloActivos {
		q += ` AND activo`
	}
	q += ` ORDER BY clase, orden, codigo`

	var rows []models.CatalogoModel
	err := r.db.WithContext(ctx).Raw(q, args...).Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	out := make([]entities.CatalogoItem, 0, len(rows))
	for _, m := range rows {
		out = append(out, *mapper.CatalogoToEntity(&m))
	}
	return out, nil
}

func (r *catalogoRepository) Activo(ctx context.Context, clase, codigo string) (bool, error) {
	var n int
	err := r.db.WithContext(ctx).Raw(`
		SELECT count(*) FROM catalogos WHERE clase = $1 AND codigo = $2 AND activo`,
		clase, codigo).Scan(&n).Error
	return n == 1, err
}

func (r *catalogoRepository) Create(ctx context.Context, clase, codigo, nombre string) (*entities.CatalogoItem, error) {
	var item *entities.CatalogoItem
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		prev, found, err := leerCatalogoPorClave(tx, clase, codigo)
		if err != nil {
			return err
		}
		var m models.CatalogoModel
		err = tx.Raw(`
			INSERT INTO catalogos (clase, codigo, nombre)
			VALUES ($1, $2, $3)
			ON CONFLICT (clase, codigo) DO UPDATE SET nombre = EXCLUDED.nombre, activo = TRUE
			RETURNING id, clase, codigo, nombre, activo, orden, provisional`,
			clase, codigo, nombre).Row().Scan(&m.ID, &m.Clase, &m.Codigo, &m.Nombre, &m.Activo, &m.Orden, &m.Provisional)
		if err != nil {
			return err
		}
		if found && prev.Nombre == m.Nombre && prev.Activo == m.Activo {
			item = mapper.CatalogoToEntity(&m)
			return nil
		}
		accion := "alta"
		var antes any
		if found {
			accion = "edicion"
			antes = snapshotCatalogo(prev)
		}
		if err := registrarCambioTx(tx, "catalogos", strconv.FormatInt(m.ID, 10), accion, antes, snapshotCatalogo(m), nil); err != nil {
			return err
		}
		item = mapper.CatalogoToEntity(&m)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return item, nil
}

func (r *catalogoRepository) Deactivate(ctx context.Context, id int64, usuarioID int64) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		m, found, err := leerCatalogoPorID(tx, id)
		if err != nil {
			return err
		}
		if !found {
			return apperrors.ErrItemNoExiste
		}
		if !m.Activo {
			return nil
		}
		if err := tx.Exec(`UPDATE catalogos SET activo = FALSE WHERE id = $1`, id).Error; err != nil {
			return err
		}
		despues := snapshotCatalogo(m)
		despues["activo"] = false
		return registrarCambioTx(tx, "catalogos", strconv.FormatInt(id, 10), "baja", snapshotCatalogo(m), despues, usuarioOpcional(usuarioID))
	})
}

func (r *catalogoRepository) Renombrar(ctx context.Context, id int64, nombre string, usuarioID int64) (*entities.CatalogoItem, error) {
	var item *entities.CatalogoItem
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		m, found, err := leerCatalogoPorID(tx, id)
		if err != nil {
			return err
		}
		if !found {
			return apperrors.ErrItemNoExiste
		}
		if m.Nombre == nombre {
			item = mapper.CatalogoToEntity(&m)
			return nil
		}
		antes := snapshotCatalogo(m)
		if err := tx.Exec(`UPDATE catalogos SET nombre = $2 WHERE id = $1`, id, nombre).Error; err != nil {
			return err
		}
		m.Nombre = nombre
		if err := registrarCambioTx(tx, "catalogos", strconv.FormatInt(id, 10), "edicion", antes, snapshotCatalogo(m), usuarioOpcional(usuarioID)); err != nil {
			return err
		}
		item = mapper.CatalogoToEntity(&m)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return item, nil
}

func leerCatalogoPorID(tx *gorm.DB, id int64) (models.CatalogoModel, bool, error) {
	var m models.CatalogoModel
	err := tx.Raw(`
		SELECT id, clase, codigo, nombre, activo, orden, provisional
		FROM catalogos WHERE id = $1`, id).Row().Scan(
		&m.ID, &m.Clase, &m.Codigo, &m.Nombre, &m.Activo, &m.Orden, &m.Provisional,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return models.CatalogoModel{}, false, nil
	}
	if err != nil {
		return models.CatalogoModel{}, false, err
	}
	return m, true, nil
}

func leerCatalogoPorClave(tx *gorm.DB, clase, codigo string) (models.CatalogoModel, bool, error) {
	var m models.CatalogoModel
	err := tx.Raw(`
		SELECT id, clase, codigo, nombre, activo, orden, provisional
		FROM catalogos WHERE clase = $1 AND codigo = $2`, clase, codigo).Row().Scan(
		&m.ID, &m.Clase, &m.Codigo, &m.Nombre, &m.Activo, &m.Orden, &m.Provisional,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return models.CatalogoModel{}, false, nil
	}
	if err != nil {
		return models.CatalogoModel{}, false, err
	}
	return m, true, nil
}

func snapshotCatalogo(m models.CatalogoModel) map[string]any {
	return map[string]any{
		"clase":       m.Clase,
		"codigo":      m.Codigo,
		"nombre":      m.Nombre,
		"activo":      m.Activo,
		"orden":       m.Orden,
		"provisional": m.Provisional,
	}
}

func usuarioOpcional(usuarioID int64) *int64 {
	if usuarioID <= 0 {
		return nil
	}
	id := usuarioID
	return &id
}
