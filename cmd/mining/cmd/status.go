package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// statusCmd represents the status command
var statusCmd = &cobra.Command{
	Use:   "status [miner_name]",
	Short: "Get status of a running miner",
	Long:  `Get detailed status information for a specific running miner.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		minerName := args[0]
		mgr := getManager()

		miner, err := mgr.GetMiner(minerName)
		if err != nil {
			return fmt.Errorf("failed to get miner: %w", err)
		}

		stats, err := miner.GetStats()
		if err != nil {
			return fmt.Errorf("failed to get miner stats: %w", err)
		}

		fmt.Printf("Miner Status for %s:\n", cases.Title(language.English).String(minerName))
		fmt.Printf("  Hash Rate:  %d H/s\n", stats.Hashrate)
		fmt.Printf("  Shares:     %d\n", stats.Shares)
		fmt.Printf("  Rejected:   %d\n", stats.Rejected)
		fmt.Printf("  Uptime:     %d seconds\n", stats.Uptime)
		fmt.Printf("  Algorithm:  %s\n", stats.Algorithm)

		return nil
	},
}

func init() {
	rootCmd.AddCommand(statusCmd)
}
