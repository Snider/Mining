package cmd

import (
	"fmt"

	"github.com/Snider/Mining/pkg/mining"
	"github.com/spf13/cobra"
)

var (
	minerName      string
	minerAlgorithm string
	minerPool      string
	minerWallet    string
)

// startCmd represents the start command
var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start a new miner",
	Long:  `Start a new miner with the specified configuration.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		config := mining.MinerConfig{
			Name:      minerName,
			Algorithm: minerAlgorithm,
			Pool:      minerPool,
			Wallet:    minerWallet,
		}

		miner, err := getManager().StartMiner(config)
		if err != nil {
			return fmt.Errorf("failed to start miner: %w", err)
		}

		fmt.Printf("Miner started successfully:\n")
		fmt.Printf("  ID:     %s\n", miner.ID)
		fmt.Printf("  Name:   %s\n", miner.Name)
		fmt.Printf("  Status: %s\n", miner.Status)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(startCmd)
	startCmd.Flags().StringVarP(&minerName, "name", "n", "", "Miner name (required)")
	startCmd.Flags().StringVarP(&minerAlgorithm, "algorithm", "a", "sha256", "Mining algorithm")
	startCmd.Flags().StringVarP(&minerPool, "pool", "p", "", "Mining pool address")
	startCmd.Flags().StringVarP(&minerWallet, "wallet", "w", "", "Wallet address")
	startCmd.MarkFlagRequired("name")
}
