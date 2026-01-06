package node

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sync"
	"time"

	"github.com/Snider/Mining/pkg/logging"
	"github.com/Snider/Poindexter"
	"github.com/adrg/xdg"
)

// Peer represents a known remote node.
type Peer struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	PublicKey string    `json:"publicKey"`
	Address   string    `json:"address"` // host:port for WebSocket connection
	Role      NodeRole  `json:"role"`
	AddedAt   time.Time `json:"addedAt"`
	LastSeen  time.Time `json:"lastSeen"`

	// Poindexter metrics (updated dynamically)
	PingMS float64 `json:"pingMs"` // Latency in milliseconds
	Hops   int     `json:"hops"`   // Network hop count
	GeoKM  float64 `json:"geoKm"`  // Geographic distance in kilometers
	Score  float64 `json:"score"`  // Reliability score 0-100

	// Connection state (not persisted)
	Connected bool `json:"-"`
}

// saveDebounceInterval is the minimum time between disk writes.
const saveDebounceInterval = 5 * time.Second

// PeerAuthMode controls how unknown peers are handled
type PeerAuthMode int

const (
	// PeerAuthOpen allows any peer to connect (original behavior)
	PeerAuthOpen PeerAuthMode = iota
	// PeerAuthAllowlist only allows pre-registered peers or those with allowed public keys
	PeerAuthAllowlist
)

// Peer name validation constants
const (
	PeerNameMinLength = 1
	PeerNameMaxLength = 64
)

