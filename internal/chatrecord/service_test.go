package chatrecord

import (
	"context"
	"testing"
	"time"

	"github.com/cloudwego/eino/schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/walterfan/lazy-ai-coder/internal/chatrecord/memory"
	"github.com/walterfan/lazy-ai-coder/pkg/models"
)

func setupServiceTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	err = db.AutoMigrate(&models.ChatRecord{})
	require.NoError(t, err)

	return db
}

func TestService_CreateRecord(t *testing.T) {
	db := setupServiceTestDB(t)
	repo := NewGormRepository(db)
	service := NewService(repo)
	ctx := context.Background()

	req := &ConfirmRequest{
		UserInput: "serendipity",
		InputType: models.InputTypeWord,
		ResponsePayload: &models.ResponsePayloadData{
			Explanation:   "意外发现美好事物的运气",
			Pronunciation: "/ˌserənˈdɪpəti/",
			Example:       "Finding that book was pure serendipity.",
		},
	}

	record, err := service.CreateRecord(ctx, req, "user-1", "realm-1")
	require.NoError(t, err)
	assert.NotEmpty(t, record.ID)
	assert.Equal(t, "serendipity", record.UserInput)
	assert.Equal(t, models.InputTypeWord, record.InputType)
	assert.Equal(t, "user-1", record.UserID)
	assert.Equal(t, "realm-1", record.RealmID)

	// Verify it was saved
	found, err := repo.FindByID(ctx, record.ID)
	require.NoError(t, err)
	assert.Equal(t, "serendipity", found.UserInput)
}

func TestService_CreateRecord_InvalidInputType(t *testing.T) {
	db := setupServiceTestDB(t)
	repo := NewGormRepository(db)
	service := NewService(repo)
	ctx := context.Background()

	req := &ConfirmRequest{
		UserInput: "test",
		InputType: "invalid-type",
		ResponsePayload: &models.ResponsePayloadData{
			Explanation: "test",
		},
	}

	_, err := service.CreateRecord(ctx, req, "user-1", "realm-1")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid input type")
}

func TestService_CreateRecord_EmptyUserInput(t *testing.T) {
	db := setupServiceTestDB(t)
	repo := NewGormRepository(db)
	service := NewService(repo)
	ctx := context.Background()

	req := &ConfirmRequest{
		UserInput: "",
		InputType: models.InputTypeWord,
		ResponsePayload: &models.ResponsePayloadData{
			Explanation: "test",
		},
	}

	_, err := service.CreateRecord(ctx, req, "user-1", "realm-1")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "user_input is required")
}

func TestService_ListRecords(t *testing.T) {
	db := setupServiceTestDB(t)
	repo := NewGormRepository(db)
	service := NewService(repo)
	ctx := context.Background()

	// Create some records
	for i := 0; i < 5; i++ {
		req := &ConfirmRequest{
			UserInput: "word" + string(rune('a'+i)),
			InputType: models.InputTypeWord,
			ResponsePayload: &models.ResponsePayloadData{
				Explanation: "test",
			},
		}
		_, err := service.CreateRecord(ctx, req, "user-1", "realm-1")
		require.NoError(t, err)
	}

	result, err := service.ListRecords(ctx, "user-1", ListFilters{}, 1, 10)
	require.NoError(t, err)
	assert.Equal(t, int64(5), result.Total)
	assert.Len(t, result.Records, 5)
}

func TestService_ListRecords_InvalidType(t *testing.T) {
	db := setupServiceTestDB(t)
	repo := NewGormRepository(db)
	service := NewService(repo)
	ctx := context.Background()

	_, err := service.ListRecords(ctx, "user-1", ListFilters{Type: "invalid"}, 1, 10)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid filter type")
}

