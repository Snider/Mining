package cmd

import (
	"fmt"
	"strings"
	"time"

	"github.com/Snider/Mining/pkg/node"
	"github.com/spf13/cobra"
)

var (
	controller *node.Controller
	transport  *node.Transport
)

// remoteCmd represents the remote parent command
var remoteCmd = &cobra.Command{
	Use:   "remote",
	Short: "Control remote mining nodes",
	Long:  `Send commands to remote worker nodes and retrieve their status.`,
}

// remoteStatusCmd shows stats from remote peers
var remoteStatusCmd = &cobra.Command{
	Use:   "status [peer-id]",
	Short: "Get mining status from remote peers",
	Long:  `Display mining statistics from all connected peers or a specific peer.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctrl, err := getController()
		if err != nil {
			return err
		}

		if len(args) > 0 {
			// Get stats from specific peer
			peerID := args[0]
			peer := findPeerByPartialID(peerID)
			if peer == nil {
				return fmt.Errorf("peer not found: %s", peerID)
			}

			stats, err := ctrl.GetRemoteStats(peer.ID)
			if err != nil {
				return fmt.Errorf("failed to get stats: %w", err)
			}

			printPeerStats(peer, stats)
		} else {
			// Get stats from all peers
			allStats := ctrl.GetAllStats()
			if len(allStats) == 0 {
				fmt.Println("No connected peers.")
				return nil
			}

			pr, _ := getPeerRegistry()
			var totalHashrate float64

			for peerID, stats := range allStats {
				peer := pr.GetPeer(peerID)
				if peer != nil {
					printPeerStats(peer, stats)
					for _, miner := range stats.Miners {
						totalHashrate += miner.Hashrate
					}
				}
			}

			fmt.Println("────────────────────────────────────")
			fmt.Printf("Total Fleet Hashrate: %.2f H/s\n", totalHashrate)
		}

		return nil
	},
}

// remoteStartCmd starts a miner on a remote peer
var remoteStartCmd = &cobra.Command{
	Use:   "start <peer-id>",
	Short: "Start miner on remote peer",
	Long:  `Start a miner on a remote peer using a profile.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		minerType, _ := cmd.Flags().GetString("type")
		if minerType == "" {
			return fmt.Errorf("--type is required (e.g., xmrig, tt-miner)")
		}
		profileID, _ := cmd.Flags().GetString("profile")

		peerID := args[0]
		peer := findPeerByPartialID(peerID)
		if peer == nil {
			return fmt.Errorf("peer not found: %s", peerID)
		}

		ctrl, err := getController()
		if err != nil {
			return err
		}

		fmt.Printf("Starting %s miner on %s with profile %s...\n", minerType, peer.Name, profileID)
		if err := ctrl.StartRemoteMiner(peer.ID, minerType, profileID, nil); err != nil {
			return fmt.Errorf("failed to start miner: %w", err)
		}

		fmt.Println("Miner started successfully.")
		return nil
	},
}

// remoteStopCmd stops a miner on a remote peer
var remoteStopCmd = &cobra.Command{
	Use:   "stop <peer-id> [miner-name]",
	Short: "Stop miner on remote peer",
	Long:  `Stop a running miner on a remote peer.`,
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		peerID := args[0]
		peer := findPeerByPartialID(peerID)
		if peer == nil {
			return fmt.Errorf("peer not found: %s", peerID)
		}

		minerName := ""
		if len(args) > 1 {
			minerName = args[1]
		} else {
			minerName, _ = cmd.Flags().GetString("miner")
		}

		if minerName == "" {
			return fmt.Errorf("miner name required (as argument or --miner flag)")
		}

		ctrl, err := getController()
		if err != nil {
			return err
		}

		fmt.Printf("Stopping miner %s on %s...\n", minerName, peer.Name)
		if err := ctrl.StopRemoteMiner(peer.ID, minerName); err != nil {
			return fmt.Errorf("failed to stop miner: %w", err)
		}

		fmt.Println("Miner stopped successfully.")
		return nil
	},
}

