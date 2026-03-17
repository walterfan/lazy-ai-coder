package memory

import (
	"testing"
	"time"

	"github.com/cloudwego/eino/schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewInMemorySessionStore(t *testing.T) {
	store := NewInMemorySessionStore(10, 30*time.Minute)
	assert.NotNil(t, store)
	assert.Equal(t, 10, store.maxMessages)
	assert.Equal(t, 30*time.Minute, store.maxAge)
}

func TestNewInMemorySessionStore_Defaults(t *testing.T) {
	store := NewInMemorySessionStore(0, 0)
	assert.Equal(t, 10, store.maxMessages)
	assert.Equal(t, 30*time.Minute, store.maxAge)
}

func TestInMemorySessionStore_AddMessage(t *testing.T) {
	store := NewInMemorySessionStore(10, 30*time.Minute)

	err := store.AddMessage("session-1", "Hello", "Hi there!")
	require.NoError(t, err)

	messages := store.GetRecentMessages("session-1")
	assert.Len(t, messages, 2)
	assert.Equal(t, schema.User, messages[0].Role)
	assert.Equal(t, "Hello", messages[0].Content)
	assert.Equal(t, schema.Assistant, messages[1].Role)
	assert.Equal(t, "Hi there!", messages[1].Content)
}

func TestInMemorySessionStore_AddUserMessage(t *testing.T) {
	store := NewInMemorySessionStore(10, 30*time.Minute)

	err := store.AddUserMessage("session-1", "Hello")
	require.NoError(t, err)

	messages := store.GetRecentMessages("session-1")
	assert.Len(t, messages, 1)
	assert.Equal(t, schema.User, messages[0].Role)
	assert.Equal(t, "Hello", messages[0].Content)
}

func TestInMemorySessionStore_AddAssistantMessage(t *testing.T) {
	store := NewInMemorySessionStore(10, 30*time.Minute)

	err := store.AddAssistantMessage("session-1", "Hi there!")
	require.NoError(t, err)

	messages := store.GetRecentMessages("session-1")
	assert.Len(t, messages, 1)
	assert.Equal(t, schema.Assistant, messages[0].Role)
	assert.Equal(t, "Hi there!", messages[0].Content)
}

func TestInMemorySessionStore_SlidingWindow(t *testing.T) {
	store := NewInMemorySessionStore(4, 30*time.Minute) // Max 4 messages

	// Add 3 message pairs (6 messages total)
	store.AddMessage("session-1", "Message 1", "Response 1")
	store.AddMessage("session-1", "Message 2", "Response 2")
	store.AddMessage("session-1", "Message 3", "Response 3")

	messages := store.GetRecentMessages("session-1")
	// Should only keep the last 4 messages
	assert.Len(t, messages, 4)
	assert.Equal(t, "Message 2", messages[0].Content) // Oldest kept
	assert.Equal(t, "Response 3", messages[3].Content) // Most recent
}

func TestInMemorySessionStore_Clear(t *testing.T) {
	store := NewInMemorySessionStore(10, 30*time.Minute)

	store.AddMessage("session-1", "Hello", "Hi!")
	assert.Equal(t, 2, store.GetSessionCount("session-1"))

	err := store.Clear("session-1")
	require.NoError(t, err)

	assert.Equal(t, 0, store.GetSessionCount("session-1"))
	messages := store.GetRecentMessages("session-1")
	assert.Nil(t, messages)
}

func TestInMemorySessionStore_GetSessionCount(t *testing.T) {
	store := NewInMemorySessionStore(10, 30*time.Minute)

	assert.Equal(t, 0, store.GetSessionCount("session-1"))

	store.AddMessage("session-1", "Hello", "Hi!")
	assert.Equal(t, 2, store.GetSessionCount("session-1"))

	store.AddUserMessage("session-1", "Another message")
	assert.Equal(t, 3, store.GetSessionCount("session-1"))
}

func TestInMemorySessionStore_MultipleSessions(t *testing.T) {
	store := NewInMemorySessionStore(10, 30*time.Minute)

	store.AddMessage("session-1", "Hello 1", "Hi 1!")
	store.AddMessage("session-2", "Hello 2", "Hi 2!")

	messages1 := store.GetRecentMessages("session-1")
	messages2 := store.GetRecentMessages("session-2")

	assert.Len(t, messages1, 2)
	assert.Len(t, messages2, 2)
	assert.Equal(t, "Hello 1", messages1[0].Content)
	assert.Equal(t, "Hello 2", messages2[0].Content)

	// Clear only session 1
	store.Clear("session-1")
	assert.Equal(t, 0, store.GetSessionCount("session-1"))
	assert.Equal(t, 2, store.GetSessionCount("session-2"))
}

func TestInMemorySessionStore_ToSchemaMessages(t *testing.T) {
	store := NewInMemorySessionStore(10, 30*time.Minute)

	store.AddMessage("session-1", "Hello", "Hi there!")
	store.AddUserMessage("session-1", "How are you?")

	schemaMessages := store.ToSchemaMessages("session-1")
	require.Len(t, schemaMessages, 3)

	assert.Equal(t, schema.User, schemaMessages[0].Role)
	assert.Equal(t, "Hello", schemaMessages[0].Content)
	assert.Equal(t, schema.Assistant, schemaMessages[1].Role)
	assert.Equal(t, "Hi there!", schemaMessages[1].Content)
	assert.Equal(t, schema.User, schemaMessages[2].Role)
	assert.Equal(t, "How are you?", schemaMessages[2].Content)
}

func TestInMemorySessionStore_ToSchemaMessages_EmptySession(t *testing.T) {
	store := NewInMemorySessionStore(10, 30*time.Minute)

	schemaMessages := store.ToSchemaMessages("nonexistent")
	assert.Nil(t, schemaMessages)
}

func TestInMemorySessionStore_GetAllSessionIDs(t *testing.T) {
	store := NewInMemorySessionStore(10, 30*time.Minute)

	store.AddMessage("session-1", "Hello", "Hi!")
	store.AddMessage("session-2", "Hey", "Hello!")
	store.AddMessage("session-3", "Hi", "Hey!")

	ids := store.GetAllSessionIDs()
	assert.Len(t, ids, 3)
	assert.Contains(t, ids, "session-1")
	assert.Contains(t, ids, "session-2")
	assert.Contains(t, ids, "session-3")
}

func TestInMemorySessionStore_Timestamps(t *testing.T) {
	store := NewInMemorySessionStore(10, 30*time.Minute)

	before := time.Now()
	store.AddMessage("session-1", "Hello", "Hi!")
	after := time.Now()

	messages := store.GetRecentMessages("session-1")
	require.Len(t, messages, 2)

	// Check timestamps are within expected range
	for _, msg := range messages {
		assert.True(t, msg.Timestamp.After(before) || msg.Timestamp.Equal(before))
		assert.True(t, msg.Timestamp.Before(after) || msg.Timestamp.Equal(after))
	}
}

// Test concurrency safety
func TestInMemorySessionStore_Concurrent(t *testing.T) {
	store := NewInMemorySessionStore(100, 30*time.Minute)

	done := make(chan bool)

	// Writer goroutine
	go func() {
		for i := 0; i < 50; i++ {
			store.AddMessage("session-1", "Hello", "Hi!")
		}
		done <- true
	}()

	// Reader goroutine
	go func() {
		for i := 0; i < 50; i++ {
			_ = store.GetRecentMessages("session-1")
		}
		done <- true
	}()

	// Wait for both to complete
	<-done
	<-done

	// Should not panic and have some messages
	assert.True(t, store.GetSessionCount("session-1") > 0)
}
