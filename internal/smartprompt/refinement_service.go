package smartprompt

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/walterfan/lazy-ai-coder/internal/log"
	"github.com/walterfan/lazy-ai-coder/internal/models"
)

// RefinementSuggestion represents a single improvement suggestion
type RefinementSuggestion struct {
	Title       string `json:"title"`
	Before      string `json:"before"`
	After       string `json:"after"`
	Impact      string `json:"impact"` // high, medium, low
	Description string `json:"description"`
}

// RefinementResponse represents the response from refinement service
type RefinementResponse struct {
	OriginalPrompt  string                 `json:"original_prompt"`
	RefinedPrompt   string                 `json:"refined_prompt"`
	Assessment      string                 `json:"assessment"`
	Suggestions     []RefinementSuggestion `json:"suggestions"`
	QualityBefore   float64                `json:"quality_before"`
	QualityAfter    float64                `json:"quality_after"`
	ImprovementTips []string               `json:"improvement_tips"`
}

// RefinePrompt uses LLM to suggest improvements to a prompt
func RefinePrompt(originalPrompt string, settings models.Settings) (*RefinementResponse, error) {
	logger := log.GetLogger()

	// Build refinement system prompt
	systemPrompt := `You are an expert prompt engineer specializing in helping developers write better prompts for code generation and technical tasks.

Your task is to analyze prompts and provide specific, actionable improvements.

When analyzing a prompt, consider:
1. **Clarity**: Is the request clear and unambiguous?
2. **Context**: Is there sufficient background information?
3. **Specificity**: Are requirements detailed enough?
4. **Structure**: Is the prompt well-organized?
5. **Output Definition**: Is the expected output clearly defined?
6. **Examples**: Would examples make the prompt more effective?
7. **Constraints**: Are important limitations or requirements specified?

Provide your analysis in this JSON format:
{
  "assessment": "1-2 sentence overall evaluation",
  "quality_before": 7.5,
  "quality_after": 9.0,
  "suggestions": [
    {
      "title": "Add specific technology versions",
      "before": "using Python",
      "after": "using Python 3.11 with FastAPI 0.104",
      "impact": "high",
      "description": "Specifying exact versions prevents ambiguity and ensures compatibility"
    }
  ],
  "refined_prompt": "Complete improved version of the prompt",
  "improvement_tips": [
    "Consider adding error scenarios",
    "Specify expected code structure"
  ]
}

Focus on practical, high-impact improvements. Limit to 3-4 suggestions.`

	// Build user prompt
	userPrompt := fmt.Sprintf(`Please analyze and improve this coding prompt:

---
%s
---

Provide specific suggestions to make this more effective for code generation.`, originalPrompt)

	// Call LLM API
	response, err := callLLMForRefinement(systemPrompt, userPrompt, settings)
	if err != nil {
		logger.Errorf("Failed to call LLM for refinement: %v", err)
		return nil, fmt.Errorf("refinement failed: %w", err)
	}

	return response, nil
}

// RefinePromptWithRequirements uses LLM to refine system and user prompts based on user requirements
func RefinePromptWithRequirements(systemPrompt, userPrompt, requirements string, settings models.Settings) (*RefineWithRequirementsResponse, error) {
	logger := log.GetLogger()

	// Build refinement system prompt
	refinementSystemPrompt := `You are an expert prompt engineer specializing in refining and improving prompts based on specific user requirements.

Your task is to analyze the existing system and user prompts, understand the user's refinement requirements, and generate improved versions of both prompts.

When refining prompts, consider:
1. **User Requirements**: Directly address all requirements specified by the user
2. **Clarity**: Ensure the prompts are clear and unambiguous
3. **Structure**: Maintain or improve the organization
4. **Specificity**: Add relevant details where needed
5. **Consistency**: Ensure system and user prompts work well together
6. **Best Practices**: Apply prompt engineering best practices

Provide your refined prompts in this JSON format:
{
  "system_prompt": "The improved system prompt",
  "user_prompt": "The improved user prompt"
}

Make sure both prompts are complete and ready to use.`

	// Build user prompt for refinement
	refinementUserPrompt := fmt.Sprintf(`Please refine the following prompts based on these user requirements:

User Requirements:
---
%s
---

Current System Prompt:
---
%s
---

Current User Prompt:
---
%s
---

Please provide refined versions of both prompts that address the user's requirements.`, requirements, systemPrompt, userPrompt)

	// Call LLM API
	response, err := callLLMForRefineWithRequirements(refinementSystemPrompt, refinementUserPrompt, settings)
	if err != nil {
		logger.Errorf("Failed to call LLM for refinement with requirements: %v", err)
		return nil, fmt.Errorf("refinement failed: %w", err)
	}

	return response, nil
}

