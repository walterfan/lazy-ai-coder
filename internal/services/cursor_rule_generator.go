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

// CursorRuleGenerator handles AI-powered generation and refinement of cursor rules
type CursorRuleGenerator struct {
}

// NewCursorRuleGenerator creates a new cursor rule generator
func NewCursorRuleGenerator() *CursorRuleGenerator {
	return &CursorRuleGenerator{}
}

// GenerateFromProject generates cursor rules based on project context
func (g *CursorRuleGenerator) GenerateFromProject(projectContext models.ProjectContext, requirements string, settings models.Settings) (string, error) {
	logger := log.GetLogger()

	llmSettings := util.ConvertToLLMSettings(settings)

	systemPrompt := `You are an expert at creating Cursor IDE rules files (.cursorrules). 
Your task is to generate comprehensive, well-structured cursor rules that help AI assistants understand 
project conventions, coding standards, and best practices.

**IMPORTANT GUIDELINES:**
- Keep rules under 500 lines - if content exceeds this, split into multiple composable rules
- Split large rules into multiple, composable rules that can be reused
- Provide concrete examples or referenced files - avoid abstract descriptions
- Avoid vague guidance - write rules like clear internal documentation
- Be specific, actionable, and include examples where helpful
- Structure rules so they can be reused when repeating prompts in chat

The rules should be written in markdown format and include:
- Project overview and architecture guidelines
- Code style and formatting rules
- Naming conventions
- File organization patterns
- Testing requirements
- Security best practices
- Performance considerations
- Common patterns and anti-patterns

Write clear, specific rules that serve as internal documentation for the project.`

	userPrompt := fmt.Sprintf(`Generate a comprehensive .cursorrules file for a project with the following context:

**Language:** %s
**Framework:** %s
**Build Tool:** %s
**Database:** %s
**Test Framework:** %s
**Dependencies:** %s

**Additional Requirements:**
%s

Please generate a complete .cursorrules file that covers all aspects of development for this project.`, 
		projectContext.Language,
		projectContext.Framework,
		projectContext.BuildTool,
		projectContext.Database,
		projectContext.TestFramework,
		strings.Join(projectContext.Dependencies, ", "),
		requirements)

	logger.Infof("Generating cursor rule from project context: Language=%s, Framework=%s", 
		projectContext.Language, projectContext.Framework)

	content, err := llm.AskLLM(systemPrompt, userPrompt, llmSettings)
	if err != nil {
		return "", fmt.Errorf("failed to generate cursor rule: %w", err)
	}

	return content, nil
}

// GenerateFromTemplate generates cursor rules from a template + user input
func (g *CursorRuleGenerator) GenerateFromTemplate(template *pkgmodels.CursorRule, requirements string, settings models.Settings) (string, error) {
	logger := log.GetLogger()

	llmSettings := util.ConvertToLLMSettings(settings)

	systemPrompt := `You are an expert at creating Cursor IDE rules files (.cursorrules). 
Your task is to customize a template cursor rule based on user requirements while maintaining 
the structure and quality of the template.

**IMPORTANT GUIDELINES:**
- Keep rules under 500 lines - if content exceeds this, split into multiple composable rules
- Split large rules into multiple, composable rules that can be reused
- Provide concrete examples or referenced files - avoid abstract descriptions
- Avoid vague guidance - write rules like clear internal documentation
- Structure rules so they can be reused when repeating prompts in chat

Adapt the template to match the user's specific needs while preserving best practices and 
comprehensive coverage.`

	userPrompt := fmt.Sprintf(`Based on this template cursor rule, generate a customized version:

**Template:**
%s

**User Requirements:**
%s

Please generate a customized .cursorrules file that incorporates the user requirements while 
maintaining the quality and structure of the template.`, 
		template.Content,
		requirements)

	logger.Infof("Generating cursor rule from template: %s", template.Name)

	content, err := llm.AskLLM(systemPrompt, userPrompt, llmSettings)
	if err != nil {
		return "", fmt.Errorf("failed to generate cursor rule from template: %w", err)
	}

	return content, nil
}

