package cmd

import (
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/Snider/Mining/pkg/node"
	"github.com/spf13/cobra"
)

var (
	nodeManager     *node.NodeManager
	nodeManagerOnce sync.Once
	nodeManagerErr  error

	peerRegistry     *node.PeerRegistry
	peerRegistryOnce sync.Once
	peerRegistryErr  error
)

// nodeCmd represents the node parent command
var nodeCmd = &cobra.Command{
	Use:   "node",
	Short: "Manage P2P node identity and connections",
	Long:  `Manage the node's identity, view status, and control P2P networking.`,
}

// nodeInitCmd initializes a new node identity
var nodeInitCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize node identity",
	Long: `Initialize a new node identity with X25519 keypair.
This creates the node's cryptographic identity for secure P2P communication.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		name, _ := cmd.Flags().GetString("name")
		role, _ := cmd.Flags().GetString("role")

		if name == "" {
			return fmt.Errorf("--name is required")
		}

		nm, err := node.NewNodeManager()
		if err != nil {
			return fmt.Errorf("failed to create node manager: %w", err)
		}

		if nm.HasIdentity() {
			return fmt.Errorf("node identity already exists. Use 'node reset' to create a new one")
		}

		var nodeRole node.NodeRole
		switch role {
		case "controller":
			nodeRole = node.RoleController
		case "worker":
			nodeRole = node.RoleWorker
		case "dual", "":
			nodeRole = node.RoleDual
		default:
			return fmt.Errorf("invalid role: %s (use controller, worker, or dual)", role)
		}

		if err := nm.GenerateIdentity(name, nodeRole); err != nil {
			return fmt.Errorf("failed to generate identity: %w", err)
		}

		identity := nm.GetIdentity()
		fmt.Println("Node identity created successfully!")
		fmt.Println()
		fmt.Printf("  ID:         %s\n", identity.ID)
		fmt.Printf("  Name:       %s\n", identity.Name)
		fmt.Printf("  Role:       %s\n", identity.Role)
		fmt.Printf("  Public Key: %s\n", identity.PublicKey)
		fmt.Printf("  Created:    %s\n", identity.CreatedAt.Format(time.RFC3339))

		return nil
	},
}

// nodeInfoCmd shows current node identity
var nodeInfoCmd = &cobra.Command{
	Use:   "info",
	Short: "Show node identity and status",
	Long:  `Display the current node's identity, role, and connection status.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		nm, err := node.NewNodeManager()
		if err != nil {
			return fmt.Errorf("failed to create node manager: %w", err)
		}

		if !nm.HasIdentity() {
			fmt.Println("No node identity found.")
			fmt.Println("Run 'node init --name <name>' to create one.")
			return nil
		}

		identity := nm.GetIdentity()
		fmt.Println("Node Identity:")
		fmt.Println()
		fmt.Printf("  ID:         %s\n", identity.ID)
		fmt.Printf("  Name:       %s\n", identity.Name)
		fmt.Printf("  Role:       %s\n", identity.Role)
		fmt.Printf("  Public Key: %s\n", identity.PublicKey)
		fmt.Printf("  Created:    %s\n", identity.CreatedAt.Format(time.RFC3339))

		// Show peer info if available
		pr, err := node.NewPeerRegistry()
		if err == nil {
			fmt.Println()
			fmt.Printf("  Registered Peers: %d\n", pr.Count())
			connected := pr.GetConnectedPeers()
			fmt.Printf("  Connected Peers:  %d\n", len(connected))
		}

		return nil
	},
}

