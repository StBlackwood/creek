package replication

import (
	"creek/internal/commons"
	"creek/internal/config"
	"creek/internal/logger"
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
func NewRepService(cfg *config.Config) *RepService {
	qs := &RepService{
		Nodes: make(map[string]*Node),
		Conf:  cfg,
		log:   logger.CreateLogger(cfg.LogLevel),
	}
	return qs
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
