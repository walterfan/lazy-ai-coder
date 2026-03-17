package models

import (
	"encoding/json"
	"time"

	"gorm.io/gorm"
)

// Prompt represents the prompt table in the database
type Prompt struct {
	ID           string         `json:"id" gorm:"primaryKey;type:text"`
	UserID       *string        `json:"user_id" gorm:"index;type:text"` // NULL for global templates
	RealmID      *string        `json:"realm_id" gorm:"index;type:text"` // NULL for global templates
	Name         string         `json:"name" gorm:"index;not null;type:text"`
	Title        string         `json:"title" gorm:"type:text"`          // Human-readable display name
	Description  string         `json:"description" gorm:"type:text"`
	SystemPrompt string         `json:"system_prompt" gorm:"column:system_prompt;type:text"`
	UserPrompt   string         `json:"user_prompt" gorm:"column:user_prompt;type:text"`
	Arguments    string         `json:"-" gorm:"type:text"`              // JSON array of PromptArgument (stored as string in DB)
	Tags         string         `json:"tags" gorm:"type:text"`
	CreatedBy    string         `json:"created_by" gorm:"type:text"`
	CreatedTime  time.Time      `json:"created_time" gorm:"autoCreateTime"`
	UpdatedBy    string         `json:"updated_by" gorm:"type:text"`
	UpdatedTime  time.Time      `json:"updated_time" gorm:"autoUpdateTime"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
}

// PromptArgument represents an argument definition for a prompt
type PromptArgument struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Required    bool   `json:"required"`
}

// MarshalJSON implements custom JSON marshaling for Prompt
// This converts the Arguments string field to an array for API responses
func (p Prompt) MarshalJSON() ([]byte, error) {
	type Alias Prompt
	args := []PromptArgument{} // Initialize as empty array, not nil

	// Only try to unmarshal if Arguments is not empty
	if p.Arguments != "" {
		if err := json.Unmarshal([]byte(p.Arguments), &args); err != nil {
			// If parsing fails, keep empty array
			args = []PromptArgument{}
		}
	}

	return json.Marshal(&struct {
		*Alias
		Arguments []PromptArgument `json:"arguments"`
	}{
		Alias:     (*Alias)(&p),
		Arguments: args,
	})
}
