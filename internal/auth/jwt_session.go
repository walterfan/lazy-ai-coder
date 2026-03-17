package auth

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// SessionClaims represents JWT claims for OAuth session tokens
type SessionClaims struct {
	UserID   string `json:"user_id"`
	RealmID  string `json:"realm_id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	jwt.RegisteredClaims
}

// SessionJWTService handles session JWT token operations using HMAC
type SessionJWTService struct {
	secretKey []byte
	issuer    string
	audience  string
}

// NewSessionJWTService creates a new session JWT service
func NewSessionJWTService(secret, issuer, audience string) *SessionJWTService {
	return &SessionJWTService{
		secretKey: []byte(secret),
		issuer:    issuer,
		audience:  audience,
	}
}

// GenerateToken creates a new JWT session token
func (s *SessionJWTService) GenerateToken(userID, realmID, username, email string, expiration time.Duration) (string, error) {
	now := time.Now()
	claims := &SessionClaims{
		UserID:   userID,
		RealmID:  realmID,
		Username: username,
		Email:    email,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    s.issuer,
			Subject:   userID,
			Audience:  []string{s.audience},
			ExpiresAt: jwt.NewNumericDate(now.Add(expiration)),
			NotBefore: jwt.NewNumericDate(now),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.secretKey)
}

// ValidateToken validates and parses a JWT token
func (s *SessionJWTService) ValidateToken(tokenString string) (*SessionClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &SessionClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Verify signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return s.secretKey, nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	if claims, ok := token.Claims.(*SessionClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("invalid token claims")
}

// RefreshToken creates a new token with extended expiration
func (s *SessionJWTService) RefreshToken(oldToken string, expiration time.Duration) (string, error) {
	claims, err := s.ValidateToken(oldToken)
	if err != nil {
		return "", fmt.Errorf("invalid token: %w", err)
	}

	// Create new token with same claims but new expiration
	return s.GenerateToken(claims.UserID, claims.RealmID, claims.Username, claims.Email, expiration)
}
