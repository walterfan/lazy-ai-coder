package mcp

import (
	"fmt"
	"sync"
	"time"

	"github.com/walterfan/lazy-ai-coder/internal/log"
	"go.uber.org/zap"
)

// SSEClient represents a connected SSE client
type SSEClient struct {
	ID       string
	Channel  chan SSEMessage
	LastSeen time.Time
}

// SSEMessage represents a message to send via SSE
type SSEMessage struct {
	Event string // Event type (e.g., "prompts/updated", "resources/updated")
	Data  string // JSON data
	ID    string // Optional message ID for client-side tracking
}

// SSEBroker manages SSE connections and broadcasts messages
type SSEBroker struct {
	clients    map[string]*SSEClient
	register   chan *SSEClient
	unregister chan *SSEClient
	broadcast  chan SSEMessage
	mu         sync.RWMutex
	logger     *zap.SugaredLogger
}

// NewSSEBroker creates a new SSE broker
func NewSSEBroker() *SSEBroker {
	return &SSEBroker{
		clients:    make(map[string]*SSEClient),
		register:   make(chan *SSEClient, 10),
		unregister: make(chan *SSEClient, 10),
		broadcast:  make(chan SSEMessage, 256),
		logger:     log.GetLogger(),
	}
}

// Run starts the SSE broker event loop
func (b *SSEBroker) Run() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	b.logger.Info("SSE broker started")

	for {
		select {
		case client := <-b.register:
			b.mu.Lock()
			b.clients[client.ID] = client
			b.mu.Unlock()
			b.logger.Infof("SSE client registered: %s (total: %d)", client.ID, len(b.clients))

		case client := <-b.unregister:
			b.mu.Lock()
			if _, ok := b.clients[client.ID]; ok {
				close(client.Channel)
				delete(b.clients, client.ID)
				b.logger.Infof("SSE client unregistered: %s (total: %d)", client.ID, len(b.clients))
			}
			b.mu.Unlock()

		case message := <-b.broadcast:
			b.mu.RLock()
			clientCount := len(b.clients)
			b.mu.RUnlock()

			if clientCount > 0 {
				b.logger.Infof("Broadcasting SSE message to %d clients: event=%s", clientCount, message.Event)
			}

			b.mu.RLock()
			for clientID, client := range b.clients {
				select {
				case client.Channel <- message:
					// Message sent successfully
				default:
					// Channel full or blocked, remove client
					b.logger.Warnf("SSE client %s channel blocked, removing", clientID)
					go func(c *SSEClient) {
						b.unregister <- c
					}(client)
				}
			}
			b.mu.RUnlock()

		case <-ticker.C:
			// Send heartbeat/keepalive to all clients
			b.sendHeartbeat()
		}
	}
}

// sendHeartbeat sends a keepalive message to all clients
func (b *SSEBroker) sendHeartbeat() {
	b.mu.RLock()
	defer b.mu.RUnlock()

	if len(b.clients) == 0 {
		return
	}

	heartbeat := SSEMessage{
		Event: "heartbeat",
		Data:  fmt.Sprintf(`{"timestamp":"%s"}`, time.Now().Format(time.RFC3339)),
	}

	for _, client := range b.clients {
		select {
		case client.Channel <- heartbeat:
			client.LastSeen = time.Now()
		default:
			// Skip if channel is full
		}
	}
}

// RegisterClient registers a new SSE client
func (b *SSEBroker) RegisterClient(client *SSEClient) {
	b.register <- client
}

// UnregisterClient removes an SSE client
func (b *SSEBroker) UnregisterClient(client *SSEClient) {
	b.unregister <- client
}

// Broadcast sends a message to all connected clients
func (b *SSEBroker) Broadcast(message SSEMessage) {
	select {
	case b.broadcast <- message:
		// Message queued successfully
	default:
		b.logger.Warn("SSE broadcast channel full, message dropped")
	}
}

// NotifyPromptsUpdated sends notification when prompts are updated
func (b *SSEBroker) NotifyPromptsUpdated(count int) {
	b.Broadcast(SSEMessage{
		Event: "prompts/updated",
		Data:  fmt.Sprintf(`{"message":"Prompts have been updated","count":%d,"timestamp":"%s"}`, count, time.Now().Format(time.RFC3339)),
		ID:    fmt.Sprintf("prompts-%d", time.Now().UnixNano()),
	})
}

// NotifyResourcesUpdated sends notification when resources are updated
func (b *SSEBroker) NotifyResourcesUpdated(count int) {
	b.Broadcast(SSEMessage{
		Event: "resources/updated",
		Data:  fmt.Sprintf(`{"message":"Resources have been updated","count":%d,"timestamp":"%s"}`, count, time.Now().Format(time.RFC3339)),
		ID:    fmt.Sprintf("resources-%d", time.Now().UnixNano()),
	})
}

// NotifyToolsUpdated sends notification when tools are updated
func (b *SSEBroker) NotifyToolsUpdated(count int) {
	b.Broadcast(SSEMessage{
		Event: "tools/updated",
		Data:  fmt.Sprintf(`{"message":"Tools have been updated","count":%d,"timestamp":"%s"}`, count, time.Now().Format(time.RFC3339)),
		ID:    fmt.Sprintf("tools-%d", time.Now().UnixNano()),
	})
}

// NotifyToolExecution sends notification about tool execution progress
func (b *SSEBroker) NotifyToolExecution(toolName string, status string, progress int, message string) {
	b.Broadcast(SSEMessage{
		Event: "tool/execution",
		Data:  fmt.Sprintf(`{"tool":"%s","status":"%s","progress":%d,"message":"%s","timestamp":"%s"}`, toolName, status, progress, message, time.Now().Format(time.RFC3339)),
		ID:    fmt.Sprintf("tool-%s-%d", toolName, time.Now().UnixNano()),
	})
}

// GetClientCount returns the number of connected clients
func (b *SSEBroker) GetClientCount() int {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return len(b.clients)
}
