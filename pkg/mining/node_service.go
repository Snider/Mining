package mining

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/Snider/Mining/pkg/node"
	"github.com/gin-gonic/gin"
)

// NodeService handles P2P node-related API endpoints.
type NodeService struct {
	nodeManager  *node.NodeManager
	peerRegistry *node.PeerRegistry
	transport    *node.Transport
	controller   *node.Controller
	worker       *node.Worker
}

// NewNodeService creates a new NodeService instance.
func NewNodeService() (*NodeService, error) {
	nm, err := node.NewNodeManager()
	if err != nil {
		return nil, err
	}

	pr, err := node.NewPeerRegistry()
	if err != nil {
		return nil, err
	}

	config := node.DefaultTransportConfig()
	transport := node.NewTransport(nm, pr, config)

	ns := &NodeService{
		nodeManager:  nm,
		peerRegistry: pr,
		transport:    transport,
	}

	// Initialize controller and worker
	ns.controller = node.NewController(nm, pr, transport)
	ns.worker = node.NewWorker(nm, transport)

	return ns, nil
}

// SetupRoutes configures all node-related API routes.
func (ns *NodeService) SetupRoutes(router *gin.RouterGroup) {
	// Node identity endpoints
	nodeGroup := router.Group("/node")
	{
		nodeGroup.GET("/info", ns.handleNodeInfo)
		nodeGroup.POST("/init", ns.handleNodeInit)
	}

	// Peer management endpoints
	peerGroup := router.Group("/peers")
	{
		peerGroup.GET("", ns.handleListPeers)
		peerGroup.POST("", ns.handleAddPeer)
		peerGroup.GET("/:id", ns.handleGetPeer)
		peerGroup.DELETE("/:id", ns.handleRemovePeer)
		peerGroup.POST("/:id/ping", ns.handlePingPeer)
		peerGroup.POST("/:id/connect", ns.handleConnectPeer)
		peerGroup.POST("/:id/disconnect", ns.handleDisconnectPeer)
	}

	// Remote operations endpoints
	remoteGroup := router.Group("/remote")
	{
		remoteGroup.GET("/stats", ns.handleRemoteStats)
		remoteGroup.GET("/:peerId/stats", ns.handlePeerStats)
		remoteGroup.POST("/:peerId/start", ns.handleRemoteStart)
		remoteGroup.POST("/:peerId/stop", ns.handleRemoteStop)
		remoteGroup.GET("/:peerId/logs/:miner", ns.handleRemoteLogs)
	}
}

// StartTransport starts the P2P transport server.
func (ns *NodeService) StartTransport() error {
	return ns.transport.Start()
}

// StopTransport stops the P2P transport server.
func (ns *NodeService) StopTransport() error {
	return ns.transport.Stop()
}

// Node Info Response
type NodeInfoResponse struct {
	HasIdentity    bool                `json:"hasIdentity"`
	Identity       *node.NodeIdentity  `json:"identity,omitempty"`
	RegisteredPeers int                `json:"registeredPeers"`
	ConnectedPeers  int                `json:"connectedPeers"`
}

// handleNodeInfo godoc
// @Summary Get node identity information
// @Description Get the current node's identity and connection status
// @Tags node
// @Produce json
// @Success 200 {object} NodeInfoResponse
// @Router /node/info [get]
func (ns *NodeService) handleNodeInfo(c *gin.Context) {
	response := NodeInfoResponse{
		HasIdentity:     ns.nodeManager.HasIdentity(),
		RegisteredPeers: ns.peerRegistry.Count(),
		ConnectedPeers:  len(ns.peerRegistry.GetConnectedPeers()),
	}

	if ns.nodeManager.HasIdentity() {
		response.Identity = ns.nodeManager.GetIdentity()
	}

	c.JSON(http.StatusOK, response)
}

// NodeInitRequest is the request body for node initialization.
type NodeInitRequest struct {
	Name string `json:"name" binding:"required"`
	Role string `json:"role"` // "controller", "worker", or "dual"
}

