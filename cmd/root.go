package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/joho/godotenv"
	"github.com/spf13/cobra"
	"github.com/walterfan/lazy-ai-coder/internal/llm"
	"github.com/walterfan/lazy-ai-coder/internal/log"
)

type PromptConfig struct {
	Name         string `mapstructure:"name" json:"name"`
	Description  string `mapstructure:"description" json:"description"`
	SystemPrompt string `mapstructure:"system_prompt" json:"system_prompt"`
	UserPrompt   string `mapstructure:"user_prompt" json:"user_prompt"`
	Tags         string `mapstructure:"tags" json:"tags"`
}

func processCommand(cmd *cobra.Command, args []string) {
	path := args[0]
	commandName, _ := cmd.Flags().GetString("command")
	streamMode, _ := cmd.Flags().GetBool("stream")
	outputLanguage, _ := cmd.Flags().GetString("language")

	code, err := os.ReadFile(path)
	if err != nil {
		fmt.Println("Error reading file:", err)
		return
	}

	selectedPrompt, err := getPromptConfigByName(commandName)
	if err != nil {
		fmt.Println(err)
		return
	}

	// Build final prompt by replacing {{code}}
	promptText := strings.Replace(selectedPrompt.UserPrompt, "{{code}}", string(code), -1)

	if outputLanguage != "" {
		promptText = strings.Replace(promptText, "{{output_language}}", outputLanguage, -1)
	}

	// Create LLM settings from environment variables
	llmSettings := llm.LLMSettings{
		BaseUrl:     os.Getenv("LLM_BASE_URL"),
		ApiKey:      os.Getenv("LLM_API_KEY"),
		Model:       os.Getenv("LLM_MODEL"),
		Temperature: 1.0,
	}

	if streamMode {
		err = llm.AskLLMWithStream(selectedPrompt.SystemPrompt, promptText, llmSettings, func(chunk string) {
			fmt.Print(chunk)
		})
	} else {
		resp, err := llm.AskLLM(selectedPrompt.SystemPrompt, promptText, llmSettings)
		if err == nil {
			fmt.Printf("📝 Result (%s):\n%s\n", commandName, resp)
		}
	}

	if err != nil {
		fmt.Println("LLM error:", err)
	}
}

var rootCmd = &cobra.Command{
	Use:   "lazy-ai-coder  <file>",
	Short: "An LLM-powered Go code assistant for explain, review, refactor",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		processCommand(cmd, args)
	},
}

func init() {

	err := log.InitLogger()
	if err != nil {
		panic(err)
	}

	if err := godotenv.Load(); err != nil {
		fmt.Fprintln(os.Stderr, "No .env file found, using environment variables")
	}

	rootCmd.Flags().StringP("command", "c", "review", "Specify the command (e.g. explain, review, refactor)")
	rootCmd.Flags().StringP("language", "l", "golang", "Specify the computer language (e.g. golang, java, python, etc.)")
	rootCmd.Flags().StringP("tongue", "t", "chinese", "Specify the output language (e.g. chinese, english, japanese, etc.)")
	rootCmd.Flags().BoolP("stream", "s", false, "Enable streaming mode for LLM response")
}

func Execute() {
	cobra.CheckErr(rootCmd.Execute())
}