func TestService_GetRecord(t *testing.T) {
	db := setupServiceTestDB(t)
	repo := NewGormRepository(db)
	service := NewService(repo)
	ctx := context.Background()

	// Create a record
	req := &ConfirmRequest{
		UserInput: "serendipity",
		InputType: models.InputTypeWord,
		ResponsePayload: &models.ResponsePayloadData{
			Explanation: "test",
		},
	}
	created, err := service.CreateRecord(ctx, req, "user-1", "realm-1")
	require.NoError(t, err)

	// Get by ID (same user)
	found, err := service.GetRecord(ctx, created.ID, "user-1")
	require.NoError(t, err)
	assert.Equal(t, "serendipity", found.UserInput)

	// Get by ID (different user - should fail)
	_, err = service.GetRecord(ctx, created.ID, "user-2")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "access denied")
}

func TestService_DeleteRecord(t *testing.T) {
	db := setupServiceTestDB(t)
	repo := NewGormRepository(db)
	service := NewService(repo)
	ctx := context.Background()

	// Create a record
	req := &ConfirmRequest{
		UserInput: "to-delete",
		InputType: models.InputTypeWord,
		ResponsePayload: &models.ResponsePayloadData{
			Explanation: "test",
		},
	}
	created, err := service.CreateRecord(ctx, req, "user-1", "realm-1")
	require.NoError(t, err)

	// Delete (same user)
	err = service.DeleteRecord(ctx, created.ID, "user-1")
	require.NoError(t, err)

	// Should not find after delete
	_, err = service.GetRecord(ctx, created.ID, "user-1")
	assert.Error(t, err)
}

func TestService_DeleteRecord_WrongUser(t *testing.T) {
	db := setupServiceTestDB(t)
	repo := NewGormRepository(db)
	service := NewService(repo)
	ctx := context.Background()

	// Create a record
	req := &ConfirmRequest{
		UserInput: "to-delete",
		InputType: models.InputTypeWord,
		ResponsePayload: &models.ResponsePayloadData{
			Explanation: "test",
		},
	}
	created, err := service.CreateRecord(ctx, req, "user-1", "realm-1")
	require.NoError(t, err)

	// Try to delete as different user
	err = service.DeleteRecord(ctx, created.ID, "user-2")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "access denied")
}

func TestService_GetStats(t *testing.T) {
	db := setupServiceTestDB(t)
	repo := NewGormRepository(db)
	service := NewService(repo)
	ctx := context.Background()

	// Create records of different types
	types := []string{
		models.InputTypeWord,
		models.InputTypeWord,
		models.InputTypeWord,
		models.InputTypeSentence,
		models.InputTypeSentence,
		models.InputTypeQuestion,
		models.InputTypeIdea,
	}

	for i, inputType := range types {
		req := &ConfirmRequest{
			UserInput: "input" + string(rune('a'+i)),
			InputType: inputType,
			ResponsePayload: &models.ResponsePayloadData{
				Explanation: "test",
			},
		}
		_, err := service.CreateRecord(ctx, req, "user-1", "realm-1")
		require.NoError(t, err)
	}

	stats, err := service.GetStats(ctx, "user-1")
	require.NoError(t, err)

	assert.Equal(t, int64(7), stats.Total)
	assert.Equal(t, int64(3), stats.ByType[models.InputTypeWord])
	assert.Equal(t, int64(2), stats.ByType[models.InputTypeSentence])
	assert.Equal(t, int64(1), stats.ByType[models.InputTypeQuestion])
	assert.Equal(t, int64(1), stats.ByType[models.InputTypeIdea])
	assert.NotNil(t, stats.LastRecordAt)
}

func TestService_GetStats_Empty(t *testing.T) {
	db := setupServiceTestDB(t)
	repo := NewGormRepository(db)
	service := NewService(repo)
	ctx := context.Background()

	stats, err := service.GetStats(ctx, "user-1")
	require.NoError(t, err)

	assert.Equal(t, int64(0), stats.Total)
	assert.Empty(t, stats.ByType)
	assert.Equal(t, 0, stats.Streak)
	assert.Nil(t, stats.LastRecordAt)
}

