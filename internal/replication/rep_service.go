package replication

import (
	"creek/internal/commons"
	"creek/internal/config"
	"creek/internal/logger"
	"fmt"
	"github.com/sirupsen/logrus"
	"net"
	"sync"
)

type RepService struct {
	Nodes map[string]*Node
	Conf  *config.Config
	mu    sync.Mutex
	log   *logrus.Logger
}

// GetNodes returns all nodes right now, once data partition is introduced this result will be based on partitionId
func (qs *RepService) GetNodes(partitionId int) []*Node {
	qs.mu.Lock()
	defer qs.mu.Unlock()

	nodes := make([]*Node, 0, len(qs.Nodes))
	for _, node := range qs.Nodes {
		nodes = append(nodes, node)
	}
	return nodes
}
func NewRepService(cfg *config.Config) (*RepService, error) {
	qs := &RepService{
		Nodes: make(map[string]*Node),
		Conf:  cfg,
		log:   logger.CreateLogger(cfg.LogLevel),
	}
	return qs, nil
}

func (qs *RepService) ConnectToFollowers() {
	if qs.Conf.ServerMode != commons.Leader {
		return
	}

	for _, addr := range qs.Conf.PeerNodes {
		go func(address string) {
			conn, err := net.Dial("tcp", address)
			if err != nil {
				qs.log.Errorf("Failed to connect to peer %s: %v", address, err)
				return
			}
			qs.addNode(&Node{
				Id:       address,
				Address:  address,
				conn:     conn,
				IsSelf:   false,
				IsLeader: false,
			})
		}(addr)
	}
}

func (qs *RepService) addNode(node *Node) {
	qs.mu.Lock()
	defer qs.mu.Unlock()
	qs.Nodes[node.Id] = node
}

func (qs *RepService) HandleRepCmdWrite(cmd *RepCmd) error {
	nodes := qs.GetNodes(cmd.PartitionId)
	for _, node := range nodes {
		err := node.SendRepCmd(cmd)
		if err != nil {
			return err
		}
	}
	return nil
}

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
