package node

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func setupTestPeerRegistry(t *testing.T) (*PeerRegistry, func()) {
	tmpDir, err := os.MkdirTemp("", "peer-registry-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}

	peersPath := filepath.Join(tmpDir, "peers.json")

	pr, err := NewPeerRegistryWithPath(peersPath)
	if err != nil {
		os.RemoveAll(tmpDir)
		t.Fatalf("failed to create peer registry: %v", err)
	}

	cleanup := func() {
		os.RemoveAll(tmpDir)
	}

	return pr, cleanup
}

func TestPeerRegistry_NewPeerRegistry(t *testing.T) {
	pr, cleanup := setupTestPeerRegistry(t)
	defer cleanup()

	if pr.Count() != 0 {
		t.Errorf("expected 0 peers, got %d", pr.Count())
	}
}

func TestPeerRegistry_AddPeer(t *testing.T) {
	pr, cleanup := setupTestPeerRegistry(t)
	defer cleanup()

	peer := &Peer{
		ID:        "test-peer-1",
		Name:      "Test Peer",
		PublicKey: "testkey123",
		Address:   "192.168.1.100:9091",
		Role:      RoleWorker,
		Score:     75,
	}

	err := pr.AddPeer(peer)
	if err != nil {
		t.Fatalf("failed to add peer: %v", err)
	}

	if pr.Count() != 1 {
		t.Errorf("expected 1 peer, got %d", pr.Count())
	}

	// Try to add duplicate
	err = pr.AddPeer(peer)
	if err == nil {
		t.Error("expected error when adding duplicate peer")
	}
}

func TestPeerRegistry_GetPeer(t *testing.T) {
	pr, cleanup := setupTestPeerRegistry(t)
	defer cleanup()

	peer := &Peer{
		ID:        "get-test-peer",
		Name:      "Get Test",
		PublicKey: "getkey123",
		Address:   "10.0.0.1:9091",
		Role:      RoleDual,
	}

	pr.AddPeer(peer)

	retrieved := pr.GetPeer("get-test-peer")
	if retrieved == nil {
		t.Fatal("failed to retrieve peer")
	}

	if retrieved.Name != "Get Test" {
		t.Errorf("expected name 'Get Test', got '%s'", retrieved.Name)
	}

	// Non-existent peer
	nonExistent := pr.GetPeer("non-existent")
	if nonExistent != nil {
		t.Error("expected nil for non-existent peer")
	}
}

func TestPeerRegistry_ListPeers(t *testing.T) {
	pr, cleanup := setupTestPeerRegistry(t)
	defer cleanup()

	peers := []*Peer{
		{ID: "list-1", Name: "Peer 1", Address: "1.1.1.1:9091", Role: RoleWorker},
		{ID: "list-2", Name: "Peer 2", Address: "2.2.2.2:9091", Role: RoleWorker},
		{ID: "list-3", Name: "Peer 3", Address: "3.3.3.3:9091", Role: RoleController},
	}

	for _, p := range peers {
		pr.AddPeer(p)
	}

	listed := pr.ListPeers()
	if len(listed) != 3 {
		t.Errorf("expected 3 peers, got %d", len(listed))
	}
}

func TestPeerRegistry_RemovePeer(t *testing.T) {
	pr, cleanup := setupTestPeerRegistry(t)
	defer cleanup()

	peer := &Peer{
		ID:      "remove-test",
		Name:    "Remove Me",
		Address: "5.5.5.5:9091",
		Role:    RoleWorker,
	}

	pr.AddPeer(peer)

	if pr.Count() != 1 {
		t.Error("peer should exist before removal")
	}

	err := pr.RemovePeer("remove-test")
	if err != nil {
		t.Fatalf("failed to remove peer: %v", err)
	}

	if pr.Count() != 0 {
		t.Error("peer should be removed")
	}

	// Remove non-existent
	err = pr.RemovePeer("non-existent")
	if err == nil {
		t.Error("expected error when removing non-existent peer")
	}
}

func TestPeerRegistry_UpdateMetrics(t *testing.T) {
	pr, cleanup := setupTestPeerRegistry(t)
	defer cleanup()

	peer := &Peer{
		ID:      "metrics-test",
		Name:    "Metrics Peer",
		Address: "6.6.6.6:9091",
		Role:    RoleWorker,
	}

	pr.AddPeer(peer)

	err := pr.UpdateMetrics("metrics-test", 50.5, 100.2, 3)
	if err != nil {
		t.Fatalf("failed to update metrics: %v", err)
	}

	updated := pr.GetPeer("metrics-test")
	if updated == nil {
		t.Fatal("expected peer to exist")
	}
	if updated.PingMS != 50.5 {
		t.Errorf("expected ping 50.5, got %f", updated.PingMS)
	}
	if updated.GeoKM != 100.2 {
		t.Errorf("expected geo 100.2, got %f", updated.GeoKM)
	}
	if updated.Hops != 3 {
		t.Errorf("expected hops 3, got %d", updated.Hops)
	}
}

func TestPeerRegistry_UpdateScore(t *testing.T) {
	pr, cleanup := setupTestPeerRegistry(t)
	defer cleanup()

	peer := &Peer{
		ID:    "score-test",
		Name:  "Score Peer",
		Score: 50,
	}

	pr.AddPeer(peer)

	err := pr.UpdateScore("score-test", 85.5)
	if err != nil {
		t.Fatalf("failed to update score: %v", err)
	}

	updated := pr.GetPeer("score-test")
	if updated == nil {
		t.Fatal("expected peer to exist")
	}
	if updated.Score != 85.5 {
		t.Errorf("expected score 85.5, got %f", updated.Score)
	}

	// Test clamping - over 100
	err = pr.UpdateScore("score-test", 150)
	if err != nil {
		t.Fatalf("failed to update score: %v", err)
	}

	updated = pr.GetPeer("score-test")
	if updated == nil {
		t.Fatal("expected peer to exist")
	}
	if updated.Score != 100 {
		t.Errorf("expected score clamped to 100, got %f", updated.Score)
	}

	// Test clamping - below 0
	err = pr.UpdateScore("score-test", -50)
	if err != nil {
		t.Fatalf("failed to update score: %v", err)
	}

	updated = pr.GetPeer("score-test")
	if updated == nil {
		t.Fatal("expected peer to exist")
	}
	if updated.Score != 0 {
		t.Errorf("expected score clamped to 0, got %f", updated.Score)
	}
}

func TestPeerRegistry_SetConnected(t *testing.T) {
	pr, cleanup := setupTestPeerRegistry(t)
	defer cleanup()

	peer := &Peer{
		ID:        "connect-test",
		Name:      "Connect Peer",
		Connected: false,
	}

	pr.AddPeer(peer)

	pr.SetConnected("connect-test", true)

	updated := pr.GetPeer("connect-test")
	if updated == nil {
		t.Fatal("expected peer to exist")
	}
	if !updated.Connected {
		t.Error("peer should be connected")
	}
	if updated.LastSeen.IsZero() {
		t.Error("LastSeen should be set when connected")
	}

	pr.SetConnected("connect-test", false)
	updated = pr.GetPeer("connect-test")
	if updated == nil {
		t.Fatal("expected peer to exist")
	}
	if updated.Connected {
		t.Error("peer should be disconnected")
	}
}

func TestPeerRegistry_GetConnectedPeers(t *testing.T) {
	pr, cleanup := setupTestPeerRegistry(t)
	defer cleanup()

	peers := []*Peer{
		{ID: "conn-1", Name: "Peer 1"},
		{ID: "conn-2", Name: "Peer 2"},
		{ID: "conn-3", Name: "Peer 3"},
	}

	for _, p := range peers {
		pr.AddPeer(p)
	}

	pr.SetConnected("conn-1", true)
	pr.SetConnected("conn-3", true)

	connected := pr.GetConnectedPeers()
	if len(connected) != 2 {
		t.Errorf("expected 2 connected peers, got %d", len(connected))
	}
}

func TestPeerRegistry_SelectOptimalPeer(t *testing.T) {
	pr, cleanup := setupTestPeerRegistry(t)
	defer cleanup()

	// Add peers with different metrics
	peers := []*Peer{
		{ID: "opt-1", Name: "Slow Peer", PingMS: 200, Hops: 5, GeoKM: 1000, Score: 50},
		{ID: "opt-2", Name: "Fast Peer", PingMS: 10, Hops: 1, GeoKM: 50, Score: 90},
		{ID: "opt-3", Name: "Medium Peer", PingMS: 50, Hops: 2, GeoKM: 200, Score: 70},
	}

	for _, p := range peers {
		pr.AddPeer(p)
	}

	optimal := pr.SelectOptimalPeer()
	if optimal == nil {
		t.Fatal("expected to find an optimal peer")
	}

	// The "Fast Peer" should be selected as optimal
	if optimal.ID != "opt-2" {
		t.Errorf("expected 'opt-2' (Fast Peer) to be optimal, got '%s' (%s)", optimal.ID, optimal.Name)
	}
}

func TestPeerRegistry_SelectNearestPeers(t *testing.T) {
	pr, cleanup := setupTestPeerRegistry(t)
	defer cleanup()

	peers := []*Peer{
		{ID: "near-1", Name: "Peer 1", PingMS: 100, Score: 50},
		{ID: "near-2", Name: "Peer 2", PingMS: 10, Score: 90},
		{ID: "near-3", Name: "Peer 3", PingMS: 50, Score: 70},
		{ID: "near-4", Name: "Peer 4", PingMS: 200, Score: 30},
	}

	for _, p := range peers {
		pr.AddPeer(p)
	}

	nearest := pr.SelectNearestPeers(2)
	if len(nearest) != 2 {
		t.Errorf("expected 2 nearest peers, got %d", len(nearest))
	}
}

func TestPeerRegistry_Persistence(t *testing.T) {
	tmpDir, _ := os.MkdirTemp("", "persist-test")
	defer os.RemoveAll(tmpDir)

	peersPath := filepath.Join(tmpDir, "peers.json")

	// Create and save
	pr1, err := NewPeerRegistryWithPath(peersPath)
	if err != nil {
		t.Fatalf("failed to create first registry: %v", err)
	}

	peer := &Peer{
		ID:      "persist-test",
		Name:    "Persistent Peer",
		Address: "7.7.7.7:9091",
		Role:    RoleDual,
		AddedAt: time.Now(),
	}

	pr1.AddPeer(peer)

	// Flush pending changes before reloading
	if err := pr1.Close(); err != nil {
		t.Fatalf("failed to close first registry: %v", err)
	}

	// Load in new registry from same path
	pr2, err := NewPeerRegistryWithPath(peersPath)
	if err != nil {
		t.Fatalf("failed to create second registry: %v", err)
	}

	if pr2.Count() != 1 {
		t.Errorf("expected 1 peer after reload, got %d", pr2.Count())
	}

	loaded := pr2.GetPeer("persist-test")
	if loaded == nil {
		t.Fatal("peer should exist after reload")
	}

	if loaded.Name != "Persistent Peer" {
		t.Errorf("expected name 'Persistent Peer', got '%s'", loaded.Name)
	}
}

// --- Security Feature Tests ---

func TestPeerRegistry_AuthMode(t *testing.T) {
	pr, cleanup := setupTestPeerRegistry(t)
	defer cleanup()

	// Default should be Open
	if pr.GetAuthMode() != PeerAuthOpen {
		t.Errorf("expected default auth mode to be Open, got %d", pr.GetAuthMode())
	}

	// Set to Allowlist
	pr.SetAuthMode(PeerAuthAllowlist)
	if pr.GetAuthMode() != PeerAuthAllowlist {
		t.Errorf("expected auth mode to be Allowlist after setting, got %d", pr.GetAuthMode())
	}

	// Set back to Open
	pr.SetAuthMode(PeerAuthOpen)
	if pr.GetAuthMode() != PeerAuthOpen {
		t.Errorf("expected auth mode to be Open after resetting, got %d", pr.GetAuthMode())
	}
}

func TestPeerRegistry_PublicKeyAllowlist(t *testing.T) {
	pr, cleanup := setupTestPeerRegistry(t)
	defer cleanup()

	testKey := "base64PublicKeyExample1234567890123456"

	// Initially key should not be allowed
	if pr.IsPublicKeyAllowed(testKey) {
		t.Error("key should not be allowed before adding")
	}

	// Add key to allowlist
	pr.AllowPublicKey(testKey)
	if !pr.IsPublicKeyAllowed(testKey) {
		t.Error("key should be allowed after adding")
	}

	// List should contain the key
	keys := pr.ListAllowedPublicKeys()
	found := false
	for _, k := range keys {
		if k == testKey {
			found = true
			break
		}
	}
	if !found {
		t.Error("ListAllowedPublicKeys should contain the added key")
	}

	// Revoke key
	pr.RevokePublicKey(testKey)
	if pr.IsPublicKeyAllowed(testKey) {
		t.Error("key should not be allowed after revoking")
	}

	// List should be empty
	keys = pr.ListAllowedPublicKeys()
	if len(keys) != 0 {
		t.Errorf("expected 0 keys after revoke, got %d", len(keys))
	}
}

func TestPeerRegistry_IsPeerAllowed_OpenMode(t *testing.T) {
	pr, cleanup := setupTestPeerRegistry(t)
	defer cleanup()

	pr.SetAuthMode(PeerAuthOpen)

	// In Open mode, any peer should be allowed
	if !pr.IsPeerAllowed("unknown-peer", "unknown-key") {
		t.Error("in Open mode, all peers should be allowed")
	}

	if !pr.IsPeerAllowed("", "") {
		t.Error("in Open mode, even empty IDs should be allowed")
	}
}

func TestPeerRegistry_IsPeerAllowed_AllowlistMode(t *testing.T) {
	pr, cleanup := setupTestPeerRegistry(t)
	defer cleanup()

	pr.SetAuthMode(PeerAuthAllowlist)

	// Unknown peer with unknown key should be rejected
	if pr.IsPeerAllowed("unknown-peer", "unknown-key") {
		t.Error("in Allowlist mode, unknown peers should be rejected")
	}

	// Pre-registered peer should be allowed
	peer := &Peer{
		ID:        "registered-peer",
		Name:      "Registered",
		PublicKey: "registered-key",
	}
	pr.AddPeer(peer)

	if !pr.IsPeerAllowed("registered-peer", "any-key") {
		t.Error("pre-registered peer should be allowed in Allowlist mode")
	}

	// Peer with allowlisted public key should be allowed
	pr.AllowPublicKey("allowed-key-1234567890")
	if !pr.IsPeerAllowed("new-peer", "allowed-key-1234567890") {
		t.Error("peer with allowlisted key should be allowed")
	}

	// Unknown peer with non-allowlisted key should still be rejected
	if pr.IsPeerAllowed("another-peer", "not-allowed-key") {
		t.Error("peer without allowlisted key should be rejected")
	}
}

func TestPeerRegistry_PeerNameValidation(t *testing.T) {
	pr, cleanup := setupTestPeerRegistry(t)
	defer cleanup()

	testCases := []struct {
		name      string
		peerName  string
		shouldErr bool
	}{
		{"empty name allowed", "", false},
		{"single char", "A", false},
		{"simple name", "MyPeer", false},
		{"name with hyphen", "my-peer", false},
		{"name with underscore", "my_peer", false},
		{"name with space", "My Peer", false},
		{"name with numbers", "Peer123", false},
		{"max length name", "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789AB", false},
		{"too long name", "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789ABC", true},
		{"starts with hyphen", "-peer", true},
		{"ends with hyphen", "peer-", true},
		{"special chars", "peer@host", true},
		{"unicode chars", "peer\u0000name", true},
	}

	for i, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			peer := &Peer{
				ID:   "test-peer-" + string(rune('A'+i)),
				Name: tc.peerName,
			}
			err := pr.AddPeer(peer)
			if tc.shouldErr && err == nil {
				t.Errorf("expected error for name '%s' but got none", tc.peerName)
			} else if !tc.shouldErr && err != nil {
				t.Errorf("unexpected error for name '%s': %v", tc.peerName, err)
			}
			// Clean up for next test
			if err == nil {
				pr.RemovePeer(peer.ID)
			}
		})
	}
}