func TestService_FindSimilar(t *testing.T) {
	db := setupServiceTestDB(t)
	repo := NewGormRepository(db)
	service := NewService(repo)
	ctx := context.Background()

	// Create records
	words := []string{"serendipity", "serendipitous", "ephemeral", "ephemeron"}
	for _, word := range words {
		req := &ConfirmRequest{
			UserInput: word,
			InputType: models.InputTypeWord,
			ResponsePayload: &models.ResponsePayloadData{
				Explanation: "test",
			},
		}
		_, err := service.CreateRecord(ctx, req, "user-1", "realm-1")
		require.NoError(t, err)
	}

	// Find similar
	similar, err := service.FindSimilar(ctx, "user-1", "serendip", 10)
	require.NoError(t, err)
	assert.Len(t, similar, 2) // serendipity and serendipitous
}

func TestService_CreateRecord_AllInputTypes(t *testing.T) {
	db := setupServiceTestDB(t)
	repo := NewGormRepository(db)
	service := NewService(repo)
	ctx := context.Background()

	testCases := []struct {
		inputType string
		payload   *models.ResponsePayloadData
	}{
		{
			inputType: models.InputTypeWord,
			payload: &models.ResponsePayloadData{
				Explanation:   "Word explanation",
				Pronunciation: "/test/",
				Example:       "Example",
			},
		},
		{
			inputType: models.InputTypeSentence,
			payload: &models.ResponsePayloadData{
				Explanation: "Sentence explanation",
				Example:     "Example",
			},
		},
		{
			inputType: models.InputTypeQuestion,
			payload: &models.ResponsePayloadData{
				Answer: "This is the answer",
			},
		},
		{
			inputType: models.InputTypeIdea,
			payload: &models.ResponsePayloadData{
				Plan: []string{"Step 1", "Step 2", "Step 3"},
			},
		},
	}

	for _, tc := range testCases {
		req := &ConfirmRequest{
			UserInput:       "Test input for " + tc.inputType,
			InputType:       tc.inputType,
			ResponsePayload: tc.payload,
		}

		record, err := service.CreateRecord(ctx, req, "user-1", "realm-1")
		require.NoError(t, err, "Failed for input type: %s", tc.inputType)
		assert.Equal(t, tc.inputType, record.InputType)

		// Verify response payload can be retrieved
		payload, err := record.GetResponsePayload()
		require.NoError(t, err)

		switch tc.inputType {
		case models.InputTypeWord:
			assert.Equal(t, "Word explanation", payload.Explanation)
			assert.Equal(t, "/test/", payload.Pronunciation)
		case models.InputTypeSentence:
			assert.Equal(t, "Sentence explanation", payload.Explanation)
		case models.InputTypeQuestion:
			assert.Equal(t, "This is the answer", payload.Answer)
		case models.InputTypeIdea:
			assert.Len(t, payload.Plan, 3)
		}
	}
}

func TestService_ListRecords_WithFilters(t *testing.T) {
	db := setupServiceTestDB(t)
	repo := NewGormRepository(db)
	service := NewService(repo)
	ctx := context.Background()

	now := time.Now()

	// Create records with different types and times
	records := []struct {
		userInput string
		inputType string
		createdAt time.Time
	}{
		{"serendipity", models.InputTypeWord, now},
		{"ephemeral", models.InputTypeWord, now.Add(-24 * time.Hour)},
		{"How does OAuth work?", models.InputTypeQuestion, now.Add(-48 * time.Hour)},
	}

	for _, r := range records {
		req := &ConfirmRequest{
			UserInput: r.userInput,
			InputType: r.inputType,
			ResponsePayload: &models.ResponsePayloadData{
				Explanation: "test",
			},
		}
		created, err := service.CreateRecord(ctx, req, "user-1", "realm-1")
		require.NoError(t, err)

		// Update created_time manually for testing
		db.Model(&models.ChatRecord{}).Where("id = ?", created.ID).Update("created_time", r.createdAt)
	}

	// Test type filter
	result, err := service.ListRecords(ctx, "user-1", ListFilters{Type: models.InputTypeWord}, 1, 10)
	require.NoError(t, err)
	assert.Equal(t, int64(2), result.Total)

	// Test search filter
	result, err = service.ListRecords(ctx, "user-1", ListFilters{Search: "OAuth"}, 1, 10)
	require.NoError(t, err)
	assert.Equal(t, int64(1), result.Total)

	// Test date filter
	yesterday := now.Add(-36 * time.Hour)
	result, err = service.ListRecords(ctx, "user-1", ListFilters{DateFrom: &yesterday}, 1, 10)
	require.NoError(t, err)
	assert.Equal(t, int64(2), result.Total) // serendipity and ephemeral
}