// GenerateFromScratch generates cursor rules from scratch based on language/framework/requirements
func (g *CursorRuleGenerator) GenerateFromScratch(language, framework, requirements string, settings models.Settings) (string, error) {
	logger := log.GetLogger()

	llmSettings := util.ConvertToLLMSettings(settings)

	systemPrompt := `You are an expert at creating Cursor IDE rules files (.cursorrules). 
Your task is to generate comprehensive, well-structured cursor rules that help AI assistants 
understand project conventions, coding standards, and best practices.

**IMPORTANT GUIDELINES:**
- Keep rules under 500 lines - if content exceeds this, split into multiple composable rules
- Split large rules into multiple, composable rules that can be reused
- Provide concrete examples or referenced files - avoid abstract descriptions
- Avoid vague guidance - write rules like clear internal documentation
- Be specific, actionable, and include examples where helpful
- Structure rules so they can be reused when repeating prompts in chat

The rules should be written in markdown format and include:
- Project overview and architecture guidelines
- Code style and formatting rules
- Naming conventions
- File organization patterns
- Testing requirements
- Security best practices
- Performance considerations
- Common patterns and anti-patterns

Write clear, specific rules that serve as internal documentation for the project.`

	userPrompt := fmt.Sprintf(`Generate a comprehensive .cursorrules file for a %s project`, language)
	if framework != "" {
		userPrompt += fmt.Sprintf(` using %s framework`, framework)
	}
	if requirements != "" {
		userPrompt += fmt.Sprintf(`\n\n**Requirements:**\n%s`, requirements)
	}
	userPrompt += "\n\nPlease generate a complete .cursorrules file that covers all aspects of development for this project."

	logger.Infof("Generating cursor rule from scratch: Language=%s, Framework=%s", language, framework)

	content, err := llm.AskLLM(systemPrompt, userPrompt, llmSettings)
	if err != nil {
		return "", fmt.Errorf("failed to generate cursor rule: %w", err)
	}

	return content, nil
}

// RefineRule improves an existing cursor rule with AI
func (g *CursorRuleGenerator) RefineRule(rule *pkgmodels.CursorRule, improvements string, focusAreas []string, settings models.Settings) (string, error) {
	logger := log.GetLogger()

	llmSettings := util.ConvertToLLMSettings(settings)

	systemPrompt := `You are an expert at improving Cursor IDE rules files (.cursorrules). 
Your task is to refine and enhance an existing cursor rule based on user feedback and focus areas.

**IMPORTANT GUIDELINES:**
- Keep rules under 500 lines - if content exceeds this, split into multiple composable rules
- Split large rules into multiple, composable rules that can be reused
- Provide concrete examples or referenced files - avoid abstract descriptions
- Avoid vague guidance - write rules like clear internal documentation
- Structure rules so they can be reused when repeating prompts in chat

Improve clarity, completeness, and usefulness while maintaining the original structure and intent.
Add missing sections, clarify ambiguous rules, and enhance examples where needed.`

	focusAreasStr := "all areas"
	if len(focusAreas) > 0 {
		focusAreasStr = strings.Join(focusAreas, ", ")
	}

	userPrompt := fmt.Sprintf(`Please refine and improve this cursor rule:

**Current Rule:**
%s

**Focus Areas:** %s

**Improvements Requested:**
%s

Please generate an improved version of the cursor rule that addresses the requested improvements 
while maintaining or enhancing the overall quality and structure.`, 
		rule.Content,
		focusAreasStr,
		improvements)

	logger.Infof("Refining cursor rule: %s", rule.Name)

	content, err := llm.AskLLM(systemPrompt, userPrompt, llmSettings)
	if err != nil {
		return "", fmt.Errorf("failed to refine cursor rule: %w", err)
	}

	return content, nil
}

// SuggestImprovements suggests improvements for a cursor rule
func (g *CursorRuleGenerator) SuggestImprovements(rule *pkgmodels.CursorRule, settings models.Settings) (string, error) {
	logger := log.GetLogger()

	llmSettings := util.ConvertToLLMSettings(settings)

	systemPrompt := `You are an expert at reviewing Cursor IDE rules files (.cursorrules). 
Your task is to analyze a cursor rule and suggest specific improvements.

Provide actionable suggestions for:
- Missing sections or topics
- Unclear or ambiguous rules
- Areas that need more detail or examples
- Best practices that could be added
- Organization and structure improvements

Be specific and constructive in your suggestions.`

	userPrompt := fmt.Sprintf(`Please analyze this cursor rule and suggest improvements:

**Current Rule:**
%s

Provide specific, actionable suggestions for improving this cursor rule.`, rule.Content)

	logger.Infof("Analyzing cursor rule for improvements: %s", rule.Name)

	suggestions, err := llm.AskLLM(systemPrompt, userPrompt, llmSettings)
	if err != nil {
		return "", fmt.Errorf("failed to suggest improvements: %w", err)
	}

	return suggestions, nil
}

