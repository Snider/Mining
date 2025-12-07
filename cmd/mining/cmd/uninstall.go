package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// uninstallCmd represents the uninstall command
var uninstallCmd = &cobra.Command{
	Use:   "uninstall [miner_type]",
	Short: "Uninstall a miner",
	Long:  `Stops the miner if it is running, removes all associated files, and updates the configuration.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		minerType := args[0]
		manager := getManager() // Assuming getManager() provides the singleton manager instance

		fmt.Printf("Uninstalling %s...\n", minerType)
		if err := manager.UninstallMiner(minerType); err != nil {
			return fmt.Errorf("failed to uninstall miner: %w", err)
		}

		fmt.Printf("%s uninstalled successfully.\n", minerType)

		// The doctor cache is implicitly updated by the manager's actions,
		// but an explicit cache update can still be beneficial.
		fmt.Println("Updating installation cache...")
		if err := updateDoctorCache(); err != nil {
			fmt.Printf("Warning: failed to update doctor cache: %v\n", err)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(uninstallCmd)
}
