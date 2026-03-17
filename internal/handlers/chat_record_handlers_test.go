package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/cloudwego/eino/schema"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/walterfan/lazy-ai-coder/internal/chatrecord"
	"github.com/walterfan/lazy-ai-coder/pkg/models"
)

func setupchatrecordTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	err = db.AutoMigrate(&models.ChatRecord{}, &models.UserRole{})
	require.NoError(t, err)

	return db
}

func setupchatrecordTestRouter(db *gorm.DB) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	// Add middleware to set user context for testing
	r.Use(func(c *gin.Context) {
		c.Set("auth_type", "jwt")
		c.Set("user_id", "test-user-id")
		c.Set("realm_id", "test-realm-id")
		c.Set("username", "testuser")
		c.Next()
	})

	handlers := NewChatRecordHandlers(db)

	// Register routes
	r.POST("/api/v1/chat-record/submit", handlers.HandleSubmit)
	r.POST("/api/v1/chat-record/confirm", handlers.HandleConfirm)
	r.GET("/api/v1/chat-record/list", handlers.HandleList)
	r.GET("/api/v1/chat-record/stats", handlers.HandleStats)
	r.GET("/api/v1/chat-record/:id", handlers.HandleGet)
	r.DELETE("/api/v1/chat-record/:id", handlers.HandleDelete)

	return r
}

func setupUnauthenticatedRouter(db *gorm.DB) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	// No user context set - simulates unauthenticated request
	r.Use(func(c *gin.Context) {
		c.Set("auth_type", "guest")
		// user_id is empty
		c.Next()
	})

	handlers := NewChatRecordHandlers(db)

	r.POST("/api/v1/chat-record/confirm", handlers.HandleConfirm)
	r.GET("/api/v1/chat-record/list", handlers.HandleList)

	return r
}

// AT-1: Confirm creates a learning record
func TestConfirm_CreatesRecord(t *testing.T) {
	db := setupchatrecordTestDB(t)
	router := setupchatrecordTestRouter(db)

	reqBody := ConfirmchatrecordRequest{
		UserInput: "serendipity",
		InputType: models.InputTypeWord,
		ResponsePayload: &models.ResponsePayloadData{
			Explanation:   "意外发现美好事物的运气",
			Pronunciation: "/ˌserənˈdɪpəti/",
			Example:       "Finding that book was pure serendipity.",
		},
	}

	body, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "/api/v1/chat-record/confirm", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var response models.ChatRecord
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.NotEmpty(t, response.ID)
	assert.Equal(t, "serendipity", response.UserInput)
	assert.Equal(t, models.InputTypeWord, response.InputType)
	assert.Equal(t, "test-user-id", response.UserID)

	// Verify record exists in DB
	var count int64
	db.Model(&models.ChatRecord{}).Where("id = ?", response.ID).Count(&count)
	assert.Equal(t, int64(1), count)
}

// AT-2: Confirm without auth returns 401
func TestConfirm_Unauthenticated_Returns401(t *testing.T) {
	db := setupchatrecordTestDB(t)
	router := setupUnauthenticatedRouter(db)

	reqBody := ConfirmchatrecordRequest{
		UserInput: "test",
		InputType: models.InputTypeWord,
		ResponsePayload: &models.ResponsePayloadData{
			Explanation: "test",
		},
	}

	body, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "/api/v1/chat-record/confirm", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	// Verify no record was created
	var count int64
	db.Model(&models.ChatRecord{}).Count(&count)
	assert.Equal(t, int64(0), count)
}

