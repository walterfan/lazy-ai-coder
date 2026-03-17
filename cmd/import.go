package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/google/uuid"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
	"github.com/walterfan/lazy-ai-coder/pkg/database"
	"github.com/walterfan/lazy-ai-coder/pkg/models"
)

var (
	realmID   string
	createdBy string
)

var importCmd = &cobra.Command{
	Use:   "import",
	Short: "Import data from config files",
}

var importProjectsCmd = &cobra.Command{
	Use:   "projects",
	Short: "Import projects from config.yaml to database",
	Long: `Import all projects defined in config/config.yaml into the SQLite database.
	
This command reads the projects section from config.yaml and creates/updates
corresponding records in the database. Existing projects with the same name
will be updated.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Initialize config
		if err := InitConfig(); err != nil {
			fmt.Printf("Failed to initialize config: %v\n", err)
			return
		}

		// Initialize database
		if err := database.InitDB(); err != nil {
			fmt.Printf("Failed to initialize database: %v\n", err)
			fmt.Println("Hint: Make sure no other processes (like the web server) are using the database")
			return
		}
		defer database.CloseDB()

		db := database.GetDB()

		// Get projects from config
		projectsMap := viper.GetStringMap("projects")
		if len(projectsMap) == 0 {
			fmt.Println("No projects found in config.yaml")
			return
		}

		fmt.Printf("Found %d projects in config.yaml\n", len(projectsMap))

		imported := 0
		updated := 0
		errors := 0

		for name, projectData := range projectsMap {
			projectMap, ok := projectData.(map[string]interface{})
			if !ok {
				fmt.Printf("❌ Invalid project format for '%s'\n", name)
				errors++
				continue
			}

			// Check if project already exists
			var existingProject models.Project
			result := db.Where("name = ? AND realm_id = ?", name, realmID).First(&existingProject)

			// Note: Viper converts all keys to lowercase
			gitUrl := getStringValue(projectMap, "giturl")
			gitRepo := getStringValue(projectMap, "project")
			gitBranch := getStringValue(projectMap, "branch")
			entryPoint := getStringValue(projectMap, "codepath")
			language := getStringValue(projectMap, "language")

			project := models.Project{
				Name:        name,
				RealmID:     realmID,
				GitURL:      gitUrl,
				GitRepo:     gitRepo,
				GitBranch:   gitBranch,
				EntryPoint:  entryPoint,
				Language:    language,
				Description: fmt.Sprintf("Imported from config.yaml"),
				CreatedBy:   createdBy,
				UpdatedBy:   createdBy,
			}

			if result.Error == nil {
				// Update existing project
				project.ID = existingProject.ID
				project.UserID = existingProject.UserID
				if err := db.Save(&project).Error; err != nil {
					fmt.Printf("❌ Failed to update project '%s': %v\n", name, err)
					errors++
					continue
				}
				fmt.Printf("✅ Updated project: %s\n", name)
				updated++
			} else {
				// Create new project
				project.ID = uuid.New().String()
				if err := db.Create(&project).Error; err != nil {
					fmt.Printf("❌ Failed to create project '%s': %v\n", name, err)
					errors++
					continue
				}
				fmt.Printf("✅ Imported project: %s\n", name)
				imported++
			}
		}

		fmt.Printf("\n📊 Summary:\n")
		fmt.Printf("  - Imported: %d\n", imported)
		fmt.Printf("  - Updated:  %d\n", updated)
		fmt.Printf("  - Errors:   %d\n", errors)
		fmt.Printf("  - Total:    %d\n", len(projectsMap))
	},
}

var (
	promptsFile string
	updateMode  bool
	dryRun      bool
)

// PromptYAML represents the structure of prompts.yaml
type PromptYAML struct {
	Prompts map[string]PromptData `yaml:"prompts"`
}

// PromptData represents individual prompt configuration
type PromptData struct {
	Title        string                 `yaml:"title"`
	Description  string                 `yaml:"description"`
	SystemPrompt string                 `yaml:"system_prompt"`
	UserPrompt   string                 `yaml:"user_prompt"`
	Arguments    []models.PromptArgument `yaml:"arguments"`
	Tags         string                 `yaml:"tags"`
}

var importPromptsCmd = &cobra.Command{
	Use:   "prompts",
	Short: "Import prompts from YAML file to database",
	Long: `Import prompts from prompts.yaml into the SQLite database.

This command reads the prompts section from the YAML file and creates/updates
corresponding records in the database.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Initialize config
		if err := InitConfig(); err != nil {
			fmt.Printf("Failed to initialize config: %v\n", err)
			return
		}

		// Initialize database
		if err := database.InitDB(); err != nil {
			fmt.Printf("Failed to initialize database: %v\n", err)
			fmt.Println("Hint: Make sure no other processes (like the web server) are using the database")
			return
		}
		defer database.CloseDB()

		// Read YAML file
		data, err := os.ReadFile(promptsFile)
		if err != nil {
			fmt.Printf("❌ Failed to read file '%s': %v\n", promptsFile, err)
			return
		}

		// Parse YAML
		var promptsYAML PromptYAML
		if err := yaml.Unmarshal(data, &promptsYAML); err != nil {
			fmt.Printf("❌ Failed to parse YAML: %v\n", err)
			return
		}

		if len(promptsYAML.Prompts) == 0 {
			fmt.Println("No prompts found in YAML file")
			return
		}

		fmt.Printf("Found %d prompts in %s\n", len(promptsYAML.Prompts), promptsFile)

		if dryRun {
			fmt.Println("\n🔍 DRY-RUN MODE - No changes will be made\n")
		}

		db := database.GetDB()
		created := 0
		updated := 0
		skipped := 0
		errors := 0

		for name, promptData := range promptsYAML.Prompts {
			// Check if prompt already exists
			var existingPrompt models.Prompt
			result := db.Where("name = ?", name).First(&existingPrompt)

			if dryRun {
				if result.Error == nil {
					if updateMode {
						fmt.Printf("🔍 [DRY-RUN] Would update: %s\n", name)
					} else {
						fmt.Printf("🔍 [DRY-RUN] Would skip (exists): %s\n", name)
					}
				} else {
					fmt.Printf("🔍 [DRY-RUN] Would create: %s\n", name)
				}
				fmt.Printf("   Description: %s\n", promptData.Description)
				fmt.Printf("   Tags: %s\n\n", promptData.Tags)
				continue
			}

			// Generate title from name if not provided (backward compatibility)
			title := promptData.Title
			if title == "" {
				// Convert snake_case to Title Case
				title = strings.ReplaceAll(name, "_", " ")
				title = strings.Title(title)
			}

			// Handle arguments - use explicit arguments or extract from template
			var argumentsJSON string
			if len(promptData.Arguments) > 0 {
				// Use explicit arguments from YAML
				argsBytes, err := json.Marshal(promptData.Arguments)
				if err != nil {
					fmt.Printf("⚠️  Warning: Failed to marshal arguments for '%s': %v\n", name, err)
				} else {
					argumentsJSON = string(argsBytes)
				}
			} else {
				// Extract arguments from template for backward compatibility
				extractedArgs := extractArgumentsFromTemplate(promptData.UserPrompt + " " + promptData.SystemPrompt)
				if len(extractedArgs) > 0 {
					argsBytes, err := json.Marshal(extractedArgs)
					if err != nil {
						fmt.Printf("⚠️  Warning: Failed to marshal extracted arguments for '%s': %v\n", name, err)
					} else {
						argumentsJSON = string(argsBytes)
					}
				}
			}

			prompt := models.Prompt{
				Name:         name,
				Title:        title,
				Description:  promptData.Description,
				SystemPrompt: promptData.SystemPrompt,
				UserPrompt:   promptData.UserPrompt,
				Arguments:    argumentsJSON,
				Tags:         promptData.Tags,
				CreatedBy:    createdBy,
				UpdatedBy:    createdBy,
			}

			if result.Error == nil {
				// Prompt exists
				if updateMode {
					// Update existing prompt
					prompt.ID = existingPrompt.ID
					prompt.UserID = existingPrompt.UserID
					prompt.RealmID = existingPrompt.RealmID
					if err := db.Save(&prompt).Error; err != nil {
						fmt.Printf("❌ Failed to update prompt '%s': %v\n", name, err)
						errors++
						continue
					}
					fmt.Printf("✅ Updated: %s\n", name)
					updated++
				} else {
					// Skip existing prompt
					fmt.Printf("⏭️  Skipped (exists): %s\n", name)
					skipped++
				}
			} else {
				// Create new prompt
				prompt.ID = uuid.New().String()
				if err := db.Create(&prompt).Error; err != nil {
					fmt.Printf("❌ Failed to create prompt '%s': %v\n", name, err)
					errors++
					continue
				}
				fmt.Printf("✅ Created: %s\n", name)
				created++
			}
		}

		if dryRun {
			fmt.Println("\n🔍 DRY-RUN MODE - No changes were made")
		} else {
			fmt.Printf("\n📊 Import Summary:\n")
			fmt.Printf("   Total:      %d\n", len(promptsYAML.Prompts))
			fmt.Printf("   ✅ Created:  %d\n", created)
			fmt.Printf("   ✅ Updated:  %d\n", updated)
			fmt.Printf("   ⏭️  Skipped:  %d\n", skipped)
			fmt.Printf("   ❌ Errors:   %d\n", errors)
		}
	},
}

