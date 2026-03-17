package auth

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// AuthService handles authentication and user management
type AuthService struct {
	userService      UserService
	passwordManager  *PasswordManager
	jwtManager       *JWTManager
	permissionEngine *PermissionEngine
}

// NewAuthService creates a new authentication service
func NewAuthService(
	userService UserService,
	passwordManager *PasswordManager,
	jwtManager *JWTManager,
	permissionEngine *PermissionEngine,
) *AuthService {
	return &AuthService{
		userService:      userService,
		passwordManager:  passwordManager,
		jwtManager:       jwtManager,
		permissionEngine: permissionEngine,
	}
}

// Login authenticates a user and returns JWT tokens
func (a *AuthService) Login(req LoginRequest) (*LoginResponse, error) {
	// Parse realm ID
	realmID, err := uuid.Parse(req.RealmID)
	if err != nil {
		return nil, ErrInvalidRealm
	}

	// Get user by username and realm
	user, err := a.userService.GetUserByUsername(req.Username, realmID)
	if err != nil {
		return nil, ErrUserNotFound
	}

	// Check if user is active
	if !user.IsActive {
		return nil, ErrUserInactive
	}

	// Verify password
	if !a.passwordManager.VerifyPassword(req.Password, user.HashedPassword) {
		return nil, ErrInvalidCredentials
	}

	// Get user roles
	roles, err := a.userService.GetUserRoles(user.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user roles: %w", err)
	}

	roleNames := make([]string, len(roles))
	for i, role := range roles {
		roleNames[i] = role.Name
	}

	// Generate tokens
	accessToken, err := a.jwtManager.GenerateToken(
		user.ID,
		user.RealmID,
		user.Username,
		user.Email,
		roleNames,
		15*time.Minute, // Access token expires in 15 minutes
	)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	refreshToken, err := a.jwtManager.GenerateToken(
		user.ID,
		user.RealmID,
		user.Username,
		user.Email,
		roleNames,
		7*24*time.Hour, // Refresh token expires in 7 days
	)
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	return &LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    15 * 60, // 15 minutes in seconds
		User:         *user,
	}, nil
}

// RefreshToken refreshes an access token using a refresh token
func (a *AuthService) RefreshToken(req RefreshRequest) (*LoginResponse, error) {
	// Validate refresh token
	claims, err := a.jwtManager.ValidateToken(req.RefreshToken)
	if err != nil {
		return nil, ErrInvalidToken
	}

	// Get user
	user, err := a.userService.GetUserByID(claims.UserID)
	if err != nil {
		return nil, ErrUserNotFound
	}

	// Check if user is active
	if !user.IsActive {
		return nil, ErrUserInactive
	}

	// Get user roles
	roles, err := a.userService.GetUserRoles(user.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user roles: %w", err)
	}

	roleNames := make([]string, len(roles))
	for i, role := range roles {
		roleNames[i] = role.Name
	}

	// Generate new access token
	accessToken, err := a.jwtManager.GenerateToken(
		user.ID,
		user.RealmID,
		user.Username,
		user.Email,
		roleNames,
		15*time.Minute,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	return &LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: req.RefreshToken, // Keep the same refresh token
		TokenType:    "Bearer",
		ExpiresIn:    15 * 60,
		User:         *user,
	}, nil
}

// RegisterUser creates a new user account
func (a *AuthService) RegisterUser(req CreateUserRequest, createdBy uuid.UUID) (*User, error) {
	// Parse realm ID
	realmID, err := uuid.Parse(req.RealmID)
	if err != nil {
		return nil, ErrInvalidRealm
	}

	// Check password strength
	if err := a.passwordManager.CheckPasswordStrength(req.Password); err != nil {
		return nil, err
	}

	// Check if username already exists in the realm
	existingUser, _ := a.userService.GetUserByUsername(req.Username, realmID)
	if existingUser != nil {
		return nil, ErrUserAlreadyExists
	}

	// Hash password
	hashedPassword, err := a.passwordManager.HashPassword(req.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Create user
	user := &User{
		ID:             uuid.New(),
		RealmID:        realmID,
		Username:       req.Username,
		Email:          req.Email,
		HashedPassword: hashedPassword,
		IsActive:       true,
		CreatedBy:      createdBy,
		CreatedTime:    time.Now(),
		UpdatedBy:      createdBy,
		UpdatedTime:    time.Now(),
	}

	if err := a.userService.CreateUser(user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return user, nil
}

// CheckPermission checks if a user has permission to perform an action
func (a *AuthService) CheckPermission(userID, realmID uuid.UUID, action, resource string, context map[string]interface{}) (bool, error) {
	check := PermissionCheck{
		Action:   action,
		Resource: resource,
		UserID:   userID,
		RealmID:  realmID,
		Context:  context,
	}

	return a.permissionEngine.CheckPermission(check)
}

// ValidateToken validates a JWT token and returns the claims
func (a *AuthService) ValidateToken(tokenString string) (*JWTClaims, error) {
	return a.jwtManager.ValidateToken(tokenString)
}

// UserService interface for dependency injection
type UserService interface {
	GetUserByID(userID uuid.UUID) (*User, error)
	GetUserByUsername(username string, realmID uuid.UUID) (*User, error)
	GetUserRoles(userID uuid.UUID) ([]*Role, error)
	CreateUser(user *User) error
	UpdateUser(user *User) error
	DeleteUser(userID uuid.UUID) error
}
