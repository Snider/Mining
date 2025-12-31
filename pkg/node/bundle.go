package node

import (
	"archive/tar"
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/Snider/Borg/pkg/datanode"
	"github.com/Snider/Borg/pkg/tim"
)

// BundleType defines the type of deployment bundle.
type BundleType string

const (
	BundleProfile BundleType = "profile" // Just config/profile JSON
	BundleMiner   BundleType = "miner"   // Miner binary + config
	BundleFull    BundleType = "full"    // Everything (miner + profiles + config)
)

// Bundle represents a deployment bundle for P2P transfer.
type Bundle struct {
	Type     BundleType `json:"type"`
	Name     string     `json:"name"`
	Data     []byte     `json:"data"`     // Encrypted STIM data or raw JSON
	Checksum string     `json:"checksum"` // SHA-256 of Data
}

// BundleManifest describes the contents of a bundle.
type BundleManifest struct {
	Type       BundleType `json:"type"`
	Name       string     `json:"name"`
	Version    string     `json:"version,omitempty"`
	MinerType  string     `json:"minerType,omitempty"`
	ProfileIDs []string   `json:"profileIds,omitempty"`
	CreatedAt  string     `json:"createdAt"`
}

// CreateProfileBundle creates an encrypted bundle containing a mining profile.
func CreateProfileBundle(profileJSON []byte, name string, password string) (*Bundle, error) {
	// Create a TIM with just the profile config
	t, err := tim.New()
	if err != nil {
		return nil, fmt.Errorf("failed to create TIM: %w", err)
	}
	t.Config = profileJSON

	// Encrypt to STIM format
	stimData, err := t.ToSigil(password)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt bundle: %w", err)
	}

	// Calculate checksum
	checksum := calculateChecksum(stimData)

	return &Bundle{
		Type:     BundleProfile,
		Name:     name,
		Data:     stimData,
		Checksum: checksum,
	}, nil
}

// CreateProfileBundleUnencrypted creates a plain JSON bundle (for testing or trusted networks).
func CreateProfileBundleUnencrypted(profileJSON []byte, name string) (*Bundle, error) {
	checksum := calculateChecksum(profileJSON)

	return &Bundle{
		Type:     BundleProfile,
		Name:     name,
		Data:     profileJSON,
		Checksum: checksum,
	}, nil
}

// CreateMinerBundle creates an encrypted bundle containing a miner binary and optional profile.
func CreateMinerBundle(minerPath string, profileJSON []byte, name string, password string) (*Bundle, error) {
	// Read miner binary
	minerData, err := os.ReadFile(minerPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read miner binary: %w", err)
	}

	// Create a tarball with the miner binary
	tarData, err := createTarball(map[string][]byte{
		filepath.Base(minerPath): minerData,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create tarball: %w", err)
	}

	// Create DataNode from tarball
	dn, err := datanode.FromTar(tarData)
	if err != nil {
		return nil, fmt.Errorf("failed to create datanode: %w", err)
	}

	// Create TIM from DataNode
	t, err := tim.FromDataNode(dn)
	if err != nil {
		return nil, fmt.Errorf("failed to create TIM: %w", err)
	}

	// Set profile as config if provided
	if profileJSON != nil {
		t.Config = profileJSON
	}

	// Encrypt to STIM format
	stimData, err := t.ToSigil(password)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt bundle: %w", err)
	}

	checksum := calculateChecksum(stimData)

	return &Bundle{
		Type:     BundleMiner,
		Name:     name,
		Data:     stimData,
		Checksum: checksum,
	}, nil
}

// ExtractProfileBundle decrypts and extracts a profile bundle.
func ExtractProfileBundle(bundle *Bundle, password string) ([]byte, error) {
	// Verify checksum first
	if calculateChecksum(bundle.Data) != bundle.Checksum {
		return nil, fmt.Errorf("checksum mismatch - bundle may be corrupted")
	}

	// If it's unencrypted JSON, just return it
	if isJSON(bundle.Data) {
		return bundle.Data, nil
	}

	// Decrypt STIM format
	t, err := tim.FromSigil(bundle.Data, password)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt bundle: %w", err)
	}

	return t.Config, nil
}

// ExtractMinerBundle decrypts and extracts a miner bundle, returning the miner path and profile.
func ExtractMinerBundle(bundle *Bundle, password string, destDir string) (string, []byte, error) {
	// Verify checksum
	if calculateChecksum(bundle.Data) != bundle.Checksum {
		return "", nil, fmt.Errorf("checksum mismatch - bundle may be corrupted")
	}

	// Decrypt STIM format
	t, err := tim.FromSigil(bundle.Data, password)
	if err != nil {
		return "", nil, fmt.Errorf("failed to decrypt bundle: %w", err)
	}

	// Convert rootfs to tarball and extract
	tarData, err := t.RootFS.ToTar()
	if err != nil {
		return "", nil, fmt.Errorf("failed to convert rootfs to tar: %w", err)
	}

	// Extract tarball to destination
	minerPath, err := extractTarball(tarData, destDir)
	if err != nil {
		return "", nil, fmt.Errorf("failed to extract tarball: %w", err)
	}

	return minerPath, t.Config, nil
}