// callLLMForRefineWithRequirements calls the LLM API to refine prompts with requirements
func callLLMForRefineWithRequirements(systemPrompt, userPrompt string, settings models.Settings) (*RefineWithRequirementsResponse, error) {
	logger := log.GetLogger()

	// Prepare API request
	apiURL := settings.LlmBaseUrl + "/chat/completions"
	if apiURL == "/chat/completions" {
		// If base URL is empty, use default
		apiURL = "https://api.openai.com/v1/chat/completions"
	}

	requestBody := map[string]interface{}{
		"model": settings.LlmModel,
		"messages": []map[string]string{
			{"role": "system", "content": systemPrompt},
			{"role": "user", "content": userPrompt},
		},
		"temperature": 0.7,
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+settings.LlmApiKey)

	// Execute request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		logger.Errorf("LLM API returned status %d: %s", resp.StatusCode, string(body))
		return nil, fmt.Errorf("LLM API error: status %d", resp.StatusCode)
	}

	// Parse LLM response
	var llmResponse struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}

	err = json.Unmarshal(body, &llmResponse)
	if err != nil {
		return nil, fmt.Errorf("failed to parse LLM response: %w", err)
	}

	if len(llmResponse.Choices) == 0 {
		return nil, fmt.Errorf("no response from LLM")
	}

	// Extract and parse the JSON content from LLM
	content := llmResponse.Choices[0].Message.Content
	logger.Debugf("LLM refinement with requirements response: %s", content)

	// Extract JSON from response (handles text before/after JSON)
	jsonContent, err := extractJSON(content)
	if err != nil {
		logger.Warnf("Failed to extract JSON from LLM response: %v", err)
		logger.Debugf("Original content: %s", content)
		return nil, fmt.Errorf("failed to extract JSON from LLM response: %w", err)
	}

	// Parse the JSON response from LLM
	var refinementResp RefineWithRequirementsResponse
	err = json.Unmarshal([]byte(jsonContent), &refinementResp)
	if err != nil {
		logger.Warnf("Failed to parse LLM JSON response: %v", err)
		logger.Debugf("Extracted JSON content: %s", jsonContent)
		return nil, fmt.Errorf("failed to parse LLM response: %w", err)
	}

	return &refinementResp, nil
}

// callLLMForRefinement calls the LLM API to get refinement suggestions
func callLLMForRefinement(systemPrompt, userPrompt string, settings models.Settings) (*RefinementResponse, error) {
	logger := log.GetLogger()

	// Prepare API request
	apiURL := settings.LlmBaseUrl + "/chat/completions"
	if apiURL == "/chat/completions" {
		// If base URL is empty, use default
		apiURL = "https://api.openai.com/v1/chat/completions"
	}

	requestBody := map[string]interface{}{
		"model": settings.LlmModel,
		"messages": []map[string]string{
			{"role": "system", "content": systemPrompt},
			{"role": "user", "content": userPrompt},
		},
		"temperature": 0.7, // Slightly creative but focused
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+settings.LlmApiKey)

	// Execute request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		logger.Errorf("LLM API returned status %d: %s", resp.StatusCode, string(body))
		return nil, fmt.Errorf("LLM API error: status %d", resp.StatusCode)
	}

	// Parse LLM response
	var llmResponse struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}

	err = json.Unmarshal(body, &llmResponse)
	if err != nil {
		return nil, fmt.Errorf("failed to parse LLM response: %w", err)
	}

	if len(llmResponse.Choices) == 0 {
		return nil, fmt.Errorf("no response from LLM")
	}

	// Extract and parse the JSON content from LLM
	content := llmResponse.Choices[0].Message.Content
	logger.Debugf("LLM refinement response: %s", content)

	// Extract JSON from response (handles text before/after JSON)
	jsonContent, err := extractJSON(content)
	if err != nil {
		// If JSON extraction fails, create a basic response
		logger.Warnf("Failed to extract JSON from LLM response, creating basic response: %v", err)
		logger.Debugf("Original content: %s", content)
		return &RefinementResponse{
			OriginalPrompt:  userPrompt,
			RefinedPrompt:   content, // Use the raw content as refined prompt
			Assessment:      "The LLM provided suggestions but in an unexpected format",
			Suggestions:     []RefinementSuggestion{},
			QualityBefore:   6.0,
			QualityAfter:    7.5,
			ImprovementTips: []string{"Review the refined prompt and adapt as needed"},
		}, nil
	}

	// Parse the JSON response from LLM
	var refinementResp RefinementResponse
	err = json.Unmarshal([]byte(jsonContent), &refinementResp)
	if err != nil {
		// If JSON parsing fails, create a basic response
		logger.Warnf("Failed to parse LLM JSON response, creating basic response: %v", err)
		logger.Debugf("Extracted JSON content: %s", jsonContent)
		return &RefinementResponse{
			OriginalPrompt:  userPrompt,
			RefinedPrompt:   content, // Use the raw content as refined prompt
			Assessment:      "The LLM provided suggestions but in an unexpected format",
			Suggestions:     []RefinementSuggestion{},
			QualityBefore:   6.0,
			QualityAfter:    7.5,
			ImprovementTips: []string{"Review the refined prompt and adapt as needed"},
		}, nil
	}

	// Ensure original prompt is set
	refinementResp.OriginalPrompt = userPrompt

	return &refinementResp, nil
}

