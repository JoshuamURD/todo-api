package hash

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

// Hasher defines the interface for password hashing operations
// It is used to hash and compare passwords
type Hasher interface {
	Hash(password string) (string, error)
	Compare(hashedPassword, plainPassword string) bool
}

// bcryptHasher implements Hasher interface
type bcryptHasher struct {
	cost int
}

// NewBcryptHasher creates a new instance of bcryptHasher
func NewBcryptHasher(cost int) Hasher {
	return &bcryptHasher{
		cost: cost,
	}
}

// Hash implements Hasher.Hash
func (b *bcryptHasher) Hash(password string) (string, error) {
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), b.cost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}
	return string(hashedBytes), nil
}

// Compare implements Hasher.Compare
func (b *bcryptHasher) Compare(hashedPassword, plainPassword string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(plainPassword))
	return err == nil
}
