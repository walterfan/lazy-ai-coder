package chatrecord

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/walterfan/lazy-ai-coder/pkg/models"
)

func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	err = db.AutoMigrate(&models.ChatRecord{})
	require.NoError(t, err)

	return db
}

func createTestRecord(id, userID, inputType, userInput string) *models.ChatRecord {
	return &models.ChatRecord{
		ID:              id,
		InputType:       inputType,
		UserInput:       userInput,
		ResponsePayload: `{"explanation":"test"}`,
		UserID:          userID,
		RealmID:         "realm-1",
	}
}

func TestGormRepository_Create(t *testing.T) {
	db := setupTestDB(t)
	repo := NewGormRepository(db)
	ctx := context.Background()

	record := createTestRecord("rec-1", "user-1", models.InputTypeWord, "test")

	err := repo.Create(ctx, record)
	require.NoError(t, err)

	// Verify record was created
	var found models.ChatRecord
	err = db.First(&found, "id = ?", "rec-1").Error
	require.NoError(t, err)
	assert.Equal(t, "test", found.UserInput)
}

func TestGormRepository_FindByID(t *testing.T) {
	db := setupTestDB(t)
	repo := NewGormRepository(db)
	ctx := context.Background()

	// Create a record
	record := createTestRecord("rec-1", "user-1", models.InputTypeWord, "serendipity")
	db.Create(record)

	// Find by ID
	found, err := repo.FindByID(ctx, "rec-1")
	require.NoError(t, err)
	assert.Equal(t, "serendipity", found.UserInput)

	// Not found
	_, err = repo.FindByID(ctx, "non-existent")
	assert.Error(t, err)
}

func TestGormRepository_FindByUserWithFilters_Pagination(t *testing.T) {
	db := setupTestDB(t)
	repo := NewGormRepository(db)
	ctx := context.Background()

	// Create 15 records for user-1
	for i := 1; i <= 15; i++ {
		record := createTestRecord(
			"rec-"+string(rune('a'+i-1)),
			"user-1",
			models.InputTypeWord,
			"word"+string(rune('a'+i-1)),
		)
		db.Create(record)
	}

	// First page
	result, err := repo.FindByUserWithFilters(ctx, "user-1", ListFilters{}, 1, 10)
	require.NoError(t, err)
	assert.Equal(t, int64(15), result.Total)
	assert.Len(t, result.Records, 10)
	assert.Equal(t, 1, result.Page)
	assert.Equal(t, 10, result.PageSize)
	assert.Equal(t, 2, result.TotalPages)

	// Second page
	result, err = repo.FindByUserWithFilters(ctx, "user-1", ListFilters{}, 2, 10)
	require.NoError(t, err)
	assert.Len(t, result.Records, 5)
	assert.Equal(t, 2, result.Page)
}

func TestGormRepository_FindByUserWithFilters_TypeFilter(t *testing.T) {
	db := setupTestDB(t)
	repo := NewGormRepository(db)
	ctx := context.Background()

	// Create records of different types
	db.Create(createTestRecord("rec-1", "user-1", models.InputTypeWord, "word1"))
	db.Create(createTestRecord("rec-2", "user-1", models.InputTypeWord, "word2"))
	db.Create(createTestRecord("rec-3", "user-1", models.InputTypeSentence, "sentence1"))
	db.Create(createTestRecord("rec-4", "user-1", models.InputTypeQuestion, "question1"))

	// Filter by word
	result, err := repo.FindByUserWithFilters(ctx, "user-1", ListFilters{Type: models.InputTypeWord}, 1, 10)
	require.NoError(t, err)
	assert.Equal(t, int64(2), result.Total)
	for _, r := range result.Records {
		assert.Equal(t, models.InputTypeWord, r.InputType)
	}

	// Filter by sentence
	result, err = repo.FindByUserWithFilters(ctx, "user-1", ListFilters{Type: models.InputTypeSentence}, 1, 10)
	require.NoError(t, err)
	assert.Equal(t, int64(1), result.Total)
}

