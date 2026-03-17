package services

import (
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/walterfan/lazy-ai-coder/pkg/authz"
	"github.com/walterfan/lazy-ai-coder/pkg/models"
)

// LLMModelService handles LLM model CRUD operations with user isolation
type LLMModelService struct {
	db *gorm.DB
}

// NewLLMModelService creates a new LLM model service
func NewLLMModelService(db *gorm.DB) *LLMModelService {
	return &LLMModelService{db: db}
}

// ListLLMModels retrieves LLM models based on scope and filters
func (s *LLMModelService) ListLLMModels(userID, realmID *string, scope models.LLMModelScope, nameFilter string, enabledOnly bool, limit, offset int) ([]models.LLMModel, int64, error) {
	query := s.db.Model(&models.LLMModel{}).Where("deleted_at IS NULL")

	// Apply scope filtering
	switch scope {
	case models.LLMScopePersonal:
		if userID == nil || *userID == "" {
			return nil, 0, errors.New("user_id required for personal scope")
		}
		query = query.Where("user_id = ?", *userID)

	case models.LLMScopeShared:
		if realmID == nil || *realmID == "" {
			return nil, 0, errors.New("realm_id required for shared scope")
		}
		query = query.Where("realm_id = ? AND user_id IS NULL", *realmID)

	case models.LLMScopeTemplates:
		query = query.Where("user_id IS NULL AND realm_id IS NULL")

	case models.LLMScopeAll:
		// Return personal + shared + templates
		if userID != nil && *userID != "" && realmID != nil && *realmID != "" {
			query = query.Where(
				"(user_id = ?) OR (realm_id = ? AND user_id IS NULL) OR (user_id IS NULL AND realm_id IS NULL)",
				*userID, *realmID,
			)
		} else if realmID != nil && *realmID != "" {
			// No user, show shared + templates
			query = query.Where(
				"(realm_id = ? AND user_id IS NULL) OR (user_id IS NULL AND realm_id IS NULL)",
				*realmID,
			)
		} else {
			// Only templates
			query = query.Where("user_id IS NULL AND realm_id IS NULL")
		}
	}

	// Apply name filter
	if nameFilter != "" {
		query = query.Where("LOWER(name) LIKE ? OR LOWER(description) LIKE ?",
			"%"+strings.ToLower(nameFilter)+"%",
			"%"+strings.ToLower(nameFilter)+"%")
	}

	// Apply enabled filter
	if enabledOnly {
		query = query.Where("is_enabled = ?", true)
	}

	// Get total count
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count LLM models: %w", err)
	}

	// Apply sorting: default first, then by name
	query = query.Order("is_default DESC, name ASC")

	// Apply pagination
	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	var llmModels []models.LLMModel
	if err := query.Find(&llmModels).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list LLM models: %w", err)
	}

	return llmModels, total, nil
}

// GetLLMModelByID retrieves an LLM model by ID
func (s *LLMModelService) GetLLMModelByID(id string, userID, realmID *string) (*models.LLMModel, error) {
	var llmModel models.LLMModel
	query := s.db.Where("id = ? AND deleted_at IS NULL", id)

	// Check if user is super_admin
	userIDStr := ""
	if userID != nil {
		userIDStr = *userID
	}
	isSuperAdmin := authz.IsSuperAdmin(s.db, userIDStr)

	// Access control: super_admin can access any model, others follow realm rules
	if !isSuperAdmin {
		if userID != nil && *userID != "" && realmID != nil && *realmID != "" {
			query = query.Where(
				"(user_id = ?) OR (realm_id = ? AND user_id IS NULL) OR (user_id IS NULL AND realm_id IS NULL)",
				*userID, *realmID,
			)
		}
	}

	if err := query.First(&llmModel).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("LLM model not found")
		}
		return nil, fmt.Errorf("failed to get LLM model: %w", err)
	}

	return &llmModel, nil
}

