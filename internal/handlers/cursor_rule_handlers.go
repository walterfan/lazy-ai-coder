package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/walterfan/lazy-ai-coder/internal/models"
	"github.com/walterfan/lazy-ai-coder/internal/services"
	"github.com/walterfan/lazy-ai-coder/internal/smartprompt"
)

// CursorRuleHandlers handles cursor rule CRUD endpoints
type CursorRuleHandlers struct {
	cursorRuleService     *services.CursorRuleService
	cursorRuleGenerator   *services.CursorRuleGenerator
	smartPromptService    *smartprompt.SmartPromptService
	db                    *gorm.DB
}

// NewCursorRuleHandlers creates new cursor rule handlers
func NewCursorRuleHandlers(db *gorm.DB) *CursorRuleHandlers {
	return &CursorRuleHandlers{
		cursorRuleService:   services.NewCursorRuleService(db),
		cursorRuleGenerator: services.NewCursorRuleGenerator(),
		smartPromptService:  smartprompt.NewSmartPromptService(),
		db:                  db,
	}
}

// ListCursorRules godoc
// @Summary List cursor rules
// @Description Get list of cursor rules with filtering and pagination
// @Tags cursor-rules
// @Produce json
// @Param scope query string false "Scope: all, personal, shared, templates" default(all)
// @Param q query string false "Search in name, description, content"
// @Param tags query string false "Filter by tags"
// @Param language query string false "Filter by language"
// @Param framework query string false "Filter by framework"
// @Param sort query string false "Sort by: created_at, updated_at, name, usage_count" default(created_at)
// @Param limit query int false "Limit results" default(50)
// @Param offset query int false "Offset for pagination" default(0)
// @Success 200 {array} models.CursorRule
// @Router /api/v1/cursor-rules [get]
func (h *CursorRuleHandlers) ListCursorRules(c *gin.Context) {
	// Get user context from flexible auth middleware
	authType, userIDStr, realmIDStr, username, _ := GetUserContext(c)

	// Parse query parameters
	scopeStr := c.DefaultQuery("scope", "all")
	scope := services.CursorRuleScope(scopeStr)
	nameFilter := c.Query("q")
	tagsFilter := c.Query("tags")
	languageFilter := c.Query("language")
	frameworkFilter := c.Query("framework")
	sortBy := c.DefaultQuery("sort", "created_at")
	limitStr := c.DefaultQuery("limit", "50")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, _ := strconv.Atoi(limitStr)
	offset, _ := strconv.Atoi(offsetStr)

	// Convert user context to pointers
	var userID, realmID *string
	if userIDStr != "" {
		userID = &userIDStr
	}
	if realmIDStr != "" {
		realmID = &realmIDStr
	}

	// Get cursor rules
	rules, total, err := h.cursorRuleService.ListCursorRules(userID, realmID, scope, nameFilter, tagsFilter, languageFilter, frameworkFilter, sortBy, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":      rules,
		"total":     total,
		"limit":     limit,
		"offset":    offset,
		"auth_type": authType,
		"username":  username,
	})
}

// GetCursorRule godoc
// @Summary Get cursor rule by ID or Name
// @Description Get a specific cursor rule by ID (UUID) or Name
// @Tags cursor-rules
// @Produce json
// @Param id path string true "Cursor Rule ID or Name"
// @Success 200 {object} models.CursorRule
// @Router /api/v1/cursor-rules/{id} [get]
func (h *CursorRuleHandlers) GetCursorRule(c *gin.Context) {
	id := c.Param("id")
	_, userIDStr, realmIDStr, _, _ := GetUserContext(c)

	var userID, realmID *string
	if userIDStr != "" {
		userID = &userIDStr
	}
	if realmIDStr != "" {
		realmID = &realmIDStr
	}

	rule, err := h.cursorRuleService.GetCursorRuleByID(id, userID, realmID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, rule)
}

// CreateCursorRuleRequest represents the request body for creating a cursor rule
type CreateCursorRuleRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
	Content     string `json:"content" binding:"required"`
	Language    string `json:"language"`
	Framework   string `json:"framework"`
	Tags        string `json:"tags"`
	IsTemplate  bool   `json:"is_template"`
	Scope       string `json:"scope"` // "personal" or "shared"
}

// CreateCursorRule godoc
// @Summary Create cursor rule
// @Description Create a new cursor rule
// @Tags cursor-rules
// @Accept json
// @Produce json
// @Param rule body CreateCursorRuleRequest true "Cursor Rule data"
// @Success 201 {object} models.CursorRule
// @Router /api/v1/cursor-rules [post]
func (h *CursorRuleHandlers) CreateCursorRule(c *gin.Context) {
	var req CreateCursorRuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	_, userIDStr, realmIDStr, username, _ := GetUserContext(c)

	// Determine ownership based on scope
	var userID, realmID *string
	if req.Scope == "shared" {
		// Shared rule: no user_id, has realm_id
		if realmIDStr != "" {
			realmID = &realmIDStr
		}
	} else {
		// Personal rule: has user_id and realm_id
		if userIDStr != "" {
			userID = &userIDStr
		}
		if realmIDStr != "" {
			realmID = &realmIDStr
		}
	}

	rule, err := h.cursorRuleService.CreateCursorRule(
		req.Name,
		req.Description,
		req.Content,
		req.Language,
		req.Framework,
		req.Tags,
		req.IsTemplate,
		userID,
		realmID,
		username,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, rule)
}