func TestPeerRegistry_ScoreRecording(t *testing.T) {
	pr, cleanup := setupTestPeerRegistry(t)
	defer cleanup()

	peer := &Peer{
		ID:    "score-record-test",
		Name:  "Score Peer",
		Score: 50, // Start at neutral
	}
	pr.AddPeer(peer)

	// Record successes - score should increase
	for i := 0; i < 5; i++ {
		pr.RecordSuccess("score-record-test")
	}
	updated := pr.GetPeer("score-record-test")
	if updated.Score <= 50 {
		t.Errorf("score should increase after successes, got %f", updated.Score)
	}

	// Record failures - score should decrease
	initialScore := updated.Score
	for i := 0; i < 3; i++ {
		pr.RecordFailure("score-record-test")
	}
	updated = pr.GetPeer("score-record-test")
	if updated.Score >= initialScore {
		t.Errorf("score should decrease after failures, got %f (was %f)", updated.Score, initialScore)
	}

	// Record timeouts - score should decrease
	initialScore = updated.Score
	pr.RecordTimeout("score-record-test")
	updated = pr.GetPeer("score-record-test")
	if updated.Score >= initialScore {
		t.Errorf("score should decrease after timeout, got %f (was %f)", updated.Score, initialScore)
	}

	// Score should be clamped to min/max
	for i := 0; i < 100; i++ {
		pr.RecordSuccess("score-record-test")
	}
	updated = pr.GetPeer("score-record-test")
	if updated.Score > ScoreMaximum {
		t.Errorf("score should be clamped to max %f, got %f", ScoreMaximum, updated.Score)
	}

	for i := 0; i < 100; i++ {
		pr.RecordFailure("score-record-test")
	}
	updated = pr.GetPeer("score-record-test")
	if updated.Score < ScoreMinimum {
		t.Errorf("score should be clamped to min %f, got %f", ScoreMinimum, updated.Score)
	}
}

func TestPeerRegistry_GetPeersByScore(t *testing.T) {
	pr, cleanup := setupTestPeerRegistry(t)
	defer cleanup()

	// Add peers with different scores
	peers := []*Peer{
		{ID: "low-score", Name: "Low", Score: 20},
		{ID: "high-score", Name: "High", Score: 90},
		{ID: "mid-score", Name: "Mid", Score: 50},
	}

	for _, p := range peers {
		pr.AddPeer(p)
	}

	sorted := pr.GetPeersByScore()
	if len(sorted) != 3 {
		t.Fatalf("expected 3 peers, got %d", len(sorted))
	}

	// Should be sorted by score descending
	if sorted[0].ID != "high-score" {
		t.Errorf("first peer should be high-score, got %s", sorted[0].ID)
	}
	if sorted[1].ID != "mid-score" {
		t.Errorf("second peer should be mid-score, got %s", sorted[1].ID)
	}
	if sorted[2].ID != "low-score" {
		t.Errorf("third peer should be low-score, got %s", sorted[2].ID)
	}
}
