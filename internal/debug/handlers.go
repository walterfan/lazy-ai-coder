package debug

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/walterfan/lazy-ai-coder/internal/llm"
	"github.com/walterfan/lazy-ai-coder/internal/mem"
)

// DebugHandlers handles debug-related HTTP requests
type DebugHandlers struct{}

// NewDebugHandlers creates a new DebugHandlers instance
func NewDebugHandlers() *DebugHandlers {
	return &DebugHandlers{}
}

// SessionSummary represents a summary of a session for list view
type SessionSummary struct {
	SessionID     string    `json:"session_id"`
	MessageCount  int       `json:"message_count"`
	TotalTokens   int       `json:"total_tokens"`
	CreatedAt     time.Time `json:"created_at"`
	LastActivity  time.Time `json:"last_activity"`
	Age           string    `json:"age"`
	IdleTime      string    `json:"idle_time"`
}

// SessionDetails represents detailed information about a session
type SessionDetails struct {
	SessionID    string               `json:"session_id"`
	MessageCount int                  `json:"message_count"`
	TotalTokens  int                  `json:"total_tokens"`
	CreatedAt    time.Time            `json:"created_at"`
	LastActivity time.Time            `json:"last_activity"`
	Age          string               `json:"age"`
	IdleTime     string               `json:"idle_time"`
	Messages     []mem.ChatMessage    `json:"messages"`
}

// ListSessions godoc
// @Summary List all memory sessions
// @Description Get a list of all active memory sessions with summary information
// @Tags debug
// @Produce json
// @Success 200 {array} SessionSummary
// @Router /api/v1/debug/memory/sessions [get]
func (h *DebugHandlers) ListSessions(c *gin.Context) {
	memoryMgr := mem.GetMemoryManager()
	sessionStats := memoryMgr.GetAllSessions()

	summaries := make([]SessionSummary, 0, len(sessionStats))
	now := time.Now()

	for _, stats := range sessionStats {
		sessionID := stats["session_id"].(string)
		messageCount := stats["message_count"].(int)
		totalTokens := stats["total_tokens"].(int)
		createdAt := stats["created_at"].(time.Time)
		lastActivity := stats["last_activity"].(time.Time)

		age := now.Sub(createdAt)
		idle := now.Sub(lastActivity)

		summaries = append(summaries, SessionSummary{
			SessionID:    sessionID,
			MessageCount: messageCount,
			TotalTokens:  totalTokens,
			CreatedAt:    createdAt,
			LastActivity: lastActivity,
			Age:          formatDuration(age),
			IdleTime:     formatDuration(idle),
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"total_sessions": len(summaries),
		"sessions":       summaries,
		"timestamp":      now.Format(time.RFC3339),
	})
}

// GetSession godoc
// @Summary Get detailed information about a specific session
// @Description Get detailed information including all messages for a specific session
// @Tags debug
// @Produce json
// @Param session_id path string true "Session ID"
// @Success 200 {object} SessionDetails
// @Failure 404 {object} map[string]string
// @Router /api/v1/debug/memory/sessions/{session_id} [get]
func (h *DebugHandlers) GetSession(c *gin.Context) {
	sessionID := c.Param("session_id")
	if sessionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "session_id is required"})
		return
	}

	memoryMgr := mem.GetMemoryManager()
	session := memoryMgr.GetSession(sessionID)
	if session == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Session not found"})
		return
	}

	now := time.Now()
	age := now.Sub(session.CreatedAt)
	idle := now.Sub(session.LastActivity)

	details := SessionDetails{
		SessionID:    session.SessionID,
		MessageCount: len(session.Messages),
		TotalTokens:  session.TotalTokens,
		CreatedAt:    session.CreatedAt,
		LastActivity: session.LastActivity,
		Age:          formatDuration(age),
		IdleTime:     formatDuration(idle),
		Messages:     session.Messages,
	}

	c.JSON(http.StatusOK, details)
}

