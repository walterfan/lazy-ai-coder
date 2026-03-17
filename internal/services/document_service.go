package services

import (
	"errors"
	"fmt"
	"strings"

	"gorm.io/gorm"

	"github.com/walterfan/lazy-ai-coder/pkg/models"
)

// DocumentService handles document CRUD operations
type DocumentService struct {
	db *gorm.DB
}

// NewDocumentService creates a new document service
func NewDocumentService(db *gorm.DB) *DocumentService {
	return &DocumentService{db: db}
}

// DocumentScope represents the scope for querying documents
type DocumentScope string

const (
	DocumentScopeAll      DocumentScope = "all"      // Personal + Shared
	DocumentScopePersonal DocumentScope = "personal" // User's personal documents
	DocumentScopeShared   DocumentScope = "shared"   // Realm shared documents
)

// ListDocuments retrieves documents based on scope and filters
func (s *DocumentService) ListDocuments(userID *string, realmID string, projectID *string, scope DocumentScope, nameFilter string, sortBy string, limit, offset int) ([]models.Document, int64, error) {
	query := s.db.Model(&models.Document{}).Where("documents.deleted_at IS NULL")

	// Apply realm filter
	query = query.Where("documents.realm_id = ?", realmID)

	// Apply project filter if specified
	if projectID != nil && *projectID != "" {
		query = query.Where("documents.project_id = ?", *projectID)
	}

	// Apply scope filtering (based on project ownership)
	switch scope {
	case DocumentScopePersonal:
		if userID == nil || *userID == "" {
			return nil, 0, errors.New("user_id required for personal scope")
		}
		// Join with projects to filter by user
		query = query.Joins("LEFT JOIN projects ON documents.project_id = projects.id").
			Where("projects.user_id = ?", *userID)

	case DocumentScopeShared:
		// Join with projects to filter shared projects
		query = query.Joins("LEFT JOIN projects ON documents.project_id = projects.id").
			Where("projects.user_id IS NULL")

	case DocumentScopeAll:
		// Return personal + shared
		if userID != nil && *userID != "" {
			query = query.Joins("LEFT JOIN projects ON documents.project_id = projects.id").
				Where("(projects.user_id = ?) OR (projects.user_id IS NULL)", *userID)
		}
		// If no userID, return all (guest mode or admin)
	}

	// Apply name filter
	if nameFilter != "" {
		query = query.Where("LOWER(documents.name) LIKE ? OR LOWER(documents.path) LIKE ? OR LOWER(documents.content) LIKE ?",
			"%"+strings.ToLower(nameFilter)+"%",
			"%"+strings.ToLower(nameFilter)+"%",
			"%"+strings.ToLower(nameFilter)+"%")
	}

	// Get total count
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count documents: %w", err)
	}

	// Apply sorting
	switch sortBy {
	case "name":
		query = query.Order("documents.name ASC")
	case "updated_at":
		query = query.Order("documents.updated_time DESC")
	default:
		query = query.Order("documents.created_time DESC")
	}

	// Apply pagination
	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	var documents []models.Document
	// Select only necessary fields (exclude large embedding field)
	// Use SUBSTR for SQLite compatibility (works in MySQL and PostgreSQL too)
	if err := query.Select("documents.id, documents.realm_id, documents.project_id, documents.name, documents.path, SUBSTR(documents.content, 1, 200) as content, documents.created_by, documents.created_time, documents.updated_by, documents.updated_time").Find(&documents).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list documents: %w", err)
	}

	return documents, total, nil
}

// GetDocumentByID retrieves a document by ID
func (s *DocumentService) GetDocumentByID(id string, userID *string, realmID string) (*models.Document, error) {
	var document models.Document
	query := s.db.Where("documents.id = ? AND documents.deleted_at IS NULL", id)

	// Access control: user can only access their own documents or shared documents
	if userID != nil && *userID != "" {
		query = query.Where("documents.realm_id = ?", realmID).
			Joins("LEFT JOIN projects ON documents.project_id = projects.id").
			Where("(projects.user_id = ?) OR (projects.user_id IS NULL)", *userID)
	} else {
		query = query.Where("documents.realm_id = ?", realmID)
	}

	if err := query.First(&document).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("document not found")
		}
		return nil, fmt.Errorf("failed to get document: %w", err)
	}

	return &document, nil
}

