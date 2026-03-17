package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/walterfan/lazy-ai-coder/internal/services"
	"github.com/walterfan/lazy-ai-coder/pkg/models"
)

// PromptHandlers handles prompt CRUD endpoints
type PromptHandlers struct {
	promptService *services.PromptService
	db            *gorm.DB
}

// NewPromptHandlers creates new prompt handlers
func NewPromptHandlers(db *gorm.DB) *PromptHandlers {
	return &PromptHandlers{
		promptService: services.NewPromptService(db),
		db:            db,
	}
}

// ListPrompts godoc
// @Summary List prompts
// @Description Get list of prompts with filtering and pagination
// @Tags prompts
// @Produce json
// @Param scope query string false "Scope: all, personal, shared, templates" default(all)
// @Param q query string false "Search in name and description"
// @Param tags query string false "Filter by tags (comma-separated for OR logic)"
// @Param sort query string false "Sort by: created_at, updated_at, name" default(created_at)
// @Param pageNumber query int false "Page number (1-based, overrides offset)" default(1)
// @Param pageSize query int false "Page size (overrides limit)" default(200)
// @Param limit query int false "Limit results (deprecated, use pageSize)" default(200)
// @Param offset query int false "Offset for pagination (deprecated, use pageNumber)" default(0)
// @Success 200 {array} models.Prompt
// @Router /api/v1/prompts [get]
func (h *PromptHandlers) ListPrompts(c *gin.Context) {
	// Get user context from flexible auth middleware
	authType, userIDStr, realmIDStr, username, _ := GetUserContext(c)

	// Parse query parameters
	scopeStr := c.DefaultQuery("scope", "all")
	scope := services.PromptScope(scopeStr)
	nameFilter := c.Query("q")
	tagsFilter := c.Query("tags")
	sortBy := c.DefaultQuery("sort", "created_at")

	// Support both old (offset/limit) and new (pageNumber/pageSize) pagination
	var limit, offset int
	pageNumberStr := c.Query("pageNumber")
	pageSizeStr := c.Query("pageSize")

	if pageNumberStr != "" || pageSizeStr != "" {
		// New pagination style with pageNumber and pageSize
		pageNumber, _ := strconv.Atoi(c.DefaultQuery("pageNumber", "1"))
		pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "200"))

		// Ensure pageNumber is at least 1
		if pageNumber < 1 {
			pageNumber = 1
		}
		// Ensure pageSize is positive
		if pageSize < 1 {
			pageSize = 200
		}

		// Convert to offset/limit
		limit = pageSize
		offset = (pageNumber - 1) * pageSize
	} else {
		// Old pagination style with offset/limit (backward compatibility)
		limitStr := c.DefaultQuery("limit", "200")
		offsetStr := c.DefaultQuery("offset", "0")
		limit, _ = strconv.Atoi(limitStr)
		offset, _ = strconv.Atoi(offsetStr)
	}

	// Convert user context to pointers
	var userID, realmID *string
	if userIDStr != "" {
		userID = &userIDStr
	}
	if realmIDStr != "" {
		realmID = &realmIDStr
	}

	// Get prompts
	prompts, total, err := h.promptService.ListPrompts(userID, realmID, scope, nameFilter, tagsFilter, sortBy, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Calculate pageNumber and pageSize from offset/limit
	pageSize := limit
	pageNumber := 1
	if limit > 0 {
		pageNumber = (offset / limit) + 1
	}

	// Calculate total pages
	totalPages := 0
	if pageSize > 0 {
		totalPages = (int(total) + pageSize - 1) / pageSize
	}

	c.JSON(http.StatusOK, gin.H{
		"data":        prompts,
		"total":       total,
		"pageNumber":  pageNumber,
		"pageSize":    pageSize,
		"totalPages":  totalPages,
		"limit":       limit,   // Keep for backward compatibility
		"offset":      offset,  // Keep for backward compatibility
		"auth_type":   authType,
		"username":    username,
	})
}

// GetPrompt godoc
// @Summary Get prompt by ID or Name
// @Description Get a specific prompt by ID (UUID) or Name
// @Tags prompts
// @Produce json
// @Param id path string true "Prompt ID or Name"
// @Success 200 {object} models.Prompt
// @Router /api/v1/prompts/{id} [get]
func (h *PromptHandlers) GetPrompt(c *gin.Context) {
	id := c.Param("id")
	_, userIDStr, realmIDStr, _, _ := GetUserContext(c)

	var userID, realmID *string
	if userIDStr != "" {
		userID = &userIDStr
	}
	if realmIDStr != "" {
		realmID = &realmIDStr
	}

	prompt, err := h.promptService.GetPromptByID(id, userID, realmID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, prompt)
}

