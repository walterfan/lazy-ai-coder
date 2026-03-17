package mem

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/walterfan/lazy-ai-coder/internal/llm"
	"github.com/walterfan/lazy-ai-coder/internal/log"
	"github.com/walterfan/lazy-ai-coder/internal/metrics"
)

// ChatMessage represents a single message in the conversation
type ChatMessage struct {
	Role      string    `json:"role"` // "user", "assistant", "system"
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
	Tokens    int       `json:"tokens"` // Estimated token count
}

// SessionMemory holds the conversation history for a session
type SessionMemory struct {
	SessionID    string        `json:"session_id"`
	Messages     []ChatMessage `json:"messages"`
	TotalTokens  int           `json:"total_tokens"`
	CreatedAt    time.Time     `json:"created_at"`
	LastActivity time.Time     `json:"last_activity"`
	mu           sync.RWMutex  `json:"-"`
}

// MemoryManager manages all session memories
type MemoryManager struct {
	sessions map[string]*SessionMemory
	mu       sync.RWMutex

	// Configuration
	MaxTokens      int           // Maximum tokens per session (default: 8000)
	MaxMessages    int           // Maximum messages per session (default: 50)
	SummaryTokens  int           // Target tokens for summary (default: 1000)
	SessionTimeout time.Duration // Session timeout (default: 24h)
}

// NewMemoryManager creates a new memory manager
func NewMemoryManager() *MemoryManager {
	return &MemoryManager{
		sessions:       make(map[string]*SessionMemory),
		MaxTokens:      8000,
		MaxMessages:    50,
		SummaryTokens:  1000,
		SessionTimeout: 24 * time.Hour,
	}
}

// Global memory manager instance
var globalMemoryManager = NewMemoryManager()

// GetMemoryManager returns the global memory manager
func GetMemoryManager() *MemoryManager {
	return globalMemoryManager
}

// GetSession retrieves or creates a session
func (mm *MemoryManager) GetSession(sessionID string) *SessionMemory {
	mm.mu.Lock()
	defer mm.mu.Unlock()

	session, exists := mm.sessions[sessionID]
	if !exists {
		session = &SessionMemory{
			SessionID:    sessionID,
			Messages:     make([]ChatMessage, 0),
			TotalTokens:  0,
			CreatedAt:    time.Now(),
			LastActivity: time.Now(),
		}
		mm.sessions[sessionID] = session
	}

	session.LastActivity = time.Now()
	return session
}

// AddMessage adds a message to the session
func (sm *SessionMemory) AddMessage(role, content string) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	tokens := EstimateTokens(content)
	message := ChatMessage{
		Role:      role,
		Content:   content,
		Timestamp: time.Now(),
		Tokens:    tokens,
	}

	sm.Messages = append(sm.Messages, message)
	sm.TotalTokens += tokens
	sm.LastActivity = time.Now()

	// Update memory metrics
	mm := GetMemoryManager()
	mm.updateMetrics()
}

// GetMessages returns all messages in the session
func (sm *SessionMemory) GetMessages() []ChatMessage {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	// Return a copy to avoid race conditions
	messages := make([]ChatMessage, len(sm.Messages))
	copy(messages, sm.Messages)
	return messages
}

// GetRecentMessages returns recent messages within token limit
func (sm *SessionMemory) GetRecentMessages(maxTokens int) []ChatMessage {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	if len(sm.Messages) == 0 {
		return []ChatMessage{}
	}

	var result []ChatMessage
	tokenCount := 0

	// Start from the most recent messages and work backwards
	for i := len(sm.Messages) - 1; i >= 0; i-- {
		message := sm.Messages[i]
		if tokenCount+message.Tokens > maxTokens && len(result) > 0 {
			break
		}
		result = append([]ChatMessage{message}, result...)
		tokenCount += message.Tokens
	}

	return result
}

// ShouldSummarize checks if the session needs summarization
func (sm *SessionMemory) ShouldSummarize(maxTokens, maxMessages int) bool {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	return sm.TotalTokens > maxTokens || len(sm.Messages) > maxMessages
}

