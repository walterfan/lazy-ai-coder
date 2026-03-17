package memory

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/walterfan/lazy-ai-coder/pkg/models"
)

// MockRepository implements Repository for testing
type MockRepository struct {
	records []models.ChatRecord
}

func (m *MockRepository) FindSimilar(ctx context.Context, userID string, input string, limit int) ([]models.ChatRecord, error) {
	// Simple mock: return all records for the user up to limit
	var result []models.ChatRecord
	for _, r := range m.records {
		if r.UserID == userID {
			result = append(result, r)
		}
	}
	if len(result) > limit {
		result = result[:limit]
	}
	return result, nil
}

func TestNewFuzzySimilarityRetriever(t *testing.T) {
	repo := &MockRepository{}
	retriever := NewFuzzySimilarityRetriever(repo)
	assert.NotNil(t, retriever)
	assert.Equal(t, 0.3, retriever.threshold)
}

func TestNewFuzzySimilarityRetrieverWithThreshold(t *testing.T) {
	repo := &MockRepository{}

	// Normal threshold
	retriever := NewFuzzySimilarityRetrieverWithThreshold(repo, 0.5)
	assert.Equal(t, 0.5, retriever.threshold)

	// Below 0 should be clamped to 0
	retriever = NewFuzzySimilarityRetrieverWithThreshold(repo, -0.5)
	assert.Equal(t, 0.0, retriever.threshold)

	// Above 1 should be clamped to 1
	retriever = NewFuzzySimilarityRetrieverWithThreshold(repo, 1.5)
	assert.Equal(t, 1.0, retriever.threshold)
}

func TestFuzzySimilarityRetriever_FindSimilar(t *testing.T) {
	repo := &MockRepository{
		records: []models.ChatRecord{
			{ID: "1", UserID: "user-1", UserInput: "serendipity", InputType: models.InputTypeWord},
			{ID: "2", UserID: "user-1", UserInput: "ephemeral", InputType: models.InputTypeWord},
			{ID: "3", UserID: "user-2", UserInput: "ubiquitous", InputType: models.InputTypeWord},
		},
	}

	retriever := NewFuzzySimilarityRetriever(repo)

	results, err := retriever.FindSimilar(context.Background(), "user-1", "serendip", 10)
	require.NoError(t, err)
	assert.Len(t, results, 2) // Only user-1's records
}

func TestFuzzySimilarityRetriever_FindSimilar_Limit(t *testing.T) {
	repo := &MockRepository{
		records: []models.ChatRecord{
			{ID: "1", UserID: "user-1", UserInput: "word1", InputType: models.InputTypeWord},
			{ID: "2", UserID: "user-1", UserInput: "word2", InputType: models.InputTypeWord},
			{ID: "3", UserID: "user-1", UserInput: "word3", InputType: models.InputTypeWord},
		},
	}

	retriever := NewFuzzySimilarityRetriever(repo)

	// Respect limit
	results, err := retriever.FindSimilar(context.Background(), "user-1", "word", 2)
	require.NoError(t, err)
	assert.Len(t, results, 2)
}

func TestFuzzySimilarityRetriever_FindSimilar_DefaultLimit(t *testing.T) {
	repo := &MockRepository{
		records: []models.ChatRecord{
			{ID: "1", UserID: "user-1", UserInput: "word1", InputType: models.InputTypeWord},
			{ID: "2", UserID: "user-1", UserInput: "word2", InputType: models.InputTypeWord},
			{ID: "3", UserID: "user-1", UserInput: "word3", InputType: models.InputTypeWord},
			{ID: "4", UserID: "user-1", UserInput: "word4", InputType: models.InputTypeWord},
		},
	}

	retriever := NewFuzzySimilarityRetriever(repo)

	// Zero or negative limit defaults to 3
	results, err := retriever.FindSimilar(context.Background(), "user-1", "word", 0)
	require.NoError(t, err)
	assert.Len(t, results, 3)
}

