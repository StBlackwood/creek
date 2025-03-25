package core

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"
)

func (s *StateMachine) recoverOnStart() error {
	// partition will be locked until state is completely recovered
	s.p.Mu.Lock()
	defer s.p.Mu.Unlock()

	logFile, err := os.OpenFile(s.conf.DataStoreDirectory+"/commit.log", os.O_RDONLY, 0644)

	if err != nil {
		return fmt.Errorf("failed to open commit log: %w", err)
	}
	defer func(logFile *os.File) {
		err := logFile.Close()
		if err != nil {
			s.log.Warnf("Error closing commit log in recovery: %v", err)
		}
	}(logFile)

	reader := bufio.NewReader(logFile)
	batchSize := 100 // Adjust based on available memory
	var batch []string

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			return fmt.Errorf("error reading commit log: %w", err)
		}

		batch = append(batch, strings.TrimSpace(line))
		if len(batch) >= batchSize {
			s.processBatch(&batch)
			batch = batch[:0] // Clear batch
		}
	}

	// Process any remaining entries
	if len(batch) > 0 {
		s.processBatch(&batch)
	}

	return nil
}

func (s *StateMachine) processBatch(entries *[]string) {
	now := time.Now().UnixNano()

	for _, line := range *entries {
		parts := strings.Fields(line)
		if len(parts) < 3 {
			continue
		}

		timestamp, err := strconv.ParseInt(parts[0], 10, 64)
		if err != nil {
			s.log.Warnf("Skipping malformed log entry: %s", line)
			continue
		}

		op := parts[1]
		args := parts[2:]

		s.processLogEntry(timestamp, op, args, now)
	}
}

func (s *StateMachine) processLogEntry(timestamp int64, operation string, args []string, now int64) {
	switch operation {
	case "SET":
		if len(args) < 2 {
			return
		}
		key, value := args[0], args[1]
		ttl := -1
		if len(args) > 2 {
			parsedTTL, err := strconv.Atoi(args[2])
			if err == nil {
				ttl = parsedTTL
			}
		}
		if ttl > 0 && (timestamp+int64(ttl)*int64(time.Second)) <= now {
			s.p.DS.Delete(key)
		} else {
			s.p.DS.Set(key, value, ttl-(int(now-timestamp)/int(time.Second)))
		}

	case "DELETE":
		if len(args) < 1 {
			return
		}
		s.p.DS.Delete(args[0])

	case "EXPIRE":
		if len(args) < 2 {
			return
		}
		key := args[0]
		ttl, err := strconv.Atoi(args[1])
		if err == nil {
			if (timestamp + int64(ttl)*int64(time.Second)) <= now {
				s.p.DS.Delete(key)
			} else {
				s.p.DS.Expire(key, ttl-(int(now-timestamp)/int(time.Second)))
			}
		}
	}
}
