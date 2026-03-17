package chat

import (
	"fmt"
	"strings"

	"github.com/walterfan/lazy-ai-coder/internal/llm"
	"github.com/walterfan/lazy-ai-coder/internal/log"
	"github.com/walterfan/lazy-ai-coder/internal/mem"
	"github.com/walterfan/lazy-ai-coder/internal/models"
	"github.com/walterfan/lazy-ai-coder/internal/util"
)

// ChatService handles chat-related business logic
type ChatService struct{}

// NewChatService creates a new ChatService
func NewChatService() *ChatService {
	return &ChatService{}
}

// PromptResult contains the built prompt text and conversation history
type PromptResult struct {
	PromptText string
	History    []mem.ChatMessage
}

// BuildPromptTextWithMemory builds the prompt text and retrieves conversation history if memory is enabled
func (s *ChatService) BuildPromptTextWithMemory(req *models.ChatAPIRequest) (*PromptResult, error) {
	promptText := req.Prompt.UserPrompt

	// Handle multiple output languages
	var outputLanguageText string
	if len(req.OutputLanguages) > 0 {
		if len(req.OutputLanguages) == 1 {
			outputLanguageText = req.OutputLanguages[0]
		} else {
			outputLanguageText = "both " + strings.Join(req.OutputLanguages, " and ")
			promptText += "\n\nnote: please provide the response in both languages: " + strings.Join(req.OutputLanguages, " and ") + ". Please format as:\n\n**English:**\n[English content]\n\n**Chinese:**\n[Chinese content]"
		}
	} else if req.OutputLanguage != "" {
		// Backward compatibility
		outputLanguageText = req.OutputLanguage
	} else {
		outputLanguageText = "English"
	}

	// Append note if {{output_language}} not present
	if !strings.Contains(promptText, "{{output_language}}") && len(req.OutputLanguages) <= 1 {
		promptText += "\n\nnote: please use " + outputLanguageText + " to output"
	}

	// Replace {{output_language}} with provided language(s)
	promptText = strings.ReplaceAll(promptText, "{{output_language}}", outputLanguageText)

	if req.ComputerLanguage != "" && strings.Contains(promptText, "{{language}}") {
		promptText = strings.ReplaceAll(promptText, "{{language}}", req.ComputerLanguage)
	}

	// Handle code insertion
	if req.GitlabMrId != "" && strings.Contains(promptText, "{{code}}") {
		// Read code from GitLab Merge Request
		// Token must be passed from frontend settings
		privateToken := req.Settings.GitlabToken
		if privateToken == "" {
			return nil, fmt.Errorf("GitLab token not configured. Please configure it in Settings page")
		}
		if req.Settings.GitlabUrl == "" {
			return nil, fmt.Errorf("GitLab URL not configured. Please configure it in Settings page")
		}
		code, err := util.GetMergeRequestChange(req.Settings.GitlabUrl, req.GitlabProject, req.GitlabMrId, privateToken)
		if err != nil {
			return nil, fmt.Errorf("failed to read GitLab Merge Request %s: %v", req.GitlabMrId, err)
		}
		log.GetLogger().Infof("Read code from GitLab MR: %s", code)
		promptText = strings.Replace(promptText, "{{code}}", code, -1)
	} else if req.CodePath != "" && strings.Contains(promptText, "{{code}}") {
		// Handle multiple file paths separated by commas
		filePaths := strings.Split(req.CodePath, ",")
		log.GetLogger().Infof("Read code from local: %s", req.CodePath)
		if strings.HasPrefix(strings.TrimSpace(filePaths[0]), "http") {
			// Handle multiple remote files
			code, err := util.ReadMultipleRemoteFiles(filePaths)
			if err != nil {
				return nil, err
			}
			promptText = strings.Replace(promptText, "{{code}}", code, -1)
		} else {
			// Handle multiple local files
			code, err := util.ReadMultipleLocalFiles(filePaths)
			if err != nil {
				return nil, err
			}
			promptText = strings.Replace(promptText, "{{code}}", code, -1)
		}
	} else if req.GitlabCodePath != "" && strings.Contains(promptText, "{{code}}") {
		// Handle multiple GitLab file paths separated by commas
		filePaths := strings.Split(req.GitlabCodePath, ",")
		// Token must be passed from frontend settings
		privateToken := req.Settings.GitlabToken
		if privateToken == "" {
			return nil, fmt.Errorf("GitLab token not configured. Please configure it in Settings page")
		}
		if req.Settings.GitlabUrl == "" {
			return nil, fmt.Errorf("GitLab URL not configured. Please configure it in Settings page")
		}
		log.GetLogger().Infof("Read code from gitlab: %s", req.GitlabCodePath)
		code, err := util.ReadMultipleGitLabFiles(req.Settings.GitlabUrl, req.GitlabProject, req.GitlabBranch, privateToken, filePaths)
		if err != nil {
			return nil, err
		}
		promptText = strings.Replace(promptText, "{{code}}", code, -1)
	}

	// Get conversation history if memory is enabled
	var history []mem.ChatMessage
	if req.Remember && req.SessionId != "" {
		memoryManager := mem.GetMemoryManager()
		session := memoryManager.GetSession(req.SessionId)

		// Check if summarization is needed
		llmSettings := util.ConvertToLLMSettings(req.Settings)
		if session.ShouldSummarize(memoryManager.MaxTokens, memoryManager.MaxMessages) {
			err := memoryManager.SummarizeOldMessages(req.SessionId, llmSettings)
			if err != nil {
				log.GetLogger().Warnf("Failed to summarize session %s: %v", req.SessionId, err)
			}
		}

		// Get recent messages within token limit (reserve space for current request)
		memoryMessages := session.GetRecentMessages(4000) // Reserve ~4000 tokens for current request

		// Convert to LLM format, excluding system messages (summaries)
		for _, msg := range memoryMessages {
			if msg.Role != "system" { // Don't include system messages (summaries) in history
				history = append(history, mem.ChatMessage{
					Role:    msg.Role,
					Content: msg.Content,
				})
			}
		}

		log.GetLogger().Infof("Retrieved %d messages from session %s", len(history), req.SessionId)
	}

	return &PromptResult{
		PromptText: promptText,
		History:    history,
	}, nil
}

