package mcp

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"

	"github.com/walterfan/lazy-ai-coder/internal/log"
	"github.com/walterfan/lazy-ai-coder/pkg/database"
	"github.com/walterfan/lazy-ai-coder/pkg/models"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

const (
	// MCP Protocol Version
	ProtocolVersion = "2024-11-05"

	// Server Information
	ServerName    = "lazy-ai-coder"
	ServerVersion = "1.0.0"
)

// Server represents an MCP server instance
type Server struct {
	tools        []Tool
	resources    []Resource
	prompts      []Prompt
	handlers     map[string]ToolHandler
	mu           sync.RWMutex
	logger       *zap.SugaredLogger
	db           *gorm.DB
	promptsMap   map[string]*models.Prompt // Cache for prompt lookup
	resourcesMap map[string]Resource       // Cache for resource lookup
	sseBroker    *SSEBroker                // SSE broker for notifications (optional, only for HTTP server)
}

// ToolHandler is a function that handles tool execution
type ToolHandler func(arguments map[string]interface{}) (*CallToolResult, error)

// NewServer creates a new MCP server instance
func NewServer() *Server {
	return &Server{
		tools:        make([]Tool, 0),
		resources:    make([]Resource, 0),
		prompts:      make([]Prompt, 0),
		handlers:     make(map[string]ToolHandler),
		logger:       log.GetLogger(),
		promptsMap:   make(map[string]*models.Prompt),
		resourcesMap: make(map[string]Resource),
	}
}

// InitializeWithDB initializes the server with a database connection
func (s *Server) InitializeWithDB() error {
	// Initialize database connection
	if err := database.InitDB(); err != nil {
		s.logger.Warnf("Failed to initialize database for prompts/resources: %v", err)
		return err
	}
	s.db = database.GetDB()

	// Load prompts from database
	if err := s.loadPromptsFromDB(); err != nil {
		s.logger.Warnf("Failed to load prompts from database: %v", err)
	}

	// Register default resources
	s.registerDefaultResources()

	return nil
}

