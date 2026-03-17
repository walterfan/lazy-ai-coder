package mcp

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/walterfan/lazy-ai-coder/internal/llm"
	"github.com/walterfan/lazy-ai-coder/internal/util"
)

// RegisterAllTools registers all available MCP tools
func (s *Server) RegisterAllTools() {
	// GitLab Tools - Keep
	s.RegisterTool(createGetGitLabFileContentTool(), s.handleGetGitLabFileContent)
	s.RegisterTool(createGetGitLabMergeRequestTool(), s.handleGetGitLabMergeRequest)
	s.RegisterTool(createGetGitLabProjectIDTool(), s.handleGetGitLabProjectID)
	s.RegisterTool(createGetGitLabMRSummaryTool(), s.handleGetGitLabMRSummary)
	s.RegisterTool(createPostGitLabMRCommentTool(), s.handlePostGitLabMRComment)

	// Code Review & Prompt Tools - New
	//s.RegisterTool(createCodeReviewTool(), s.handleCodeReview)
	s.RegisterTool(createBuildPromptTool(), s.handleBuildPrompt)
	s.RegisterTool(createQueryPromptTool(), s.handleQueryPrompt)
}

// GitLab Tool Definitions

func createGetGitLabFileContentTool() Tool {
	return Tool{
		Name:        "get_gitlab_file_content",
		Description: "Retrieve the content of a file from a GitLab repository",
		InputSchema: InputSchema{
			Type: "object",
			Properties: map[string]Property{
				"gitlab_url": {
					Type:        "string",
					Description: "GitLab base URL (e.g., https://gitlab.com)",
				},
				"project": {
					Type:        "string",
					Description: "GitLab project path (e.g., namespace/project-name) or numeric project ID",
				},
				"file_path": {
					Type:        "string",
					Description: "Path to the file in the repository",
				},
				"branch": {
					Type:        "string",
					Description: "Branch name (default: main)",
				},
			},
			Required: []string{"project", "file_path"},
		},
	}
}

func createGetGitLabMergeRequestTool() Tool {
	return Tool{
		Name:        "get_gitlab_merge_request",
		Description: "Retrieve the changes from a GitLab merge request",
		InputSchema: InputSchema{
			Type: "object",
			Properties: map[string]Property{
				"gitlab_url": {
					Type:        "string",
					Description: "GitLab base URL (e.g., https://gitlab.com)",
				},
				"project": {
					Type:        "string",
					Description: "GitLab project path (e.g., namespace/project-name) or numeric project ID",
				},
				"merge_request_id": {
					Type:        "string",
					Description: "Merge request IID (the number shown in the UI, not the internal ID)",
				},
			},
			Required: []string{"project", "merge_request_id"},
		},
	}
}

func createGetGitLabProjectIDTool() Tool {
	return Tool{
		Name:        "get_gitlab_project_id",
		Description: "Get the numeric project ID from a GitLab project path",
		InputSchema: InputSchema{
			Type: "object",
			Properties: map[string]Property{
				"gitlab_url": {
					Type:        "string",
					Description: "GitLab base URL (e.g., https://gitlab.com)",
				},
				"project_name": {
					Type:        "string",
					Description: "GitLab project path (e.g., namespace/project-name)",
				},
			},
			Required: []string{"project_name"},
		},
	}
}

func createGetGitLabMRSummaryTool() Tool {
	return Tool{
		Name:        "get_gitlab_mr_summary",
		Description: "Get a summary table of all file changes in a GitLab merge request with statistics (added/deleted/changed lines)",
		InputSchema: InputSchema{
			Type: "object",
			Properties: map[string]Property{
				"gitlab_url": {
					Type:        "string",
					Description: "GitLab base URL (e.g., https://gitlab.com)",
				},
				"project": {
					Type:        "string",
					Description: "GitLab project path (e.g., namespace/project-name) or numeric project ID",
				},
				"merge_request_id": {
					Type:        "string",
					Description: "Merge request ID (e.g., 123)",
				},
				"format": {
					Type:        "string",
					Description: "Output format: markdown (table) or json",
					Enum:        []string{"markdown", "json"},
				},
			},
			Required: []string{"project", "merge_request_id"},
		},
	}
}

