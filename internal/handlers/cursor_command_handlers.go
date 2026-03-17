package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/walterfan/lazy-ai-coder/internal/models"
	"github.com/walterfan/lazy-ai-coder/internal/services"
)

// CursorCommandHandlers handles cursor command CRUD endpoints
type CursorCommandHandlers struct {
	cursorCommandService   *services.CursorCommandService
	cursorCommandGenerator *services.CursorCommandGenerator
	db                     *gorm.DB
}

// NewCursorCommandHandlers creates new cursor command handlers
func NewCursorCommandHandlers(db *gorm.DB) *CursorCommandHandlers {
	return &CursorCommandHandlers{
		cursorCommandService:   services.NewCursorCommandService(db),
		cursorCommandGenerator: services.NewCursorCommandGenerator(),
		db:                     db,
	}
}

// ListCursorCommands godoc
// @Summary List cursor commands
// @Description Get list of cursor commands with filtering and pagination
// @Tags cursor-commands
// @Produce json
// @Param scope query string false "Scope: all, personal, shared, templates" default(all)
// @Param q query string false "Search in name, description, command"
// @Param tags query string false "Filter by tags"
// @Param category query string false "Filter by category"
// @Param language query string false "Filter by language"
// @Param framework query string false "Filter by framework"
// @Param sort query string false "Sort by: created_at, updated_at, name, usage_count" default(created_at)
// @Param limit query int false "Limit results" default(50)
// @Param offset query int false "Offset for pagination" default(0)
// @Success 200 {array} models.CursorCommand
// @Router /api/v1/cursor-commands [get]
func (h *CursorCommandHandlers) ListCursorCommands(c *gin.Context) {
	authType, userIDStr, realmIDStr, username, _ := GetUserContext(c)

	scopeStr := c.DefaultQuery("scope", "all")
	scope := services.CursorCommandScope(scopeStr)
	nameFilter := c.Query("q")
	tagsFilter := c.Query("tags")
	categoryFilter := c.Query("category")
	languageFilter := c.Query("language")
	frameworkFilter := c.Query("framework")
	sortBy := c.DefaultQuery("sort", "created_at")
	limitStr := c.DefaultQuery("limit", "50")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, _ := strconv.Atoi(limitStr)
	offset, _ := strconv.Atoi(offsetStr)

	var userID, realmID *string
	if userIDStr != "" {
		userID = &userIDStr
	}
	if realmIDStr != "" {
		realmID = &realmIDStr
	}

	commands, total, err := h.cursorCommandService.ListCursorCommands(userID, realmID, scope, nameFilter, tagsFilter, categoryFilter, languageFilter, frameworkFilter, sortBy, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":      commands,
		"total":     total,
		"limit":     limit,
		"offset":    offset,
		"auth_type": authType,
		"username":  username,
	})
}

// GetCursorCommand godoc
// @Summary Get cursor command by ID or Name
// @Description Get a specific cursor command by ID (UUID) or Name
// @Tags cursor-commands
// @Produce json
// @Param id path string true "Cursor Command ID or Name"
// @Success 200 {object} models.CursorCommand
// @Router /api/v1/cursor-commands/{id} [get]
func (h *CursorCommandHandlers) GetCursorCommand(c *gin.Context) {
	id := c.Param("id")
	_, userIDStr, realmIDStr, _, _ := GetUserContext(c)

	var userID, realmID *string
	if userIDStr != "" {
		userID = &userIDStr
	}
	if realmIDStr != "" {
		realmID = &realmIDStr
	}

	cmd, err := h.cursorCommandService.GetCursorCommandByID(id, userID, realmID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, cmd)
}

// CreateCursorCommandRequest represents the request body for creating a cursor command
type CreateCursorCommandRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
	Command     string `json:"command" binding:"required"`
	Category    string `json:"category"`
	Language    string `json:"language"`
	Framework   string `json:"framework"`
	Tags        string `json:"tags"`
	IsTemplate  bool   `json:"is_template"`
	Scope       string `json:"scope"` // "personal" or "shared"
}