// UpdateCursorRuleRequest represents the request body for updating a cursor rule
type UpdateCursorRuleRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
	Content     string `json:"content" binding:"required"`
	Language    string `json:"language"`
	Framework   string `json:"framework"`
	Tags        string `json:"tags"`
	IsTemplate  bool   `json:"is_template"`
}

// UpdateCursorRule godoc
// @Summary Update cursor rule
// @Description Update an existing cursor rule by ID (UUID) or Name
// @Tags cursor-rules
// @Accept json
// @Produce json
// @Param id path string true "Cursor Rule ID or Name"
// @Param rule body UpdateCursorRuleRequest true "Cursor Rule data"
// @Success 200 {object} models.CursorRule
// @Router /api/v1/cursor-rules/{id} [put]
func (h *CursorRuleHandlers) UpdateCursorRule(c *gin.Context) {
	id := c.Param("id")

	var req UpdateCursorRuleRequest
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

	rule, err := h.cursorRuleService.UpdateCursorRule(
		id,
		req.Name,
		req.Description,
		req.Content,
		req.Language,
		req.Framework,
		req.Tags,
		req.IsTemplate,
		username,
		userID,
		realmID,
	)

	if err != nil {
		errMsg := err.Error()
		if errMsg == "unauthorized: you can only update your own cursor rules" ||
		   errMsg == "unauthorized: cursor rule belongs to different realm" {
			c.JSON(http.StatusForbidden, gin.H{"error": errMsg})
		} else if errMsg == "cursor rule not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": errMsg})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": errMsg})
		}
		return
	}

	c.JSON(http.StatusOK, rule)
}

// DeleteCursorRule godoc
// @Summary Delete cursor rule
// @Description Delete a cursor rule (soft delete) by ID (UUID) or Name
// @Tags cursor-rules
// @Produce json
// @Param id path string true "Cursor Rule ID or Name"
// @Success 204
// @Router /api/v1/cursor-rules/{id} [delete]
func (h *CursorRuleHandlers) DeleteCursorRule(c *gin.Context) {
	id := c.Param("id")
	_, userIDStr, realmIDStr, _, _ := GetUserContext(c)

	var userID, realmID *string
	if userIDStr != "" {
		userID = &userIDStr
	}
	if realmIDStr != "" {
		realmID = &realmIDStr
	}

	err := h.cursorRuleService.DeleteCursorRule(id, userID, realmID)
	if err != nil {
		errMsg := err.Error()
		if errMsg == "unauthorized: you can only delete your own cursor rules" ||
		   errMsg == "unauthorized: cursor rule belongs to different realm" {
			c.JSON(http.StatusForbidden, gin.H{"error": errMsg})
		} else if errMsg == "cursor rule not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": errMsg})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": errMsg})
		}
		return
	}

	c.Status(http.StatusNoContent)
}

// GenerateCursorRuleRequest represents the request body for generating a cursor rule
type GenerateCursorRuleRequest struct {
	ProjectContext *models.ProjectContext `json:"project_context,omitempty"`
	Language       string                 `json:"language,omitempty"`
	Framework      string                 `json:"framework,omitempty"`
	Requirements   string                 `json:"requirements,omitempty"`
	TemplateID     string                 `json:"template_id,omitempty"`
	Settings       models.Settings         `json:"settings"`
}

// GenerateCursorRule godoc
// @Summary Generate cursor rule
// @Description Generate a new cursor rule using AI
// @Tags cursor-rules
// @Accept json
// @Produce json
// @Param request body GenerateCursorRuleRequest true "Generation request"
// @Success 200 {object} map[string]string
// @Router /api/v1/cursor-rules/generate [post]
func (h *CursorRuleHandlers) GenerateCursorRule(c *gin.Context) {
	var req GenerateCursorRuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var content string
	var err error

	// Determine generation method
	if req.TemplateID != "" {
		// Generate from template
		template, err := h.cursorRuleService.GetCursorRuleByID(req.TemplateID, nil, nil)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Template not found"})
			return
		}
		content, err = h.cursorRuleGenerator.GenerateFromTemplate(template, req.Requirements, req.Settings)
	} else if req.ProjectContext != nil {
		// Generate from project context
		content, err = h.cursorRuleGenerator.GenerateFromProject(*req.ProjectContext, req.Requirements, req.Settings)
	} else {
		// Generate from scratch
		content, err = h.cursorRuleGenerator.GenerateFromScratch(req.Language, req.Framework, req.Requirements, req.Settings)
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"content": content})
}

