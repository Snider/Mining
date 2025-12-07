package mining

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/adrg/xdg"
)

// BaseMiner provides a foundation for specific miner implementations.
type BaseMiner struct {
	Name                  string `json:"name"`
	Version               string `json:"version"`
	URL                   string `json:"url"`
	Path                  string `json:"path"`
	MinerBinary           string `json:"miner_binary"`
	ExecutableName        string `json:"-"`
	Running               bool   `json:"running"`
	ConfigPath            string `json:"configPath"`
	API                   *API   `json:"api"`
	mu                    sync.RWMutex
	cmd                   *exec.Cmd
	HashrateHistory       []HashratePoint `json:"hashrateHistory"`
	LowResHashrateHistory []HashratePoint `json:"lowResHashrateHistory"`
	LastLowResAggregation time.Time       `json:"-"`
}

// GetName returns the name of the miner.
func (b *BaseMiner) GetName() string {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.Name
}

// GetPath returns the base installation directory for the miner type.
// It uses the stable ExecutableName field to ensure the correct path.
func (b *BaseMiner) GetPath() string {
	dataPath, err := xdg.DataFile(fmt.Sprintf("lethean-desktop/miners/%s", b.ExecutableName))
	if err != nil {
		home, err := os.UserHomeDir()
		if err != nil {
			return ""
		}
		return filepath.Join(home, ".lethean-desktop", "miners", b.ExecutableName)
	}
	return dataPath
}

// GetBinaryPath returns the full path to the miner's executable file.
func (b *BaseMiner) GetBinaryPath() string {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.MinerBinary
}

// Stop terminates the miner process.
func (b *BaseMiner) Stop() error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if !b.Running || b.cmd == nil {
		return errors.New("miner is not running")
	}

	return b.cmd.Process.Kill()
}

// Uninstall removes all files related to the miner.
func (b *BaseMiner) Uninstall() error {
	return os.RemoveAll(b.GetPath())
}

// InstallFromURL handles the generic download and extraction process for a miner.
func (b *BaseMiner) InstallFromURL(url string) error {
	tmpfile, err := os.CreateTemp("", b.ExecutableName+"-")
	if err != nil {
		return err
	}
	defer os.Remove(tmpfile.Name())
	defer tmpfile.Close()

	resp, err := httpClient.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to download release: unexpected status code %d", resp.StatusCode)
	}

	if _, err := io.Copy(tmpfile, resp.Body); err != nil {
		return err
	}

	baseInstallPath := b.GetPath()
	if err := os.MkdirAll(baseInstallPath, 0755); err != nil {
		return err
	}

	if strings.HasSuffix(url, ".zip") {
		err = b.unzip(tmpfile.Name(), baseInstallPath)
	} else {
		err = b.untar(tmpfile.Name(), baseInstallPath)
	}
	if err != nil {
		return fmt.Errorf("failed to extract miner: %w", err)
	}

	return nil
}

// parseVersion parses a version string (e.g., "6.24.0") into a slice of integers for comparison.
func parseVersion(v string) []int {
	parts := strings.Split(v, ".")
	intParts := make([]int, len(parts))
	for i, p := range parts {
		val, err := strconv.Atoi(p)
		if err != nil {
			return []int{0} // Malformed version, treat as very old
		}
		intParts[i] = val
	}
	return intParts
}

// compareVersions compares two version slices. Returns 1 if v1 > v2, -1 if v1 < v2, 0 if equal.
func compareVersions(v1, v2 []int) int {
	minLen := len(v1)
	if len(v2) < minLen {
		minLen = len(v2)
	}

	for i := 0; i < minLen; i++ {
		if v1[i] > v2[i] {
			return 1
		}
		if v1[i] < v2[i] {
			return -1
		}
	}

	if len(v1) > len(v2) {
		return 1
	}
	if len(v1) < len(v2) {
		return -1
	}
	return 0
}

