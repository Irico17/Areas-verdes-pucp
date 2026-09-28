package contracts

// IHasher hashes and verifies passwords.
type IHasher interface {
	Hash(password string) (string, error)
	Compare(hashedPassword, password string) error
}
