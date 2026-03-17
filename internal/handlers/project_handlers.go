package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/walterfan/lazy-ai-coder/internal/services"
)

// ProjectHandlers handles project CRUD endpoints
type ProjectHandlers struct {
	projectService *services.ProjectService
	db             *gorm.DB
}

// NewProjectHandlers creates new project handlers
func NewProjectHandlers(db *gorm.DB) *ProjectHandlers {
	return &ProjectHandlers{
		projectService: services.NewProjectService(db),
		db:             db,
	}
}

// ListProjects godoc
// @Summary List projects
// @Description Get list of projects with filtering and pagination
// @Tags projects
// @Produce json
// @Param scope query string false "Scope: all, personal, shared" default(all)
// @Param q query string false "Search in name and description"
// @Param language query string false "Filter by language"
// @Param sort query string false "Sort by: created_at, updated_at, name" default(created_at)
// @Param limit query int false "Limit results" default(50)
// @Param offset query int false "Offset for pagination" default(0)
// @Success 200 {array} models.Project
// @Router /api/v1/projects [get]
func (h *ProjectHandlers) ListProjects(c *gin.Context) {
	// Get user context from flexible auth middleware
	authType, userIDStr, realmIDStr, username, _ := GetUserContext(c)

	// Parse query parameters
	scopeStr := c.DefaultQuery("scope", "all")
	scope := services.ProjectScope(scopeStr)
	nameFilter := c.Query("q")
	languageFilter := c.Query("language")
	sortBy := c.DefaultQuery("sort", "created_at")
	limitStr := c.DefaultQuery("limit", "50")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, _ := strconv.Atoi(limitStr)
	offset, _ := strconv.Atoi(offsetStr)

	// Convert user context to pointers
	var userID *string
	if userIDStr != "" {
		userID = &userIDStr
	}

	// Get projects
	projects, total, err := h.projectService.ListProjects(userID, realmIDStr, scope, nameFilter, languageFilter, sortBy, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":      projects,
		"total":     total,
		"limit":     limit,
		"offset":    offset,
		"auth_type": authType,
		"username":  username,
	})
}

// GetProject godoc
// @Summary Get project by ID
// @Description Get a specific project by ID
// @Tags projects
// @Produce json
// @Param id path string true "Project ID"
// @Success 200 {object} models.Project
// @Router /api/v1/projects/{id} [get]
func (h *ProjectHandlers) GetProject(c *gin.Context) {
	id := c.Param("id")
	_, userIDStr, realmIDStr, _, _ := GetUserContext(c)

	var userID *string
	if userIDStr != "" {
		userID = &userIDStr
	}

	project, err := h.projectService.GetProjectByID(id, userID, realmIDStr)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, project)
}

// CreateProjectRequest represents the request body for creating a project
type CreateProjectRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
	GitURL      string `json:"git_url"`
	GitRepo     string `json:"git_repo"`
	GitBranch   string `json:"git_branch"`
	Language    string `json:"language"`
	EntryPoint  string `json:"entry_point"`
	Scope       string `json:"scope"` // "personal" or "shared"
}

// CreateProject godoc
// @Summary Create project
// @Description Create a new project
// @Tags projects
// @Accept json
// @Produce json
// @Param project body CreateProjectRequest true "Project data"
// @Success 201 {object} models.Project
// @Router /api/v1/projects [post]
func (h *ProjectHandlers) CreateProject(c *gin.Context) {
	var req CreateProjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	_, userIDStr, realmIDStr, username, _ := GetUserContext(c)

	// Determine ownership based on scope
	var userID *string
	if req.Scope != "shared" {
		// Personal project: has user_id
		if userIDStr != "" {
			userID = &userIDStr
		}
	}
	// Shared project: no user_id

	project, err := h.projectService.CreateProject(
		req.Name,
		req.Description,
		req.GitURL,
		req.GitRepo,
		req.GitBranch,
		req.Language,
		req.EntryPoint,
		userID,
		realmIDStr,
		username,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, project)
}

// UpdateProjectRequest represents the request body for updating a project
type UpdateProjectRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
	GitURL      string `json:"git_url"`
	GitRepo     string `json:"git_repo"`
	GitBranch   string `json:"git_branch"`
	Language    string `json:"language"`
	EntryPoint  string `json:"entry_point"`
}

