package chatrecord

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/cloudwego/eino/schema"
	"github.com/google/uuid"

	"github.com/walterfan/lazy-ai-coder/internal/chatrecord/memory"
	"github.com/walterfan/lazy-ai-coder/pkg/models"
)

// ConfirmRequest is the request to confirm and save a learning record
type ConfirmRequest struct {
	UserInput       string                      `json:"user_input" binding:"required"`
	InputType       string                      `json:"input_type" binding:"required"`
	ResponsePayload *models.ResponsePayloadData `json:"response_payload" binding:"required"`
}

// SubmitRequest is the request to submit user input for classification and response
type SubmitRequest struct {
	UserInput string `json:"user_input" binding:"required"`
	SessionID string `json:"session_id,omitempty"`
}

// SubmitResult is the result of submitting user input
type SubmitResult struct {
	InputType       string                      `json:"input_type"`
	ResponsePayload *models.ResponsePayloadData `json:"response_payload"`
	SimilarRecords  []models.ChatRecord         `json:"similar_records,omitempty"`
	SessionID       string                      `json:"session_id,omitempty"`
}

// Service handles learning record business logic
type Service struct {
	repo         Repository
	agent        Agent
	sessionStore memory.SessionStore
}

// NewService creates a new learning record service
func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

// NewServiceWithAgent creates a new learning record service with an agent
func NewServiceWithAgent(repo Repository, agent Agent) *Service {
	return &Service{repo: repo, agent: agent}
}

// NewServiceWithMemory creates a service with agent and session memory
func NewServiceWithMemory(repo Repository, agent Agent, sessionStore memory.SessionStore) *Service {
	return &Service{repo: repo, agent: agent, sessionStore: sessionStore}
}

// SetAgent sets the agent for the service (useful for lazy initialization)
func (s *Service) SetAgent(agent Agent) {
	s.agent = agent
}

// SetSessionStore sets the session store for the service
func (s *Service) SetSessionStore(store memory.SessionStore) {
	s.sessionStore = store
}

// ErrAgentNotConfigured is returned when no agent is available and no request config provided
var ErrAgentNotConfigured = fmt.Errorf("agent not configured")

// SubmitInput processes user input through the agent and returns classification and response
// It does NOT persist to database - that happens only on confirm
func (s *Service) SubmitInput(ctx context.Context, req *SubmitRequest, userID string) (*SubmitResult, error) {
	return s.SubmitInputWithConfig(ctx, req, userID, nil)
}

// SubmitInputWithConfig processes user input using the service agent or, if nil, an agent created from config.
// When config is non-nil and has APIKey, a one-off agent is created for this request (e.g. from request settings).
func (s *Service) SubmitInputWithConfig(ctx context.Context, req *SubmitRequest, userID string, config *AgentConfig) (*SubmitResult, error) {
	if req.UserInput == "" {
		return nil, fmt.Errorf("user_input is required")
	}

	agent := s.agent
	if agent == nil && config != nil && config.APIKey != "" {
		var err error
		agent, err = NewEinoAgent(ctx, config)
		if err != nil {
			return nil, fmt.Errorf("agent from request config failed: %w", err)
		}
	}
	if agent == nil {
		return nil, ErrAgentNotConfigured
	}

	// Generate session ID if not provided
	sessionID := req.SessionID
	if sessionID == "" {
		sessionID = uuid.New().String()
	}

	// Multi-turn: get conversation history from session if available
	var history []*schema.Message
	if s.sessionStore != nil {
		history = s.sessionStore.ToSchemaMessages(sessionID)
	}

	// Process input through agent (with history for follow-up context)
	processResult, err := agent.ProcessWithHistory(ctx, req.UserInput, history)
	if err != nil {
		return nil, fmt.Errorf("agent process failed: %w", err)
	}

	// Find similar records for reinforcement (optional, ignore errors)
	var similarRecords []models.ChatRecord
	if userID != "" {
		similarRecords, _ = s.repo.FindSimilar(ctx, userID, req.UserInput, 3)
	}

	// Store in session memory for context in subsequent calls
	if s.sessionStore != nil {
		// Create a summary of the response for session context
		responseSummary := formatResponseSummary(processResult.InputType, processResult.ResponsePayload)
		s.sessionStore.AddMessage(sessionID, req.UserInput, responseSummary)
	}

	return &SubmitResult{
		InputType:       processResult.InputType,
		ResponsePayload: processResult.ResponsePayload,
		SimilarRecords:  similarRecords,
		SessionID:       sessionID,
	}, nil
}

