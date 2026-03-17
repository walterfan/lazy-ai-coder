package models

import (
	"time"

	"gorm.io/gorm"
)

// User represents the app_user table in the database
type User struct {
	ID                  string         `json:"id" gorm:"primaryKey;type:text"`
	RealmID             string         `json:"realm_id" gorm:"not null;type:text;index"`
	Username            string         `json:"username" gorm:"not null;type:text;index"`
	Email               string         `json:"email" gorm:"unique;type:text;index"`
	HashedPassword      *string        `json:"-" gorm:"column:hashed_password;type:text"` // Nullable for OAuth users
	IsActive            bool           `json:"is_active" gorm:"default:true"`
	// OAuth fields
	GitLabUserID        *int           `json:"gitlab_user_id" gorm:"unique;column:gitlab_user_id"` // GitLab user ID
	Name                string         `json:"name" gorm:"type:text"`                               // Full name from GitLab
	AvatarURL           string         `json:"avatar_url" gorm:"column:avatar_url;type:text"`       // Avatar URL
	GitLabAccessToken   *string        `json:"-" gorm:"column:gitlab_access_token;type:text"`       // Encrypted OAuth token
	GitLabRefreshToken  *string        `json:"-" gorm:"column:gitlab_refresh_token;type:text"`      // Encrypted refresh token
	TokenExpiresAt      *time.Time     `json:"token_expires_at" gorm:"column:token_expires_at"`     // Token expiration
	LastLoginAt         *time.Time     `json:"last_login_at" gorm:"column:last_login_at"`           // Last login time
	// Audit fields
	CreatedBy           string         `json:"created_by" gorm:"type:text"`
	CreatedTime         time.Time      `json:"created_time" gorm:"autoCreateTime"`
	UpdatedBy           string         `json:"updated_by" gorm:"type:text"`
	UpdatedTime         time.Time      `json:"updated_time" gorm:"autoUpdateTime"`
	DeletedAt           gorm.DeletedAt `gorm:"index" json:"-"`
}

func (User) TableName() string {
	return "app_user"
}