// loadPromptsFromDB loads prompts from the database and registers them as MCP prompts
func (s *Server) loadPromptsFromDB() error {
	if s.db == nil {
		return fmt.Errorf("database not initialized")
	}

	var dbPrompts []models.Prompt
	if err := s.db.Find(&dbPrompts).Error; err != nil {
		return fmt.Errorf("failed to query prompts: %w", err)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	for i := range dbPrompts {
		prompt := &dbPrompts[i]

		// Parse tags for logging
		tags := strings.Split(prompt.Tags, ",")
		var arguments []PromptArgument

		// Use explicit arguments from database if available
		if prompt.Arguments != "" {
			// Parse JSON arguments from database
			var dbArgs []models.PromptArgument
			if err := json.Unmarshal([]byte(prompt.Arguments), &dbArgs); err != nil {
				s.logger.Warnf("Failed to parse arguments for prompt '%s': %v, falling back to extraction", prompt.Name, err)
				// Fallback to extracting from template
				variables := extractVariables(prompt.UserPrompt + " " + prompt.SystemPrompt)
				for _, varName := range variables {
					arguments = append(arguments, PromptArgument{
						Name:        varName,
						Description: fmt.Sprintf("Value for %s", varName),
						Required:    true,
					})
				}
			} else {
				// Convert models.PromptArgument to MCP PromptArgument
				for _, arg := range dbArgs {
					arguments = append(arguments, PromptArgument{
						Name:        arg.Name,
						Description: arg.Description,
						Required:    arg.Required,
					})
				}
			}
		} else {
			// No explicit arguments, extract from template for backward compatibility
			variables := extractVariables(prompt.UserPrompt + " " + prompt.SystemPrompt)
			for _, varName := range variables {
				arguments = append(arguments, PromptArgument{
					Name:        varName,
					Description: fmt.Sprintf("Value for %s", varName),
					Required:    true,
				})
			}
		}

		mcpPrompt := Prompt{
			Name:        prompt.Name,
			Description: prompt.Description,
			Arguments:   arguments,
		}

		s.prompts = append(s.prompts, mcpPrompt)
		s.promptsMap[prompt.Name] = prompt
		s.logger.Infof("Loaded prompt from DB: %s (%d arguments, tags: %s)",
			prompt.Name, len(arguments), strings.Join(tags, ", "))
	}

	s.logger.Infof("Loaded %d prompts from database", len(dbPrompts))

	// Notify SSE clients if broker is available
	if s.sseBroker != nil {
		s.sseBroker.NotifyPromptsUpdated(len(dbPrompts))
	}

	return nil
}

// extractVariables extracts {{variable}} patterns from a template string
func extractVariables(template string) []string {
	var variables []string
	seen := make(map[string]bool)

	for i := 0; i < len(template)-1; i++ {
		if template[i] == '{' && template[i+1] == '{' {
			// Found opening {{
			start := i + 2
			end := start

			// Find closing }}
			for j := start; j < len(template)-1; j++ {
				if template[j] == '}' && template[j+1] == '}' {
					end = j
					break
				}
			}

			if end > start {
				varName := strings.TrimSpace(template[start:end])
				if varName != "" && !seen[varName] {
					variables = append(variables, varName)
					seen[varName] = true
				}
				i = end + 1 // Skip past the closing }}
			}
		}
	}

	return variables
}

// registerDefaultResources registers default resources like GitLab projects and config files
func (s *Server) registerDefaultResources() {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Register GitLab project as a resource
	gitlabURL := os.Getenv("GITLAB_BASE_URL")
	if gitlabURL != "" {
		resource := Resource{
			URI:         "gitlab://projects",
			Name:        "GitLab Projects",
			Description: "Access to GitLab projects and repositories",
			MimeType:    "application/json",
		}
		s.resources = append(s.resources, resource)
		s.resourcesMap[resource.URI] = resource
		s.logger.Info("Registered GitLab projects resource")
	}

	// Register config files as resources
	configResource := Resource{
		URI:         "file://config/prompts.yaml",
		Name:        "Prompt Templates",
		Description: "System prompt template configuration",
		MimeType:    "text/yaml",
	}
	s.resources = append(s.resources, configResource)
	s.resourcesMap[configResource.URI] = configResource

	projectConfigResource := Resource{
		URI:         "file://config/config.yaml",
		Name:        "Application Configuration",
		Description: "Main application configuration file",
		MimeType:    "text/yaml",
	}
	s.resources = append(s.resources, projectConfigResource)
	s.resourcesMap[projectConfigResource.URI] = projectConfigResource

	s.logger.Infof("Registered %d default resources", len(s.resources))

	// Notify SSE clients if broker is available
	if s.sseBroker != nil {
		s.sseBroker.NotifyResourcesUpdated(len(s.resources))
	}
}

// RegisterTool registers a new tool with its handler
func (s *Server) RegisterTool(tool Tool, handler ToolHandler) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.tools = append(s.tools, tool)
	s.handlers[tool.Name] = handler
	s.logger.Infof("Registered MCP tool: %s", tool.Name)
}

// RegisterResource registers a new resource
func (s *Server) RegisterResource(resource Resource) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.resources = append(s.resources, resource)
	s.logger.Infof("Registered MCP resource: %s", resource.Name)
}

// RegisterPrompt registers a new prompt
func (s *Server) RegisterPrompt(prompt Prompt) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.prompts = append(s.prompts, prompt)
	s.logger.Infof("Registered MCP prompt: %s", prompt.Name)
}

// Start starts the MCP server using stdio transport
func (s *Server) Start() error {
	s.logger.Info("Starting MCP server on stdio...")

	reader := bufio.NewReader(os.Stdin)
	writer := bufio.NewWriter(os.Stdout)
	defer writer.Flush()

	for {
		// Read JSON-RPC request from stdin
		line, err := reader.ReadBytes('\n')
		if err != nil {
			if err == io.EOF {
				s.logger.Info("MCP server shutting down (EOF)")
				return nil
			}
			s.logger.Errorf("Error reading from stdin: %v", err)
			return err
		}

		// Parse JSON-RPC request
		var req JSONRPCRequest
		if err := json.Unmarshal(line, &req); err != nil {
			s.logger.Errorf("Error parsing JSON-RPC request: %v", err)
			// Use -1 as ID for parse errors (Cursor doesn't accept null)
			s.writeError(writer, -1, -32700, "Parse error", nil)
			continue
		}

		s.logger.Infof("Received MCP request: method=%s, id=%v", req.Method, req.ID)

		// Handle request
		response := s.handleRequest(&req)

		// Only send response if it's not nil (don't respond to notifications)
		if response != nil {
			// Write response to stdout
			responseJSON, err := json.Marshal(response)
			if err != nil {
				s.logger.Errorf("Error marshaling response: %v", err)
				continue
			}

			if _, err := writer.Write(responseJSON); err != nil {
				s.logger.Errorf("Error writing response: %v", err)
				return err
			}
			if _, err := writer.Write([]byte("\n")); err != nil {
				s.logger.Errorf("Error writing newline: %v", err)
				return err
			}
			if err := writer.Flush(); err != nil {
				s.logger.Errorf("Error flushing writer: %v", err)
				return err
			}

			s.logger.Infof("Sent MCP response for method: %s", req.Method)
		}
	}
}