// AutoFillFrameworkFields uses LLM to auto-fill framework fields based on user input
func AutoFillFrameworkFields(framework *Framework, userInput string, settings models.Settings) (map[string]string, error) {
	logger := log.GetLogger()

	// Build field descriptions for the prompt with IDs and examples
	fieldDescriptions := ""
	exampleFields := make(map[string]string)
	for _, field := range framework.Fields {
		required := ""
		if field.Required {
			required = " (REQUIRED)"
		}
		fieldDescriptions += fmt.Sprintf("- Field ID: \"%s\" - Label: \"%s\"%s\n  Description: %s\n",
			field.ID, field.Label, required, field.Description)
		exampleFields[field.ID] = fmt.Sprintf("value for %s", field.Label)
	}

	// Create example JSON to show exact format expected
	exampleJSON, _ := json.MarshalIndent(map[string]interface{}{
		"fields": exampleFields,
	}, "", "  ")
	//TODO: move the prompt to prompt config table
	// Build system prompt
	systemPrompt := fmt.Sprintf(`You are an expert prompt engineer helping users create effective prompts using the %s framework.

The %s framework has the following fields:
%s

Your task is to analyze the user's requirement and generate appropriate values for each field based on their input.

CRITICAL INSTRUCTIONS:
1. You must use the exact Field IDs as keys in your JSON response, NOT the labels!
2. Return ONLY valid JSON in your response - NO explanatory text, NO thinking process, NO markdown
3. Do NOT include any text before or after the JSON object

Example response format (return EXACTLY this structure):
%s

Guidelines:
- Use Field IDs (like "capacity", "request") as JSON keys, NOT labels (like "Capacity/Role", "Request")
- All REQUIRED fields must have meaningful, non-empty values
- Values should be specific and relevant to the user's requirement
- Keep values concise but informative
- Use the field descriptions to understand what each field expects
- If a field seems irrelevant, provide a reasonable default or placeholder

Remember: Your response must be PURE JSON ONLY, starting with { and ending with }`,
		framework.Name, framework.Name, fieldDescriptions, string(exampleJSON))

	// Build user prompt
	userPrompt := fmt.Sprintf(`Based on this user requirement, please fill in the framework fields:

User Requirement:
---
%s
---

Please generate appropriate values for each field in the %s framework.`, userInput, framework.Name)

	// Call LLM API
	response, err := callLLMForAutoFill(systemPrompt, userPrompt, framework, settings)
	if err != nil {
		logger.Errorf("Failed to call LLM for auto-fill: %v", err)
		return nil, fmt.Errorf("auto-fill failed: %w", err)
	}

	return response, nil
}

