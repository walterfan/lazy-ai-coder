package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/walterfan/lazy-ai-coder/internal/chatrecord"
	"github.com/walterfan/lazy-ai-coder/internal/chatrecord/memory"
	"github.com/walterfan/lazy-ai-coder/internal/llm"
	"github.com/walterfan/lazy-ai-coder/pkg/models"
)

const defaultLLMBaseURL = "https://api.openai.com/v1"
const defaultLLMModel = "gpt-4"
const defaultLLMTemperature = 0.7

// ChatRecordHandlers handles learning record CRUD endpoints
type ChatRecordHandlers struct {
	service *chatrecord.Service
}

// NewChatRecordHandlers creates new learning record handlers
func NewChatRecordHandlers(db *gorm.DB) *ChatRecordHandlers {
	repo := chatrecord.NewGormRepository(db)
	service := chatrecord.NewService(repo)
	return &ChatRecordHandlers{
		service: service,
	}
}

// NewChatRecordHandlersWithAgent creates new learning record handlers with an agent
func NewChatRecordHandlersWithAgent(db *gorm.DB, agent chatrecord.Agent) *ChatRecordHandlers {
	repo := chatrecord.NewGormRepository(db)
	service := chatrecord.NewServiceWithAgent(repo, agent)
	return &ChatRecordHandlers{
		service: service,
	}
}

// NewChatRecordHandlersWithAgentAndMemory creates handlers with agent and session store for multi-turn
func NewChatRecordHandlersWithAgentAndMemory(db *gorm.DB, agent chatrecord.Agent, sessionStore memory.SessionStore) *ChatRecordHandlers {
	repo := chatrecord.NewGormRepository(db)
	service := chatrecord.NewServiceWithMemory(repo, agent, sessionStore)
	return &ChatRecordHandlers{
		service: service,
	}
}

// SetAgent sets the agent for the handlers' service
func (h *ChatRecordHandlers) SetAgent(agent chatrecord.Agent) {
	h.service.SetAgent(agent)
}

// SubmitChatRecordRequest is the request body for submitting input
type SubmitChatRecordRequest struct {
	UserInput string `json:"user_input" binding:"required"`
	SessionID string `json:"session_id,omitempty"`
	// Optional skill context (SKILL.md content) to guide the LLM conversation
	SkillContext string `json:"skill_context,omitempty"`
	// Optional LLM settings from client (e.g. Settings page). Used when server has no LLM_API_KEY set.
	LLMApiKey      string `json:"LLM_API_KEY,omitempty"`
	LLMBaseURL     string `json:"LLM_BASE_URL,omitempty"`
	LLMModel       string `json:"LLM_MODEL,omitempty"`
	LLMTemperature string `json:"LLM_TEMPERATURE,omitempty"`
}

// SubmitchatrecordResponse is the response for submit endpoint
type SubmitchatrecordResponse struct {
	InputType       string                      `json:"input_type"`
	ResponsePayload *models.ResponsePayloadData `json:"response_payload"`
	SimilarRecords  []models.ChatRecordSummary  `json:"similar_records,omitempty"`
	SessionID       string                      `json:"session_id,omitempty"`
}

// ConfirmRequest is the request body for confirming a learning record
type ConfirmchatrecordRequest struct {
	UserInput       string                      `json:"user_input" binding:"required"`
	InputType       string                      `json:"input_type" binding:"required"`
	ResponsePayload *models.ResponsePayloadData `json:"response_payload" binding:"required"`
}

// ListchatrecordsResponse is the response for listing learning records
type ListchatrecordsResponse struct {
	Records    []models.ChatRecordSummary `json:"records"`
	Total      int64                      `json:"total"`
	Page       int                        `json:"page"`
	PageSize   int                        `json:"page_size"`
	TotalPages int                        `json:"total_pages"`
}

// StatsResponse is the response for learning record statistics
type StatsResponse struct {
	Total        int64            `json:"total"`
	ByType       map[string]int64 `json:"by_type"`
	Streak       int              `json:"streak"`
	LastRecordAt *string          `json:"last_record_at,omitempty"`
}

