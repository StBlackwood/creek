package core

import (
	"creek/internal/commons"
	"creek/internal/config"
	"creek/internal/datastore"
	"creek/internal/logger"
	"creek/internal/partition"
	"creek/internal/replication"
	"fmt"
	"github.com/sirupsen/logrus"
)

type StateMachine struct {
	p *partition.Partition

	NodeId    string
	WriteMode commons.ReplicaMode

	log  *logrus.Logger
	conf *config.Config
}

func NewStateMachine(NodeId string, cfg *config.Config) (*StateMachine, error) {

	store := datastore.NewDataStore(cfg)
	p, err := partition.NewPartition(0, NodeId, cfg, store)
	if err != nil {
		panic(err)
	}

	sm := &StateMachine{
		p:      p,
		log:    logger.CreateLogger(cfg.LogLevel),
		conf:   cfg,
		NodeId: NodeId,
	}
	return sm, nil
}

func (s *StateMachine) Start() error {
	return s.p.Start()
}

func (s *StateMachine) Stop() error {
	// Perform any necessary cleanup or shutdown operations
	return s.p.StopPartition()
}

func (s *StateMachine) AttachRepCmdWriteHandlerToPartitions(handler partition.RepCmdWriteHandler) {
	// loop through all partitions
	s.p.AttachRepCmdWriteHandler(handler)
}

func (s *StateMachine) getPartitionFromKey(key string) (*partition.Partition, error) {
	return s.p, nil
}

func (s *StateMachine) getPartitionFromId(partitionId int) (*partition.Partition, error) {
	return s.p, nil
}

func (s *StateMachine) Get(key string) (string, error) {
	p, err := s.getPartitionFromKey(key)
	if err != nil {
		return "", err
	}
	return p.Get(key)
}

func (s *StateMachine) Set(key, value string, ttl int) error {
	p, err := s.getPartitionFromKey(key)
	if err != nil {
		return err
	}
	if s.WriteMode == commons.ReadOnlyReplication && p.PartitionMode == commons.Follower {
		return fmt.Errorf("write mode is read-only for the follower partition")
	}
	return p.Set(key, value, ttl)
}

func (s *StateMachine) Delete(key string) error {
	p, err := s.getPartitionFromKey(key)
	if err != nil {
		return err
	}
	if s.WriteMode == commons.ReadOnlyReplication && p.PartitionMode == commons.Follower {
		return fmt.Errorf("write mode is read-only for the follower partition")
	}
	return p.Delete(key)
}

func (s *StateMachine) Expire(key string, ttl int) error {
	p, err := s.getPartitionFromKey(key)
	if err != nil {
		return err
	}
	if s.WriteMode == commons.ReadOnlyReplication && p.PartitionMode == commons.Follower {
		return fmt.Errorf("write mode is read-only for the follower partition")
	}
	return p.Expire(key, ttl)
}

func (s *StateMachine) TTL(key string) (int, error) {
	p, err := s.getPartitionFromKey(key)
	if err != nil {
		return 0, err
	}
	return p.TTL(key)
}

func (s *StateMachine) ProcessRepCmd(cmd *replication.RepCmd) error {
	p, err := s.getPartitionFromId(cmd.PartitionId)
	if err != nil {
		return err
	}

	return p.ProcessRepCmd(cmd)
}
