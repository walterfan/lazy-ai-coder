package models

import (
	"time"

	"gorm.io/gorm"
)

// CursorCommand represents the cursor_commands table in the database
type CursorCommand struct {
	ID          string         `json:"id" gorm:"primaryKey;type:text"`
	UserID      *string        `json:"user_id" gorm:"index;type:text"` // NULL for global templates
	RealmID     *string        `json:"realm_id" gorm:"index;type:text"` // NULL for global templates
	Name        string         `json:"name" gorm:"index;not null;type:text"`
	Description string         `json:"description" gorm:"type:text"`
	Command     string         `json:"command" gorm:"type:text"` // The actual command/prompt text
	Category    string         `json:"category" gorm:"type:text"` // e.g., "refactor", "debug", "generate", "review"
	Language    string         `json:"language" gorm:"type:text"` // e.g., "go", "typescript", "general"
	Framework   string         `json:"framework" gorm:"type:text"` // e.g., "gin", "vue", "general"
	Tags        string         `json:"tags" gorm:"type:text"` // Comma-separated
	IsTemplate  bool           `json:"is_template" gorm:"default:false"` // Template for generation
	UsageCount  int            `json:"usage_count" gorm:"default:0"` // Track popularity
	CreatedBy   string         `json:"created_by" gorm:"type:text"`
	CreatedTime time.Time      `json:"created_time" gorm:"autoCreateTime"`
	UpdatedBy   string         `json:"updated_by" gorm:"type:text"`
	UpdatedTime time.Time      `json:"updated_time" gorm:"autoUpdateTime"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

