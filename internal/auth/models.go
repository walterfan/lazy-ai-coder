package auth

import (
	"time"

	"github.com/google/uuid"
)

// Realm represents a multi-tenant organization
type Realm struct {
	ID          uuid.UUID `json:"id" db:"id"`
	Name        string    `json:"name" db:"name"`
	Description string    `json:"description" db:"description"`
	CreatedBy   uuid.UUID `json:"created_by" db:"created_by"`
	CreatedTime time.Time `json:"created_time" db:"created_time"`
	UpdatedBy   uuid.UUID `json:"updated_by" db:"updated_by"`
	UpdatedTime time.Time `json:"updated_time" db:"updated_time"`
}

// User represents an application user
type User struct {
	ID             uuid.UUID `json:"id" db:"id"`
	RealmID        uuid.UUID `json:"realm_id" db:"realm_id"`
	Username       string    `json:"username" db:"username"`
	Email          string    `json:"email" db:"email"`
	HashedPassword string    `json:"-" db:"hashed_password"`
	IsActive       bool      `json:"is_active" db:"is_active"`
	CreatedBy      uuid.UUID `json:"created_by" db:"created_by"`
	CreatedTime    time.Time `json:"created_time" db:"created_time"`
	UpdatedBy      uuid.UUID `json:"updated_by" db:"updated_by"`
	UpdatedTime    time.Time `json:"updated_time" db:"updated_time"`
}

// Role represents a user role within a realm
type Role struct {
	ID          uuid.UUID `json:"id" db:"id"`
	RealmID     uuid.UUID `json:"realm_id" db:"realm_id"`
	Name        string    `json:"name" db:"name"`
	Description string    `json:"description" db:"description"`
	CreatedBy   uuid.UUID `json:"created_by" db:"created_by"`
	CreatedTime time.Time `json:"created_time" db:"created_time"`
	UpdatedBy   uuid.UUID `json:"updated_by" db:"updated_by"`
	UpdatedTime time.Time `json:"updated_time" db:"updated_time"`
}

// Policy represents an AWS-style access policy
type Policy struct {
	ID          uuid.UUID `json:"id" db:"id"`
	RealmID     uuid.UUID `json:"realm_id" db:"realm_id"`
	Name        string    `json:"name" db:"name"`
	Description string    `json:"description" db:"description"`
	Version     string    `json:"version" db:"version"`
	CreatedBy   uuid.UUID `json:"created_by" db:"created_by"`
	CreatedTime time.Time `json:"created_time" db:"created_time"`
	UpdatedBy   uuid.UUID `json:"updated_by" db:"updated_by"`
	UpdatedTime time.Time `json:"updated_time" db:"updated_time"`
}

// Statement represents a policy statement (Allow/Deny)
type Statement struct {
	ID          uuid.UUID              `json:"id" db:"id"`
	PolicyID    uuid.UUID              `json:"policy_id" db:"policy_id"`
	SID         string                 `json:"sid" db:"sid"`
	Effect      string                 `json:"effect" db:"effect"`
	Actions     []string               `json:"actions" db:"actions"`
	Resources   []string               `json:"resources" db:"resources"`
	Conditions  map[string]interface{} `json:"conditions" db:"conditions"`
	CreatedBy   uuid.UUID              `json:"created_by" db:"created_by"`
	CreatedTime time.Time              `json:"created_time" db:"created_time"`
	UpdatedBy   uuid.UUID              `json:"updated_by" db:"updated_by"`
	UpdatedTime time.Time              `json:"updated_time" db:"updated_time"`
}

// UserRole represents the many-to-many relationship between users and roles
type UserRole struct {
	UserID uuid.UUID `json:"user_id" db:"user_id"`
	RoleID uuid.UUID `json:"role_id" db:"role_id"`
}

// RolePolicy represents the many-to-many relationship between roles and policies
type RolePolicy struct {
	RoleID   uuid.UUID `json:"role_id" db:"role_id"`
	PolicyID uuid.UUID `json:"policy_id" db:"policy_id"`
}

// UserPolicy represents direct policy attachments to users
type UserPolicy struct {
	UserID   uuid.UUID `json:"user_id" db:"user_id"`
	PolicyID uuid.UUID `json:"policy_id" db:"policy_id"`
}

// ResourcePolicy represents resource-based policies
type ResourcePolicy struct {
	ID           uuid.UUID `json:"id" db:"id"`
	RealmID      uuid.UUID `json:"realm_id" db:"realm_id"`
	ResourceType string    `json:"resource_type" db:"resource_type"`
	ResourceID   uuid.UUID `json:"resource_id" db:"resource_id"`
	PolicyID     uuid.UUID `json:"policy_id" db:"policy_id"`
	CreatedBy    uuid.UUID `json:"created_by" db:"created_by"`
	CreatedTime  time.Time `json:"created_time" db:"created_time"`
	UpdatedBy    uuid.UUID `json:"updated_by" db:"updated_by"`
	UpdatedTime  time.Time `json:"updated_time" db:"updated_time"`
}

// LoginRequest represents a login request
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
	RealmID  string `json:"realm_id" binding:"required"`
}

// LoginResponse represents a successful login response
type LoginResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int64  `json:"expires_in"`
	User         User   `json:"user"`
}

// RefreshRequest represents a token refresh request
type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// CreateUserRequest represents a user creation request
type CreateUserRequest struct {
	Username string `json:"username" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
	RealmID  string `json:"realm_id" binding:"required"`
}

// UpdateUserRequest represents a user update request
type UpdateUserRequest struct {
	Username string `json:"username"`
	Email    string `json:"email" binding:"omitempty,email"`
	IsActive *bool  `json:"is_active"`
}

// PermissionCheck represents a permission check request
type PermissionCheck struct {
	Action   string                 `json:"action" binding:"required"`
	Resource string                 `json:"resource" binding:"required"`
	UserID   uuid.UUID              `json:"user_id" binding:"required"`
	RealmID  uuid.UUID              `json:"realm_id" binding:"required"`
	Context  map[string]interface{} `json:"context,omitempty"`
}
