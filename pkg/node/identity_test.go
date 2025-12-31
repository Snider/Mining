package node

import (
	"os"
	"path/filepath"
	"testing"
)

// setupTestNodeManager creates a NodeManager with paths in a temp directory.
func setupTestNodeManager(t *testing.T) (*NodeManager, func()) {
	tmpDir, err := os.MkdirTemp("", "node-identity-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}

	keyPath := filepath.Join(tmpDir, "private.key")
	configPath := filepath.Join(tmpDir, "node.json")

	nm, err := NewNodeManagerWithPaths(keyPath, configPath)
	if err != nil {
		os.RemoveAll(tmpDir)
		t.Fatalf("failed to create node manager: %v", err)
	}

	cleanup := func() {
		os.RemoveAll(tmpDir)
	}

	return nm, cleanup
}

func TestNodeIdentity(t *testing.T) {
	t.Run("NewNodeManager", func(t *testing.T) {
		nm, cleanup := setupTestNodeManager(t)
		defer cleanup()

		if nm.HasIdentity() {
			t.Error("new node manager should not have identity")
		}
	})

	t.Run("GenerateIdentity", func(t *testing.T) {
		nm, cleanup := setupTestNodeManager(t)
		defer cleanup()

		err := nm.GenerateIdentity("test-node", RoleDual)
		if err != nil {
			t.Fatalf("failed to generate identity: %v", err)
		}

		if !nm.HasIdentity() {
			t.Error("node manager should have identity after generation")
		}

		identity := nm.GetIdentity()
		if identity == nil {
			t.Fatal("identity should not be nil")
		}

		if identity.Name != "test-node" {
			t.Errorf("expected name 'test-node', got '%s'", identity.Name)
		}

		if identity.Role != RoleDual {
			t.Errorf("expected role Dual, got '%s'", identity.Role)
		}

		if identity.ID == "" {
			t.Error("identity ID should not be empty")
		}

		if identity.PublicKey == "" {
			t.Error("public key should not be empty")
		}
	})

	t.Run("LoadExistingIdentity", func(t *testing.T) {
		tmpDir, err := os.MkdirTemp("", "node-load-test")
		if err != nil {
			t.Fatalf("failed to create temp dir: %v", err)
		}
		defer os.RemoveAll(tmpDir)

		keyPath := filepath.Join(tmpDir, "private.key")
		configPath := filepath.Join(tmpDir, "node.json")

		// First, create an identity
		nm1, err := NewNodeManagerWithPaths(keyPath, configPath)
		if err != nil {
			t.Fatalf("failed to create first node manager: %v", err)
		}

		err = nm1.GenerateIdentity("persistent-node", RoleWorker)
		if err != nil {
			t.Fatalf("failed to generate identity: %v", err)
		}

		originalID := nm1.GetIdentity().ID
		originalPubKey := nm1.GetIdentity().PublicKey

		// Create a new manager - should load existing identity
		nm2, err := NewNodeManagerWithPaths(keyPath, configPath)
		if err != nil {
			t.Fatalf("failed to create second node manager: %v", err)
		}

		if !nm2.HasIdentity() {
			t.Error("second node manager should have loaded existing identity")
		}

		identity := nm2.GetIdentity()
		if identity.ID != originalID {
			t.Errorf("expected ID '%s', got '%s'", originalID, identity.ID)
		}

		if identity.PublicKey != originalPubKey {
			t.Error("public key mismatch after reload")
		}
	})

	t.Run("DeriveSharedSecret", func(t *testing.T) {
		// Create two node managers with separate temp directories
		tmpDir1, _ := os.MkdirTemp("", "node1")
		tmpDir2, _ := os.MkdirTemp("", "node2")
		defer os.RemoveAll(tmpDir1)
		defer os.RemoveAll(tmpDir2)

		// Node 1
		nm1, err := NewNodeManagerWithPaths(
			filepath.Join(tmpDir1, "private.key"),
			filepath.Join(tmpDir1, "node.json"),
		)
		if err != nil {
			t.Fatalf("failed to create node manager 1: %v", err)
		}
		err = nm1.GenerateIdentity("node1", RoleDual)
		if err != nil {
			t.Fatalf("failed to generate identity 1: %v", err)
		}

		// Node 2
		nm2, err := NewNodeManagerWithPaths(
			filepath.Join(tmpDir2, "private.key"),
			filepath.Join(tmpDir2, "node.json"),
		)
		if err != nil {
			t.Fatalf("failed to create node manager 2: %v", err)
		}
		err = nm2.GenerateIdentity("node2", RoleDual)
		if err != nil {
			t.Fatalf("failed to generate identity 2: %v", err)
		}

		// Derive shared secrets - should be identical
		secret1, err := nm1.DeriveSharedSecret(nm2.GetIdentity().PublicKey)
		if err != nil {
			t.Fatalf("failed to derive shared secret from node 1: %v", err)
		}

		secret2, err := nm2.DeriveSharedSecret(nm1.GetIdentity().PublicKey)
		if err != nil {
			t.Fatalf("failed to derive shared secret from node 2: %v", err)
		}

		if len(secret1) != len(secret2) {
			t.Errorf("shared secrets have different lengths: %d vs %d", len(secret1), len(secret2))
		}

		for i := range secret1 {
			if secret1[i] != secret2[i] {
				t.Error("shared secrets do not match")
				break
			}
		}
	})

	t.Run("DeleteIdentity", func(t *testing.T) {
		nm, cleanup := setupTestNodeManager(t)
		defer cleanup()

		err := nm.GenerateIdentity("delete-me", RoleDual)
		if err != nil {
			t.Fatalf("failed to generate identity: %v", err)
		}

		if !nm.HasIdentity() {
			t.Error("should have identity before delete")
		}

		err = nm.Delete()
		if err != nil {
			t.Fatalf("failed to delete identity: %v", err)
		}

		if nm.HasIdentity() {
			t.Error("should not have identity after delete")
		}
	})
}

func TestNodeRoles(t *testing.T) {
	tests := []struct {
		role     NodeRole
		expected string
	}{
		{RoleController, "controller"},
		{RoleWorker, "worker"},
		{RoleDual, "dual"},
	}

	for _, tt := range tests {
		t.Run(string(tt.role), func(t *testing.T) {
			if string(tt.role) != tt.expected {
				t.Errorf("expected '%s', got '%s'", tt.expected, string(tt.role))
			}
		})
	}
}

func TestChallengeResponse(t *testing.T) {
	t.Run("GenerateChallenge", func(t *testing.T) {
		challenge, err := GenerateChallenge()
		if err != nil {
			t.Fatalf("failed to generate challenge: %v", err)
		}

		if len(challenge) != ChallengeSize {
			t.Errorf("expected challenge size %d, got %d", ChallengeSize, len(challenge))
		}

		// Ensure challenges are unique (not all zeros)
		allZero := true
		for _, b := range challenge {
			if b != 0 {
				allZero = false
				break
			}
		}
		if allZero {
			t.Error("challenge should not be all zeros")
		}

		// Generate another and ensure they're different
		challenge2, err := GenerateChallenge()
		if err != nil {
			t.Fatalf("failed to generate second challenge: %v", err)
		}

		same := true
		for i := range challenge {
			if challenge[i] != challenge2[i] {
				same = false
				break
			}
		}
		if same {
			t.Error("two generated challenges should be different")
		}
	})

	t.Run("SignAndVerifyChallenge", func(t *testing.T) {
		challenge, _ := GenerateChallenge()
		sharedSecret := []byte("test-secret-key-32-bytes-long!!")

		// Sign the challenge
		signature := SignChallenge(challenge, sharedSecret)

		if len(signature) == 0 {
			t.Error("signature should not be empty")
		}

		// Verify should succeed with correct parameters
		if !VerifyChallenge(challenge, signature, sharedSecret) {
			t.Error("verification should succeed with correct parameters")
		}

		// Verify should fail with wrong challenge
		wrongChallenge, _ := GenerateChallenge()
		if VerifyChallenge(wrongChallenge, signature, sharedSecret) {
			t.Error("verification should fail with wrong challenge")
		}

		// Verify should fail with wrong secret
		wrongSecret := []byte("wrong-secret-key-32-bytes-long!")
		if VerifyChallenge(challenge, signature, wrongSecret) {
			t.Error("verification should fail with wrong secret")
		}

		// Verify should fail with tampered signature
		tamperedSig := make([]byte, len(signature))
		copy(tamperedSig, signature)
		tamperedSig[0] ^= 0xFF // Flip bits
		if VerifyChallenge(challenge, tamperedSig, sharedSecret) {
			t.Error("verification should fail with tampered signature")
		}
	})

	t.Run("SignatureIsDeterministic", func(t *testing.T) {
		challenge := []byte("fixed-challenge-for-testing")
		sharedSecret := []byte("fixed-secret-key-for-testing")

		sig1 := SignChallenge(challenge, sharedSecret)
		sig2 := SignChallenge(challenge, sharedSecret)

		if len(sig1) != len(sig2) {
			t.Fatal("signatures should have same length")
		}

		for i := range sig1 {
			if sig1[i] != sig2[i] {
				t.Fatal("signatures should be identical for same inputs")
			}
		}
	})

	t.Run("IntegrationWithSharedSecret", func(t *testing.T) {
		// Create two nodes and test end-to-end challenge-response
		tmpDir1, _ := os.MkdirTemp("", "node-challenge-1")
		tmpDir2, _ := os.MkdirTemp("", "node-challenge-2")
		defer os.RemoveAll(tmpDir1)
		defer os.RemoveAll(tmpDir2)

		nm1, _ := NewNodeManagerWithPaths(
			filepath.Join(tmpDir1, "private.key"),
			filepath.Join(tmpDir1, "node.json"),
		)
		nm1.GenerateIdentity("challenger", RoleDual)

		nm2, _ := NewNodeManagerWithPaths(
			filepath.Join(tmpDir2, "private.key"),
			filepath.Join(tmpDir2, "node.json"),
		)
		nm2.GenerateIdentity("responder", RoleDual)

		// Challenger generates challenge
		challenge, err := GenerateChallenge()
		if err != nil {
			t.Fatalf("failed to generate challenge: %v", err)
		}

		// Both derive the same shared secret
		secret1, _ := nm1.DeriveSharedSecret(nm2.GetIdentity().PublicKey)
		secret2, _ := nm2.DeriveSharedSecret(nm1.GetIdentity().PublicKey)

		// Responder signs challenge with their derived secret
		response := SignChallenge(challenge, secret2)

		// Challenger verifies with their derived secret
		if !VerifyChallenge(challenge, response, secret1) {
			t.Error("challenge-response should verify with matching shared secrets")
		}
	})
}