// CreateCursorCommand godoc
// @Summary Create cursor command
// @Description Create a new cursor command
// @Tags cursor-commands
// @Accept json
// @Produce json
// @Param command body CreateCursorCommandRequest true "Cursor Command data"
// @Success 201 {object} models.CursorCommand
// @Router /api/v1/cursor-commands [post]
func (h *CursorCommandHandlers) CreateCursorCommand(c *gin.Context) {
	var req CreateCursorCommandRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	_, userIDStr, realmIDStr, username, _ := GetUserContext(c)

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

	cmd, err := h.cursorCommandService.CreateCursorCommand(
		req.Name,
		req.Description,
		req.Command,
		req.Category,
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

	c.JSON(http.StatusCreated, cmd)
}

// UpdateCursorCommandRequest represents the request body for updating a cursor command
type UpdateCursorCommandRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
	Command     string `json:"command" binding:"required"`
	Category    string `json:"category"`
	Language    string `json:"language"`
	Framework   string `json:"framework"`
	Tags        string `json:"tags"`
	IsTemplate  bool   `json:"is_template"`
}

// UpdateCursorCommand godoc
// @Summary Update cursor command
// @Description Update an existing cursor command by ID (UUID) or Name
// @Tags cursor-commands
// @Accept json
// @Produce json
// @Param id path string true "Cursor Command ID or Name"
// @Param command body UpdateCursorCommandRequest true "Cursor Command data"
// @Success 200 {object} models.CursorCommand
// @Router /api/v1/cursor-commands/{id} [put]
func (h *CursorCommandHandlers) UpdateCursorCommand(c *gin.Context) {
	id := c.Param("id")

	var req UpdateCursorCommandRequest
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

	cmd, err := h.cursorCommandService.UpdateCursorCommand(
		id,
		req.Name,
		req.Description,
		req.Command,
		req.Category,
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
		if errMsg == "unauthorized: you can only update your own cursor commands" ||
		   errMsg == "unauthorized: cursor command belongs to different realm" {
			c.JSON(http.StatusForbidden, gin.H{"error": errMsg})
		} else if errMsg == "cursor command not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": errMsg})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": errMsg})
		}
		return
	}

	c.JSON(http.StatusOK, cmd)
}

// DeleteCursorCommand godoc
// @Summary Delete cursor command
// @Description Delete a cursor command (soft delete) by ID (UUID) or Name
// @Tags cursor-commands
// @Produce json
// @Param id path string true "Cursor Command ID or Name"
// @Success 204
// @Router /api/v1/cursor-commands/{id} [delete]
func (h *CursorCommandHandlers) DeleteCursorCommand(c *gin.Context) {
	id := c.Param("id")
	_, userIDStr, realmIDStr, _, _ := GetUserContext(c)

	var userID, realmID *string
	if userIDStr != "" {
		userID = &userIDStr
	}
	if realmIDStr != "" {
		realmID = &realmIDStr
	}

	err := h.cursorCommandService.DeleteCursorCommand(id, userID, realmID)
	if err != nil {
		errMsg := err.Error()
		if errMsg == "unauthorized: you can only delete your own cursor commands" ||
		   errMsg == "unauthorized: cursor command belongs to different realm" {
			c.JSON(http.StatusForbidden, gin.H{"error": errMsg})
		} else if errMsg == "cursor command not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": errMsg})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": errMsg})
		}
		return
	}

	c.Status(http.StatusNoContent)
}

// GenerateCursorCommandRequest represents the request body for generating a cursor command
type GenerateCursorCommandRequest struct {
	Category    string          `json:"category,omitempty"`
	Language    string          `json:"language,omitempty"`
	Framework   string          `json:"framework,omitempty"`
	Requirements string         `json:"requirements,omitempty"`
	TemplateID  string          `json:"template_id,omitempty"`
	Settings    models.Settings `json:"settings"`
}

