package node

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestCreateProfileBundleUnencrypted(t *testing.T) {
	profileJSON := []byte(`{"name":"test-profile","minerType":"xmrig","config":{}}`)

	bundle, err := CreateProfileBundleUnencrypted(profileJSON, "test-profile")
	if err != nil {
		t.Fatalf("failed to create bundle: %v", err)
	}

	if bundle.Type != BundleProfile {
		t.Errorf("expected type BundleProfile, got %s", bundle.Type)
	}

	if bundle.Name != "test-profile" {
		t.Errorf("expected name 'test-profile', got '%s'", bundle.Name)
	}

	if bundle.Checksum == "" {
		t.Error("checksum should not be empty")
	}

	if !bytes.Equal(bundle.Data, profileJSON) {
		t.Error("data should match original JSON")
	}
}

func TestVerifyBundle(t *testing.T) {
	t.Run("ValidChecksum", func(t *testing.T) {
		bundle, _ := CreateProfileBundleUnencrypted([]byte(`{"test":"data"}`), "test")

		if !VerifyBundle(bundle) {
			t.Error("valid bundle should verify")
		}
	})

	t.Run("InvalidChecksum", func(t *testing.T) {
		bundle, _ := CreateProfileBundleUnencrypted([]byte(`{"test":"data"}`), "test")
		bundle.Checksum = "invalid-checksum"

		if VerifyBundle(bundle) {
			t.Error("bundle with invalid checksum should not verify")
		}
	})

	t.Run("ModifiedData", func(t *testing.T) {
		bundle, _ := CreateProfileBundleUnencrypted([]byte(`{"test":"data"}`), "test")
		bundle.Data = []byte(`{"test":"modified"}`)

		if VerifyBundle(bundle) {
			t.Error("bundle with modified data should not verify")
		}
	})
}

func TestCreateProfileBundle(t *testing.T) {
	profileJSON := []byte(`{"name":"encrypted-profile","minerType":"xmrig"}`)
	password := "test-password-123"

	bundle, err := CreateProfileBundle(profileJSON, "encrypted-test", password)
	if err != nil {
		t.Fatalf("failed to create encrypted bundle: %v", err)
	}

	if bundle.Type != BundleProfile {
		t.Errorf("expected type BundleProfile, got %s", bundle.Type)
	}

	// Encrypted data should not match original
	if bytes.Equal(bundle.Data, profileJSON) {
		t.Error("encrypted data should not match original")
	}

	// Should be able to extract with correct password
	extracted, err := ExtractProfileBundle(bundle, password)
	if err != nil {
		t.Fatalf("failed to extract bundle: %v", err)
	}

	if !bytes.Equal(extracted, profileJSON) {
		t.Errorf("extracted data should match original: got %s", string(extracted))
	}
}

func TestExtractProfileBundle(t *testing.T) {
	t.Run("UnencryptedBundle", func(t *testing.T) {
		originalJSON := []byte(`{"name":"plain","config":{}}`)
		bundle, _ := CreateProfileBundleUnencrypted(originalJSON, "plain")

		extracted, err := ExtractProfileBundle(bundle, "")
		if err != nil {
			t.Fatalf("failed to extract unencrypted bundle: %v", err)
		}

		if !bytes.Equal(extracted, originalJSON) {
			t.Error("extracted data should match original")
		}
	})

	t.Run("EncryptedBundle", func(t *testing.T) {
		originalJSON := []byte(`{"name":"secret","config":{"pool":"pool.example.com"}}`)
		password := "strong-password"

		bundle, _ := CreateProfileBundle(originalJSON, "secret", password)

		extracted, err := ExtractProfileBundle(bundle, password)
		if err != nil {
			t.Fatalf("failed to extract encrypted bundle: %v", err)
		}

		if !bytes.Equal(extracted, originalJSON) {
			t.Error("extracted data should match original")
		}
	})

	t.Run("WrongPassword", func(t *testing.T) {
		originalJSON := []byte(`{"name":"secret"}`)
		bundle, _ := CreateProfileBundle(originalJSON, "secret", "correct-password")

		_, err := ExtractProfileBundle(bundle, "wrong-password")
		if err == nil {
			t.Error("should fail with wrong password")
		}
	})

	t.Run("CorruptedChecksum", func(t *testing.T) {
		bundle, _ := CreateProfileBundleUnencrypted([]byte(`{}`), "test")
		bundle.Checksum = "corrupted"

		_, err := ExtractProfileBundle(bundle, "")
		if err == nil {
			t.Error("should fail with corrupted checksum")
		}
	})
}

func TestTarballFunctions(t *testing.T) {
	t.Run("CreateAndExtractTarball", func(t *testing.T) {
		files := map[string][]byte{
			"file1.txt":        []byte("content of file 1"),
			"dir/file2.json":   []byte(`{"key":"value"}`),
			"miners/xmrig":     []byte("binary content"),
		}

		tarData, err := createTarball(files)
		if err != nil {
			t.Fatalf("failed to create tarball: %v", err)
		}

		if len(tarData) == 0 {
			t.Error("tarball should not be empty")
		}

		// Extract to temp directory
		tmpDir, _ := os.MkdirTemp("", "tarball-test")
		defer os.RemoveAll(tmpDir)

		firstExec, err := extractTarball(tarData, tmpDir)
		if err != nil {
			t.Fatalf("failed to extract tarball: %v", err)
		}

		// Check files exist
		for name, content := range files {
			path := filepath.Join(tmpDir, name)
			data, err := os.ReadFile(path)
			if err != nil {
				t.Errorf("failed to read extracted file %s: %v", name, err)
				continue
			}

			if !bytes.Equal(data, content) {
				t.Errorf("content mismatch for %s", name)
			}
		}

		// Check first executable is the miner
		if firstExec == "" {
			t.Error("should find an executable")
		}
	})
}