func createPostGitLabMRCommentTool() Tool {
	return Tool{
		Name:        "post_gitlab_mr_comment",
		Description: "Post a comment to a GitLab merge request",
		InputSchema: InputSchema{
			Type: "object",
			Properties: map[string]Property{
				"gitlab_url": {
					Type:        "string",
					Description: "GitLab base URL (e.g., https://gitlab.com)",
				},
				"project": {
					Type:        "string",
					Description: "GitLab project path (e.g., namespace/project-name) or numeric project ID",
				},
				"merge_request_id": {
					Type:        "string",
					Description: "Merge request ID (e.g., 123)",
				},
				"comment": {
					Type:        "string",
					Description: "Comment text to post (supports Markdown formatting)",
				},
			},
			Required: []string{"project", "merge_request_id", "comment"},
		},
	}
}

// Code Review Tool Definition

func createCodeReviewTool() Tool {
	return Tool{
		Name:        "code_review",
		Description: "Perform comprehensive code review with analysis of quality, security, performance, and best practices",
		InputSchema: InputSchema{
			Type: "object",
			Properties: map[string]Property{
				"code": {
					Type:        "string",
					Description: "Code content to review",
				},
				"language": {
					Type:        "string",
					Description: "Programming language (e.g., go, java, python, javascript)",
				},
				"focus": {
					Type:        "string",
					Description: "Review focus area: all, security, performance, quality, or style",
					Enum:        []string{"all", "security", "performance", "quality", "style"},
				},
			},
			Required: []string{"code"},
		},
	}
}

// Prompt Building Tool Definition

func createBuildPromptTool() Tool {
	return Tool{
		Name:        "build_prompt",
		Description: "Build and enhance a prompt using LLM to make it more effective and well-structured",
		InputSchema: InputSchema{
			Type: "object",
			Properties: map[string]Property{
				"input_text": {
					Type:        "string",
					Description: "The input text or rough prompt idea to be enhanced",
				},
				"purpose": {
					Type:        "string",
					Description: "Purpose of the prompt: code_generation, analysis, documentation, review, or general",
				},
			},
			Required: []string{"input_text"},
		},
	}
}

// Query Prompt Tool Definition

func createQueryPromptTool() Tool {
	return Tool{
		Name:        "query_prompt",
		Description: "Query and retrieve saved prompts by keyword, category, or ID",
		InputSchema: InputSchema{
			Type: "object",
			Properties: map[string]Property{
				"query": {
					Type:        "string",
					Description: "Search keyword or prompt ID",
				},
				"category": {
					Type:        "string",
					Description: "Prompt category: code_review, code_generation, documentation, analysis, or all",
					Enum:        []string{"all", "code_review", "code_generation", "documentation", "analysis"},
				},
			},
			Required: []string{"query"},
		},
	}
}

// Tool Handler Implementations

func (s *Server) handleGetGitLabFileContent(args map[string]interface{}) (*CallToolResult, error) {
	// Extract parameters - credentials must be passed from frontend
	gitlabURL := getStringArg(args, "gitlab_url", "")
	project := getStringArg(args, "project", "")
	filePath := getStringArg(args, "file_path", "")
	branch := getStringArg(args, "branch", "main")
	privateToken := getStringArg(args, "gitlab_token", "")

	if project == "" || filePath == "" {
		return createErrorResult("Missing required parameters: project and file_path"), nil
	}

	if gitlabURL == "" {
		return createErrorResult("GitLab URL not configured. Please configure it in Settings page"), nil
	}

	if privateToken == "" {
		return createErrorResult("GitLab token not configured. Please configure it in Settings page"), nil
	}

	s.logger.Infof("Fetching GitLab file: project=%s, path=%s, branch=%s", project, filePath, branch)

	// Call GitLab API
	content, err := util.GetGitLabFileContent(gitlabURL, project, filePath, branch, privateToken)
	if err != nil {
		return createErrorResult(fmt.Sprintf("Failed to fetch GitLab file: %v", err)), nil
	}

	return &CallToolResult{
		Content: []ContentItem{
			{
				Type: "text",
				Text: fmt.Sprintf("# File: %s (branch: %s)\n\n```\n%s\n```", filePath, branch, content),
			},
		},
	}, nil
}