// GetDefaultLLMModel retrieves the default LLM model for a user/realm
func (s *LLMModelService) GetDefaultLLMModel(userID, realmID *string) (*models.LLMModel, error) {
	var llmModel models.LLMModel
	query := s.db.Where("is_default = ? AND is_enabled = ? AND deleted_at IS NULL", true, true)

	// Check user's personal default first, then realm default, then global template default
	if userID != nil && *userID != "" {
		// Try user's personal default
		err := query.Where("user_id = ?", *userID).First(&llmModel).Error
		if err == nil {
			return &llmModel, nil
		}
	}

	if realmID != nil && *realmID != "" {
		// Try realm shared default
		err := s.db.Where("is_default = ? AND is_enabled = ? AND deleted_at IS NULL AND realm_id = ? AND user_id IS NULL",
			true, true, *realmID).First(&llmModel).Error
		if err == nil {
			return &llmModel, nil
		}
	}

	// Try global template default
	err := s.db.Where("is_default = ? AND is_enabled = ? AND deleted_at IS NULL AND user_id IS NULL AND realm_id IS NULL",
		true, true).First(&llmModel).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // No default model found - will use legacy settings
		}
		return nil, fmt.Errorf("failed to get default LLM model: %w", err)
	}

	return &llmModel, nil
}

// CreateLLMModel creates a new LLM model
func (s *LLMModelService) CreateLLMModel(name, llmType, baseURL, model, extraParams string, temperature float64, maxTokens int, isDefault, isEnabled bool, description string, userID, realmID *string, createdBy string) (*models.LLMModel, error) {
	llmModel := &models.LLMModel{
		ID:          uuid.New().String(),
		Name:        name,
		LLMType:     llmType,
		BaseURL:     baseURL,
		Model:       model,
		ExtraParams: extraParams,
		Temperature: temperature,
		MaxTokens:   maxTokens,
		IsDefault:   isDefault,
		IsEnabled:   isEnabled,
		Description: description,
		UserID:      userID,
		RealmID:     realmID,
		CreatedBy:   createdBy,
		UpdatedBy:   createdBy,
	}

	// If setting as default, clear other defaults in the same scope
	if isDefault {
		if err := s.clearDefaultInScope(userID, realmID); err != nil {
			return nil, err
		}
	}

	if err := s.db.Create(llmModel).Error; err != nil {
		return nil, fmt.Errorf("failed to create LLM model: %w", err)
	}

	return llmModel, nil
}

// UpdateLLMModel updates an existing LLM model
func (s *LLMModelService) UpdateLLMModel(id, name, llmType, baseURL, model, extraParams string, temperature float64, maxTokens int, isDefault, isEnabled bool, description string, updatedBy string, userID, realmID *string) (*models.LLMModel, error) {
	// Get existing model
	llmModel, err := s.GetLLMModelByID(id, userID, realmID)
	if err != nil {
		return nil, err
	}

	// Authorization check
	userIDStr := ""
	if userID != nil {
		userIDStr = *userID
	}

	// Super admins can edit any model
	isSuperAdmin := authz.IsSuperAdmin(s.db, userIDStr)

	if !isSuperAdmin {
		// Non-super-admin authorization: user can only update their own models
		if userID != nil && *userID != "" {
			if llmModel.UserID == nil || *llmModel.UserID != *userID {
				return nil, errors.New("unauthorized: you can only update your own LLM models")
			}
		}

		// Additional realm isolation check for non-super-admins
		if realmID != nil && *realmID != "" {
			if llmModel.RealmID == nil || *llmModel.RealmID != *realmID {
				return nil, errors.New("unauthorized: LLM model belongs to different realm")
			}
		}
	}

	// If setting as default, clear other defaults in the same scope
	if isDefault && !llmModel.IsDefault {
		if err := s.clearDefaultInScope(llmModel.UserID, llmModel.RealmID); err != nil {
			return nil, err
		}
	}

	// Update fields
	updates := map[string]interface{}{
		"name":         name,
		"llm_type":     llmType,
		"base_url":     baseURL,
		"model":        model,
		"extra_params": extraParams,
		"temperature":  temperature,
		"max_tokens":   maxTokens,
		"is_default":   isDefault,
		"is_enabled":   isEnabled,
		"description":  description,
		"updated_by":   updatedBy,
	}

	if err := s.db.Model(&models.LLMModel{}).Where("id = ?", llmModel.ID).Updates(updates).Error; err != nil {
		return nil, fmt.Errorf("failed to update LLM model: %w", err)
	}

	// Fetch updated model
	return s.GetLLMModelByID(llmModel.ID, userID, realmID)
}

