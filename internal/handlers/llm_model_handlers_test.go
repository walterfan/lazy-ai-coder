package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/walterfan/lazy-ai-coder/pkg/models"
)

func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	// Auto-migrate all required tables
	err = db.AutoMigrate(&models.LLMModel{}, &models.UserRole{})
	require.NoError(t, err)

	return db
}

func setupTestRouter(db *gorm.DB) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	// Add a simple middleware to set user context for testing
	r.Use(func(c *gin.Context) {
		c.Set("auth_type", "guest")
		c.Set("user_id", "test-user-id")
		c.Set("realm_id", "test-realm-id")
		c.Set("username", "testuser")
		c.Next()
	})

	handlers := NewLLMModelHandlers(db)

	// Register routes
	r.GET("/api/v1/llm-models", handlers.ListLLMModels)
	r.GET("/api/v1/llm-models/default", handlers.GetDefaultLLMModel)
	r.GET("/api/v1/llm-models/:id", handlers.GetLLMModel)
	r.POST("/api/v1/llm-models", handlers.CreateLLMModel)
	r.PUT("/api/v1/llm-models/:id", handlers.UpdateLLMModel)
	r.DELETE("/api/v1/llm-models/:id", handlers.DeleteLLMModel)
	r.POST("/api/v1/llm-models/:id/default", handlers.SetDefaultLLMModel)
	r.POST("/api/v1/llm-models/:id/toggle", handlers.ToggleLLMModelEnabled)

	return r
}

func TestListLLMModels_Empty(t *testing.T) {
	db := setupTestDB(t)
	router := setupTestRouter(db)

	req, _ := http.NewRequest("GET", "/api/v1/llm-models", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, float64(0), response["total"])
	assert.NotNil(t, response["data"])
}

func TestCreateLLMModel(t *testing.T) {
	db := setupTestDB(t)
	router := setupTestRouter(db)

	model := CreateLLMModelRequest{
		Name:        "Test GPT-4",
		LLMType:     "openai",
		BaseURL:     "https://api.openai.com/v1",
		Model:       "gpt-4",
		Temperature: 0.7,
		MaxTokens:   4096,
		IsEnabled:   true,
		Description: "Test model",
		Scope:       "personal",
	}

	body, _ := json.Marshal(model)
	req, _ := http.NewRequest("POST", "/api/v1/llm-models", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var response models.LLMModel
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "Test GPT-4", response.Name)
	assert.Equal(t, "openai", response.LLMType)
	assert.Equal(t, "gpt-4", response.Model)
	assert.True(t, response.IsEnabled)
}

func TestGetLLMModel(t *testing.T) {
	db := setupTestDB(t)
	router := setupTestRouter(db)

	// First create a model
	testModel := models.LLMModel{
		ID:          "test-model-id",
		Name:        "Test Model",
		LLMType:     "openai",
		BaseURL:     "https://api.openai.com/v1",
		Model:       "gpt-4",
		Temperature: 0.7,
		MaxTokens:   4096,
		IsEnabled:   true,
	}
	db.Create(&testModel)

	req, _ := http.NewRequest("GET", "/api/v1/llm-models/test-model-id", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response models.LLMModel
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "Test Model", response.Name)
}

