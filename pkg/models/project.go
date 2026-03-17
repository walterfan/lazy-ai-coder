package models

import (
	"time"

	"gorm.io/gorm"
)

// Project represents the project table in the database
type Project struct {
	ID          string         `json:"id" gorm:"primaryKey;type:text"`
	UserID      *string        `json:"user_id" gorm:"index;type:text"` // NULL for shared projects
	RealmID     string         `json:"realm_id" gorm:"not null;type:text;index"`
	Name        string         `json:"name" gorm:"not null;type:text"`
	Description string         `json:"description" gorm:"type:text"`
	GitURL      string         `json:"git_url" gorm:"column:git_url;type:text"`
	GitRepo     string         `json:"git_repo" gorm:"column:git_repo;type:text"`
	GitBranch   string         `json:"git_branch" gorm:"column:git_branch;type:text"`
	Language    string         `json:"language" gorm:"type:text"`
	EntryPoint  string         `json:"entry_point" gorm:"column:entry_point;type:text"`
	CreatedBy   string         `json:"created_by" gorm:"type:text"`
	CreatedTime time.Time      `json:"created_time" gorm:"autoCreateTime"`
	UpdatedBy   string         `json:"updated_by" gorm:"type:text"`
	UpdatedTime time.Time      `json:"updated_time" gorm:"autoUpdateTime"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

// Code represents the code table in the database
type Code struct {
	ID              string         `json:"id" gorm:"primaryKey;type:text"`
	RealmID         string         `json:"realm_id" gorm:"not null;type:text;index"`
	ProjectID       string         `json:"project_id" gorm:"not null;type:text;index"`
	Path            string         `json:"path" gorm:"not null;type:text"`
	Code            string         `json:"code" gorm:"not null;type:text"`
	VectorEmbedding string         `json:"vector_embedding" gorm:"column:vector_embedding;type:text"` // Store vector as JSON string
	CreatedBy       string         `json:"created_by" gorm:"type:text"`
	CreatedTime     time.Time      `json:"created_time" gorm:"autoCreateTime"`
	UpdatedBy       string         `json:"updated_by" gorm:"type:text"`
	UpdatedTime     time.Time      `json:"updated_time" gorm:"autoUpdateTime"`
	DeletedAt       gorm.DeletedAt `gorm:"index" json:"-"`
}

// Document represents the document table in the database
type Document struct {
	ID              string         `json:"id" gorm:"primaryKey;type:text"`
	RealmID         string         `json:"realm_id" gorm:"not null;type:text;index"`
	ProjectID       string         `json:"project_id" gorm:"not null;type:text;index"`
	Name            string         `json:"name" gorm:"not null;type:text"`
	Path            string         `json:"path" gorm:"not null;type:text"`
	Content         string         `json:"content" gorm:"not null;type:text"`
	VectorEmbedding string         `json:"vector_embedding" gorm:"column:vector_embedding;type:text"` // Store vector as JSON string
	CreatedBy       string         `json:"created_by" gorm:"type:text"`
	CreatedTime     time.Time      `json:"created_time" gorm:"autoCreateTime"`
	UpdatedBy       string         `json:"updated_by" gorm:"type:text"`
	UpdatedTime     time.Time      `json:"updated_time" gorm:"autoUpdateTime"`
	DeletedAt       gorm.DeletedAt `gorm:"index" json:"-"`
}
