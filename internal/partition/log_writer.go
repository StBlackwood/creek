package partition

import (
	"fmt"
	"os"
	"strings"
	"sync"
)

// LogEntry represents a single operation in the transaction log.
type LogEntry struct {
	Timestamp int64 // timestamp
	Version   int
	Operation string // e.g., "set", "delete"
	Args      []string
}

// LogEntryWriter handles appending operations to a log file and replaying it to recreate the datastore state.
type LogEntryWriter struct {
	mu          sync.Mutex
	logFile     *os.File
	logFilePath string
}

// newLogEntryWriter initializes a transaction log and opens the file for writing.
func newLogEntryWriter(filePath string) (*LogEntryWriter, error) {
	file, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %w", err)
	}

	return &LogEntryWriter{
		logFile:     file,
		logFilePath: filePath,
	}, nil
}

// Append adds an operation to the transaction log.
func (t *LogEntryWriter) Append(entry LogEntry) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	// Format log entry as a string with timestamp and arguments
	logLine := fmt.Sprintf("%d %d %s %s\n",
		entry.Timestamp, entry.Version, entry.Operation, strings.Join(entry.Args, " "))

	if _, err := t.logFile.Write([]byte(logLine)); err != nil {
		return fmt.Errorf("failed to write log buffer to file: %w", err)
	}

	return nil
}

// Close releases resources related to the log file.
func (t *LogEntryWriter) Close() error {
	return t.logFile.Close()
}

// Flush ensures all buffered data is written to the log file.
func (t *LogEntryWriter) Flush() error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if err := t.logFile.Sync(); err != nil {
		return fmt.Errorf("failed to flush log file: %w", err)
	}

	return nil
}