// formatResponseSummary creates a brief summary of the response for session memory
func formatResponseSummary(inputType string, payload *models.ResponsePayloadData) string {
	if payload == nil {
		return ""
	}
	var summary string
	switch inputType {
	case models.InputTypeResearchSolution:
		summary = payload.Summary
		if summary == "" {
			summary = payload.Recommendation
		}
	case models.InputTypeLearnTech, models.InputTypeTopic:
		summary = payload.Introduction
	case models.InputTypeTechDesign, models.InputTypeIdea:
		if len(payload.Plan) > 0 {
			summary = "Plan: " + payload.Plan[0]
		} else {
			summary = payload.ProblemStatement
			if summary == "" {
				summary = payload.ChosenApproach
			}
		}
	case models.InputTypeWord, models.InputTypeSentence:
		summary = payload.Explanation
	case models.InputTypeQuestion:
		summary = payload.Answer
	default:
		summary = payload.Explanation
		if summary == "" {
			summary = payload.Answer
		}
	}
	if len(summary) > 200 {
		summary = summary[:200] + "..."
	}
	return summary
}

// GetSessionContext returns the session context as schema messages (for agent use)
func (s *Service) GetSessionContext(sessionID string) []memory.SessionMessage {
	if s.sessionStore == nil {
		return nil
	}
	return s.sessionStore.GetRecentMessages(sessionID)
}

// ClearSession clears the session memory for a given session ID
func (s *Service) ClearSession(sessionID string) error {
	if s.sessionStore == nil {
		return nil
	}
	return s.sessionStore.Clear(sessionID)
}

// CreateRecord creates a new learning record from a confirm request
func (s *Service) CreateRecord(ctx context.Context, req *ConfirmRequest, userID, realmID string) (*models.ChatRecord, error) {
	// Validate input type
	if !models.IsValidInputType(req.InputType) {
		return nil, fmt.Errorf("invalid input type: %s", req.InputType)
	}

	// Validate required fields
	if req.UserInput == "" {
		return nil, fmt.Errorf("user_input is required")
	}

	// Marshal response payload to JSON string
	payloadJSON, err := json.Marshal(req.ResponsePayload)
	if err != nil {
		return nil, fmt.Errorf("marshal response payload: %w", err)
	}

	record := &models.ChatRecord{
		ID:              uuid.New().String(),
		InputType:       req.InputType,
		UserInput:       req.UserInput,
		ResponsePayload: string(payloadJSON),
		UserID:          userID,
		RealmID:         realmID,
		CreatedBy:       userID,
	}

	if err := s.repo.Create(ctx, record); err != nil {
		return nil, fmt.Errorf("create record: %w", err)
	}

	return record, nil
}

// ListRecords returns paginated learning records for a user
func (s *Service) ListRecords(ctx context.Context, userID string, filters ListFilters, page, pageSize int) (*ListResult, error) {
	// Validate type filter if provided
	if filters.Type != "" && !models.IsValidInputType(filters.Type) {
		return nil, fmt.Errorf("invalid filter type: %s", filters.Type)
	}

	return s.repo.FindByUserWithFilters(ctx, userID, filters, page, pageSize)
}

// GetRecord returns a single learning record by ID
func (s *Service) GetRecord(ctx context.Context, id, userID string) (*models.ChatRecord, error) {
	record, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Verify ownership
	if record.UserID != userID {
		return nil, fmt.Errorf("access denied: record belongs to another user")
	}

	return record, nil
}

// DeleteRecord soft-deletes a learning record
func (s *Service) DeleteRecord(ctx context.Context, id, userID string) error {
	// First verify ownership
	record, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return err
	}

	if record.UserID != userID {
		return fmt.Errorf("access denied: record belongs to another user")
	}

	return s.repo.SoftDelete(ctx, id)
}

// GetStats returns learning record statistics for a user
func (s *Service) GetStats(ctx context.Context, userID string) (*Stats, error) {
	// Get counts by type
	byType, err := s.repo.CountByType(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("count by type: %w", err)
	}

	// Calculate total
	var total int64
	for _, count := range byType {
		total += count
	}

	// Get streak
	streak, err := s.repo.CountStreak(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("count streak: %w", err)
	}

	// Get last record time
	lastRecordAt, err := s.repo.GetLastRecordTime(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get last record time: %w", err)
	}

	return &Stats{
		Total:        total,
		ByType:       byType,
		Streak:       streak,
		LastRecordAt: lastRecordAt,
	}, nil
}

// FindSimilar finds learning records similar to the given input
func (s *Service) FindSimilar(ctx context.Context, userID, input string, limit int) ([]models.ChatRecord, error) {
	return s.repo.FindSimilar(ctx, userID, input, limit)
}
