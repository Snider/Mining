package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List running and available miners",
	Long:  `List all running miners and their status, as well as all miners that are available to be installed and started.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		manager := getManager()

		// List running miners
		runningMiners := manager.ListMiners()
		fmt.Println("Running Miners:")
		if len(runningMiners) == 0 {
			fmt.Println("  No running miners found.")
		} else {
			fmt.Printf("  %-20s\n", "Name")
			fmt.Println("  --------------------")
			for _, miner := range runningMiners {
				fmt.Printf("  %-20s\n", miner.GetName())
			}
		}

		fmt.Println()

		// List available miners
		availableMiners := manager.ListAvailableMiners()
		fmt.Println("Available Miners:")
		if len(availableMiners) == 0 {
			fmt.Println("  No available miners found.")
		} else {
			fmt.Printf("  %-20s %s\n", "Name", "Description")
			fmt.Println("  -----------------------------------------------------------------")
			for _, miner := range availableMiners {
				fmt.Printf("  %-20s %s\n", miner.Name, miner.Description)
			}
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}