// findMinerBinary searches for the miner's executable file.
// It returns the absolute path to the executable if found, prioritizing the highest versioned installation.
func (b *BaseMiner) findMinerBinary() (string, error) {
	executableName := b.ExecutableName
	if runtime.GOOS == "windows" {
		executableName += ".exe"
	}

	baseInstallPath := b.GetPath()
	searchedPaths := []string{}

	var highestVersion []int
	var highestVersionDir string

	// 1. Check the standard installation directory first
	if _, err := os.Stat(baseInstallPath); err == nil {
		dirs, err := os.ReadDir(baseInstallPath)
		if err == nil {
			for _, d := range dirs {
				if d.IsDir() && strings.HasPrefix(d.Name(), b.ExecutableName+"-") {
					// Extract version string, e.g., "xmrig-6.24.0" -> "6.24.0"
					versionStr := strings.TrimPrefix(d.Name(), b.ExecutableName+"-")
					currentVersion := parseVersion(versionStr)

					if highestVersionDir == "" || compareVersions(currentVersion, highestVersion) > 0 {
						highestVersion = currentVersion
						highestVersionDir = d.Name()
					}
					versionedPath := filepath.Join(baseInstallPath, d.Name())
					fullPath := filepath.Join(versionedPath, executableName)
					searchedPaths = append(searchedPaths, fullPath)
				}
			}
		}

		if highestVersionDir != "" {
			fullPath := filepath.Join(baseInstallPath, highestVersionDir, executableName)
			if _, err := os.Stat(fullPath); err == nil {
				log.Printf("Found miner binary at highest versioned path: %s", fullPath)
				return fullPath, nil
			}
		}
	}

	// 2. Fallback to searching the system PATH
	path, err := exec.LookPath(executableName)
	if err == nil {
		absPath, err := filepath.Abs(path)
		if err != nil {
			return "", fmt.Errorf("failed to get absolute path for '%s': %w", path, err)
		}
		log.Printf("Found miner binary in system PATH: %s", absPath)
		return absPath, nil
	}

	// If not found, return a detailed error
	return "", fmt.Errorf("miner executable '%s' not found. Searched in: %s and system PATH", executableName, strings.Join(searchedPaths, ", "))
}

// CheckInstallation verifies if the miner is installed correctly.
func (b *BaseMiner) CheckInstallation() (*InstallationDetails, error) {
	binaryPath, err := b.findMinerBinary()
	if err != nil {
		return &InstallationDetails{IsInstalled: false}, err
	}

	b.MinerBinary = binaryPath
	b.Path = filepath.Dir(binaryPath)

	cmd := exec.Command(binaryPath, "--version")
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		b.Version = "Unknown (could not run executable)"
	} else {
		fields := strings.Fields(out.String())
		if len(fields) >= 2 {
			b.Version = fields[1]
		} else {
			b.Version = "Unknown (could not parse version)"
		}
	}

	return &InstallationDetails{
		IsInstalled: true,
		MinerBinary: b.MinerBinary,
		Path:        b.Path,
		Version:     b.Version,
	}, nil
}

// GetHashrateHistory returns the combined hashrate history.
func (b *BaseMiner) GetHashrateHistory() []HashratePoint {
	b.mu.RLock()
	defer b.mu.RUnlock()
	combinedHistory := make([]HashratePoint, 0, len(b.LowResHashrateHistory)+len(b.HashrateHistory))
	combinedHistory = append(combinedHistory, b.LowResHashrateHistory...)
	combinedHistory = append(combinedHistory, b.HashrateHistory...)
	return combinedHistory
}

// AddHashratePoint adds a new hashrate measurement.
func (b *BaseMiner) AddHashratePoint(point HashratePoint) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.HashrateHistory = append(b.HashrateHistory, point)
}

