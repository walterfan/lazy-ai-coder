package auth

import (
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/walterfan/lazy-ai-coder/pkg/models"
)

// OAuthUserService handles OAuth-specific user database operations
type OAuthUserService struct {
	db *gorm.DB
}

// NewOAuthUserService creates a new OAuth user service
func NewOAuthUserService(db *gorm.DB) *OAuthUserService {
	return &OAuthUserService{db: db}
}

// GetUserByID retrieves a user by their ID
func (s *OAuthUserService) GetUserByID(userID string) (*models.User, error) {
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

// GetUserByGitLabID retrieves a user by their GitLab user ID
func (s *OAuthUserService) GetUserByGitLabID(gitlabUserID int) (*models.User, error) {
	var user models.User
	result := s.db.Where("gitlab_user_id = ? AND is_active = ?", gitlabUserID, true).First(&user)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get user by GitLab ID: %w", result.Error)
	}
	return &user, nil
}

// GetUserByEmail retrieves a user by their email
func (s *OAuthUserService) GetUserByEmail(email string) (*models.User, error) {
	var user models.User
	result := s.db.Where("email = ? AND is_active = ?", email, true).First(&user)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get user by email: %w", result.Error)
	}
	return &user, nil
}

// CreateOrUpdateOAuthUser creates a new user from OAuth data or updates existing user
func (s *OAuthUserService) CreateOrUpdateOAuthUser(gitlabUserID int, username, email, name, avatarURL, accessToken, refreshToken string, tokenExpiresAt time.Time, realmID string) (*models.User, error) {
	// Try to find existing user by GitLab ID
	existingUser, err := s.GetUserByGitLabID(gitlabUserID)

	now := time.Now()

	if err != nil && !errors.Is(err, ErrUserNotFound) {
		return nil, fmt.Errorf("failed to check existing user: %w", err)
	}

	if existingUser != nil {
		// Update existing user
		existingUser.Username = username
		existingUser.Email = email
		existingUser.Name = name
		existingUser.AvatarURL = avatarURL
		existingUser.GitLabAccessToken = &accessToken
		existingUser.GitLabRefreshToken = &refreshToken
		existingUser.TokenExpiresAt = &tokenExpiresAt
		existingUser.LastLoginAt = &now
		existingUser.UpdatedBy = username
		existingUser.UpdatedTime = now

		if err := s.db.Save(existingUser).Error; err != nil {
			return nil, fmt.Errorf("failed to update user: %w", err)
		}

		return existingUser, nil
	}

	// Create new user (inactive until admin approval, no realm assigned yet)
	newUser := &models.User{
		ID:                 uuid.New().String(),
		RealmID:            "", // No realm assigned yet - will be assigned by admin on approval
		GitLabUserID:       &gitlabUserID,
		Username:           username,
		Email:              email,
		Name:               name,
		AvatarURL:          avatarURL,
		GitLabAccessToken:  &accessToken,
		GitLabRefreshToken: &refreshToken,
		TokenExpiresAt:     &tokenExpiresAt,
		LastLoginAt:        nil,   // Not logged in yet until approved
		IsActive:           false, // Requires admin approval
		CreatedBy:          username,
		CreatedTime:        now,
		UpdatedBy:          username,
		UpdatedTime:        now,
	}

	if err := s.db.Create(newUser).Error; err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return newUser, nil
}

// UpdateLastLogin updates the user's last login timestamp
func (s *OAuthUserService) UpdateLastLogin(userID string) error {
	now := time.Now()
	result := s.db.Model(&models.User{}).
		Where("id = ?", userID).
		Update("last_login_at", now)

	if result.Error != nil {
		return fmt.Errorf("failed to update last login: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return ErrUserNotFound
	}

	return nil
}

// UpdateOAuthTokens updates user's OAuth access and refresh tokens
func (s *OAuthUserService) UpdateOAuthTokens(userID, accessToken, refreshToken string, expiresAt time.Time) error {
	result := s.db.Model(&models.User{}).
		Where("id = ?", userID).
		Updates(map[string]interface{}{
			"gitlab_access_token":  accessToken,
			"gitlab_refresh_token": refreshToken,
			"token_expires_at":     expiresAt,
			"updated_time":         time.Now(),
		})

	if result.Error != nil {
		return fmt.Errorf("failed to update OAuth tokens: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return ErrUserNotFound
	}

	return nil
}

// DeactivateUser soft-deletes a user by setting is_active to false
func (s *OAuthUserService) DeactivateUser(userID string) error {
	result := s.db.Model(&models.User{}).
		Where("id = ?", userID).
		Update("is_active", false)

	if result.Error != nil {
		return fmt.Errorf("failed to deactivate user: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return ErrUserNotFound
	}

	return nil
}

// GetActiveUsersInRealm retrieves all active users in a realm
func (s *OAuthUserService) GetActiveUsersInRealm(realmID string) ([]models.User, error) {
	var users []models.User
	result := s.db.Where("realm_id = ? AND is_active = ?", realmID, true).Find(&users)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to get users in realm: %w", result.Error)
	}
	return users, nil
}
