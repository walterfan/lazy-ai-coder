package cmd

import (
	"fmt"
	"os"
	"sort"

	"github.com/spf13/viper"
	"github.com/walterfan/lazy-ai-coder/internal/models"
)

type GitRepoConfig struct {
	GitlabUrl      string `json:"gitUrl"`
	GitlabProject  string `json:"project"`
	GitlabBranch   string `json:"branch"`
	GitlabCodePath string `json:"codePath"`
	Language       string `json:"language"`
}

var (
	port              string
	cachedPromptMap   map[string]interface{}
	cachedPromptList  []string // new field to preserve order
	configInitialized = false
	codeRepoConfig    map[string]GitRepoConfig
)

// loadPromptsFromFile loads prompts from the external prompts.yaml file
func loadPromptsFromFile() error {
	// Get the prompts file path from config
	promptsFile := viper.GetString("prompts_file")
	if promptsFile == "" {
		promptsFile = "config/prompts.yaml" // Default path
	}

	// Create a new viper instance for prompts
	promptViper := viper.New()
	promptViper.SetConfigFile(promptsFile)
	promptViper.SetConfigType("yaml")

	if err := promptViper.ReadInConfig(); err != nil {
		return fmt.Errorf("error reading prompts file %s: %v", promptsFile, err)
	}

	// Load prompts map
	cachedPromptMap = promptViper.GetStringMap("prompts")
	if cachedPromptMap == nil {
		return fmt.Errorf("prompts configuration is missing in %s", promptsFile)
	}

	// Get the prompts node to preserve order
	promptNode := promptViper.Get("prompts")
	if promptNode == nil {
		return fmt.Errorf("prompts configuration is missing in %s", promptsFile)
	}

	m, ok := promptNode.(map[string]interface{})
	if !ok {
		return fmt.Errorf("prompts must be a map[string]interface{} in %s", promptsFile)
	}

	// Build the prompt list preserving order
	cachedPromptList = make([]string, 0, len(m))
	for k := range m {
		cachedPromptList = append(cachedPromptList, k)
	}

	// Sort alphabetically since we removed sequence numbers
	sort.Strings(cachedPromptList)

	return nil
}

// getStringValue safely extracts a string value from a map, returning empty string if not found or nil
func getStringValue(data map[string]interface{}, key string) string {
	if val, exists := data[key]; exists && val != nil {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return ""
}

func getPromptConfigByName(name string) (*models.PromptRequest, error) {
	item, ok := cachedPromptMap[name]
	if !ok {
		return nil, fmt.Errorf("prompt not found: %s", name)
	}

	itemMap, ok := item.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid prompt format for: %s", name)
	}

	return &models.PromptRequest{
		Name:         name,
		SystemPrompt: itemMap["system_prompt"].(string),
		UserPrompt:   itemMap["user_prompt"].(string),
	}, nil
}

// InitConfig initializes viper configuration only once
func InitConfig() error {
	if configInitialized {
		return nil
	}

	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./config")

	if err := viper.ReadInConfig(); err != nil {
		return fmt.Errorf("error reading config file: %v", err)
	}

	// Load prompts from external file
	if err := loadPromptsFromFile(); err != nil {
		return fmt.Errorf("error loading prompts: %v", err)
	}

	projectsMap := viper.GetStringMap("projects")
	if projectsMap == nil {
		panic("gitlab configuration is missing")
	}

	codeRepoConfig = make(map[string]GitRepoConfig, len(projectsMap))
	for k, v := range projectsMap {

		data, ok := v.(map[string]interface{})
		if !ok {
			fmt.Fprintf(os.Stderr, "Unexpected type for project '%s': %T, value: %+v\n", k, v, v)
			panic(fmt.Sprintf("project '%s' must be a map[string]string in config", k))
		}

		gitRepoConfig := GitRepoConfig{
			GitlabUrl:      getStringValue(data, "giturl"),
			GitlabProject:  getStringValue(data, "project"),
			GitlabBranch:   getStringValue(data, "branch"),
			GitlabCodePath: getStringValue(data, "codepath"),
			Language:       getStringValue(data, "language"),
		}
		codeRepoConfig[k] = gitRepoConfig
	}

	configInitialized = true
	return nil
}