// remoteLogsCmd gets logs from a remote miner
var remoteLogsCmd = &cobra.Command{
	Use:   "logs <peer-id> <miner-name>",
	Short: "Get console logs from remote miner",
	Long:  `Retrieve console output logs from a miner running on a remote peer.`,
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		peerID := args[0]
		minerName := args[1]
		lines, _ := cmd.Flags().GetInt("lines")

		peer := findPeerByPartialID(peerID)
		if peer == nil {
			return fmt.Errorf("peer not found: %s", peerID)
		}

		ctrl, err := getController()
		if err != nil {
			return err
		}

		logLines, err := ctrl.GetRemoteLogs(peer.ID, minerName, lines)
		if err != nil {
			return fmt.Errorf("failed to get logs: %w", err)
		}

		fmt.Printf("Logs from %s on %s (%d lines):\n", minerName, peer.Name, len(logLines))
		fmt.Println("────────────────────────────────────")
		for _, line := range logLines {
			fmt.Println(line)
		}

		return nil
	},
}

// remoteConnectCmd connects to a peer
var remoteConnectCmd = &cobra.Command{
	Use:   "connect <peer-id>",
	Short: "Connect to a remote peer",
	Long:  `Establish a WebSocket connection to a registered peer.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		peerID := args[0]
		peer := findPeerByPartialID(peerID)
		if peer == nil {
			return fmt.Errorf("peer not found: %s", peerID)
		}

		ctrl, err := getController()
		if err != nil {
			return err
		}

		fmt.Printf("Connecting to %s at %s...\n", peer.Name, peer.Address)
		if err := ctrl.ConnectToPeer(peer.ID); err != nil {
			return fmt.Errorf("failed to connect: %w", err)
		}

		fmt.Println("Connected successfully.")
		return nil
	},
}

// remoteDisconnectCmd disconnects from a peer
var remoteDisconnectCmd = &cobra.Command{
	Use:   "disconnect <peer-id>",
	Short: "Disconnect from a remote peer",
	Long:  `Close the connection to a peer.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		peerID := args[0]
		peer := findPeerByPartialID(peerID)
		if peer == nil {
			return fmt.Errorf("peer not found: %s", peerID)
		}

		ctrl, err := getController()
		if err != nil {
			return err
		}

		fmt.Printf("Disconnecting from %s...\n", peer.Name)
		if err := ctrl.DisconnectFromPeer(peer.ID); err != nil {
			return fmt.Errorf("failed to disconnect: %w", err)
		}

		fmt.Println("Disconnected.")
		return nil
	},
}

