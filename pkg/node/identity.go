// Package node provides P2P node identity and communication for multi-node mining management.
package node

import (
	"crypto/ecdh"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/Snider/Borg/pkg/stmf"
	"github.com/adrg/xdg"
)

// NodeRole defines the operational mode of a node.
type NodeRole string

const (
	// RoleController manages remote worker nodes.
	RoleController NodeRole = "controller"
	// RoleWorker receives commands and runs miners.
	RoleWorker NodeRole = "worker"
	// RoleDual operates as both controller and worker (default).
	RoleDual NodeRole = "dual"
)

// NodeIdentity represents the public identity of a node.
type NodeIdentity struct {
	ID        string    `json:"id"`        // Derived from public key (first 16 bytes hex)
	Name      string    `json:"name"`      // Human-friendly name
	PublicKey string    `json:"publicKey"` // X25519 base64
	CreatedAt time.Time `json:"createdAt"`
	Role      NodeRole  `json:"role"`
}

// NodeManager handles node identity operations including key generation and storage.
type NodeManager struct {
	identity   *NodeIdentity
	privateKey []byte // Never serialized to JSON
	keyPair    *stmf.KeyPair
	keyPath    string // ~/.local/share/lethean-desktop/node/private.key
	configPath string // ~/.config/lethean-desktop/node.json
	mu         sync.RWMutex
}

// NewNodeManager creates a new NodeManager, loading existing identity if available.
func NewNodeManager() (*NodeManager, error) {
	keyPath, err := xdg.DataFile("lethean-desktop/node/private.key")
	if err != nil {
		return nil, fmt.Errorf("failed to get key path: %w", err)
	}

	configPath, err := xdg.ConfigFile("lethean-desktop/node.json")
	if err != nil {
		return nil, fmt.Errorf("failed to get config path: %w", err)
	}

	return NewNodeManagerWithPaths(keyPath, configPath)
}

// NewNodeManagerWithPaths creates a NodeManager with custom paths.
// This is primarily useful for testing to avoid xdg path caching issues.
func NewNodeManagerWithPaths(keyPath, configPath string) (*NodeManager, error) {
	nm := &NodeManager{
		keyPath:    keyPath,
		configPath: configPath,
	}

	// Try to load existing identity
	if err := nm.loadIdentity(); err != nil {
		// Identity doesn't exist yet, that's ok
		return nm, nil
	}

	return nm, nil
}

// HasIdentity returns true if a node identity has been initialized.
func (n *NodeManager) HasIdentity() bool {
	n.mu.RLock()
	defer n.mu.RUnlock()
	return n.identity != nil
}

// GetIdentity returns the node's public identity.
func (n *NodeManager) GetIdentity() *NodeIdentity {
	n.mu.RLock()
	defer n.mu.RUnlock()
	if n.identity == nil {
		return nil
	}
	// Return a copy to prevent mutation
	identity := *n.identity
	return &identity
}

// GenerateIdentity creates a new node identity with the given name and role.
func (n *NodeManager) GenerateIdentity(name string, role NodeRole) error {
	n.mu.Lock()
	defer n.mu.Unlock()

	// Generate X25519 keypair using STMF
	keyPair, err := stmf.GenerateKeyPair()
	if err != nil {
		return fmt.Errorf("failed to generate keypair: %w", err)
	}

	// Derive node ID from public key (first 16 bytes as hex = 32 char ID)
	pubKeyBytes := keyPair.PublicKey()
	hash := sha256.Sum256(pubKeyBytes)
	nodeID := hex.EncodeToString(hash[:16])

	n.identity = &NodeIdentity{
		ID:        nodeID,
		Name:      name,
		PublicKey: keyPair.PublicKeyBase64(),
		CreatedAt: time.Now(),
		Role:      role,
	}

	n.keyPair = keyPair
	n.privateKey = keyPair.PrivateKey()

	// Save private key
	if err := n.savePrivateKey(); err != nil {
		return fmt.Errorf("failed to save private key: %w", err)
	}

	// Save identity config
	if err := n.saveIdentity(); err != nil {
		return fmt.Errorf("failed to save identity: %w", err)
	}

	return nil
}

