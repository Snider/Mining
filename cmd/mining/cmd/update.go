package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Masterminds/semver/v3"
	"github.com/Snider/Mining/pkg/mining"
	"github.com/adrg/xdg"
	"github.com/spf13/cobra"
)

// validateUpdateConfigPath validates that a config path is within the expected XDG config directory
func validateUpdateConfigPath(configPath string) error {
	expectedBase := filepath.Join(xdg.ConfigHome, "lethean-desktop")
	cleanPath := filepath.Clean(configPath)
	if !strings.HasPrefix(cleanPath, expectedBase+string(os.PathSeparator)) && cleanPath != expectedBase {
		return fmt.Errorf("invalid config path: must be within %s", expectedBase)
	}
	return nil
}

// updateCmd represents the update command
var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Check for updates to installed miners",
	Long:  `Checks for new versions of all installed miners and notifies you if an update is available.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("Checking for updates...")

		homeDir, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("could not get home directory: %w", err)
		}
		signpostPath := filepath.Join(homeDir, signpostFilename)

		if _, err := os.Stat(signpostPath); os.IsNotExist(err) {
			fmt.Println("No miners installed yet. Run 'doctor' or 'install' first.")
			return nil
		}

		configPathBytes, err := os.ReadFile(signpostPath)
		if err != nil {
			return fmt.Errorf("could not read signpost file: %w", err)
		}
		configPath := strings.TrimSpace(string(configPathBytes))

		// Security: Validate that the config path is within the expected directory
		if err := validateUpdateConfigPath(configPath); err != nil {
			return fmt.Errorf("security error: %w", err)
		}

		cacheBytes, err := os.ReadFile(configPath)
		if err != nil {
			return fmt.Errorf("could not read cache file from %s: %w", configPath, err)
		}

		// Fix: Use SystemInfo type (matches what doctor.go saves)
		var systemInfo mining.SystemInfo
		if err := json.Unmarshal(cacheBytes, &systemInfo); err != nil {
			return fmt.Errorf("could not parse cache file: %w", err)
		}

		updatesFound := false
		for _, details := range systemInfo.InstalledMinersInfo {
			if !details.IsInstalled {
				continue
			}

			var miner mining.Miner
			var minerName string
			if filepath.Base(details.Path) == "xmrig" {
				minerName = "xmrig"
				miner = mining.NewXMRigMiner()
			} else {
				continue // Skip unknown miners
			}

			fmt.Printf("Checking %s... ", minerName)
			latestVersionStr, err := miner.GetLatestVersion()
			if err != nil {
				fmt.Printf("Error getting latest version: %v\n", err)
				continue
			}

			latestVersion, err := semver.NewVersion(latestVersionStr)
			if err != nil {
				fmt.Printf("Error parsing latest version '%s': %v\n", latestVersionStr, err)
				continue
			}

			installedVersion, err := semver.NewVersion(details.Version)
			if err != nil {
				fmt.Printf("Error parsing installed version '%s': %v\n", details.Version, err)
				continue
			}

			if latestVersion.GreaterThan(installedVersion) {
				fmt.Printf("Update available! %s -> %s\n", installedVersion, latestVersion)
				fmt.Printf("  To update, run: install %s\n", minerName)
				updatesFound = true
			} else {
				fmt.Println("You are on the latest version.")
			}
		}

		if !updatesFound {
			fmt.Println("\nAll installed miners are up to date.")
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(updateCmd)
}
