package cmd

import (
	"fmt"
	"runtime"
	"time"

	"github.com/Masterminds/semver/v3"
	"github.com/Snider/Mining/pkg/mining"
	"github.com/spf13/cobra"
)

// installCmd represents the install command
var installCmd = &cobra.Command{
	Use:   "install [miner_type]",
	Short: "Install or update a miner",
	Long:  `Download and install a new miner, or update an existing one to the latest version.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		minerType := args[0]

		var miner mining.Miner
		switch minerType {
		case "xmrig":
			miner = mining.NewXMRigMiner()
		default:
			return fmt.Errorf("unknown miner type: %s", minerType)
		}

		// Check if it's already installed and up-to-date
		details, err := miner.CheckInstallation()
		if err == nil && details.IsInstalled {
			latestVersionStr, err := miner.GetLatestVersion()
			if err == nil {
				latestVersion, err := semver.NewVersion(latestVersionStr)
				if err == nil {
					installedVersion, err := semver.NewVersion(details.Version)
					if err == nil && !latestVersion.GreaterThan(installedVersion) {
						fmt.Printf("%s is already installed and up to date (version %s).\n", miner.GetName(), installedVersion)
						return nil
					}
					fmt.Printf("Updating %s from %s to %s...\n", miner.GetName(), installedVersion, latestVersion)
				}
			}
		} else {
			fmt.Printf("Installing %s...\n", miner.GetName())
		}

		if err := miner.Install(); err != nil {
			return fmt.Errorf("failed to install/update miner: %w", err)
		}

		// Get fresh details after installation
		finalDetails, err := miner.CheckInstallation()
		if err != nil {
			return fmt.Errorf("failed to verify installation: %w", err)
		}

		fmt.Printf("%s installed successfully to %s (version %s).\n", miner.GetName(), finalDetails.Path, finalDetails.Version)

		// Update the cache after a successful installation
		fmt.Println("Updating installation cache...")
		if err := updateDoctorCache(); err != nil {
			fmt.Printf("Warning: failed to update doctor cache: %v\n", err)
		}

		return nil
	},
}

// updateDoctorCache runs the core logic of the doctor command to refresh the cache.
func updateDoctorCache() error {
	manager := getManager()
	availableMiners := manager.ListAvailableMiners()
	if len(availableMiners) == 0 {
		return nil
	}

	var allDetails []*mining.InstallationDetails
	for _, availableMiner := range availableMiners {
		var miner mining.Miner
		switch availableMiner.Name {
		case "xmrig":
			miner = mining.NewXMRigMiner()
		default:
			continue
		}
		details, err := miner.CheckInstallation()
		if err != nil {
			continue // Ignore errors for this background update
		}
		allDetails = append(allDetails, details)
	}

	// Create the SystemInfo struct that the /info endpoint expects
	systemInfo := &mining.SystemInfo{
		Timestamp:           time.Now(),
		OS:                  runtime.GOOS,
		Architecture:        runtime.GOARCH,
		GoVersion:           runtime.Version(),
		AvailableCPUCores:   runtime.NumCPU(),
		InstalledMinersInfo: allDetails,
	}

	return saveResultsToCache(systemInfo)
}

func init() {
	rootCmd.AddCommand(installCmd)
}
