package cmd

import (
	"fmt"
	"time"

	"github.com/Snider/Mining/pkg/node"
	"github.com/spf13/cobra"
)

// Note: findPeerByPartialID is defined in remote.go and used for peer lookup

// peerCmd represents the peer parent command
var peerCmd = &cobra.Command{
	Use:   "peer",
	Short: "Manage peer nodes",
	Long:  `Add, remove, and manage connections to peer nodes.`,
}

// peerAddCmd adds a new peer
var peerAddCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a peer node",
	Long: `Add a new peer node by address. This will initiate a handshake
to exchange public keys and establish a secure connection.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		address, _ := cmd.Flags().GetString("address")
		name, _ := cmd.Flags().GetString("name")

		if address == "" {
			return fmt.Errorf("--address is required")
		}

		nm, err := getNodeManager()
		if err != nil {
			return fmt.Errorf("failed to get node manager: %w", err)
		}

		if !nm.HasIdentity() {
			return fmt.Errorf("no node identity found. Run 'node init' first")
		}

		pr, err := getPeerRegistry()
		if err != nil {
			return fmt.Errorf("failed to get peer registry: %w", err)
		}

		// For now, just add to registry - actual connection happens with 'node serve'
		// In a full implementation, we'd connect here and get the peer's identity
		peer := &node.Peer{
			ID:      fmt.Sprintf("pending-%d", time.Now().UnixNano()),
			Name:    name,
			Address: address,
			Role:    node.RoleDual,
			AddedAt: time.Now(),
			Score:   50,
		}

		if err := pr.AddPeer(peer); err != nil {
			return fmt.Errorf("failed to add peer: %w", err)
		}

		fmt.Printf("Peer added: %s at %s\n", name, address)
		fmt.Println("Connect using 'node serve' to complete handshake.")
		return nil
	},
}

// peerListCmd lists all registered peers
var peerListCmd = &cobra.Command{
	Use:   "list",
	Short: "List registered peers",
	Long:  `Display all registered peers with their connection status.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		pr, err := getPeerRegistry()
		if err != nil {
			return fmt.Errorf("failed to get peer registry: %w", err)
		}

		peers := pr.ListPeers()
		if len(peers) == 0 {
			fmt.Println("No peers registered.")
			fmt.Println("Use 'peer add --address <host:port> --name <name>' to add one.")
			return nil
		}

		fmt.Printf("Registered Peers (%d):\n\n", len(peers))
		for _, peer := range peers {
			status := "offline"
			if peer.Connected {
				status = "online"
			}

			fmt.Printf("  %s (%s)\n", peer.Name, peer.ID[:16])
			fmt.Printf("    Address:  %s\n", peer.Address)
			fmt.Printf("    Role:     %s\n", peer.Role)
			fmt.Printf("    Status:   %s\n", status)
			fmt.Printf("    Ping:     %.1f ms\n", peer.PingMS)
			fmt.Printf("    Score:    %.1f\n", peer.Score)
			if !peer.LastSeen.IsZero() {
				fmt.Printf("    Last Seen: %s\n", peer.LastSeen.Format(time.RFC3339))
			}
			fmt.Println()
		}

		return nil
	},
}

// peerRemoveCmd removes a peer
var peerRemoveCmd = &cobra.Command{
	Use:   "remove <peer-id>",
	Short: "Remove a peer from registry",
	Long:  `Remove a peer node from the registry. This will disconnect if connected.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		peerID := args[0]

		peer := findPeerByPartialID(peerID)
		if peer == nil {
			return fmt.Errorf("peer not found: %s", peerID)
		}

		pr, err := getPeerRegistry()
		if err != nil {
			return fmt.Errorf("failed to get peer registry: %w", err)
		}

		if err := pr.RemovePeer(peer.ID); err != nil {
			return fmt.Errorf("failed to remove peer: %w", err)
		}

		fmt.Printf("Peer removed: %s (%s)\n", peer.Name, peer.ID[:16])
		return nil
	},
}

// peerPingCmd pings a peer
var peerPingCmd = &cobra.Command{
	Use:   "ping <peer-id>",
	Short: "Ping a peer and update metrics",
	Long:  `Send a ping to a peer and measure round-trip latency.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		peerID := args[0]

		peer := findPeerByPartialID(peerID)
		if peer == nil {
			return fmt.Errorf("peer not found: %s", peerID)
		}

		if !peer.Connected {
			return fmt.Errorf("peer not connected: %s", peer.Name)
		}

		fmt.Printf("Pinging %s (%s)...\n", peer.Name, peer.Address)
		// TODO: Actually send ping via transport
		fmt.Println("Ping functionality requires active connection via 'node serve'")
		return nil
	},
}

// peerOptimalCmd shows the optimal peer based on metrics
var peerOptimalCmd = &cobra.Command{
	Use:   "optimal",
	Short: "Show the optimal peer based on metrics",
	Long: `Use the Poindexter KD-tree to find the best peer based on
ping latency, hop count, geographic distance, and reliability score.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		count, _ := cmd.Flags().GetInt("count")

		pr, err := getPeerRegistry()
		if err != nil {
			return fmt.Errorf("failed to get peer registry: %w", err)
		}

		if pr.Count() == 0 {
			fmt.Println("No peers registered.")
			return nil
		}

		if count == 1 {
			peer := pr.SelectOptimalPeer()
			if peer == nil {
				fmt.Println("No optimal peer found.")
				return nil
			}

			fmt.Println("Optimal Peer:")
			fmt.Printf("  %s (%s)\n", peer.Name, peer.ID[:16])
			fmt.Printf("  Address: %s\n", peer.Address)
			fmt.Printf("  Ping:    %.1f ms\n", peer.PingMS)
			fmt.Printf("  Hops:    %d\n", peer.Hops)
			fmt.Printf("  Geo:     %.1f km\n", peer.GeoKM)
			fmt.Printf("  Score:   %.1f\n", peer.Score)
		} else {
			peers := pr.SelectNearestPeers(count)
			if len(peers) == 0 {
				fmt.Println("No peers found.")
				return nil
			}

			fmt.Printf("Top %d Peers (by multi-factor optimization):\n\n", len(peers))
			for i, peer := range peers {
				fmt.Printf("  %d. %s (%s)\n", i+1, peer.Name, peer.ID[:16])
				fmt.Printf("     Ping: %.1f ms | Hops: %d | Geo: %.1f km | Score: %.1f\n",
					peer.PingMS, peer.Hops, peer.GeoKM, peer.Score)
			}
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(peerCmd)

	// peer add
	peerCmd.AddCommand(peerAddCmd)
	peerAddCmd.Flags().StringP("address", "a", "", "Peer address (host:port)")
	peerAddCmd.Flags().StringP("name", "n", "", "Peer name")

	// peer list
	peerCmd.AddCommand(peerListCmd)

	// peer remove
	peerCmd.AddCommand(peerRemoveCmd)

	// peer ping
	peerCmd.AddCommand(peerPingCmd)

	// peer optimal
	peerCmd.AddCommand(peerOptimalCmd)
	peerOptimalCmd.Flags().IntP("count", "c", 1, "Number of optimal peers to show")
}