func (s *Server) handleGetGitLabMergeRequest(args map[string]interface{}) (*CallToolResult, error) {
	// Extract parameters - credentials must be passed from frontend
	gitlabURL := getStringArg(args, "gitlab_url", "")
	project := getStringArg(args, "project", "")
	mrID := getStringArg(args, "merge_request_id", "")
	privateToken := getStringArg(args, "gitlab_token", "")

	if project == "" || mrID == "" {
		return createErrorResult("Missing required parameters: project and merge_request_id"), nil
	}

	if gitlabURL == "" {
		return createErrorResult("GitLab URL not configured. Please configure it in Settings page"), nil
	}

	if privateToken == "" {
		return createErrorResult("GitLab token not configured. Please configure it in Settings page"), nil
	}

	if gitlabURL == "" {
		return createErrorResult("GitLab URL not configured. Set GITLAB_BASE_URL environment variable or provide gitlab_url parameter"), nil
	}

	if privateToken == "" {
		return createErrorResult("GitLab token not configured. Set GITLAB_TOKEN environment variable"), nil
	}

	s.logger.Infof("Fetching GitLab MR: project=%s, mr_id=%s", project, mrID)

	// Call GitLab API
	changes, err := util.GetMergeRequestChange(gitlabURL, project, mrID, privateToken)
	if err != nil {
		return createErrorResult(fmt.Sprintf("Failed to fetch GitLab merge request: %v", err)), nil
	}

	return &CallToolResult{
		Content: []ContentItem{
			{
				Type: "text",
				Text: changes,
			},
		},
	}, nil
}

func (s *Server) handleGetGitLabProjectID(args map[string]interface{}) (*CallToolResult, error) {
	// Extract parameters - credentials must be passed from frontend
	gitlabURL := getStringArg(args, "gitlab_url", "")
	projectName := getStringArg(args, "project_name", "")
	privateToken := getStringArg(args, "gitlab_token", "")

	if projectName == "" {
		return createErrorResult("Missing required parameter: project_name"), nil
	}

	if gitlabURL == "" {
		return createErrorResult("GitLab URL not configured. Please configure it in Settings page"), nil
	}

	if privateToken == "" {
		return createErrorResult("GitLab token not configured. Please configure it in Settings page"), nil
	}

	if gitlabURL == "" {
		return createErrorResult("GitLab URL not configured. Set GITLAB_BASE_URL environment variable or provide gitlab_url parameter"), nil
	}

	if privateToken == "" {
		return createErrorResult("GitLab token not configured. Set GITLAB_TOKEN environment variable"), nil
	}

	s.logger.Infof("Fetching GitLab project ID: project_name=%s", projectName)

	// Call GitLab API
	projectID, err := util.GetProjectIDByName(gitlabURL, projectName, privateToken)
	if err != nil {
		return createErrorResult(fmt.Sprintf("Failed to fetch GitLab project ID: %v", err)), nil
	}

	return &CallToolResult{
		Content: []ContentItem{
			{
				Type: "text",
				Text: fmt.Sprintf("Project '%s' has ID: %s", projectName, projectID),
			},
		},
	}, nil
}

func (s *Server) handleGetGitLabMRSummary(args map[string]interface{}) (*CallToolResult, error) {
	// Extract parameters - credentials must be passed from frontend
	gitlabURL := getStringArg(args, "gitlab_url", "")
	project := getStringArg(args, "project", "")
	mrID := getStringArg(args, "merge_request_id", "")
	format := getStringArg(args, "format", "markdown")
	privateToken := getStringArg(args, "gitlab_token", "")

	if project == "" || mrID == "" {
		return createErrorResult("Missing required parameters: project and merge_request_id"), nil
	}

	if gitlabURL == "" {
		return createErrorResult("GitLab URL not configured. Please configure it in Settings page"), nil
	}

	if privateToken == "" {
		return createErrorResult("GitLab token not configured. Please configure it in Settings page"), nil
	}

	s.logger.Infof("Fetching GitLab MR summary: project=%s, mr_id=%s, format=%s", project, mrID, format)

	// Get the raw MR data
	mrInfo, err := util.GetMergeRequestInfo(gitlabURL, project, mrID, privateToken)
	if err != nil {
		return createErrorResult(fmt.Sprintf("Failed to fetch merge request info: %v", err)), nil
	}

	// Convert to summary
	summary := util.ConvertToMRSummary(mrInfo)

	var output string
	if format == "json" {
		jsonData, err := json.MarshalIndent(summary, "", "  ")
		if err != nil {
			return createErrorResult(fmt.Sprintf("Failed to marshal JSON: %v", err)), nil
		}
		output = string(jsonData)
	} else {
		output = util.MRSummaryToMarkdownTable(summary)
	}

	return &CallToolResult{
		Content: []ContentItem{
			{
				Type: "text",
				Text: output,
			},
		},
	}, nil
}