// ProcessChatRequest processes a chat request and returns the answer
func (s *ChatService) ProcessChatRequest(req *models.ChatAPIRequest, promptResult *PromptResult) (string, error) {
	llmSettings := util.ConvertToLLMSettings(req.Settings)
	var answer string
	var err error

	// Store user message in memory if remember is enabled
	if req.Remember && req.SessionId != "" {
		memoryManager := mem.GetMemoryManager()
		session := memoryManager.GetSession(req.SessionId)
		session.AddMessage("user", promptResult.PromptText)
	}

	if req.Stream {
		if len(promptResult.History) > 0 {
			// Convert mem.ChatMessage to llm.ChatMessage
			llmHistory := make([]llm.ChatMessage, len(promptResult.History))
			for i, msg := range promptResult.History {
				llmHistory[i] = llm.ChatMessage{
					Role:    msg.Role,
					Content: msg.Content,
				}
			}
			err = llm.AskLLMWithStreamAndMemory(req.Prompt.SystemPrompt, promptResult.PromptText, llmHistory, llmSettings, func(chunk string) {
				answer += chunk
			})
		} else {
			err = llm.AskLLMWithStream(req.Prompt.SystemPrompt, promptResult.PromptText, llmSettings, func(chunk string) {
				answer += chunk
			})
		}
	} else {
		if len(promptResult.History) > 0 {
			// Convert mem.ChatMessage to llm.ChatMessage
			llmHistory := make([]llm.ChatMessage, len(promptResult.History))
			for i, msg := range promptResult.History {
				llmHistory[i] = llm.ChatMessage{
					Role:    msg.Role,
					Content: msg.Content,
				}
			}
			answer, err = llm.AskLLMWithMemory(req.Prompt.SystemPrompt, promptResult.PromptText, llmHistory, llmSettings)
		} else {
			answer, err = llm.AskLLM(req.Prompt.SystemPrompt, promptResult.PromptText, llmSettings)
		}
	}

	// Store assistant response in memory if remember is enabled
	if req.Remember && req.SessionId != "" && err == nil {
		memoryManager := mem.GetMemoryManager()
		session := memoryManager.GetSession(req.SessionId)
		// Remove the <answer> tags before storing
		cleanAnswer := strings.TrimPrefix(answer, "<answer>")
		cleanAnswer = strings.TrimSuffix(cleanAnswer, "</answer>")
		session.AddMessage("assistant", cleanAnswer)
	}

	return answer, err
}

// ProcessStreamingChat processes a streaming chat request with a callback function
func (s *ChatService) ProcessStreamingChat(req *models.ChatAPIRequest, promptResult *PromptResult, callback func(string)) error {
	llmSettings := util.ConvertToLLMSettings(req.Settings)

	// Store user message in memory if remember is enabled
	if req.Remember && req.SessionId != "" {
		memoryManager := mem.GetMemoryManager()
		session := memoryManager.GetSession(req.SessionId)
		session.AddMessage("user", promptResult.PromptText)
	}

	// Collect the full answer for storage
	var fullAnswer strings.Builder

	// Wrap callback to collect answer
	wrappedCallback := func(chunk string) {
		fullAnswer.WriteString(chunk)
		callback(chunk)
	}

	var err error
	if len(promptResult.History) > 0 {
		// Convert mem.ChatMessage to llm.ChatMessage
		llmHistory := make([]llm.ChatMessage, len(promptResult.History))
		for i, msg := range promptResult.History {
			llmHistory[i] = llm.ChatMessage{
				Role:    msg.Role,
				Content: msg.Content,
			}
		}
		err = llm.AskLLMWithStreamAndMemory(req.Prompt.SystemPrompt, promptResult.PromptText, llmHistory, llmSettings, wrappedCallback)
	} else {
		err = llm.AskLLMWithStream(req.Prompt.SystemPrompt, promptResult.PromptText, llmSettings, wrappedCallback)
	}

	// Store assistant response in memory if remember is enabled
	if req.Remember && req.SessionId != "" && err == nil {
		memoryManager := mem.GetMemoryManager()
		session := memoryManager.GetSession(req.SessionId)
		// Remove the <answer> tags before storing
		cleanAnswer := strings.TrimPrefix(fullAnswer.String(), "<answer>")
		cleanAnswer = strings.TrimSuffix(cleanAnswer, "</answer>")
		session.AddMessage("assistant", cleanAnswer)
	}

	return err
}
