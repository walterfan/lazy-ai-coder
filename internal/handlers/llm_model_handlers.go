package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/walterfan/lazy-ai-coder/internal/services"
	"github.com/walterfan/lazy-ai-coder/pkg/authz"
	"github.com/walterfan/lazy-ai-coder/pkg/models"
)

// LLMModelHandlers handles LLM model CRUD endpoints
type LLMModelHandlers struct {
	llmModelService *services.LLMModelService
	db              *gorm.DB
}

// NewLLMModelHandlers creates new LLM model handlers
func NewLLMModelHandlers(db *gorm.DB) *LLMModelHandlers {
	return &LLMModelHandlers{
		llmModelService: services.NewLLMModelService(db),
		db:              db,
	}
}

// ListLLMModels godoc
// @Summary List LLM models
// @Description Get list of LLM models with filtering and pagination
// @Tags llm-models
// @Produce json
// @Param scope query string false "Scope: all, personal, shared, templates" default(all)
// @Param q query string false "Search in name and description"
// @Param enabled_only query bool false "Only show enabled models" default(false)
// @Param pageNumber query int false "Page number (1-based)" default(1)
// @Param pageSize query int false "Page size" default(50)
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/llm-models [get]
func (h *LLMModelHandlers) ListLLMModels(c *gin.Context) {
	// Get user context from flexible auth middleware
	authType, userIDStr, realmIDStr, username, _ := GetUserContext(c)

	// Parse query parameters
	scopeStr := c.DefaultQuery("scope", "all")
	scope := models.LLMModelScope(scopeStr)
	nameFilter := c.Query("q")
	enabledOnlyStr := c.DefaultQuery("enabled_only", "false")
	enabledOnly := enabledOnlyStr == "true" || enabledOnlyStr == "1"

	// Parse pagination
	pageNumber, _ := strconv.Atoi(c.DefaultQuery("pageNumber", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "50"))

	if pageNumber < 1 {
		pageNumber = 1
	}
	if pageSize < 1 {
		pageSize = 50
	}

	limit := pageSize
	offset := (pageNumber - 1) * pageSize

	// Convert user context to pointers
	var userID, realmID *string
	if userIDStr != "" {
		userID = &userIDStr
	}
	if realmIDStr != "" {
		realmID = &realmIDStr
	}

	// Get LLM models
	llmModels, total, err := h.llmModelService.ListLLMModels(userID, realmID, scope, nameFilter, enabledOnly, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Calculate total pages
	totalPages := 0
	if pageSize > 0 {
		totalPages = (int(total) + pageSize - 1) / pageSize
	}

	c.JSON(http.StatusOK, gin.H{
		"data":       llmModels,
		"total":      total,
		"pageNumber": pageNumber,
		"pageSize":   pageSize,
		"totalPages": totalPages,
		"auth_type":  authType,
		"username":   username,
	})
}

// GetLLMModel godoc
// @Summary Get LLM model by ID
// @Description Get a specific LLM model by ID
// @Tags llm-models
// @Produce json
// @Param id path string true "LLM Model ID"
// @Success 200 {object} models.LLMModel
// @Router /api/v1/llm-models/{id} [get]
func (h *LLMModelHandlers) GetLLMModel(c *gin.Context) {
	id := c.Param("id")
	_, userIDStr, realmIDStr, _, _ := GetUserContext(c)

	var userID, realmID *string
	if userIDStr != "" {
		userID = &userIDStr
	}
	if realmIDStr != "" {
		realmID = &realmIDStr
	}

	llmModel, err := h.llmModelService.GetLLMModelByID(id, userID, realmID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, llmModel)
}

// GetDefaultLLMModel godoc
// @Summary Get default LLM model
// @Description Get the default LLM model for the current user/realm
// @Tags llm-models
// @Produce json
// @Success 200 {object} models.LLMModel
// @Router /api/v1/llm-models/default [get]
func (h *LLMModelHandlers) GetDefaultLLMModel(c *gin.Context) {
	_, userIDStr, realmIDStr, _, _ := GetUserContext(c)

	var userID, realmID *string
	if userIDStr != "" {
		userID = &userIDStr
	}
	if realmIDStr != "" {
		realmID = &realmIDStr
	}

	llmModel, err := h.llmModelService.GetDefaultLLMModel(userID, realmID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if llmModel == nil {
		// No default model found - return null to indicate legacy settings should be used
		c.JSON(http.StatusOK, gin.H{
			"model":       nil,
			"use_legacy":  true,
			"description": "No default model configured. Using legacy settings from localStorage.",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"model":      llmModel,
		"use_legacy": false,
	})
}

// CreateLLMModelRequest represents the request body for creating an LLM model
type CreateLLMModelRequest struct {
	Name        string  `json:"name" binding:"required"`
	LLMType     string  `json:"llm_type" binding:"required"`
	BaseURL     string  `json:"base_url" binding:"required"`
	Model       string  `json:"model" binding:"required"`
	ExtraParams string  `json:"extra_params"`
	Temperature float64 `json:"temperature"`
	MaxTokens   int     `json:"max_tokens"`
	IsDefault   bool    `json:"is_default"`
	IsEnabled   bool    `json:"is_enabled"`
	Description string  `json:"description"`
	Scope       string  `json:"scope"` // "personal" or "shared"
}

// CreateLLMModel godoc
// @Summary Create LLM model
// @Description Create a new LLM model configuration
// @Tags llm-models
// @Accept json
// @Produce json
// @Param llm_model body CreateLLMModelRequest true "LLM Model data"
// @Success 201 {object} models.LLMModel
// @Router /api/v1/llm-models [post]
func (h *LLMModelHandlers) CreateLLMModel(c *gin.Context) {
	var req CreateLLMModelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	_, userIDStr, realmIDStr, username, _ := GetUserContext(c)

	// Determine ownership based on scope
	var userID, realmID *string
	switch req.Scope {
	case "templates":
		// Global template: no user_id, no realm_id (visible to everyone)
		// Only super_admin can create templates
		if !authz.IsSuperAdmin(h.db, userIDStr) {
			c.JSON(http.StatusForbidden, gin.H{"error": "Only super admin can create global templates"})
			return
		}
		// Both userID and realmID remain nil for templates
	case "shared":
		// Shared model: no user_id, has realm_id
		if realmIDStr != "" {
			realmID = &realmIDStr
		}
	default:
		// Personal model: has user_id and realm_id
		if userIDStr != "" {
			userID = &userIDStr
		}
		if realmIDStr != "" {
			realmID = &realmIDStr
		}
	}

	// Set defaults
	if req.Temperature == 0 {
		req.Temperature = 0.7
	}
	if req.MaxTokens == 0 {
		req.MaxTokens = 4096
	}

	llmModel, err := h.llmModelService.CreateLLMModel(
		req.Name,
		req.LLMType,
		req.BaseURL,
		req.Model,
		req.ExtraParams,
		req.Temperature,
		req.MaxTokens,
		req.IsDefault,
		req.IsEnabled,
		req.Description,
		userID,
		realmID,
		username,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, llmModel)
}

// UpdateLLMModelRequest represents the request body for updating an LLM model
type UpdateLLMModelRequest struct {
	Name        string  `json:"name" binding:"required"`
	LLMType     string  `json:"llm_type" binding:"required"`
	BaseURL     string  `json:"base_url" binding:"required"`
	Model       string  `json:"model" binding:"required"`
	ExtraParams string  `json:"extra_params"`
	Temperature float64 `json:"temperature"`
	MaxTokens   int     `json:"max_tokens"`
	IsDefault   bool    `json:"is_default"`
	IsEnabled   bool    `json:"is_enabled"`
	Description string  `json:"description"`
}

// UpdateLLMModel godoc
// @Summary Update LLM model
// @Description Update an existing LLM model by ID
// @Tags llm-models
// @Accept json
// @Produce json
// @Param id path string true "LLM Model ID"
// @Param llm_model body UpdateLLMModelRequest true "LLM Model data"
// @Success 200 {object} models.LLMModel
// @Router /api/v1/llm-models/{id} [put]
func (h *LLMModelHandlers) UpdateLLMModel(c *gin.Context) {
	id := c.Param("id")

	var req UpdateLLMModelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	_, userIDStr, realmIDStr, username, _ := GetUserContext(c)

	var userID, realmID *string
	if userIDStr != "" {
		userID = &userIDStr
	}
	if realmIDStr != "" {
		realmID = &realmIDStr
	}

	llmModel, err := h.llmModelService.UpdateLLMModel(
		id,
		req.Name,
		req.LLMType,
		req.BaseURL,
		req.Model,
		req.ExtraParams,
		req.Temperature,
		req.MaxTokens,
		req.IsDefault,
		req.IsEnabled,
		req.Description,
		username,
		userID,
		realmID,
	)

	if err != nil {
		errMsg := err.Error()
		if errMsg == "unauthorized: you can only update your own LLM models" ||
			errMsg == "unauthorized: LLM model belongs to different realm" {
			c.JSON(http.StatusForbidden, gin.H{"error": errMsg})
		} else if errMsg == "LLM model not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": errMsg})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": errMsg})
		}
		return
	}

	c.JSON(http.StatusOK, llmModel)
}

