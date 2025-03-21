package core

import (
	"creek/internal/commons"
	"creek/internal/config"
	"creek/internal/datastore"
	"creek/internal/logger"
	"creek/internal/replication"
	"github.com/sirupsen/logrus"
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
		log:       logger.GetLogger(),
		conf:      cfg,

		stopLWFlush: make(chan struct{}),
		stopGC:      make(chan struct{}),
	}
	return sm, nil
}

func (s *StateMachine) Start() error {
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
				s.p.mu.Lock()
				s.p.ds.CleanExpiredKeys()
				s.p.mu.Unlock()
			case <-s.stopGC:
				s.log.Info("Stopping datastore garbage collection...")
				return
			}
		}
	}()
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
	return s.p.ds.Get(key)
}
func (s *StateMachine) Set(key, value string, ttl int) error {
	s.p.ds.Set(key, value, ttl)
	return nil
}

func (s *StateMachine) Delete(key string) error {
	return s.p.ds.Delete(key)
}

func (s *StateMachine) Stop() error {
	// Perform any necessary cleanup or shutdown operations
	s.p.ds.Stop()
	close(s.stopLWFlush)
	close(s.stopGC)
	return s.p.lw.Close()
}

func (s *StateMachine) Expire(key string) error {
	return s.p.ds.Expire(key, 0)
}

func (s *StateMachine) TTL(key string) (int, error) {
	return s.p.ds.TTL(key)
}
