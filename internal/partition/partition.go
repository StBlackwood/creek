package partition

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

type RepCmdWriteHandler func(cmd *replication.RepCmd) error

type Partition struct {
	Id         int
	SelfNodeId string

	lw *LogEntryWriter
	ds *datastore.DataStore
	mu sync.Mutex

	PartitionMode commons.PartitionMode
	WriteMode     commons.WriteConsistencyMode

	Version int

	log *logrus.Logger

	writeChan   chan *replication.RepCmd
	stopLWFlush chan struct{}
	stopGC      chan struct{}
}

func (p *Partition) SendWriteCommand(cmd *replication.RepCmd) {
	cmd.PartitionId = p.Id

	select {
	case p.writeChan <- cmd:
		// Successfully written
	case <-time.After(1 * time.Second): // Timeout after 1 second
		logrus.Error("WriteCommand timed out: channel full")
	}
}

// NewPartition initializes a Partition with a custom handler for processing commands.
func NewPartition(id int, nodeId string, cfg *config.Config, ds *datastore.DataStore) (*Partition, error) {
	logFileName := "commit.log"
	logFilePath := cfg.DataStoreDirectory + "/" + logFileName
	writer, err := newLogEntryWriter(logFilePath)
	if err != nil {
		return nil, err
	}

	p := &Partition{
		Id:            id,
		SelfNodeId:    nodeId,
		lw:            writer, // Assume LogEntryWriter is initialized elsewhere
		ds:            ds,
		PartitionMode: cfg.ServerMode,
		log:           logger.CreateLogger(cfg.LogLevel),
		Version:       0,
		writeChan:     make(chan *replication.RepCmd, 100), // Buffered channel for async writes
		WriteMode:     cfg.WriteConsistencyMode,
		stopLWFlush:   make(chan struct{}),
		stopGC:        make(chan struct{}),
	}

	return p, nil
}

func (p *Partition) Start() error {
	err := p.recoverOnStart()
	if err != nil {
		return err
	}
	p.startLWFlush()
	if p.PartitionMode == commons.Leader {
		p.startGC() // Start garbage collection only in leader mode. followers will receive expire deletes from leader
	}
	return nil
}

func (p *Partition) AttachRepCmdWriteHandler(handler RepCmdWriteHandler) {
	go p.listenForWrites(handler)
}

// listenForWrites listens for commands and delegates handling to the provided function.
func (p *Partition) listenForWrites(handler RepCmdWriteHandler) {
	for cmd := range p.writeChan {
		err := handler(cmd)
		if err != nil {
			p.log.Errorf("Error handling rep write command: %v", err)
			continue
		}
	}
}

// StopPartition ensures graceful shutdown.
func (p *Partition) StopPartition() error {
	p.ds.Stop()
	close(p.stopLWFlush)
	close(p.stopGC)
	close(p.writeChan)
	return p.lw.Close()
}

func (p *Partition) startLWFlush() {
	go func() {
		ticker := time.NewTicker(5 * time.Second) // Adjust interval as needed
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				err := p.lw.Flush()
				if err != nil {
					p.log.Error("Error flushing log entries: ", err)
					return
				}
			case <-p.stopLWFlush:
				p.log.Info("Stopping log entry writer...")
				return
			}
		}
	}()
}

func (p *Partition) startGC() {
	go func() {
		ticker := time.NewTicker(10 * time.Second) // Adjust interval as needed
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				p.cleanExpiredKeys()
			case <-p.stopGC:
				p.log.Info("Stopping datastore garbage collection...")
				return
			}
		}
	}()
}

func (p *Partition) cleanExpiredKeys() {
	p.mu.Lock()
	defer p.mu.Unlock()
	expiredKeys := p.ds.GetExpiredKeys()
	for _, key := range expiredKeys {
		err := p.deleteWithoutLock(key)
		if err != nil {
			p.log.Warnf("Error deleting expired key: %v", err)
			continue
		}
	}
}

