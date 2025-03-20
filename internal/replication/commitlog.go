package replication

import (
	"bufio"
	"encoding/json"
	"os"
	"sync"
	"time"
)

// LogEntry represents a single entry in the commit log.
type LogEntry struct {
	Timestamp time.Time `json:"timestamp"`
	Key       string    `json:"key"`
	Value     string    `json:"value"`
	Action    string    `json:"action"` // e.g., "insert", "update", "delete"
}

// CommitLog is responsible for buffering and persisting log entries.
type CommitLog struct {
	mu         sync.Mutex
	buffer     []LogEntry
	logFile    *os.File
	writer     *bufio.Writer
	bufferSize int
}

// NewCommitLog initializes a new CommitLog.
func NewCommitLog(filePath string, bufferSize int) (*CommitLog, error) {
	file, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}

	return &CommitLog{
		buffer:     make([]LogEntry, 0, bufferSize),
		logFile:    file,
		writer:     bufio.NewWriter(file),
		bufferSize: bufferSize,
	}, nil
}

// Append adds a new LogEntry to the buffer.
func (cl *CommitLog) Append(entry LogEntry) error {
	cl.mu.Lock()
	defer cl.mu.Unlock()

	cl.buffer = append(cl.buffer, entry)
	if len(cl.buffer) >= cl.bufferSize {
		return cl.flush()
	}

	return nil
}

// flush writes the buffered log entries to the transaction log file.
func (cl *CommitLog) flush() error {
	for _, entry := range cl.buffer {
		data, err := json.Marshal(entry)
		if err != nil {
			return err
		}
		_, err = cl.writer.Write(data)
		if err != nil {
			return err
		}
		_, err = cl.writer.WriteString("\n")
		if err != nil {
			return err
		}
	}

	// Clear the buffer after flushing
	cl.buffer = cl.buffer[:0]
	return cl.writer.Flush()
}

// Close closes the transaction log file and ensures all data is flushed.
func (cl *CommitLog) Close() error {
	cl.mu.Lock()
	defer cl.mu.Unlock()

	if err := cl.flush(); err != nil {
		return err
	}

	return cl.logFile.Close()
}
