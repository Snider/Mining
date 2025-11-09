package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// stopCmd represents the stop command
var stopCmd = &cobra.Command{
	Use:   "stop [miner_name]",
	Short: "Stop a running miner",
	Long:  `Stop a running miner by its name.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		minerName := args[0]
		mgr := getManager()

		if err := mgr.StopMiner(minerName); err != nil {
			return fmt.Errorf("failed to stop miner: %w", err)
		}

		fmt.Printf("Miner %s stopped successfully\n", minerName)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(stopCmd)
}