// callLLMForAutoFill calls the LLM API to auto-fill fields
func callLLMForAutoFill(systemPrompt, userPrompt string, framework *Framework, settings models.Settings) (map[string]string, error) {
	logger := log.GetLogger()

	// Prepare API request
	apiURL := settings.LlmBaseUrl + "/chat/completions"
	if apiURL == "/chat/completions" {
		apiURL = "https://api.openai.com/v1/chat/completions"
	}
	logger.Debugf("LLM API request body: %s, %s", systemPrompt, userPrompt)

	requestBody := map[string]interface{}{
		"model": settings.LlmModel,
		"messages": []map[string]string{
			{"role": "system", "content": systemPrompt},
			{"role": "user", "content": userPrompt},
		},
		"temperature": 0.7,
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+settings.LlmApiKey)

	// Execute request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		logger.Errorf("LLM API returned status %d: %s", resp.StatusCode, string(body))
		return nil, fmt.Errorf("LLM API error: status %d", resp.StatusCode)
	}

	// Parse LLM response
	var llmResponse struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}

	err = json.Unmarshal(body, &llmResponse)
	if err != nil {
		return nil, fmt.Errorf("failed to parse LLM response: %w", err)
	}

	if len(llmResponse.Choices) == 0 {
		return nil, fmt.Errorf("no response from LLM")
	}

	// Extract and parse the JSON content from LLM
	content := llmResponse.Choices[0].Message.Content
	logger.Debugf("LLM auto-fill response: %s", content)

	// Extract JSON from response (handles text before/after JSON)
	jsonContent, err := extractJSON(content)
	if err != nil {
		logger.Warnf("Failed to extract JSON from LLM response: %v", err)
		logger.Debugf("Original content: %s", content)
		// Try to provide basic fallback values
		fields := make(map[string]string)
		for _, field := range framework.Fields {
			fields[field.ID] = ""
		}
		return fields, nil
	}

	// Parse the JSON response from LLM
	var autoFillResp struct {
		Fields map[string]string `json:"fields"`
	}
	err = json.Unmarshal([]byte(jsonContent), &autoFillResp)
	if err != nil {
		logger.Warnf("Failed to parse LLM JSON response: %v", err)
		logger.Debugf("Extracted JSON content: %s", jsonContent)
		// Try to provide basic fallback values
		fields := make(map[string]string)
		for _, field := range framework.Fields {
			fields[field.ID] = ""
		}
		return fields, nil
	}

	// Normalize field keys: map labels to IDs if LLM used labels instead
	normalizedFields := make(map[string]string)
	labelToIDMap := make(map[string]string)

	// Build label-to-ID mapping with fuzzy matching
	for _, field := range framework.Fields {
		// Exact match
		labelToIDMap[field.Label] = field.ID
		// Lowercase match
		labelToIDMap[strings.ToLower(field.Label)] = field.ID
		// Normalized match (remove special chars and spaces)
		normalized := strings.ToLower(strings.ReplaceAll(strings.ReplaceAll(field.Label, "/", ""), " ", ""))
		labelToIDMap[normalized] = field.ID
	}

	// Try to normalize keys
	for key, value := range autoFillResp.Fields {
		// Check if key is already a valid field ID
		isValidID := false
		for _, field := range framework.Fields {
			if field.ID == key {
				isValidID = true
				normalizedFields[key] = value
				break
			}
		}

		// If not a valid ID, try to map from label with fuzzy matching
		if !isValidID {
			// Try exact match
			if fieldID, exists := labelToIDMap[key]; exists {
				logger.Infof("Mapped label '%s' to field ID '%s'", key, fieldID)
				normalizedFields[fieldID] = value
			} else {
				// Try normalized match (remove special chars)
				normalizedKey := strings.ToLower(strings.ReplaceAll(strings.ReplaceAll(key, "/", ""), " ", ""))
				if fieldID, exists := labelToIDMap[normalizedKey]; exists {
					logger.Infof("Mapped normalized label '%s' (from '%s') to field ID '%s'", normalizedKey, key, fieldID)
					normalizedFields[fieldID] = value
				} else {
					logger.Warnf("Unknown field key '%s' in LLM response, ignoring", key)
				}
			}
		}
	}

	// Ensure all required fields have values
	for _, field := range framework.Fields {
		if field.Required {
			if val, exists := normalizedFields[field.ID]; !exists || val == "" {
				logger.Warnf("Required field %s is missing or empty, using placeholder", field.ID)
				normalizedFields[field.ID] = fmt.Sprintf("[Please specify %s]", field.Label)
			}
		}
	}

	return normalizedFields, nil
}

