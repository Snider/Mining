package node

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"sync"
	"sync/atomic"
	"time"

	"github.com/Snider/Borg/pkg/smsg"
	"github.com/Snider/Mining/pkg/logging"
	"github.com/gorilla/websocket"
)

// debugLogCounter tracks message counts for rate limiting debug logs
var debugLogCounter atomic.Int64

// debugLogInterval controls how often we log debug messages in hot paths (1 in N)
const debugLogInterval = 100

// DefaultMaxMessageSize is the default maximum message size (1MB)
const DefaultMaxMessageSize int64 = 1 << 20 // 1MB

// TransportConfig configures the WebSocket transport.
type TransportConfig struct {
	ListenAddr     string // ":9091" default
	WSPath         string // "/ws" - WebSocket endpoint path
	TLSCertPath    string // Optional TLS for wss://
	TLSKeyPath     string
	MaxConns       int           // Maximum concurrent connections
	MaxMessageSize int64         // Maximum message size in bytes (0 = 1MB default)
	PingInterval   time.Duration // WebSocket keepalive interval
	PongTimeout    time.Duration // Timeout waiting for pong
}

// DefaultTransportConfig returns sensible defaults.
func DefaultTransportConfig() TransportConfig {
	return TransportConfig{
		ListenAddr:     ":9091",
		WSPath:         "/ws",
		MaxConns:       100,
		MaxMessageSize: DefaultMaxMessageSize,
		PingInterval:   30 * time.Second,
		PongTimeout:    10 * time.Second,
	}
}

// MessageHandler processes incoming messages.
type MessageHandler func(conn *PeerConnection, msg *Message)

// MessageDeduplicator tracks seen message IDs to prevent duplicate processing
type MessageDeduplicator struct {
	seen map[string]time.Time
	mu   sync.RWMutex
	ttl  time.Duration
}

// NewMessageDeduplicator creates a deduplicator with specified TTL
func NewMessageDeduplicator(ttl time.Duration) *MessageDeduplicator {
	d := &MessageDeduplicator{
		seen: make(map[string]time.Time),
		ttl:  ttl,
	}
	return d
}

// IsDuplicate checks if a message ID has been seen recently
func (d *MessageDeduplicator) IsDuplicate(msgID string) bool {
	d.mu.RLock()
	_, exists := d.seen[msgID]
	d.mu.RUnlock()
	return exists
}

// Mark records a message ID as seen
func (d *MessageDeduplicator) Mark(msgID string) {
	d.mu.Lock()
	d.seen[msgID] = time.Now()
	d.mu.Unlock()
}

// Cleanup removes expired entries
func (d *MessageDeduplicator) Cleanup() {
	d.mu.Lock()
	defer d.mu.Unlock()
	now := time.Now()
	for id, seen := range d.seen {
		if now.Sub(seen) > d.ttl {
			delete(d.seen, id)
		}
	}
}

// Transport manages WebSocket connections with SMSG encryption.
type Transport struct {
	config       TransportConfig
	server       *http.Server
	upgrader     websocket.Upgrader
	conns        map[string]*PeerConnection // peer ID -> connection
	pendingConns atomic.Int32               // tracks connections during handshake
	node         *NodeManager
	registry     *PeerRegistry
	handler      MessageHandler
	dedup        *MessageDeduplicator // Message deduplication
	mu           sync.RWMutex
	ctx          context.Context
	cancel       context.CancelFunc
	wg           sync.WaitGroup
}

// PeerRateLimiter implements a simple token bucket rate limiter per peer
type PeerRateLimiter struct {
	tokens     int
	maxTokens  int
	refillRate int // tokens per second
	lastRefill time.Time
	mu         sync.Mutex
}

// NewPeerRateLimiter creates a rate limiter with specified messages/second
func NewPeerRateLimiter(maxTokens, refillRate int) *PeerRateLimiter {
	return &PeerRateLimiter{
		tokens:     maxTokens,
		maxTokens:  maxTokens,
		refillRate: refillRate,
		lastRefill: time.Now(),
	}
}

