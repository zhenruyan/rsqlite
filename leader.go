package rsqlite

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/rqlite/gorqlite"
)

// LeaderInfo holds information about the current leader
type LeaderInfo struct {
	Leader string   `json:"leader"`
	Peers  []string `json:"peers"`
}

// NodeStatus represents the status of a node
type NodeStatus struct {
	Addr      string `json:"addr"`
	Leader    bool   `json:"leader"`
	Reachable bool   `json:"reachable"`
}

// ClusterManager manages cluster discovery and leader selection
type ClusterManager struct {
	nodes          []string
	leader         string
	peers          []string
	mu             sync.RWMutex
	lastUpdate     time.Time
	updateInterval time.Duration
	client         *http.Client
}

// NewClusterManager creates a new cluster manager
func NewClusterManager(nodes []string) *ClusterManager {
	return &ClusterManager{
		nodes:          nodes,
		updateInterval: 30 * time.Second,
		client:         &http.Client{Timeout: 10 * time.Second},
	}
}

// DiscoverLeader discovers the current leader and peers
func (cm *ClusterManager) DiscoverLeader(ctx context.Context) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	// If we recently updated, skip
	if time.Since(cm.lastUpdate) < cm.updateInterval {
		return nil
	}

	var lastErr error
	for _, node := range cm.nodes {
		leader, peers, err := cm.queryNodeStatus(ctx, node)
		if err != nil {
			lastErr = err
			continue
		}

		cm.leader = leader
		cm.peers = peers
		cm.lastUpdate = time.Now()
		return nil
	}

	return fmt.Errorf("failed to discover leader from any node: %w", lastErr)
}

// queryNodeStatus queries a node for its status
func (cm *ClusterManager) queryNodeStatus(ctx context.Context, node string) (string, []string, error) {
	statusURL := fmt.Sprintf("%s/status", node)

	req, err := http.NewRequestWithContext(ctx, "GET", statusURL, nil)
	if err != nil {
		return "", nil, err
	}

	resp, err := cm.client.Do(req)
	if err != nil {
		return "", nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", nil, fmt.Errorf("status request failed: %d", resp.StatusCode)
	}

	var status map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&status); err != nil {
		return "", nil, err
	}

	// Extract cluster info from status
	cluster, ok := status["cluster"].(map[string]interface{})
	if !ok {
		return "", nil, fmt.Errorf("invalid cluster info in status")
	}

	leader, _ := cluster["leader"].(string)
	if leader == "" {
		return "", nil, fmt.Errorf("no leader found")
	}

	// Extract peers
	var peers []string
	if peerList, ok := cluster["peers"].([]interface{}); ok {
		for _, peer := range peerList {
			if peerStr, ok := peer.(string); ok {
				peers = append(peers, peerStr)
			}
		}
	}

	return leader, peers, nil
}

// GetLeader returns the current leader
func (cm *ClusterManager) GetLeader() string {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	return cm.leader
}

// GetPeers returns the current peers
func (cm *ClusterManager) GetPeers() []string {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	return append([]string{}, cm.peers...)
}

// GetAllNodes returns all known nodes (leader + peers)
func (cm *ClusterManager) GetAllNodes() []string {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	nodes := make(map[string]bool)
	if cm.leader != "" {
		nodes[cm.leader] = true
	}
	for _, peer := range cm.peers {
		nodes[peer] = true
	}

	var result []string
	for node := range nodes {
		result = append(result, node)
	}

	return result
}

// SelectBestNode selects the best node to connect to based on consistency level
func (cm *ClusterManager) SelectBestNode(consistencyLevel string) string {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	// For strong consistency, always use leader
	if consistencyLevel == "strong" {
		if cm.leader != "" {
			return cm.leader
		}
	}

	// For weak/none consistency, we can use any node
	// Prefer leader if available, otherwise use any peer
	if cm.leader != "" {
		return cm.leader
	}

	if len(cm.peers) > 0 {
		return cm.peers[0]
	}

	// Fallback to original nodes
	if len(cm.nodes) > 0 {
		return cm.nodes[0]
	}

	return ""
}

// IsLeaderHealthy checks if the current leader is healthy
func (cm *ClusterManager) IsLeaderHealthy(ctx context.Context) bool {
	leader := cm.GetLeader()
	if leader == "" {
		return false
	}

	// Try to connect to leader
	client, err := gorqlite.Open(leader)
	if err != nil {
		return false
	}
	defer client.Close()

	// Simple health check
	_, err = client.QueryOneContext(ctx, "SELECT 1")
	return err == nil
}

// ForceRefresh forces a refresh of cluster information
func (cm *ClusterManager) ForceRefresh(ctx context.Context) error {
	cm.mu.Lock()
	cm.lastUpdate = time.Time{} // Reset last update time
	cm.mu.Unlock()

	return cm.DiscoverLeader(ctx)
}