// AT-3: List returns user's records with pagination
func TestList_ReturnsPaginatedRecords(t *testing.T) {
	db := setupchatrecordTestDB(t)
	router := setupchatrecordTestRouter(db)

	// Create 15 records
	for i := 0; i < 15; i++ {
		record := &models.ChatRecord{
			ID:              "rec-" + string(rune('a'+i)),
			InputType:       models.InputTypeWord,
			UserInput:       "word" + string(rune('a'+i)),
			ResponsePayload: `{"explanation":"test"}`,
			UserID:          "test-user-id",
			RealmID:         "test-realm-id",
		}
		db.Create(record)
	}

	// First page
	req, _ := http.NewRequest("GET", "/api/v1/chat-record/list?page=1&page_size=10", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response ListchatrecordsResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, int64(15), response.Total)
	assert.Len(t, response.Records, 10)
	assert.Equal(t, 1, response.Page)
	assert.Equal(t, 10, response.PageSize)
	assert.Equal(t, 2, response.TotalPages)

	// Second page
	req, _ = http.NewRequest("GET", "/api/v1/chat-record/list?page=2&page_size=10", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Len(t, response.Records, 5)
	assert.Equal(t, 2, response.Page)
}

// AT-4: List filters by type
func TestList_FiltersByType(t *testing.T) {
	db := setupchatrecordTestDB(t)
	router := setupchatrecordTestRouter(db)

	// Create records of different types
	records := []struct {
		id        string
		inputType string
	}{
		{"rec-1", models.InputTypeWord},
		{"rec-2", models.InputTypeWord},
		{"rec-3", models.InputTypeSentence},
		{"rec-4", models.InputTypeQuestion},
	}

	for _, r := range records {
		record := &models.ChatRecord{
			ID:              r.id,
			InputType:       r.inputType,
			UserInput:       "test",
			ResponsePayload: `{"explanation":"test"}`,
			UserID:          "test-user-id",
		}
		db.Create(record)
	}

	// Filter by word
	req, _ := http.NewRequest("GET", "/api/v1/chat-record/list?type=word", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response ListchatrecordsResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, int64(2), response.Total)
	for _, r := range response.Records {
		assert.Equal(t, models.InputTypeWord, r.InputType)
	}
}

// AT-5: Get single record by ID
func TestGet_ReturnsRecordDetail(t *testing.T) {
	db := setupchatrecordTestDB(t)
	router := setupchatrecordTestRouter(db)

	// Create a record
	record := &models.ChatRecord{
		ID:              "test-record-id",
		InputType:       models.InputTypeWord,
		UserInput:       "serendipity",
		ResponsePayload: `{"explanation":"test explanation","pronunciation":"/test/"}`,
		UserID:          "test-user-id",
	}
	db.Create(record)

	req, _ := http.NewRequest("GET", "/api/v1/chat-record/test-record-id", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "test-record-id", response["id"])
	assert.Equal(t, "serendipity", response["user_input"])
}

// AT-5b: Get returns 404 for non-existent record
func TestGet_NotFound(t *testing.T) {
	db := setupchatrecordTestDB(t)
	router := setupchatrecordTestRouter(db)

	req, _ := http.NewRequest("GET", "/api/v1/chat-record/non-existent", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

// AT-5c: Get returns 403 for record owned by another user
func TestGet_OtherUserRecord_Forbidden(t *testing.T) {
	db := setupchatrecordTestDB(t)
	router := setupchatrecordTestRouter(db)

	// Create a record owned by a different user
	record := &models.ChatRecord{
		ID:              "other-user-record",
		InputType:       models.InputTypeWord,
		UserInput:       "test",
		ResponsePayload: `{"explanation":"test"}`,
		UserID:          "other-user-id", // Different user
	}
	db.Create(record)

	req, _ := http.NewRequest("GET", "/api/v1/chat-record/other-user-record", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

// AT-6: Delete soft-deletes a record
func TestDelete_SoftDeletesRecord(t *testing.T) {
	db := setupchatrecordTestDB(t)
	router := setupchatrecordTestRouter(db)

	// Create a record
	record := &models.ChatRecord{
		ID:              "to-delete",
		InputType:       models.InputTypeWord,
		UserInput:       "test",
		ResponsePayload: `{"explanation":"test"}`,
		UserID:          "test-user-id",
	}
	db.Create(record)

	// Delete
	req, _ := http.NewRequest("DELETE", "/api/v1/chat-record/to-delete", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)

	// Verify soft-deleted (not in normal query)
	var count int64
	db.Model(&models.ChatRecord{}).Where("id = ?", "to-delete").Count(&count)
	assert.Equal(t, int64(0), count)

	// But exists with Unscoped
	db.Unscoped().Model(&models.ChatRecord{}).Where("id = ?", "to-delete").Count(&count)
	assert.Equal(t, int64(1), count)

	// Get returns 404 after delete
	req, _ = http.NewRequest("GET", "/api/v1/chat-record/to-delete", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

// AT-7: Stats returns counts by type
func TestStats_ReturnsCountsByType(t *testing.T) {
	db := setupchatrecordTestDB(t)
	router := setupchatrecordTestRouter(db)

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
		record := &models.ChatRecord{
			ID:              "rec-" + string(rune('a'+i)),
			InputType:       inputType,
			UserInput:       "test",
			ResponsePayload: `{"explanation":"test"}`,
			UserID:          "test-user-id",
		}
		db.Create(record)
	}

	req, _ := http.NewRequest("GET", "/api/v1/chat-record/stats", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response StatsResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, int64(7), response.Total)
	assert.Equal(t, int64(3), response.ByType[models.InputTypeWord])
	assert.Equal(t, int64(2), response.ByType[models.InputTypeSentence])
	assert.Equal(t, int64(1), response.ByType[models.InputTypeQuestion])
	assert.Equal(t, int64(1), response.ByType[models.InputTypeIdea])
}

// Test different input types
func TestConfirm_AllInputTypes(t *testing.T) {
	db := setupchatrecordTestDB(t)
	router := setupchatrecordTestRouter(db)

	testCases := []struct {
		name      string
		inputType string
		payload   *models.ResponsePayloadData
	}{
		{
			name:      "Word",
			inputType: models.InputTypeWord,
			payload: &models.ResponsePayloadData{
				Explanation:   "Word explanation",
				Pronunciation: "/test/",
				Example:       "Example usage",
			},
		},
		{
			name:      "Sentence",
			inputType: models.InputTypeSentence,
			payload: &models.ResponsePayloadData{
				Explanation: "Sentence explanation",
				Example:     "Example context",
			},
		},
		{
			name:      "Question",
			inputType: models.InputTypeQuestion,
			payload: &models.ResponsePayloadData{
				Answer: "This is the answer",
			},
		},
		{
			name:      "Idea",
			inputType: models.InputTypeIdea,
			payload: &models.ResponsePayloadData{
				Plan: []string{"Step 1", "Step 2", "Step 3"},
			},
		},
		{
			name:      "Topic",
			inputType: models.InputTypeTopic,
			payload: &models.ResponsePayloadData{
				Introduction: "Kubernetes is a container orchestration platform",
				KeyConcepts: []models.ConceptItem{
					{Name: "Pod", Description: "Smallest deployable unit"},
				},
				LearningPath: []models.LearningStep{
					{Order: 1, Title: "Learn Docker basics", Duration: "1 week"},
				},
				Resources: []models.ResourceItem{
					{Type: "documentation", Title: "K8s Docs", URL: "https://kubernetes.io"},
				},
				Prerequisites: []string{"Docker", "Linux"},
				TimeEstimate:  "4-6 weeks",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			reqBody := ConfirmchatrecordRequest{
				UserInput:       "Test input for " + tc.inputType,
				InputType:       tc.inputType,
				ResponsePayload: tc.payload,
			}

			body, _ := json.Marshal(reqBody)
			req, _ := http.NewRequest("POST", "/api/v1/chat-record/confirm", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusCreated, w.Code, "Failed for input type: %s", tc.inputType)

			var response models.ChatRecord
			err := json.Unmarshal(w.Body.Bytes(), &response)
			require.NoError(t, err)

			assert.Equal(t, tc.inputType, response.InputType)
		})
	}
}

// Test search filter
func TestList_SearchFilter(t *testing.T) {
	db := setupchatrecordTestDB(t)
	router := setupchatrecordTestRouter(db)

	// Create records
	records := []struct {
		id    string
		input string
	}{
		{"rec-1", "serendipity"},
		{"rec-2", "ephemeral"},
		{"rec-3", "What is OAuth2?"},
	}

	for _, r := range records {
		record := &models.ChatRecord{
			ID:              r.id,
			InputType:       models.InputTypeWord,
			UserInput:       r.input,
			ResponsePayload: `{"explanation":"test"}`,
			UserID:          "test-user-id",
		}
		db.Create(record)
	}

	// Search for "OAuth"
	req, _ := http.NewRequest("GET", "/api/v1/chat-record/list?search=OAuth", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response ListchatrecordsResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, int64(1), response.Total)
}

// Test user isolation - user can only see their own records
func TestList_UserIsolation(t *testing.T) {
	db := setupchatrecordTestDB(t)
	router := setupchatrecordTestRouter(db)

	// Create records for different users
	db.Create(&models.ChatRecord{
		ID:              "my-record",
		InputType:       models.InputTypeWord,
		UserInput:       "my word",
		ResponsePayload: `{"explanation":"test"}`,
		UserID:          "test-user-id",
	})
	db.Create(&models.ChatRecord{
		ID:              "other-record",
		InputType:       models.InputTypeWord,
		UserInput:       "other word",
		ResponsePayload: `{"explanation":"test"}`,
		UserID:          "other-user-id",
	})

	req, _ := http.NewRequest("GET", "/api/v1/chat-record/list", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response ListchatrecordsResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, int64(1), response.Total)
	assert.Equal(t, "my word", response.Records[0].UserInput)
}

// ==================== Submit Endpoint Tests ====================

// MockAgent for testing submit endpoint
type MockAgent struct {
	classifyAs string
	payload    *models.ResponsePayloadData
}

func (m *MockAgent) Process(ctx context.Context, input string) (*chatrecord.ProcessResult, error) {
	return m.ProcessWithHistory(ctx, input, nil)
}

func (m *MockAgent) ProcessWithHistory(ctx context.Context, _ string, _ []*schema.Message) (*chatrecord.ProcessResult, error) {
	return &chatrecord.ProcessResult{
		InputType:       m.classifyAs,
		ResponsePayload: m.payload,
	}, nil
}

func setupSubmitTestRouter(db *gorm.DB, agent chatrecord.Agent) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	// Add middleware to set user context for testing
	r.Use(func(c *gin.Context) {
		c.Set("auth_type", "jwt")
		c.Set("user_id", "test-user-id")
		c.Set("realm_id", "test-realm-id")
		c.Set("username", "testuser")
		c.Next()
	})

	handlers := NewChatRecordHandlersWithAgent(db, agent)

	// Register routes
	r.POST("/api/v1/chat-record/submit", handlers.HandleSubmit)
	r.POST("/api/v1/chat-record/confirm", handlers.HandleConfirm)
	r.GET("/api/v1/chat-record/list", handlers.HandleList)

	return r
}

// AT-8: Submit with word input returns inputType="word" and explanation
func TestSubmit_Word(t *testing.T) {
	db := setupchatrecordTestDB(t)
	mockAgent := &MockAgent{
		classifyAs: models.InputTypeWord,
		payload: &models.ResponsePayloadData{
			Explanation:   "意外发现美好事物的运气",
			Pronunciation: "/ˌserənˈdɪpəti/",
			Example:       "Finding that book was pure serendipity.",
		},
	}
	router := setupSubmitTestRouter(db, mockAgent)

	reqBody := SubmitChatRecordRequest{
		UserInput: "serendipity",
	}

	body, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "/api/v1/chat-record/submit", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response SubmitchatrecordResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, models.InputTypeWord, response.InputType)
	assert.NotEmpty(t, response.ResponsePayload.Explanation)
	assert.NotEmpty(t, response.ResponsePayload.Pronunciation)
	assert.NotEmpty(t, response.ResponsePayload.Example)
}

// AT-9: Submit with sentence input returns inputType="sentence"
func TestSubmit_Sentence(t *testing.T) {
	db := setupchatrecordTestDB(t)
	mockAgent := &MockAgent{
		classifyAs: models.InputTypeSentence,
		payload: &models.ResponsePayloadData{
			Explanation: "时间像箭一样飞逝，形容时间过得很快",
			Example:     "Time really flies when you're having fun.",
		},
	}
	router := setupSubmitTestRouter(db, mockAgent)

	reqBody := SubmitChatRecordRequest{
		UserInput: "Time flies like an arrow",
	}

	body, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "/api/v1/chat-record/submit", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response SubmitchatrecordResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, models.InputTypeSentence, response.InputType)
	assert.NotEmpty(t, response.ResponsePayload.Explanation)
}

// AT-10: Submit with question input returns inputType="question" and answer
func TestSubmit_Question(t *testing.T) {
	db := setupchatrecordTestDB(t)
	mockAgent := &MockAgent{
		classifyAs: models.InputTypeQuestion,
		payload: &models.ResponsePayloadData{
			Answer: "OAuth2 is an authorization framework that enables applications to obtain limited access to user accounts on an HTTP service.",
		},
	}
	router := setupSubmitTestRouter(db, mockAgent)

	reqBody := SubmitChatRecordRequest{
		UserInput: "How does OAuth2 work?",
	}

	body, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "/api/v1/chat-record/submit", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response SubmitchatrecordResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, models.InputTypeQuestion, response.InputType)
	assert.NotEmpty(t, response.ResponsePayload.Answer)
}

// AT-11: Submit with idea input returns inputType="idea" and plan
func TestSubmit_Idea(t *testing.T) {
	db := setupchatrecordTestDB(t)
	mockAgent := &MockAgent{
		classifyAs: models.InputTypeIdea,
		payload: &models.ResponsePayloadData{
			Plan: []string{
				"Step 1: Define the habit data model",
				"Step 2: Create the database schema",
				"Step 3: Build the API endpoints",
			},
		},
	}
	router := setupSubmitTestRouter(db, mockAgent)

	reqBody := SubmitChatRecordRequest{
		UserInput: "I want to build a habit tracker",
	}

	body, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "/api/v1/chat-record/submit", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response SubmitchatrecordResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, models.InputTypeIdea, response.InputType)
	assert.NotEmpty(t, response.ResponsePayload.Plan)
	assert.GreaterOrEqual(t, len(response.ResponsePayload.Plan), 1)
}

// Test submit with topic input returns comprehensive learning plan
func TestSubmit_Topic(t *testing.T) {
	db := setupchatrecordTestDB(t)
	mockAgent := &MockAgent{
		classifyAs: models.InputTypeTopic,
		payload: &models.ResponsePayloadData{
			Introduction: "Kubernetes is a container orchestration platform",
			KeyConcepts: []models.ConceptItem{
				{Name: "Pod", Description: "Smallest deployable unit"},
				{Name: "Service", Description: "Network abstraction"},
			},
			LearningPath: []models.LearningStep{
				{Order: 1, Title: "Learn Docker", Duration: "1 week"},
				{Order: 2, Title: "K8s Basics", Duration: "2 weeks"},
			},
			Resources: []models.ResourceItem{
				{Type: "documentation", Title: "K8s Docs", URL: "https://kubernetes.io/docs"},
			},
			Prerequisites: []string{"Docker", "Linux"},
			TimeEstimate:  "4-6 weeks",
		},
	}
	router := setupSubmitTestRouter(db, mockAgent)

	reqBody := SubmitChatRecordRequest{
		UserInput: "Kubernetes",
	}

	body, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "/api/v1/chat-record/submit", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response SubmitchatrecordResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, models.InputTypeTopic, response.InputType)
	assert.NotEmpty(t, response.ResponsePayload.Introduction)
	assert.NotEmpty(t, response.ResponsePayload.KeyConcepts)
	assert.NotEmpty(t, response.ResponsePayload.LearningPath)
	assert.NotEmpty(t, response.ResponsePayload.Resources)
	assert.NotEmpty(t, response.ResponsePayload.Prerequisites)
	assert.NotEmpty(t, response.ResponsePayload.TimeEstimate)
}

