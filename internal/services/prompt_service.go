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

// PromptService handles prompt CRUD operations with user isolation
type PromptService struct {
	db *gorm.DB
}

// NewPromptService creates a new prompt service
func NewPromptService(db *gorm.DB) *PromptService {
	return &PromptService{db: db}
}

// PromptScope represents the scope for querying prompts
type PromptScope string

const (
	ScopeAll       PromptScope = "all"       // Personal + Shared + Templates
	ScopePersonal  PromptScope = "personal"  // User's personal prompts
	ScopeShared    PromptScope = "shared"    // Realm shared prompts
	ScopeTemplates PromptScope = "templates" // Global templates
)

// ListPrompts retrieves prompts based on scope and filters
func (s *PromptService) ListPrompts(userID, realmID *string, scope PromptScope, nameFilter, tagsFilter string, sortBy string, limit, offset int) ([]models.Prompt, int64, error) {
	query := s.db.Model(&models.Prompt{}).Where("deleted_at IS NULL")

	// Apply scope filtering
	switch scope {
	case ScopePersonal:
		if userID == nil || *userID == "" {
			return nil, 0, errors.New("user_id required for personal scope")
		}
		query = query.Where("user_id = ?", *userID)

	case ScopeShared:
		if realmID == nil || *realmID == "" {
			return nil, 0, errors.New("realm_id required for shared scope")
		}
		query = query.Where("realm_id = ? AND user_id IS NULL", *realmID)

	case ScopeTemplates:
		query = query.Where("user_id IS NULL AND realm_id IS NULL")

	case ScopeAll:
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

	// Apply tags filter (supports comma-separated tags with OR logic)
	if tagsFilter != "" {
		tags := strings.Split(tagsFilter, ",")
		if len(tags) == 1 {
			// Single tag
			query = query.Where("LOWER(tags) LIKE ?", "%"+strings.ToLower(strings.TrimSpace(tags[0]))+"%")
		} else {
			// Multiple tags - use OR logic
			var conditions []string
			var args []interface{}
			for _, tag := range tags {
				trimmedTag := strings.TrimSpace(tag)
				if trimmedTag != "" {
					conditions = append(conditions, "LOWER(tags) LIKE ?")
					args = append(args, "%"+strings.ToLower(trimmedTag)+"%")
				}
			}
			if len(conditions) > 0 {
				query = query.Where(strings.Join(conditions, " OR "), args...)
			}
		}
	}

	// Get total count
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count prompts: %w", err)
	}

	// Apply sorting
	switch sortBy {
	case "name":
		query = query.Order("name ASC")
	case "updated_at":
		query = query.Order("updated_time DESC")
	default:
		query = query.Order("created_time DESC")
	}

	// Apply pagination
	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	var prompts []models.Prompt
	if err := query.Find(&prompts).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list prompts: %w", err)
	}

	return prompts, total, nil
}

// GetPromptByID retrieves a prompt by ID or Name
func (s *PromptService) GetPromptByID(id string, userID, realmID *string) (*models.Prompt, error) {
	var prompt models.Prompt
	// Support lookup by both ID (UUID) and Name for backward compatibility
	query := s.db.Where("(id = ? OR name = ?) AND deleted_at IS NULL", id, id)

	// Check if user is super_admin
	userIDStr := ""
	if userID != nil {
		userIDStr = *userID
	}
	isSuperAdmin := authz.IsSuperAdmin(s.db, userIDStr)

	// Access control: super_admin can access any prompt, others follow realm rules
	if !isSuperAdmin {
		if userID != nil && *userID != "" && realmID != nil && *realmID != "" {
			query = query.Where(
				"(user_id = ?) OR (realm_id = ? AND user_id IS NULL) OR (user_id IS NULL AND realm_id IS NULL)",
				*userID, *realmID,
			)
		}
	}

	if err := query.First(&prompt).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("prompt not found")
		}
		return nil, fmt.Errorf("failed to get prompt: %w", err)
	}

	return &prompt, nil
}