func TestGormRepository_FindByUserWithFilters_Search(t *testing.T) {
	db := setupTestDB(t)
	repo := NewGormRepository(db)
	ctx := context.Background()

	// Create records
	db.Create(createTestRecord("rec-1", "user-1", models.InputTypeWord, "serendipity"))
	db.Create(createTestRecord("rec-2", "user-1", models.InputTypeWord, "ephemeral"))
	db.Create(createTestRecord("rec-3", "user-1", models.InputTypeQuestion, "What is OAuth?"))

	// Search for "serendipity"
	result, err := repo.FindByUserWithFilters(ctx, "user-1", ListFilters{Search: "serendipity"}, 1, 10)
	require.NoError(t, err)
	assert.Equal(t, int64(1), result.Total)
	assert.Equal(t, "serendipity", result.Records[0].UserInput)

	// Search for "OAuth"
	result, err = repo.FindByUserWithFilters(ctx, "user-1", ListFilters{Search: "OAuth"}, 1, 10)
	require.NoError(t, err)
	assert.Equal(t, int64(1), result.Total)
}

func TestGormRepository_FindByUserWithFilters_DateRange(t *testing.T) {
	db := setupTestDB(t)
	repo := NewGormRepository(db)
	ctx := context.Background()

	now := time.Now()
	yesterday := now.Add(-24 * time.Hour)
	lastWeek := now.Add(-7 * 24 * time.Hour)

	// Create records with specific times
	rec1 := createTestRecord("rec-1", "user-1", models.InputTypeWord, "today")
	rec1.CreatedTime = now
	db.Create(rec1)

	rec2 := createTestRecord("rec-2", "user-1", models.InputTypeWord, "yesterday")
	rec2.CreatedTime = yesterday
	db.Create(rec2)

	rec3 := createTestRecord("rec-3", "user-1", models.InputTypeWord, "lastweek")
	rec3.CreatedTime = lastWeek
	db.Create(rec3)

	// Filter from yesterday
	dateFrom := yesterday.Add(-time.Hour)
	result, err := repo.FindByUserWithFilters(ctx, "user-1", ListFilters{DateFrom: &dateFrom}, 1, 10)
	require.NoError(t, err)
	assert.Equal(t, int64(2), result.Total) // today and yesterday
}

func TestGormRepository_FindByUserWithFilters_UserIsolation(t *testing.T) {
	db := setupTestDB(t)
	repo := NewGormRepository(db)
	ctx := context.Background()

	// Create records for different users
	db.Create(createTestRecord("rec-1", "user-1", models.InputTypeWord, "user1-word"))
	db.Create(createTestRecord("rec-2", "user-2", models.InputTypeWord, "user2-word"))

	// User-1 should only see their own records
	result, err := repo.FindByUserWithFilters(ctx, "user-1", ListFilters{}, 1, 10)
	require.NoError(t, err)
	assert.Equal(t, int64(1), result.Total)
	assert.Equal(t, "user1-word", result.Records[0].UserInput)
}

func TestGormRepository_SoftDelete(t *testing.T) {
	db := setupTestDB(t)
	repo := NewGormRepository(db)
	ctx := context.Background()

	// Create a record
	record := createTestRecord("rec-1", "user-1", models.InputTypeWord, "to-delete")
	db.Create(record)

	// Soft delete
	err := repo.SoftDelete(ctx, "rec-1")
	require.NoError(t, err)

	// Should not find with normal query
	_, err = repo.FindByID(ctx, "rec-1")
	assert.Error(t, err)

	// Should find with unscoped
	var found models.ChatRecord
	err = db.Unscoped().First(&found, "id = ?", "rec-1").Error
	require.NoError(t, err)
	assert.NotNil(t, found.DeletedAt)
}

func TestGormRepository_SoftDelete_NotFound(t *testing.T) {
	db := setupTestDB(t)
	repo := NewGormRepository(db)
	ctx := context.Background()

	err := repo.SoftDelete(ctx, "non-existent")
	assert.Error(t, err)
}

