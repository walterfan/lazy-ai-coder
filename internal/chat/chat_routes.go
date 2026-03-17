package chat

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/walterfan/lazy-ai-coder/internal/log"
	"github.com/walterfan/lazy-ai-coder/internal/models"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins for simplicity
	},
}

// ChatHandlers handles chat-related HTTP requests
type ChatHandlers struct {
	service *ChatService
}

// NewChatHandlers creates a new ChatHandlers
func NewChatHandlers(service *ChatService) *ChatHandlers {
	return &ChatHandlers{
		service: service,
	}
}

// HandleChatRequest godoc
// @Summary Process LLM chat request
// @Description Process LLM requests with optional memory, GitLab integration, and streaming
// @Tags llm
// @Accept json
// @Produce json
// @Param request body models.ChatAPIRequest true "Chat Request"
// @Success 200 {object} models.ChatAPIResponse
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /process [post]
func (h *ChatHandlers) HandleChatRequest(c *gin.Context) {
	logger := log.GetLogger()
	var req models.ChatAPIRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	logger.Infof("received HTTP request with settings - BaseUrl: %s, Model: %s, Remember: %v, SessionId: %s",
		req.Settings.LlmBaseUrl, req.Settings.LlmModel, req.Remember, req.SessionId)

	// Build prompt with memory
	promptResult, err := h.service.BuildPromptTextWithMemory(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Process chat request
	answer, err := h.service.ProcessChatRequest(&req, promptResult)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "LLM processing failed"})
		return
	}

	c.JSON(http.StatusOK, models.ChatAPIResponse{
		Answer: answer,
	})
}

// HandleWebSocket handles WebSocket connections for streaming chat
func (h *ChatHandlers) HandleWebSocket(c *gin.Context) {
	logger := log.GetLogger()
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		logger.Error("Failed to upgrade connection")
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Failed to upgrade connection"})
		return
	}
	defer conn.Close()

	var req models.ChatAPIRequest
	if err := conn.ReadJSON(&req); err != nil {
		logger.Error("Invalid request")
		conn.WriteJSON(gin.H{"error": "Invalid request"})
		return
	}
	logger.Infof("received WebSocket request with settings - BaseUrl: %s, Model: %s, Remember: %v, SessionId: %s",
		req.Settings.LlmBaseUrl, req.Settings.LlmModel, req.Remember, req.SessionId)

	// Build prompt with memory
	promptResult, err := h.service.BuildPromptTextWithMemory(&req)
	if err != nil {
		conn.WriteJSON(gin.H{"error": err.Error()})
		return
	}

	// Process streaming chat request with WebSocket callback
	err = h.service.ProcessStreamingChat(&req, promptResult, func(chunk string) {
		_ = conn.WriteMessage(websocket.TextMessage, []byte(chunk))
	})

	if err != nil {
		_ = conn.WriteJSON(gin.H{"error": "LLM processing failed"})
	}
}