func (s *Server) handlePostGitLabMRComment(args map[string]interface{}) (*CallToolResult, error) {
	// Extract parameters - credentials must be passed from frontend
	gitlabURL := getStringArg(args, "gitlab_url", "")
	project := getStringArg(args, "project", "")
	mrID := getStringArg(args, "merge_request_id", "")
	comment := getStringArg(args, "comment", "")
	privateToken := getStringArg(args, "gitlab_token", "")

	if project == "" || mrID == "" || comment == "" {
		return createErrorResult("Missing required parameters: project, merge_request_id, and comment"), nil
	}

	if gitlabURL == "" {
		return createErrorResult("GitLab URL not configured. Please configure it in Settings page"), nil
	}

	if privateToken == "" {
		return createErrorResult("GitLab token not configured. Please configure it in Settings page"), nil
	}

	s.logger.Infof("Posting comment to GitLab MR: project=%s, mr_id=%s", project, mrID)

	// Post the comment
	err := util.PostMergeRequestComment(gitlabURL, project, mrID, comment, privateToken)
	if err != nil {
		return createErrorResult(fmt.Sprintf("Failed to post comment: %v", err)), nil
	}

	return &CallToolResult{
		Content: []ContentItem{
			{
				Type: "text",
				Text: fmt.Sprintf("Comment posted successfully to MR #%s in project %s", mrID, project),
			},
		},
	}, nil
}

func (s *Server) handleCodeReview(args map[string]interface{}) (*CallToolResult, error) {
	// Extract parameters
	code := getStringArg(args, "code", "")
	language := getStringArg(args, "language", "")
	focus := getStringArg(args, "focus", "all")

	if code == "" {
		return createErrorResult("Missing required parameter: code"), nil
	}

	// Build system prompt based on focus area
	var systemPrompt string
	switch focus {
	case "security":
		systemPrompt = "You are a security expert. Perform a comprehensive security review of the code. Identify vulnerabilities (SQL injection, XSS, auth issues, data exposure), suggest fixes, and rate severity (Critical/High/Medium/Low)."
	case "performance":
		systemPrompt = "You are a performance optimization expert. Analyze the code for performance issues (inefficient algorithms, memory leaks, unnecessary operations, database N+1 queries). Suggest optimizations with expected impact."
	case "quality":
		systemPrompt = "You are a code quality expert. Review the code for maintainability, readability, design patterns, SOLID principles, and technical debt. Suggest improvements."
	case "style":
		systemPrompt = "You are a code style expert. Review the code for naming conventions, formatting, documentation, and adherence to language-specific best practices."
	default: // "all"
		systemPrompt = "You are an expert code reviewer. Perform a comprehensive code review covering: 1) Security vulnerabilities, 2) Performance issues, 3) Code quality and maintainability, 4) Style and best practices. Provide specific, actionable feedback with examples."
	}

	// Build user prompt
	userPrompt := fmt.Sprintf("Review this %s code:\n\n```%s\n%s\n```\n\nProvide detailed feedback with specific line references where applicable.", language, language, code)

	// Build LLM settings - use args if provided, fallback to env variables
	settings := llm.LLMSettings{
		BaseUrl:     getStringArg(args, "llm_base_url", os.Getenv("LLM_BASE_URL")),
		ApiKey:      getStringArg(args, "llm_api_key", os.Getenv("LLM_API_KEY")),
		Model:       getStringArg(args, "llm_model", os.Getenv("LLM_MODEL")),
		Temperature: 0.3, // Lower temperature for more focused, consistent reviews
	}

	s.logger.Infof("Performing code review: language=%s, focus=%s, code_length=%d", language, focus, len(code))

	// Call LLM
	response, err := llm.AskLLM(systemPrompt, userPrompt, settings)
	if err != nil {
		return createErrorResult(fmt.Sprintf("Code review failed: %v", err)), nil
	}

	return &CallToolResult{
		Content: []ContentItem{
			{
				Type: "text",
				Text: response,
			},
		},
	}, nil
}