// Allow checks if a message is allowed and consumes a token if so
func (r *PeerRateLimiter) Allow() bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Refill tokens based on elapsed time
	now := time.Now()
	elapsed := now.Sub(r.lastRefill)
	tokensToAdd := int(elapsed.Seconds()) * r.refillRate
	if tokensToAdd > 0 {
		r.tokens = min(r.tokens+tokensToAdd, r.maxTokens)
		r.lastRefill = now
	}

	// Check if we have tokens available
	if r.tokens > 0 {
		r.tokens--
		return true
	}
	return false
}

// PeerConnection represents an active connection to a peer.
type PeerConnection struct {
	Peer         *Peer
	Conn         *websocket.Conn
	SharedSecret []byte // Derived via X25519 ECDH, used for SMSG
	LastActivity time.Time
	writeMu      sync.Mutex // Serialize WebSocket writes
	transport    *Transport
	closeOnce    sync.Once        // Ensure Close() is only called once
	rateLimiter  *PeerRateLimiter // Per-peer message rate limiting
}

// NewTransport creates a new WebSocket transport.
func NewTransport(node *NodeManager, registry *PeerRegistry, config TransportConfig) *Transport {
	ctx, cancel := context.WithCancel(context.Background())

	return &Transport{
		config:   config,
		node:     node,
		registry: registry,
		conns:    make(map[string]*PeerConnection),
		dedup:    NewMessageDeduplicator(5 * time.Minute), // 5 minute TTL for dedup
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(r *http.Request) bool {
				// Allow local connections only for security
				origin := r.Header.Get("Origin")
				if origin == "" {
					return true // No origin header (non-browser client)
				}
				// Allow localhost and 127.0.0.1 origins
				u, err := url.Parse(origin)
				if err != nil {
					return false
				}
				host := u.Hostname()
				return host == "localhost" || host == "127.0.0.1" || host == "::1"
			},
		},
		ctx:    ctx,
		cancel: cancel,
	}
}

// Start begins listening for incoming connections.
func (t *Transport) Start() error {
	mux := http.NewServeMux()
	mux.HandleFunc(t.config.WSPath, t.handleWSUpgrade)

	t.server = &http.Server{
		Addr:    t.config.ListenAddr,
		Handler: mux,
	}

	t.wg.Add(1)
	go func() {
		defer t.wg.Done()
		var err error
		if t.config.TLSCertPath != "" && t.config.TLSKeyPath != "" {
			err = t.server.ListenAndServeTLS(t.config.TLSCertPath, t.config.TLSKeyPath)
		} else {
			err = t.server.ListenAndServe()
		}
		if err != nil && err != http.ErrServerClosed {
			logging.Error("HTTP server error", logging.Fields{"error": err, "addr": t.config.ListenAddr})
		}
	}()

	// Start message deduplication cleanup goroutine
	t.wg.Add(1)
	go func() {
		defer t.wg.Done()
		ticker := time.NewTicker(time.Minute)
		defer ticker.Stop()
		for {
			select {
			case <-t.ctx.Done():
				return
			case <-ticker.C:
				t.dedup.Cleanup()
			}
		}
	}()

	return nil
}

// Stop gracefully shuts down the transport.
func (t *Transport) Stop() error {
	t.cancel()

	// Gracefully close all connections with shutdown message
	t.mu.Lock()
	for _, pc := range t.conns {
		pc.GracefulClose("server shutdown", DisconnectShutdown)
	}
	t.mu.Unlock()

	// Shutdown HTTP server if it was started
	if t.server != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := t.server.Shutdown(ctx); err != nil {
			return fmt.Errorf("server shutdown error: %w", err)
		}
	}

	t.wg.Wait()
	return nil
}

// OnMessage sets the handler for incoming messages.
// Must be called before Start() to avoid races.
func (t *Transport) OnMessage(handler MessageHandler) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.handler = handler
}

