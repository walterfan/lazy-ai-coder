package chatrecord

import (
	"context"
	"strings"
	"testing"

	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/walterfan/lazy-ai-coder/pkg/models"
)

// MockChatModel implements model.ChatModel for testing
type MockChatModel struct {
	classifyResponse  string
	generateResponse  string
	generateResponses map[string]string // type -> response
}

func (m *MockChatModel) Generate(ctx context.Context, messages []*schema.Message, opts ...model.Option) (*schema.Message, error) {
	// Determine if this is a classification or generation call based on system prompt
	systemPrompt := ""
	for _, msg := range messages {
		if msg.Role == schema.System {
			systemPrompt = msg.Content
		}
	}

	// Classification call - returns type
	if strings.Contains(systemPrompt, "input classifier") {
		return &schema.Message{
			Role:    schema.Assistant,
			Content: m.classifyResponse,
		}, nil
	}

	// Generation call - returns response based on type
	if m.generateResponses != nil {
		for typeKey, response := range m.generateResponses {
			if strings.Contains(systemPrompt, typeKey) {
				return &schema.Message{
					Role:    schema.Assistant,
					Content: response,
				}, nil
			}
		}
	}

	// Default response
	return &schema.Message{
		Role:    schema.Assistant,
		Content: m.generateResponse,
	}, nil
}

func (m *MockChatModel) Stream(ctx context.Context, messages []*schema.Message, opts ...model.Option) (*schema.StreamReader[*schema.Message], error) {
	return nil, nil
}

func (m *MockChatModel) BindTools(tools []*schema.ToolInfo) error {
	return nil
}

// Ensure MockChatModel implements model.ChatModel
var _ model.ChatModel = (*MockChatModel)(nil)

// Test classification of different input types (Code Mate: research_solution, learn_tech, tech_design)
func TestEinoAgent_Classify_ResearchSolution(t *testing.T) {
	mockModel := &MockChatModel{
		classifyResponse: "research_solution",
		generateResponse: `{"summary": "gRPC vs REST comparison", "recommendation": "Use gRPC for internal services", "options": [{"name": "gRPC", "pros": "Fast", "cons": "Complex"}]}`,
	}

	agent := NewEinoAgentWithModel(mockModel, nil)
	result, err := agent.Process(context.Background(), "gRPC vs REST for microservices")

	require.NoError(t, err)
	assert.Equal(t, models.InputTypeResearchSolution, result.InputType)
	assert.NotEmpty(t, result.ResponsePayload.Summary)
	assert.NotEmpty(t, result.ResponsePayload.Recommendation)
}

func TestEinoAgent_Classify_LearnTech(t *testing.T) {
	mockModel := &MockChatModel{
		classifyResponse: "learn_tech",
		generateResponse: `{"introduction": "Go is a statically typed language", "time_estimate": "2-4 weeks", "key_concepts": [{"name": "Goroutines", "description": "Lightweight threads"}]}`,
	}

	agent := NewEinoAgentWithModel(mockModel, nil)
	result, err := agent.Process(context.Background(), "Learn Go")

	require.NoError(t, err)
	assert.Equal(t, models.InputTypeLearnTech, result.InputType)
	assert.NotEmpty(t, result.ResponsePayload.Introduction)
	assert.NotEmpty(t, result.ResponsePayload.TimeEstimate)
}

func TestEinoAgent_Classify_TechDesign(t *testing.T) {
	mockModel := &MockChatModel{
		classifyResponse: "tech_design",
		generateResponse: `{"problem_statement": "Need file sync API", "chosen_approach": "REST with webhooks", "components": ["API server", "sync engine"], "risks": "Conflict resolution"}`,
	}

	agent := NewEinoAgentWithModel(mockModel, nil)
	result, err := agent.Process(context.Background(), "Design API for file sync")

	require.NoError(t, err)
	assert.Equal(t, models.InputTypeTechDesign, result.InputType)
	assert.NotEmpty(t, result.ResponsePayload.ProblemStatement)
	assert.NotEmpty(t, result.ResponsePayload.ChosenApproach)
}

func TestEinoAgent_InvalidClassification_FallbackToResearchSolution(t *testing.T) {
	mockModel := &MockChatModel{
		classifyResponse: "invalid_type",
		generateResponse: `{"summary": "Summary", "recommendation": "Use X"}`,
	}

	agent := NewEinoAgentWithModel(mockModel, nil)
	result, err := agent.Process(context.Background(), "something unclear")

	require.NoError(t, err)
	assert.Equal(t, models.InputTypeResearchSolution, result.InputType)
}

