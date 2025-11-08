package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// statusCmd represents the status command
var statusCmd = &cobra.Command{
	Use:   "status [miner-id]",
	Short: "Get status of a miner",
	Long:  `Get detailed status information for a specific miner.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		minerID := args[0]

		miner, err := getManager().GetMiner(minerID)
		if err != nil {
			return fmt.Errorf("failed to get miner: %w", err)
		}

		fmt.Printf("Miner Status:\n")
		fmt.Printf("  ID:         %s\n", miner.ID)
		fmt.Printf("  Name:       %s\n", miner.Name)
		fmt.Printf("  Status:     %s\n", miner.Status)
		fmt.Printf("  Start Time: %s\n", miner.StartTime.Format("2006-01-02 15:04:05"))
		fmt.Printf("  Hash Rate:  %.2f H/s\n", miner.HashRate)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(statusCmd)
}