// AT-12: Submit does NOT write to database
func TestSubmit_DoesNotWriteToDB(t *testing.T) {
	db := setupchatrecordTestDB(t)
	mockAgent := &MockAgent{
		classifyAs: models.InputTypeWord,
		payload: &models.ResponsePayloadData{
			Explanation: "Test explanation",
		},
	}
	router := setupSubmitTestRouter(db, mockAgent)

	// Count records before
	var countBefore int64
	db.Model(&models.ChatRecord{}).Count(&countBefore)

	reqBody := SubmitChatRecordRequest{
		UserInput: "test",
	}

	body, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "/api/v1/chat-record/submit", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// Count records after - should be the same
	var countAfter int64
	db.Model(&models.ChatRecord{}).Count(&countAfter)

	assert.Equal(t, countBefore, countAfter, "Submit should NOT write to database")
}

// Test submit returns similar records
func TestSubmit_ReturnsSimilarRecords(t *testing.T) {
	db := setupchatrecordTestDB(t)

	// Create existing records
	db.Create(&models.ChatRecord{
		ID:              "existing-1",
		InputType:       models.InputTypeWord,
		UserInput:       "serendipity",
		ResponsePayload: `{"explanation":"意外发现"}`,
		UserID:          "test-user-id",
	})

	mockAgent := &MockAgent{
		classifyAs: models.InputTypeWord,
		payload: &models.ResponsePayloadData{
			Explanation: "Test",
		},
	}
	router := setupSubmitTestRouter(db, mockAgent)

	reqBody := SubmitChatRecordRequest{
		UserInput: "serendipity", // Same as existing record
	}

	body, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "/api/v1/chat-record/submit", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response SubmitchatrecordResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.NotEmpty(t, response.SimilarRecords, "Should return similar records")
}