// Connect establishes a connection to a peer.
func (t *Transport) Connect(peer *Peer) (*PeerConnection, error) {
	// Build WebSocket URL
	scheme := "ws"
	if t.config.TLSCertPath != "" {
		scheme = "wss"
	}
	u := url.URL{Scheme: scheme, Host: peer.Address, Path: t.config.WSPath}

	// Dial the peer with timeout to prevent hanging on unresponsive peers
	dialer := websocket.Dialer{
		HandshakeTimeout: 10 * time.Second,
	}
	conn, _, err := dialer.Dial(u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to peer: %w", err)
	}

	pc := &PeerConnection{
		Peer:         peer,
		Conn:         conn,
		LastActivity: time.Now(),
		transport:    t,
		rateLimiter:  NewPeerRateLimiter(100, 50), // 100 burst, 50/sec refill
	}

	// Perform handshake with challenge-response authentication
	// This also derives and stores the shared secret in pc.SharedSecret
	if err := t.performHandshake(pc); err != nil {
		conn.Close()
		return nil, fmt.Errorf("handshake failed: %w", err)
	}

	// Store connection using the real peer ID from handshake
	t.mu.Lock()
	t.conns[pc.Peer.ID] = pc
	t.mu.Unlock()

	logging.Debug("connected to peer", logging.Fields{"peer_id": pc.Peer.ID, "secret_len": len(pc.SharedSecret)})

	// Update registry
	t.registry.SetConnected(pc.Peer.ID, true)

	// Start read loop
	t.wg.Add(1)
	go t.readLoop(pc)

	logging.Debug("started readLoop for peer", logging.Fields{"peer_id": pc.Peer.ID})

	// Start keepalive
	t.wg.Add(1)
	go t.keepalive(pc)

	return pc, nil
}

// Send sends a message to a specific peer.
func (t *Transport) Send(peerID string, msg *Message) error {
	t.mu.RLock()
	pc, exists := t.conns[peerID]
	t.mu.RUnlock()

	if !exists {
		return fmt.Errorf("peer %s not connected", peerID)
	}

	return pc.Send(msg)
}

// Broadcast sends a message to all connected peers except the sender.
// The sender is identified by msg.From and excluded to prevent echo.
func (t *Transport) Broadcast(msg *Message) error {
	t.mu.RLock()
	conns := make([]*PeerConnection, 0, len(t.conns))
	for _, pc := range t.conns {
		// Exclude sender from broadcast to prevent echo (P2P-MED-6)
		if pc.Peer != nil && pc.Peer.ID == msg.From {
			continue
		}
		conns = append(conns, pc)
	}
	t.mu.RUnlock()

	var lastErr error
	for _, pc := range conns {
		if err := pc.Send(msg); err != nil {
			lastErr = err
		}
	}
	return lastErr
}

// GetConnection returns an active connection to a peer.
func (t *Transport) GetConnection(peerID string) *PeerConnection {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.conns[peerID]
}