func (p *Partition) Set(key, value string, ttl int) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.Version++

	entry := LogEntry{
		Timestamp: time.Now().UnixNano(),
		Version:   p.Version,
		Operation: commons.CmdDataSet,
		Args:      []string{key, value, fmt.Sprintf("%d", ttl)},
	}

	err := p.lw.Append(entry)
	if err != nil {
		return err
	}
	if p.WriteMode == commons.StrongConsistency {
		err := p.lw.Flush()
		if err != nil {
			return err
		}
	}

	p.ds.Set(key, value, ttl)
	p.SendWriteCommand(
		&replication.RepCmd{
			Origin:      p.SelfNodeId,
			PartitionId: p.Id,
			Timestamp:   time.Now().UnixNano(),
			Version:     p.Version,
			Operation:   commons.CmdDataSet,
			Args:        []string{key, value, fmt.Sprintf("%d", ttl)},
		},
	)
	return nil
}

func (p *Partition) Get(key string) (string, error) {
	return p.ds.Get(key), nil
}

func (p *Partition) Delete(key string) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.deleteWithoutLock(key)
}

func (p *Partition) deleteWithoutLock(key string) error {
	p.Version++
	entry := LogEntry{
		Timestamp: time.Now().UnixNano(),
		Version:   p.Version,
		Operation: commons.CmdDataDel,
		Args:      []string{key},
	}

	err := p.lw.Append(entry)
	if err != nil {
		return err
	}
	if p.WriteMode == commons.StrongConsistency {
		err := p.lw.Flush()
		if err != nil {
			return err
		}
	}
	p.ds.Delete(key)

	p.SendWriteCommand(
		&replication.RepCmd{
			Origin:      p.SelfNodeId,
			PartitionId: p.Id,
			Timestamp:   time.Now().UnixNano(),
			Version:     p.Version,
			Operation:   commons.CmdDataDel,
			Args:        []string{key},
		})
	return nil
}

func (p *Partition) Expire(key string, ttl int) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.Version++

	entry := LogEntry{
		Timestamp: time.Now().UnixNano(),
		Version:   p.Version,
		Operation: commons.CmdDataEXP,
		Args:      []string{key, strconv.Itoa(ttl)},
	}

	err := p.lw.Append(entry)
	if err != nil {
		return err
	}
	if p.WriteMode == commons.StrongConsistency {
		err := p.lw.Flush()
		if err != nil {
			return err
		}
	}
	p.ds.Expire(key, ttl)

	p.SendWriteCommand(
		&replication.RepCmd{
			Origin:      p.SelfNodeId,
			PartitionId: p.Id,
			Timestamp:   time.Now().UnixNano(),
			Version:     p.Version,
			Operation:   commons.CmdDataEXP,
			Args:        []string{key, strconv.Itoa(ttl)},
		})
	return nil
}

func (p *Partition) TTL(key string) (int, error) {
	return p.ds.TTL(key), nil
}

func (p *Partition) ProcessRepCmd(cmd *replication.RepCmd) error {
	if p.PartitionMode == commons.Leader {
		return fmt.Errorf("partition is not in follower mode to accept replication commands")
	}

	switch cmd.Operation {
	case commons.CmdDataSet:
		if len(cmd.Args) < 2 {
			return fmt.Errorf("invalid args in rep command: %s", cmd.String())
		}
		key, value := cmd.Args[0], cmd.Args[1]
		ttl := -1
		if len(cmd.Args) > 2 {
			parsedTTL, err := strconv.Atoi(cmd.Args[2])
			if err == nil {
				ttl = parsedTTL
			}
		}
		return p.Set(key, value, ttl)

	case commons.CmdDataDel:
		if len(cmd.Args) < 1 {
			return fmt.Errorf("invalid args in rep command: %s", cmd.String())
		}
		return p.Delete(cmd.Args[0])

	case commons.CmdDataEXP:
		if len(cmd.Args) < 2 {
			return fmt.Errorf("invalid args in rep command: %s", cmd.String())
		}
		key := cmd.Args[0]
		ttl, err := strconv.Atoi(cmd.Args[1])
		if err != nil {
			return fmt.Errorf("invalid args in rep command: %s", cmd.String())
		}
		return p.Expire(key, ttl)

	default:
		return fmt.Errorf("invalid operation in rep command: %s", cmd.String())
	}
}