// CreatePrompt creates a new prompt
func (s *PromptService) CreatePrompt(name, title, description, systemPrompt, userPrompt, arguments, tags string, userID, realmID *string, createdBy string) (*models.Prompt, error) {
	prompt := &models.Prompt{
		ID:           uuid.New().String(),
		UserID:       userID,
		RealmID:      realmID,
		Name:         name,
		Title:        title,
		Description:  description,
		SystemPrompt: systemPrompt,
		UserPrompt:   userPrompt,
		Arguments:    arguments,
		Tags:         tags,
		CreatedBy:    createdBy,
		UpdatedBy:    createdBy,
	}

	if err := s.db.Create(prompt).Error; err != nil {
		return nil, fmt.Errorf("failed to create prompt: %w", err)
	}

	return prompt, nil
}

// UpdatePrompt updates an existing prompt
func (s *PromptService) UpdatePrompt(id, name, title, description, systemPrompt, userPrompt, arguments, tags string, updatedBy string, userID, realmID *string) (*models.Prompt, error) {
	// Get existing prompt (supports lookup by ID or name)
	prompt, err := s.GetPromptByID(id, userID, realmID)
	if err != nil {
		return nil, err
	}

	// Authorization check
	userIDStr := ""
	if userID != nil {
		userIDStr = *userID
	}

	// Super admins can edit any prompt
	isSuperAdmin := authz.IsSuperAdmin(s.db, userIDStr)

	if !isSuperAdmin {
		// Non-super-admin authorization: user can only update their own prompts
		if userID != nil && *userID != "" {
			if prompt.UserID == nil || *prompt.UserID != *userID {
				return nil, errors.New("unauthorized: you can only update your own prompts")
			}
		}

		// Additional realm isolation check for non-super-admins
		if realmID != nil && *realmID != "" {
			if prompt.RealmID == nil || *prompt.RealmID != *realmID {
				return nil, errors.New("unauthorized: prompt belongs to different realm")
			}
		}
	}

	// Update fields - use the actual UUID ID from the fetched prompt
	updates := map[string]interface{}{
		"name":          name,
		"title":         title,
		"description":   description,
		"system_prompt": systemPrompt,
		"user_prompt":   userPrompt,
		"arguments":     arguments,
		"tags":          tags,
		"updated_by":    updatedBy,
	}

	if err := s.db.Model(&models.Prompt{}).Where("id = ?", prompt.ID).Updates(updates).Error; err != nil {
		return nil, fmt.Errorf("failed to update prompt: %w", err)
	}

	// Fetch updated prompt using the actual ID
	return s.GetPromptByID(prompt.ID, userID, realmID)
}

// DeletePrompt soft-deletes a prompt
func (s *PromptService) DeletePrompt(id string, userID, realmID *string) error {
	// Get existing prompt (supports lookup by ID or name)
	prompt, err := s.GetPromptByID(id, userID, realmID)
	if err != nil {
		return err
	}

	// Authorization check
	userIDStr := ""
	if userID != nil {
		userIDStr = *userID
	}

	// Super admins can delete any prompt
	isSuperAdmin := authz.IsSuperAdmin(s.db, userIDStr)

	if !isSuperAdmin {
		// Non-super-admin authorization: user can only delete their own prompts
		if userID != nil && *userID != "" {
			if prompt.UserID == nil || *prompt.UserID != *userID {
				return errors.New("unauthorized: you can only delete your own prompts")
			}
		}

		// Additional realm isolation check for non-super-admins
		if realmID != nil && *realmID != "" {
			if prompt.RealmID == nil || *prompt.RealmID != *realmID {
				return errors.New("unauthorized: prompt belongs to different realm")
			}
		}
	}

	// Soft delete - use the actual UUID ID from the fetched prompt
	if err := s.db.Model(&models.Prompt{}).Where("id = ?", prompt.ID).Update("deleted_at", gorm.Expr("CURRENT_TIMESTAMP")).Error; err != nil {
		return fmt.Errorf("failed to delete prompt: %w", err)
	}

	return nil
}
