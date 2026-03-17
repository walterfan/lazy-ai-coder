package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/walterfan/lazy-ai-coder/internal/log"
	"github.com/walterfan/lazy-ai-coder/internal/mcp"
	"github.com/walterfan/lazy-ai-coder/pkg/database"
)

var mcpCmd = &cobra.Command{
	Use:   "mcp",
	Short: "Start an MCP (Model Context Protocol) server",
	Long: `Start an MCP server that exposes GitLab and LLM capabilities as MCP tools.

The MCP server uses stdio transport and follows the Model Context Protocol specification.
It can be used with MCP clients like Claude Desktop, VSCode extensions, or custom integrations.

Available Tools:
  - get_gitlab_file_content: Retrieve file content from GitLab repositories
  - get_gitlab_merge_request: Retrieve merge request changes
  - get_gitlab_project_id: Get numeric project ID from project path
  - llm_chat: Send messages to LLM and get responses
  - llm_analyze_code: Analyze code using LLM
  - generate_plantuml: Generate PlantUML diagrams

Environment Variables Required:
  - GITLAB_BASE_URL: GitLab instance URL (e.g., https://gitlab.com)
  - GITLAB_TOKEN: GitLab private token for authentication
  - LLM_BASE_URL: LLM API base URL
  - LLM_API_KEY: LLM API key
  - LLM_MODEL: LLM model name (e.g., gpt-4, claude-3-sonnet)
  - PLANTUML_URL: PlantUML server URL (optional)

Example Usage:
  # Start MCP server
  lazy-ai-coder mcp

  # Use with Claude Desktop - add to config.json:
  {
    "mcpServers": {
      "lazy-ai-coder": {
        "command": "/path/to/lazy-ai-coder",
        "args": ["mcp"],
        "env": {
          "GITLAB_BASE_URL": "https://gitlab.com",
          "GITLAB_TOKEN": "your-token",
          "LLM_BASE_URL": "https://api.openai.com/v1",
          "LLM_API_KEY": "your-api-key",
          "LLM_MODEL": "gpt-4"
        }
      }
    }
  }
`,
	Run: func(cmd *cobra.Command, args []string) {
		// MCP command works entirely from environment variables
		// No config.yaml required - this allows running from any directory

	// Enable silent mode for database logging (MCP uses stdio for protocol)
	database.SetSilentMode(true)

	// Initialize logger with file-only output (no console output for MCP mode)
	// Don't print to stderr even if it fails - MCP uses stdio for protocol
	_ = log.InitLoggerFileOnly()

		// Validate environment variables
		requiredEnvVars := []string{
			"LLM_BASE_URL",
			"LLM_API_KEY",
			"LLM_MODEL",
		}

		missingVars := []string{}
		invalidVars := []string{}
		for _, envVar := range requiredEnvVars {
			value := os.Getenv(envVar)
			if value == "" {
				missingVars = append(missingVars, envVar)
			} else if strings.HasPrefix(value, "${") && strings.HasSuffix(value, "}") {
				// Detect unexpanded environment variable placeholders like ${env:VAR}
				invalidVars = append(invalidVars, fmt.Sprintf("%s=%s", envVar, value))
			}
		}

		if len(missingVars) > 0 {
			fmt.Fprintf(os.Stderr, "Error: Missing required environment variables: %v\n", missingVars)
			fmt.Fprintf(os.Stderr, "Please set these variables before starting the MCP server.\n")
			os.Exit(1)
		}

		if len(invalidVars) > 0 {
			fmt.Fprintf(os.Stderr, "Error: Environment variables contain unexpanded placeholders: %v\n", invalidVars)
			fmt.Fprintf(os.Stderr, "Your MCP client may not support ${env:VAR} syntax.\n")
			fmt.Fprintf(os.Stderr, "Try:\n")
			fmt.Fprintf(os.Stderr, "  1. Use actual values in your MCP config\n")
			fmt.Fprintf(os.Stderr, "  2. Set environment variables in your shell\n")
			fmt.Fprintf(os.Stderr, "  3. Use a wrapper script to source environment variables\n")
			os.Exit(1)
		}

		// Create and configure MCP server
		server := mcp.NewServer()

		// Initialize database and load prompts/resources
		if err := server.InitializeWithDB(); err != nil {
			// Don't print to stderr - it interferes with MCP protocol
			// Error is already logged to file
			// Continue anyway - tools will still work
		}

		// Register all tools
		server.RegisterAllTools()

		// Start the MCP server (uses stdio transport)
		if err := server.Start(); err != nil {
			// Only print fatal errors to stderr before exit
			fmt.Fprintf(os.Stderr, "MCP server error: %v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(mcpCmd)
}
