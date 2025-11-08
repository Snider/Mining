package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all miners",
	Long:  `List all miners and their current status.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		miners := getManager().ListMiners()

		if len(miners) == 0 {
			fmt.Println("No miners found")
			return nil
		}

		fmt.Printf("%-20s %-20s %-10s %-12s\n", "ID", "Name", "Status", "Hash Rate")
		fmt.Println("--------------------------------------------------------------------------------")
		for _, miner := range miners {
			fmt.Printf("%-20s %-20s %-10s %-12.2f\n",
				miner.ID,
				miner.Name,
				miner.Status,
				miner.HashRate,
			)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}
