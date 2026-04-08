package health

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/walterfan/lazy-ai-coder/internal/mem"
	"github.com/walterfan/lazy-ai-coder/pkg/database"
)

// Status represents the health status of a component
type Status string

const (
	StatusHealthy   Status = "healthy"
	StatusUnhealthy Status = "unhealthy"
	StatusDegraded  Status = "degraded"
)

// ComponentHealth represents the health status of a single component
type ComponentHealth struct {
	Name    string                 `json:"name"`
	Status  Status                 `json:"status"`
	Message string                 `json:"message,omitempty"`
	Details map[string]interface{} `json:"details,omitempty"`
}

// HealthCheckResponse represents the overall health check response
type HealthCheckResponse struct {
	Status     Status            `json:"status"`
	Timestamp  string            `json:"timestamp"`
	Version    string            `json:"version,omitempty"`
	Components []ComponentHealth `json:"components"`
}

// HealthChecker performs health checks on various components
type HealthChecker struct {
	version string
}

// NewHealthChecker creates a new health checker
func NewHealthChecker(version string) *HealthChecker {
	return &HealthChecker{
		version: version,
	}
}

// BasicHealthCheck performs a basic health check
func (hc *HealthChecker) BasicHealthCheck() *HealthCheckResponse {
	return &HealthCheckResponse{
		Status:    StatusHealthy,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Version:   hc.version,
	}
}

// DetailedHealthCheck performs a detailed health check of all components
func (hc *HealthChecker) DetailedHealthCheck() *HealthCheckResponse {
	components := []ComponentHealth{
		hc.checkDatabase(),
		hc.checkMemoryManager(),
	}

	// Determine overall status
	overallStatus := StatusHealthy
	for _, comp := range components {
		if comp.Status == StatusUnhealthy {
			overallStatus = StatusUnhealthy
			break
		} else if comp.Status == StatusDegraded && overallStatus == StatusHealthy {
			overallStatus = StatusDegraded
		}
	}

	return &HealthCheckResponse{
		Status:     overallStatus,
		Timestamp:  time.Now().UTC().Format(time.RFC3339),
		Version:    hc.version,
		Components: components,
	}
}

// checkDatabase checks database connectivity and basic operations
func (hc *HealthChecker) checkDatabase() ComponentHealth {
	db := database.GetDB()
	if db == nil {
		return ComponentHealth{
			Name:    "database",
			Status:  StatusUnhealthy,
			Message: "Database connection not initialized",
		}
	}

	// Get underlying SQL DB
	sqlDB, err := db.DB()
	if err != nil {
		return ComponentHealth{
			Name:    "database",
			Status:  StatusUnhealthy,
			Message: fmt.Sprintf("Failed to get database instance: %v", err),
		}
	}

	// Test connection with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := sqlDB.PingContext(ctx); err != nil {
		return ComponentHealth{
			Name:    "database",
			Status:  StatusUnhealthy,
			Message: fmt.Sprintf("Database ping failed: %v", err),
		}
	}

	// Get connection stats
	stats := sqlDB.Stats()
	details := map[string]interface{}{
		"open_connections": stats.OpenConnections,
		"in_use":           stats.InUse,
		"idle":             stats.Idle,
		"max_open":         stats.MaxOpenConnections,
	}

	// Check if we're running low on connections
	status := StatusHealthy
	message := "Database is healthy"
	if stats.MaxOpenConnections > 0 {
		utilization := float64(stats.InUse) / float64(stats.MaxOpenConnections)
		if utilization > 0.9 {
			status = StatusDegraded
			message = fmt.Sprintf("High connection utilization: %.0f%%", utilization*100)
		}
		details["utilization_percent"] = fmt.Sprintf("%.1f", utilization*100)
	}

	return ComponentHealth{
		Name:    "database",
		Status:  status,
		Message: message,
		Details: details,
	}
}

// checkMemoryManager checks the memory manager status
func (hc *HealthChecker) checkMemoryManager() ComponentHealth {
	memoryMgr := mem.GetMemoryManager()
	if memoryMgr == nil {
		return ComponentHealth{
			Name:    "memory_manager",
			Status:  StatusUnhealthy,
			Message: "Memory manager not initialized",
		}
	}

	// Get all sessions
	sessionStats := memoryMgr.GetAllSessions()

	totalMessages := 0
	totalTokens := int64(0)
	oldestSession := time.Now()
	newestSession := time.Time{}

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
	}

	details := map[string]interface{}{
		"active_sessions": len(sessionStats),
		"total_messages":  totalMessages,
		"total_tokens":    totalTokens,
		"max_tokens":      memoryMgr.MaxTokens,
		"max_messages":    memoryMgr.MaxMessages,
	}

	if len(sessionStats) > 0 {
		details["oldest_session_age_seconds"] = int(time.Since(oldestSession).Seconds())
		details["newest_session_age_seconds"] = int(time.Since(newestSession).Seconds())
		details["avg_messages_per_session"] = fmt.Sprintf("%.1f", float64(totalMessages)/float64(len(sessionStats)))
		details["avg_tokens_per_session"] = fmt.Sprintf("%.0f", float64(totalTokens)/float64(len(sessionStats)))
	}

	// Determine status based on session count
	status := StatusHealthy
	message := "Memory manager is healthy"

	if len(sessionStats) > 1000 {
		status = StatusDegraded
		message = fmt.Sprintf("High number of active sessions: %d", len(sessionStats))
	}

	return ComponentHealth{
		Name:    "memory_manager",
		Status:  status,
		Message: message,
		Details: details,
	}
}