// handleNodeInit godoc
// @Summary Initialize node identity
// @Description Create a new node identity with X25519 keypair
// @Tags node
// @Accept json
// @Produce json
// @Param request body NodeInitRequest true "Node initialization parameters"
// @Success 200 {object} node.NodeIdentity
// @Router /node/init [post]
func (ns *NodeService) handleNodeInit(c *gin.Context) {
	var req NodeInitRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if ns.nodeManager.HasIdentity() {
		c.JSON(http.StatusConflict, gin.H{"error": "node identity already exists"})
		return
	}

	role := node.RoleDual
	switch req.Role {
	case "controller":
		role = node.RoleController
	case "worker":
		role = node.RoleWorker
	case "dual", "":
		role = node.RoleDual
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid role"})
		return
	}

	if err := ns.nodeManager.GenerateIdentity(req.Name, role); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, ns.nodeManager.GetIdentity())
}

// handleListPeers godoc
// @Summary List registered peers
// @Description Get a list of all registered peers with their status
// @Tags peers
// @Produce json
// @Success 200 {array} node.Peer
// @Router /peers [get]
func (ns *NodeService) handleListPeers(c *gin.Context) {
	peers := ns.peerRegistry.ListPeers()
	c.JSON(http.StatusOK, peers)
}

// AddPeerRequest is the request body for adding a peer.
type AddPeerRequest struct {
	Address string `json:"address" binding:"required"`
	Name    string `json:"name"`
}

// handleAddPeer godoc
// @Summary Add a new peer
// @Description Register a new peer node by address
// @Tags peers
// @Accept json
// @Produce json
// @Param request body AddPeerRequest true "Peer information"
// @Success 201 {object} node.Peer
// @Router /peers [post]
func (ns *NodeService) handleAddPeer(c *gin.Context) {
	var req AddPeerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	peer := &node.Peer{
		ID:      "pending-" + req.Address, // Will be updated on handshake
		Name:    req.Name,
		Address: req.Address,
		Role:    node.RoleDual,
		Score:   50,
	}

	if err := ns.peerRegistry.AddPeer(peer); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, peer)
}

// handleGetPeer godoc
// @Summary Get peer information
// @Description Get information about a specific peer
// @Tags peers
// @Produce json
// @Param id path string true "Peer ID"
// @Success 200 {object} node.Peer
// @Router /peers/{id} [get]
func (ns *NodeService) handleGetPeer(c *gin.Context) {
	peerID := c.Param("id")
	peer := ns.peerRegistry.GetPeer(peerID)
	if peer == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "peer not found"})
		return
	}
	c.JSON(http.StatusOK, peer)
}

// handleRemovePeer godoc
// @Summary Remove a peer
// @Description Remove a peer from the registry
// @Tags peers
// @Produce json
// @Param id path string true "Peer ID"
// @Success 200 {object} map[string]string
// @Router /peers/{id} [delete]
func (ns *NodeService) handleRemovePeer(c *gin.Context) {
	peerID := c.Param("id")
	if err := ns.peerRegistry.RemovePeer(peerID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "peer removed"})
}

// handlePingPeer godoc
// @Summary Ping a peer
// @Description Send a ping to a peer and measure latency
// @Tags peers
// @Produce json
// @Param id path string true "Peer ID"
// @Success 200 {object} map[string]float64
// @Router /peers/{id}/ping [post]
func (ns *NodeService) handlePingPeer(c *gin.Context) {
	peerID := c.Param("id")
	rtt, err := ns.controller.PingPeer(peerID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"rtt_ms": rtt})
}

// handleConnectPeer godoc
// @Summary Connect to a peer
// @Description Establish a WebSocket connection to a peer
// @Tags peers
// @Produce json
// @Param id path string true "Peer ID"
// @Success 200 {object} map[string]string
// @Router /peers/{id}/connect [post]
func (ns *NodeService) handleConnectPeer(c *gin.Context) {
	peerID := c.Param("id")
	if err := ns.controller.ConnectToPeer(peerID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "connected"})
}