// GetDocumentChunks retrieves all chunks for a specific document (by path and project)
func (s *DocumentService) GetDocumentChunks(projectID, documentPath string, userID *string, realmID string) ([]models.Document, error) {
	query := s.db.Where("documents.project_id = ? AND documents.path = ? AND documents.deleted_at IS NULL", projectID, documentPath)

	// Access control
	query = query.Where("documents.realm_id = ?", realmID)
	if userID != nil && *userID != "" {
		query = query.Joins("LEFT JOIN projects ON documents.project_id = projects.id").
			Where("(projects.user_id = ?) OR (projects.user_id IS NULL)", *userID)
	}

	var chunks []models.Document
	if err := query.Order("documents.created_time ASC").Find(&chunks).Error; err != nil {
		return nil, fmt.Errorf("failed to get document chunks: %w", err)
	}

	return chunks, nil
}

// DeleteDocument soft-deletes a document
func (s *DocumentService) DeleteDocument(id string, userID *string, realmID string) error {
	// Get existing document
	document, err := s.GetDocumentByID(id, userID, realmID)
	if err != nil {
		return err
	}

	// Authorization: check via project ownership
	if userID != nil && *userID != "" {
		var project models.Project
		if err := s.db.Where("id = ?", document.ProjectID).First(&project).Error; err != nil {
			return fmt.Errorf("failed to check project ownership: %w", err)
		}

		if project.UserID != nil && *project.UserID != *userID {
			return errors.New("unauthorized: you can only delete your own documents")
		}
	}

	// Soft delete
	if err := s.db.Model(&models.Document{}).Where("id = ?", id).Update("deleted_at", gorm.Expr("CURRENT_TIMESTAMP")).Error; err != nil {
		return fmt.Errorf("failed to delete document: %w", err)
	}

	return nil
}

// DeleteDocumentsByPath deletes all chunks for a specific document (by path and project)
func (s *DocumentService) DeleteDocumentsByPath(projectID, documentPath string, userID *string, realmID string) (int64, error) {
	// Authorization: check via project ownership
	if userID != nil && *userID != "" {
		var project models.Project
		if err := s.db.Where("id = ?", projectID).First(&project).Error; err != nil {
			return 0, fmt.Errorf("failed to check project ownership: %w", err)
		}

		if project.UserID != nil && *project.UserID != *userID {
			return 0, errors.New("unauthorized: you can only delete your own documents")
		}
	}

	// Soft delete all chunks
	result := s.db.Model(&models.Document{}).
		Where("project_id = ? AND path = ? AND realm_id = ? AND deleted_at IS NULL", projectID, documentPath, realmID).
		Update("deleted_at", gorm.Expr("CURRENT_TIMESTAMP"))

	if result.Error != nil {
		return 0, fmt.Errorf("failed to delete document chunks: %w", result.Error)
	}

	return result.RowsAffected, nil
}

// GetDocumentStats retrieves statistics about documents
func (s *DocumentService) GetDocumentStats(projectID *string, userID *string, realmID string) (map[string]interface{}, error) {
	query := s.db.Model(&models.Document{}).Where("documents.deleted_at IS NULL AND documents.realm_id = ?", realmID)

	if projectID != nil && *projectID != "" {
		query = query.Where("documents.project_id = ?", *projectID)
	}

	// Apply access control
	if userID != nil && *userID != "" {
		query = query.Joins("LEFT JOIN projects ON documents.project_id = projects.id").
			Where("(projects.user_id = ?) OR (projects.user_id IS NULL)", *userID)
	}

	var totalChunks int64
	if err := query.Count(&totalChunks).Error; err != nil {
		return nil, fmt.Errorf("failed to count documents: %w", err)
	}

	// Count unique documents (by path)
	var uniqueDocs int64
	if err := query.Distinct("documents.path").Count(&uniqueDocs).Error; err != nil {
		return nil, fmt.Errorf("failed to count unique documents: %w", err)
	}

	stats := map[string]interface{}{
		"total_chunks":     totalChunks,
		"unique_documents": uniqueDocs,
	}

	return stats, nil
}
