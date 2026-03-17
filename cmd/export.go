package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
	"github.com/walterfan/lazy-ai-coder/pkg/database"
	"github.com/walterfan/lazy-ai-coder/pkg/models"
)

var (
	exportFile string
	exportAll  bool
)

var exportCmd = &cobra.Command{
	Use:   "export",
	Short: "Export data to files",
}

var exportPromptsCmd = &cobra.Command{
	Use:   "prompts",
	Short: "Export prompts from database to YAML file",
	Long: `Export prompts from the SQLite database to a YAML file.

This command reads all prompts from the database and exports them to a YAML file
with format: prompts_YYYY-MM-DD.yaml`,
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

		db := database.GetDB()

		// Fetch all prompts from database
		var prompts []models.Prompt
		query := db.Order("name ASC")

		if !exportAll {
			// Only export global prompts (without user_id and realm_id)
			query = query.Where("user_id IS NULL AND realm_id IS NULL")
		}

		if err := query.Find(&prompts).Error; err != nil {
			fmt.Printf("❌ Failed to fetch prompts: %v\n", err)
			return
		}

		if len(prompts) == 0 {
			fmt.Println("No prompts found in database")
			return
		}

		fmt.Printf("Found %d prompts in database\n", len(prompts))

		// Generate default filename with date if not provided
		if exportFile == "" {
			exportFile = fmt.Sprintf("prompts_%s.yaml", time.Now().Format("2006-01-02"))
		}

		// Convert to YAML structure
		promptsYAML := PromptYAML{
			Prompts: make(map[string]PromptData),
		}

		for _, prompt := range prompts {
			// Parse arguments from JSON string if available
			var arguments []models.PromptArgument
			if prompt.Arguments != "" {
				if err := json.Unmarshal([]byte(prompt.Arguments), &arguments); err != nil {
					fmt.Printf("⚠️  Warning: Failed to parse arguments for '%s': %v\n", prompt.Name, err)
				}
			}

			promptsYAML.Prompts[prompt.Name] = PromptData{
				Title:        prompt.Title,
				Description:  prompt.Description,
				SystemPrompt: prompt.SystemPrompt,
				UserPrompt:   prompt.UserPrompt,
				Arguments:    arguments,
				Tags:         prompt.Tags,
			}
		}

		// Marshal to YAML
		data, err := yaml.Marshal(&promptsYAML)
		if err != nil {
			fmt.Printf("❌ Failed to marshal YAML: %v\n", err)
			return
		}

		// Write to file
		if err := os.WriteFile(exportFile, data, 0644); err != nil {
			fmt.Printf("❌ Failed to write file '%s': %v\n", exportFile, err)
			return
		}

		fmt.Printf("✅ Successfully exported %d prompts to '%s'\n", len(prompts), exportFile)
		fmt.Printf("\n📊 Export Summary:\n")
		fmt.Printf("   Total Exported: %d\n", len(prompts))
		fmt.Printf("   Output File:    %s\n", exportFile)
	},
}

func init() {
	// Add flags for export prompts command
	exportPromptsCmd.Flags().StringVarP(&exportFile, "output", "o", "", "Output file path (default: prompts_YYYY-MM-DD.yaml)")
	exportPromptsCmd.Flags().BoolVarP(&exportAll, "all", "a", false, "Export all prompts including user-specific ones")

	// Add subcommands
	exportCmd.AddCommand(exportPromptsCmd)
	rootCmd.AddCommand(exportCmd)
}
