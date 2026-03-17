package auth

import "errors"

// Authentication and authorization errors
var (
	ErrInvalidCredentials      = errors.New("invalid credentials")
	ErrUserNotFound            = errors.New("user not found")
	ErrUserInactive            = errors.New("user is inactive")
	ErrInvalidToken            = errors.New("invalid token")
	ErrExpiredToken            = errors.New("token has expired")
	ErrInsufficientPermissions = errors.New("insufficient permissions")
	ErrRealmNotFound           = errors.New("realm not found")
	ErrPasswordTooShort        = errors.New("password too short")
	ErrUserAlreadyExists       = errors.New("user already exists")
	ErrInvalidRealm            = errors.New("invalid realm")
	ErrAccessDenied            = errors.New("access denied")
)
