// Package seguridad provides security adapters such as password hashing.
package seguridad

import (
	"golang.org/x/crypto/bcrypt"

	"github.com/GRUPO-12-DP2/-areas-verdes-pucp/backend/internal/application/contracts"
)

type bcryptHasher struct{}

// NewBcryptHasher creates a password hasher backed by bcrypt.
func NewBcryptHasher() contracts.IHasher {
	return &bcryptHasher{}
}

func (h *bcryptHasher) Hash(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

func (h *bcryptHasher) Compare(hashedPassword, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}
