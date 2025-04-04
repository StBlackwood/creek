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
	Origin      string
	Timestamp   int64
	Operation   string
	Args        []string
	Version     int
}

func (rm *RepCmd) String() string {
	return fmt.Sprintf(
		"%s %d %s %d %d %s %s\n",
		commons.CmdSysRep,
		rm.PartitionId,
		rm.Origin,
		rm.Timestamp,
		rm.Version,
		rm.Operation,
		strings.Join(rm.Args, " "),
	)
}

func RepCmdFromString(s string) (*RepCmd, error) {
	s = strings.TrimSpace(s)

	// Extract CmdSysRep
	sysRepEnd := strings.IndexByte(s, ' ')
	if sysRepEnd == -1 {
		return nil, fmt.Errorf("invalid format: missing fields")
	}
	sysRepMsg := s[:sysRepEnd]
	if sysRepMsg != commons.CmdSysRep {
		return nil, fmt.Errorf("invalid format: invalid CmdSysRep")
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

	// Extract Origin
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

	versionStart := timestampEnd + 1
	versionEnd := strings.IndexByte(s[versionStart:], ' ')
	if versionEnd == -1 {
		return nil, fmt.Errorf("invalid format: missing fields")
	}
	versionEnd += versionStart
	version, err := strconv.Atoi(s[versionStart:versionEnd])
	if err != nil {
		return nil, fmt.Errorf("invalid Version: %v", err)
	}

	// Extract Operation
	operationStart := versionEnd + 1
	operationEnd := strings.IndexByte(s[operationStart:], ' ')
	if operationEnd == -1 {
		// No arguments, only operation
		return &RepCmd{
			PartitionId: partitionId,
			Origin:      nodeId,
			Timestamp:   timestamp,
			Operation:   s[operationStart:],
			Version:     version,
			Args:        nil,
		}, nil
	}
	operationEnd += operationStart
	operation := s[operationStart:operationEnd]

	// Extract Args (if any)
	args := strings.Fields(s[operationEnd+1:])

	return &RepCmd{
		PartitionId: partitionId,
		Origin:      nodeId,
		Timestamp:   timestamp,
		Operation:   operation,
		Args:        args,
		Version:     version,
	}, nil
}

func (rm *RepCmd) Equals(other *RepCmd) bool {
	if other == nil {
		return false
	}
	return rm.PartitionId == other.PartitionId &&
		rm.Origin == other.Origin &&
		rm.Timestamp == other.Timestamp &&
		rm.Operation == other.Operation &&
		reflect.DeepEqual(rm.Args, other.Args) &&
		rm.Version == other.Version
}

func RepCmdFromArgs(args []string) (*RepCmd, error) {
	if len(args) < 6 {
		return nil, fmt.Errorf("invalid input: expected at least 6 arguments, got %d", len(args))
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

	// Extract Version
	version, err := strconv.Atoi(args[3])
	if err != nil {
		return nil, fmt.Errorf("invalid Version: %v", err)
	}

	return &RepCmd{
		PartitionId: partitionId,
		Origin:      args[1],
		Timestamp:   timestamp,
		Operation:   args[4],
		Args:        args[5:],
		Version:     version,
	}, nil
}