func (s *Server) handleBuildPrompt(args map[string]interface{}) (*CallToolResult, error) {
	// Extract parameters
	inputText := getStringArg(args, "input_text", "")
	purpose := getStringArg(args, "purpose", "general")

	if inputText == "" {
		return createErrorResult("Missing required parameter: input_text"), nil
	}

	// Build system prompt for prompt enhancement
	systemPrompt := `You are an expert prompt engineer. Transform rough prompt ideas into well-structured, effective prompts that get better LLM results.

Guidelines:
1. Clarity: Make the intent crystal clear
2. Context: Add necessary background information
3. Structure: Use sections (task, constraints, format, examples)
4. Specificity: Define expected output format precisely
5. Examples: Include examples when helpful
6. Constraints: State any limitations or requirements

Return ONLY the enhanced prompt without meta-commentary.`

	// Customize based on purpose
	var purposeGuidance string
	switch purpose {
	case "code_generation":
		purposeGuidance = "\n\nFocus: This prompt is for code generation. Include language, framework, requirements, edge cases, and expected output format."
	case "analysis":
		purposeGuidance = "\n\nFocus: This prompt is for analysis. Specify what to analyze, analysis criteria, and expected insights format."
	case "documentation":
		purposeGuidance = "\n\nFocus: This prompt is for documentation. Specify documentation type, audience, structure, and level of detail."
	case "review":
		purposeGuidance = "\n\nFocus: This prompt is for code/content review. Specify review criteria, focus areas, and feedback format."
	default:
		purposeGuidance = ""
	}

	// Build user prompt
	userPrompt := fmt.Sprintf("Enhance this prompt:%s\n\nInput prompt:\n%s", purposeGuidance, inputText)

	// Build LLM settings - use args if provided, fallback to env variables
	settings := llm.LLMSettings{
		BaseUrl:     getStringArg(args, "llm_base_url", os.Getenv("LLM_BASE_URL")),
		ApiKey:      getStringArg(args, "llm_api_key", os.Getenv("LLM_API_KEY")),
		Model:       getStringArg(args, "llm_model", os.Getenv("LLM_MODEL")),
		Temperature: 0.7,
	}

	s.logger.Infof("Building enhanced prompt: purpose=%s, input_length=%d", purpose, len(inputText))

	// Call LLM
	enhancedPrompt, err := llm.AskLLM(systemPrompt, userPrompt, settings)
	if err != nil {
		return createErrorResult(fmt.Sprintf("Failed to build prompt: %v", err)), nil
	}

	return &CallToolResult{
		Content: []ContentItem{
			{
				Type: "text",
				Text: enhancedPrompt,
			},
		},
	}, nil
}

