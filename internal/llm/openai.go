package llm

import (
	"bufio"
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/walterfan/lazy-ai-coder/internal/log"
	"github.com/walterfan/lazy-ai-coder/internal/metrics"
)

type LLMSettings struct {
	BaseUrl     string  `json:"base_url"`
	ApiKey      string  `json:"api_key"`
	Model       string  `json:"model"`
	Temperature float64 `json:"temperature"`
}

type ChatRequest struct {
	Model       string      `json:"model"`
	Messages    []ChatEntry `json:"messages"`
	Stream      bool        `json:"stream"`
	Temperature float64     `json:"temperature,omitempty"` // Default: 1.0
	//TopP        float64     `json:"top_p,omitempty"`       // Default: 1.0
	//MaxTokens   int         `json:"max_tokens,omitempty"`  // Default: 4096
}

type ChatEntry struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatResponse struct {
	Choices []struct {
		Message ChatEntry `json:"message"`
	} `json:"choices"`
}

func createClient() (*http.Client, error) {
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	return &http.Client{Transport: transport}, nil
}

func buildChatRequest(systemPrompt, userPrompt, model string, stream bool, temperature float64) (*ChatRequest, error) {
	return &ChatRequest{
		Model:  model,
		Stream: stream,
		Messages: []ChatEntry{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: userPrompt},
		},
		Temperature: temperature,
	}, nil
}

func loadBaseConfig() (string, string, string, float64, error) {
	baseUrl := os.Getenv("LLM_BASE_URL")
	apiKey := os.Getenv("LLM_API_KEY")
	model := os.Getenv("LLM_MODEL")

	temperatureStr := os.Getenv("LLM_TEMPERATURE")
	if temperatureStr == "" {
		temperatureStr = "1.0"
	}
	temperature, err := strconv.ParseFloat(temperatureStr, 64)
	if err != nil {
		temperature = 1.0
	}

	return baseUrl, apiKey, model, temperature, nil
}
func AskLLM(systemPrompt string, userPrompt string, settings LLMSettings) (string, error) {
	logger := log.GetLogger()

	// Use settings from web request first, only fall back to environment variables if empty
	baseUrl := settings.BaseUrl
	apiKey := settings.ApiKey
	model := settings.Model
	temperature := settings.Temperature

	// Only use environment variables as fallback when web settings are empty
	if baseUrl == "" {
		baseUrl = os.Getenv("LLM_BASE_URL")
	}
	if apiKey == "" {
		apiKey = os.Getenv("LLM_API_KEY")
	}
	if model == "" {
		model = os.Getenv("LLM_MODEL")
	}
	if temperature == 0 {
		temperatureStr := os.Getenv("LLM_TEMPERATURE")
		if temperatureStr == "" {
			temperatureStr = "1.0"
		}
		temperature, _ = strconv.ParseFloat(temperatureStr, 64)
		if temperature == 0 {
			temperature = 1.0
		}
	}

	logger.Infof("Using LLM settings - BaseUrl: %s, Model: %s, Temperature: %.1f", baseUrl, model, temperature)

	// Start timing for metrics
	startTime := time.Now()

	// Log detailed LLM request
	logger.With(
		"model", model,
		"base_url", baseUrl,
		"temperature", temperature,
		"system_prompt_length", len(systemPrompt),
		"user_prompt_length", len(userPrompt),
		"stream", false,
	).Info("LLM request initiated")

	req, err := buildChatRequest(systemPrompt, userPrompt, model, false, temperature)
	if err != nil {
		metrics.RecordLLMError(model, "build_request_error")
		logger.With("model", model, "error", err.Error()).Error("Failed to build LLM request")
		return "", err
	}

	body, _ := json.Marshal(req)

	httpReq, err := http.NewRequest("POST", fmt.Sprintf("%s/chat/completions", baseUrl), bytes.NewBuffer(body))
	if err != nil {
		metrics.RecordLLMError(model, "http_request_error")
		logger.With("model", model, "error", err.Error()).Error("Failed to create HTTP request")
		return "", err
	}
	httpReq.Header.Set("X-Api-Key", apiKey)
	httpReq.Header.Set("Authorization", "Bearer "+apiKey)
	httpReq.Header.Set("Content-Type", "application/json")

	client, err := createClient()
	if err != nil {
		metrics.RecordLLMError(model, "client_creation_error")
		return "", err
	}

	resp, err := client.Do(httpReq)
	if err != nil {
		duration := time.Since(startTime).Seconds()
		metrics.RecordLLMRequest(model, "openai", "error", duration, 0, 0)
		metrics.RecordLLMError(model, "api_call_error")
		logger.With(
			"model", model,
			"base_url", baseUrl,
			"duration_seconds", duration,
			"error", err.Error(),
		).Error("LLM API call failed")
		return "", err
	}
	defer resp.Body.Close()

	var out ChatResponse
	err = json.NewDecoder(resp.Body).Decode(&out)
	if err != nil {
		duration := time.Since(startTime).Seconds()
		metrics.RecordLLMRequest(model, "openai", "decode_error", duration, 0, 0)
		metrics.RecordLLMError(model, "decode_error")
		logger.With(
			"model", model,
			"duration_seconds", duration,
			"status_code", resp.StatusCode,
			"error", err.Error(),
		).Error("Failed to decode LLM response")
		return "", err
	}

	// Record successful request metrics
	duration := time.Since(startTime).Seconds()
	// Estimate tokens (rough approximation)
	promptTokens := float64(estimateTokenCount(systemPrompt + userPrompt))
	completionTokens := float64(estimateTokenCount(out.Choices[0].Message.Content))
	metrics.RecordLLMRequest(model, "openai", "success", duration, promptTokens, completionTokens)

	// Log detailed LLM response
	logger.With(
		"model", model,
		"duration_seconds", duration,
		"prompt_tokens_estimated", int(promptTokens),
		"completion_tokens_estimated", int(completionTokens),
		"response_length", len(out.Choices[0].Message.Content),
		"status", "success",
	).Info("LLM response received")

	// Log full request/response for debugging (only if DEBUG env var is set)
	if os.Getenv("LLM_DEBUG") == "true" {
		logger.With(
			"model", model,
			"system_prompt", systemPrompt,
			"user_prompt", userPrompt,
			"response", out.Choices[0].Message.Content,
			"duration_seconds", duration,
		).Debug("LLM full request/response")
	}

	return out.Choices[0].Message.Content, nil
}