// GenerateCursorCommand godoc
// @Summary Generate cursor command
// @Description Generate a new cursor command using AI
// @Tags cursor-commands
// @Accept json
// @Produce json
// @Param request body GenerateCursorCommandRequest true "Generation request"
// @Success 200 {object} map[string]string
// @Router /api/v1/cursor-commands/generate [post]
func (h *CursorCommandHandlers) GenerateCursorCommand(c *gin.Context) {
	var req GenerateCursorCommandRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var command string
	var err error

	if req.TemplateID != "" {
		template, err := h.cursorCommandService.GetCursorCommandByID(req.TemplateID, nil, nil)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Template not found"})
			return
		}
		command, err = h.cursorCommandGenerator.GenerateFromTemplate(template, req.Requirements, req.Settings)
	} else {
		command, err = h.cursorCommandGenerator.GenerateFromScratch(req.Category, req.Language, req.Framework, req.Requirements, req.Settings)
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"command": command})
}

// RefineCursorCommandRequest represents the request body for refining a cursor command
type RefineCursorCommandRequest struct {
	Improvements string          `json:"improvements,omitempty"`
	FocusAreas   []string        `json:"focus_areas,omitempty"`
	Settings     models.Settings `json:"settings"`
}

// RefineCursorCommand godoc
// @Summary Refine cursor command
// @Description Refine an existing cursor command using AI
// @Tags cursor-commands
// @Accept json
// @Produce json
// @Param id path string true "Cursor Command ID"
// @Param request body RefineCursorCommandRequest true "Refinement request"
// @Success 200 {object} map[string]string
// @Router /api/v1/cursor-commands/{id}/refine [post]
func (h *CursorCommandHandlers) RefineCursorCommand(c *gin.Context) {
	id := c.Param("id")

	var req RefineCursorCommandRequest
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

	cmd, err := h.cursorCommandService.GetCursorCommandByID(id, userID, realmID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	command, err := h.cursorCommandGenerator.RefineCommand(cmd, req.Improvements, req.FocusAreas, req.Settings)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"command": command})
}

// ExportCursorCommand godoc
// @Summary Export cursor command
// @Description Export a cursor command as text file
// @Tags cursor-commands
// @Produce text/plain
// @Param id path string true "Cursor Command ID"
// @Success 200 {string} string
// @Router /api/v1/cursor-commands/{id}/export [get]
func (h *CursorCommandHandlers) ExportCursorCommand(c *gin.Context) {
	id := c.Param("id")
	_, userIDStr, realmIDStr, _, _ := GetUserContext(c)

	var userID, realmID *string
	if userIDStr != "" {
		userID = &userIDStr
	}
	if realmIDStr != "" {
		realmID = &realmIDStr
	}

	cmd, err := h.cursorCommandService.GetCursorCommandByID(id, userID, realmID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.Header("Content-Type", "text/plain")
	c.Header("Content-Disposition", "attachment; filename="+cmd.Name+".txt")
	c.String(http.StatusOK, cmd.Command)
}

// ImportCursorCommandRequest represents the request body for importing a cursor command
type ImportCursorCommandRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
	Command     string `json:"command" binding:"required"`
	Category    string `json:"category"`
	Language    string `json:"language"`
	Framework   string `json:"framework"`
	Tags        string `json:"tags"`
	Scope       string `json:"scope"` // "personal" or "shared"
}

// ImportCursorCommand godoc
// @Summary Import cursor command
// @Description Import a cursor command from text file content
// @Tags cursor-commands
// @Accept json
// @Produce json
// @Param command body ImportCursorCommandRequest true "Cursor Command data"
// @Success 201 {object} models.CursorCommand
// @Router /api/v1/cursor-commands/import [post]
func (h *CursorCommandHandlers) ImportCursorCommand(c *gin.Context) {
	var req ImportCursorCommandRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	_, userIDStr, realmIDStr, username, _ := GetUserContext(c)

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

	cmd, err := h.cursorCommandService.CreateCursorCommand(
		req.Name,
		req.Description,
		req.Command,
		req.Category,
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

	c.JSON(http.StatusCreated, cmd)
}

