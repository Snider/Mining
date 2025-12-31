package cmd

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/Snider/Mining/pkg/mining"
	"github.com/spf13/cobra"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

var (
	host      string
	port      int
	namespace string
)

// serveCmd represents the serve command
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the mining service and interactive shell",
	Long:  `Start the mining service, which provides a RESTful API for managing miners, and an interactive shell for CLI commands.`,
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

		// Use the global manager instance
		mgr := getManager() // This ensures we get the manager initialized by initManager

		service, err := mining.NewService(mgr, listenAddr, displayAddr, namespace) // Pass the global manager
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

		// Handle graceful shutdown on Ctrl+C
		signalChan := make(chan os.Signal, 1)
		signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

		// Start interactive shell in a goroutine
		go func() {
			fmt.Printf("Mining service started on http://%s:%d\n", displayHost, port)
			fmt.Printf("Swagger documentation is available at http://%s:%d%s/index.html\n", displayHost, port, service.SwaggerUIPath)
			fmt.Println("Entering interactive shell. Type 'exit' or 'quit' to stop.")
			fmt.Print(">> ")

			scanner := bufio.NewScanner(os.Stdin)
			for scanner.Scan() {
				line := scanner.Text()
				if line == "" {
					fmt.Print(">> ")
					continue
				}
				if strings.ToLower(line) == "exit" || strings.ToLower(line) == "quit" {
					fmt.Println("Exiting...")
					cancel()
					return
				}

				parts := strings.Fields(line)
				if len(parts) == 0 {
					fmt.Print(">> ")
					continue
				}

				command := strings.ToLower(parts[0])
				cmdArgs := parts[1:]

				switch command {
				case "start":
					if len(cmdArgs) < 3 {
						fmt.Println("Usage: start <miner_type> <pool> <wallet>")
						fmt.Println("Example: start xmrig stratum+tcp://pool.example.com:3333 YOUR_WALLET_ADDRESS")
					} else {
						minerType := cmdArgs[0]
						pool := cmdArgs[1]
						wallet := cmdArgs[2]

						// Validate pool URL format
						if !strings.HasPrefix(pool, "stratum+tcp://") &&
							!strings.HasPrefix(pool, "stratum+ssl://") &&
							!strings.HasPrefix(pool, "stratum://") {
							fmt.Fprintf(os.Stderr, "Error: Invalid pool URL (must start with stratum+tcp://, stratum+ssl://, or stratum://)\n")
							fmt.Print(">> ")
							continue
						}
						if len(pool) > 256 {
							fmt.Fprintf(os.Stderr, "Error: Pool URL too long (max 256 chars)\n")
							fmt.Print(">> ")
							continue
						}

						// Validate wallet address length
						if len(wallet) > 256 {
							fmt.Fprintf(os.Stderr, "Error: Wallet address too long (max 256 chars)\n")
							fmt.Print(">> ")
							continue
						}

						config := &mining.Config{
							Pool:      pool,
							Wallet:    wallet,
							LogOutput: true,
						}

						// Validate config before starting
						if err := config.Validate(); err != nil {
							fmt.Fprintf(os.Stderr, "Error: Invalid configuration: %v\n", err)
							fmt.Print(">> ")
							continue
						}

						miner, err := mgr.StartMiner(context.Background(), minerType, config)
						if err != nil {
							fmt.Fprintf(os.Stderr, "Error starting miner: %v\n", err)
						} else {
							fmt.Printf("Miner %s started successfully.\n", miner.GetName())
						}
					}
				case "status":
					if len(cmdArgs) < 1 {
						fmt.Println("Error: status command requires miner name (e.g., 'status xmrig')")
					} else {
						minerName := cmdArgs[0]
						miner, err := mgr.GetMiner(minerName)
						if err != nil {
							fmt.Fprintf(os.Stderr, "Error getting miner status: %v\n", err)
						} else {
							stats, err := miner.GetStats(context.Background())
							if err != nil {
								fmt.Fprintf(os.Stderr, "Error getting miner stats: %v\n", err)
							} else {
								fmt.Printf("Miner Status for %s:\n", cases.Title(language.English).String(minerName))
								fmt.Printf("  Hash Rate:  %d H/s\n", stats.Hashrate)
								fmt.Printf("  Shares:     %d\n", stats.Shares)
								fmt.Printf("  Rejected:   %d\n", stats.Rejected)
								fmt.Printf("  Uptime:     %d seconds\n", stats.Uptime)
								fmt.Printf("  Algorithm:  %s\n", stats.Algorithm)
							}
						}
					}
				case "stop":
					if len(cmdArgs) < 1 {
						fmt.Println("Error: stop command requires miner name (e.g., 'stop xmrig')")
					} else {
						minerName := cmdArgs[0]
						err := mgr.StopMiner(context.Background(), minerName)
						if err != nil {
							fmt.Fprintf(os.Stderr, "Error stopping miner: %v\n", err)
						} else {
							fmt.Printf("Miner %s stopped successfully.\n", minerName)
						}
					}
				case "list":
					miners := mgr.ListMiners()
					if len(miners) == 0 {
						fmt.Println("No miners currently running.")
					} else {
						fmt.Println("Running Miners:")
						for _, miner := range miners {
							fmt.Printf("  - %s\n", miner.GetName())
						}
					}
				default:
					fmt.Fprintf(os.Stderr, "Unknown command: %s. Only 'start', 'status', 'stop', 'list' are directly supported in this shell.\n", command)
					fmt.Fprintf(os.Stderr, "For other commands, please run them directly from your terminal (e.g., 'miner-ctrl doctor').\n")
				}
				fmt.Print(">> ")
			}

			// Check for scanner errors (I/O issues)
			if err := scanner.Err(); err != nil {
				fmt.Fprintf(os.Stderr, "Error reading input: %v\n", err)
			}
		}()

		select {
		case <-signalChan:
			fmt.Println("\nReceived shutdown signal, stopping service...")
			cancel()
		case <-ctx.Done():
		}

		// Explicit cleanup of manager resources
		mgr.Stop()

		fmt.Println("Mining service stopped.")
		return nil
	},
}

func init() {
	serveCmd.Flags().StringVar(&host, "host", "127.0.0.1", "Host to listen on")
	serveCmd.Flags().IntVarP(&port, "port", "p", 9090, "Port to listen on")
	serveCmd.Flags().StringVarP(&namespace, "namespace", "n", "/api/v1/mining", "API namespace for the swagger UI")
	rootCmd.AddCommand(serveCmd)
}

func getLocalIP() (string, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "", err
	}
	for _, address := range addrs {
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String(), nil
			}
		}
	}
	return "localhost", nil
}
