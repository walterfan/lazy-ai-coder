package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

// helpCmd represents the help command
var helpCmd = &cobra.Command{
	Use:   "help",
	Short: "Show help information and examples",
	Long: `Show detailed help information and examples for lazy-ai-coder.

This command provides examples and usage information for all available commands.`,
	Run: func(cmd *cobra.Command, args []string) {
		showHelp()
	},
}

func init() {
	rootCmd.AddCommand(helpCmd)
}

// showHelp displays comprehensive help information
func showHelp() {
	fmt.Println("🚀 lazy-ai-coder - LLM-powered Go code assistant")
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println()

	fmt.Println("📖 Available Commands:")
	fmt.Println("  lazy-ai-coder <file>           # Process a file with LLM (default)")
	fmt.Println("  lazy-ai-coder cmd <command>    # Execute shell commands")
	fmt.Println("  lazy-ai-coder help             # Show this help message")
	fmt.Println()

	fmt.Println("🔧 Main Command (File Processing):")
	fmt.Println("  Usage: lazy-ai-coder <file> [flags]")
	fmt.Println("  Flags:")
	fmt.Println("    -c, --command string    Specify the command (e.g. explain, review, refactor)")
	fmt.Println("    -l, --language string   Specify the computer language (e.g. golang, java, python)")
	fmt.Println("    -t, --tongue string     Specify the output language (e.g. chinese, english, japanese)")
	fmt.Println("    -s, --stream            Enable streaming mode for LLM response")
	fmt.Println()

	fmt.Println("  Examples:")
	fmt.Println("    lazy-ai-coder main.go --command review --language golang")
	fmt.Println("    lazy-ai-coder app.py --command explain --tongue english")
	fmt.Println("    lazy-ai-coder index.js --command refactor --stream")
	fmt.Println()

	fmt.Println("💻 CMD Command (Shell Execution):")
	fmt.Println("  Usage: lazy-ai-coder cmd <command> [flags]")
	fmt.Println("  Flags:")
	fmt.Println("    -t, --timeout int       Command timeout in seconds (default: 300)")
	fmt.Println("    -s, --shell string      Shell to use (auto-detect if not specified)")
	fmt.Println("    -w, --workdir string    Working directory for command execution")
	fmt.Println("    -e, --env stringArray   Environment variables in KEY=VALUE format")
	fmt.Println()

	fmt.Println("  Examples:")
	fmt.Println("    lazy-ai-coder cmd 'ls -la'")
	fmt.Println("    lazy-ai-coder cmd 'git status'")
	fmt.Println("    lazy-ai-coder cmd 'docker ps'")
	fmt.Println("    lazy-ai-coder cmd 'go build .'")
	fmt.Println("    lazy-ai-coder cmd 'npm install'")
	fmt.Println("    lazy-ai-coder cmd --workdir /tmp 'pwd'")
	fmt.Println("    lazy-ai-coder cmd --env DEBUG=true 'echo $DEBUG'")
	fmt.Println("    lazy-ai-coder cmd --shell bash 'for i in {1..5}; do echo $i; done'")
	fmt.Println()

	fmt.Println("🌐 Web Interface:")
	fmt.Println("  Start the web server with: lazy-ai-coder web")
	fmt.Println("  Access at: http://localhost:8888")
	fmt.Println()

	fmt.Println("⚙️  Configuration:")
	fmt.Println("  Config file: config/config.yaml")
	fmt.Println("  Prompts file: config/prompts.yaml")
	fmt.Println("  Environment: .env file or environment variables")
	fmt.Println()

	fmt.Println("🔑 Environment Variables:")
	fmt.Println("  LLM_BASE_URL    # LLM API base URL")
	fmt.Println("  LLM_API_KEY     # LLM API key")
	fmt.Println("  LLM_MODEL       # LLM model name")
	fmt.Println()

	fmt.Println("📚 For more information, visit:")
	fmt.Println("  https://github.com/walterfan/lazy-ai-coder")
	fmt.Println()
}

// showCmdHelp shows help specifically for the cmd command
func showCmdHelp() {
	fmt.Println("💻 CMD Command - Execute shell commands")
	fmt.Println(strings.Repeat("=", 40))
	fmt.Println()

	fmt.Println("Usage:")
	fmt.Println("  lazy-ai-coder cmd <command> [flags]")
	fmt.Println()

	fmt.Println("Description:")
	fmt.Println("  Execute shell commands with support for different shells, working directories,")
	fmt.Println("  environment variables, and timeout settings.")
	fmt.Println()

	fmt.Println("Arguments:")
	fmt.Println("  command    The shell command to execute (required)")
	fmt.Println()

	fmt.Println("Flags:")
	fmt.Println("  -t, --timeout int       Command timeout in seconds (default: 300)")
	fmt.Println("  -s, --shell string      Shell to use (auto-detect if not specified)")
	fmt.Println("  -w, --workdir string    Working directory for command execution")
	fmt.Println("  -e, --env stringArray   Environment variables in KEY=VALUE format")
	fmt.Println()

	fmt.Println("Supported Shells:")
	fmt.Println("  bash, sh, zsh, powershell, cmd")
	fmt.Println()

	fmt.Println("Examples:")
	ExampleCommands()
	fmt.Println()

	fmt.Println("Notes:")
	fmt.Println("  - Commands are executed in the specified working directory")
	fmt.Println("  - Environment variables can be overridden with --env flag")
	fmt.Println("  - Output is streamed in real-time with STDOUT/STDERR prefixes")
	fmt.Println("  - Exit codes are preserved and displayed")
	fmt.Println()
}
