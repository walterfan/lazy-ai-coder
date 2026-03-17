package services

import (
	"fmt"
	"strings"

	"github.com/walterfan/lazy-ai-coder/internal/llm"
	"github.com/walterfan/lazy-ai-coder/internal/log"
	"github.com/walterfan/lazy-ai-coder/internal/models"
	"github.com/walterfan/lazy-ai-coder/internal/util"
	pkgmodels "github.com/walterfan/lazy-ai-coder/pkg/models"
)

// CursorCommandGenerator handles AI-powered generation and refinement of cursor commands
type CursorCommandGenerator struct {
}

// NewCursorCommandGenerator creates a new cursor command generator
func NewCursorCommandGenerator() *CursorCommandGenerator {
	return &CursorCommandGenerator{}
}

// GenerateFromScratch generates cursor commands from scratch based on category/language/framework/requirements
func (g *CursorCommandGenerator) GenerateFromScratch(category, language, framework, requirements string, settings models.Settings) (string, error) {
	logger := log.GetLogger()

	llmSettings := util.ConvertToLLMSettings(settings)

	systemPrompt := `You are an expert at creating effective Cursor IDE commands/prompts. 
Your task is to generate clear, actionable commands that users can directly use in Cursor's chat interface.

**IMPORTANT GUIDELINES:**
- Commands should be concise but complete - typically 1-3 sentences
- Be specific and actionable - avoid vague requests
- Include context when needed (file paths, function names, etc.)
- Use clear language that the AI assistant can understand
- Consider the category (refactor, debug, generate, review, test, etc.)
- Provide concrete examples or references when helpful
- Make commands reusable with placeholders if appropriate

The command should be ready to use directly in Cursor's chat.`

	categoryDesc := getCategoryDescription(category)
	
	userPrompt := fmt.Sprintf(`Generate a Cursor IDE command/prompt for: %s`, categoryDesc)
	if language != "" {
		userPrompt += fmt.Sprintf(`\n**Language:** %s`, language)
	}
	if framework != "" {
		userPrompt += fmt.Sprintf(`\n**Framework:** %s`, framework)
	}
	if requirements != "" {
		userPrompt += fmt.Sprintf(`\n**Specific Requirements:**\n%s`, requirements)
	}
	userPrompt += "\n\nGenerate a clear, actionable command that can be used directly in Cursor's chat interface."

	logger.Infof("Generating cursor command from scratch: Category=%s, Language=%s, Framework=%s", category, language, framework)

	content, err := llm.AskLLM(systemPrompt, userPrompt, llmSettings)
	if err != nil {
		return "", fmt.Errorf("failed to generate cursor command: %w", err)
	}

	return strings.TrimSpace(content), nil
}

// GenerateFromTemplate generates cursor commands from a template + user input
func (g *CursorCommandGenerator) GenerateFromTemplate(template *pkgmodels.CursorCommand, requirements string, settings models.Settings) (string, error) {
	logger := log.GetLogger()

	llmSettings := util.ConvertToLLMSettings(settings)

	systemPrompt := `You are an expert at creating effective Cursor IDE commands/prompts. 
Your task is to customize a template command based on user requirements while maintaining 
the clarity and effectiveness of the template.

**IMPORTANT GUIDELINES:**
- Commands should be concise but complete - typically 1-3 sentences
- Be specific and actionable - avoid vague requests
- Include context when needed (file paths, function names, etc.)
- Use clear language that the AI assistant can understand
- Adapt the template to match the user's specific needs while maintaining quality

The command should be ready to use directly in Cursor's chat.`

	userPrompt := fmt.Sprintf(`Based on this template command, generate a customized version:

**Template:**
%s

**User Requirements:**
%s

Please generate a customized command that incorporates the user requirements while 
maintaining the clarity and effectiveness of the template.`, 
		template.Command,
		requirements)

	logger.Infof("Generating cursor command from template: %s", template.Name)

	content, err := llm.AskLLM(systemPrompt, userPrompt, llmSettings)
	if err != nil {
		return "", fmt.Errorf("failed to generate cursor command from template: %w", err)
	}

	return strings.TrimSpace(content), nil
}

// RefineCommand improves an existing cursor command with AI
func (g *CursorCommandGenerator) RefineCommand(cmd *pkgmodels.CursorCommand, improvements string, focusAreas []string, settings models.Settings) (string, error) {
	logger := log.GetLogger()

	llmSettings := util.ConvertToLLMSettings(settings)

	systemPrompt := `You are an expert at improving Cursor IDE commands/prompts. 
Your task is to refine and enhance an existing command based on user feedback and focus areas.

**IMPORTANT GUIDELINES:**
- Commands should be concise but complete - typically 1-3 sentences
- Be specific and actionable - avoid vague requests
- Include context when needed (file paths, function names, etc.)
- Use clear language that the AI assistant can understand
- Improve clarity, completeness, and usefulness while maintaining the original intent

The refined command should be ready to use directly in Cursor's chat.`

	focusAreasStr := "all areas"
	if len(focusAreas) > 0 {
		focusAreasStr = strings.Join(focusAreas, ", ")
	}

	userPrompt := fmt.Sprintf(`Please refine and improve this cursor command:

**Current Command:**
%s

**Focus Areas:** %s

**Improvements Requested:**
%s

Please generate an improved version of the command that addresses the requested improvements 
while maintaining or enhancing the overall clarity and effectiveness.`, 
		cmd.Command,
		focusAreasStr,
		improvements)

	logger.Infof("Refining cursor command: %s", cmd.Name)

	content, err := llm.AskLLM(systemPrompt, userPrompt, llmSettings)
	if err != nil {
		return "", fmt.Errorf("failed to refine cursor command: %w", err)
	}

	return strings.TrimSpace(content), nil
}

// getCategoryDescription returns a description for the category
func getCategoryDescription(category string) string {
	descriptions := map[string]string{
		"refactor":   "refactoring code to improve structure, readability, or maintainability",
		"debug":      "debugging and fixing issues in code",
		"generate":   "generating new code, functions, or components",
		"review":     "reviewing code for quality, security, or best practices",
		"test":       "creating or improving tests",
		"document":   "documenting code, APIs, or functionality",
		"optimize":   "optimizing code for performance",
		"migrate":    "migrating code to new versions or frameworks",
		"analyze":    "analyzing code structure, dependencies, or patterns",
		"general":    "general purpose command",
	}
	
	if desc, ok := descriptions[strings.ToLower(category)]; ok {
		return desc
	}
	return category
}