// handleWSUpgrade handles incoming WebSocket connections.
func (t *Transport) handleWSUpgrade(w http.ResponseWriter, r *http.Request) {
	// Enforce MaxConns limit (including pending connections during handshake)
	t.mu.RLock()
	currentConns := len(t.conns)
	t.mu.RUnlock()
	pendingConns := int(t.pendingConns.Load())

	totalConns := currentConns + pendingConns
	if totalConns >= t.config.MaxConns {
		http.Error(w, "Too many connections", http.StatusServiceUnavailable)
		return
	}

	// Track this connection as pending during handshake
	t.pendingConns.Add(1)
	defer t.pendingConns.Add(-1)

	conn, err := t.upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}

	// Apply message size limit during handshake to prevent memory exhaustion
	maxSize := t.config.MaxMessageSize
	if maxSize <= 0 {
		maxSize = DefaultMaxMessageSize
	}
	conn.SetReadLimit(maxSize)

	// Set handshake timeout to prevent slow/malicious clients from blocking
	handshakeTimeout := 10 * time.Second
	conn.SetReadDeadline(time.Now().Add(handshakeTimeout))

	// Wait for handshake from client
	_, data, err := conn.ReadMessage()
	if err != nil {
		conn.Close()
		return
	}

	// Decode handshake message (not encrypted yet, contains public key)
	var msg Message
	if err := json.Unmarshal(data, &msg); err != nil {
		conn.Close()
		return
	}

	if msg.Type != MsgHandshake {
		conn.Close()
		return
	}

	var payload HandshakePayload
	if err := msg.ParsePayload(&payload); err != nil {
		conn.Close()
		return
	}

	// Check protocol version compatibility (P2P-MED-1)
	if !IsProtocolVersionSupported(payload.Version) {
		logging.Warn("peer connection rejected: incompatible protocol version", logging.Fields{
			"peer_version":       payload.Version,
			"supported_versions": SupportedProtocolVersions,
			"peer_id":            payload.Identity.ID,
		})
		identity := t.node.GetIdentity()
		if identity != nil {
			rejectPayload := HandshakeAckPayload{
				Identity: *identity,
				Accepted: false,
				Reason:   fmt.Sprintf("incompatible protocol version %s, supported: %v", payload.Version, SupportedProtocolVersions),
			}
			rejectMsg, _ := NewMessage(MsgHandshakeAck, identity.ID, payload.Identity.ID, rejectPayload)
			if rejectData, err := MarshalJSON(rejectMsg); err == nil {
				conn.WriteMessage(websocket.TextMessage, rejectData)
			}
		}
		conn.Close()
		return
	}

	// Derive shared secret from peer's public key
	sharedSecret, err := t.node.DeriveSharedSecret(payload.Identity.PublicKey)
	if err != nil {
		conn.Close()
		return
	}

	// Check if peer is allowed to connect (allowlist check)
	if !t.registry.IsPeerAllowed(payload.Identity.ID, payload.Identity.PublicKey) {
		logging.Warn("peer connection rejected: not in allowlist", logging.Fields{
			"peer_id":    payload.Identity.ID,
			"peer_name":  payload.Identity.Name,
			"public_key": payload.Identity.PublicKey[:16] + "...",
		})
		// Send rejection before closing
		identity := t.node.GetIdentity()
		if identity != nil {
			rejectPayload := HandshakeAckPayload{
				Identity: *identity,
				Accepted: false,
				Reason:   "peer not authorized",
			}
			rejectMsg, _ := NewMessage(MsgHandshakeAck, identity.ID, payload.Identity.ID, rejectPayload)
			if rejectData, err := MarshalJSON(rejectMsg); err == nil {
				conn.WriteMessage(websocket.TextMessage, rejectData)
			}
		}
		conn.Close()
		return
	}

	// Create peer if not exists (only if auth passed)
	peer := t.registry.GetPeer(payload.Identity.ID)
	if peer == nil {
		// Auto-register the peer since they passed allowlist check
		peer = &Peer{
			ID:        payload.Identity.ID,
			Name:      payload.Identity.Name,
			PublicKey: payload.Identity.PublicKey,
			Role:      payload.Identity.Role,
			AddedAt:   time.Now(),
			Score:     50,
		}
		t.registry.AddPeer(peer)
		logging.Info("auto-registered new peer", logging.Fields{
			"peer_id":   peer.ID,
			"peer_name": peer.Name,
		})
	}

	pc := &PeerConnection{
		Peer:         peer,
		Conn:         conn,
		SharedSecret: sharedSecret,
		LastActivity: time.Now(),
		transport:    t,
		rateLimiter:  NewPeerRateLimiter(100, 50), // 100 burst, 50/sec refill
	}

	// Send handshake acknowledgment
	identity := t.node.GetIdentity()
	if identity == nil {
		conn.Close()
		return
	}

	// Sign the client's challenge to prove we have the matching private key
	var challengeResponse []byte
	if len(payload.Challenge) > 0 {
		challengeResponse = SignChallenge(payload.Challenge, sharedSecret)
	}

	ackPayload := HandshakeAckPayload{
		Identity:          *identity,
		ChallengeResponse: challengeResponse,
		Accepted:          true,
	}

	ackMsg, err := NewMessage(MsgHandshakeAck, identity.ID, peer.ID, ackPayload)
	if err != nil {
		conn.Close()
		return
	}

	// First ack is unencrypted (peer needs to know our public key)
	ackData, err := MarshalJSON(ackMsg)
	if err != nil {
		conn.Close()
		return
	}

	if err := conn.WriteMessage(websocket.TextMessage, ackData); err != nil {
		conn.Close()
		return
	}

	// Store connection
	t.mu.Lock()
	t.conns[peer.ID] = pc
	t.mu.Unlock()

	// Update registry
	t.registry.SetConnected(peer.ID, true)

	// Start read loop
	t.wg.Add(1)
	go t.readLoop(pc)

	// Start keepalive
	t.wg.Add(1)
	go t.keepalive(pc)
}

