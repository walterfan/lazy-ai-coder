package chatrecord

import (
	"context"
	"fmt"
	"time"

	"github.com/walterfan/lazy-ai-coder/pkg/models"
	"gorm.io/gorm"
)

// GormRepository implements Repository using GORM
type GormRepository struct {
	db *gorm.DB
}

// NewGormRepository creates a new GORM-based repository
func NewGormRepository(db *gorm.DB) *GormRepository {
	return &GormRepository{db: db}
}

// Create creates a new learning record
func (r *GormRepository) Create(ctx context.Context, record *models.ChatRecord) error {
	return r.db.WithContext(ctx).Create(record).Error
}

// FindByID finds a learning record by ID
func (r *GormRepository) FindByID(ctx context.Context, id string) (*models.ChatRecord, error) {
	var record models.ChatRecord
	if err := r.db.WithContext(ctx).First(&record, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &record, nil
}

// FindByUserWithFilters finds learning records for a user with filters and pagination
func (r *GormRepository) FindByUserWithFilters(ctx context.Context, userID string, filters ListFilters, page, pageSize int) (*ListResult, error) {
	query := r.db.WithContext(ctx).Model(&models.ChatRecord{}).Where("user_id = ?", userID)

	// Apply type filter
	if filters.Type != "" {
		query = query.Where("input_type = ?", filters.Type)
	}

	// Apply search filter (search in user_input and response_payload)
	if filters.Search != "" {
		searchPattern := "%" + filters.Search + "%"
		query = query.Where("user_input LIKE ? OR response_payload LIKE ?", searchPattern, searchPattern)
	}

	// Apply date filters
	if filters.DateFrom != nil {
		query = query.Where("created_time >= ?", filters.DateFrom)
	}
	if filters.DateTo != nil {
		query = query.Where("created_time <= ?", filters.DateTo)
	}

	// Count total matching records
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}

	// Apply pagination
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}
	offset := (page - 1) * pageSize

	var records []models.ChatRecord
	if err := query.Order("created_time DESC").Offset(offset).Limit(pageSize).Find(&records).Error; err != nil {
		return nil, err
	}

	totalPages := int(total) / pageSize
	if int(total)%pageSize > 0 {
		totalPages++
	}

	return &ListResult{
		Records:    records,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}, nil
}

// SoftDelete soft-deletes a learning record
func (r *GormRepository) SoftDelete(ctx context.Context, id string) error {
	result := r.db.WithContext(ctx).Delete(&models.ChatRecord{}, "id = ?", id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// CountByType counts learning records by input type for a user
func (r *GormRepository) CountByType(ctx context.Context, userID string) (map[string]int64, error) {
	type countResult struct {
		InputType string
		Count     int64
	}

	var results []countResult
	err := r.db.WithContext(ctx).
		Model(&models.ChatRecord{}).
		Select("input_type, COUNT(*) as count").
		Where("user_id = ?", userID).
		Group("input_type").
		Scan(&results).Error

	if err != nil {
		return nil, err
	}

	counts := make(map[string]int64)
	for _, r := range results {
		counts[r.InputType] = r.Count
	}

	return counts, nil
}

// CountStreak counts the consecutive days with at least one record
func (r *GormRepository) CountStreak(ctx context.Context, userID string) (int, error) {
	// Get unique dates (in descending order) when records were created
	type dateResult struct {
		RecordDate string
	}

	var results []dateResult
	err := r.db.WithContext(ctx).
		Model(&models.ChatRecord{}).
		Select("DATE(created_time) as record_date").
		Where("user_id = ?", userID).
		Group("DATE(created_time)").
		Order("record_date DESC").
		Scan(&results).Error

	if err != nil {
		return 0, err
	}

	if len(results) == 0 {
		return 0, nil
	}

	// Count consecutive days starting from today (or most recent)
	streak := 0
	today := time.Now().Truncate(24 * time.Hour)

	for i, result := range results {
		recordDate, err := time.Parse("2006-01-02", result.RecordDate)
		if err != nil {
			continue
		}
		recordDate = recordDate.Truncate(24 * time.Hour)

		expectedDate := today.AddDate(0, 0, -i)

		// Allow for the first record to be today or yesterday
		if i == 0 {
			daysDiff := int(today.Sub(recordDate).Hours() / 24)
			if daysDiff > 1 {
				// Most recent record is more than 1 day ago, streak is 0
				return 0, nil
			}
			if daysDiff == 1 {
				// Adjust expected date if most recent is yesterday
				expectedDate = recordDate
				today = recordDate
			}
		}

		if recordDate.Equal(expectedDate) || (i == 0 && recordDate.Equal(today)) {
			streak++
		} else {
			break
		}
	}

	return streak, nil
}

// GetLastRecordTime gets the time of the most recent record for a user
func (r *GormRepository) GetLastRecordTime(ctx context.Context, userID string) (*time.Time, error) {
	var record models.ChatRecord
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_time DESC").
		First(&record).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}

	return &record.CreatedTime, nil
}

// FindSimilar finds learning records with similar user_input using fuzzy matching
func (r *GormRepository) FindSimilar(ctx context.Context, userID string, input string, limit int) ([]models.ChatRecord, error) {
	if limit < 1 {
		limit = 3
	}

	var records []models.ChatRecord

	// For SQLite, use LIKE for basic fuzzy matching
	// First try exact match, then partial match
	searchPattern := "%" + input + "%"

	err := r.db.WithContext(ctx).
		Where("user_id = ? AND user_input LIKE ?", userID, searchPattern).
		Order("created_time DESC").
		Limit(limit).
		Find(&records).Error

	if err != nil {
		return nil, fmt.Errorf("find similar records: %w", err)
	}

	return records, nil
}

// Ensure GormRepository implements Repository interface
var _ Repository = (*GormRepository)(nil)