// handleRequest handles a JSON-RPC request
func (s *Server) handleRequest(req *JSONRPCRequest) *JSONRPCResponse {
	var response *JSONRPCResponse

	switch req.Method {
	case "initialize":
		response = s.handleInitialize(req)
	case "notifications/initialized":
		s.logger.Info("Client initialized")
		return nil
	case "notifications/cancelled":
		s.logger.Info("Client cancelled request")
		return nil
	case "tools/list":
		response = s.handleListTools(req)
	case "tools/call":
		response = s.handleCallTool(req)
	case "resources/list":
		response = s.handleListResources(req)
	case "resources/read":
		response = s.handleReadResource(req)
	case "prompts/list":
		response = s.handleListPrompts(req)
	case "prompts/get":
		response = s.handleGetPrompt(req)
	default:
		// If ID is missing/null, it's a notification - do not reply
		if req.ID == nil {
			s.logger.Infof("Received notification: %s", req.Method)
			return nil
		}
		return s.createErrorResponse(req.ID, -32601, fmt.Sprintf("Method not found: %s", req.Method), nil)
	}

	// If request ID is nil (notification), do not send response
	if req.ID == nil {
		return nil
	}

	return response
}

// handleInitialize handles the initialize request
func (s *Server) handleInitialize(req *JSONRPCRequest) *JSONRPCResponse {
	var initReq InitializeRequest
	if err := json.Unmarshal(req.Params, &initReq); err != nil {
		return s.createErrorResponse(req.ID, -32602, "Invalid params", err.Error())
	}

	s.logger.Infof("Client connected: %s v%s", initReq.ClientInfo.Name, initReq.ClientInfo.Version)

	result := InitializeResult{
		ProtocolVersion: ProtocolVersion,
		ServerInfo: ServerInfo{
			Name:    ServerName,
			Version: ServerVersion,
		},
		Capabilities: ServerCapabilities{
			Tools: &ToolsCapability{
				ListChanged: false,
			},
			Resources: &ResourcesCapability{
				Subscribe:   false,
				ListChanged: false,
			},
			Prompts: &PromptsCapability{
				ListChanged: false,
			},
		},
	}

	return &JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result:  result,
	}
}

// handleListTools handles the tools/list request
func (s *Server) handleListTools(req *JSONRPCRequest) *JSONRPCResponse {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := ListToolsResult{
		Tools: s.tools,
	}

	s.logger.Infof("Listing %d tools", len(s.tools))

	return &JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result:  result,
	}
}

// handleCallTool handles the tools/call request
func (s *Server) handleCallTool(req *JSONRPCRequest) *JSONRPCResponse {
	var callReq CallToolRequest
	if err := json.Unmarshal(req.Params, &callReq); err != nil {
		return s.createErrorResponse(req.ID, -32602, "Invalid params", err.Error())
	}

	s.mu.RLock()
	handler, exists := s.handlers[callReq.Name]
	s.mu.RUnlock()

	if !exists {
		return s.createErrorResponse(req.ID, -32602, fmt.Sprintf("Tool not found: %s", callReq.Name), nil)
	}

	s.logger.Infof("Calling tool: %s with args: %v", callReq.Name, callReq.Arguments)

	result, err := handler(callReq.Arguments)
	if err != nil {
		s.logger.Errorf("Tool execution error: %v", err)
		return s.createErrorResponse(req.ID, -32603, "Tool execution failed", err.Error())
	}

	return &JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result:  result,
	}
}

// handleListResources handles the resources/list request
func (s *Server) handleListResources(req *JSONRPCRequest) *JSONRPCResponse {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := ListResourcesResult{
		Resources: s.resources,
	}

	s.logger.Infof("Listing %d resources", len(s.resources))

	return &JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result:  result,
	}
}

