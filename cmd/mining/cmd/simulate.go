package cmd

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/Snider/Mining/pkg/mining"
	"github.com/spf13/cobra"
)

var (
	simCount     int
	simPreset    string
	simHashrate  int
	simAlgorithm string
)

// simulateCmd represents the simulate command
var simulateCmd = &cobra.Command{
	Use:   "simulate",
	Short: "Start the service with simulated miners for UI testing",
	Long: `Start the mining service with simulated miners that generate realistic
hashrate data and statistics. This is useful for UI development and testing
without requiring actual mining hardware.

Examples:
  # Start with 3 medium-hashrate CPU miners
  miner-ctrl simulate --count 3 --preset cpu-medium

  # Start with custom hashrate
  miner-ctrl simulate --count 2 --hashrate 8000 --algorithm rx/0

  # Start with a mix of presets
  miner-ctrl simulate --count 1 --preset gpu-ethash

Available presets:
  cpu-low      - Low-end CPU (500 H/s, rx/0)
  cpu-medium   - Medium CPU (5 kH/s, rx/0)
  cpu-high     - High-end CPU (15 kH/s, rx/0)
  gpu-ethash   - GPU mining ETH (30 MH/s, ethash)
  gpu-kawpow   - GPU mining RVN (15 MH/s, kawpow)`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		displayHost := host
		if displayHost == "0.0.0.0" {
			var err error
			displayHost, err = getLocalIP()
			if err != nil {
				displayHost = "localhost"
			}
		}
		displayAddr := fmt.Sprintf("%s:%d", displayHost, port)
		listenAddr := fmt.Sprintf("%s:%d", host, port)

		// Create a new manager for simulation (skips autostart of real miners)
		mgr := mining.NewManagerForSimulation()

		// Create and start simulated miners
		for i := 0; i < simCount; i++ {
			config := getSimulatedConfig(i)
			simMiner := mining.NewSimulatedMiner(config)

			// Start the simulated miner
			if err := simMiner.Start(&mining.Config{}); err != nil {
				return fmt.Errorf("failed to start simulated miner %d: %w", i, err)
			}

			// Register with manager
			if err := mgr.RegisterMiner(simMiner); err != nil {
				return fmt.Errorf("failed to register simulated miner %d: %w", i, err)
			}

			fmt.Printf("Started simulated miner: %s (%s, ~%d H/s)\n",
				config.Name, config.Algorithm, config.BaseHashrate)
		}

		// Create and start the service
		service, err := mining.NewService(mgr, listenAddr, displayAddr, namespace)
		if err != nil {
			return fmt.Errorf("failed to create new service: %w", err)
		}

		// Start the server in a goroutine
		go func() {
			if err := service.ServiceStartup(ctx); err != nil {
				fmt.Fprintf(os.Stderr, "Failed to start service: %v\n", err)
				cancel()
			}
		}()

		fmt.Printf("\n=== SIMULATION MODE ===\n")
		fmt.Printf("Mining service started on http://%s:%d\n", displayHost, port)
		fmt.Printf("Swagger documentation is available at http://%s:%d%s/swagger/index.html\n", displayHost, port, namespace)
		fmt.Printf("\nSimulating %d miner(s). Press Ctrl+C to stop.\n", simCount)
		fmt.Printf("Note: All data is simulated - no actual mining is occurring.\n\n")

		// Handle graceful shutdown on Ctrl+C
		signalChan := make(chan os.Signal, 1)
		signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

		select {
		case <-signalChan:
			fmt.Println("\nReceived shutdown signal, stopping simulation...")
			cancel()
		case <-ctx.Done():
		}

		// Stop all simulated miners
		for _, miner := range mgr.ListMiners() {
			mgr.StopMiner(miner.GetName())
		}

		fmt.Println("Simulation stopped.")
		return nil
	},
}

// getSimulatedConfig returns configuration for a simulated miner based on flags.
func getSimulatedConfig(index int) mining.SimulatedMinerConfig {
	// Generate unique name
	name := fmt.Sprintf("sim-%s-%03d", simPreset, index+1)

	// Start with preset if specified
	var config mining.SimulatedMinerConfig
	if preset, ok := mining.SimulatedMinerPresets[simPreset]; ok {
		config = preset
	} else {
		// Default preset
		config = mining.SimulatedMinerPresets["cpu-medium"]
	}

	config.Name = name

	// Override with custom values if provided
	if simHashrate > 0 {
		config.BaseHashrate = simHashrate
	}
	if simAlgorithm != "" {
		config.Algorithm = simAlgorithm
	}

	// Add some variance between miners
	variance := 0.1 + rand.Float64()*0.1 // 10-20% variance
	config.BaseHashrate = int(float64(config.BaseHashrate) * (0.9 + rand.Float64()*0.2))
	config.Variance = variance

	return config
}

func init() {
	// Seed random for varied simulation
	rand.Seed(time.Now().UnixNano())

	simulateCmd.Flags().IntVarP(&simCount, "count", "c", 1, "Number of simulated miners to create")
	simulateCmd.Flags().StringVar(&simPreset, "preset", "cpu-medium", "Miner preset (cpu-low, cpu-medium, cpu-high, gpu-ethash, gpu-kawpow)")
	simulateCmd.Flags().IntVar(&simHashrate, "hashrate", 0, "Custom base hashrate (overrides preset)")
	simulateCmd.Flags().StringVar(&simAlgorithm, "algorithm", "", "Custom algorithm (overrides preset)")

	// Reuse serve command flags
	simulateCmd.Flags().StringVar(&host, "host", "127.0.0.1", "Host to listen on")
	simulateCmd.Flags().IntVarP(&port, "port", "p", 9090, "Port to listen on")
	simulateCmd.Flags().StringVarP(&namespace, "namespace", "n", "/api/v1/mining", "API namespace")

	rootCmd.AddCommand(simulateCmd)
}

// Helper function to format hashrate
func formatHashrate(h int) string {
	if h >= 1000000000 {
		return strconv.FormatFloat(float64(h)/1000000000, 'f', 2, 64) + " GH/s"
	}
	if h >= 1000000 {
		return strconv.FormatFloat(float64(h)/1000000, 'f', 2, 64) + " MH/s"
	}
	if h >= 1000 {
		return strconv.FormatFloat(float64(h)/1000, 'f', 2, 64) + " kH/s"
	}
	return strconv.Itoa(h) + " H/s"
}
