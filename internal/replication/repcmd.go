package replication

import (
	"creek/internal/commons"
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

type RepCmd struct {
	PartitionId int
	NodeId      string
	Timestamp   int64
	Operation   string
	Args        []string
}

func (rm *RepCmd) String() string {
	return fmt.Sprintf(
		"%s %d %s %d %s %s\n",
		commons.SysRepMsg,
		rm.PartitionId,
		rm.NodeId,
		rm.Timestamp,
		rm.Operation,
		strings.Join(rm.Args, " "),
	)
}

func RepCmdFromString(s string) (*RepCmd, error) {
	s = strings.TrimSpace(s)

	// Extract SysRepMsg
	sysRepEnd := strings.IndexByte(s, ' ')
	if sysRepEnd == -1 {
		return nil, fmt.Errorf("invalid format: missing fields")
	}
	sysRepMsg := s[:sysRepEnd]
	if sysRepMsg != commons.SysRepMsg {
		return nil, fmt.Errorf("invalid format: invalid SysRepMsg")
	}

	// Extract PartitionId
	partitionStart := sysRepEnd + 1
	partitionEnd := strings.IndexByte(s[partitionStart:], ' ')
	if partitionEnd == -1 {
		return nil, fmt.Errorf("invalid format: missing fields")
	}
	partitionEnd += partitionStart
	partitionId, err := strconv.Atoi(s[partitionStart:partitionEnd])
	if err != nil {
		return nil, fmt.Errorf("invalid PartitionId: %v", err)
	}

	// Extract NodeId
	nodeIdStart := partitionEnd + 1
	nodeIdEnd := strings.IndexByte(s[nodeIdStart:], ' ')
	if nodeIdEnd == -1 {
		return nil, fmt.Errorf("invalid format: missing fields")
	}
	nodeIdEnd += nodeIdStart
	nodeId := s[nodeIdStart:nodeIdEnd]

	// Extract Timestamp
	timestampStart := nodeIdEnd + 1
	timestampEnd := strings.IndexByte(s[timestampStart:], ' ')
	if timestampEnd == -1 {
		return nil, fmt.Errorf("invalid format: missing fields")
	}
	timestampEnd += timestampStart
	timestamp, err := strconv.ParseInt(s[timestampStart:timestampEnd], 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid Timestamp: %v", err)
	}

	// Extract Operation
	operationStart := timestampEnd + 1
	operationEnd := strings.IndexByte(s[operationStart:], ' ')
	if operationEnd == -1 {
		// No arguments, only operation
		return &RepCmd{
			PartitionId: partitionId,
			NodeId:      nodeId,
			Timestamp:   timestamp,
			Operation:   s[operationStart:],
			Args:        nil,
		}, nil
	}
	operationEnd += operationStart
	operation := s[operationStart:operationEnd]

	// Extract Args (if any)
	args := strings.Fields(s[operationEnd+1:])

	return &RepCmd{
		PartitionId: partitionId,
		NodeId:      nodeId,
		Timestamp:   timestamp,
		Operation:   operation,
		Args:        args,
	}, nil
}

func (rm *RepCmd) Equals(other *RepCmd) bool {
	if other == nil {
		return false
	}
	return rm.PartitionId == other.PartitionId &&
		rm.NodeId == other.NodeId &&
		rm.Timestamp == other.Timestamp &&
		rm.Operation == other.Operation &&
		reflect.DeepEqual(rm.Args, other.Args)
}

func RepCmdFromArgs(args []string) (*RepCmd, error) {
	if len(args) < 5 {
		return nil, fmt.Errorf("invalid input: expected at least 5 arguments, got %d", len(args))
	}

	// Extract PartitionId
	partitionId, err := strconv.Atoi(args[0])
	if err != nil {
		return nil, fmt.Errorf("invalid PartitionId: %v", err)
	}

	// Extract Timestamp
	timestamp, err := strconv.ParseInt(args[2], 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid Timestamp: %v", err)
	}

	return &RepCmd{
		PartitionId: partitionId,
		NodeId:      args[1],
		Timestamp:   timestamp,
		Operation:   args[3],
		Args:        args[4:],
	}, nil
}