// HandleSubmit godoc
// @Summary Submit input for classification and response generation
// @Description Process user input through AI agent to classify and generate type-appropriate response (does NOT persist to database)
// @Tags chat-record
// @Accept json
// @Produce json
// @Param request body SubmitChatRecordRequest true "User input to process"
// @Success 200 {object} SubmitchatrecordResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/chat-record/submit [post]
func (h *ChatRecordHandlers) HandleSubmit(c *gin.Context) {
	_, userID, _, _, _ := GetUserContext(c)
	if userID == "" {
		userID = "anonymous"
	}

	var req SubmitChatRecordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Convert to service request
	serviceReq := &chatrecord.SubmitRequest{
		UserInput:    req.UserInput,
		SessionID:    req.SessionID,
		SkillContext: req.SkillContext,
	}

	// Build optional agent config from request (client can send LLM settings when server has none)
	var agentConfig *chatrecord.AgentConfig
	if req.LLMApiKey != "" {
		agentConfig = &chatrecord.AgentConfig{
			APIKey:      req.LLMApiKey,
			BaseURL:     req.LLMBaseURL,
			Model:       req.LLMModel,
			Temperature: float32(defaultLLMTemperature),
			MaxTokens:   2048,
		}
		if agentConfig.BaseURL == "" {
			agentConfig.BaseURL = defaultLLMBaseURL
		}
		if agentConfig.Model == "" {
			agentConfig.Model = defaultLLMModel
		}
		if req.LLMTemperature != "" {
			if f, err := strconv.ParseFloat(req.LLMTemperature, 32); err == nil {
				agentConfig.Temperature = float32(f)
			}
		}
	}

	result, err := h.service.SubmitInputWithConfig(c.Request.Context(), serviceReq, userID, agentConfig)
	if err != nil {
		if err == chatrecord.ErrAgentNotConfigured {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"error": "LLM is not configured. Set LLM_API_KEY in the server environment or configure LLM in Settings and ensure it is sent with the request.",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Convert similar records to summaries
	var similarSummaries []models.ChatRecordSummary
	for _, record := range result.SimilarRecords {
		similarSummaries = append(similarSummaries, record.ToSummary(100))
	}

	c.JSON(http.StatusOK, SubmitchatrecordResponse{
		InputType:       result.InputType,
		ResponsePayload: result.ResponsePayload,
		SimilarRecords:  similarSummaries,
		SessionID:       result.SessionID,
	})
}

// HandleConfirm godoc
// @Summary Confirm and save a learning record
// @Description Create a new learning record after user confirms
// @Tags chat-record
// @Accept json
// @Produce json
// @Param request body ConfirmchatrecordRequest true "Learning record to confirm"
// @Success 201 {object} models.ChatRecord
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/chat-record/confirm [post]
func (h *ChatRecordHandlers) HandleConfirm(c *gin.Context) {
	// Get user context
	_, userID, realmID, _, _ := GetUserContext(c)
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	var req ConfirmchatrecordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Convert to service request
	serviceReq := &chatrecord.ConfirmRequest{
		UserInput:       req.UserInput,
		InputType:       req.InputType,
		ResponsePayload: req.ResponsePayload,
	}

	record, err := h.service.CreateRecord(c.Request.Context(), serviceReq, userID, realmID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, record)
}

// HandleList godoc
// @Summary List learning records
// @Description Get paginated list of learning records with optional filters
// @Tags chat-record
// @Produce json
// @Param type query string false "Filter by input type (word, sentence, question, idea)"
// @Param search query string false "Search in user_input and response"
// @Param date_from query string false "Filter records created after this date (RFC3339)"
// @Param date_to query string false "Filter records created before this date (RFC3339)"
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(10)
// @Success 200 {object} ListchatrecordsResponse
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/chat-record/list [get]
func (h *ChatRecordHandlers) HandleList(c *gin.Context) {
	// Get user context
	_, userID, _, _, _ := GetUserContext(c)
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	// Parse filters
	filters := chatrecord.ListFilters{
		Type:   c.Query("type"),
		Search: c.Query("search"),
	}

	// Parse date filters
	if dateFromStr := c.Query("date_from"); dateFromStr != "" {
		if dateFrom, err := time.Parse(time.RFC3339, dateFromStr); err == nil {
			filters.DateFrom = &dateFrom
		}
	}
	if dateToStr := c.Query("date_to"); dateToStr != "" {
		if dateTo, err := time.Parse(time.RFC3339, dateToStr); err == nil {
			filters.DateTo = &dateTo
		}
	}

	// Parse pagination
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}
	if pageSize > 100 {
		pageSize = 100
	}

	result, err := h.service.ListRecords(c.Request.Context(), userID, filters, page, pageSize)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Convert to summaries
	summaries := make([]models.ChatRecordSummary, len(result.Records))
	for i, record := range result.Records {
		summaries[i] = record.ToSummary(100)
	}

	c.JSON(http.StatusOK, ListchatrecordsResponse{
		Records:    summaries,
		Total:      result.Total,
		Page:       result.Page,
		PageSize:   result.PageSize,
		TotalPages: result.TotalPages,
	})
}

// HandleGet godoc
// @Summary Get a learning record by ID
// @Description Get a single learning record by its ID
// @Tags chat-record
// @Produce json
// @Param id path string true "Record ID"
// @Success 200 {object} models.ChatRecord
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /api/v1/chat-record/{id} [get]
func (h *ChatRecordHandlers) HandleGet(c *gin.Context) {
	// Get user context
	_, userID, _, _, _ := GetUserContext(c)
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id is required"})
		return
	}

	record, err := h.service.GetRecord(c.Request.Context(), id, userID)
	if err != nil {
		if err.Error() == "record not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "record not found"})
			return
		}
		if err.Error() == "access denied: record belongs to another user" {
			c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, record)
}

