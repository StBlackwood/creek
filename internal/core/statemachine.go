package core

import (
	"creek/internal/config"
	"creek/internal/datastore"
	"creek/internal/logger"
	"creek/internal/partition"
	"github.com/sirupsen/logrus"
)

type StateMachine struct {
	p *partition.Partition

	log  *logrus.Logger
	conf *config.Config
}

func NewStateMachine(cfg *config.Config) (*StateMachine, error) {

	store := datastore.NewDataStore(cfg)
	p, _ := partition.NewPartition(0, cfg, store)

	sm := &StateMachine{
		p:    p,
		log:  logger.CreateLogger(cfg.LogLevel),
		conf: cfg,
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

func (s *StateMachine) getPartition(key string) (*partition.Partition, error) {
	return s.p, nil
}

func (s *StateMachine) Get(key string) (string, error) {
	partition, err := s.getPartition(key)
	if err != nil {
		return "", err
	}
	return partition.Get(key)
}

func (s *StateMachine) Set(key, value string, ttl int) error {
	partition, err := s.getPartition(key)
	if err != nil {
		return err
	}
	return partition.Set(key, value, ttl)
}

func (s *StateMachine) Delete(key string) error {
	partition, err := s.getPartition(key)
	if err != nil {
		return err
	}
	return partition.Delete(key)
}

func (s *StateMachine) Expire(key string, ttl int) error {
	partition, err := s.getPartition(key)
	if err != nil {
		return err
	}
	return partition.Expire(key, ttl)
}

func (s *StateMachine) TTL(key string) (int, error) {
	partition, err := s.getPartition(key)
	if err != nil {
		return 0, err
	}
	return partition.TTL(key)
}
