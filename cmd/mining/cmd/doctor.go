package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Snider/Mining/pkg/mining"
	"github.com/adrg/xdg"
	"github.com/spf13/cobra"
)

const signpostFilename = ".installed-miners"

// validateConfigPath validates that a config path is within the expected XDG config directory
// This prevents path traversal attacks via manipulated signpost files
func validateConfigPath(configPath string) error {
	// Get the expected XDG config base directory
	expectedBase := filepath.Join(xdg.ConfigHome, "lethean-desktop")

	// Clean and resolve the config path
	cleanPath := filepath.Clean(configPath)

	// Check if the path is within the expected directory
	if !strings.HasPrefix(cleanPath, expectedBase+string(os.PathSeparator)) && cleanPath != expectedBase {
		return fmt.Errorf("invalid config path: must be within %s", expectedBase)
	}

	return nil
}

// doctorCmd represents the doctor command
var doctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Check and refresh the status of installed miners",
	Long:  `Performs a live check for installed miners, displays their status, and updates the local cache.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("--- Mining Doctor ---")
		fmt.Println("Performing live check and refreshing cache...")
		fmt.Println()

		if err := updateDoctorCache(); err != nil {
			return fmt.Errorf("failed to run doctor check: %w", err)
		}
		// After updating the cache, display the fresh results
		_, err := loadAndDisplayCache()
		return err
	},
}

func loadAndDisplayCache() (bool, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return false, fmt.Errorf("could not get home directory: %w", err)
	}
	signpostPath := filepath.Join(homeDir, signpostFilename)

	if _, err := os.Stat(signpostPath); os.IsNotExist(err) {
		fmt.Println("No cached data found. Run 'install' for a miner first.")
		return false, nil // No cache to load
	}

	configPathBytes, err := os.ReadFile(signpostPath)
	if err != nil {
		return false, fmt.Errorf("could not read signpost file: %w", err)
	}
	configPath := strings.TrimSpace(string(configPathBytes))

	// Security: Validate that the config path is within the expected directory
	if err := validateConfigPath(configPath); err != nil {
		return false, fmt.Errorf("security error: %w", err)
	}

	cacheBytes, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Println("No cached data found. Run 'install' for a miner first.")
			return false, nil
		}
		return false, fmt.Errorf("could not read cache file from %s: %w", configPath, err)
	}

	var systemInfo mining.SystemInfo
	if err := json.Unmarshal(cacheBytes, &systemInfo); err != nil {
		return false, fmt.Errorf("could not parse cache file: %w", err)
	}

	fmt.Printf("System Info (cached at %s):\n", systemInfo.Timestamp.Format(time.RFC1123))
	fmt.Printf("  OS: %s, Arch: %s\n", systemInfo.OS, systemInfo.Architecture)
	fmt.Println()

	for _, details := range systemInfo.InstalledMinersInfo {
		// Infer miner name from path for display purposes
		var minerName string
		if details.Path != "" {
			if strings.Contains(details.Path, "xmrig") {
				minerName = "XMRig"
			} else {
				minerName = "Unknown Miner"
			}
		} else {
			minerName = "Unknown Miner"
		}
		displayDetails(minerName, details)
	}

	return true, nil
}

func saveResultsToCache(systemInfo *mining.SystemInfo) error {
	configDir, err := xdg.ConfigFile("lethean-desktop/miners")
	if err != nil {
		return fmt.Errorf("could not get config directory: %w", err)
	}
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("could not create config directory: %w", err)
	}
	configPath := filepath.Join(configDir, "config.json")

	data, err := json.MarshalIndent(systemInfo, "", "  ")
	if err != nil {
		return fmt.Errorf("could not marshal cache data: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0600); err != nil {
		return fmt.Errorf("could not write cache file: %w", err)
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("could not get home directory for signpost: %w", err)
	}
	signpostPath := filepath.Join(homeDir, signpostFilename)
	if err := os.WriteFile(signpostPath, []byte(configPath), 0600); err != nil {
		return fmt.Errorf("could not write signpost file: %w", err)
	}

	fmt.Printf("\n(Cache updated at %s)\n", configPath)
	return nil
}

func displayDetails(minerName string, details *mining.InstallationDetails) {
	fmt.Printf("--- %s ---\n", minerName)
	if details.IsInstalled {
		fmt.Printf("  Status:      Installed\n")
		fmt.Printf("  Version:     %s\n", details.Version)
		fmt.Printf("  Install Path: %s\n", details.Path)
		if details.MinerBinary != "" {
			fmt.Printf("  Miner Binary: %s\n", details.MinerBinary)
		}
		fmt.Println("  (Add this path to your AV scanner's whitelist to prevent interference)")
	} else {
		fmt.Printf("  Status:      Not Installed\n")
		fmt.Printf("  To install, run: install %s\n", strings.ToLower(minerName))
	}
	fmt.Println()
}

func init() {
	rootCmd.AddCommand(doctorCmd)
}
