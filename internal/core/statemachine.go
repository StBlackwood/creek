package core

import (
	"creek/internal/commons"
	"creek/internal/config"
	"creek/internal/datastore"
	"creek/internal/logger"
	"creek/internal/replication"
	"fmt"
	"github.com/sirupsen/logrus"
	"strconv"
	"sync"
	"time"
)

type Partition struct {
	lw *replication.LogEntryWriter
	ds *datastore.DataStore
	mu sync.Mutex
}

type StateMachine struct {
	p *Partition

	writeMode commons.WriteConsistencyMode
	log       *logrus.Logger
	conf      *config.Config

	stopLWFlush chan struct{}
	stopGC      chan struct{}
}

func NewStateMachine(cfg *config.Config) (*StateMachine, error) {

	logFileName := "commit.log"
	logFilePath := cfg.DataStoreDirectory + "/" + logFileName
	writer, err := replication.NewLogEntryWriter(logFilePath)
	if err != nil {
		return nil, err
	}

	p := &Partition{
		lw: writer,
		ds: datastore.NewDataStore(cfg),
	}
	sm := &StateMachine{
		p:         p,
		writeMode: cfg.WriteConsistencyMode,
		log:       logger.CreateLogger(cfg.LogLevel),
		conf:      cfg,

		stopLWFlush: make(chan struct{}),
		stopGC:      make(chan struct{}),
	}
	return sm, nil
}

func (s *StateMachine) Start() error {
	err := s.recoverOnStart()
	if err != nil {
		return err
	}
	s.startLWFlush()
	s.startGC() // Start garbage collection
	return nil
}

// startGC runs garbage collection to remove expired keys
func (s *StateMachine) startGC() {
	go func() {
		ticker := time.NewTicker(10 * time.Second) // Adjust interval as needed
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				s.cleanExpiredKeys()
			case <-s.stopGC:
				s.log.Info("Stopping datastore garbage collection...")
				return
			}
		}
	}()
}

func (s *StateMachine) cleanExpiredKeys() {
	s.p.mu.Lock()
	defer s.p.mu.Unlock()
	expiredKeys := s.p.ds.GetExpiredKeys()
	for _, key := range expiredKeys {
		err := s.deleteWithoutLock(key)
		if err != nil {
			s.log.Warnf("Error deleting expired key: %v", err)
			continue
		}
	}
}

func (s *StateMachine) startLWFlush() {
	go func() {
		ticker := time.NewTicker(5 * time.Second) // Adjust interval as needed
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				err := s.p.lw.Flush()
				if err != nil {
					s.log.Error("Error flushing log entries: ", err)
					return
				}
			case <-s.stopLWFlush:
				s.log.Info("Stopping log entry writer...")
				return
			}
		}
	}()
}

func (s *StateMachine) Get(key string) (string, error) {
	return s.p.ds.Get(key), nil
}

func (s *StateMachine) Set(key, value string, ttl int) error {
	s.p.mu.Lock()
	defer s.p.mu.Unlock()
	entry := replication.LogEntry{
		Timestamp: time.Now().UnixNano(),
		Operation: "SET",
		Args:      []string{key, value, fmt.Sprintf("%d", ttl)},
	}

	err := s.p.lw.Append(entry)
	if err != nil {
		return err
	}
	if s.writeMode == commons.StrongConsistency {
		err := s.p.lw.Flush()
		if err != nil {
			return err
		}
	}

	s.p.ds.Set(key, value, ttl)
	return nil
}

func (s *StateMachine) Delete(key string) error {
	s.p.mu.Lock()
	defer s.p.mu.Unlock()
	return s.deleteWithoutLock(key)
}

func (s *StateMachine) deleteWithoutLock(key string) error {
	entry := replication.LogEntry{
		Timestamp: time.Now().UnixNano(),
		Operation: "DELETE",
		Args:      []string{key},
	}

	err := s.p.lw.Append(entry)
	if err != nil {
		return err
	}
	if s.writeMode == commons.StrongConsistency {
		err := s.p.lw.Flush()
		if err != nil {
			return err
		}
	}
	s.p.ds.Delete(key)
	return nil
}

func (s *StateMachine) Stop() error {
	// Perform any necessary cleanup or shutdown operations
	s.p.ds.Stop()
	close(s.stopLWFlush)
	close(s.stopGC)
	return s.p.lw.Close()
}

func (s *StateMachine) Expire(key string, ttl int) error {
	s.p.mu.Lock()
	defer s.p.mu.Unlock()

	entry := replication.LogEntry{
		Timestamp: time.Now().UnixNano(),
		Operation: "EXPIRE",
		Args:      []string{key, strconv.Itoa(ttl)},
	}

	err := s.p.lw.Append(entry)
	if err != nil {
		return err
	}
	if s.writeMode == commons.StrongConsistency {
		err := s.p.lw.Flush()
		if err != nil {
			return err
		}
	}
	s.p.ds.Expire(key, ttl)
	return nil
}

func (s *StateMachine) TTL(key string) (int, error) {
	return s.p.ds.TTL(key), nil
}