// Test submit without agent configured returns error
func TestSubmit_NoAgent_ReturnsError(t *testing.T) {
	db := setupchatrecordTestDB(t)

	// Use handler without agent
	router := setupchatrecordTestRouter(db)

	reqBody := SubmitChatRecordRequest{
		UserInput: "test",
	}

	body, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "/api/v1/chat-record/submit", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Should return 503 (service unavailable) when no agent and no request LLM settings
	assert.Equal(t, http.StatusServiceUnavailable, w.Code)

	// Verify error message
	var response map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Contains(t, response["error"], "not configured")
}

// Test the full submit -> confirm flow
func TestSubmitThenConfirm_Flow(t *testing.T) {
	db := setupchatrecordTestDB(t)
	mockAgent := &MockAgent{
		classifyAs: models.InputTypeWord,
		payload: &models.ResponsePayloadData{
			Explanation:   "意外发现美好事物的运气",
			Pronunciation: "/ˌserənˈdɪpəti/",
			Example:       "Finding that book was pure serendipity.",
		},
	}
	router := setupSubmitTestRouter(db, mockAgent)

	// Step 1: Submit
	submitReq := SubmitChatRecordRequest{
		UserInput: "serendipity",
	}

	body, _ := json.Marshal(submitReq)
	req, _ := http.NewRequest("POST", "/api/v1/chat-record/submit", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var submitResponse SubmitchatrecordResponse
	err := json.Unmarshal(w.Body.Bytes(), &submitResponse)
	require.NoError(t, err)

	// Step 2: Confirm with the response from submit
	confirmReq := ConfirmchatrecordRequest{
		UserInput:       "serendipity",
		InputType:       submitResponse.InputType,
		ResponsePayload: submitResponse.ResponsePayload,
	}

	body, _ = json.Marshal(confirmReq)
	req, _ = http.NewRequest("POST", "/api/v1/chat-record/confirm", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var confirmResponse models.ChatRecord
	err = json.Unmarshal(w.Body.Bytes(), &confirmResponse)
	require.NoError(t, err)

	assert.Equal(t, "serendipity", confirmResponse.UserInput)
	assert.Equal(t, models.InputTypeWord, confirmResponse.InputType)

	// Step 3: Verify record is in database via list
	req, _ = http.NewRequest("GET", "/api/v1/chat-record/list", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var listResponse ListchatrecordsResponse
	err = json.Unmarshal(w.Body.Bytes(), &listResponse)
	require.NoError(t, err)

	assert.Equal(t, int64(1), listResponse.Total)
	assert.Equal(t, "serendipity", listResponse.Records[0].UserInput)
}
