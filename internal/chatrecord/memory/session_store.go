package memory

import (
	"sync"
	"time"

	"github.com/cloudwego/eino/schema"
)

// SessionMessage represents a message in session memory
type SessionMessage struct {
	Role      schema.RoleType `json:"role"`
	Content   string          `json:"content"`
	Timestamp time.Time       `json:"timestamp"`
}

// SessionStore defines the interface for session memory management
type SessionStore interface {
	// GetRecentMessages returns the most recent messages for a session
	GetRecentMessages(sessionID string) []SessionMessage

	// AddMessage adds a user-assistant message pair to the session
	AddMessage(sessionID string, userMsg, assistantMsg string) error

	// AddUserMessage adds just the user message
	AddUserMessage(sessionID string, content string) error

	// AddAssistantMessage adds just the assistant message
	AddAssistantMessage(sessionID string, content string) error

	// Clear removes all messages for a session
	Clear(sessionID string) error

	// GetSessionCount returns the number of messages in a session
	GetSessionCount(sessionID string) int

	// ToSchemaMessages converts session messages to Eino schema messages
	ToSchemaMessages(sessionID string) []*schema.Message
}

// InMemorySessionStore implements SessionStore with an in-memory sliding window
type InMemorySessionStore struct {
	sessions    map[string][]SessionMessage
	maxMessages int
	maxAge      time.Duration
	mu          sync.RWMutex
}

// NewInMemorySessionStore creates a new in-memory session store
func NewInMemorySessionStore(maxMessages int, maxAge time.Duration) *InMemorySessionStore {
	if maxMessages < 1 {
		maxMessages = 10 // default: keep last 10 messages
	}
	if maxAge == 0 {
		maxAge = 30 * time.Minute // default: 30 minute session timeout
	}

	store := &InMemorySessionStore{
		sessions:    make(map[string][]SessionMessage),
		maxMessages: maxMessages,
		maxAge:      maxAge,
	}

	// Start background cleanup goroutine
	go store.cleanupLoop()

	return store
}

// GetRecentMessages returns the most recent messages for a session
func (s *InMemorySessionStore) GetRecentMessages(sessionID string) []SessionMessage {
	s.mu.RLock()
	defer s.mu.RUnlock()

	messages, exists := s.sessions[sessionID]
	if !exists {
		return nil
	}

	// Filter out expired messages
	now := time.Now()
	var validMessages []SessionMessage
	for _, msg := range messages {
		if now.Sub(msg.Timestamp) <= s.maxAge {
			validMessages = append(validMessages, msg)
		}
	}

	return validMessages
}

// AddMessage adds a user-assistant message pair to the session
func (s *InMemorySessionStore) AddMessage(sessionID string, userMsg, assistantMsg string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()

	// Initialize session if needed
	if _, exists := s.sessions[sessionID]; !exists {
		s.sessions[sessionID] = make([]SessionMessage, 0, s.maxMessages)
	}

	// Add user message
	s.sessions[sessionID] = append(s.sessions[sessionID], SessionMessage{
		Role:      schema.User,
		Content:   userMsg,
		Timestamp: now,
	})

	// Add assistant message
	s.sessions[sessionID] = append(s.sessions[sessionID], SessionMessage{
		Role:      schema.Assistant,
		Content:   assistantMsg,
		Timestamp: now,
	})

	// Enforce sliding window
	s.evict(sessionID)

	return nil
}

// AddUserMessage adds just the user message
func (s *InMemorySessionStore) AddUserMessage(sessionID string, content string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.sessions[sessionID]; !exists {
		s.sessions[sessionID] = make([]SessionMessage, 0, s.maxMessages)
	}

	s.sessions[sessionID] = append(s.sessions[sessionID], SessionMessage{
		Role:      schema.User,
		Content:   content,
		Timestamp: time.Now(),
	})

	s.evict(sessionID)
	return nil
}

// AddAssistantMessage adds just the assistant message
func (s *InMemorySessionStore) AddAssistantMessage(sessionID string, content string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.sessions[sessionID]; !exists {
		s.sessions[sessionID] = make([]SessionMessage, 0, s.maxMessages)
	}

	s.sessions[sessionID] = append(s.sessions[sessionID], SessionMessage{
		Role:      schema.Assistant,
		Content:   content,
		Timestamp: time.Now(),
	})

	s.evict(sessionID)
	return nil
}

// Clear removes all messages for a session
func (s *InMemorySessionStore) Clear(sessionID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.sessions, sessionID)
	return nil
}

// GetSessionCount returns the number of messages in a session
func (s *InMemorySessionStore) GetSessionCount(sessionID string) int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	messages, exists := s.sessions[sessionID]
	if !exists {
		return 0
	}
	return len(messages)
}

// ToSchemaMessages converts session messages to Eino schema messages
func (s *InMemorySessionStore) ToSchemaMessages(sessionID string) []*schema.Message {
	messages := s.GetRecentMessages(sessionID)
	if len(messages) == 0 {
		return nil
	}

	schemaMessages := make([]*schema.Message, len(messages))
	for i, msg := range messages {
		schemaMessages[i] = &schema.Message{
			Role:    msg.Role,
			Content: msg.Content,
		}
	}
	return schemaMessages
}

// evict removes oldest messages to stay within maxMessages limit (must be called with lock held)
func (s *InMemorySessionStore) evict(sessionID string) {
	messages := s.sessions[sessionID]
	if len(messages) > s.maxMessages {
		// Keep only the most recent maxMessages
		s.sessions[sessionID] = messages[len(messages)-s.maxMessages:]
	}
}

// cleanupLoop periodically removes expired sessions
func (s *InMemorySessionStore) cleanupLoop() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		s.cleanup()
	}
}

// cleanup removes sessions with no recent activity
func (s *InMemorySessionStore) cleanup() {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	for sessionID, messages := range s.sessions {
		if len(messages) == 0 {
			delete(s.sessions, sessionID)
			continue
		}

		// Check if the most recent message is expired
		lastMsg := messages[len(messages)-1]
		if now.Sub(lastMsg.Timestamp) > s.maxAge {
			delete(s.sessions, sessionID)
		}
	}
}

// GetAllSessionIDs returns all active session IDs (for testing/debugging)
func (s *InMemorySessionStore) GetAllSessionIDs() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	ids := make([]string, 0, len(s.sessions))
	for id := range s.sessions {
		ids = append(ids, id)
	}
	return ids
}
