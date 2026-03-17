package util

import (
	"strconv"

	"github.com/walterfan/lazy-ai-coder/internal/llm"
	"github.com/walterfan/lazy-ai-coder/internal/models"
)

// ConvertToLLMSettings converts Settings to LLMSettings
func ConvertToLLMSettings(settings models.Settings) llm.LLMSettings {
	temperature := 1.0
	if settings.LlmTemperature != "" {
		if temp, err := strconv.ParseFloat(settings.LlmTemperature, 64); err == nil {
			temperature = temp
		}
	}

	return llm.LLMSettings{
		ApiKey:      settings.LlmApiKey,
		Model:       settings.LlmModel,
		BaseUrl:     settings.LlmBaseUrl,
		Temperature: temperature,
	}
}