// handleDisconnectPeer godoc
// @Summary Disconnect from a peer
// @Description Close the connection to a peer
// @Tags peers
// @Produce json
// @Param id path string true "Peer ID"
// @Success 200 {object} map[string]string
// @Router /peers/{id}/disconnect [post]
func (ns *NodeService) handleDisconnectPeer(c *gin.Context) {
	peerID := c.Param("id")
	if err := ns.controller.DisconnectFromPeer(peerID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "disconnected"})
}

// handleRemoteStats godoc
// @Summary Get stats from all remote peers
// @Description Fetch mining statistics from all connected peers
// @Tags remote
// @Produce json
// @Success 200 {object} map[string]node.StatsPayload
// @Router /remote/stats [get]
func (ns *NodeService) handleRemoteStats(c *gin.Context) {
	stats := ns.controller.GetAllStats()
	c.JSON(http.StatusOK, stats)
}

// handlePeerStats godoc
// @Summary Get stats from a specific peer
// @Description Fetch mining statistics from a specific peer
// @Tags remote
// @Produce json
// @Param peerId path string true "Peer ID"
// @Success 200 {object} node.StatsPayload
// @Router /remote/{peerId}/stats [get]
func (ns *NodeService) handlePeerStats(c *gin.Context) {
	peerID := c.Param("peerId")
	stats, err := ns.controller.GetRemoteStats(peerID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, stats)
}

// RemoteStartRequest is the request body for starting a remote miner.
type RemoteStartRequest struct {
	MinerType string          `json:"minerType" binding:"required"`
	ProfileID string          `json:"profileId,omitempty"`
	Config    json.RawMessage `json:"config,omitempty"`
}

// handleRemoteStart godoc
// @Summary Start miner on remote peer
// @Description Start a miner on a remote peer using a profile
// @Tags remote
// @Accept json
// @Produce json
// @Param peerId path string true "Peer ID"
// @Param request body RemoteStartRequest true "Start parameters"
// @Success 200 {object} map[string]string
// @Router /remote/{peerId}/start [post]
func (ns *NodeService) handleRemoteStart(c *gin.Context) {
	peerID := c.Param("peerId")
	var req RemoteStartRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := ns.controller.StartRemoteMiner(peerID, req.MinerType, req.ProfileID, req.Config); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "miner started"})
}

// RemoteStopRequest is the request body for stopping a remote miner.
type RemoteStopRequest struct {
	MinerName string `json:"minerName" binding:"required"`
}

// handleRemoteStop godoc
// @Summary Stop miner on remote peer
// @Description Stop a running miner on a remote peer
// @Tags remote
// @Accept json
// @Produce json
// @Param peerId path string true "Peer ID"
// @Param request body RemoteStopRequest true "Stop parameters"
// @Success 200 {object} map[string]string
// @Router /remote/{peerId}/stop [post]
func (ns *NodeService) handleRemoteStop(c *gin.Context) {
	peerID := c.Param("peerId")
	var req RemoteStopRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := ns.controller.StopRemoteMiner(peerID, req.MinerName); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "miner stopped"})
}

// handleRemoteLogs godoc
// @Summary Get logs from remote miner
// @Description Retrieve console logs from a miner on a remote peer
// @Tags remote
// @Produce json
// @Param peerId path string true "Peer ID"
// @Param miner path string true "Miner Name"
// @Param lines query int false "Number of lines" default(100)
// @Success 200 {array} string
// @Router /remote/{peerId}/logs/{miner} [get]
func (ns *NodeService) handleRemoteLogs(c *gin.Context) {
	peerID := c.Param("peerId")
	minerName := c.Param("miner")
	lines := 100
	if l := c.Query("lines"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
			lines = parsed
		}
	}

	logs, err := ns.controller.GetRemoteLogs(peerID, minerName, lines)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, logs)
}