// SetDefaultLLMModel godoc
// @Summary Set default LLM model
// @Description Set an LLM model as the default
// @Tags llm-models
// @Produce json
// @Param id path string true "LLM Model ID"
// @Success 200 {object} models.LLMModel
// @Router /api/v1/llm-models/{id}/default [post]
func (h *LLMModelHandlers) SetDefaultLLMModel(c *gin.Context) {
	id := c.Param("id")
	_, userIDStr, realmIDStr, username, _ := GetUserContext(c)

	var userID, realmID *string
	if userIDStr != "" {
		userID = &userIDStr
	}
	if realmIDStr != "" {
		realmID = &realmIDStr
	}

	llmModel, err := h.llmModelService.SetDefaultLLMModel(id, userID, realmID, username)
	if err != nil {
		errMsg := err.Error()
		if errMsg == "unauthorized: you can only set default for your own LLM models" {
			c.JSON(http.StatusForbidden, gin.H{"error": errMsg})
		} else if errMsg == "LLM model not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": errMsg})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": errMsg})
		}
		return
	}

	c.JSON(http.StatusOK, llmModel)
}

// ToggleLLMModelRequest represents the request body for toggling model enabled status
type ToggleLLMModelRequest struct {
	Enabled bool `json:"enabled"`
}

// ToggleLLMModelEnabled godoc
// @Summary Toggle LLM model enabled status
// @Description Enable or disable an LLM model
// @Tags llm-models
// @Accept json
// @Produce json
// @Param id path string true "LLM Model ID"
// @Param body body ToggleLLMModelRequest true "Toggle data"
// @Success 200 {object} models.LLMModel
// @Router /api/v1/llm-models/{id}/toggle [post]
func (h *LLMModelHandlers) ToggleLLMModelEnabled(c *gin.Context) {
	id := c.Param("id")

	var req ToggleLLMModelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	_, userIDStr, realmIDStr, username, _ := GetUserContext(c)

	var userID, realmID *string
	if userIDStr != "" {
		userID = &userIDStr
	}
	if realmIDStr != "" {
		realmID = &realmIDStr
	}

	llmModel, err := h.llmModelService.ToggleLLMModelEnabled(id, req.Enabled, userID, realmID, username)
	if err != nil {
		errMsg := err.Error()
		if errMsg == "unauthorized: you can only toggle your own LLM models" {
			c.JSON(http.StatusForbidden, gin.H{"error": errMsg})
		} else if errMsg == "LLM model not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": errMsg})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": errMsg})
		}
		return
	}

	c.JSON(http.StatusOK, llmModel)
}

// DeleteLLMModel godoc
// @Summary Delete LLM model
// @Description Delete an LLM model (soft delete) by ID
// @Tags llm-models
// @Produce json
// @Param id path string true "LLM Model ID"
// @Success 204
// @Router /api/v1/llm-models/{id} [delete]
func (h *LLMModelHandlers) DeleteLLMModel(c *gin.Context) {
	id := c.Param("id")
	_, userIDStr, realmIDStr, _, _ := GetUserContext(c)

	var userID, realmID *string
	if userIDStr != "" {
		userID = &userIDStr
	}
	if realmIDStr != "" {
		realmID = &realmIDStr
	}

	err := h.llmModelService.DeleteLLMModel(id, userID, realmID)
	if err != nil {
		errMsg := err.Error()
		if errMsg == "unauthorized: you can only delete your own LLM models" ||
			errMsg == "unauthorized: LLM model belongs to different realm" {
			c.JSON(http.StatusForbidden, gin.H{"error": errMsg})
		} else if errMsg == "LLM model not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": errMsg})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": errMsg})
		}
		return
	}

	c.Status(http.StatusNoContent)
}
