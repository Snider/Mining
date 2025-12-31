package mining

import (
	"encoding/json"
	"sync"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

func TestNewEventHub(t *testing.T) {
	hub := NewEventHub()
	if hub == nil {
		t.Fatal("NewEventHub returned nil")
	}

	if hub.clients == nil {
		t.Error("clients map should be initialized")
	}

	if hub.maxConnections != DefaultMaxConnections {
		t.Errorf("Expected maxConnections %d, got %d", DefaultMaxConnections, hub.maxConnections)
	}
}

func TestNewEventHubWithOptions(t *testing.T) {
	hub := NewEventHubWithOptions(50)
	if hub.maxConnections != 50 {
		t.Errorf("Expected maxConnections 50, got %d", hub.maxConnections)
	}

	// Test with invalid value
	hub2 := NewEventHubWithOptions(0)
	if hub2.maxConnections != DefaultMaxConnections {
		t.Errorf("Expected default maxConnections for 0, got %d", hub2.maxConnections)
	}

	hub3 := NewEventHubWithOptions(-1)
	if hub3.maxConnections != DefaultMaxConnections {
		t.Errorf("Expected default maxConnections for -1, got %d", hub3.maxConnections)
	}
}

func TestEventHubBroadcast(t *testing.T) {
	hub := NewEventHub()
	go hub.Run()
	defer hub.Stop()

	// Create an event
	event := Event{
		Type:      EventMinerStarted,
		Timestamp: time.Now(),
		Data:      MinerEventData{Name: "test-miner"},
	}

	// Broadcast should not block even with no clients
	done := make(chan struct{})
	go func() {
		hub.Broadcast(event)
		close(done)
	}()

	select {
	case <-done:
		// Success
	case <-time.After(time.Second):
		t.Error("Broadcast blocked unexpectedly")
	}
}

func TestEventHubClientCount(t *testing.T) {
	hub := NewEventHub()
	go hub.Run()
	defer hub.Stop()

	// Initial count should be 0
	if count := hub.ClientCount(); count != 0 {
		t.Errorf("Expected 0 clients, got %d", count)
	}
}

func TestEventHubStop(t *testing.T) {
	hub := NewEventHub()
	go hub.Run()

	// Stop should not panic
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Stop panicked: %v", r)
		}
	}()

	hub.Stop()

	// Give time for cleanup
	time.Sleep(50 * time.Millisecond)
}

func TestNewEvent(t *testing.T) {
	data := MinerEventData{Name: "test-miner"}
	event := NewEvent(EventMinerStarted, data)

	if event.Type != EventMinerStarted {
		t.Errorf("Expected type %s, got %s", EventMinerStarted, event.Type)
	}

	if event.Timestamp.IsZero() {
		t.Error("Timestamp should not be zero")
	}

	eventData, ok := event.Data.(MinerEventData)
	if !ok {
		t.Error("Data should be MinerEventData")
	}
	if eventData.Name != "test-miner" {
		t.Errorf("Expected miner name 'test-miner', got '%s'", eventData.Name)
	}
}

func TestEventJSON(t *testing.T) {
	event := Event{
		Type:      EventMinerStats,
		Timestamp: time.Now(),
		Data: MinerStatsData{
			Name:     "test-miner",
			Hashrate: 1000,
			Shares:   10,
			Rejected: 1,
			Uptime:   3600,
		},
	}

	data, err := json.Marshal(event)
	if err != nil {
		t.Fatalf("Failed to marshal event: %v", err)
	}

	var decoded Event
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal event: %v", err)
	}

	if decoded.Type != EventMinerStats {
		t.Errorf("Expected type %s, got %s", EventMinerStats, decoded.Type)
	}
}

func TestSetStateProvider(t *testing.T) {
	hub := NewEventHub()
	go hub.Run()
	defer hub.Stop()

	called := false
	var mu sync.Mutex

	provider := func() interface{} {
		mu.Lock()
		called = true
		mu.Unlock()
		return map[string]string{"status": "ok"}
	}

	hub.SetStateProvider(provider)

	// The provider should be set but not called until a client connects
	mu.Lock()
	wasCalled := called
	mu.Unlock()

	if wasCalled {
		t.Error("Provider should not be called until client connects")
	}
}

// MockWebSocketConn provides a minimal mock for testing
type MockWebSocketConn struct {
	websocket.Conn
	written [][]byte
	mu      sync.Mutex
}

func TestEventTypes(t *testing.T) {
	types := []EventType{
		EventMinerStarting,
		EventMinerStarted,
		EventMinerStopping,
		EventMinerStopped,
		EventMinerStats,
		EventMinerError,
		EventMinerConnected,
		EventPong,
		EventStateSync,
	}

	for _, et := range types {
		if et == "" {
			t.Error("Event type should not be empty")
		}
	}
}
