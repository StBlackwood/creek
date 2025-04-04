package replication

import (
	"creek/internal/commons"
	"creek/internal/config"
	"creek/internal/logger"
	"fmt"
	"github.com/sirupsen/logrus"
	"net"
	"sync"
	"time"
)

const maxAttempts = 5
const delayBetweenAttempts = time.Second * 5

// RepService represents a replication service that manages the communication between nodes in a distributed system.
type RepService struct {
	Nodes map[string]*Node // A map of connected nodes, keyed by their IDs.
	Conf  *config.Config   // The configuration for this replication service.
	mu    sync.Mutex       // A mutex to protect access to the Nodes map.
	log   *logrus.Logger   // A logger for logging messages related to this replication service.
}

// GetNodes returns all nodes right now, once data partition is introduced this result will be based on partitionId.
func (qs *RepService) GetNodes(partitionId int) []*Node {
	qs.mu.Lock()
	defer qs.mu.Unlock()

	nodes := make([]*Node, 0, len(qs.Nodes))
	for _, node := range qs.Nodes {
		nodes = append(nodes, node)
	}
	return nodes
}

// NewRepService creates a new replication service with the given configuration.
func NewRepService(cfg *config.Config) (*RepService, error) {
	qs := &RepService{
		Nodes: make(map[string]*Node),
		Conf:  cfg,
		log:   logger.CreateLogger(cfg.LogLevel),
	}
	return qs, nil
}

// ConnectToFollowers connects to all follower nodes in the distributed system.
func (qs *RepService) ConnectToFollowers() {
	if qs.Conf.ServerMode != commons.Leader {
		return
	}

	for _, address := range qs.Conf.PeerNodes {
		attempts := 0
		for attempts < maxAttempts {
			conn, err := net.Dial("tcp", address)
			if err != nil {
				qs.log.Warnf("Failed to connect to peer %s: %v", address, err)
				attempts++
				time.Sleep(delayBetweenAttempts)
			} else {
				node := &Node{
					Id:       address,
					Address:  address,
					conn:     conn,
					IsSelf:   false,
					IsLeader: false,
				}
				qs.addNode(node)
				break
			}
		}
	}
}

// addNode adds a new node to the replication service.
func (qs *RepService) addNode(node *Node) {
	qs.mu.Lock()
	defer qs.mu.Unlock()
	qs.Nodes[node.Id] = node
}

// HandleRepCmdWrite handles a write command by sending it to all nodes in the distributed system.
func (qs *RepService) HandleRepCmdWrite(cmd *RepCmd) error {
	nodes := qs.GetNodes(cmd.PartitionId)
	for _, node := range nodes {
		qs.log.Tracef("Sending write command to node: %v", node)
		err := node.SendRepCmd(cmd)
		if err != nil {
			return err
		}
	}
	return nil
}

// Stop stops the replication service and disconnects from all nodes in the distributed system.
func (qs *RepService) Stop() error {
	qs.mu.Lock()
	defer qs.mu.Unlock()

	isError := false

	for id, node := range qs.Nodes {
		err := node.Close()
		if err != nil {
			qs.log.Errorf("Error closing node %s: %v", id, err)
			isError = true
			continue
		}
		delete(qs.Nodes, id)
	}
	qs.log.Info("Disconnected from all nodes.")
	if isError {
		return fmt.Errorf("error closing few nodes")
	}
	return nil
}

// GetSelfNodeId returns the ID of the current node.
func (qs *RepService) GetSelfNodeId() string {
	return qs.Conf.ServerAddress
}