func TestGormRepository_CountByType(t *testing.T) {
	db := setupTestDB(t)
	repo := NewGormRepository(db)
	ctx := context.Background()

	// Create records of different types
	db.Create(createTestRecord("rec-1", "user-1", models.InputTypeWord, "word1"))
	db.Create(createTestRecord("rec-2", "user-1", models.InputTypeWord, "word2"))
	db.Create(createTestRecord("rec-3", "user-1", models.InputTypeWord, "word3"))
	db.Create(createTestRecord("rec-4", "user-1", models.InputTypeSentence, "sentence1"))
	db.Create(createTestRecord("rec-5", "user-1", models.InputTypeSentence, "sentence2"))
	db.Create(createTestRecord("rec-6", "user-1", models.InputTypeQuestion, "question1"))
	db.Create(createTestRecord("rec-7", "user-1", models.InputTypeIdea, "idea1"))

	// Different user's records shouldn't count
	db.Create(createTestRecord("rec-8", "user-2", models.InputTypeWord, "other-user"))

	counts, err := repo.CountByType(ctx, "user-1")
	require.NoError(t, err)

	assert.Equal(t, int64(3), counts[models.InputTypeWord])
	assert.Equal(t, int64(2), counts[models.InputTypeSentence])
	assert.Equal(t, int64(1), counts[models.InputTypeQuestion])
	assert.Equal(t, int64(1), counts[models.InputTypeIdea])
}

func TestGormRepository_GetLastRecordTime(t *testing.T) {
	db := setupTestDB(t)
	repo := NewGormRepository(db)
	ctx := context.Background()

	// No records
	lastTime, err := repo.GetLastRecordTime(ctx, "user-1")
	require.NoError(t, err)
	assert.Nil(t, lastTime)

	// Create records
	now := time.Now()
	rec1 := createTestRecord("rec-1", "user-1", models.InputTypeWord, "old")
	rec1.CreatedTime = now.Add(-24 * time.Hour)
	db.Create(rec1)

	rec2 := createTestRecord("rec-2", "user-1", models.InputTypeWord, "new")
	rec2.CreatedTime = now
	db.Create(rec2)

	lastTime, err = repo.GetLastRecordTime(ctx, "user-1")
	require.NoError(t, err)
	require.NotNil(t, lastTime)
	// Check that it's approximately equal to now (within 1 second)
	assert.True(t, lastTime.Sub(now) < time.Second)
}

func TestGormRepository_FindSimilar(t *testing.T) {
	db := setupTestDB(t)
	repo := NewGormRepository(db)
	ctx := context.Background()

	// Create records
	db.Create(createTestRecord("rec-1", "user-1", models.InputTypeWord, "serendipity"))
	db.Create(createTestRecord("rec-2", "user-1", models.InputTypeWord, "ephemeral"))
	db.Create(createTestRecord("rec-3", "user-1", models.InputTypeWord, "serendipitous"))
	db.Create(createTestRecord("rec-4", "user-2", models.InputTypeWord, "serendipity")) // different user

	// Find similar to "serendipity"
	records, err := repo.FindSimilar(ctx, "user-1", "serendip", 10)
	require.NoError(t, err)
	assert.Len(t, records, 2) // serendipity and serendipitous

	// Respect limit
	records, err = repo.FindSimilar(ctx, "user-1", "serendip", 1)
	require.NoError(t, err)
	assert.Len(t, records, 1)
}

func TestGormRepository_CountStreak_NoRecords(t *testing.T) {
	db := setupTestDB(t)
	repo := NewGormRepository(db)
	ctx := context.Background()

	streak, err := repo.CountStreak(ctx, "user-1")
	require.NoError(t, err)
	assert.Equal(t, 0, streak)
}

func TestGormRepository_CountStreak_ConsecutiveDays(t *testing.T) {
	db := setupTestDB(t)
	repo := NewGormRepository(db)
	ctx := context.Background()

	now := time.Now()

	// Create records for consecutive days (today, yesterday, day before)
	for i := 0; i < 3; i++ {
		rec := createTestRecord("rec-"+string(rune('a'+i)), "user-1", models.InputTypeWord, "word")
		rec.CreatedTime = now.AddDate(0, 0, -i)
		db.Create(rec)
	}

	streak, err := repo.CountStreak(ctx, "user-1")
	require.NoError(t, err)
	assert.Equal(t, 3, streak)
}
