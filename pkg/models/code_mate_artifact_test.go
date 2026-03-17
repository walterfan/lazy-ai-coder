package models

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupCodeMateArtifactTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	err = db.AutoMigrate(&CodeMateArtifact{})
	require.NoError(t, err)
	return db
}

func TestCodeMateArtifact_TableName(t *testing.T) {
	record := CodeMateArtifact{}
	assert.Equal(t, "code_mate_artifacts", record.TableName())
}

func TestCodeMateArtifact_CreateAndRead(t *testing.T) {
	db := setupCodeMateArtifactTestDB(t)
	payload := &CodeMateResponsePayload{
		Summary:        "gRPC is better for internal services",
		Recommendation: "Use gRPC for microservices",
	}
	payloadJSON, err := json.Marshal(payload)
	require.NoError(t, err)

	record := CodeMateArtifact{
		ID:              "test-artifact-1",
		InputType:       InputTypeResearchSolution,
		UserInput:       "gRPC vs REST",
		ResponsePayload: string(payloadJSON),
		UserID:          "user-123",
		RealmID:         "realm-456",
		CreatedBy:       "user-123",
	}
	err = db.Create(&record).Error
	require.NoError(t, err)

	var retrieved CodeMateArtifact
	err = db.First(&retrieved, "id = ?", "test-artifact-1").Error
	require.NoError(t, err)
	assert.Equal(t, "test-artifact-1", retrieved.ID)
	assert.Equal(t, InputTypeResearchSolution, retrieved.InputType)
	assert.Equal(t, "gRPC vs REST", retrieved.UserInput)
	assert.Equal(t, "user-123", retrieved.UserID)
	assert.Equal(t, "realm-456", retrieved.RealmID)
	assert.NotZero(t, retrieved.CreatedTime)
}

func TestCodeMateArtifact_GetSetResponsePayload(t *testing.T) {
	record := &CodeMateArtifact{
		ID:        "test-1",
		InputType: InputTypeLearnTech,
		UserInput: "Learn Go",
	}
	payload := &CodeMateResponsePayload{
		Introduction: "Go is a statically typed language",
		TimeEstimate: "2-4 weeks",
	}
	err := record.SetResponsePayload(payload)
	require.NoError(t, err)
	assert.NotEmpty(t, record.ResponsePayload)

	retrieved, err := record.GetResponsePayload()
	require.NoError(t, err)
	assert.Equal(t, "Go is a statically typed language", retrieved.Introduction)
	assert.Equal(t, "2-4 weeks", retrieved.TimeEstimate)
}

func TestCodeMateArtifact_GetResponsePayload_Empty(t *testing.T) {
	record := &CodeMateArtifact{ID: "test-1", ResponsePayload: ""}
	payload, err := record.GetResponsePayload()
	require.NoError(t, err)
	assert.NotNil(t, payload)
	assert.Empty(t, payload.Summary)
}

func TestCodeMateArtifact_SetResponsePayload_Nil(t *testing.T) {
	record := &CodeMateArtifact{ID: "test-1", ResponsePayload: "x"}
	err := record.SetResponsePayload(nil)
	require.NoError(t, err)
	assert.Empty(t, record.ResponsePayload)
}

func TestCodeMateArtifact_MarshalJSON(t *testing.T) {
	payload := &CodeMateResponsePayload{Summary: "Summary text", Recommendation: "Use X"}
	payloadJSON, _ := json.Marshal(payload)
	record := CodeMateArtifact{
		ID:              "test-id",
		InputType:       InputTypeResearchSolution,
		UserInput:       "test",
		ResponsePayload: string(payloadJSON),
		UserID:          "user-1",
		RealmID:         "realm-1",
		CreatedTime:     time.Now(),
	}
	data, err := json.Marshal(record)
	require.NoError(t, err)
	var result map[string]interface{}
	err = json.Unmarshal(data, &result)
	require.NoError(t, err)
	responsePayload, ok := result["response_payload"].(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, "Summary text", responsePayload["summary"])
	assert.Equal(t, "Use X", responsePayload["recommendation"])
}