// performHandshake initiates handshake with a peer.
func (t *Transport) performHandshake(pc *PeerConnection) error {
	// Set handshake timeout
	handshakeTimeout := 10 * time.Second
	pc.Conn.SetWriteDeadline(time.Now().Add(handshakeTimeout))
	pc.Conn.SetReadDeadline(time.Now().Add(handshakeTimeout))
	defer func() {
		// Reset deadlines after handshake
		pc.Conn.SetWriteDeadline(time.Time{})
		pc.Conn.SetReadDeadline(time.Time{})
	}()

	identity := t.node.GetIdentity()
	if identity == nil {
		return fmt.Errorf("node identity not initialized")
	}

	// Generate challenge for the server to prove it has the matching private key
	challenge, err := GenerateChallenge()
	if err != nil {
		return fmt.Errorf("generate challenge: %w", err)
	}

	payload := HandshakePayload{
		Identity:  *identity,
		Challenge: challenge,
		Version:   ProtocolVersion,
	}

	msg, err := NewMessage(MsgHandshake, identity.ID, pc.Peer.ID, payload)
	if err != nil {
		return fmt.Errorf("create handshake message: %w", err)
	}

	// First message is unencrypted (peer needs our public key)
	data, err := MarshalJSON(msg)
	if err != nil {
		return fmt.Errorf("marshal handshake message: %w", err)
	}

	if err := pc.Conn.WriteMessage(websocket.TextMessage, data); err != nil {
		return fmt.Errorf("send handshake: %w", err)
	}

	// Wait for ack
	_, ackData, err := pc.Conn.ReadMessage()
	if err != nil {
		return fmt.Errorf("read handshake ack: %w", err)
	}

	var ackMsg Message
	if err := json.Unmarshal(ackData, &ackMsg); err != nil {
		return fmt.Errorf("unmarshal handshake ack: %w", err)
	}

	if ackMsg.Type != MsgHandshakeAck {
		return fmt.Errorf("expected handshake_ack, got %s", ackMsg.Type)
	}

	var ackPayload HandshakeAckPayload
	if err := ackMsg.ParsePayload(&ackPayload); err != nil {
		return fmt.Errorf("parse handshake ack payload: %w", err)
	}

	if !ackPayload.Accepted {
		return fmt.Errorf("handshake rejected: %s", ackPayload.Reason)
	}

	// Update peer with the received identity info
	pc.Peer.ID = ackPayload.Identity.ID
	pc.Peer.PublicKey = ackPayload.Identity.PublicKey
	pc.Peer.Name = ackPayload.Identity.Name
	pc.Peer.Role = ackPayload.Identity.Role

	// Verify challenge response - derive shared secret first using the peer's public key
	sharedSecret, err := t.node.DeriveSharedSecret(pc.Peer.PublicKey)
	if err != nil {
		return fmt.Errorf("derive shared secret for challenge verification: %w", err)
	}

	// Verify the server's response to our challenge
	if len(ackPayload.ChallengeResponse) == 0 {
		return fmt.Errorf("server did not provide challenge response")
	}
	if !VerifyChallenge(challenge, ackPayload.ChallengeResponse, sharedSecret) {
		return fmt.Errorf("challenge response verification failed: server may not have matching private key")
	}

	// Store the shared secret for later use
	pc.SharedSecret = sharedSecret

	// Update the peer in registry with the real identity
	if err := t.registry.UpdatePeer(pc.Peer); err != nil {
		// If update fails (peer not found with old ID), add as new
		t.registry.AddPeer(pc.Peer)
	}

	logging.Debug("handshake completed with challenge-response verification", logging.Fields{
		"peer_id":   pc.Peer.ID,
		"peer_name": pc.Peer.Name,
	})

	return nil
}