func (s *Server) handleQueryPrompt(args map[string]interface{}) (*CallToolResult, error) {
	// Extract parameters
	query := getStringArg(args, "query", "")
	category := getStringArg(args, "category", "all")

	if query == "" {
		return createErrorResult("Missing required parameter: query"), nil
	}

	s.logger.Infof("Querying prompts: query=%s, category=%s", query, category)

	// Build predefined prompt library
	prompts := map[string]map[string]string{
		"code_review": {
			"security":    "Perform a comprehensive security review of the code. Identify vulnerabilities including SQL injection, XSS, authentication issues, data exposure, insecure dependencies. For each issue found, provide: 1) Severity rating (Critical/High/Medium/Low), 2) Affected code lines, 3) Explanation of the vulnerability, 4) Recommended fix with code example.",
			"performance": "Analyze the code for performance bottlenecks and optimization opportunities. Check for: 1) Inefficient algorithms (O(n²) where O(n) possible), 2) Memory leaks or excessive allocations, 3) Database N+1 queries, 4) Unnecessary API calls, 5) Missing indexes or caching. Prioritize by impact.",
			"general":     "Conduct a thorough code review evaluating: 1) Security vulnerabilities, 2) Performance issues, 3) Code quality (maintainability, readability, SOLID principles), 4) Style and conventions. Provide specific, actionable feedback with code examples.",
		},
		"code_generation": {
			"api":      "Generate a RESTful API endpoint with the following requirements: [SPECIFY REQUIREMENTS]. Include: 1) Route definition, 2) Request/response DTOs, 3) Validation logic, 4) Error handling, 5) Unit tests. Follow [LANGUAGE/FRAMEWORK] best practices.",
			"function": "Write a function that [DESCRIBE FUNCTIONALITY]. Requirements: 1) Input parameters: [LIST], 2) Output: [DESCRIBE], 3) Handle edge cases: [LIST], 4) Include error handling, 5) Add comprehensive docstrings/comments, 6) Include unit tests.",
		},
		"documentation": {
			"readme": "Create a README.md for this project covering: 1) Project overview and purpose, 2) Installation instructions, 3) Usage examples with code, 4) Configuration options, 5) API documentation, 6) Contributing guidelines, 7) License information.",
			"api":    "Document this API endpoint: [ENDPOINT]. Include: 1) Purpose and use case, 2) HTTP method and path, 3) Request parameters (query, path, body), 4) Response format with examples, 5) Error codes and handling, 6) Authentication requirements, 7) Rate limits.",
		},
		"analysis": {
			"architecture": "Analyze the system architecture. Evaluate: 1) Component organization and responsibilities, 2) Data flow and dependencies, 3) Scalability considerations, 4) Security architecture, 5) Performance bottlenecks, 6) Improvement recommendations. Provide diagrams where helpful.",
			"complexity":   "Analyze code complexity and technical debt. Report on: 1) Cyclomatic complexity, 2) Code duplication, 3) Long methods/classes, 4) Deep nesting, 5) Coupling issues, 6) Refactoring opportunities with priority.",
		},
	}

	// Search through prompts
	var results []string
	searchCategory := category
	if category == "all" {
		for cat := range prompts {
			for key, prompt := range prompts[cat] {
				if contains(key, query) || contains(prompt, query) {
					results = append(results, fmt.Sprintf("**[%s/%s]**\n%s\n", cat, key, prompt))
				}
			}
		}
	} else {
		if catPrompts, ok := prompts[searchCategory]; ok {
			for key, prompt := range catPrompts {
				if contains(key, query) || contains(prompt, query) {
					results = append(results, fmt.Sprintf("**[%s/%s]**\n%s\n", searchCategory, key, prompt))
				}
			}
		}
	}

	if len(results) == 0 {
		return &CallToolResult{
			Content: []ContentItem{
				{
					Type: "text",
					Text: fmt.Sprintf("No prompts found matching query '%s' in category '%s'.\n\nAvailable categories: code_review, code_generation, documentation, analysis", query, category),
				},
			},
		}, nil
	}

	responseText := fmt.Sprintf("Found %d prompt(s):\n\n%s", len(results), joinStrings(results, "\n---\n\n"))

	return &CallToolResult{
		Content: []ContentItem{
			{
				Type: "text",
				Text: responseText,
			},
		},
	}, nil
}

// Helper Functions

// Helper function for case-insensitive string search
func contains(str, substr string) bool {
	return len(substr) > 0 && (str == substr ||
		len(str) >= len(substr) &&
			findSubstring(str, substr))
}

func findSubstring(str, substr string) bool {
	str = toLower(str)
	substr = toLower(substr)
	for i := 0; i <= len(str)-len(substr); i++ {
		if str[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func toLower(s string) string {
	result := make([]byte, len(s))
	for i := 0; i < len(s); i++ {
		if s[i] >= 'A' && s[i] <= 'Z' {
			result[i] = s[i] + 32
		} else {
			result[i] = s[i]
		}
	}
	return string(result)
}

func joinStrings(strs []string, sep string) string {
	if len(strs) == 0 {
		return ""
	}
	result := strs[0]
	for i := 1; i < len(strs); i++ {
		result += sep + strs[i]
	}
	return result
}

func getStringArg(args map[string]interface{}, key string, defaultValue string) string {
	if val, ok := args[key]; ok {
		if strVal, ok := val.(string); ok {
			return strVal
		}
	}
	return defaultValue
}

func createErrorResult(message string) *CallToolResult {
	return &CallToolResult{
		Content: []ContentItem{
			{
				Type: "text",
				Text: fmt.Sprintf("Error: %s", message),
			},
		},
		IsError: true,
	}
}

// MarshalJSON for CallToolResult to ensure proper JSON encoding
func (r *CallToolResult) MarshalJSON() ([]byte, error) {
	type Alias CallToolResult
	return json.Marshal(&struct {
		*Alias
	}{
		Alias: (*Alias)(r),
	})
}