func AskLLMWithStream(systemPrompt string, userPrompt string, settings LLMSettings, processChunk func(string)) error {
	logger := log.GetLogger()

	// Use settings from web request first, only fall back to environment variables if empty
	baseUrl := settings.BaseUrl
	apiKey := settings.ApiKey
	model := settings.Model
	temperature := settings.Temperature

	// Only use environment variables as fallback when web settings are empty
	if baseUrl == "" {
		baseUrl = os.Getenv("LLM_BASE_URL")
	}
	if apiKey == "" {
		apiKey = os.Getenv("LLM_API_KEY")
	}
	if model == "" {
		model = os.Getenv("LLM_MODEL")
	}
	if temperature == 0 {
		temperatureStr := os.Getenv("LLM_TEMPERATURE")
		if temperatureStr == "" {
			temperatureStr = "1.0"
		}
		temperature, _ = strconv.ParseFloat(temperatureStr, 64)
		if temperature == 0 {
			temperature = 1.0
		}
	}

	logger.Infof("Using LLM settings - BaseUrl: %s, Model: %s, Temperature: %.1f", baseUrl, model, temperature)

	fmt.Println("userPrompt:", userPrompt)
	req, err := buildChatRequest(systemPrompt, userPrompt, model, true, temperature)
	if err != nil {
		return err
	}

	body, _ := json.Marshal(req)

	httpReq, err := http.NewRequest("POST", fmt.Sprintf("%s/chat/completions", baseUrl), bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	httpReq.Header.Set("Authorization", "Bearer "+apiKey)
	httpReq.Header.Set("Content-Type", "application/json")

	client, err := createClient()
	if err != nil {
		return err
	}

	resp, err := client.Do(httpReq)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, errRead := io.ReadAll(resp.Body)
		if errRead != nil {
			logger.Errorf("Failed to read response body: %v", errRead)
		} else {
			logger.Errorf("LLM request failed with status code: %d, Response Body: %s", resp.StatusCode, bodyBytes)
		}
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	// Send opening tag first
	processChunk("<answer>")

	reader := bufio.NewReader(resp.Body)
	for {
		line, err := reader.ReadBytes('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			// Send closing tag on error before returning
			processChunk("</answer>")
			return err
		}

		trimmedLine := bytes.TrimSpace(line)
		if !bytes.HasPrefix(trimmedLine, []byte("data: ")) {
			continue
		}

		data := trimmedLine[6:]
		if len(data) == 0 || bytes.Equal(data, []byte("[DONE]")) {
			continue
		}

		var chunk map[string]interface{}
		if err := json.Unmarshal(data, &chunk); err != nil {
			logger.Errorf("JSON decode error: %v (raw data: %s)", err, data)
			continue
		}

		if choices, ok := chunk["choices"].([]interface{}); ok && len(choices) > 0 {
			if choice, ok := choices[0].(map[string]interface{}); ok {
				if delta, ok := choice["delta"].(map[string]interface{}); ok {
					if content, ok := delta["content"].(string); ok && content != "" {
						processChunk(content)
					}
				}
			}
		}
	}

	// Send closing tag at the end
	processChunk("</answer>")
	return nil
}

// ChatMessage represents a message in conversation history
type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// AskLLMWithMemory sends a request with conversation history
func AskLLMWithMemory(systemPrompt string, userPrompt string, history []ChatMessage, settings LLMSettings) (string, error) {
	logger := log.GetLogger()

	// Use settings from web request first, only fall back to environment variables if empty
	baseUrl := settings.BaseUrl
	apiKey := settings.ApiKey
	model := settings.Model
	temperature := settings.Temperature

	// Only use environment variables as fallback when web settings are empty
	if baseUrl == "" {
		baseUrl = os.Getenv("LLM_BASE_URL")
	}
	if apiKey == "" {
		apiKey = os.Getenv("LLM_API_KEY")
	}
	if model == "" {
		model = os.Getenv("LLM_MODEL")
	}
	if temperature == 0 {
		temperatureStr := os.Getenv("LLM_TEMPERATURE")
		if temperatureStr == "" {
			temperatureStr = "1.0"
		}
		temperature, _ = strconv.ParseFloat(temperatureStr, 64)
		if temperature == 0 {
			temperature = 1.0
		}
	}

	logger.Infof("Using LLM settings with memory - BaseUrl: %s, Model: %s, Temperature: %.1f, History: %d messages", baseUrl, model, temperature, len(history))

	// Build messages array starting with system prompt
	messages := []ChatEntry{{Role: "system", Content: systemPrompt}}

	// Add conversation history
	for _, msg := range history {
		messages = append(messages, ChatEntry{Role: msg.Role, Content: msg.Content})
	}

	// Add current user prompt
	messages = append(messages, ChatEntry{Role: "user", Content: userPrompt})

	req := &ChatRequest{
		Model:       model,
		Messages:    messages,
		Stream:      false,
		Temperature: temperature,
	}

	body, _ := json.Marshal(req)

	httpReq, err := http.NewRequest("POST", fmt.Sprintf("%s/chat/completions", baseUrl), bytes.NewBuffer(body))
	if err != nil {
		return "", err
	}
	httpReq.Header.Set("Authorization", "Bearer "+apiKey)
	httpReq.Header.Set("Content-Type", "application/json")

	client, err := createClient()
	if err != nil {
		return "", err
	}

	resp, err := client.Do(httpReq)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var out ChatResponse
	err = json.NewDecoder(resp.Body).Decode(&out)
	if err != nil {
		logger.Infof("Decode error: %v", err)
		return "", err
	}

	return out.Choices[0].Message.Content, nil
}

// AskLLMWithStreamAndMemory sends a streaming request with conversation history
func AskLLMWithStreamAndMemory(systemPrompt string, userPrompt string, history []ChatMessage, settings LLMSettings, processChunk func(string)) error {
	logger := log.GetLogger()

	// Use settings from web request first, only fall back to environment variables if empty
	baseUrl := settings.BaseUrl
	apiKey := settings.ApiKey
	model := settings.Model
	temperature := settings.Temperature

	// Only use environment variables as fallback when web settings are empty
	if baseUrl == "" {
		baseUrl = os.Getenv("LLM_BASE_URL")
	}
	if apiKey == "" {
		apiKey = os.Getenv("LLM_API_KEY")
	}
	if model == "" {
		model = os.Getenv("LLM_MODEL")
	}
	if temperature == 0 {
		temperatureStr := os.Getenv("LLM_TEMPERATURE")
		if temperatureStr == "" {
			temperatureStr = "1.0"
		}
		temperature, _ = strconv.ParseFloat(temperatureStr, 64)
		if temperature == 0 {
			temperature = 1.0
		}
	}

	logger.Infof("Using LLM settings with memory - BaseUrl: %s, Model: %s, Temperature: %.1f, History: %d messages", baseUrl, model, temperature, len(history))

	// Build messages array starting with system prompt
	messages := []ChatEntry{{Role: "system", Content: systemPrompt}}

	// Add conversation history
	for _, msg := range history {
		messages = append(messages, ChatEntry{Role: msg.Role, Content: msg.Content})
	}

	// Add current user prompt
	messages = append(messages, ChatEntry{Role: "user", Content: userPrompt})

	req := &ChatRequest{
		Model:       model,
		Messages:    messages,
		Stream:      true,
		Temperature: temperature,
	}

	body, _ := json.Marshal(req)

	httpReq, err := http.NewRequest("POST", fmt.Sprintf("%s/chat/completions", baseUrl), bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	httpReq.Header.Set("Authorization", "Bearer "+apiKey)
	httpReq.Header.Set("Content-Type", "application/json")

	client, err := createClient()
	if err != nil {
		return err
	}

	resp, err := client.Do(httpReq)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, errRead := io.ReadAll(resp.Body)
		if errRead != nil {
			logger.Errorf("Failed to read response body: %v", errRead)
		} else {
			logger.Errorf("LLM request failed with status code: %d, Response Body: %s", resp.StatusCode, bodyBytes)
		}
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	// Send opening tag first
	processChunk("<answer>")

	reader := bufio.NewReader(resp.Body)
	for {
		line, err := reader.ReadBytes('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			// Send closing tag on error before returning
			processChunk("</answer>")
			return err
		}

		trimmedLine := bytes.TrimSpace(line)
		if !bytes.HasPrefix(trimmedLine, []byte("data: ")) {
			continue
		}

		data := trimmedLine[6:]
		if len(data) == 0 || bytes.Equal(data, []byte("[DONE]")) {
			continue
		}

		var chunk map[string]interface{}
		if err := json.Unmarshal(data, &chunk); err != nil {
			logger.Errorf("JSON decode error: %v (raw data: %s)", err, data)
			continue
		}

		if choices, ok := chunk["choices"].([]interface{}); ok && len(choices) > 0 {
			if choice, ok := choices[0].(map[string]interface{}); ok {
				if delta, ok := choice["delta"].(map[string]interface{}); ok {
					if content, ok := delta["content"].(string); ok && content != "" {
						processChunk(content)
					}
				}
			}
		}
	}

	// Send closing tag at the end
	processChunk("</answer>")
	return nil
}

// estimateTokenCount provides a rough estimate of token count for metrics
// This is a simple approximation: ~4 characters per token for English text
func estimateTokenCount(text string) int {
	charCount := len(text)
	tokenCount := charCount / 4
	if tokenCount < 1 && charCount > 0 {
		tokenCount = 1
	}
	return tokenCount
}
