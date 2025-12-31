package mining

import (
	"encoding/json"
	"sync"
	"time"

	"github.com/Snider/Mining/pkg/logging"
	"github.com/gorilla/websocket"
)

// EventType represents the type of mining event
type EventType string

const (
	// Miner lifecycle events
	EventMinerStarting  EventType = "miner.starting"
	EventMinerStarted   EventType = "miner.started"
	EventMinerStopping  EventType = "miner.stopping"
	EventMinerStopped   EventType = "miner.stopped"
	EventMinerStats     EventType = "miner.stats"
	EventMinerError     EventType = "miner.error"
	EventMinerConnected EventType = "miner.connected"

	// System events
	EventPong      EventType = "pong"
	EventStateSync EventType = "state.sync" // Initial state on connect/reconnect
)

// Event represents a mining event that can be broadcast to clients
type Event struct {
	Type      EventType   `json:"type"`
	Timestamp time.Time   `json:"timestamp"`
	Data      interface{} `json:"data,omitempty"`
}

// MinerStatsData contains stats data for a miner event
type MinerStatsData struct {
	Name        string `json:"name"`
	Hashrate    int    `json:"hashrate"`
	Shares      int    `json:"shares"`
	Rejected    int    `json:"rejected"`
	Uptime      int    `json:"uptime"`
	Algorithm   string `json:"algorithm,omitempty"`
	DiffCurrent int    `json:"diffCurrent,omitempty"`
}

// MinerEventData contains basic miner event data
type MinerEventData struct {
	Name      string `json:"name"`
	ProfileID string `json:"profileId,omitempty"`
	Reason    string `json:"reason,omitempty"`
	Error     string `json:"error,omitempty"`
	Pool      string `json:"pool,omitempty"`
}

// wsClient represents a WebSocket client connection
type wsClient struct {
	conn      *websocket.Conn
	send      chan []byte
	hub       *EventHub
	miners    map[string]bool // subscribed miners, "*" for all
	minersMu  sync.RWMutex    // protects miners map from concurrent access
	closeOnce sync.Once
}

// safeClose closes the send channel exactly once to prevent panic on double close
func (c *wsClient) safeClose() {
	c.closeOnce.Do(func() {
		close(c.send)
	})
}

// StateProvider is a function that returns the current state for sync
type StateProvider func() interface{}

// EventHub manages WebSocket connections and event broadcasting
type EventHub struct {
	// Registered clients
	clients map[*wsClient]bool

	// Inbound events to broadcast
	broadcast chan Event

	// Register requests from clients
	register chan *wsClient

	// Unregister requests from clients
	unregister chan *wsClient

	// Mutex for thread-safe access
	mu sync.RWMutex

	// Stop signal
	stop chan struct{}

	// Ensure Stop() is called only once
	stopOnce sync.Once

	// Connection limits
	maxConnections int

	// State provider for sync on connect
	stateProvider StateProvider
}

// DefaultMaxConnections is the default maximum WebSocket connections
const DefaultMaxConnections = 100

// NewEventHub creates a new EventHub with default settings
func NewEventHub() *EventHub {
	return NewEventHubWithOptions(DefaultMaxConnections)
}

// NewEventHubWithOptions creates a new EventHub with custom settings
func NewEventHubWithOptions(maxConnections int) *EventHub {
	if maxConnections <= 0 {
		maxConnections = DefaultMaxConnections
	}
	return &EventHub{
		clients:        make(map[*wsClient]bool),
		broadcast:      make(chan Event, 256),
		register:       make(chan *wsClient, 16),
		unregister:     make(chan *wsClient, 16), // Buffered to prevent goroutine leaks on shutdown
		stop:           make(chan struct{}),
		maxConnections: maxConnections,
	}
}

// Run starts the EventHub's main loop
func (h *EventHub) Run() {
	for {
		select {
		case <-h.stop:
			// Close all client connections
			h.mu.Lock()
			for client := range h.clients {
				client.safeClose()
				delete(h.clients, client)
			}
			h.mu.Unlock()
			return

		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = true
			stateProvider := h.stateProvider
			h.mu.Unlock()
			logging.Debug("client connected", logging.Fields{"total": len(h.clients)})

			// Send initial state sync if provider is set
			if stateProvider != nil {
				go func(c *wsClient) {
					defer func() {
						if r := recover(); r != nil {
							logging.Error("panic in state sync goroutine", logging.Fields{"panic": r})
						}
					}()
					state := stateProvider()
					if state != nil {
						event := Event{
							Type:      EventStateSync,
							Timestamp: time.Now(),
							Data:      state,
						}
						data, err := MarshalJSON(event)
						if err != nil {
							logging.Error("failed to marshal state sync", logging.Fields{"error": err})
							return
						}
						select {
						case c.send <- data:
						default:
							// Client buffer full
						}
					}
				}(client)
			}

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				client.safeClose()
				// Decrement WebSocket connection metrics
				RecordWSConnection(false)
			}
			h.mu.Unlock()
			logging.Debug("client disconnected", logging.Fields{"total": len(h.clients)})

		case event := <-h.broadcast:
			data, err := MarshalJSON(event)
			if err != nil {
				logging.Error("failed to marshal event", logging.Fields{"error": err})
				continue
			}

			h.mu.RLock()
			for client := range h.clients {
				// Check if client is subscribed to this miner
				if h.shouldSendToClient(client, event) {
					select {
					case client.send <- data:
					default:
						// Client buffer full, close connection
						go func(c *wsClient) {
							h.unregister <- c
						}(client)
					}
				}
			}
			h.mu.RUnlock()
		}
	}
}