// remotePingCmd pings a peer
var remotePingCmd = &cobra.Command{
	Use:   "ping <peer-id>",
	Short: "Ping a remote peer",
	Long:  `Send a ping to a peer and measure round-trip latency.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		count, _ := cmd.Flags().GetInt("count")

		peerID := args[0]
		peer := findPeerByPartialID(peerID)
		if peer == nil {
			return fmt.Errorf("peer not found: %s", peerID)
		}

		ctrl, err := getController()
		if err != nil {
			return err
		}

		fmt.Printf("Pinging %s (%s)...\n", peer.Name, peer.Address)

		var totalRTT float64
		var successful int

		for i := 0; i < count; i++ {
			rtt, err := ctrl.PingPeer(peer.ID)
			if err != nil {
				fmt.Printf("  Ping %d: timeout\n", i+1)
				continue
			}
			fmt.Printf("  Ping %d: %.2f ms\n", i+1, rtt)
			totalRTT += rtt
			successful++

			if i < count-1 {
				time.Sleep(time.Second)
			}
		}

		if successful > 0 {
			fmt.Printf("\nAverage: %.2f ms (%d/%d successful)\n", totalRTT/float64(successful), successful, count)
		} else {
			fmt.Println("\nAll pings failed.")
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(remoteCmd)

	// remote status
	remoteCmd.AddCommand(remoteStatusCmd)

	// remote start
	remoteCmd.AddCommand(remoteStartCmd)
	remoteStartCmd.Flags().StringP("profile", "p", "", "Profile ID to use for starting the miner")
	remoteStartCmd.Flags().StringP("type", "t", "", "Miner type (e.g., xmrig, tt-miner)")

	// remote stop
	remoteCmd.AddCommand(remoteStopCmd)
	remoteStopCmd.Flags().StringP("miner", "m", "", "Miner name to stop")

	// remote logs
	remoteCmd.AddCommand(remoteLogsCmd)
	remoteLogsCmd.Flags().IntP("lines", "n", 100, "Number of log lines to retrieve")

	// remote connect
	remoteCmd.AddCommand(remoteConnectCmd)

	// remote disconnect
	remoteCmd.AddCommand(remoteDisconnectCmd)

	// remote ping
	remoteCmd.AddCommand(remotePingCmd)
	remotePingCmd.Flags().IntP("count", "c", 4, "Number of pings to send")
}

// getController returns or creates the controller instance.
func getController() (*node.Controller, error) {
	if controller != nil {
		return controller, nil
	}

	nm, err := getNodeManager()
	if err != nil {
		return nil, fmt.Errorf("failed to get node manager: %w", err)
	}

	if !nm.HasIdentity() {
		return nil, fmt.Errorf("no node identity found. Run 'node init' first")
	}

	pr, err := getPeerRegistry()
	if err != nil {
		return nil, fmt.Errorf("failed to get peer registry: %w", err)
	}

	// Initialize transport if not done
	if transport == nil {
		config := node.DefaultTransportConfig()
		transport = node.NewTransport(nm, pr, config)
	}

	controller = node.NewController(nm, pr, transport)
	return controller, nil
}

// findPeerByPartialID finds a peer by full or partial ID.
func findPeerByPartialID(partialID string) *node.Peer {
	pr, err := getPeerRegistry()
	if err != nil {
		return nil
	}

	// Try exact match first
	peer := pr.GetPeer(partialID)
	if peer != nil {
		return peer
	}

	// Try partial match
	for _, p := range pr.ListPeers() {
		if strings.HasPrefix(p.ID, partialID) {
			return p
		}
		// Also try matching by name
		if strings.EqualFold(p.Name, partialID) {
			return p
		}
	}

	return nil
}

// printPeerStats prints formatted stats for a peer.
func printPeerStats(peer *node.Peer, stats *node.StatsPayload) {
	fmt.Printf("\n%s (%s)\n", peer.Name, peer.ID[:16])
	fmt.Printf("  Address: %s\n", peer.Address)
	fmt.Printf("  Uptime:  %s\n", formatDuration(time.Duration(stats.Uptime)*time.Second))
	fmt.Printf("  Miners:  %d\n", len(stats.Miners))

	if len(stats.Miners) > 0 {
		fmt.Println()
		for _, miner := range stats.Miners {
			fmt.Printf("    %s (%s)\n", miner.Name, miner.Type)
			fmt.Printf("      Hashrate:  %.2f H/s\n", miner.Hashrate)
			fmt.Printf("      Shares:    %d (rejected: %d)\n", miner.Shares, miner.Rejected)
			fmt.Printf("      Algorithm: %s\n", miner.Algorithm)
			fmt.Printf("      Pool:      %s\n", miner.Pool)
		}
	}
}

// formatDuration formats a duration into a human-readable string.
func formatDuration(d time.Duration) string {
	days := int(d.Hours() / 24)
	hours := int(d.Hours()) % 24
	minutes := int(d.Minutes()) % 60

	if days > 0 {
		return fmt.Sprintf("%dd %dh %dm", days, hours, minutes)
	}
	if hours > 0 {
		return fmt.Sprintf("%dh %dm", hours, minutes)
	}
	return fmt.Sprintf("%dm", minutes)
}