// CreatePromptRequest represents the request body for creating a prompt
type CreatePromptRequest struct {
	Name         string                  `json:"name" binding:"required"`
	Title        string                  `json:"title"`
	Description  string                  `json:"description"`
	SystemPrompt string                  `json:"system_prompt"`
	UserPrompt   string                  `json:"user_prompt"`
	Arguments    []models.PromptArgument `json:"arguments"`
	Tags         string                  `json:"tags"`
	Scope        string                  `json:"scope"` // "personal" or "shared"
}

// CreatePrompt godoc
// @Summary Create prompt
// @Description Create a new prompt
// @Tags prompts
// @Accept json
// @Produce json
// @Param prompt body CreatePromptRequest true "Prompt data"
// @Success 201 {object} models.Prompt
// @Router /api/v1/prompts [post]
func (h *PromptHandlers) CreatePrompt(c *gin.Context) {
	var req CreatePromptRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	_, userIDStr, realmIDStr, username, _ := GetUserContext(c)

	// Determine ownership based on scope
	var userID, realmID *string
	if req.Scope == "shared" {
		// Shared prompt: no user_id, has realm_id
		if realmIDStr != "" {
			realmID = &realmIDStr
		}
	} else {
		// Personal prompt: has user_id and realm_id
		if userIDStr != "" {
			userID = &userIDStr
		}
		if realmIDStr != "" {
			realmID = &realmIDStr
		}
	}

	// Marshal arguments to JSON string
	var argumentsJSON string
	if len(req.Arguments) > 0 {
		argsBytes, err := json.Marshal(req.Arguments)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to marshal arguments"})
			return
		}
		argumentsJSON = string(argsBytes)
	}

	prompt, err := h.promptService.CreatePrompt(
		req.Name,
		req.Title,
		req.Description,
		req.SystemPrompt,
		req.UserPrompt,
		argumentsJSON,
		req.Tags,
		userID,
		realmID,
		username,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, prompt)
}

// UpdatePromptRequest represents the request body for updating a prompt
type UpdatePromptRequest struct {
	Name         string                  `json:"name" binding:"required"`
	Title        string                  `json:"title"`
	Description  string                  `json:"description"`
	SystemPrompt string                  `json:"system_prompt"`
	UserPrompt   string                  `json:"user_prompt"`
	Arguments    []models.PromptArgument `json:"arguments"`
	Tags         string                  `json:"tags"`
}

// UpdatePrompt godoc
// @Summary Update prompt
// @Description Update an existing prompt by ID (UUID) or Name
// @Tags prompts
// @Accept json
// @Produce json
// @Param id path string true "Prompt ID or Name"
// @Param prompt body UpdatePromptRequest true "Prompt data"
// @Success 200 {object} models.Prompt
// @Router /api/v1/prompts/{id} [put]
func (h *PromptHandlers) UpdatePrompt(c *gin.Context) {
	id := c.Param("id")

	var req UpdatePromptRequest
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

	// Marshal arguments to JSON string
	var argumentsJSON string
	if len(req.Arguments) > 0 {
		argsBytes, err := json.Marshal(req.Arguments)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to marshal arguments"})
			return
		}
		argumentsJSON = string(argsBytes)
	}

	prompt, err := h.promptService.UpdatePrompt(
		id,
		req.Name,
		req.Title,
		req.Description,
		req.SystemPrompt,
		req.UserPrompt,
		argumentsJSON,
		req.Tags,
		username,
		userID,
		realmID,
	)

	if err != nil {
		errMsg := err.Error()
		if errMsg == "unauthorized: you can only update your own prompts" ||
		   errMsg == "unauthorized: prompt belongs to different realm" {
			c.JSON(http.StatusForbidden, gin.H{"error": errMsg})
		} else if errMsg == "prompt not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": errMsg})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": errMsg})
		}
		return
	}

	c.JSON(http.StatusOK, prompt)
}

// DeletePrompt godoc
// @Summary Delete prompt
// @Description Delete a prompt (soft delete) by ID (UUID) or Name
// @Tags prompts
// @Produce json
// @Param id path string true "Prompt ID or Name"
// @Success 204
// @Router /api/v1/prompts/{id} [delete]
func (h *PromptHandlers) DeletePrompt(c *gin.Context) {
	id := c.Param("id")
	_, userIDStr, realmIDStr, _, _ := GetUserContext(c)

	var userID, realmID *string
	if userIDStr != "" {
		userID = &userIDStr
	}
	if realmIDStr != "" {
		realmID = &realmIDStr
	}

	err := h.promptService.DeletePrompt(id, userID, realmID)
	if err != nil {
		errMsg := err.Error()
		if errMsg == "unauthorized: you can only delete your own prompts" ||
		   errMsg == "unauthorized: prompt belongs to different realm" {
			c.JSON(http.StatusForbidden, gin.H{"error": errMsg})
		} else if errMsg == "prompt not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": errMsg})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": errMsg})
		}
		return
	}

	c.Status(http.StatusNoContent)
}