// VerifyBundle checks if a bundle's checksum is valid.
func VerifyBundle(bundle *Bundle) bool {
	return calculateChecksum(bundle.Data) == bundle.Checksum
}

// calculateChecksum computes SHA-256 checksum of data.
func calculateChecksum(data []byte) string {
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:])
}

// isJSON checks if data starts with JSON characters.
func isJSON(data []byte) bool {
	if len(data) == 0 {
		return false
	}
	// JSON typically starts with { or [
	return data[0] == '{' || data[0] == '['
}

// createTarball creates a tar archive from a map of filename -> content.
func createTarball(files map[string][]byte) ([]byte, error) {
	var buf bytes.Buffer
	tw := tar.NewWriter(&buf)

	// Track directories we've created
	dirs := make(map[string]bool)

	for name, content := range files {
		// Create parent directories if needed
		dir := filepath.Dir(name)
		if dir != "." && !dirs[dir] {
			hdr := &tar.Header{
				Name:     dir + "/",
				Mode:     0755,
				Typeflag: tar.TypeDir,
			}
			if err := tw.WriteHeader(hdr); err != nil {
				return nil, err
			}
			dirs[dir] = true
		}

		// Determine file mode (executable for binaries in miners/)
		mode := int64(0644)
		if filepath.Dir(name) == "miners" || !isJSON(content) {
			mode = 0755
		}

		hdr := &tar.Header{
			Name: name,
			Mode: mode,
			Size: int64(len(content)),
		}
		if err := tw.WriteHeader(hdr); err != nil {
			return nil, err
		}
		if _, err := tw.Write(content); err != nil {
			return nil, err
		}
	}

	if err := tw.Close(); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// extractTarball extracts a tar archive to a directory, returns first executable found.
func extractTarball(tarData []byte, destDir string) (string, error) {
	// Ensure destDir is an absolute, clean path for security checks
	absDestDir, err := filepath.Abs(destDir)
	if err != nil {
		return "", fmt.Errorf("failed to resolve destination directory: %w", err)
	}
	absDestDir = filepath.Clean(absDestDir)

	if err := os.MkdirAll(absDestDir, 0755); err != nil {
		return "", err
	}

	tr := tar.NewReader(bytes.NewReader(tarData))
	var firstExecutable string

	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", err
		}

		// Security: Sanitize the tar entry name to prevent path traversal (Zip Slip)
		cleanName := filepath.Clean(hdr.Name)

		// Reject absolute paths
		if filepath.IsAbs(cleanName) {
			return "", fmt.Errorf("invalid tar entry: absolute path not allowed: %s", hdr.Name)
		}

		// Reject paths that escape the destination directory
		if strings.HasPrefix(cleanName, ".."+string(os.PathSeparator)) || cleanName == ".." {
			return "", fmt.Errorf("invalid tar entry: path traversal attempt: %s", hdr.Name)
		}

		// Build the full path and verify it's within destDir
		fullPath := filepath.Join(absDestDir, cleanName)
		fullPath = filepath.Clean(fullPath)

		// Final security check: ensure the path is still within destDir
		if !strings.HasPrefix(fullPath, absDestDir+string(os.PathSeparator)) && fullPath != absDestDir {
			return "", fmt.Errorf("invalid tar entry: path escape attempt: %s", hdr.Name)
		}

		switch hdr.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(fullPath, os.FileMode(hdr.Mode)); err != nil {
				return "", err
			}
		case tar.TypeReg:
			// Ensure parent directory exists
			if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
				return "", err
			}

			f, err := os.OpenFile(fullPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, os.FileMode(hdr.Mode))
			if err != nil {
				return "", err
			}

			// Limit file size to prevent decompression bombs (100MB max per file)
			const maxFileSize int64 = 100 * 1024 * 1024
			limitedReader := io.LimitReader(tr, maxFileSize+1)
			written, err := io.Copy(f, limitedReader)
			f.Close()
			if err != nil {
				return "", err
			}
			if written > maxFileSize {
				os.Remove(fullPath)
				return "", fmt.Errorf("file %s exceeds maximum size of %d bytes", hdr.Name, maxFileSize)
			}

			// Track first executable
			if hdr.Mode&0111 != 0 && firstExecutable == "" {
				firstExecutable = fullPath
			}
		// Explicitly ignore symlinks and hard links to prevent symlink attacks
		case tar.TypeSymlink, tar.TypeLink:
			// Skip symlinks and hard links for security
			continue
		}
	}

	return firstExecutable, nil
}

// StreamBundle writes a bundle to a writer (for large transfers).
func StreamBundle(bundle *Bundle, w io.Writer) error {
	encoder := json.NewEncoder(w)
	return encoder.Encode(bundle)
}

// ReadBundle reads a bundle from a reader.
func ReadBundle(r io.Reader) (*Bundle, error) {
	var bundle Bundle
	decoder := json.NewDecoder(r)
	if err := decoder.Decode(&bundle); err != nil {
		return nil, err
	}
	return &bundle, nil
}
