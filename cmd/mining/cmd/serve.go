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

		manager := mining.NewManager()
		service := mining.NewService(manager, listenAddr, displayAddr, namespace)

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

				rootCmd.SetArgs(strings.Fields(line))
				if err := rootCmd.Execute(); err != nil {
					fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				}
				rootCmd.SetArgs([]string{})
				fmt.Print(">> ")
			}
		}()

		select {
		case <-signalChan:
			fmt.Println("\nReceived shutdown signal, stopping service...")
			cancel()
		case <-ctx.Done():
		}

		fmt.Println("Mining service stopped.")
		return nil
	},
}

func init() {
	serveCmd.Flags().StringVar(&host, "host", "0.0.0.0", "Host to listen on")
	serveCmd.Flags().IntVarP(&port, "port", "p", 8080, "Port to listen on")
	serveCmd.Flags().StringVarP(&namespace, "namespace", "n", "/", "API namespace for the swagger UI")
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