// DeleteSession godoc
// @Summary Delete a specific memory session
// @Description Remove a session and all its messages from memory
// @Tags debug
// @Produce json
// @Param session_id path string true "Session ID"
// @Success 200 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /api/v1/debug/memory/sessions/{session_id} [delete]
func (h *DebugHandlers) DeleteSession(c *gin.Context) {
	sessionID := c.Param("session_id")
	if sessionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "session_id is required"})
		return
	}

	memoryMgr := mem.GetMemoryManager()

	// Check if session exists first
	session := memoryMgr.GetSession(sessionID)
	if session == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Session not found"})
		return
	}

	// Delete the session
	memoryMgr.DeleteSession(sessionID)

	c.JSON(http.StatusOK, gin.H{
		"message":    "Session deleted successfully",
		"session_id": sessionID,
	})
}

// TriggerSummarization godoc
// @Summary Trigger manual summarization for a session
// @Description Force summarization of old messages in a session
// @Tags debug
// @Produce json
// @Param session_id path string true "Session ID"
// @Success 200 {object} map[string]interface{}
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/debug/memory/sessions/{session_id}/summarize [post]
func (h *DebugHandlers) TriggerSummarization(c *gin.Context) {
	sessionID := c.Param("session_id")
	if sessionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "session_id is required"})
		return
	}

	memoryMgr := mem.GetMemoryManager()
	session := memoryMgr.GetSession(sessionID)
	if session == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Session not found"})
		return
	}

	// Store pre-summarization stats
	messagesBefore := len(session.Messages)
	tokensBefore := session.TotalTokens

	// Get LLM settings from environment/config
	llmSettings := h.getLLMSettings()

	// Trigger summarization
	err := memoryMgr.SummarizeOldMessages(sessionID, llmSettings)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Get updated session
	session = memoryMgr.GetSession(sessionID)

	c.JSON(http.StatusOK, gin.H{
		"message":        "Summarization completed",
		"session_id":     sessionID,
		"before": gin.H{
			"messages": messagesBefore,
			"tokens":   tokensBefore,
		},
		"after": gin.H{
			"messages": len(session.Messages),
			"tokens":   session.TotalTokens,
		},
		"reduction": gin.H{
			"messages": messagesBefore - len(session.Messages),
			"tokens":   tokensBefore - session.TotalTokens,
		},
	})
}