// DeriveSharedSecret derives a shared secret with a peer using X25519 ECDH.
// The result is hashed with SHA-256 for use as a symmetric key.
func (n *NodeManager) DeriveSharedSecret(peerPubKeyBase64 string) ([]byte, error) {
	n.mu.RLock()
	defer n.mu.RUnlock()

	if n.privateKey == nil {
		return nil, fmt.Errorf("node identity not initialized")
	}

	// Load peer's public key
	peerPubKey, err := stmf.LoadPublicKeyBase64(peerPubKeyBase64)
	if err != nil {
		return nil, fmt.Errorf("failed to load peer public key: %w", err)
	}

	// Load our private key
	privateKey, err := ecdh.X25519().NewPrivateKey(n.privateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to load private key: %w", err)
	}

	// Derive shared secret using ECDH
	sharedSecret, err := privateKey.ECDH(peerPubKey)
	if err != nil {
		return nil, fmt.Errorf("failed to derive shared secret: %w", err)
	}

	// Hash the shared secret using SHA-256 (same pattern as Borg/trix)
	hash := sha256.Sum256(sharedSecret)
	return hash[:], nil
}

// GetPublicKey returns the node's public key in base64 format.
func (n *NodeManager) GetPublicKey() string {
	n.mu.RLock()
	defer n.mu.RUnlock()
	if n.identity == nil {
		return ""
	}
	return n.identity.PublicKey
}

// savePrivateKey saves the private key to disk with restricted permissions.
func (n *NodeManager) savePrivateKey() error {
	// Ensure directory exists
	dir := filepath.Dir(n.keyPath)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("failed to create key directory: %w", err)
	}

	// Write private key with restricted permissions (0600)
	if err := os.WriteFile(n.keyPath, n.privateKey, 0600); err != nil {
		return fmt.Errorf("failed to write private key: %w", err)
	}

	return nil
}

// saveIdentity saves the public identity to the config file.
func (n *NodeManager) saveIdentity() error {
	// Ensure directory exists
	dir := filepath.Dir(n.configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := json.MarshalIndent(n.identity, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal identity: %w", err)
	}

	if err := os.WriteFile(n.configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write identity: %w", err)
	}

	return nil
}

// loadIdentity loads the node identity from disk.
func (n *NodeManager) loadIdentity() error {
	// Load identity config
	data, err := os.ReadFile(n.configPath)
	if err != nil {
		return fmt.Errorf("failed to read identity: %w", err)
	}

	var identity NodeIdentity
	if err := json.Unmarshal(data, &identity); err != nil {
		return fmt.Errorf("failed to unmarshal identity: %w", err)
	}

	// Load private key
	privateKey, err := os.ReadFile(n.keyPath)
	if err != nil {
		return fmt.Errorf("failed to read private key: %w", err)
	}

	// Reconstruct keypair from private key
	keyPair, err := stmf.LoadKeyPair(privateKey)
	if err != nil {
		return fmt.Errorf("failed to load keypair: %w", err)
	}

	n.identity = &identity
	n.privateKey = privateKey
	n.keyPair = keyPair

	return nil
}

// UpdateName updates the node's display name.
func (n *NodeManager) UpdateName(name string) error {
	n.mu.Lock()
	defer n.mu.Unlock()

	if n.identity == nil {
		return fmt.Errorf("node identity not initialized")
	}

	n.identity.Name = name
	return n.saveIdentity()
}

// UpdateRole updates the node's operational role.
func (n *NodeManager) UpdateRole(role NodeRole) error {
	n.mu.Lock()
	defer n.mu.Unlock()

	if n.identity == nil {
		return fmt.Errorf("node identity not initialized")
	}

	n.identity.Role = role
	return n.saveIdentity()
}

// Delete removes the node identity and keys from disk.
func (n *NodeManager) Delete() error {
	n.mu.Lock()
	defer n.mu.Unlock()

	// Remove private key
	if err := os.Remove(n.keyPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove private key: %w", err)
	}

	// Remove identity config
	if err := os.Remove(n.configPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove identity: %w", err)
	}

	n.identity = nil
	n.privateKey = nil
	n.keyPair = nil

	return nil
}