func TestFormatSimilarRecordsForPrompt_Empty(t *testing.T) {
	result := FormatSimilarRecordsForPrompt(nil, 3)
	assert.Empty(t, result)

	result = FormatSimilarRecordsForPrompt([]models.ChatRecord{}, 3)
	assert.Empty(t, result)
}

func TestFormatSimilarRecordsForPrompt_WithRecords(t *testing.T) {
	records := []models.ChatRecord{
		{
			ID:              "1",
			UserInput:       "serendipity",
			InputType:       models.InputTypeWord,
			ResponsePayload: `{"explanation": "意外发现"}`,
		},
		{
			ID:              "2",
			UserInput:       "How does OAuth2 work?",
			InputType:       models.InputTypeQuestion,
			ResponsePayload: `{"answer": "OAuth2 is an authorization framework"}`,
		},
	}

	result := FormatSimilarRecordsForPrompt(records, 3)
	assert.Contains(t, result, "previously recorded")
	assert.Contains(t, result, "serendipity")
	assert.Contains(t, result, "word")
	assert.Contains(t, result, "OAuth2")
	assert.Contains(t, result, "question")
}

func TestFormatSimilarRecordsForPrompt_MaxRecords(t *testing.T) {
	records := []models.ChatRecord{
		{ID: "1", UserInput: "word1", InputType: models.InputTypeWord, ResponsePayload: `{"explanation": "test"}`},
		{ID: "2", UserInput: "word2", InputType: models.InputTypeWord, ResponsePayload: `{"explanation": "test"}`},
		{ID: "3", UserInput: "word3", InputType: models.InputTypeWord, ResponsePayload: `{"explanation": "test"}`},
		{ID: "4", UserInput: "word4", InputType: models.InputTypeWord, ResponsePayload: `{"explanation": "test"}`},
	}

	result := FormatSimilarRecordsForPrompt(records, 2)
	// Should only include 2 records
	assert.Contains(t, result, "word1")
	assert.Contains(t, result, "word2")
	assert.NotContains(t, result, "word3")
	assert.NotContains(t, result, "word4")
}

func TestFormatSimilarRecordsForPrompt_LongSummary(t *testing.T) {
	longExplanation := "This is a very long explanation that should be truncated because it exceeds the maximum length allowed for summaries in the prompt formatting function which is 100 characters"

	records := []models.ChatRecord{
		{
			ID:              "1",
			UserInput:       "test",
			InputType:       models.InputTypeWord,
			ResponsePayload: `{"explanation": "` + longExplanation + `"}`,
		},
	}

	result := FormatSimilarRecordsForPrompt(records, 3)
	// Should be truncated with "..."
	assert.Contains(t, result, "...")
	assert.True(t, len(result) < len(longExplanation)+100) // Reasonably truncated
}

func TestFormatSimilarRecordsForPrompt_DifferentTypes(t *testing.T) {
	records := []models.ChatRecord{
		{
			ID:              "1",
			UserInput:       "test word",
			InputType:       models.InputTypeWord,
			ResponsePayload: `{"explanation": "Word explanation"}`,
		},
		{
			ID:              "2",
			UserInput:       "test question",
			InputType:       models.InputTypeQuestion,
			ResponsePayload: `{"answer": "Question answer"}`,
		},
		{
			ID:              "3",
			UserInput:       "test idea",
			InputType:       models.InputTypeIdea,
			ResponsePayload: `{"plan": ["Step 1", "Step 2"]}`,
		},
		{
			ID:              "4",
			UserInput:       "test topic",
			InputType:       models.InputTypeTopic,
			ResponsePayload: `{"introduction": "Topic intro"}`,
		},
	}

	result := FormatSimilarRecordsForPrompt(records, 10)
	assert.Contains(t, result, "Word explanation")
	assert.Contains(t, result, "Question answer")
	assert.Contains(t, result, "Step 1")
	assert.Contains(t, result, "Topic intro")
}
