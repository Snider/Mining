package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// stopCmd represents the stop command
var stopCmd = &cobra.Command{
	Use:   "stop [miner-id]",
	Short: "Stop a running miner",
	Long:  `Stop a running miner by its ID.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		minerID := args[0]

		err := getManager().StopMiner(minerID)
		if err != nil {
			return fmt.Errorf("failed to stop miner: %w", err)
		}

		fmt.Printf("Miner %s stopped successfully\n", minerID)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(stopCmd)
}
