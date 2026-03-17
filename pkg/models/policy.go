package models

import (
	"time"

	"gorm.io/gorm"
)

// Policy represents the policy table in the database
type Policy struct {
	ID          string         `json:"id" gorm:"primaryKey;type:text"`
	RealmID     string         `json:"realm_id" gorm:"not null;type:text;index"`
	Name        string         `json:"name" gorm:"not null;type:text"`
	Description string         `json:"description" gorm:"type:text"`
	Version     string         `json:"version" gorm:"default:'2012-10-17';type:text"` // AWS policy version format
	CreatedBy   string         `json:"created_by" gorm:"type:text"`
	CreatedTime time.Time      `json:"created_time" gorm:"autoCreateTime"`
	UpdatedBy   string         `json:"updated_by" gorm:"type:text"`
	UpdatedTime time.Time      `json:"updated_time" gorm:"autoUpdateTime"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

// Statement represents AWS-style policy statements
type Statement struct {
	ID          string         `json:"id" gorm:"primaryKey;type:text"`
	PolicyID    string         `json:"policy_id" gorm:"not null;type:text;index"`
	SID         string         `json:"sid" gorm:"type:text"` // Statement ID (optional)
	Effect      string         `json:"effect" gorm:"not null;type:text;check:effect IN ('Allow', 'Deny')"`
	Actions     string         `json:"actions" gorm:"not null;type:text"`   // JSON array as string
	Resources   string         `json:"resources" gorm:"not null;type:text"` // JSON array as string
	Conditions  string         `json:"conditions" gorm:"type:text"`         // JSON string (optional)
	CreatedBy   string         `json:"created_by" gorm:"type:text"`
	CreatedTime time.Time      `json:"created_time" gorm:"autoCreateTime"`
	UpdatedBy   string         `json:"updated_by" gorm:"type:text"`
	UpdatedTime time.Time      `json:"updated_time" gorm:"autoUpdateTime"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

// UserRole represents the user_role join table with audit fields
type UserRole struct {
	ID          string         `json:"id" gorm:"primaryKey;type:text"`
	UserID      string         `json:"user_id" gorm:"not null;type:text;index;uniqueIndex:idx_user_role"`
	RoleID      string         `json:"role_id" gorm:"not null;type:text;index;uniqueIndex:idx_user_role"`
	CreatedBy   string         `json:"created_by" gorm:"type:text"`
	CreatedTime time.Time      `json:"created_time" gorm:"autoCreateTime"`
	UpdatedBy   string         `json:"updated_by" gorm:"type:text"`
	UpdatedTime time.Time      `json:"updated_time" gorm:"autoUpdateTime"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

// RolePolicy represents the role_policy join table
type RolePolicy struct {
	RoleID   string `json:"role_id" gorm:"primaryKey;type:text"`
	PolicyID string `json:"policy_id" gorm:"primaryKey;type:text"`
}

// UserPolicy represents the user_policy join table
type UserPolicy struct {
	UserID   string `json:"user_id" gorm:"primaryKey;type:text"`
	PolicyID string `json:"policy_id" gorm:"primaryKey;type:text"`
}

// ResourcePolicy represents resource-based policies
type ResourcePolicy struct {
	ID           string         `json:"id" gorm:"primaryKey;type:text"`
	RealmID      string         `json:"realm_id" gorm:"not null;type:text;index"`
	ResourceType string         `json:"resource_type" gorm:"not null;type:text"` // 'project', 'document', 'code', etc.
	ResourceID   string         `json:"resource_id" gorm:"not null;type:text"`   // Reference to the actual resource
	PolicyID     string         `json:"policy_id" gorm:"not null;type:text"`
	CreatedBy    string         `json:"created_by" gorm:"type:text"`
	CreatedTime  time.Time      `json:"created_time" gorm:"autoCreateTime"`
	UpdatedBy    string         `json:"updated_by" gorm:"type:text"`
	UpdatedTime  time.Time      `json:"updated_time" gorm:"autoUpdateTime"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
}