// ReduceHashrateHistory aggregates and trims hashrate data.
func (b *BaseMiner) ReduceHashrateHistory(now time.Time) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if !b.LastLowResAggregation.IsZero() && now.Sub(b.LastLowResAggregation) < LowResolutionInterval {
		return
	}

	var pointsToAggregate []HashratePoint
	var newHighResHistory []HashratePoint
	cutoff := now.Add(-HighResolutionDuration)

	for _, p := range b.HashrateHistory {
		if p.Timestamp.Before(cutoff) {
			pointsToAggregate = append(pointsToAggregate, p)
		} else {
			newHighResHistory = append(newHighResHistory, p)
		}
	}
	b.HashrateHistory = newHighResHistory

	if len(pointsToAggregate) == 0 {
		b.LastLowResAggregation = now
		return
	}

	minuteGroups := make(map[time.Time][]int)
	for _, p := range pointsToAggregate {
		minute := p.Timestamp.Truncate(LowResolutionInterval)
		minuteGroups[minute] = append(minuteGroups[minute], p.Hashrate)
	}

	var newLowResPoints []HashratePoint
	for minute, hashrates := range minuteGroups {
		if len(hashrates) > 0 {
			totalHashrate := 0
			for _, hr := range hashrates {
				totalHashrate += hr
			}
			avgHashrate := totalHashrate / len(hashrates)
			newLowResPoints = append(newLowResPoints, HashratePoint{Timestamp: minute, Hashrate: avgHashrate})
		}
	}

	sort.Slice(newLowResPoints, func(i, j int) bool {
		return newLowResPoints[i].Timestamp.Before(newLowResPoints[j].Timestamp)
	})

	b.LowResHashrateHistory = append(b.LowResHashrateHistory, newLowResPoints...)

	lowResCutoff := now.Add(-LowResHistoryRetention)
	firstValidLowResIndex := 0
	for i, p := range b.LowResHashrateHistory {
		if p.Timestamp.After(lowResCutoff) || p.Timestamp.Equal(lowResCutoff) {
			firstValidLowResIndex = i
			break
		}
		if i == len(b.LowResHashrateHistory)-1 {
			firstValidLowResIndex = len(b.LowResHashrateHistory)
		}
	}
	b.LowResHashrateHistory = b.LowResHashrateHistory[firstValidLowResIndex:]
	b.LastLowResAggregation = now
}

// unzip extracts a zip archive.
func (b *BaseMiner) unzip(src, dest string) error {
	r, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer r.Close()

	for _, f := range r.File {
		fpath := filepath.Join(dest, f.Name)
		if !strings.HasPrefix(fpath, filepath.Clean(dest)+string(os.PathSeparator)) {
			return fmt.Errorf("%s: illegal file path", fpath)
		}
		if f.FileInfo().IsDir() {
			os.MkdirAll(fpath, os.ModePerm)
			continue
		}

		if err = os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
			return err
		}
		outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return err
		}
		rc, err := f.Open()
		if err != nil {
			outFile.Close()
			return err
		}
		_, err = io.Copy(outFile, rc)
		outFile.Close()
		rc.Close()
		if err != nil {
			return err
		}
	}
	return nil
}

// untar extracts a tar.gz archive.
func (b *BaseMiner) untar(src, dest string) error {
	file, err := os.Open(src)
	if err != nil {
		return err
	}
	defer file.Close()

	gzr, err := gzip.NewReader(file)
	if err != nil {
		return err
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)

	for {
		header, err := tr.Next()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}

		target := filepath.Join(dest, header.Name)
		if !strings.HasPrefix(target, filepath.Clean(dest)+string(os.PathSeparator)) {
			continue
		}

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, 0755); err != nil {
				return err
			}
		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
				return err
			}
			f, err := os.OpenFile(target, os.O_CREATE|os.O_RDWR|os.O_TRUNC, os.FileMode(header.Mode))
			if err != nil {
				return err
			}
			if _, err := io.Copy(f, tr); err != nil {
				f.Close()
				return err
			}
			f.Close()
		}
	}
}
