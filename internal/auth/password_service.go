package auth

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"github.com/walterfan/lazy-ai-coder/pkg/models"
)

var (
	ErrWeakPassword    = errors.New("password is too weak")
	ErrInvalidUsername = errors.New("invalid username format")
	ErrInvalidEmail    = errors.New("invalid email format")
	ErrUsernameTaken   = errors.New("username already taken")
	ErrEmailTaken      = errors.New("email already taken")

	// Username validation: alphanumeric and underscores, 3-20 characters
	usernameRegex = regexp.MustCompile(`^[a-zA-Z0-9_]{3,20}$`)

	// Email validation: basic email format
	emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
)

// PasswordAuthService handles username/password authentication
type PasswordAuthService struct {
	db *gorm.DB
}

// NewPasswordAuthService creates a new password auth service
func NewPasswordAuthService(db *gorm.DB) *PasswordAuthService {
	return &PasswordAuthService{db: db}
}

// ValidatePassword checks if password meets minimum requirements
func ValidatePassword(password string) error {
	if len(password) < 8 {
		return fmt.Errorf("%w: password must be at least 8 characters", ErrWeakPassword)
	}
	// Additional checks can be added here:
	// - Require uppercase, lowercase, numbers, symbols
	// - Check against common password list
	return nil
}

// ValidateUsername checks if username meets format requirements
func ValidateUsername(username string) error {
	if !usernameRegex.MatchString(username) {
		return fmt.Errorf("%w: username must be 3-20 characters, alphanumeric and underscores only", ErrInvalidUsername)
	}
	return nil
}

// ValidateEmail checks if email is valid format
func ValidateEmail(email string) error {
	if !emailRegex.MatchString(email) {
		return fmt.Errorf("%w: email must be a valid email address", ErrInvalidEmail)
	}
	return nil
}

// HashPassword hashes a password using bcrypt
func HashPassword(password string) (string, error) {
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}
	return string(hashedBytes), nil
}

// ComparePassword compares a password with its hash
func ComparePassword(hashedPassword, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}

// SignUp creates a new user with username/password
func (s *PasswordAuthService) SignUp(username, email, password, name string) (*models.User, error) {
	// Normalize username and email
	username = strings.ToLower(strings.TrimSpace(username))
	email = strings.ToLower(strings.TrimSpace(email))
	name = strings.TrimSpace(name)

	// Validate inputs
	if err := ValidateUsername(username); err != nil {
		return nil, err
	}

	if err := ValidateEmail(email); err != nil {
		return nil, err
	}

	if err := ValidatePassword(password); err != nil {
		return nil, err
	}

	// Check if username already exists
	var existingUser models.User
	result := s.db.Where("LOWER(username) = ?", username).First(&existingUser)
	if result.Error == nil {
		return nil, ErrUsernameTaken
	} else if !errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("failed to check username: %w", result.Error)
	}

	// Check if email already exists
	result = s.db.Where("LOWER(email) = ?", email).First(&existingUser)
	if result.Error == nil {
		return nil, ErrEmailTaken
	} else if !errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("failed to check email: %w", result.Error)
	}

	// Hash password
	hashedPassword, err := HashPassword(password)
	if err != nil {
		return nil, err
	}

	// Create new user (inactive until admin approval, no realm assigned yet)
	now := time.Now()
	newUser := &models.User{
		ID:             uuid.New().String(),
		RealmID:        "", // No realm assigned yet - will be assigned by admin on approval
		Username:       username,
		Email:          email,
		Name:           name,
		HashedPassword: &hashedPassword,
		IsActive:       false, // Requires admin approval
		LastLoginAt:    nil,   // Not logged in yet
		CreatedBy:      username,
		CreatedTime:    now,
		UpdatedBy:      username,
		UpdatedTime:    now,
	}

	if err := s.db.Create(newUser).Error; err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return newUser, nil
}

// GetUserByID retrieves a user by their ID
func (s *PasswordAuthService) GetUserByID(userID string) (*models.User, error) {
	var user models.User
	result := s.db.Where("id = ? AND is_active = ?", userID, true).First(&user)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get user: %w", result.Error)
	}
	return &user, nil
}

// SignIn authenticates a user with username/password
func (s *PasswordAuthService) SignIn(username, password string) (*models.User, error) {
	// Normalize username
	username = strings.ToLower(strings.TrimSpace(username))

	// Find user by username (check both active and inactive)
	var user models.User
	result := s.db.Where("LOWER(username) = ?", username).First(&user)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, ErrInvalidCredentials
		}
		return nil, fmt.Errorf("failed to find user: %w", result.Error)
	}

	// Check if user account is pending approval
	if !user.IsActive {
		return nil, fmt.Errorf("your account is pending admin approval. Please wait for approval before signing in")
	}

	// Check if user has a password (might be OAuth-only user)
	if user.HashedPassword == nil {
		return nil, fmt.Errorf("this account uses OAuth authentication. Please sign in with GitLab")
	}

	// Verify password
	if err := ComparePassword(*user.HashedPassword, password); err != nil {
		return nil, ErrInvalidCredentials
	}

	// Update last login time
	now := time.Now()
	user.LastLoginAt = &now
	s.db.Model(&user).Update("last_login_at", now)

	return &user, nil
}

// ChangePassword changes a user's password
func (s *PasswordAuthService) ChangePassword(userID, oldPassword, newPassword string) error {
	// Find user
	var user models.User
	result := s.db.Where("id = ? AND is_active = ?", userID, true).First(&user)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return ErrUserNotFound
		}
		return fmt.Errorf("failed to find user: %w", result.Error)
	}

	// Check if user has a password
	if user.HashedPassword == nil {
		return fmt.Errorf("this account uses OAuth authentication and has no password")
	}

	// Verify old password
	if err := ComparePassword(*user.HashedPassword, oldPassword); err != nil {
		return ErrInvalidCredentials
	}

	// Validate new password
	if err := ValidatePassword(newPassword); err != nil {
		return err
	}

	// Hash new password
	hashedPassword, err := HashPassword(newPassword)
	if err != nil {
		return err
	}

	// Update password
	user.HashedPassword = &hashedPassword
	user.UpdatedTime = time.Now()
	if err := s.db.Save(&user).Error; err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	return nil
}
