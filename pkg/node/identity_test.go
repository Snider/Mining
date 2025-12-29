package node

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNodeIdentity(t *testing.T) {
	// Create temp directory for test
	tmpDir, err := os.MkdirTemp("", "node-identity-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Override XDG paths for testing
	os.Setenv("XDG_CONFIG_HOME", filepath.Join(tmpDir, "config"))
	os.Setenv("XDG_DATA_HOME", filepath.Join(tmpDir, "data"))
	defer func() {
		os.Unsetenv("XDG_CONFIG_HOME")
		os.Unsetenv("XDG_DATA_HOME")
	}()

	t.Run("NewNodeManager", func(t *testing.T) {
		nm, err := NewNodeManager()
		if err != nil {
			t.Fatalf("failed to create node manager: %v", err)
		}

		if nm.HasIdentity() {
			t.Error("new node manager should not have identity")
		}
	})

	t.Run("GenerateIdentity", func(t *testing.T) {
		nm, err := NewNodeManager()
		if err != nil {
			t.Fatalf("failed to create node manager: %v", err)
		}

		err = nm.GenerateIdentity("test-node", RoleDual)
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
		// First, create an identity
		nm1, err := NewNodeManager()
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
		nm2, err := NewNodeManager()
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
		// Create two node managers with fresh XDG paths
		tmpDir1, _ := os.MkdirTemp("", "node1")
		tmpDir2, _ := os.MkdirTemp("", "node2")
		defer os.RemoveAll(tmpDir1)
		defer os.RemoveAll(tmpDir2)

		// Node 1
		os.Setenv("XDG_CONFIG_HOME", filepath.Join(tmpDir1, "config"))
		os.Setenv("XDG_DATA_HOME", filepath.Join(tmpDir1, "data"))

		nm1, err := NewNodeManager()
		if err != nil {
			t.Fatalf("failed to create node manager 1: %v", err)
		}
		err = nm1.GenerateIdentity("node1", RoleDual)
		if err != nil {
			t.Fatalf("failed to generate identity 1: %v", err)
		}

		// Node 2
		os.Setenv("XDG_CONFIG_HOME", filepath.Join(tmpDir2, "config"))
		os.Setenv("XDG_DATA_HOME", filepath.Join(tmpDir2, "data"))

		nm2, err := NewNodeManager()
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
		tmpDirDel, _ := os.MkdirTemp("", "node-delete")
		defer os.RemoveAll(tmpDirDel)

		os.Setenv("XDG_CONFIG_HOME", filepath.Join(tmpDirDel, "config"))
		os.Setenv("XDG_DATA_HOME", filepath.Join(tmpDirDel, "data"))

		nm, err := NewNodeManager()
		if err != nil {
			t.Fatalf("failed to create node manager: %v", err)
		}

		err = nm.GenerateIdentity("delete-me", RoleDual)
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
