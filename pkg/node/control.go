package node

import (
	"fmt"
	"net"
	"net/rpc"
	"os"
	"path/filepath"
	"time"

	"github.com/Snider/Mining/pkg/logging"
	"github.com/adrg/xdg"
)

// ControlService exposes node commands via RPC.
type ControlService struct {
	controller *Controller
}

// PingArgs represents arguments for the Ping command.
type PingArgs struct {
	PeerID string
}

// PingReply represents the response from the Ping command.
type PingReply struct {
	LatencyMS float64
}

// Ping sends a ping to the specified peer.
func (s *ControlService) Ping(args *PingArgs, reply *PingReply) error {
	if s.controller == nil {
		return fmt.Errorf("controller not initialized")
	}

	latency, err := s.controller.PingPeer(args.PeerID)
	if err != nil {
		return err
	}

	reply.LatencyMS = latency
	return nil
}

// StartControlServer starts the RPC server on a Unix socket.
// Returns the listener, which should be closed by the caller.
func StartControlServer(controller *Controller) (net.Listener, error) {
	service := &ControlService{controller: controller}
	server := rpc.NewServer()
	if err := server.Register(service); err != nil {
		return nil, fmt.Errorf("failed to register control service: %w", err)
	}

	sockPath, err := getControlSocketPath()
	if err != nil {
		return nil, err
	}

	// Remove stale socket file if it exists
	if _, err := os.Stat(sockPath); err == nil {
		// Try to connect to see if it's active
		if conn, err := net.DialTimeout("unix", sockPath, 100*time.Millisecond); err == nil {
			conn.Close()
			return nil, fmt.Errorf("control socket already in use: %s", sockPath)
		}
		// Not active, remove it
		if err := os.Remove(sockPath); err != nil {
			return nil, fmt.Errorf("failed to remove stale socket: %w", err)
		}
	}

	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(sockPath), 0755); err != nil {
		return nil, fmt.Errorf("failed to create socket directory: %w", err)
	}

	listener, err := net.Listen("unix", sockPath)
	if err != nil {
		return nil, fmt.Errorf("failed to listen on socket %s: %w", sockPath, err)
	}

	// Set permissions so only user can access
	if err := os.Chmod(sockPath, 0600); err != nil {
		listener.Close()
		return nil, fmt.Errorf("failed to set socket permissions: %w", err)
	}

	logging.Info("control server started", logging.Fields{"address": sockPath})

	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				// Check if listener was closed (ErrNetClosing is not exported, check string or type)
				// Simply returning on error is usually fine for this use case
				return
			}
			go server.ServeConn(conn)
		}
	}()

	return listener, nil
}

// NewControlClient creates a new RPC client connected to the control server.
func NewControlClient() (*rpc.Client, error) {
	sockPath, err := getControlSocketPath()
	if err != nil {
		return nil, err
	}

	// Use DialTimeout to fail fast if server is not running
	conn, err := net.DialTimeout("unix", sockPath, 2*time.Second)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to control socket (is 'node serve' running?): %w", err)
	}

	return rpc.NewClient(conn), nil
}

// getControlSocketPath returns the path to the control socket.
func getControlSocketPath() (string, error) {
	return xdg.RuntimeFile("lethean-desktop/node.sock")
}
