package models

import (
	"time"

	"gorm.io/gorm"
)

// LLMModel represents an LLM model configuration in the database
type LLMModel struct {
	ID          string         `json:"id" gorm:"primaryKey;type:text"`
	Name        string         `json:"name" gorm:"not null;type:text"`
	LLMType     string         `json:"llm_type" gorm:"column:llm_type;not null;type:text;default:openai"` // openai, anthropic, google, alibaba, deepseek
	BaseURL     string         `json:"base_url" gorm:"column:base_url;not null;type:text"`
	Model       string         `json:"model" gorm:"not null;type:text"`                 // Model identifier (e.g., gpt-4, claude-3-opus)
	ExtraParams string         `json:"extra_params" gorm:"column:extra_params;type:text"` // JSON for additional provider-specific parameters
	Temperature float64        `json:"temperature" gorm:"default:0.7"`
	MaxTokens   int            `json:"max_tokens" gorm:"column:max_tokens;default:4096"`
	IsDefault   bool           `json:"is_default" gorm:"column:is_default;default:false"`
	IsEnabled   bool           `json:"is_enabled" gorm:"column:is_enabled;default:true"`
	Description string         `json:"description" gorm:"type:text"`
	UserID      *string        `json:"user_id" gorm:"index;type:text"`   // NULL for realm-shared models
	RealmID     *string        `json:"realm_id" gorm:"index;type:text"`  // NULL for global/template models
	CreatedBy   string         `json:"created_by" gorm:"type:text"`
	CreatedTime time.Time      `json:"created_time" gorm:"autoCreateTime"`
	UpdatedBy   string         `json:"updated_by" gorm:"type:text"`
	UpdatedTime time.Time      `json:"updated_time" gorm:"autoUpdateTime"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName specifies the table name for LLMModel
func (LLMModel) TableName() string {
	return "llm_models"
}

// LLMModelScope represents the scope for querying LLM models
type LLMModelScope string

const (
	LLMScopeAll       LLMModelScope = "all"       // Personal + Shared + Templates
	LLMScopePersonal  LLMModelScope = "personal"  // User's personal models
	LLMScopeShared    LLMModelScope = "shared"    // Realm shared models
	LLMScopeTemplates LLMModelScope = "templates" // Global templates (no user/realm)
)