func TestCodeMateArtifact_ToSummary(t *testing.T) {
	payload := &CodeMateResponsePayload{Introduction: "Long intro that should be truncated"}
	payloadJSON, _ := json.Marshal(payload)
	record := &CodeMateArtifact{
		ID:              "test-id",
		InputType:       InputTypeLearnTech,
		UserInput:       "Long user input to truncate",
		ResponsePayload: string(payloadJSON),
		CreatedTime:     time.Now(),
	}
	summary := record.ToSummary(20)
	assert.Equal(t, "test-id", summary.ID)
	assert.Equal(t, InputTypeLearnTech, summary.InputType)
	assert.True(t, len(summary.UserInput) <= 23 && strings.HasSuffix(summary.UserInput, "..."), "UserInput truncated: %q", summary.UserInput)
	assert.True(t, len(summary.ResponseSummary) <= 23 && strings.HasSuffix(summary.ResponseSummary, "..."), "ResponseSummary truncated: %q", summary.ResponseSummary)
}

func TestCodeMateArtifact_SoftDelete(t *testing.T) {
	db := setupCodeMateArtifactTestDB(t)
	record := CodeMateArtifact{
		ID:        "test-delete",
		InputType: InputTypeTechDesign,
		UserInput: "Design API",
		UserID:    "user-1",
	}
	err := db.Create(&record).Error
	require.NoError(t, err)
	err = db.Delete(&record).Error
	require.NoError(t, err)

	var notFound CodeMateArtifact
	err = db.First(&notFound, "id = ?", "test-delete").Error
	assert.Error(t, err)

	var found CodeMateArtifact
	err = db.Unscoped().Where("id = ?", "test-delete").First(&found).Error
	require.NoError(t, err)
	assert.NotNil(t, found.DeletedAt)
}

func TestValidInputTypesCodeMate(t *testing.T) {
	types := ValidInputTypesCodeMate()
	assert.Len(t, types, 3)
	assert.Contains(t, types, InputTypeResearchSolution)
	assert.Contains(t, types, InputTypeLearnTech)
	assert.Contains(t, types, InputTypeTechDesign)
}

func TestIsValidInputTypeCodeMate(t *testing.T) {
	assert.True(t, IsValidInputTypeCodeMate(InputTypeResearchSolution))
	assert.True(t, IsValidInputTypeCodeMate(InputTypeLearnTech))
	assert.True(t, IsValidInputTypeCodeMate(InputTypeTechDesign))
	assert.False(t, IsValidInputTypeCodeMate("invalid"))
	assert.False(t, IsValidInputTypeCodeMate(""))
}

func TestCodeMateArtifact_DifferentInputTypes(t *testing.T) {
	db := setupCodeMateArtifactTestDB(t)
	testCases := []struct {
		inputType string
		payload   CodeMateResponsePayload
	}{
		{
			inputType: InputTypeResearchSolution,
			payload:   CodeMateResponsePayload{Summary: "Research summary", Recommendation: "Use option A"},
		},
		{
			inputType: InputTypeLearnTech,
			payload:   CodeMateResponsePayload{Introduction: "Intro to tech", TimeEstimate: "2 weeks"},
		},
		{
			inputType: InputTypeTechDesign,
			payload:   CodeMateResponsePayload{ProblemStatement: "Need sync API", ChosenApproach: "REST"},
		},
	}
	for i, tc := range testCases {
		record := &CodeMateArtifact{
			ID:        "artifact-" + tc.inputType,
			InputType: tc.inputType,
			UserInput: "Input " + tc.inputType,
			UserID:    "user-1",
		}
		err := record.SetResponsePayload(&tc.payload)
		require.NoError(t, err, "Case %d", i)
		err = db.Create(record).Error
		require.NoError(t, err, "Case %d", i)

		var retrieved CodeMateArtifact
		err = db.First(&retrieved, "id = ?", record.ID).Error
		require.NoError(t, err, "Case %d", i)
		assert.Equal(t, tc.inputType, retrieved.InputType, "Case %d", i)
	}
}