// CursorRulesYAML represents the structure of cursor_rules.yaml
type CursorRulesYAML struct {
	CursorRules map[string]struct {
		Description string `yaml:"description"`
		Content     string `yaml:"content"`
		Language    string `yaml:"language"`
		Framework   string `yaml:"framework"`
		Tags        string `yaml:"tags"`
		IsTemplate  bool   `yaml:"is_template"`
	} `yaml:"cursor_rules"`
}

var importCursorRulesCmd = &cobra.Command{
	Use:   "cursor-rules",
	Short: "Import cursor rules from YAML file to database",
	Long:  `Import cursor rules from cursor_rules.yaml into the database as global templates.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Initialize config
		if err := InitConfig(); err != nil {
			fmt.Printf("Failed to initialize config: %v\n", err)
			return
		}

		// Initialize database
		if err := database.InitDB(); err != nil {
			fmt.Printf("Failed to initialize database: %v\n", err)
			return
		}
		defer database.CloseDB()

		// Read YAML file
		rulesFile := "config/cursor_rules.yaml"
		data, err := os.ReadFile(rulesFile)
		if err != nil {
			fmt.Printf("❌ Failed to read file '%s': %v\n", rulesFile, err)
			return
		}

		// Parse YAML
		var rulesYAML CursorRulesYAML
		if err := yaml.Unmarshal(data, &rulesYAML); err != nil {
			fmt.Printf("❌ Failed to parse YAML: %v\n", err)
			return
		}

		if len(rulesYAML.CursorRules) == 0 {
			fmt.Println("No cursor rules found in YAML file")
			return
		}

		fmt.Printf("Found %d cursor rules in %s\n", len(rulesYAML.CursorRules), rulesFile)

		db := database.GetDB()
		created := 0
		updated := 0
		skipped := 0
		errors := 0

		for name, ruleData := range rulesYAML.CursorRules {
			// Check if rule already exists
			var existingRule models.CursorRule
			result := db.Where("name = ?", name).First(&existingRule)

			rule := models.CursorRule{
				Name:        name,
				Description: ruleData.Description,
				Content:     ruleData.Content,
				Language:    ruleData.Language,
				Framework:   ruleData.Framework,
				Tags:        ruleData.Tags,
				IsTemplate:  ruleData.IsTemplate,
				UsageCount:  0,
				CreatedBy:   createdBy,
				UpdatedBy:   createdBy,
			}

			if result.Error == nil {
				// Rule exists
				if updateMode {
					// Update existing rule
					rule.ID = existingRule.ID
					rule.UserID = existingRule.UserID
					rule.RealmID = existingRule.RealmID
					if err := db.Save(&rule).Error; err != nil {
						fmt.Printf("❌ Failed to update rule '%s': %v\n", name, err)
						errors++
						continue
					}
					fmt.Printf("✅ Updated: %s\n", name)
					updated++
				} else {
					// Skip existing rule
					fmt.Printf("⏭️  Skipped (exists): %s\n", name)
					skipped++
				}
			} else {
				// Create new rule (as global template: user_id=NULL, realm_id=NULL)
				rule.ID = uuid.New().String()
				if err := db.Create(&rule).Error; err != nil {
					fmt.Printf("❌ Failed to create rule '%s': %v\n", name, err)
					errors++
					continue
				}
				fmt.Printf("✅ Created: %s\n", name)
				created++
			}
		}

		fmt.Printf("\n📊 Import Summary:\n")
		fmt.Printf("   Total:      %d\n", len(rulesYAML.CursorRules))
		fmt.Printf("   ✅ Created:  %d\n", created)
		fmt.Printf("   ✅ Updated:  %d\n", updated)
		fmt.Printf("   ⏭️  Skipped:  %d\n", skipped)
		fmt.Printf("   ❌ Errors:   %d\n", errors)
	},
}

// CursorCommandsYAML represents the structure of cursor_commands.yaml
type CursorCommandsYAML struct {
	CursorCommands map[string]struct {
		Description string `yaml:"description"`
		Command     string `yaml:"command"`
		Category    string `yaml:"category"`
		Language    string `yaml:"language"`
		Framework   string `yaml:"framework"`
		Tags        string `yaml:"tags"`
		IsTemplate  bool   `yaml:"is_template"`
	} `yaml:"cursor_commands"`
}

var importCursorCommandsCmd = &cobra.Command{
	Use:   "cursor-commands",
	Short: "Import cursor commands from YAML file to database",
	Long:  `Import cursor commands from cursor_commands.yaml into the database as global templates.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Initialize config
		if err := InitConfig(); err != nil {
			fmt.Printf("Failed to initialize config: %v\n", err)
			return
		}

		// Initialize database
		if err := database.InitDB(); err != nil {
			fmt.Printf("Failed to initialize database: %v\n", err)
			return
		}
		defer database.CloseDB()

		// Read YAML file
		commandsFile := "config/cursor_commands.yaml"
		data, err := os.ReadFile(commandsFile)
		if err != nil {
			fmt.Printf("❌ Failed to read file '%s': %v\n", commandsFile, err)
			return
		}

		// Parse YAML
		var commandsYAML CursorCommandsYAML
		if err := yaml.Unmarshal(data, &commandsYAML); err != nil {
			fmt.Printf("❌ Failed to parse YAML: %v\n", err)
			return
		}

		if len(commandsYAML.CursorCommands) == 0 {
			fmt.Println("No cursor commands found in YAML file")
			return
		}

		fmt.Printf("Found %d cursor commands in %s\n", len(commandsYAML.CursorCommands), commandsFile)

		db := database.GetDB()
		created := 0
		updated := 0
		skipped := 0
		errors := 0

		for name, cmdData := range commandsYAML.CursorCommands {
			// Check if command already exists
			var existingCmd models.CursorCommand
			result := db.Where("name = ?", name).First(&existingCmd)

			cmd := models.CursorCommand{
				Name:        name,
				Description: cmdData.Description,
				Command:     cmdData.Command,
				Category:    cmdData.Category,
				Language:    cmdData.Language,
				Framework:   cmdData.Framework,
				Tags:        cmdData.Tags,
				IsTemplate:  cmdData.IsTemplate,
				UsageCount:  0,
				CreatedBy:   createdBy,
				UpdatedBy:   createdBy,
			}

			if result.Error == nil {
				// Command exists
				if updateMode {
					// Update existing command
					cmd.ID = existingCmd.ID
					cmd.UserID = existingCmd.UserID
					cmd.RealmID = existingCmd.RealmID
					if err := db.Save(&cmd).Error; err != nil {
						fmt.Printf("❌ Failed to update command '%s': %v\n", name, err)
						errors++
						continue
					}
					fmt.Printf("✅ Updated: %s\n", name)
					updated++
				} else {
					// Skip existing command
					fmt.Printf("⏭️  Skipped (exists): %s\n", name)
					skipped++
				}
			} else {
				// Create new command (as global template: user_id=NULL, realm_id=NULL)
				cmd.ID = uuid.New().String()
				if err := db.Create(&cmd).Error; err != nil {
					fmt.Printf("❌ Failed to create command '%s': %v\n", name, err)
					errors++
					continue
				}
				fmt.Printf("✅ Created: %s\n", name)
				created++
			}
		}

		fmt.Printf("\n📊 Import Summary:\n")
		fmt.Printf("   Total:      %d\n", len(commandsYAML.CursorCommands))
		fmt.Printf("   ✅ Created:  %d\n", created)
		fmt.Printf("   ✅ Updated:  %d\n", updated)
		fmt.Printf("   ⏭️  Skipped:  %d\n", skipped)
		fmt.Printf("   ❌ Errors:   %d\n", errors)
	},
}