// stripMarkdownCodeBlocks removes markdown code block markers from JSON content
// Handles formats like:
//
//	```json\n{...}\n```
//	```\n{...}\n```
//	{... } (no markers)
func stripMarkdownCodeBlocks(content string) string {
	// Trim whitespace
	content = strings.TrimSpace(content)

	// Remove opening markdown code block with optional language
	if strings.HasPrefix(content, "```") {
		// Find the first newline after ```
		firstNewline := strings.Index(content, "\n")
		if firstNewline > 0 {
			content = content[firstNewline+1:]
		}
	}

	// Remove closing markdown code block
	if strings.HasSuffix(content, "```") {
		content = strings.TrimSuffix(content, "```")
	}

	// Trim again after removing markers
	return strings.TrimSpace(content)
}

// extractJSON attempts to extract JSON object from LLM response
// Handles cases where LLM includes explanatory text before/after JSON
func extractJSON(content string) (string, error) {
	logger := log.GetLogger()

	// First try stripping markdown code blocks
	content = stripMarkdownCodeBlocks(content)

	// Try to parse as-is first
	var testJSON map[string]interface{}
	if err := json.Unmarshal([]byte(content), &testJSON); err == nil {
		return content, nil
	}

	// Look for JSON object by finding valid JSON start patterns
	// Common patterns: {\n, { ", {", etc.
	// We need to find a '{' that's likely the start of a JSON object, not a template placeholder like {{topic}}

	// Strategy: Look for patterns that indicate JSON object start
	jsonStartPatterns := []string{
		"{\n  \"", // Pretty-printed JSON
		"{\n\"",   // Compact pretty-printed
		"{\"",     // Minified JSON
		"{ \"",    // JSON with space after brace
		"{\n\t\"", // Tab-indented JSON
	}

	startIdx := -1
	for _, pattern := range jsonStartPatterns {
		idx := strings.Index(content, pattern)
		if idx != -1 {
			// Found a potential JSON start, verify it's at the beginning of the object
			startIdx = idx
			break
		}
	}

	// If no pattern found, fall back to looking for standalone '{' followed by quote on next non-whitespace
	if startIdx == -1 {
		for i := 0; i < len(content); i++ {
			if content[i] == '{' {
				// Check if this could be the start of a JSON object
				// Skip whitespace after '{'
				j := i + 1
				for j < len(content) && (content[j] == ' ' || content[j] == '\n' || content[j] == '\t' || content[j] == '\r') {
					j++
				}
				// If next non-whitespace is '"', this is likely JSON start
				if j < len(content) && content[j] == '"' {
					startIdx = i
					break
				}
			}
		}
	}

	if startIdx == -1 {
		logger.Warnf("No JSON object found in content (length: %d)", len(content))
		if len(content) < 500 {
			logger.Debugf("Content: %s", content)
		}
		return "", fmt.Errorf("no JSON object found in response")
	}

	// Now find the matching closing brace by counting brace depth
	braceDepth := 0
	inString := false
	escaped := false
	endIdx := -1

	for i := startIdx; i < len(content); i++ {
		char := content[i]

		if escaped {
			escaped = false
			continue
		}

		if char == '\\' {
			escaped = true
			continue
		}

		if char == '"' {
			inString = !inString
			continue
		}

		if inString {
			continue
		}

		if char == '{' {
			braceDepth++
		} else if char == '}' {
			braceDepth--
			if braceDepth == 0 {
				endIdx = i
				break
			}
		}
	}

	if endIdx == -1 || endIdx < startIdx {
		logger.Warnf("Malformed JSON in content: couldn't find matching closing brace")
		return "", fmt.Errorf("malformed JSON in response")
	}

	// Extract the JSON portion
	jsonContent := content[startIdx : endIdx+1]

	// Validate it's proper JSON
	if err := json.Unmarshal([]byte(jsonContent), &testJSON); err != nil {
		logger.Warnf("Extracted content is not valid JSON, error: %v", err)
		if len(jsonContent) < 500 {
			logger.Debugf("Extracted content: %s", jsonContent)
		}
		return "", fmt.Errorf("invalid JSON: %w", err)
	}

	logger.Debugf("Successfully extracted JSON from LLM response")
	return jsonContent, nil
}