// UpdateProject godoc
// @Summary Update project
// @Description Update an existing project
// @Tags projects
// @Accept json
// @Produce json
// @Param id path string true "Project ID"
// @Param project body UpdateProjectRequest true "Project data"
// @Success 200 {object} models.Project
// @Router /api/v1/projects/{id} [put]
func (h *ProjectHandlers) UpdateProject(c *gin.Context) {
	id := c.Param("id")

	var req UpdateProjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	_, userIDStr, realmIDStr, username, _ := GetUserContext(c)

	var userID *string
	if userIDStr != "" {
		userID = &userIDStr
	}

	project, err := h.projectService.UpdateProject(
		id,
		req.Name,
		req.Description,
		req.GitURL,
		req.GitRepo,
		req.GitBranch,
		req.Language,
		req.EntryPoint,
		username,
		userID,
		realmIDStr,
	)

	if err != nil {
		errMsg := err.Error()
		if errMsg == "unauthorized: you can only update your own projects" ||
		   errMsg == "unauthorized: project belongs to different realm" {
			c.JSON(http.StatusForbidden, gin.H{"error": errMsg})
		} else if errMsg == "project not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": errMsg})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": errMsg})
		}
		return
	}

	c.JSON(http.StatusOK, project)
}

// DeleteProject godoc
// @Summary Delete project
// @Description Delete a project (soft delete)
// @Tags projects
// @Produce json
// @Param id path string true "Project ID"
// @Success 204
// @Router /api/v1/projects/{id} [delete]
func (h *ProjectHandlers) DeleteProject(c *gin.Context) {
	id := c.Param("id")
	_, userIDStr, realmIDStr, _, _ := GetUserContext(c)

	var userID *string
	if userIDStr != "" {
		userID = &userIDStr
	}

	err := h.projectService.DeleteProject(id, userID, realmIDStr)
	if err != nil {
		errMsg := err.Error()
		if errMsg == "unauthorized: you can only delete your own projects" ||
		   errMsg == "unauthorized: project belongs to different realm" {
			c.JSON(http.StatusForbidden, gin.H{"error": errMsg})
		} else if errMsg == "project not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": errMsg})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": errMsg})
		}
		return
	}

	c.Status(http.StatusNoContent)
}

// ExportProjects godoc
// @Summary Export projects to YAML
// @Description Export all accessible projects to YAML format
// @Tags projects
// @Produce application/x-yaml
// @Param scope query string false "Scope: all, personal, shared" default(all)
// @Success 200 {string} string "YAML file"
// @Router /api/v1/projects/export [get]
func (h *ProjectHandlers) ExportProjects(c *gin.Context) {
	_, userIDStr, realmIDStr, _, _ := GetUserContext(c)

	var userID *string
	if userIDStr != "" {
		userID = &userIDStr
	}

	scopeStr := c.DefaultQuery("scope", "all")
	scope := services.ProjectScope(scopeStr)

	projectsYAML, err := h.projectService.ExportProjects(userID, realmIDStr, scope)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Set headers for file download
	filename := fmt.Sprintf("projects_%s.yaml", time.Now().Format("2006-01-02"))
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	c.YAML(http.StatusOK, gin.H{"projects": projectsYAML})
}

// ImportProjectsRequest represents the request body for importing projects
type ImportProjectsRequest struct {
	Projects       []services.ProjectYAML `json:"projects" binding:"required"`
	UpdateExisting bool                   `json:"update_existing"`
	Scope          string                 `json:"scope"` // "personal" or "shared"
}

// ImportProjects godoc
// @Summary Import projects from YAML
// @Description Import projects from YAML data
// @Tags projects
// @Accept json
// @Produce json
// @Param request body ImportProjectsRequest true "Projects data"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/projects/import [post]
func (h *ProjectHandlers) ImportProjects(c *gin.Context) {
	var req ImportProjectsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	_, userIDStr, realmIDStr, username, _ := GetUserContext(c)

	// Determine ownership based on scope
	var userID *string
	if req.Scope != "shared" {
		// Personal project: has user_id
		if userIDStr != "" {
			userID = &userIDStr
		}
	}
	// Shared project: no user_id

	created, updated, skipped, err := h.projectService.ImportProjects(
		req.Projects,
		userID,
		realmIDStr,
		username,
		req.UpdateExisting,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Import completed",
		"created": created,
		"updated": updated,
		"skipped": skipped,
		"total":   len(req.Projects),
	})
}