func TestGetLLMModel_NotFound(t *testing.T) {
	db := setupTestDB(t)
	router := setupTestRouter(db)

	req, _ := http.NewRequest("GET", "/api/v1/llm-models/non-existent-id", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestUpdateLLMModel(t *testing.T) {
	db := setupTestDB(t)
	router := setupTestRouter(db)

	// First create a model owned by the test user in the test realm
	userID := "test-user-id"
	realmID := "test-realm-id"
	testModel := models.LLMModel{
		ID:          "test-model-id",
		Name:        "Original Name",
		LLMType:     "openai",
		BaseURL:     "https://api.openai.com/v1",
		Model:       "gpt-4",
		Temperature: 0.7,
		MaxTokens:   4096,
		IsEnabled:   true,
		UserID:      &userID,
		RealmID:     &realmID,
	}
	db.Create(&testModel)

	updateReq := UpdateLLMModelRequest{
		Name:        "Updated Name",
		LLMType:     "openai",
		BaseURL:     "https://api.openai.com/v1",
		Model:       "gpt-4-turbo",
		Temperature: 0.8,
		MaxTokens:   8192,
		IsEnabled:   true,
		Description: "Updated description",
	}

	body, _ := json.Marshal(updateReq)
	req, _ := http.NewRequest("PUT", "/api/v1/llm-models/test-model-id", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response models.LLMModel
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "Updated Name", response.Name)
	assert.Equal(t, "gpt-4-turbo", response.Model)
}

func TestDeleteLLMModel(t *testing.T) {
	db := setupTestDB(t)
	router := setupTestRouter(db)

	// First create a model owned by the test user in the test realm
	userID := "test-user-id"
	realmID := "test-realm-id"
	testModel := models.LLMModel{
		ID:          "test-model-id",
		Name:        "To Delete",
		LLMType:     "openai",
		BaseURL:     "https://api.openai.com/v1",
		Model:       "gpt-4",
		Temperature: 0.7,
		MaxTokens:   4096,
		IsEnabled:   true,
		UserID:      &userID,
		RealmID:     &realmID,
	}
	db.Create(&testModel)

	req, _ := http.NewRequest("DELETE", "/api/v1/llm-models/test-model-id", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)

	// Verify soft delete
	var count int64
	db.Model(&models.LLMModel{}).Where("id = ? AND deleted_at IS NULL", "test-model-id").Count(&count)
	assert.Equal(t, int64(0), count)
}

func TestSetDefaultLLMModel(t *testing.T) {
	db := setupTestDB(t)
	router := setupTestRouter(db)

	// Create two models
	userID := "test-user-id"
	model1 := models.LLMModel{
		ID:        "model-1",
		Name:      "Model 1",
		LLMType:   "openai",
		BaseURL:   "https://api.openai.com/v1",
		Model:     "gpt-4",
		IsEnabled: true,
		IsDefault: false,
		UserID:    &userID,
	}
	model2 := models.LLMModel{
		ID:        "model-2",
		Name:      "Model 2",
		LLMType:   "openai",
		BaseURL:   "https://api.openai.com/v1",
		Model:     "gpt-3.5-turbo",
		IsEnabled: true,
		IsDefault: true,
		UserID:    &userID,
	}
	db.Create(&model1)
	db.Create(&model2)

	// Set model-1 as default
	req, _ := http.NewRequest("POST", "/api/v1/llm-models/model-1/default", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// Verify model-1 is now default
	var updatedModel1 models.LLMModel
	db.First(&updatedModel1, "id = ?", "model-1")
	assert.True(t, updatedModel1.IsDefault)

	// Verify model-2 is no longer default
	var updatedModel2 models.LLMModel
	db.First(&updatedModel2, "id = ?", "model-2")
	assert.False(t, updatedModel2.IsDefault)
}

func TestToggleLLMModelEnabled(t *testing.T) {
	db := setupTestDB(t)
	router := setupTestRouter(db)

	// Create a model
	userID := "test-user-id"
	testModel := models.LLMModel{
		ID:        "test-model-id",
		Name:      "Test Model",
		LLMType:   "openai",
		BaseURL:   "https://api.openai.com/v1",
		Model:     "gpt-4",
		IsEnabled: true,
		UserID:    &userID,
	}
	db.Create(&testModel)

	// Disable the model
	toggleReq := ToggleLLMModelRequest{Enabled: false}
	body, _ := json.Marshal(toggleReq)
	req, _ := http.NewRequest("POST", "/api/v1/llm-models/test-model-id/toggle", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response models.LLMModel
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.False(t, response.IsEnabled)
}

func TestGetDefaultLLMModel_NoDefault(t *testing.T) {
	db := setupTestDB(t)
	router := setupTestRouter(db)

	req, _ := http.NewRequest("GET", "/api/v1/llm-models/default", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Nil(t, response["model"])
	assert.True(t, response["use_legacy"].(bool))
}

func TestListLLMModels_WithFilters(t *testing.T) {
	db := setupTestDB(t)
	router := setupTestRouter(db)

	// Create some models (global templates - no user/realm)
	model1 := models.LLMModel{
		ID:        "model-1",
		Name:      "GPT-4",
		LLMType:   "openai",
		BaseURL:   "https://api.openai.com/v1",
		Model:     "gpt-4",
		IsEnabled: true,
		// No UserID or RealmID = global template
	}
	model2 := models.LLMModel{
		ID:        "model-2",
		Name:      "Claude",
		LLMType:   "anthropic",
		BaseURL:   "https://api.anthropic.com/v1",
		Model:     "claude-3-opus",
		IsEnabled: false,
		// No UserID or RealmID = global template
	}
	db.Create(&model1)
	db.Create(&model2)

	// Test search filter (should return only GPT-4 when searching for "GPT")
	req, _ := http.NewRequest("GET", "/api/v1/llm-models?q=GPT", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	// GPT search should return 1 result
	assert.Equal(t, float64(1), response["total"])

	// Test search for "Claude" (should return only Claude)
	req, _ = http.NewRequest("GET", "/api/v1/llm-models?q=Claude", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, float64(1), response["total"])
}