// QuickRefineWithFramework provides quick refinement suggestions based on framework
func QuickRefineWithFramework(promptText string, frameworkID string) []string {
	suggestions := []string{}

	framework := GetFrameworkByID(frameworkID)
	if framework == nil {
		return suggestions
	}

	// Provide framework-specific suggestions
	switch frameworkID {
	case "crispe":
		suggestions = append(suggestions, "✓ Consider adding a specific role/capacity (e.g., 'senior developer with 10 years experience')")
		suggestions = append(suggestions, "✓ Break down the task into clear steps")
		suggestions = append(suggestions, "✓ Specify performance requirements (error handling, testing, documentation)")
		suggestions = append(suggestions, "✓ Include an example if similar code exists")

	case "risen":
		suggestions = append(suggestions, "✓ Define the AI's role clearly at the start")
		suggestions = append(suggestions, "✓ Provide step-by-step instructions")
		suggestions = append(suggestions, "✓ Clearly state the end goal and deliverables")
		suggestions = append(suggestions, "✓ Add specific constraints and limitations")

	case "costar":
		suggestions = append(suggestions, "✓ Provide comprehensive context and background")
		suggestions = append(suggestions, "✓ Define the objective/goal explicitly")
		suggestions = append(suggestions, "✓ Specify the style and tone (professional, casual, etc.)")
		suggestions = append(suggestions, "✓ Identify the target audience")
		suggestions = append(suggestions, "✓ Describe the expected response format in detail")

	case "ape":
		suggestions = append(suggestions, "✓ State the action to be performed clearly")
		suggestions = append(suggestions, "✓ Explain why you're doing this (the purpose/problem)")
		suggestions = append(suggestions, "✓ Define what success looks like with specific criteria")

	case "car":
		suggestions = append(suggestions, "✓ Provide technical context (language, framework, tools)")
		suggestions = append(suggestions, "✓ Clearly define the action/task")
		suggestions = append(suggestions, "✓ Specify acceptance criteria and expected results")
	}

	// General suggestions for all frameworks
	suggestions = append(suggestions, "✓ Include specific technology versions")
	suggestions = append(suggestions, "✓ Mention testing requirements")
	suggestions = append(suggestions, "✓ Request error handling")

	return suggestions
}

// EstimatePromptQuality provides a quick quality estimate (0-10)
func EstimatePromptQuality(promptText string) float64 {
	score := 0.0

	// Length check (longer prompts tend to be more detailed)
	if len(promptText) > 100 {
		score += 1.0
	}
	if len(promptText) > 300 {
		score += 1.0
	}

	// Check for technical terms
	technicalTerms := []string{"framework", "database", "API", "test", "error", "authentication", "validation"}
	for _, term := range technicalTerms {
		if containsIgnoreCase(promptText, term) {
			score += 0.3
			if score > 10 {
				break
			}
		}
	}

	// Check for specific requirements
	requirementWords := []string{"should", "must", "include", "implement", "create", "generate"}
	for _, word := range requirementWords {
		if containsIgnoreCase(promptText, word) {
			score += 0.2
			if score > 10 {
				break
			}
		}
	}

	// Check for examples
	if containsIgnoreCase(promptText, "example") || containsIgnoreCase(promptText, "like") {
		score += 0.5
	}

	// Check for constraints
	constraintWords := []string{"using", "with", "without", "constraint", "limit", "requirement"}
	for _, word := range constraintWords {
		if containsIgnoreCase(promptText, word) {
			score += 0.3
			if score > 10 {
				break
			}
		}
	}

	// Cap at 10
	if score > 10 {
		score = 10
	}

	return score
}

// Helper function to check if string contains substring (case-insensitive)
func containsIgnoreCase(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr ||
		len(s) > 0 && len(substr) > 0 &&
			contains(toLower(s), toLower(substr)))
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && findSubstring(s, substr) >= 0
}

func findSubstring(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

func toLower(s string) string {
	result := make([]byte, len(s))
	for i := 0; i < len(s); i++ {
		if s[i] >= 'A' && s[i] <= 'Z' {
			result[i] = s[i] + 32
		} else {
			result[i] = s[i]
		}
	}
	return string(result)
}