// SetDefaultLLMModel sets a model as the default
func (s *LLMModelService) SetDefaultLLMModel(id string, userID, realmID *string, updatedBy string) (*models.LLMModel, error) {
	// Get existing model
	llmModel, err := s.GetLLMModelByID(id, userID, realmID)
	if err != nil {
		return nil, err
	}

	// Authorization check
	userIDStr := ""
	if userID != nil {
		userIDStr = *userID
	}

	// Super admins can set default for any model
	isSuperAdmin := authz.IsSuperAdmin(s.db, userIDStr)

	if !isSuperAdmin {
		// Non-super-admin authorization
		if userID != nil && *userID != "" {
			if llmModel.UserID == nil || *llmModel.UserID != *userID {
				return nil, errors.New("unauthorized: you can only set default for your own LLM models")
			}
		}
	}

	// Clear other defaults in the same scope
	if err := s.clearDefaultInScope(llmModel.UserID, llmModel.RealmID); err != nil {
		return nil, err
	}

	// Set this model as default
	if err := s.db.Model(&models.LLMModel{}).Where("id = ?", llmModel.ID).Updates(map[string]interface{}{
		"is_default": true,
		"updated_by": updatedBy,
	}).Error; err != nil {
		return nil, fmt.Errorf("failed to set default LLM model: %w", err)
	}

	// Fetch updated model
	return s.GetLLMModelByID(llmModel.ID, userID, realmID)
}

// ToggleLLMModelEnabled toggles the enabled status of a model
func (s *LLMModelService) ToggleLLMModelEnabled(id string, enabled bool, userID, realmID *string, updatedBy string) (*models.LLMModel, error) {
	// Get existing model
	llmModel, err := s.GetLLMModelByID(id, userID, realmID)
	if err != nil {
		return nil, err
	}

	// Authorization check
	userIDStr := ""
	if userID != nil {
		userIDStr = *userID
	}

	// Super admins can toggle any model
	isSuperAdmin := authz.IsSuperAdmin(s.db, userIDStr)

	if !isSuperAdmin {
		if userID != nil && *userID != "" {
			if llmModel.UserID == nil || *llmModel.UserID != *userID {
				return nil, errors.New("unauthorized: you can only toggle your own LLM models")
			}
		}
	}

	// Update enabled status
	if err := s.db.Model(&models.LLMModel{}).Where("id = ?", llmModel.ID).Updates(map[string]interface{}{
		"is_enabled": enabled,
		"updated_by": updatedBy,
	}).Error; err != nil {
		return nil, fmt.Errorf("failed to toggle LLM model: %w", err)
	}

	// Fetch updated model
	return s.GetLLMModelByID(llmModel.ID, userID, realmID)
}

// DeleteLLMModel soft-deletes an LLM model
func (s *LLMModelService) DeleteLLMModel(id string, userID, realmID *string) error {
	// Get existing model
	llmModel, err := s.GetLLMModelByID(id, userID, realmID)
	if err != nil {
		return err
	}

	// Authorization check
	userIDStr := ""
	if userID != nil {
		userIDStr = *userID
	}

	// Super admins can delete any model
	isSuperAdmin := authz.IsSuperAdmin(s.db, userIDStr)

	if !isSuperAdmin {
		// Non-super-admin authorization: user can only delete their own models
		if userID != nil && *userID != "" {
			if llmModel.UserID == nil || *llmModel.UserID != *userID {
				return errors.New("unauthorized: you can only delete your own LLM models")
			}
		}

		// Additional realm isolation check for non-super-admins
		if realmID != nil && *realmID != "" {
			if llmModel.RealmID == nil || *llmModel.RealmID != *realmID {
				return errors.New("unauthorized: LLM model belongs to different realm")
			}
		}
	}

	// Soft delete
	if err := s.db.Model(&models.LLMModel{}).Where("id = ?", llmModel.ID).Update("deleted_at", gorm.Expr("CURRENT_TIMESTAMP")).Error; err != nil {
		return fmt.Errorf("failed to delete LLM model: %w", err)
	}

	return nil
}