func TestEinoAgent_ParseResponse_ResearchSolutionWithMarkdown(t *testing.T) {
	mockModel := &MockChatModel{
		classifyResponse: "research_solution",
		generateResponse: "```json\n{\"summary\": \"Summary text\", \"recommendation\": \"Use A\"}\n```",
	}

	agent := NewEinoAgentWithModel(mockModel, nil)
	result, err := agent.Process(context.Background(), "compare A vs B")

	require.NoError(t, err)
	assert.Equal(t, models.InputTypeResearchSolution, result.InputType)
	assert.Equal(t, "Summary text", result.ResponsePayload.Summary)
	assert.Equal(t, "Use A", result.ResponsePayload.Recommendation)
}

func TestEinoAgent_FallbackParse_TechDesign(t *testing.T) {
	mockModel := &MockChatModel{
		classifyResponse: "tech_design",
		generateResponse: "Plain text problem and approach",
	}

	agent := NewEinoAgentWithModel(mockModel, nil)
	result, err := agent.Process(context.Background(), "Design a system")

	require.NoError(t, err)
	assert.Equal(t, models.InputTypeTechDesign, result.InputType)
	assert.Contains(t, result.ResponsePayload.ProblemStatement, "Plain text")
}

func TestEinoAgent_FallbackParse_LearnTech(t *testing.T) {
	mockModel := &MockChatModel{
		classifyResponse: "learn_tech",
		generateResponse: "Plain text introduction to the topic.",
	}

	agent := NewEinoAgentWithModel(mockModel, nil)
	result, err := agent.Process(context.Background(), "Learn Kubernetes")

	require.NoError(t, err)
	assert.Equal(t, models.InputTypeLearnTech, result.InputType)
	assert.Contains(t, result.ResponsePayload.Introduction, "Plain text")
}

// Test different response formats for Code Mate types
func TestEinoAgent_GenerateResponse_AllTypes(t *testing.T) {
	testCases := []struct {
		name         string
		inputType    string
		response     string
		checkPayload func(t *testing.T, p *models.ResponsePayloadData)
	}{
		{
			name:      "Research solution response",
			inputType: "research_solution",
			response:  `{"summary": "Comparison summary", "recommendation": "Use option A", "options": [{"name": "A", "pros": "Fast", "cons": "Complex"}]}`,
			checkPayload: func(t *testing.T, p *models.ResponsePayloadData) {
				assert.NotEmpty(t, p.Summary)
				assert.NotEmpty(t, p.Recommendation)
			},
		},
		{
			name:      "Learn tech response",
			inputType: "learn_tech",
			response:  `{"introduction": "Intro to tech", "time_estimate": "2 weeks", "key_concepts": [{"name": "Concept", "description": "Desc"}]}`,
			checkPayload: func(t *testing.T, p *models.ResponsePayloadData) {
				assert.NotEmpty(t, p.Introduction)
				assert.NotEmpty(t, p.TimeEstimate)
			},
		},
		{
			name:      "Tech design response",
			inputType: "tech_design",
			response:  `{"problem_statement": "Need API", "chosen_approach": "REST", "components": ["API", "DB"], "risks": "None"}`,
			checkPayload: func(t *testing.T, p *models.ResponsePayloadData) {
				assert.NotEmpty(t, p.ProblemStatement)
				assert.NotEmpty(t, p.ChosenApproach)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockModel := &MockChatModel{
				classifyResponse: tc.inputType,
				generateResponse: tc.response,
			}

			agent := NewEinoAgentWithModel(mockModel, nil)
			result, err := agent.Process(context.Background(), "test input")

			require.NoError(t, err)
			assert.Equal(t, tc.inputType, result.InputType)
			tc.checkPayload(t, result.ResponsePayload)
		})
	}
}

// Test DefaultAgentConfig
func TestDefaultAgentConfig(t *testing.T) {
	config := DefaultAgentConfig()

	assert.Equal(t, "https://api.openai.com/v1", config.BaseURL)
	assert.Equal(t, "gpt-4", config.Model)
	assert.Equal(t, float32(0.7), config.Temperature)
	assert.Equal(t, 2048, config.MaxTokens)
}
