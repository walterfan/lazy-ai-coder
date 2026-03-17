package models

// ErrorResponse represents a standard error response
type ErrorResponse struct {
	Error string `json:"error"`
}

// PromptRequest represents a prompt request with system, user, and optional assistant prompts
type PromptRequest struct {
	Name            string `json:"name"`
	SystemPrompt    string `json:"system_prompt"`
	UserPrompt      string `json:"user_prompt"`
	AssistantPrompt string `json:"assistant_prompt,omitempty"`
}

// Command represents a command with prompts and metadata
type Command struct {
	Name         string `json:"name"`
	Description  string `json:"description"`
	SystemPrompt string `json:"system_prompt"`
	UserPrompt   string `json:"user_prompt"`
	Tags         string `json:"tags"`
}

// CommandsResponse wraps a list of commands
type CommandsResponse struct {
	Commands []Command `json:"commands"`
}

// ProjectConfig represents project configuration
type ProjectConfig struct {
	Name     string `json:"name"`
	GitUrl   string `json:"gitUrl"`
	Project  string `json:"project"`
	Branch   string `json:"branch"`
	CodePath string `json:"codePath"`
	Language string `json:"language"`
}

// Settings represents LLM and external service settings
type Settings struct {
	LlmApiKey      string `json:"LLM_API_KEY"`
	LlmModel       string `json:"LLM_MODEL"`
	LlmBaseUrl     string `json:"LLM_BASE_URL"`
	LlmTemperature string `json:"LLM_TEMPERATURE"`
	GitlabUrl      string `json:"GITLAB_BASE_URL"`
	GitlabToken    string `json:"GITLAB_TOKEN"`
}

// ChatAPIRequest represents a chat request with all parameters
type ChatAPIRequest struct {
	Settings         Settings      `json:"settings"`
	Prompt           PromptRequest `json:"prompt"`
	ComputerLanguage string        `json:"computer_language,omitempty"`
	OutputLanguage   string        `json:"output_language,omitempty"`  // Keep for backward compatibility
	OutputLanguages  []string      `json:"output_languages,omitempty"` // New field for multiple languages
	Stream           bool          `json:"stream"`
	Remember         bool          `json:"remember"`
	SessionId        string        `json:"session_id"`
	CodePath         string        `json:"codePath"`

	GitlabBranch   string `json:"gitlab_code_branch"`
	GitlabProject  string `json:"gitlab_code_repo"`
	GitlabCodePath string `json:"gitlab_code_path"`
	GitlabMrId     string `json:"gitlab_mr_id"`
}

// ChatAPIResponse represents a chat response
type ChatAPIResponse struct {
	Answer string `json:"answer"`
}

// DrawAPIRequest represents a diagram drawing request
type DrawAPIRequest struct {
	PromptRequest
}

// DrawAPIResponse represents a diagram drawing response
type DrawAPIResponse struct {
	ImageUrl    string `json:"url"`
	ImagePath   string `json:"path"`
	ImageType   string `json:"type,omitempty"`   // "uml" or "mindmap"
	ImageScript string `json:"script,omitempty"` // PlantUML script if needed
}

// SmartPromptRequest represents a smart prompt generation request
type SmartPromptRequest struct {
	Input          string   `json:"input"`
	GitlabProject  string   `json:"gitlab_project"`
	GitlabBranch   string   `json:"gitlab_branch"`
	GitlabCodePath string   `json:"gitlab_code_path"`
	AnalyzeContext bool     `json:"analyze_context"`
	PresetID       string   `json:"preset_id"`
	Settings       Settings `json:"settings"`
}

// ProjectContext represents detected project context
type ProjectContext struct {
	Language         string   `json:"language"`
	Framework        string   `json:"framework"`
	FrameworkVersion string   `json:"framework_version"`
	BuildTool        string   `json:"build_tool"`
	Database         string   `json:"database"`
	HasTests         bool     `json:"has_tests"`
	TestFramework    string   `json:"test_framework"`
	Dependencies     []string `json:"dependencies"`
}

// CodeExample represents a code example
type CodeExample struct {
	Title       string `json:"title"`
	Language    string `json:"language"`
	Code        string `json:"code"`
	Description string `json:"description"`
}

// QualityScore represents prompt quality evaluation
type QualityScore struct {
	Score       float64  `json:"score"`
	MaxScore    float64  `json:"max_score"`
	Feedback    []string `json:"feedback"`
	Suggestions []string `json:"suggestions"`
}

// SmartPromptResponse represents a smart prompt generation response
type SmartPromptResponse struct {
	Context      string         `json:"context"`
	Action       string         `json:"action"`
	Result       string         `json:"result"`
	FullPrompt   string         `json:"full_prompt"`
	DetectedCtx  ProjectContext `json:"detected_context"`
	Examples     []CodeExample  `json:"examples"`
	QualityScore QualityScore   `json:"quality_score"`
}

// Preset represents a prompt generation preset
type Preset struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	Language     string `json:"language"`
	Framework    string `json:"framework"`
	ContextHints string `json:"context_hints"`
	ResultHints  string `json:"result_hints"`
}

// FrameworkExample represents an example for a prompt framework
type FrameworkExample struct {
	UseCase         string `json:"use_case"`
	Input           string `json:"input"`
	GeneratedPrompt string `json:"generated_prompt"`
}

