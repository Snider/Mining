package cmd

import (
	"github.com/Snider/Mining/pkg/mining"
	"github.com/spf13/cobra"
)

var (
	manager *mining.Manager
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "mining",
	Short: "Mining CLI - Manage miners with RESTful control",
	Long: `Mining is a CLI tool for managing cryptocurrency miners.
It provides commands to start, stop, list, and manage miners with RESTful control capabilities.`,
	Version: mining.GetVersion(),
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(initManager)
}

// initManager initializes the miner manager
func initManager() {
	if manager == nil {
		manager = mining.NewManager()
	}
}

// getManager returns the singleton manager instance
func getManager() *mining.Manager {
	if manager == nil {
		manager = mining.NewManager()
	}
	return manager
}
