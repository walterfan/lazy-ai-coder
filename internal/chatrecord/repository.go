package chatrecord

import (
	"context"
	"time"

	"github.com/walterfan/lazy-ai-coder/pkg/models"
)

// ListFilters contains filter options for listing learning records
type ListFilters struct {
	Type     string     // Filter by input type (word, sentence, question, idea)
	Search   string     // Search in user_input and response_payload
	DateFrom *time.Time // Filter records created after this time
	DateTo   *time.Time // Filter records created before this time
}

// ListResult contains the paginated list result
type ListResult struct {
	Records    []models.ChatRecord
	Total      int64
	Page       int
	PageSize   int
	TotalPages int
}

// Stats contains learning record statistics
type Stats struct {
	Total        int64            `json:"total"`
	ByType       map[string]int64 `json:"by_type"`
	Streak       int              `json:"streak"`         // Consecutive days with at least one record
	LastRecordAt *time.Time       `json:"last_record_at"` // Time of the most recent record
}

// Repository defines the interface for learning record data access
type Repository interface {
	// Create creates a new learning record
	Create(ctx context.Context, record *models.ChatRecord) error

	// FindByID finds a learning record by ID
	FindByID(ctx context.Context, id string) (*models.ChatRecord, error)

	// FindByUserWithFilters finds learning records for a user with filters and pagination
	FindByUserWithFilters(ctx context.Context, userID string, filters ListFilters, page, pageSize int) (*ListResult, error)

	// SoftDelete soft-deletes a learning record
	SoftDelete(ctx context.Context, id string) error

	// CountByType counts learning records by input type for a user
	CountByType(ctx context.Context, userID string) (map[string]int64, error)

	// CountStreak counts the consecutive days with at least one record
	CountStreak(ctx context.Context, userID string) (int, error)

	// GetLastRecordTime gets the time of the most recent record for a user
	GetLastRecordTime(ctx context.Context, userID string) (*time.Time, error)

	// FindSimilar finds learning records with similar user_input (for similarity retriever)
	FindSimilar(ctx context.Context, userID string, input string, limit int) ([]models.ChatRecord, error)
}