// GetMemoryStats godoc
// @Summary Get overall memory manager statistics
// @Description Get aggregated statistics across all sessions
// @Tags debug
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/debug/memory/stats [get]
func (h *DebugHandlers) GetMemoryStats(c *gin.Context) {
	memoryMgr := mem.GetMemoryManager()
	sessionStats := memoryMgr.GetAllSessions()

	totalMessages := 0
	totalTokens := int64(0)
	oldestSession := time.Now()
	newestSession := time.Time{}

	sessionsByAge := make(map[string]int)
	messageDistribution := make(map[string]int) // "0-10", "11-50", "51-100", "100+"
	tokenDistribution := make(map[string]int)   // "0-1000", "1001-5000", "5001-10000", "10000+"

	for _, stats := range sessionStats {
		messageCount := stats["message_count"].(int)
		tokens := stats["total_tokens"].(int)
		createdAt := stats["created_at"].(time.Time)

		totalMessages += messageCount
		totalTokens += int64(tokens)

		if createdAt.Before(oldestSession) {
			oldestSession = createdAt
		}
		if createdAt.After(newestSession) {
			newestSession = createdAt
		}

		// Categorize by age
		age := time.Since(createdAt)
		if age < time.Hour {
			sessionsByAge["< 1 hour"]++
		} else if age < 24*time.Hour {
			sessionsByAge["1-24 hours"]++
		} else if age < 7*24*time.Hour {
			sessionsByAge["1-7 days"]++
		} else {
			sessionsByAge["> 7 days"]++
		}

		// Categorize by message count
		if messageCount <= 10 {
			messageDistribution["0-10"]++
		} else if messageCount <= 50 {
			messageDistribution["11-50"]++
		} else if messageCount <= 100 {
			messageDistribution["51-100"]++
		} else {
			messageDistribution["100+"]++
		}

		// Categorize by token count
		if tokens <= 1000 {
			tokenDistribution["0-1000"]++
		} else if tokens <= 5000 {
			tokenDistribution["1001-5000"]++
		} else if tokens <= 10000 {
			tokenDistribution["5001-10000"]++
		} else {
			tokenDistribution["10000+"]++
		}
	}

	responseStats := gin.H{
		"total_sessions":  len(sessionStats),
		"total_messages":  totalMessages,
		"total_tokens":    totalTokens,
		"configuration": gin.H{
			"max_tokens":       memoryMgr.MaxTokens,
			"max_messages":     memoryMgr.MaxMessages,
			"summary_tokens":   memoryMgr.SummaryTokens,
			"session_timeout":  memoryMgr.SessionTimeout.String(),
		},
	}

	if len(sessionStats) > 0 {
		responseStats["averages"] = gin.H{
			"messages_per_session": float64(totalMessages) / float64(len(sessionStats)),
			"tokens_per_session":   float64(totalTokens) / float64(len(sessionStats)),
		}
		responseStats["oldest_session_age"] = formatDuration(time.Since(oldestSession))
		responseStats["newest_session_age"] = formatDuration(time.Since(newestSession))
	}

	responseStats["distribution"] = gin.H{
		"by_age":      sessionsByAge,
		"by_messages": messageDistribution,
		"by_tokens":   tokenDistribution,
	}

	c.JSON(http.StatusOK, responseStats)
}

// CleanupExpiredSessions godoc
// @Summary Trigger manual cleanup of expired sessions
// @Description Force cleanup of sessions that have exceeded the timeout
// @Tags debug
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/debug/memory/cleanup [post]
func (h *DebugHandlers) CleanupExpiredSessions(c *gin.Context) {
	memoryMgr := mem.GetMemoryManager()

	// Get session count before cleanup
	sessionsBefore := len(memoryMgr.GetAllSessions())

	// Trigger cleanup
	memoryMgr.CleanupExpiredSessions()

	// Get session count after cleanup
	sessionsAfter := len(memoryMgr.GetAllSessions())

	c.JSON(http.StatusOK, gin.H{
		"message":          "Cleanup completed",
		"sessions_before":  sessionsBefore,
		"sessions_after":   sessionsAfter,
		"sessions_removed": sessionsBefore - sessionsAfter,
		"timeout":          memoryMgr.SessionTimeout.String(),
	})
}

// formatDuration formats a duration into a human-readable string
func formatDuration(d time.Duration) string {
	if d < time.Minute {
		return d.Round(time.Second).String()
	}
	if d < time.Hour {
		return d.Round(time.Minute).String()
	}
	if d < 24*time.Hour {
		return d.Round(time.Hour).String()
	}
	days := int(d.Hours() / 24)
	hours := int(d.Hours()) % 24
	if hours > 0 {
		return fmt.Sprintf("%dd %dh", days, hours)
	}
	return fmt.Sprintf("%dd", days)
}

// getLLMSettings returns LLM settings from environment variables
func (h *DebugHandlers) getLLMSettings() llm.LLMSettings {
	temperature := 1.0
	if val := os.Getenv("LLM_TEMPERATURE"); val != "" {
		if parsed, err := strconv.ParseFloat(val, 64); err == nil {
			temperature = parsed
		}
	}

	return llm.LLMSettings{
		BaseUrl:     os.Getenv("LLM_BASE_URL"),
		ApiKey:      os.Getenv("LLM_API_KEY"),
		Model:       os.Getenv("LLM_MODEL"),
		Temperature: temperature,
	}
}
