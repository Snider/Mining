package cmd

import (
	"fmt"

	"github.com/Snider/Mining/pkg/mining"
	"github.com/spf13/cobra"
)

var (
	minerPool   string
	minerWallet string
)

// startCmd represents the start command
var startCmd = &cobra.Command{
	Use:   "start [miner_name]",
	Short: "Start a new miner",
	Long:  `Start a new miner with the specified configuration.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		minerType := args[0]
		config := &mining.Config{
			Pool:   minerPool,
			Wallet: minerWallet,
		}

		miner, err := getManager().StartMiner(minerType, config)
		if err != nil {
			return fmt.Errorf("failed to start miner: %w", err)
		}

		fmt.Printf("Miner started successfully:\n")
		fmt.Printf("  Name:   %s\n", miner.GetName())
		return nil
	},
}

func init() {
	rootCmd.AddCommand(startCmd)
	startCmd.Flags().StringVarP(&minerPool, "pool", "p", "pool.hashvault.pro", "Mining pool address")
	startCmd.Flags().StringVarP(&minerWallet, "wallet", "w", "888tNkZrPN6JsEgekjMnABU4TBzc2Dt29EPAvkRxbANsAnjyPbb3iQ1YBRk1UXcdRsiKc9dhwMVgN5S9cQUiyoogDavup3H", "Wallet address")
	// Removed MarkFlagRequired as we now have default values
}
