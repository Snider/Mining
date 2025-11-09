package cmd

import (
	"fmt"

	"github.com/Snider/Mining/pkg/mining"
	"github.com/spf13/cobra"
)

// uninstallCmd represents the uninstall command
var uninstallCmd = &cobra.Command{
	Use:   "uninstall [miner_type]",
	Short: "Uninstall a miner",
	Long:  `Remove all files associated with a specific miner.`,
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

		fmt.Printf("Uninstalling %s...\n", miner.GetName())
		if err := miner.Uninstall(); err != nil {
			return fmt.Errorf("failed to uninstall miner: %w", err)
		}

		fmt.Printf("%s uninstalled successfully.\n", miner.GetName())

		// Update the cache after a successful uninstallation
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