// RefineCursorRuleRequest represents the request body for refining a cursor rule
type RefineCursorRuleRequest struct {
	Improvements string   `json:"improvements,omitempty"`
	FocusAreas   []string `json:"focus_areas,omitempty"`
	Settings     models.Settings `json:"settings"`
}

// RefineCursorRule godoc
// @Summary Refine cursor rule
// @Description Refine an existing cursor rule using AI
// @Tags cursor-rules
// @Accept json
// @Produce json
// @Param id path string true "Cursor Rule ID"
// @Param request body RefineCursorRuleRequest true "Refinement request"
// @Success 200 {object} map[string]string
// @Router /api/v1/cursor-rules/{id}/refine [post]
func (h *CursorRuleHandlers) RefineCursorRule(c *gin.Context) {
	id := c.Param("id")

	var req RefineCursorRuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	_, userIDStr, realmIDStr, _, _ := GetUserContext(c)

	var userID, realmID *string
	if userIDStr != "" {
		userID = &userIDStr
	}
	if realmIDStr != "" {
		realmID = &realmIDStr
	}

	rule, err := h.cursorRuleService.GetCursorRuleByID(id, userID, realmID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	content, err := h.cursorRuleGenerator.RefineRule(rule, req.Improvements, req.FocusAreas, req.Settings)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"content": content})
}

// ExportCursorRule godoc
// @Summary Export cursor rule
// @Description Export a cursor rule as .cursorrules file content
// @Tags cursor-rules
// @Produce text/plain
// @Param id path string true "Cursor Rule ID"
// @Success 200 {string} string
// @Router /api/v1/cursor-rules/{id}/export [get]
func (h *CursorRuleHandlers) ExportCursorRule(c *gin.Context) {
	id := c.Param("id")
	_, userIDStr, realmIDStr, _, _ := GetUserContext(c)

	var userID, realmID *string
	if userIDStr != "" {
		userID = &userIDStr
	}
	if realmIDStr != "" {
		realmID = &realmIDStr
	}

	rule, err := h.cursorRuleService.GetCursorRuleByID(id, userID, realmID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.Header("Content-Type", "text/plain")
	c.Header("Content-Disposition", "attachment; filename="+rule.Name+".cursorrules")
	c.String(http.StatusOK, rule.Content)
}

// ImportCursorRuleRequest represents the request body for importing a cursor rule
type ImportCursorRuleRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
	Content     string `json:"content" binding:"required"`
	Language    string `json:"language"`
	Framework   string `json:"framework"`
	Tags        string `json:"tags"`
	Scope       string `json:"scope"` // "personal" or "shared"
}

// ImportCursorRule godoc
// @Summary Import cursor rule
// @Description Import a cursor rule from .cursorrules file content
// @Tags cursor-rules
// @Accept json
// @Produce json
// @Param rule body ImportCursorRuleRequest true "Cursor Rule data"
// @Success 201 {object} models.CursorRule
// @Router /api/v1/cursor-rules/import [post]
func (h *CursorRuleHandlers) ImportCursorRule(c *gin.Context) {
	var req ImportCursorRuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	_, userIDStr, realmIDStr, username, _ := GetUserContext(c)

	// Determine ownership based on scope
	var userID, realmID *string
	if req.Scope == "shared" {
		if realmIDStr != "" {
			realmID = &realmIDStr
		}
	} else {
		if userIDStr != "" {
			userID = &userIDStr
		}
		if realmIDStr != "" {
			realmID = &realmIDStr
		}
	}

	rule, err := h.cursorRuleService.CreateCursorRule(
		req.Name,
		req.Description,
		req.Content,
		req.Language,
		req.Framework,
		req.Tags,
		false, // Not a template by default
		userID,
		realmID,
		username,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, rule)
}

// ValidateCursorRuleRequest represents the request body for validating a cursor rule
type ValidateCursorRuleRequest struct {
	Content string `json:"content" binding:"required"`
}

// ValidateCursorRule godoc
// @Summary Validate cursor rule
// @Description Validate cursor rule syntax and structure
// @Tags cursor-rules
// @Accept json
// @Produce json
// @Param request body ValidateCursorRuleRequest true "Validation request"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/cursor-rules/validate [post]
func (h *CursorRuleHandlers) ValidateCursorRule(c *gin.Context) {
	var req ValidateCursorRuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Basic validation - check if content is not empty
	isValid := len(strings.TrimSpace(req.Content)) > 0
	errors := []string{}

	if !isValid {
		errors = append(errors, "Content cannot be empty")
	}

	// Check for basic markdown structure (optional)
	if isValid && !strings.Contains(req.Content, "#") {
		// Not an error, just a warning
	}

	c.JSON(http.StatusOK, gin.H{
		"valid": isValid,
		"errors": errors,
	})
}

