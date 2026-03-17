package mcp

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/walterfan/lazy-ai-coder/internal/log"
	"go.uber.org/zap"
)

// HTTPServer wraps the MCP server for HTTP transport
type HTTPServer struct {
	server    *Server
	logger    *zap.SugaredLogger
	sseBroker *SSEBroker
}

// NewHTTPServer creates a new HTTP-based MCP server
func NewHTTPServer() *HTTPServer {
	broker := NewSSEBroker()
	go broker.Run() // Start SSE broker in background

	server := NewServer()
	server.sseBroker = broker // Connect broker to server for notifications

	return &HTTPServer{
		server:    server,
		logger:    log.GetLogger(),
		sseBroker: broker,
	}
}

// RegisterAllTools registers all MCP tools
func (h *HTTPServer) RegisterAllTools() {
	h.server.RegisterAllTools()
}

// InitializeWithDB initializes the server with a database connection
func (h *HTTPServer) InitializeWithDB() error {
	return h.server.InitializeWithDB()
}

// HandleMCPRequest godoc
// @Summary MCP JSON-RPC endpoint
// @Description Handle MCP (Model Context Protocol) JSON-RPC 2.0 requests for tool calls
// @Tags mcp
// @Accept json
// @Produce json
// @Param request body JSONRPCRequest true "JSON-RPC Request"
// @Success 200 {object} JSONRPCResponse
// @Failure 400 {object} JSONRPCResponse
// @Router /mcp [post]
func (h *HTTPServer) HandleMCPRequest(c *gin.Context) {
	var req JSONRPCRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Errorf("Invalid MCP request: %v", err)
		// Use -1 as ID for parse errors (Cursor doesn't accept null)
		c.JSON(http.StatusBadRequest, h.server.createErrorResponse(-1, -32700, "Parse error", err.Error()))
		return
	}

	h.logger.Infof("Received HTTP MCP request: method=%s, id=%v", req.Method, req.ID)

	// Handle the request using the existing server logic
	response := h.server.handleRequest(&req)

	// Send JSON response
	c.JSON(http.StatusOK, response)

	h.logger.Infof("Sent HTTP MCP response for method: %s", req.Method)
}

// HandleServerInfo godoc
// @Summary Get MCP server info
// @Description Get MCP server information including version, protocol version, and capabilities
// @Tags mcp
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /mcp/info [get]
func (h *HTTPServer) HandleServerInfo(c *gin.Context) {
	h.server.mu.RLock()
	hasPrompts := len(h.server.prompts) > 0
	hasResources := len(h.server.resources) > 0
	h.server.mu.RUnlock()

	info := map[string]interface{}{
		"name":            ServerName,
		"version":         ServerVersion,
		"protocolVersion": ProtocolVersion,
		"transport":       []string{"http", "sse"},
		"capabilities": map[string]interface{}{
			"tools":     true,
			"resources": hasResources,
			"prompts":   hasPrompts,
		},
		"sse": map[string]interface{}{
			"endpoint":       "/api/v1/mcp/events",
			"connectedClients": h.sseBroker.GetClientCount(),
		},
	}

	c.JSON(http.StatusOK, info)
}