// SummarizeOldMessages summarizes old messages to free up space
func (mm *MemoryManager) SummarizeOldMessages(sessionID string, llmSettings llm.LLMSettings) error {
	session := mm.GetSession(sessionID)

	if !session.ShouldSummarize(mm.MaxTokens, mm.MaxMessages) {
		return nil
	}

	session.mu.Lock()
	defer session.mu.Unlock()

	logger := log.GetLogger()
	logger.Infof("Summarizing session %s with %d messages and %d tokens", sessionID, len(session.Messages), session.TotalTokens)

	// Keep the most recent messages and summarize the older ones
	recentMessages := session.Messages[len(session.Messages)-10:] // Keep last 10 messages
	oldMessages := session.Messages[:len(session.Messages)-10]    // Summarize the rest

	if len(oldMessages) == 0 {
		return nil
	}

	// Build conversation text for summarization
	var conversationText strings.Builder
	for _, msg := range oldMessages {
		conversationText.WriteString(fmt.Sprintf("%s: %s\n", strings.Title(msg.Role), msg.Content))
	}

	// Create summarization prompt
	systemPrompt := "You are a conversation summarizer. Summarize the following conversation history while preserving key information, context, and important details. Keep the summary concise but comprehensive."
	userPrompt := fmt.Sprintf("Please summarize this conversation:\n\n%s\n\nSummary:", conversationText.String())

	// Get summary from LLM
	summary, err := llm.AskLLM(systemPrompt, userPrompt, llmSettings)
	if err != nil {
		logger.Errorf("Failed to summarize conversation: %v", err)
		return err
	}

	// Create summary message
	summaryMessage := ChatMessage{
		Role:      "system",
		Content:   fmt.Sprintf("[CONVERSATION SUMMARY]: %s", summary),
		Timestamp: time.Now(),
		Tokens:    EstimateTokens(summary),
	}

	// Replace old messages with summary + recent messages
	session.Messages = append([]ChatMessage{summaryMessage}, recentMessages...)

	// Recalculate total tokens
	session.TotalTokens = 0
	for _, msg := range session.Messages {
		session.TotalTokens += msg.Tokens
	}

	logger.Infof("Session %s summarized: %d messages, %d tokens", sessionID, len(session.Messages), session.TotalTokens)

	// Record summarization metric
	metrics.RecordMemorySummarization()
	mm.updateMetrics()

	return nil
}

// CleanupExpiredSessions removes old sessions
func (mm *MemoryManager) CleanupExpiredSessions() {
	mm.mu.Lock()
	defer mm.mu.Unlock()

	now := time.Now()
	cleanedCount := 0
	for sessionID, session := range mm.sessions {
		if now.Sub(session.LastActivity) > mm.SessionTimeout {
			// Record session duration before deletion
			duration := now.Sub(session.CreatedAt).Seconds()
			metrics.RecordMemorySessionDuration(duration)

			delete(mm.sessions, sessionID)
			log.GetLogger().Infof("Cleaned up expired session: %s", sessionID)
			cleanedCount++
		}
	}

	if cleanedCount > 0 {
		metrics.RecordMemoryCleanup()
		mm.updateMetrics()
	}
}

// EstimateTokens provides a rough estimate of token count
// This is a simple approximation: ~4 characters per token for English text
func EstimateTokens(text string) int {
	// Simple tokenization estimate
	// More accurate would be to use tiktoken or similar
	charCount := len(text)
	tokenCount := charCount / 4
	if tokenCount < 1 && charCount > 0 {
		tokenCount = 1
	}
	return tokenCount
}

// GetSessionStats returns session statistics
func (sm *SessionMemory) GetStats() map[string]interface{} {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	return map[string]interface{}{
		"session_id":    sm.SessionID,
		"message_count": len(sm.Messages),
		"total_tokens":  sm.TotalTokens,
		"created_at":    sm.CreatedAt,
		"last_activity": sm.LastActivity,
	}
}

// DeleteSession removes a session from memory
func (mm *MemoryManager) DeleteSession(sessionID string) {
	mm.mu.Lock()
	defer mm.mu.Unlock()

	// Record session duration if it exists
	if session, exists := mm.sessions[sessionID]; exists {
		duration := time.Since(session.CreatedAt).Seconds()
		metrics.RecordMemorySessionDuration(duration)
	}

	delete(mm.sessions, sessionID)
	mm.updateMetrics()
}

// GetAllSessions returns all session statistics
func (mm *MemoryManager) GetAllSessions() []map[string]interface{} {
	mm.mu.RLock()
	defer mm.mu.RUnlock()

	sessions := make([]map[string]interface{}, 0)
	for _, session := range mm.sessions {
		sessions = append(sessions, session.GetStats())
	}

	return sessions
}

// updateMetrics updates memory-related Prometheus metrics
func (mm *MemoryManager) updateMetrics() {
	mm.mu.RLock()
	defer mm.mu.RUnlock()

	totalMessages := 0
	totalTokens := int64(0)

	for _, session := range mm.sessions {
		session.mu.RLock()
		totalMessages += len(session.Messages)
		totalTokens += int64(session.TotalTokens)
		session.mu.RUnlock()
	}

	metrics.UpdateMemoryMetrics(len(mm.sessions), totalMessages, totalTokens)
}