// readLoop reads messages from a peer connection.
func (t *Transport) readLoop(pc *PeerConnection) {
	defer t.wg.Done()
	defer t.removeConnection(pc)

	// Apply message size limit to prevent memory exhaustion attacks
	maxSize := t.config.MaxMessageSize
	if maxSize <= 0 {
		maxSize = DefaultMaxMessageSize
	}
	pc.Conn.SetReadLimit(maxSize)

	for {
		select {
		case <-t.ctx.Done():
			return
		default:
		}

		// Set read deadline to prevent blocking forever on unresponsive connections
		readDeadline := t.config.PingInterval + t.config.PongTimeout
		if err := pc.Conn.SetReadDeadline(time.Now().Add(readDeadline)); err != nil {
			logging.Error("SetReadDeadline error", logging.Fields{"peer_id": pc.Peer.ID, "error": err})
			return
		}

		_, data, err := pc.Conn.ReadMessage()
		if err != nil {
			logging.Debug("read error from peer", logging.Fields{"peer_id": pc.Peer.ID, "error": err})
			return
		}

		pc.LastActivity = time.Now()

		// Check rate limit before processing
		if pc.rateLimiter != nil && !pc.rateLimiter.Allow() {
			logging.Warn("peer rate limited, dropping message", logging.Fields{"peer_id": pc.Peer.ID})
			continue // Drop message from rate-limited peer
		}

		// Decrypt message using SMSG with shared secret
		msg, err := t.decryptMessage(data, pc.SharedSecret)
		if err != nil {
			logging.Debug("decrypt error from peer", logging.Fields{"peer_id": pc.Peer.ID, "error": err, "data_len": len(data)})
			continue // Skip invalid messages
		}

		// Check for duplicate messages (prevents amplification attacks)
		if t.dedup.IsDuplicate(msg.ID) {
			logging.Debug("dropping duplicate message", logging.Fields{"msg_id": msg.ID, "peer_id": pc.Peer.ID})
			continue
		}
		t.dedup.Mark(msg.ID)

		// Rate limit debug logs in hot path to reduce noise (log 1 in N messages)
		if debugLogCounter.Add(1)%debugLogInterval == 0 {
			logging.Debug("received message from peer", logging.Fields{"type": msg.Type, "peer_id": pc.Peer.ID, "reply_to": msg.ReplyTo, "sample": "1/100"})
		}

		// Dispatch to handler (read handler under lock to avoid race)
		t.mu.RLock()
		handler := t.handler
		t.mu.RUnlock()
		if handler != nil {
			handler(pc, msg)
		}
	}
}

// keepalive sends periodic pings.
func (t *Transport) keepalive(pc *PeerConnection) {
	defer t.wg.Done()

	ticker := time.NewTicker(t.config.PingInterval)
	defer ticker.Stop()

	for {
		select {
		case <-t.ctx.Done():
			return
		case <-ticker.C:
			// Check if connection is still alive
			if time.Since(pc.LastActivity) > t.config.PingInterval+t.config.PongTimeout {
				t.removeConnection(pc)
				return
			}

			// Send ping
			identity := t.node.GetIdentity()
			pingMsg, err := NewMessage(MsgPing, identity.ID, pc.Peer.ID, PingPayload{
				SentAt: time.Now().UnixMilli(),
			})
			if err != nil {
				continue
			}

			if err := pc.Send(pingMsg); err != nil {
				t.removeConnection(pc)
				return
			}
		}
	}
}