// CheckLLMAPI tests LLM API connectivity by calling the /models endpoint.
func (hc *HealthChecker) CheckLLMAPI(baseURL, apiKey string) ComponentHealth {
	if baseURL == "" || apiKey == "" {
		return ComponentHealth{
			Name:    "llm_api",
			Status:  StatusUnhealthy,
			Message: "Base URL and API key are required",
		}
	}

	client := &http.Client{
		Timeout: 15 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	req, err := http.NewRequest("GET", baseURL+"/models", nil)
	if err != nil {
		return ComponentHealth{
			Name:    "llm_api",
			Status:  StatusUnhealthy,
			Message: fmt.Sprintf("Failed to create request: %v", err),
		}
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)

	resp, err := client.Do(req)
	if err != nil {
		return ComponentHealth{
			Name:    "llm_api",
			Status:  StatusUnhealthy,
			Message: fmt.Sprintf("Failed to connect: %v", err),
		}
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))

	if resp.StatusCode != http.StatusOK {
		return ComponentHealth{
			Name:    "llm_api",
			Status:  StatusUnhealthy,
			Message: fmt.Sprintf("LLM API returned status %d", resp.StatusCode),
			Details: map[string]interface{}{
				"base_url":    baseURL,
				"status_code": resp.StatusCode,
			},
		}
	}

	details := map[string]interface{}{"base_url": baseURL}
	var modelsResp struct {
		Data []struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	if json.Unmarshal(body, &modelsResp) == nil && len(modelsResp.Data) > 0 {
		modelIDs := make([]string, 0, min(len(modelsResp.Data), 10))
		for i, m := range modelsResp.Data {
			if i >= 10 {
				break
			}
			modelIDs = append(modelIDs, m.ID)
		}
		details["available_models"] = modelIDs
		details["model_count"] = len(modelsResp.Data)
	}

	return ComponentHealth{
		Name:    "llm_api",
		Status:  StatusHealthy,
		Message: "LLM API is accessible",
		Details: details,
	}
}

// CheckGitLabAPI tests GitLab API connectivity by calling /api/v4/user.
func (hc *HealthChecker) CheckGitLabAPI(baseURL, token string) ComponentHealth {
	if baseURL == "" || token == "" {
		return ComponentHealth{
			Name:    "gitlab_api",
			Status:  StatusUnhealthy,
			Message: "Base URL and token are required",
		}
	}

	client := &http.Client{
		Timeout: 15 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	req, err := http.NewRequest("GET", baseURL+"/api/v4/user", nil)
	if err != nil {
		return ComponentHealth{
			Name:    "gitlab_api",
			Status:  StatusUnhealthy,
			Message: fmt.Sprintf("Failed to create request: %v", err),
		}
	}
	req.Header.Set("PRIVATE-TOKEN", token)

	resp, err := client.Do(req)
	if err != nil {
		return ComponentHealth{
			Name:    "gitlab_api",
			Status:  StatusUnhealthy,
			Message: fmt.Sprintf("Failed to connect: %v", err),
		}
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))

	if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
		return ComponentHealth{
			Name:    "gitlab_api",
			Status:  StatusUnhealthy,
			Message: fmt.Sprintf("Authentication failed (status %d) — check your token", resp.StatusCode),
			Details: map[string]interface{}{"base_url": baseURL},
		}
	}

	if resp.StatusCode != http.StatusOK {
		return ComponentHealth{
			Name:    "gitlab_api",
			Status:  StatusUnhealthy,
			Message: fmt.Sprintf("GitLab API returned status %d", resp.StatusCode),
			Details: map[string]interface{}{"base_url": baseURL, "status_code": resp.StatusCode},
		}
	}

	details := map[string]interface{}{"base_url": baseURL}
	var userResp struct {
		Username string `json:"username"`
		Name     string `json:"name"`
		Email    string `json:"email"`
	}
	if json.Unmarshal(body, &userResp) == nil && userResp.Username != "" {
		details["username"] = userResp.Username
		details["name"] = userResp.Name
	}

	return ComponentHealth{
		Name:    "gitlab_api",
		Status:  StatusHealthy,
		Message: "GitLab API is accessible",
		Details: details,
	}
}