// peerNameRegex validates peer names: alphanumeric, hyphens, underscores, and spaces
var peerNameRegex = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9\-_ ]{0,62}[a-zA-Z0-9]$|^[a-zA-Z0-9]$`)

// safeKeyPrefix returns a truncated key for logging, handling short keys safely
func safeKeyPrefix(key string) string {
	if len(key) >= 16 {
		return key[:16] + "..."
	}
	if len(key) == 0 {
		return "(empty)"
	}
	return key
}

// validatePeerName checks if a peer name is valid.
// Peer names must be 1-64 characters, start and end with alphanumeric,
// and contain only alphanumeric, hyphens, underscores, and spaces.
func validatePeerName(name string) error {
	if name == "" {
		return nil // Empty names are allowed (optional field)
	}
	if len(name) < PeerNameMinLength {
		return fmt.Errorf("peer name too short (min %d characters)", PeerNameMinLength)
	}
	if len(name) > PeerNameMaxLength {
		return fmt.Errorf("peer name too long (max %d characters)", PeerNameMaxLength)
	}
	if !peerNameRegex.MatchString(name) {
		return fmt.Errorf("peer name contains invalid characters (use alphanumeric, hyphens, underscores, spaces)")
	}
	return nil
}

// PeerRegistry manages known peers with KD-tree based selection.
type PeerRegistry struct {
	peers  map[string]*Peer
	kdTree *poindexter.KDTree[string] // KD-tree with peer ID as payload
	path   string
	mu     sync.RWMutex

	// Authentication settings
	authMode           PeerAuthMode    // How to handle unknown peers
	allowedPublicKeys  map[string]bool // Allowlist of public keys (when authMode is Allowlist)
	allowedPublicKeyMu sync.RWMutex    // Protects allowedPublicKeys

	// Debounce disk writes
	dirty        bool          // Whether there are unsaved changes
	saveTimer    *time.Timer   // Timer for debounced save
	saveMu       sync.Mutex    // Protects dirty and saveTimer
	stopChan     chan struct{} // Signal to stop background save
	saveStopOnce sync.Once     // Ensure stopChan is closed only once
}

// Dimension weights for peer selection
// Lower ping, hops, geo are better; higher score is better
var (
	pingWeight  = 1.0
	hopsWeight  = 0.7
	geoWeight   = 0.2
	scoreWeight = 1.2
)

// NewPeerRegistry creates a new PeerRegistry, loading existing peers if available.
func NewPeerRegistry() (*PeerRegistry, error) {
	peersPath, err := xdg.ConfigFile("lethean-desktop/peers.json")
	if err != nil {
		return nil, fmt.Errorf("failed to get peers path: %w", err)
	}

	return NewPeerRegistryWithPath(peersPath)
}

// NewPeerRegistryWithPath creates a new PeerRegistry with a custom path.
// This is primarily useful for testing to avoid xdg path caching issues.
func NewPeerRegistryWithPath(peersPath string) (*PeerRegistry, error) {
	pr := &PeerRegistry{
		peers:             make(map[string]*Peer),
		path:              peersPath,
		stopChan:          make(chan struct{}),
		authMode:          PeerAuthOpen, // Default to open for backward compatibility
		allowedPublicKeys: make(map[string]bool),
	}

	// Try to load existing peers
	if err := pr.load(); err != nil {
		// No existing peers, that's ok
		pr.rebuildKDTree()
		return pr, nil
	}

	pr.rebuildKDTree()
	return pr, nil
}

// SetAuthMode sets the authentication mode for peer connections.
func (r *PeerRegistry) SetAuthMode(mode PeerAuthMode) {
	r.allowedPublicKeyMu.Lock()
	defer r.allowedPublicKeyMu.Unlock()
	r.authMode = mode
	logging.Info("peer auth mode changed", logging.Fields{"mode": mode})
}

// GetAuthMode returns the current authentication mode.
func (r *PeerRegistry) GetAuthMode() PeerAuthMode {
	r.allowedPublicKeyMu.RLock()
	defer r.allowedPublicKeyMu.RUnlock()
	return r.authMode
}

// AllowPublicKey adds a public key to the allowlist.
func (r *PeerRegistry) AllowPublicKey(publicKey string) {
	r.allowedPublicKeyMu.Lock()
	defer r.allowedPublicKeyMu.Unlock()
	r.allowedPublicKeys[publicKey] = true
	logging.Debug("public key added to allowlist", logging.Fields{"key": safeKeyPrefix(publicKey)})
}

// RevokePublicKey removes a public key from the allowlist.
func (r *PeerRegistry) RevokePublicKey(publicKey string) {
	r.allowedPublicKeyMu.Lock()
	defer r.allowedPublicKeyMu.Unlock()
	delete(r.allowedPublicKeys, publicKey)
	logging.Debug("public key removed from allowlist", logging.Fields{"key": safeKeyPrefix(publicKey)})
}

// IsPublicKeyAllowed checks if a public key is in the allowlist.
func (r *PeerRegistry) IsPublicKeyAllowed(publicKey string) bool {
	r.allowedPublicKeyMu.RLock()
	defer r.allowedPublicKeyMu.RUnlock()
	return r.allowedPublicKeys[publicKey]
}

// IsPeerAllowed checks if a peer is allowed to connect based on auth mode.
// Returns true if:
// - AuthMode is Open (allow all)
// - AuthMode is Allowlist AND (peer is pre-registered OR public key is allowlisted)
func (r *PeerRegistry) IsPeerAllowed(peerID string, publicKey string) bool {
	r.allowedPublicKeyMu.RLock()
	authMode := r.authMode
	keyAllowed := r.allowedPublicKeys[publicKey]
	r.allowedPublicKeyMu.RUnlock()

	// Open mode allows everyone
	if authMode == PeerAuthOpen {
		return true
	}

	// Allowlist mode: check if peer is pre-registered
	r.mu.RLock()
	_, isRegistered := r.peers[peerID]
	r.mu.RUnlock()

	if isRegistered {
		return true
	}

	// Check if public key is allowlisted
	return keyAllowed
}

// ListAllowedPublicKeys returns all allowlisted public keys.
func (r *PeerRegistry) ListAllowedPublicKeys() []string {
	r.allowedPublicKeyMu.RLock()
	defer r.allowedPublicKeyMu.RUnlock()

	keys := make([]string, 0, len(r.allowedPublicKeys))
	for key := range r.allowedPublicKeys {
		keys = append(keys, key)
	}
	return keys
}

// AddPeer adds a new peer to the registry.
// Note: Persistence is debounced (writes batched every 5s). Call Close() to ensure
// all changes are flushed to disk before shutdown.
func (r *PeerRegistry) AddPeer(peer *Peer) error {
	r.mu.Lock()

	if peer.ID == "" {
		r.mu.Unlock()
		return fmt.Errorf("peer ID is required")
	}

	// Validate peer name (P2P-LOW-3)
	if err := validatePeerName(peer.Name); err != nil {
		r.mu.Unlock()
		return err
	}

	if _, exists := r.peers[peer.ID]; exists {
		r.mu.Unlock()
		return fmt.Errorf("peer %s already exists", peer.ID)
	}

	// Set defaults
	if peer.AddedAt.IsZero() {
		peer.AddedAt = time.Now()
	}
	if peer.Score == 0 {
		peer.Score = 50 // Default neutral score
	}

	r.peers[peer.ID] = peer
	r.rebuildKDTree()
	r.mu.Unlock()

	return r.save()
}

// UpdatePeer updates an existing peer's information.
// Note: Persistence is debounced. Call Close() to flush before shutdown.
func (r *PeerRegistry) UpdatePeer(peer *Peer) error {
	r.mu.Lock()

	if _, exists := r.peers[peer.ID]; !exists {
		r.mu.Unlock()
		return fmt.Errorf("peer %s not found", peer.ID)
	}

	r.peers[peer.ID] = peer
	r.rebuildKDTree()
	r.mu.Unlock()

	return r.save()
}

// RemovePeer removes a peer from the registry.
// Note: Persistence is debounced. Call Close() to flush before shutdown.
func (r *PeerRegistry) RemovePeer(id string) error {
	r.mu.Lock()

	if _, exists := r.peers[id]; !exists {
		r.mu.Unlock()
		return fmt.Errorf("peer %s not found", id)
	}

	delete(r.peers, id)
	r.rebuildKDTree()
	r.mu.Unlock()

	return r.save()
}

// GetPeer returns a peer by ID.
func (r *PeerRegistry) GetPeer(id string) *Peer {
	r.mu.RLock()
	defer r.mu.RUnlock()

	peer, exists := r.peers[id]
	if !exists {
		return nil
	}

	// Return a copy
	peerCopy := *peer
	return &peerCopy
}

// ListPeers returns all registered peers.
func (r *PeerRegistry) ListPeers() []*Peer {
	r.mu.RLock()
	defer r.mu.RUnlock()

	peers := make([]*Peer, 0, len(r.peers))
	for _, peer := range r.peers {
		peerCopy := *peer
		peers = append(peers, &peerCopy)
	}
	return peers
}

// UpdateMetrics updates a peer's performance metrics.
// Note: Persistence is debounced. Call Close() to flush before shutdown.
func (r *PeerRegistry) UpdateMetrics(id string, pingMS, geoKM float64, hops int) error {
	r.mu.Lock()

	peer, exists := r.peers[id]
	if !exists {
		r.mu.Unlock()
		return fmt.Errorf("peer %s not found", id)
	}

	peer.PingMS = pingMS
	peer.GeoKM = geoKM
	peer.Hops = hops
	peer.LastSeen = time.Now()

	r.rebuildKDTree()
	r.mu.Unlock()

	return r.save()
}

// UpdateScore updates a peer's reliability score.
// Note: Persistence is debounced. Call Close() to flush before shutdown.
func (r *PeerRegistry) UpdateScore(id string, score float64) error {
	r.mu.Lock()

	peer, exists := r.peers[id]
	if !exists {
		r.mu.Unlock()
		return fmt.Errorf("peer %s not found", id)
	}

	// Clamp score to 0-100
	if score < 0 {
		score = 0
	} else if score > 100 {
		score = 100
	}

	peer.Score = score
	r.rebuildKDTree()
	r.mu.Unlock()

	return r.save()
}

// SetConnected updates a peer's connection state.
func (r *PeerRegistry) SetConnected(id string, connected bool) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if peer, exists := r.peers[id]; exists {
		peer.Connected = connected
		if connected {
			peer.LastSeen = time.Now()
		}
	}
}

// Score adjustment constants
const (
	ScoreSuccessIncrement = 1.0   // Increment for successful interaction
	ScoreFailureDecrement = 5.0   // Decrement for failed interaction
	ScoreTimeoutDecrement = 3.0   // Decrement for timeout
	ScoreMinimum          = 0.0   // Minimum score
	ScoreMaximum          = 100.0 // Maximum score
	ScoreDefault          = 50.0  // Default score for new peers
)

// RecordSuccess records a successful interaction with a peer, improving their score.
func (r *PeerRegistry) RecordSuccess(id string) {
	r.mu.Lock()
	peer, exists := r.peers[id]
	if !exists {
		r.mu.Unlock()
		return
	}

	peer.Score = min(peer.Score+ScoreSuccessIncrement, ScoreMaximum)
	peer.LastSeen = time.Now()
	r.mu.Unlock()
	r.save()
}

// RecordFailure records a failed interaction with a peer, reducing their score.
func (r *PeerRegistry) RecordFailure(id string) {
	r.mu.Lock()
	peer, exists := r.peers[id]
	if !exists {
		r.mu.Unlock()
		return
	}

	peer.Score = max(peer.Score-ScoreFailureDecrement, ScoreMinimum)
	newScore := peer.Score
	r.mu.Unlock()
	r.save()

	logging.Debug("peer score decreased", logging.Fields{
		"peer_id":   id,
		"new_score": newScore,
		"reason":    "failure",
	})
}

// RecordTimeout records a timeout when communicating with a peer.
func (r *PeerRegistry) RecordTimeout(id string) {
	r.mu.Lock()
	peer, exists := r.peers[id]
	if !exists {
		r.mu.Unlock()
		return
	}

	peer.Score = max(peer.Score-ScoreTimeoutDecrement, ScoreMinimum)
	newScore := peer.Score
	r.mu.Unlock()
	r.save()

	logging.Debug("peer score decreased", logging.Fields{
		"peer_id":   id,
		"new_score": newScore,
		"reason":    "timeout",
	})
}

// GetPeersByScore returns peers sorted by score (highest first).
func (r *PeerRegistry) GetPeersByScore() []*Peer {
	r.mu.RLock()
	defer r.mu.RUnlock()

	peers := make([]*Peer, 0, len(r.peers))
	for _, p := range r.peers {
		peers = append(peers, p)
	}

	// Sort by score descending
	for i := 0; i < len(peers)-1; i++ {
		for j := i + 1; j < len(peers); j++ {
			if peers[j].Score > peers[i].Score {
				peers[i], peers[j] = peers[j], peers[i]
			}
		}
	}

	return peers
}

// SelectOptimalPeer returns the best peer based on multi-factor optimization.
// Uses Poindexter KD-tree to find the peer closest to ideal metrics.
func (r *PeerRegistry) SelectOptimalPeer() *Peer {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if r.kdTree == nil || len(r.peers) == 0 {
		return nil
	}

	// Target: ideal peer (0 ping, 0 hops, 0 geo, 100 score)
	// Score is inverted (100 - score) so lower is better in the tree
	target := []float64{0, 0, 0, 0}

	result, _, found := r.kdTree.Nearest(target)
	if !found {
		return nil
	}

	peer, exists := r.peers[result.Value]
	if !exists {
		return nil
	}

	peerCopy := *peer
	return &peerCopy
}

// SelectNearestPeers returns the n best peers based on multi-factor optimization.
func (r *PeerRegistry) SelectNearestPeers(n int) []*Peer {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if r.kdTree == nil || len(r.peers) == 0 {
		return nil
	}

	// Target: ideal peer
	target := []float64{0, 0, 0, 0}

	results, _ := r.kdTree.KNearest(target, n)

	peers := make([]*Peer, 0, len(results))
	for _, result := range results {
		if peer, exists := r.peers[result.Value]; exists {
			peerCopy := *peer
			peers = append(peers, &peerCopy)
		}
	}

	return peers
}

// GetConnectedPeers returns all currently connected peers.
func (r *PeerRegistry) GetConnectedPeers() []*Peer {
	r.mu.RLock()
	defer r.mu.RUnlock()

	peers := make([]*Peer, 0)
	for _, peer := range r.peers {
		if peer.Connected {
			peerCopy := *peer
			peers = append(peers, &peerCopy)
		}
	}
	return peers
}

// Count returns the number of registered peers.
func (r *PeerRegistry) Count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.peers)
}

// rebuildKDTree rebuilds the KD-tree from current peers.
// Must be called with lock held.
func (r *PeerRegistry) rebuildKDTree() {
	if len(r.peers) == 0 {
		r.kdTree = nil
		return
	}

	points := make([]poindexter.KDPoint[string], 0, len(r.peers))
	for _, peer := range r.peers {
		// Build 4D point with weighted, normalized values
		// Invert score so that higher score = lower value (better)
		point := poindexter.KDPoint[string]{
			ID: peer.ID,
			Coords: []float64{
				peer.PingMS * pingWeight,
				float64(peer.Hops) * hopsWeight,
				peer.GeoKM * geoWeight,
				(100 - peer.Score) * scoreWeight, // Invert score
			},
			Value: peer.ID,
		}
		points = append(points, point)
	}

	// Build KD-tree with Euclidean distance
	tree, err := poindexter.NewKDTree(points, poindexter.WithMetric(poindexter.EuclideanDistance{}))
	if err != nil {
		// Log error but continue - worst case we don't have optimal selection
		return
	}

	r.kdTree = tree
}

// scheduleSave schedules a debounced save operation.
// Multiple calls within saveDebounceInterval will be coalesced into a single save.
// Must NOT be called with r.mu held.
func (r *PeerRegistry) scheduleSave() {
	r.saveMu.Lock()
	defer r.saveMu.Unlock()

	r.dirty = true

	// If timer already running, let it handle the save
	if r.saveTimer != nil {
		return
	}

	// Start a new timer
	r.saveTimer = time.AfterFunc(saveDebounceInterval, func() {
		r.saveMu.Lock()
		r.saveTimer = nil
		shouldSave := r.dirty
		r.dirty = false
		r.saveMu.Unlock()

		if shouldSave {
			r.mu.RLock()
			err := r.saveNow()
			r.mu.RUnlock()
			if err != nil {
				// Log error but continue - best effort persistence
				logging.Warn("failed to save peer registry", logging.Fields{"error": err})
			}
		}
	})
}

// saveNow persists peers to disk immediately.
// Must be called with r.mu held (at least RLock).
func (r *PeerRegistry) saveNow() error {
	// Ensure directory exists
	dir := filepath.Dir(r.path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create peers directory: %w", err)
	}

	// Convert to slice for JSON
	peers := make([]*Peer, 0, len(r.peers))
	for _, peer := range r.peers {
		peers = append(peers, peer)
	}

	data, err := json.MarshalIndent(peers, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal peers: %w", err)
	}

	// Use atomic write pattern: write to temp file, then rename
	tmpPath := r.path + ".tmp"
	if err := os.WriteFile(tmpPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write peers temp file: %w", err)
	}

	if err := os.Rename(tmpPath, r.path); err != nil {
		os.Remove(tmpPath) // Clean up temp file
		return fmt.Errorf("failed to rename peers file: %w", err)
	}

	return nil
}

// Close flushes any pending changes and releases resources.
func (r *PeerRegistry) Close() error {
	r.saveStopOnce.Do(func() {
		close(r.stopChan)
	})

	// Cancel pending timer and save immediately if dirty
	r.saveMu.Lock()
	if r.saveTimer != nil {
		r.saveTimer.Stop()
		r.saveTimer = nil
	}
	shouldSave := r.dirty
	r.dirty = false
	r.saveMu.Unlock()

	if shouldSave {
		r.mu.RLock()
		err := r.saveNow()
		r.mu.RUnlock()
		return err
	}

	return nil
}

// save is a helper that schedules a debounced save.
// Kept for backward compatibility but now debounces writes.
// Must NOT be called with r.mu held.
func (r *PeerRegistry) save() error {
	r.scheduleSave()
	return nil // Errors will be logged asynchronously
}

// load reads peers from disk.
func (r *PeerRegistry) load() error {
	data, err := os.ReadFile(r.path)
	if err != nil {
		return fmt.Errorf("failed to read peers: %w", err)
	}

	var peers []*Peer
	if err := json.Unmarshal(data, &peers); err != nil {
		return fmt.Errorf("failed to unmarshal peers: %w", err)
	}

	r.peers = make(map[string]*Peer)
	for _, peer := range peers {
		r.peers[peer.ID] = peer
	}

	return nil
}

// Example usage inside a connection handler