// removeConnection removes and cleans up a connection.
func (t *Transport) removeConnection(pc *PeerConnection) {
	t.mu.Lock()
	delete(t.conns, pc.Peer.ID)
	t.mu.Unlock()

	t.registry.SetConnected(pc.Peer.ID, false)
	pc.Close()
}

// Send sends an encrypted message over the connection.
func (pc *PeerConnection) Send(msg *Message) error {
	pc.writeMu.Lock()
	defer pc.writeMu.Unlock()

	// Encrypt message using SMSG
	data, err := pc.transport.encryptMessage(msg, pc.SharedSecret)
	if err != nil {
		return err
	}

	// Set write deadline to prevent blocking forever
	if err := pc.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second)); err != nil {
		return fmt.Errorf("failed to set write deadline: %w", err)
	}
	defer pc.Conn.SetWriteDeadline(time.Time{}) // Reset deadline after send

	return pc.Conn.WriteMessage(websocket.BinaryMessage, data)
}

// Close closes the connection.
func (pc *PeerConnection) Close() error {
	var err error
	pc.closeOnce.Do(func() {
		err = pc.Conn.Close()
	})
	return err
}

// DisconnectPayload contains reason for disconnect.
type DisconnectPayload struct {
	Reason string `json:"reason"`
	Code   int    `json:"code"` // Optional disconnect code
}

// Disconnect codes
const (
	DisconnectNormal      = 1000 // Normal closure
	DisconnectGoingAway   = 1001 // Server/peer going away
	DisconnectProtocolErr = 1002 // Protocol error
	DisconnectTimeout     = 1003 // Idle timeout
	DisconnectShutdown    = 1004 // Server shutdown
)

// GracefulClose sends a disconnect message before closing the connection.
func (pc *PeerConnection) GracefulClose(reason string, code int) error {
	var err error
	pc.closeOnce.Do(func() {
		// Try to send disconnect message (best effort)
		if pc.transport != nil && pc.SharedSecret != nil {
			identity := pc.transport.node.GetIdentity()
			if identity != nil {
				payload := DisconnectPayload{
					Reason: reason,
					Code:   code,
				}
				msg, msgErr := NewMessage(MsgDisconnect, identity.ID, pc.Peer.ID, payload)
				if msgErr == nil {
					// Set short deadline for disconnect message
					pc.Conn.SetWriteDeadline(time.Now().Add(2 * time.Second))
					pc.Send(msg)
				}
			}
		}

		// Close the underlying connection
		err = pc.Conn.Close()
	})
	return err
}

// encryptMessage encrypts a message using SMSG with the shared secret.
func (t *Transport) encryptMessage(msg *Message, sharedSecret []byte) ([]byte, error) {
	// Serialize message to JSON (using pooled buffer for efficiency)
	msgData, err := MarshalJSON(msg)
	if err != nil {
		return nil, err
	}

	// Create SMSG message
	smsgMsg := smsg.NewMessage(string(msgData))

	// Encrypt using shared secret as password (base64 encoded)
	password := base64.StdEncoding.EncodeToString(sharedSecret)
	encrypted, err := smsg.Encrypt(smsgMsg, password)
	if err != nil {
		return nil, err
	}

	return encrypted, nil
}

// decryptMessage decrypts a message using SMSG with the shared secret.
func (t *Transport) decryptMessage(data []byte, sharedSecret []byte) (*Message, error) {
	// Decrypt using shared secret as password
	password := base64.StdEncoding.EncodeToString(sharedSecret)
	smsgMsg, err := smsg.Decrypt(data, password)
	if err != nil {
		return nil, err
	}

	// Parse message from JSON
	var msg Message
	if err := json.Unmarshal([]byte(smsgMsg.Body), &msg); err != nil {
		return nil, err
	}

	return &msg, nil
}

// ConnectedPeers returns the number of connected peers.
func (t *Transport) ConnectedPeers() int {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return len(t.conns)
}
