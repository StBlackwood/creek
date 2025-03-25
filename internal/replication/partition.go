package replication

import (
	"creek/internal/commons"
	"creek/internal/config"
	"creek/internal/datastore"
	"creek/internal/logger"
	"github.com/sirupsen/logrus"
	"sync"
	"time"
)

type PartitionRepCmdWriteHandler func(cmd *RepCmd) error

type Partition struct {
	LW   *LogEntryWriter
	DS   *datastore.DataStore
	Mu   sync.Mutex
	Id   int
	Mode commons.PartitionMode
	log  *logrus.Logger

	writeChan chan *RepCmd
}

func (p *Partition) SendWriteCommand(cmd *RepCmd) {
	cmd.PartitionId = p.Id

	select {
	case p.writeChan <- cmd:
		// Successfully written
	case <-time.After(1 * time.Second): // Timeout after 1 second
		logrus.Error("WriteCommand timed out: channel full")
	}
}

// NewPartition initializes a Partition with a custom handler for processing commands.
func NewPartition(id int, cfg *config.Config, ds *datastore.DataStore) (*Partition, error) {
	logFileName := "commit.log"
	logFilePath := cfg.DataStoreDirectory + "/" + logFileName
	writer, err := newLogEntryWriter(logFilePath)
	if err != nil {
		return nil, err
	}

	p := &Partition{
		LW:        writer, // Assume LogEntryWriter is initialized elsewhere
		DS:        ds,
		Id:        id,
		Mode:      cfg.ServerMode,
		log:       logger.CreateLogger(cfg.LogLevel),
		writeChan: make(chan *RepCmd, 100), // Buffered channel for async writes
	}

	return p, nil
}

func (p *Partition) AttachRepCmdWriteHandler(handler PartitionRepCmdWriteHandler) {
	go p.listenForWrites(handler)
}

// listenForWrites listens for commands and delegates handling to the provided function.
func (p *Partition) listenForWrites(handler PartitionRepCmdWriteHandler) {
	for cmd := range p.writeChan {
		err := handler(cmd)
		if err != nil {
			p.log.Errorf("Error handling rep write command: %v", err)
			continue
		}
	}
}

// StopPartition ensures graceful shutdown.
func (p *Partition) StopPartition() {
	close(p.writeChan)
}