// HandleListTools godoc
// @Summary List MCP tools
// @Description List all available MCP tools with their names and descriptions
// @Tags mcp
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /mcp/tools [get]
func (h *HTTPServer) HandleListTools(c *gin.Context) {
	h.server.mu.RLock()
	defer h.server.mu.RUnlock()

	tools := make([]map[string]interface{}, 0, len(h.server.tools))
	for _, tool := range h.server.tools {
		tools = append(tools, map[string]interface{}{
			"name":        tool.Name,
			"description": tool.Description,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"tools": tools,
		"count": len(tools),
	})
}

// HandleListPrompts godoc
// @Summary List MCP prompts
// @Description List all available MCP prompts with their names and descriptions
// @Tags mcp
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /mcp/prompts [get]
func (h *HTTPServer) HandleListPrompts(c *gin.Context) {
	h.server.mu.RLock()
	defer h.server.mu.RUnlock()

	prompts := make([]map[string]interface{}, 0, len(h.server.prompts))
	for _, prompt := range h.server.prompts {
		prompts = append(prompts, map[string]interface{}{
			"name":        prompt.Name,
			"description": prompt.Description,
			"arguments":   prompt.Arguments,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"prompts": prompts,
		"count":   len(prompts),
	})
}

// HandleListResources godoc
// @Summary List MCP resources
// @Description List all available MCP resources with their URIs and descriptions
// @Tags mcp
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /mcp/resources [get]
func (h *HTTPServer) HandleListResources(c *gin.Context) {
	h.server.mu.RLock()
	defer h.server.mu.RUnlock()

	resources := make([]map[string]interface{}, 0, len(h.server.resources))
	for _, resource := range h.server.resources {
		resources = append(resources, map[string]interface{}{
			"uri":         resource.URI,
			"name":        resource.Name,
			"description": resource.Description,
			"mimeType":    resource.MimeType,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"resources": resources,
		"count":     len(resources),
	})
}

// HandleSSE handles Server-Sent Events connections
// @Summary MCP SSE endpoint
// @Description Subscribe to real-time MCP server events (prompts/resources updates, tool execution)
// @Tags mcp
// @Produce text/event-stream
// @Success 200 {string} string "SSE event stream"
// @Router /mcp/events [get]
func (h *HTTPServer) HandleSSE(c *gin.Context) {
	// Set SSE headers
	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
	c.Writer.Header().Set("X-Accel-Buffering", "no") // Disable nginx buffering

	// Create client
	clientID := fmt.Sprintf("client_%d", time.Now().UnixNano())
	client := &SSEClient{
		ID:       clientID,
		Channel:  make(chan SSEMessage, 256),
		LastSeen: time.Now(),
	}

	// Register client
	h.sseBroker.RegisterClient(client)
	defer h.sseBroker.UnregisterClient(client)

	// Send initial connection message
	fmt.Fprintf(c.Writer, "event: connected\n")
	fmt.Fprintf(c.Writer, "data: {\"clientId\":\"%s\",\"timestamp\":\"%s\"}\n\n", clientID, time.Now().Format(time.RFC3339))
	c.Writer.Flush()

	h.logger.Infof("SSE client connected: %s", clientID)

	// Create a channel for client disconnect
	clientGone := c.Request.Context().Done()

	// Stream messages
	for {
		select {
		case <-clientGone:
			h.logger.Infof("SSE client disconnected: %s", clientID)
			return

		case message, ok := <-client.Channel:
			if !ok {
				h.logger.Infof("SSE client channel closed: %s", clientID)
				return
			}

			// Write event
			if message.Event != "" {
				fmt.Fprintf(c.Writer, "event: %s\n", message.Event)
			}
			if message.ID != "" {
				fmt.Fprintf(c.Writer, "id: %s\n", message.ID)
			}
			fmt.Fprintf(c.Writer, "data: %s\n\n", message.Data)

			// Flush the response writer
			if flusher, ok := c.Writer.(http.Flusher); ok {
				flusher.Flush()
			}
		}
	}
}

// ToolCallRequest represents a request to call an MCP tool
type ToolCallRequest struct {
	Name      string                 `json:"name" binding:"required"`
	Arguments map[string]interface{} `json:"arguments"`
	// Settings passed from frontend for authentication
	Settings *ToolCallSettings `json:"settings,omitempty"`
}

// ToolCallSettings contains credentials passed from frontend
type ToolCallSettings struct {
	GitlabToken string `json:"gitlab_token,omitempty"`
	GitlabURL   string `json:"gitlab_url,omitempty"`
	LLMApiKey   string `json:"llm_api_key,omitempty"`
	LLMBaseURL  string `json:"llm_base_url,omitempty"`
	LLMModel    string `json:"llm_model,omitempty"`
}

// HandleCallTool godoc
// @Summary Call a specific MCP tool
// @Description Simplified endpoint to call a specific tool by name with arguments
// @Tags mcp
// @Accept json
// @Produce json
// @Param request body ToolCallRequest true "Tool call request with name, arguments, and optional settings"
// @Success 200 {object} CallToolResult
// @Failure 400 {object} map[string]interface{}
// @Router /mcp/v1/call-tool [post]
func (h *HTTPServer) HandleCallTool(c *gin.Context) {
	var req ToolCallRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Errorf("Invalid tool call request: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request",
			"message": err.Error(),
		})
		return
	}

	h.logger.Infof("Calling tool: %s with args: %v", req.Name, req.Arguments)

	// Inject settings into arguments if provided from frontend
	if req.Settings != nil {
		if req.Arguments == nil {
			req.Arguments = make(map[string]interface{})
		}
		// Inject GitLab credentials if provided and not already in arguments
		if req.Settings.GitlabToken != "" {
			if _, exists := req.Arguments["gitlab_token"]; !exists {
				req.Arguments["gitlab_token"] = req.Settings.GitlabToken
			}
		}
		if req.Settings.GitlabURL != "" {
			if _, exists := req.Arguments["gitlab_url"]; !exists {
				req.Arguments["gitlab_url"] = req.Settings.GitlabURL
			}
		}
		// Inject LLM credentials if provided and not already in arguments
		if req.Settings.LLMApiKey != "" {
			if _, exists := req.Arguments["llm_api_key"]; !exists {
				req.Arguments["llm_api_key"] = req.Settings.LLMApiKey
			}
		}
		if req.Settings.LLMBaseURL != "" {
			if _, exists := req.Arguments["llm_base_url"]; !exists {
				req.Arguments["llm_base_url"] = req.Settings.LLMBaseURL
			}
		}
		if req.Settings.LLMModel != "" {
			if _, exists := req.Arguments["llm_model"]; !exists {
				req.Arguments["llm_model"] = req.Settings.LLMModel
			}
		}
	}

	// Get the tool handler
	h.server.mu.RLock()
	handler, exists := h.server.handlers[req.Name]
	h.server.mu.RUnlock()

	if !exists {
		h.logger.Errorf("Tool not found: %s", req.Name)
		c.JSON(http.StatusOK, gin.H{
			"isError": true,
			"content": []map[string]interface{}{
				{
					"type": "text",
					"text": fmt.Sprintf("Tool not found: %s", req.Name),
				},
			},
		})
		return
	}

	// Call the tool handler
	result, err := handler(req.Arguments)
	if err != nil {
		h.logger.Errorf("Tool call failed: %v", err)
		c.JSON(http.StatusOK, gin.H{
			"isError": true,
			"content": []map[string]interface{}{
				{
					"type": "text",
					"text": fmt.Sprintf("Error: %v", err),
				},
			},
		})
		return
	}

	c.JSON(http.StatusOK, result)
}

// SetupRoutes adds MCP routes to the Gin router
func (h *HTTPServer) SetupRoutes(r *gin.Engine) {
	// MCP Protocol endpoint (POST for JSON-RPC)
	r.POST("/api/v1/mcp", h.HandleMCPRequest)

	// Convenience GET endpoints
	r.GET("/api/v1/mcp/info", h.HandleServerInfo)
	r.GET("/api/v1/mcp/tools", h.HandleListTools)
	r.GET("/api/v1/mcp/prompts", h.HandleListPrompts)
	r.GET("/api/v1/mcp/resources", h.HandleListResources)

	// SSE endpoint for real-time updates
	r.GET("/api/v1/mcp/events", h.HandleSSE)

	// Simplified tool call endpoint
	r.POST("/mcp/v1/call-tool", h.HandleCallTool)

	h.logger.Info("MCP HTTP endpoints registered at /api/v1/mcp")
	h.logger.Info("MCP tool call endpoint registered at /mcp/v1/call-tool")
	h.logger.Info("MCP SSE endpoint registered at /api/v1/mcp/events")
}