// ==================== Session Memory Tests ====================

func TestService_SubmitInput_WithSessionMemory(t *testing.T) {
	db := setupServiceTestDB(t)
	repo := NewGormRepository(db)

	// Create a mock agent
	mockAgent := &MockAgent{
		classifyAs: models.InputTypeWord,
		payload: &models.ResponsePayloadData{
			Explanation:   "Test explanation",
			Pronunciation: "/test/",
			Example:       "Test example",
		},
	}

	// Create session store
	sessionStore := memory.NewInMemorySessionStore(10, 30*time.Minute)

	// Create service with memory
	service := NewServiceWithMemory(repo, mockAgent, sessionStore)

	ctx := context.Background()

	// First submit
	req := &SubmitRequest{
		UserInput: "serendipity",
		SessionID: "test-session",
	}

	result, err := service.SubmitInput(ctx, req, "user-1")
	require.NoError(t, err)
	assert.Equal(t, "test-session", result.SessionID)
	assert.Equal(t, models.InputTypeWord, result.InputType)

	// Verify session has context
	messages := service.GetSessionContext("test-session")
	assert.Len(t, messages, 2) // user message + assistant response
	assert.Equal(t, "serendipity", messages[0].Content)
	assert.Contains(t, messages[1].Content, "Test explanation")
}

func TestService_SubmitInput_GeneratesSessionID(t *testing.T) {
	db := setupServiceTestDB(t)
	repo := NewGormRepository(db)

	mockAgent := &MockAgent{
		classifyAs: models.InputTypeQuestion,
		payload: &models.ResponsePayloadData{
			Answer: "This is the answer",
		},
	}

	sessionStore := memory.NewInMemorySessionStore(10, 30*time.Minute)
	service := NewServiceWithMemory(repo, mockAgent, sessionStore)

	ctx := context.Background()

	// Submit without session ID
	req := &SubmitRequest{
		UserInput: "How does OAuth2 work?",
		// SessionID not provided
	}

	result, err := service.SubmitInput(ctx, req, "user-1")
	require.NoError(t, err)
	assert.NotEmpty(t, result.SessionID) // Should be generated
}

func TestService_SubmitInput_SessionContext_MultipleInputs(t *testing.T) {
	db := setupServiceTestDB(t)
	repo := NewGormRepository(db)

	callCount := 0
	mockAgent := &MockAgent{
		classifyAs: models.InputTypeWord,
		payload: &models.ResponsePayloadData{
			Explanation: "Explanation",
		},
	}

	sessionStore := memory.NewInMemorySessionStore(10, 30*time.Minute)
	service := NewServiceWithMemory(repo, mockAgent, sessionStore)

	ctx := context.Background()
	sessionID := "multi-test-session"

	// Submit multiple inputs to the same session
	inputs := []string{"serendipity", "ephemeral", "ubiquitous"}
	for _, input := range inputs {
		req := &SubmitRequest{
			UserInput: input,
			SessionID: sessionID,
		}
		_, err := service.SubmitInput(ctx, req, "user-1")
		require.NoError(t, err)
		callCount++
	}

	// Verify session has all context (2 messages per input)
	messages := service.GetSessionContext(sessionID)
	assert.Len(t, messages, 6) // 3 inputs * 2 messages each
}

