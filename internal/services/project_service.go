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

// ProjectService handles project CRUD operations with user isolation
type ProjectService struct {
	db *gorm.DB
}

// NewProjectService creates a new project service
func NewProjectService(db *gorm.DB) *ProjectService {
	return &ProjectService{db: db}
}

// ProjectScope represents the scope for querying projects
type ProjectScope string

const (
	ProjectScopeAll      ProjectScope = "all"      // Personal + Shared
	ProjectScopePersonal ProjectScope = "personal" // User's personal projects
	ProjectScopeShared   ProjectScope = "shared"   // Realm shared projects
)

// ListProjects retrieves projects based on scope and filters
func (s *ProjectService) ListProjects(userID *string, realmID string, scope ProjectScope, nameFilter, languageFilter string, sortBy string, limit, offset int) ([]models.Project, int64, error) {
	query := s.db.Model(&models.Project{}).Where("deleted_at IS NULL")

	// Apply scope filtering
	switch scope {
	case ProjectScopePersonal:
		if userID == nil || *userID == "" {
			return nil, 0, errors.New("user_id required for personal scope")
		}
		query = query.Where("user_id = ?", *userID)

	case ProjectScopeShared:
		query = query.Where("realm_id = ? AND user_id IS NULL", realmID)

	case ProjectScopeAll:
		// Return personal + shared
		if userID != nil && *userID != "" {
			query = query.Where(
				"(user_id = ?) OR (realm_id = ? AND user_id IS NULL)",
				*userID, realmID,
			)
		} else {
			// Only shared projects
			query = query.Where("realm_id = ? AND user_id IS NULL", realmID)
		}
	}

	// Apply name filter
	if nameFilter != "" {
		query = query.Where("LOWER(name) LIKE ? OR LOWER(description) LIKE ?",
			"%"+strings.ToLower(nameFilter)+"%",
			"%"+strings.ToLower(nameFilter)+"%")
	}

	// Apply language filter
	if languageFilter != "" {
		query = query.Where("LOWER(language) = ?", strings.ToLower(languageFilter))
	}

	// Get total count
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count projects: %w", err)
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

	var projects []models.Project
	if err := query.Find(&projects).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list projects: %w", err)
	}

	return projects, total, nil
}

// GetProjectByID retrieves a project by ID
func (s *ProjectService) GetProjectByID(id string, userID *string, realmID string) (*models.Project, error) {
	var project models.Project
	query := s.db.Where("id = ? AND deleted_at IS NULL", id)

	// Access control: user can only access their own projects or shared projects
	if userID != nil && *userID != "" {
		query = query.Where(
			"(user_id = ?) OR (realm_id = ? AND user_id IS NULL)",
			*userID, realmID,
		)
	} else {
		query = query.Where("realm_id = ? AND user_id IS NULL", realmID)
	}

	if err := query.First(&project).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("project not found")
		}
		return nil, fmt.Errorf("failed to get project: %w", err)
	}

	return &project, nil
}

// CreateProject creates a new project
func (s *ProjectService) CreateProject(
	name, description, gitURL, gitRepo, gitBranch, language, entryPoint string,
	userID *string, realmID, createdBy string,
) (*models.Project, error) {
	project := &models.Project{
		ID:          uuid.New().String(),
		UserID:      userID,
		RealmID:     realmID,
		Name:        name,
		Description: description,
		GitURL:      gitURL,
		GitRepo:     gitRepo,
		GitBranch:   gitBranch,
		Language:    language,
		EntryPoint:  entryPoint,
		CreatedBy:   createdBy,
		UpdatedBy:   createdBy,
	}

	if err := s.db.Create(project).Error; err != nil {
		return nil, fmt.Errorf("failed to create project: %w", err)
	}

	return project, nil
}

// UpdateProject updates an existing project
func (s *ProjectService) UpdateProject(
	id, name, description, gitURL, gitRepo, gitBranch, language, entryPoint, updatedBy string,
	userID *string, realmID string,
) (*models.Project, error) {
	// Get existing project
	project, err := s.GetProjectByID(id, userID, realmID)
	if err != nil {
		return nil, err
	}

	// Authorization check
	userIDStr := ""
	if userID != nil {
		userIDStr = *userID
	}

	// Super admins can edit any project
	isSuperAdmin := authz.IsSuperAdmin(s.db, userIDStr)

	if !isSuperAdmin {
		// Non-super-admin authorization: user can only update their own projects
		if userID != nil && *userID != "" {
			if project.UserID == nil || *project.UserID != *userID {
				return nil, errors.New("unauthorized: you can only update your own projects")
			}
		}

		// Additional realm isolation check for non-super-admins
		if realmID != "" {
			if project.RealmID != realmID {
				return nil, errors.New("unauthorized: project belongs to different realm")
			}
		}
	}

	// Update fields
	updates := map[string]interface{}{
		"name":        name,
		"description": description,
		"git_url":     gitURL,
		"git_repo":    gitRepo,
		"git_branch":  gitBranch,
		"language":    language,
		"entry_point": entryPoint,
		"updated_by":  updatedBy,
	}

	if err := s.db.Model(&models.Project{}).Where("id = ?", id).Updates(updates).Error; err != nil {
		return nil, fmt.Errorf("failed to update project: %w", err)
	}

	// Fetch updated project
	return s.GetProjectByID(id, userID, realmID)
}