// extractArgumentsFromTemplate extracts {{variable}} patterns from template string
// and returns an array of PromptArgument for backward compatibility
func extractArgumentsFromTemplate(template string) []models.PromptArgument {
	var arguments []models.PromptArgument
	seen := make(map[string]bool)

	for i := 0; i < len(template)-1; i++ {
		if template[i] == '{' && template[i+1] == '{' {
			// Found opening {{
			start := i + 2
			end := start

			// Find closing }}
			for end < len(template)-1 {
				if template[end] == '}' && template[end+1] == '}' {
					// Found closing }}
					varName := strings.TrimSpace(template[start:end])
					if varName != "" && !seen[varName] {
						seen[varName] = true
						arguments = append(arguments, models.PromptArgument{
							Name:        varName,
							Description: fmt.Sprintf("Value for %s", varName),
							Required:    true,
						})
					}
					i = end + 1
					break
				}
				end++
			}
		}
	}

	return arguments
}

func init() {
	// Add flags for import prompts command
	importPromptsCmd.Flags().StringVarP(&promptsFile, "file", "f", "config/prompts.yaml", "Path to prompts.yaml file")
	importPromptsCmd.Flags().BoolVarP(&updateMode, "update", "u", false, "Update existing prompts")
	importPromptsCmd.Flags().BoolVarP(&dryRun, "dry-run", "d", false, "Preview changes without applying")
	importPromptsCmd.Flags().StringVarP(&createdBy, "user", "", "admin", "User who created/updated the prompts")

	// Add flags for import cursor rules command
	importCursorRulesCmd.Flags().BoolVarP(&updateMode, "update", "u", false, "Update existing cursor rules")
	importCursorRulesCmd.Flags().StringVarP(&createdBy, "user", "", "system", "User who created/updated the cursor rules")

	// Add flags for import cursor commands command
	importCursorCommandsCmd.Flags().BoolVarP(&updateMode, "update", "u", false, "Update existing cursor commands")
	importCursorCommandsCmd.Flags().StringVarP(&createdBy, "user", "", "system", "User who created/updated the cursor commands")

	// Add flags for import projects command
	importProjectsCmd.Flags().StringVarP(&realmID, "realm", "r", "default", "Realm ID for projects")
	importProjectsCmd.Flags().StringVarP(&createdBy, "user", "u", "admin", "User who created the projects")

	// Add subcommands
	importCmd.AddCommand(importPromptsCmd)
	importCmd.AddCommand(importProjectsCmd)
	importCmd.AddCommand(importCursorRulesCmd)
	importCmd.AddCommand(importCursorCommandsCmd)
	rootCmd.AddCommand(importCmd)
}