// clearDefaultInScope clears the default flag for all models in the same scope
func (s *LLMModelService) clearDefaultInScope(userID, realmID *string) error {
	query := s.db.Model(&models.LLMModel{}).Where("is_default = ? AND deleted_at IS NULL", true)

	if userID != nil && *userID != "" {
		// Clear user's personal defaults
		query = query.Where("user_id = ?", *userID)
	} else if realmID != nil && *realmID != "" {
		// Clear realm shared defaults
		query = query.Where("realm_id = ? AND user_id IS NULL", *realmID)
	} else {
		// Clear global template defaults
		query = query.Where("user_id IS NULL AND realm_id IS NULL")
	}

	if err := query.Update("is_default", false).Error; err != nil {
		return fmt.Errorf("failed to clear default LLM models: %w", err)
	}

	return nil
}

// InitDefaultLLMModels initializes the database with default LLM model configurations
func (s *LLMModelService) InitDefaultLLMModels() error {
	var count int64
	s.db.Model(&models.LLMModel{}).Count(&count)
	if count > 0 {
		return nil // Already initialized
	}

	defaultModels := []models.LLMModel{
		{ID: "llm_openai_gpt4", Name: "GPT-4", LLMType: "openai", BaseURL: "https://api.openai.com/v1", Model: "gpt-4", Temperature: 0.7, MaxTokens: 8192, IsEnabled: true, Description: "OpenAI GPT-4 - Most capable model for complex tasks", CreatedBy: "system"},
		{ID: "llm_openai_gpt4_turbo", Name: "GPT-4 Turbo", LLMType: "openai", BaseURL: "https://api.openai.com/v1", Model: "gpt-4-turbo-preview", Temperature: 0.7, MaxTokens: 128000, IsEnabled: true, Description: "OpenAI GPT-4 Turbo - Faster with larger context window", CreatedBy: "system"},
		{ID: "llm_openai_gpt35", Name: "GPT-3.5 Turbo", LLMType: "openai", BaseURL: "https://api.openai.com/v1", Model: "gpt-3.5-turbo", Temperature: 0.7, MaxTokens: 16384, IsEnabled: true, Description: "OpenAI GPT-3.5 Turbo - Fast and cost-effective", CreatedBy: "system"},
		{ID: "llm_anthropic_claude3_opus", Name: "Claude 3 Opus", LLMType: "anthropic", BaseURL: "https://api.anthropic.com/v1", Model: "claude-3-opus-20240229", Temperature: 0.7, MaxTokens: 4096, IsEnabled: true, Description: "Anthropic Claude 3 Opus - Most intelligent model", CreatedBy: "system"},
		{ID: "llm_anthropic_claude3_sonnet", Name: "Claude 3.5 Sonnet", LLMType: "anthropic", BaseURL: "https://api.anthropic.com/v1", Model: "claude-3-5-sonnet-20241022", Temperature: 0.7, MaxTokens: 8192, IsEnabled: true, Description: "Anthropic Claude 3.5 Sonnet - Balanced performance", CreatedBy: "system"},
		{ID: "llm_google_gemini_pro", Name: "Gemini Pro", LLMType: "google", BaseURL: "https://generativelanguage.googleapis.com/v1beta", Model: "gemini-pro", Temperature: 0.7, MaxTokens: 32768, IsEnabled: true, Description: "Google Gemini Pro - Multimodal capabilities", CreatedBy: "system"},
		{ID: "llm_alibaba_qwen_max", Name: "Qwen Max", LLMType: "alibaba", BaseURL: "https://dashscope.aliyuncs.com/compatible-mode/v1", Model: "qwen-max", Temperature: 0.7, MaxTokens: 8192, IsEnabled: true, Description: "Alibaba Qwen Max - Excellent for Chinese and English", CreatedBy: "system"},
		{ID: "llm_deepseek_chat", Name: "DeepSeek Chat", LLMType: "deepseek", BaseURL: "https://api.deepseek.com/v1", Model: "deepseek-chat", Temperature: 0.7, MaxTokens: 32768, IsEnabled: true, Description: "DeepSeek Chat - Strong coding and reasoning", CreatedBy: "system"},
	}

	for _, m := range defaultModels {
		if err := s.db.Create(&m).Error; err != nil {
			return fmt.Errorf("failed to create default LLM model %s: %w", m.Name, err)
		}
	}

	return nil
}