// handleReadResource handles the resources/read request
func (s *Server) handleReadResource(req *JSONRPCRequest) *JSONRPCResponse {
	var readReq ReadResourceRequest
	if err := json.Unmarshal(req.Params, &readReq); err != nil {
		return s.createErrorResponse(req.ID, -32602, "Invalid params", err.Error())
	}

	s.mu.RLock()
	resource, exists := s.resourcesMap[readReq.URI]
	s.mu.RUnlock()

	if !exists {
		return s.createErrorResponse(req.ID, -32602, fmt.Sprintf("Resource not found: %s", readReq.URI), nil)
	}

	s.logger.Infof("Reading resource: %s", readReq.URI)

	var content string
	var err error

	// Handle different resource types based on URI scheme
	if strings.HasPrefix(readReq.URI, "file://") {
		// Read local file
		filePath := strings.TrimPrefix(readReq.URI, "file://")
		contentBytes, readErr := os.ReadFile(filePath)
		if readErr != nil {
			return s.createErrorResponse(req.ID, -32603, fmt.Sprintf("Failed to read file: %v", readErr), nil)
		}
		content = string(contentBytes)
	} else if strings.HasPrefix(readReq.URI, "gitlab://") {
		// Handle GitLab resources
		if readReq.URI == "gitlab://projects" {
			content = `{
  "description": "GitLab projects resource",
  "usage": "Use get_gitlab_file_content or get_gitlab_merge_request tools to access GitLab data",
  "note": "GitLab URL and token must be configured in the Settings page"
}`
		} else {
			return s.createErrorResponse(req.ID, -32602, fmt.Sprintf("Unknown GitLab resource: %s", readReq.URI), nil)
		}
	} else {
		return s.createErrorResponse(req.ID, -32602, fmt.Sprintf("Unsupported URI scheme: %s", readReq.URI), nil)
	}

	if err != nil {
		return s.createErrorResponse(req.ID, -32603, fmt.Sprintf("Failed to read resource: %v", err), nil)
	}

	result := ReadResourceResult{
		Contents: []ResourceContent{
			{
				URI:      readReq.URI,
				MimeType: resource.MimeType,
				Text:     content,
			},
		},
	}

	return &JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result:  result,
	}
}

// handleListPrompts handles the prompts/list request
func (s *Server) handleListPrompts(req *JSONRPCRequest) *JSONRPCResponse {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := ListPromptsResult{
		Prompts: s.prompts,
	}

	s.logger.Infof("Listing %d prompts", len(s.prompts))

	return &JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result:  result,
	}
}

// handleGetPrompt handles the prompts/get request
func (s *Server) handleGetPrompt(req *JSONRPCRequest) *JSONRPCResponse {
	var getReq GetPromptRequest
	if err := json.Unmarshal(req.Params, &getReq); err != nil {
		return s.createErrorResponse(req.ID, -32602, "Invalid params", err.Error())
	}

	s.mu.RLock()
	promptData, exists := s.promptsMap[getReq.Name]
	s.mu.RUnlock()

	if !exists {
		return s.createErrorResponse(req.ID, -32602, fmt.Sprintf("Prompt not found: %s", getReq.Name), nil)
	}

	// Substitute variables in the prompt template
	systemPrompt := promptData.SystemPrompt
	userPrompt := promptData.UserPrompt

	// Replace {{variable}} with actual values from arguments
	for key, value := range getReq.Arguments {
		placeholder := fmt.Sprintf("{{%s}}", key)
		valueStr := fmt.Sprintf("%v", value)
		userPrompt = strings.ReplaceAll(userPrompt, placeholder, valueStr)
		systemPrompt = strings.ReplaceAll(systemPrompt, placeholder, valueStr)
	}

	result := GetPromptResult{
		Description: promptData.Description,
		Messages: []PromptMessage{
			{
				Role: "system",
				Content: ContentItem{
					Type: "text",
					Text: systemPrompt,
				},
			},
			{
				Role: "user",
				Content: ContentItem{
					Type: "text",
					Text: userPrompt,
				},
			},
		},
	}

	s.logger.Infof("Retrieved prompt: %s with %d arguments", getReq.Name, len(getReq.Arguments))

	return &JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result:  result,
	}
}

// Helper methods

func (s *Server) createErrorResponse(id interface{}, code int, message string, data interface{}) *JSONRPCResponse {
	return &JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      id,
		Error: &RPCError{
			Code:    code,
			Message: message,
			Data:    data,
		},
	}
}

func (s *Server) writeError(writer *bufio.Writer, id interface{}, code int, message string, data interface{}) {
	response := s.createErrorResponse(id, code, message, data)
	responseJSON, _ := json.Marshal(response)
	writer.Write(responseJSON)
	writer.Write([]byte("\n"))
	writer.Flush()
}
