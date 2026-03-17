package cmd

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"syscall"

	"github.com/spf13/cobra"
)

var (
	cmdTimeout int
	cmdShell   string
	cmdWorkDir string
	cmdEnv     []string
)

// cmdCmd represents the cmd command
var cmdCmd = &cobra.Command{
	Use:   "cmd [command]",
	Short: "Execute a shell command",
	Long: `Execute a shell command on the command line.

Examples:
  lazy-ai-coder cmd "ls -la"                    # Execute ls command
  lazy-ai-coder cmd "git status"                # Check git status
  lazy-ai-coder cmd "docker ps"                 # List docker containers
  lazy-ai-coder cmd "go build ."                # Build Go project
  lazy-ai-coder cmd "npm install"               # Install npm packages
  lazy-ai-coder cmd "echo 'Hello World'"        # Simple echo command

Flags:
  --timeout int        Command timeout in seconds (default: 300)
  --shell string       Shell to use (default: auto-detect)
  --workdir string     Working directory for command execution
  --env stringArray    Environment variables in KEY=VALUE format`,
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		executeCommand(args)
	},
}

func init() {
	rootCmd.AddCommand(cmdCmd)

	// Add flags
	cmdCmd.Flags().IntVarP(&cmdTimeout, "timeout", "t", 300, "Command timeout in seconds")
	cmdCmd.Flags().StringVarP(&cmdShell, "shell", "s", "", "Shell to use (auto-detect if not specified)")
	cmdCmd.Flags().StringVarP(&cmdWorkDir, "workdir", "w", "", "Working directory for command execution")
	cmdCmd.Flags().StringArrayVarP(&cmdEnv, "env", "e", []string{}, "Environment variables in KEY=VALUE format")
}

// executeCommand executes the given shell command
func executeCommand(args []string) {
	// Join all arguments into a single command string
	commandStr := strings.Join(args, " ")

	fmt.Printf("🚀 Executing command: %s\n", commandStr)
	fmt.Printf("📍 Working directory: %s\n", getWorkingDir())
	fmt.Printf("⏱️  Timeout: %d seconds\n", cmdTimeout)

	// Parse environment variables
	envVars := parseEnvironmentVariables()

	// Determine shell to use
	shell, shellArgs := determineShell(commandStr)

	// Create command
	var cmd *exec.Cmd
	if shell != "" {
		// Use specified shell
		cmd = exec.Command(shell, shellArgs...)
	} else {
		// Use system default shell
		cmd = exec.Command(commandStr)
	}

	// Set working directory
	if cmdWorkDir != "" {
		cmd.Dir = cmdWorkDir
	}

	// Set environment variables
	cmd.Env = append(os.Environ(), envVars...)

	// Set up pipes for output
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		fmt.Printf("❌ Error setting up stdout pipe: %v\n", err)
		return
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		fmt.Printf("❌ Error setting up stderr pipe: %v\n", err)
		return
	}

	// Start command
	fmt.Println("🔄 Starting command execution...")
	if err := cmd.Start(); err != nil {
		fmt.Printf("❌ Error starting command: %v\n", err)
		return
	}

	// Read output streams concurrently
	go readOutput(stdout, "STDOUT")
	go readOutput(stderr, "STDERR")

	// Wait for command to complete
	if err := cmd.Wait(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			if status, ok := exitErr.Sys().(syscall.WaitStatus); ok {
				fmt.Printf("⚠️  Command exited with status: %d\n", status.ExitStatus())
			} else {
				fmt.Printf("⚠️  Command exited with error: %v\n", err)
			}
		} else {
			fmt.Printf("❌ Command failed: %v\n", err)
		}
	} else {
		fmt.Println("✅ Command completed successfully")
	}
}

// determineShell determines which shell to use and returns shell name and arguments
func determineShell(commandStr string) (string, []string) {
	if cmdShell != "" {
		// Use specified shell
		switch cmdShell {
		case "bash":
			return "bash", []string{"-c", commandStr}
		case "sh":
			return "sh", []string{"-c", commandStr}
		case "zsh":
			return "zsh", []string{"-c", commandStr}
		case "powershell":
			return "powershell", []string{"-Command", commandStr}
		case "cmd":
			return "cmd", []string{"/C", commandStr}
		default:
			return cmdShell, []string{"-c", commandStr}
		}
	}

	// Auto-detect shell based on OS
	if isWindows() {
		// On Windows, try PowerShell first, then cmd
		if _, err := exec.LookPath("powershell"); err == nil {
			return "powershell", []string{"-Command", commandStr}
		}
		return "cmd", []string{"/C", commandStr}
	}

	// On Unix-like systems, try common shells
	shells := []string{"bash", "zsh", "sh"}
	for _, shell := range shells {
		if _, err := exec.LookPath(shell); err == nil {
			return shell, []string{"-c", commandStr}
		}
	}

	// Fallback to system default
	return "", nil
}

// getWorkingDir returns the working directory for command execution
func getWorkingDir() string {
	if cmdWorkDir != "" {
		return cmdWorkDir
	}

	currentDir, err := os.Getwd()
	if err != nil {
		return "unknown"
	}
	return currentDir
}

// parseEnvironmentVariables parses the --env flag values
func parseEnvironmentVariables() []string {
	var envVars []string

	for _, envVar := range cmdEnv {
		if strings.Contains(envVar, "=") {
			envVars = append(envVars, envVar)
		} else {
			fmt.Printf("⚠️  Invalid environment variable format: %s (should be KEY=VALUE)\n", envVar)
		}
	}

	return envVars
}

// readOutput reads from a pipe and prints the output with a prefix
func readOutput(pipe io.ReadCloser, prefix string) {
	scanner := bufio.NewScanner(pipe)
	for scanner.Scan() {
		line := scanner.Text()
		fmt.Printf("[%s] %s\n", prefix, line)
	}

	if err := scanner.Err(); err != nil {
		fmt.Printf("❌ Error reading %s: %v\n", prefix, err)
	}
}

// isWindows checks if the current OS is Windows
func isWindows() bool {
	return os.PathSeparator == '\\' && os.PathListSeparator == ';'
}

// Example usage functions for common commands
func ExampleCommands() {
	examples := []string{
		"lazy-ai-coder cmd 'ls -la'",
		"lazy-ai-coder cmd 'git status'",
		"lazy-ai-coder cmd 'docker ps'",
		"lazy-ai-coder cmd 'go build .'",
		"lazy-ai-coder cmd 'npm install'",
		"lazy-ai-coder cmd 'echo \"Hello World\"'",
		"lazy-ai-coder cmd --workdir /tmp 'pwd'",
		"lazy-ai-coder cmd --env PATH=/usr/local/bin --env DEBUG=true 'echo $PATH'",
		"lazy-ai-coder cmd --shell bash 'for i in {1..5}; do echo $i; done'",
		"lazy-ai-coder cmd --timeout 60 'sleep 30'",
	}

	fmt.Println("📚 Example commands:")
	for _, example := range examples {
		fmt.Printf("  %s\n", example)
	}
}
