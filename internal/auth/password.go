package auth

import (
	"golang.org/x/crypto/bcrypt"
)

// PasswordManager handles password hashing and verification
type PasswordManager struct {
	cost int
}

// NewPasswordManager creates a new password manager with specified bcrypt cost
func NewPasswordManager(cost int) *PasswordManager {
	if cost < bcrypt.MinCost {
		cost = bcrypt.DefaultCost
	}
	return &PasswordManager{cost: cost}
}

// HashPassword creates a bcrypt hash of the password
func (p *PasswordManager) HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), p.cost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

// VerifyPassword checks if a password matches its hash
func (p *PasswordManager) VerifyPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// CheckPasswordStrength validates password requirements
func (p *PasswordManager) CheckPasswordStrength(password string) error {
	if len(password) < 8 {
		return ErrPasswordTooShort
	}

	// Add more password strength checks as needed
	// e.g., require uppercase, lowercase, numbers, special characters

	return nil
}