// DeleteProject soft-deletes a project
func (s *ProjectService) DeleteProject(id string, userID *string, realmID string) error {
	// Get existing project
	project, err := s.GetProjectByID(id, userID, realmID)
	if err != nil {
		return err
	}

	// Authorization check
	userIDStr := ""
	if userID != nil {
		userIDStr = *userID
	}

	// Super admins can delete any project
	isSuperAdmin := authz.IsSuperAdmin(s.db, userIDStr)

	if !isSuperAdmin {
		// Non-super-admin authorization: user can only delete their own projects
		if userID != nil && *userID != "" {
			if project.UserID == nil || *project.UserID != *userID {
				return errors.New("unauthorized: you can only delete your own projects")
			}
		}

		// Additional realm isolation check for non-super-admins
		if realmID != "" {
			if project.RealmID != realmID {
				return errors.New("unauthorized: project belongs to different realm")
			}
		}
	}

	// Soft delete
	if err := s.db.Model(&models.Project{}).Where("id = ?", id).Update("deleted_at", gorm.Expr("CURRENT_TIMESTAMP")).Error; err != nil {
		return fmt.Errorf("failed to delete project: %w", err)
	}

	return nil
}

// ProjectYAML represents the structure of projects.yaml
type ProjectYAML struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
	GitURL      string `yaml:"gitUrl"`
	GitRepo     string `yaml:"project"`
	GitBranch   string `yaml:"branch"`
	EntryPoint  string `yaml:"codePath"`
	Language    string `yaml:"language"`
}

// ExportProjects exports all projects to YAML format
func (s *ProjectService) ExportProjects(userID *string, realmID string, scope ProjectScope) ([]ProjectYAML, error) {
	projects, _, err := s.ListProjects(userID, realmID, scope, "", "", "name", 0, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to list projects: %w", err)
	}

	var projectsYAML []ProjectYAML
	for _, project := range projects {
		projectsYAML = append(projectsYAML, ProjectYAML{
			Name:        project.Name,
			Description: project.Description,
			GitURL:      project.GitURL,
			GitRepo:     project.GitRepo,
			GitBranch:   project.GitBranch,
			EntryPoint:  project.EntryPoint,
			Language:    project.Language,
		})
	}

	return projectsYAML, nil
}

// ImportProjects imports projects from YAML data
func (s *ProjectService) ImportProjects(projectsYAML []ProjectYAML, userID *string, realmID, createdBy string, updateExisting bool) (int, int, int, error) {
	created := 0
	updated := 0
	skipped := 0

	for _, projectData := range projectsYAML {
		// Check if project already exists
		var existingProject models.Project
		result := s.db.Where("name = ? AND realm_id = ? AND deleted_at IS NULL", projectData.Name, realmID).First(&existingProject)

		if result.Error == nil {
			// Project exists
			if updateExisting {
				// Update existing project
				updates := map[string]interface{}{
					"description": projectData.Description,
					"git_url":     projectData.GitURL,
					"git_repo":    projectData.GitRepo,
					"git_branch":  projectData.GitBranch,
					"language":    projectData.Language,
					"entry_point": projectData.EntryPoint,
					"updated_by":  createdBy,
				}

				if err := s.db.Model(&models.Project{}).Where("id = ?", existingProject.ID).Updates(updates).Error; err != nil {
					return created, updated, skipped, fmt.Errorf("failed to update project '%s': %w", projectData.Name, err)
				}
				updated++
			} else {
				// Skip existing project
				skipped++
			}
		} else {
			// Create new project
			project := &models.Project{
				ID:          uuid.New().String(),
				UserID:      userID,
				RealmID:     realmID,
				Name:        projectData.Name,
				Description: projectData.Description,
				GitURL:      projectData.GitURL,
				GitRepo:     projectData.GitRepo,
				GitBranch:   projectData.GitBranch,
				Language:    projectData.Language,
				EntryPoint:  projectData.EntryPoint,
				CreatedBy:   createdBy,
				UpdatedBy:   createdBy,
			}

			if err := s.db.Create(project).Error; err != nil {
				return created, updated, skipped, fmt.Errorf("failed to create project '%s': %w", projectData.Name, err)
			}
			created++
		}
	}

	return created, updated, skipped, nil
}
