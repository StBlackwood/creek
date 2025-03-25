package core

import (
	"creek/internal/config"
	"creek/internal/datastore"
	"creek/internal/logger"
	"creek/internal/partition"
	"creek/internal/replication"
	"github.com/sirupsen/logrus"
)

type StateMachine struct {
	p *partition.Partition

	NodeId string

	log  *logrus.Logger
	conf *config.Config
}

func NewStateMachine(NodeId string, cfg *config.Config) (*StateMachine, error) {

	store := datastore.NewDataStore(cfg)
	p, _ := partition.NewPartition(0, NodeId, cfg, store)

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
	return p.Set(key, value, ttl)
}

func (s *StateMachine) Delete(key string) error {
	p, err := s.getPartitionFromKey(key)
	if err != nil {
		return err
	}
	return p.Delete(key)
}

func (s *StateMachine) Expire(key string, ttl int) error {
	p, err := s.getPartitionFromKey(key)
	if err != nil {
		return err
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