// HandleDelete godoc
// @Summary Delete a learning record
// @Description Soft-delete a learning record by ID
// @Tags chat-record
// @Param id path string true "Record ID"
// @Success 204 "No Content"
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /api/v1/chat-record/{id} [delete]
func (h *ChatRecordHandlers) HandleDelete(c *gin.Context) {
	// Get user context
	_, userID, _, _, _ := GetUserContext(c)
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id is required"})
		return
	}

	err := h.service.DeleteRecord(c.Request.Context(), id, userID)
	if err != nil {
		if err.Error() == "record not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "record not found"})
			return
		}
		if err.Error() == "access denied: record belongs to another user" {
			c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

// HandleStats godoc
// @Summary Get learning record statistics
// @Description Get statistics about the user's learning records
// @Tags chat-record
// @Produce json
// @Success 200 {object} StatsResponse
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/chat-record/stats [get]
func (h *ChatRecordHandlers) HandleStats(c *gin.Context) {
	// Get user context
	_, userID, _, _, _ := GetUserContext(c)
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	stats, err := h.service.GetStats(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var lastRecordAtStr *string
	if stats.LastRecordAt != nil {
		s := stats.LastRecordAt.Format(time.RFC3339)
		lastRecordAtStr = &s
	}

	c.JSON(http.StatusOK, StatsResponse{
		Total:        stats.Total,
		ByType:       stats.ByType,
		Streak:       stats.Streak,
		LastRecordAt: lastRecordAtStr,
	})
}

// HandleStreamSubmit streams an LLM response via Server-Sent Events.
// It reuses the client-side LLM settings and session memory but bypasses
// the classify-then-JSON-generate pipeline, returning raw markdown instead.
func (h *ChatRecordHandlers) HandleStreamSubmit(c *gin.Context) {
	_, userID, _, _, _ := GetUserContext(c)
	if userID == "" {
		userID = "anonymous"
	}

	var req SubmitChatRecordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.UserInput == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user_input is required"})
		return
	}

	sessionID := req.SessionID
	if sessionID == "" {
		sessionID = uuid.New().String()
	}

	// Collect conversation history from session store
	var history []llm.ChatMessage
	for _, msg := range h.service.GetSessionContext(sessionID) {
		history = append(history, llm.ChatMessage{
			Role:    string(msg.Role),
			Content: msg.Content,
		})
	}

	// Build system prompt
	systemPrompt := "You are a helpful coding assistant. Provide clear, well-structured answers in markdown format."
	if req.SkillContext != "" {
		systemPrompt = req.SkillContext + "\n\n" + systemPrompt
	}

	// Build LLM settings from the client-provided config
	settings := llm.LLMSettings{}
	if req.LLMApiKey != "" {
		settings.ApiKey = req.LLMApiKey
		settings.BaseUrl = req.LLMBaseURL
		settings.Model = req.LLMModel
		if req.LLMTemperature != "" {
			if f, err := strconv.ParseFloat(req.LLMTemperature, 64); err == nil {
				settings.Temperature = f
			}
		}
	}

	// Set SSE headers
	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Header().Set("X-Accel-Buffering", "no")

	// Send the session_id as the first event so the frontend can track it
	fmt.Fprintf(c.Writer, "data: {\"session_id\":%q}\n\n", sessionID)
	c.Writer.Flush()

	var fullResponse strings.Builder

	err := llm.AskLLMWithStreamAndMemory(systemPrompt, req.UserInput, history, settings, func(chunk string) {
		// Filter out the wrapper tags injected by AskLLMWithStreamAndMemory
		if chunk == "<answer>" || chunk == "</answer>" {
			return
		}
		fullResponse.WriteString(chunk)
		fmt.Fprintf(c.Writer, "data: {\"content\":%q}\n\n", chunk)
		c.Writer.Flush()
	})

	if err != nil {
		fmt.Fprintf(c.Writer, "data: {\"error\":%q}\n\n", err.Error())
		c.Writer.Flush()
	}

	// Signal completion
	fmt.Fprintf(c.Writer, "data: [DONE]\n\n")
	c.Writer.Flush()

	// Persist the exchange in session memory for multi-turn context
	h.service.SaveToSession(sessionID, req.UserInput, fullResponse.String())
}