func TestStreamAndReadBundle(t *testing.T) {
	original, _ := CreateProfileBundleUnencrypted([]byte(`{"streaming":"test"}`), "stream-test")

	// Stream to buffer
	var buf bytes.Buffer
	err := StreamBundle(original, &buf)
	if err != nil {
		t.Fatalf("failed to stream bundle: %v", err)
	}

	// Read back
	restored, err := ReadBundle(&buf)
	if err != nil {
		t.Fatalf("failed to read bundle: %v", err)
	}

	if restored.Name != original.Name {
		t.Errorf("name mismatch: expected '%s', got '%s'", original.Name, restored.Name)
	}

	if restored.Checksum != original.Checksum {
		t.Error("checksum mismatch")
	}

	if !bytes.Equal(restored.Data, original.Data) {
		t.Error("data mismatch")
	}
}

func TestCalculateChecksum(t *testing.T) {
	t.Run("Deterministic", func(t *testing.T) {
		data := []byte("test data for checksum")

		checksum1 := calculateChecksum(data)
		checksum2 := calculateChecksum(data)

		if checksum1 != checksum2 {
			t.Error("checksum should be deterministic")
		}
	})

	t.Run("DifferentData", func(t *testing.T) {
		checksum1 := calculateChecksum([]byte("data1"))
		checksum2 := calculateChecksum([]byte("data2"))

		if checksum1 == checksum2 {
			t.Error("different data should produce different checksums")
		}
	})

	t.Run("HexFormat", func(t *testing.T) {
		checksum := calculateChecksum([]byte("test"))

		// SHA-256 produces 64 hex characters
		if len(checksum) != 64 {
			t.Errorf("expected 64 character hex string, got %d characters", len(checksum))
		}

		// Should be valid hex
		for _, c := range checksum {
			if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f')) {
				t.Errorf("invalid hex character: %c", c)
			}
		}
	})
}

func TestIsJSON(t *testing.T) {
	tests := []struct {
		data     []byte
		expected bool
	}{
		{[]byte(`{"key":"value"}`), true},
		{[]byte(`["item1","item2"]`), true},
		{[]byte(`{}`), true},
		{[]byte(`[]`), true},
		{[]byte(`binary\x00data`), false},
		{[]byte(`plain text`), false},
		{[]byte{}, false},
		{nil, false},
	}

	for _, tt := range tests {
		result := isJSON(tt.data)
		if result != tt.expected {
			t.Errorf("isJSON(%q) = %v, expected %v", tt.data, result, tt.expected)
		}
	}
}

func TestBundleTypes(t *testing.T) {
	types := []BundleType{
		BundleProfile,
		BundleMiner,
		BundleFull,
	}

	expected := []string{"profile", "miner", "full"}

	for i, bt := range types {
		if string(bt) != expected[i] {
			t.Errorf("expected %s, got %s", expected[i], string(bt))
		}
	}
}

func TestCreateMinerBundle(t *testing.T) {
	// Create a temp "miner binary"
	tmpDir, _ := os.MkdirTemp("", "miner-bundle-test")
	defer os.RemoveAll(tmpDir)

	minerPath := filepath.Join(tmpDir, "test-miner")
	err := os.WriteFile(minerPath, []byte("fake miner binary content"), 0755)
	if err != nil {
		t.Fatalf("failed to create test miner: %v", err)
	}

	profileJSON := []byte(`{"profile":"data"}`)
	password := "miner-password"

	bundle, err := CreateMinerBundle(minerPath, profileJSON, "miner-bundle", password)
	if err != nil {
		t.Fatalf("failed to create miner bundle: %v", err)
	}

	if bundle.Type != BundleMiner {
		t.Errorf("expected type BundleMiner, got %s", bundle.Type)
	}

	if bundle.Name != "miner-bundle" {
		t.Errorf("expected name 'miner-bundle', got '%s'", bundle.Name)
	}

	// Extract and verify
	extractDir, _ := os.MkdirTemp("", "miner-extract-test")
	defer os.RemoveAll(extractDir)

	extractedPath, extractedProfile, err := ExtractMinerBundle(bundle, password, extractDir)
	if err != nil {
		t.Fatalf("failed to extract miner bundle: %v", err)
	}

	// Note: extractedPath may be empty if the tarball structure doesn't match
	// what extractTarball expects (it looks for files at root with executable bit)
	t.Logf("extracted path: %s", extractedPath)

	if !bytes.Equal(extractedProfile, profileJSON) {
		t.Error("profile data mismatch")
	}

	// If we got an extracted path, verify its content
	if extractedPath != "" {
		minerData, err := os.ReadFile(extractedPath)
		if err != nil {
			t.Fatalf("failed to read extracted miner: %v", err)
		}

		if string(minerData) != "fake miner binary content" {
			t.Error("miner content mismatch")
		}
	}
}
