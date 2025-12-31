package mining

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/Snider/Mining/pkg/logging"
	"github.com/adrg/xdg"
)

// LogBuffer is a thread-safe ring buffer for capturing miner output.
type LogBuffer struct {
	lines    []string
	maxLines int
	mu       sync.RWMutex
}

// NewLogBuffer creates a new log buffer with the specified max lines.
func NewLogBuffer(maxLines int) *LogBuffer {
	return &LogBuffer{
		lines:    make([]string, 0, maxLines),
		maxLines: maxLines,
	}
}

// maxLineLength is the maximum length of a single log line to prevent memory bloat.
const maxLineLength = 2000

// Write implements io.Writer for capturing output.
func (lb *LogBuffer) Write(p []byte) (n int, err error) {
	lb.mu.Lock()
	defer lb.mu.Unlock()

	// Split input into lines
	text := string(p)
	newLines := strings.Split(text, "\n")

	for _, line := range newLines {
		if line == "" {
			continue
		}
		// Truncate excessively long lines to prevent memory bloat
		if len(line) > maxLineLength {
			line = line[:maxLineLength] + "... [truncated]"
		}
		// Add timestamp prefix
		timestampedLine := fmt.Sprintf("[%s] %s", time.Now().Format("15:04:05"), line)
		lb.lines = append(lb.lines, timestampedLine)

		// Trim if over max - force reallocation to release memory
		if len(lb.lines) > lb.maxLines {
			newSlice := make([]string, lb.maxLines)
			copy(newSlice, lb.lines[len(lb.lines)-lb.maxLines:])
			lb.lines = newSlice
		}
	}
	return len(p), nil
}

// GetLines returns all captured log lines.
func (lb *LogBuffer) GetLines() []string {
	lb.mu.RLock()
	defer lb.mu.RUnlock()
	result := make([]string, len(lb.lines))
	copy(result, lb.lines)
	return result
}

// Clear clears the log buffer.
func (lb *LogBuffer) Clear() {
	lb.mu.Lock()
	defer lb.mu.Unlock()
	lb.lines = lb.lines[:0]
}

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
	stdinPipe             io.WriteCloser  `json:"-"`
	HashrateHistory       []HashratePoint `json:"hashrateHistory"`
	LowResHashrateHistory []HashratePoint `json:"lowResHashrateHistory"`
	LastLowResAggregation time.Time       `json:"-"`
	LogBuffer             *LogBuffer      `json:"-"`
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

// Stop terminates the miner process gracefully.
// It first tries SIGTERM to allow cleanup, then SIGKILL if needed.
func (b *BaseMiner) Stop() error {
	b.mu.Lock()

	if !b.Running || b.cmd == nil {
		b.mu.Unlock()
		return errors.New("miner is not running")
	}

	// Close stdin pipe if open
	if b.stdinPipe != nil {
		b.stdinPipe.Close()
		b.stdinPipe = nil
	}

	// Capture cmd locally to avoid race with Wait() goroutine
	cmd := b.cmd
	process := cmd.Process

	// Mark as not running immediately to prevent concurrent Stop() calls
	b.Running = false
	b.cmd = nil
	b.mu.Unlock()

	// Try graceful shutdown with SIGTERM first (Unix only)
	if runtime.GOOS != "windows" {
		if err := process.Signal(syscall.SIGTERM); err == nil {
			// Wait up to 3 seconds for graceful shutdown
			done := make(chan struct{})
			go func() {
				process.Wait()
				close(done)
			}()

			select {
			case <-done:
				return nil
			case <-time.After(3 * time.Second):
				// Process didn't exit gracefully, force kill below
			}
		}
	}

	// Force kill and wait for process to exit
	if err := process.Kill(); err != nil {
		return err
	}

	// Wait for process to fully terminate to avoid zombies
	process.Wait()
	return nil
}

// stdinWriteTimeout is the maximum time to wait for stdin write to complete.
const stdinWriteTimeout = 5 * time.Second

// WriteStdin sends input to the miner's stdin (for console commands).
func (b *BaseMiner) WriteStdin(input string) error {
	b.mu.RLock()
	stdinPipe := b.stdinPipe
	running := b.Running
	b.mu.RUnlock()

	if !running || stdinPipe == nil {
		return errors.New("miner is not running or stdin not available")
	}

	// Append newline if not present
	if !strings.HasSuffix(input, "\n") {
		input += "\n"
	}

	// Write with timeout to prevent blocking indefinitely
	done := make(chan error, 1)
	go func() {
		_, err := stdinPipe.Write([]byte(input))
		done <- err
	}()

	select {
	case err := <-done:
		return err
	case <-time.After(stdinWriteTimeout):
		return errors.New("stdin write timeout: miner may be unresponsive")
	}
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

	resp, err := getHTTPClient().Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		io.Copy(io.Discard, resp.Body) // Drain body to allow connection reuse
		return fmt.Errorf("failed to download release: unexpected status code %d", resp.StatusCode)
	}

	if _, err := io.Copy(tmpfile, resp.Body); err != nil {
		// Drain remaining body to allow connection reuse
		io.Copy(io.Discard, resp.Body)
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
				logging.Debug("found miner binary at highest versioned path", logging.Fields{"path": fullPath})
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
		logging.Debug("found miner binary in system PATH", logging.Fields{"path": absPath})
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

// GetHighResHistoryLength returns the number of high-resolution hashrate points.
func (b *BaseMiner) GetHighResHistoryLength() int {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return len(b.HashrateHistory)
}

// GetLowResHistoryLength returns the number of low-resolution hashrate points.
func (b *BaseMiner) GetLowResHistoryLength() int {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return len(b.LowResHashrateHistory)
}

// GetLogs returns the captured log output from the miner process.
func (b *BaseMiner) GetLogs() []string {
	b.mu.RLock()
	logBuffer := b.LogBuffer
	b.mu.RUnlock()

	if logBuffer == nil {
		return []string{}
	}
	return logBuffer.GetLines()
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
	// Force reallocation if significantly oversized to free memory
	if cap(b.HashrateHistory) > 1000 && len(newHighResHistory) < cap(b.HashrateHistory)/2 {
		trimmed := make([]HashratePoint, len(newHighResHistory))
		copy(trimmed, newHighResHistory)
		b.HashrateHistory = trimmed
	} else {
		b.HashrateHistory = newHighResHistory
	}

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

	// Force reallocation if significantly oversized to free memory
	newLowResLen := len(b.LowResHashrateHistory) - firstValidLowResIndex
	if cap(b.LowResHashrateHistory) > 1000 && newLowResLen < cap(b.LowResHashrateHistory)/2 {
		trimmed := make([]HashratePoint, newLowResLen)
		copy(trimmed, b.LowResHashrateHistory[firstValidLowResIndex:])
		b.LowResHashrateHistory = trimmed
	} else {
		b.LowResHashrateHistory = b.LowResHashrateHistory[firstValidLowResIndex:]
	}
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
			return fmt.Errorf("%s: illegal file path in archive", header.Name)
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
