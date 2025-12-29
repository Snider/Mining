package node

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

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

// PeerRegistry manages known peers with KD-tree based selection.
type PeerRegistry struct {
	peers  map[string]*Peer
	kdTree *poindexter.KDTree[string] // KD-tree with peer ID as payload
	path   string
	mu     sync.RWMutex
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
		peers: make(map[string]*Peer),
		path:  peersPath,
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

// AddPeer adds a new peer to the registry.
func (r *PeerRegistry) AddPeer(peer *Peer) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if peer.ID == "" {
		return fmt.Errorf("peer ID is required")
	}

	if _, exists := r.peers[peer.ID]; exists {
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

	return r.save()
}

// UpdatePeer updates an existing peer's information.
func (r *PeerRegistry) UpdatePeer(peer *Peer) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.peers[peer.ID]; !exists {
		return fmt.Errorf("peer %s not found", peer.ID)
	}

	r.peers[peer.ID] = peer
	r.rebuildKDTree()

	return r.save()
}

// RemovePeer removes a peer from the registry.
func (r *PeerRegistry) RemovePeer(id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.peers[id]; !exists {
		return fmt.Errorf("peer %s not found", id)
	}

	delete(r.peers, id)
	r.rebuildKDTree()

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
func (r *PeerRegistry) UpdateMetrics(id string, pingMS, geoKM float64, hops int) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	peer, exists := r.peers[id]
	if !exists {
		return fmt.Errorf("peer %s not found", id)
	}

	peer.PingMS = pingMS
	peer.GeoKM = geoKM
	peer.Hops = hops
	peer.LastSeen = time.Now()

	r.rebuildKDTree()

	return r.save()
}

// UpdateScore updates a peer's reliability score.
func (r *PeerRegistry) UpdateScore(id string, score float64) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	peer, exists := r.peers[id]
	if !exists {
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

// save persists peers to disk.
func (r *PeerRegistry) save() error {
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

	if err := os.WriteFile(r.path, data, 0644); err != nil {
		return fmt.Errorf("failed to write peers: %w", err)
	}

	return nil
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