// shouldSendToClient checks if an event should be sent to a client
func (h *EventHub) shouldSendToClient(client *wsClient, event Event) bool {
	// Always send pong and system events
	if event.Type == EventPong {
		return true
	}

	// Check miner subscription for miner events (protected by mutex)
	client.minersMu.RLock()
	defer client.minersMu.RUnlock()

	if client.miners == nil || len(client.miners) == 0 {
		// No subscription filter, send all
		return true
	}

	// Check for wildcard subscription
	if client.miners["*"] {
		return true
	}

	// Extract miner name from event data
	minerName := ""
	switch data := event.Data.(type) {
	case MinerStatsData:
		minerName = data.Name
	case MinerEventData:
		minerName = data.Name
	case map[string]interface{}:
		if name, ok := data["name"].(string); ok {
			minerName = name
		}
	}

	if minerName == "" {
		// Non-miner event, send to all
		return true
	}

	return client.miners[minerName]
}

// Stop stops the EventHub (safe to call multiple times)
func (h *EventHub) Stop() {
	h.stopOnce.Do(func() {
		close(h.stop)
	})
}

// SetStateProvider sets the function that provides current state for new clients
func (h *EventHub) SetStateProvider(provider StateProvider) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.stateProvider = provider
}

// Broadcast sends an event to all subscribed clients
func (h *EventHub) Broadcast(event Event) {
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now()
	}
	select {
	case h.broadcast <- event:
	default:
		logging.Warn("broadcast channel full, dropping event", logging.Fields{"type": event.Type})
	}
}

// ClientCount returns the number of connected clients
func (h *EventHub) ClientCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients)
}

// NewEvent creates a new event with the current timestamp
func NewEvent(eventType EventType, data interface{}) Event {
	return Event{
		Type:      eventType,
		Timestamp: time.Now(),
		Data:      data,
	}
}

// writePump pumps messages from the hub to the websocket connection
func (c *wsClient) writePump() {
	ticker := time.NewTicker(30 * time.Second)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				// Hub closed the channel
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			if _, err := w.Write(message); err != nil {
				logging.Debug("WebSocket write error", logging.Fields{"error": err})
				return
			}

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// readPump pumps messages from the websocket connection to the hub
func (c *wsClient) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadLimit(512)
	c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				logging.Debug("WebSocket error", logging.Fields{"error": err})
			}
			break
		}

		// Parse client message
		var msg struct {
			Type   string   `json:"type"`
			Miners []string `json:"miners,omitempty"`
		}
		if err := json.Unmarshal(message, &msg); err != nil {
			continue
		}

		switch msg.Type {
		case "subscribe":
			// Update miner subscription (protected by mutex)
			c.minersMu.Lock()
			c.miners = make(map[string]bool)
			for _, m := range msg.Miners {
				c.miners[m] = true
			}
			c.minersMu.Unlock()
			logging.Debug("client subscribed to miners", logging.Fields{"miners": msg.Miners})

		case "ping":
			// Respond with pong
			c.hub.Broadcast(Event{
				Type:      EventPong,
				Timestamp: time.Now(),
			})
		}
	}
}

// ServeWs handles websocket requests from clients.
// Returns false if the connection was rejected due to limits.
func (h *EventHub) ServeWs(conn *websocket.Conn) bool {
	// Check connection limit
	h.mu.RLock()
	currentCount := len(h.clients)
	h.mu.RUnlock()

	if currentCount >= h.maxConnections {
		logging.Warn("connection rejected: limit reached", logging.Fields{"current": currentCount, "max": h.maxConnections})
		conn.WriteMessage(websocket.CloseMessage,
			websocket.FormatCloseMessage(websocket.CloseTryAgainLater, "connection limit reached"))
		conn.Close()
		return false
	}

	client := &wsClient{
		conn:   conn,
		send:   make(chan []byte, 256),
		hub:    h,
		miners: map[string]bool{"*": true}, // Subscribe to all by default
	}

	h.register <- client

	// Start read/write pumps
	go client.writePump()
	go client.readPump()
	return true
}