// nodeServeCmd starts the P2P server
var nodeServeCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start P2P server for remote connections",
	Long: `Start the P2P WebSocket server to accept connections from other nodes.
This allows other nodes to connect, send commands, and receive stats.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		listen, _ := cmd.Flags().GetString("listen")

		nm, err := node.NewNodeManager()
		if err != nil {
			return fmt.Errorf("failed to create node manager: %w", err)
		}

		if !nm.HasIdentity() {
			return fmt.Errorf("no node identity found. Run 'node init --name <name>' first")
		}

		pr, err := node.NewPeerRegistry()
		if err != nil {
			return fmt.Errorf("failed to create peer registry: %w", err)
		}

		config := node.DefaultTransportConfig()
		if listen != "" {
			config.ListenAddr = listen
		}

		transport := node.NewTransport(nm, pr, config)

		// Create worker to handle incoming messages
		worker := node.NewWorker(nm, transport)
		worker.RegisterWithTransport()

		if err := transport.Start(); err != nil {
			return fmt.Errorf("failed to start transport: %w", err)
		}

		identity := nm.GetIdentity()
		fmt.Printf("P2P server started on %s\n", config.ListenAddr)
		fmt.Printf("Node ID: %s (%s)\n", identity.ID, identity.Name)
		fmt.Printf("Role: %s\n", identity.Role)
		fmt.Println()
		fmt.Println("Press Ctrl+C to stop...")

		// Set up signal handling for graceful shutdown (including SIGHUP for terminal disconnect)
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM, syscall.SIGHUP)

		// Wait for shutdown signal
		sig := <-sigChan
		fmt.Printf("\nReceived signal %v, shutting down...\n", sig)

		// Graceful shutdown: stop transport and cleanup resources
		if err := transport.Stop(); err != nil {
			fmt.Printf("Warning: error during transport shutdown: %v\n", err)
			// Force cleanup on Stop() failure
			fmt.Println("Forcing resource cleanup...")
			for _, peer := range pr.GetConnectedPeers() {
				pr.SetConnected(peer.ID, false)
			}
		}

		// Ensure peer registry is flushed to disk
		if err := pr.Close(); err != nil {
			fmt.Printf("Warning: error closing peer registry: %v\n", err)
		}

		fmt.Println("P2P server stopped.")
		return nil
	},
}

// nodeResetCmd deletes the node identity
var nodeResetCmd = &cobra.Command{
	Use:   "reset",
	Short: "Delete node identity and start fresh",
	Long:  `Remove the current node identity, keys, and all peer data. Use with caution!`,
	RunE: func(cmd *cobra.Command, args []string) error {
		force, _ := cmd.Flags().GetBool("force")

		nm, err := node.NewNodeManager()
		if err != nil {
			return fmt.Errorf("failed to create node manager: %w", err)
		}

		if !nm.HasIdentity() {
			fmt.Println("No node identity to reset.")
			return nil
		}

		if !force {
			fmt.Println("This will permanently delete your node identity and keys.")
			fmt.Println("All peers will need to re-register with your new identity.")
			fmt.Println()
			fmt.Println("Run with --force to confirm.")
			return nil
		}

		if err := nm.Delete(); err != nil {
			return fmt.Errorf("failed to delete identity: %w", err)
		}

		fmt.Println("Node identity deleted successfully.")
		fmt.Println("Run 'node init --name <name>' to create a new identity.")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(nodeCmd)

	// node init
	nodeCmd.AddCommand(nodeInitCmd)
	nodeInitCmd.Flags().StringP("name", "n", "", "Node name (required)")
	nodeInitCmd.Flags().StringP("role", "r", "dual", "Node role: controller, worker, or dual (default)")

	// node info
	nodeCmd.AddCommand(nodeInfoCmd)

	// node serve
	nodeCmd.AddCommand(nodeServeCmd)
	nodeServeCmd.Flags().StringP("listen", "l", ":9091", "Address to listen on")

	// node reset
	nodeCmd.AddCommand(nodeResetCmd)
	nodeResetCmd.Flags().BoolP("force", "f", false, "Force reset without confirmation")
}

// getNodeManager returns the singleton node manager (thread-safe)
func getNodeManager() (*node.NodeManager, error) {
	nodeManagerOnce.Do(func() {
		nodeManager, nodeManagerErr = node.NewNodeManager()
	})
	return nodeManager, nodeManagerErr
}

// getPeerRegistry returns the singleton peer registry (thread-safe)
func getPeerRegistry() (*node.PeerRegistry, error) {
	peerRegistryOnce.Do(func() {
		peerRegistry, peerRegistryErr = node.NewPeerRegistry()
	})
	return peerRegistry, peerRegistryErr
}
