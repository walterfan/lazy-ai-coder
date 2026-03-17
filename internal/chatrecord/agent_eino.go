package chatrecord

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
	"github.com/walterfan/lazy-ai-coder/pkg/models"
)

// EinoAgent implements the Agent interface using the Eino framework
type EinoAgent struct {
	chatModel model.ChatModel
	config    *AgentConfig
}

// NewEinoAgent creates a new Eino-based learning record agent
func NewEinoAgent(ctx context.Context, config *AgentConfig) (*EinoAgent, error) {
	if config == nil {
		config = DefaultAgentConfig()
	}

	chatModel, err := openai.NewChatModel(ctx, &openai.ChatModelConfig{
		BaseURL:     config.BaseURL,
		APIKey:      config.APIKey,
		Model:       config.Model,
		Temperature: &config.Temperature,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create chat model: %w", err)
	}

	return &EinoAgent{
		chatModel: chatModel,
		config:    config,
	}, nil
}

// NewEinoAgentWithModel creates a new Eino agent with an existing chat model (for testing)
func NewEinoAgentWithModel(chatModel model.ChatModel, config *AgentConfig) *EinoAgent {
	if config == nil {
		config = DefaultAgentConfig()
	}
	return &EinoAgent{
		chatModel: chatModel,
		config:    config,
	}
}

// Process classifies user input and generates a type-appropriate response
func (a *EinoAgent) Process(ctx context.Context, input string) (*ProcessResult, error) {
	return a.ProcessWithHistory(ctx, input, nil)
}

// ProcessWithHistory supports multi-turn: uses history for context when present
func (a *EinoAgent) ProcessWithHistory(ctx context.Context, input string, history []*schema.Message) (*ProcessResult, error) {
	// Step 1: Classify the input (with history for context on follow-ups)
	inputType, err := a.classifyWithHistory(ctx, input, history)
	if err != nil {
		return nil, fmt.Errorf("classification failed: %w", err)
	}

	// Step 2: Generate response based on type (with history for continuity)
	payload, err := a.generateResponseWithHistory(ctx, input, inputType, history)
	if err != nil {
		return nil, fmt.Errorf("response generation failed: %w", err)
	}

	return &ProcessResult{
		InputType:       inputType,
		ResponsePayload: payload,
	}, nil
}

// classifyWithHistory classifies the latest input; history provides context for follow-up questions
func (a *EinoAgent) classifyWithHistory(ctx context.Context, input string, history []*schema.Message) (string, error) {
	systemPrompt := `You are an input classifier for a Code Mate agent. Classify the user's input into exactly one of these categories:
- "research_solution": Comparing or researching tech options, tools, or approaches (e.g., "gRPC vs REST for microservices", "best database for real-time")
- "learn_tech": Learning a new technology, framework, or concept (e.g., "Learn Kubernetes", "WebAssembly basics", "How does OAuth2 work?")
- "tech_design": Designing a technical solution, API, or system (e.g., "Design API for file sync", "Architecture for a habit tracker")

Respond with ONLY the category name in lowercase, nothing else.

Classification rules:
1. Questions about comparing/choosing technologies or approaches → "research_solution"
2. Questions about learning or understanding a technology/topic → "learn_tech"
3. Requests to design, plan, or architect a system/feature → "tech_design"
When uncertain, prefer "research_solution".`

	messages := make([]*schema.Message, 0, 4+len(history))
	messages = append(messages, &schema.Message{Role: schema.System, Content: systemPrompt})
	if len(history) > 0 {
		messages = append(messages, history...)
	}
	messages = append(messages, &schema.Message{Role: schema.User, Content: input})

	resp, err := a.chatModel.Generate(ctx, messages)
	if err != nil {
		return models.InputTypeResearchSolution, fmt.Errorf("LLM call failed: %w", err)
	}

	// Parse the classification result
	classification := strings.TrimSpace(strings.ToLower(resp.Content))
	validTypes := models.ValidInputTypesCodeMate()
	for _, t := range validTypes {
		if classification == t {
			return classification, nil
		}
	}
	return models.InputTypeResearchSolution, nil
}

// generateResponseWithHistory generates a type-appropriate response; history gives prior turns for continuity
func (a *EinoAgent) generateResponseWithHistory(ctx context.Context, input, inputType string, history []*schema.Message) (*models.ResponsePayloadData, error) {
	var systemPrompt string
	var responseFormat string

	switch inputType {
	case models.InputTypeResearchSolution:
		systemPrompt = `You are a tech research assistant. For the user's question about technologies, tools, or approaches, provide:
1. A brief summary
2. Options with pros and cons for each
3. Trade-offs to consider
4. A clear recommendation

Respond in JSON format with: "summary", "options" (array of objects with "name", "pros", "cons", "summary"), "trade_offs", "recommendation", optional "references".`
		responseFormat = "research_solution"

	case models.InputTypeLearnTech:
		systemPrompt = `You are a technical learning advisor. For the given topic or technology, provide:
1. Brief introduction
2. Key concepts (name, description, importance)
3. Learning path (order, title, description, duration, objectives)
4. Recommended resources (type, title, url, description, difficulty)
5. Prerequisites
6. Time estimate

Respond in JSON format with: "introduction", "key_concepts", "learning_path", "resources", "prerequisites", "time_estimate".`
		responseFormat = "learn_tech"

	case models.InputTypeTechDesign:
		systemPrompt = `You are a technical design assistant. For the user's design request, provide:
1. Problem statement
2. Approach options (brief list)
3. Chosen approach and rationale
4. Components or APIs (list)
5. Risks and mitigations

Respond in JSON format with: "problem_statement", "approach_options", "chosen_approach", "components", "risks".`
		responseFormat = "tech_design"

	default:
		return nil, fmt.Errorf("unknown input type: %s", inputType)
	}

	messages := make([]*schema.Message, 0, 4+len(history))
	messages = append(messages, &schema.Message{Role: schema.System, Content: systemPrompt})
	if len(history) > 0 {
		messages = append(messages, history...)
	}
	messages = append(messages, &schema.Message{Role: schema.User, Content: input})

	resp, err := a.chatModel.Generate(ctx, messages)
	if err != nil {
		return nil, fmt.Errorf("LLM call failed: %w", err)
	}

	// Parse the JSON response
	payload, err := a.parseResponse(resp.Content, responseFormat)
	if err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return payload, nil
}

// parseResponse parses the LLM JSON response into ResponsePayloadData
func (a *EinoAgent) parseResponse(content, responseFormat string) (*models.ResponsePayloadData, error) {
	// Clean up the response - remove markdown code blocks if present
	content = strings.TrimSpace(content)
	if strings.HasPrefix(content, "```json") {
		content = strings.TrimPrefix(content, "```json")
		content = strings.TrimSuffix(content, "```")
		content = strings.TrimSpace(content)
	} else if strings.HasPrefix(content, "```") {
		content = strings.TrimPrefix(content, "```")
		content = strings.TrimSuffix(content, "```")
		content = strings.TrimSpace(content)
	}

	payload := &models.ResponsePayloadData{}
	if err := json.Unmarshal([]byte(content), payload); err != nil {
		// If JSON parsing fails, try to extract meaningful content
		return a.fallbackParse(content, responseFormat)
	}

	return payload, nil
}

// fallbackParse handles cases where the LLM doesn't return valid JSON
func (a *EinoAgent) fallbackParse(content, responseFormat string) (*models.ResponsePayloadData, error) {
	payload := &models.ResponsePayloadData{}
	switch responseFormat {
	case "research_solution":
		payload.Summary = content
		payload.Recommendation = content
	case "learn_tech":
		payload.Introduction = content
	case "tech_design":
		payload.ProblemStatement = content
		payload.ChosenApproach = content
	default:
		payload.Explanation = content
	}
	return payload, nil
}