func TestService_ClearSession(t *testing.T) {
	db := setupServiceTestDB(t)
	repo := NewGormRepository(db)

	mockAgent := &MockAgent{
		classifyAs: models.InputTypeWord,
		payload: &models.ResponsePayloadData{
			Explanation: "Test",
		},
	}

	sessionStore := memory.NewInMemorySessionStore(10, 30*time.Minute)
	service := NewServiceWithMemory(repo, mockAgent, sessionStore)

	ctx := context.Background()
	sessionID := "clear-test-session"

	// Add some messages
	req := &SubmitRequest{
		UserInput: "test",
		SessionID: sessionID,
	}
	_, err := service.SubmitInput(ctx, req, "user-1")
	require.NoError(t, err)

	// Verify context exists
	messages := service.GetSessionContext(sessionID)
	assert.Len(t, messages, 2)

	// Clear session
	err = service.ClearSession(sessionID)
	require.NoError(t, err)

	// Verify context is cleared
	messages = service.GetSessionContext(sessionID)
	assert.Nil(t, messages)
}

func TestService_GetSessionContext_NilStore(t *testing.T) {
	db := setupServiceTestDB(t)
	repo := NewGormRepository(db)
	service := NewService(repo) // No session store

	// Should return nil without error
	messages := service.GetSessionContext("any-session")
	assert.Nil(t, messages)
}

func TestService_ClearSession_NilStore(t *testing.T) {
	db := setupServiceTestDB(t)
	repo := NewGormRepository(db)
	service := NewService(repo) // No session store

	// Should return nil without error
	err := service.ClearSession("any-session")
	assert.Nil(t, err)
}

func TestService_SetSessionStore(t *testing.T) {
	db := setupServiceTestDB(t)
	repo := NewGormRepository(db)
	service := NewService(repo)

	// Initially no session store
	assert.Nil(t, service.GetSessionContext("test"))

	// Set session store
	sessionStore := memory.NewInMemorySessionStore(10, 30*time.Minute)
	service.SetSessionStore(sessionStore)

	// Now it should work (return empty but not nil store)
	messages := service.GetSessionContext("test")
	assert.Nil(t, messages) // No messages yet, but store is set
}

func TestFormatResponseSummary(t *testing.T) {
	testCases := []struct {
		name      string
		inputType string
		payload   *models.ResponsePayloadData
		expected  string
	}{
		{
			name:      "Word",
			inputType: models.InputTypeWord,
			payload:   &models.ResponsePayloadData{Explanation: "Word explanation"},
			expected:  "Word explanation",
		},
		{
			name:      "Question",
			inputType: models.InputTypeQuestion,
			payload:   &models.ResponsePayloadData{Answer: "This is the answer"},
			expected:  "This is the answer",
		},
		{
			name:      "Idea",
			inputType: models.InputTypeIdea,
			payload:   &models.ResponsePayloadData{Plan: []string{"Step 1", "Step 2"}},
			expected:  "Plan: Step 1",
		},
		{
			name:      "Topic",
			inputType: models.InputTypeTopic,
			payload:   &models.ResponsePayloadData{Introduction: "Topic introduction"},
			expected:  "Topic introduction",
		},
		{
			name:      "Nil payload",
			inputType: models.InputTypeWord,
			payload:   nil,
			expected:  "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := formatResponseSummary(tc.inputType, tc.payload)
			assert.Contains(t, result, tc.expected)
		})
	}
}

// MockAgent for testing
type MockAgent struct {
	classifyAs string
	payload    *models.ResponsePayloadData
}

func (m *MockAgent) Process(ctx context.Context, input string) (*ProcessResult, error) {
	return m.ProcessWithHistory(ctx, input, nil, "")
}

func (m *MockAgent) ProcessWithHistory(ctx context.Context, _ string, _ []*schema.Message, _ string) (*ProcessResult, error) {
	return &ProcessResult{
		InputType:       m.classifyAs,
		ResponsePayload: m.payload,
	}, nil
}
