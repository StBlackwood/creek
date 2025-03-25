package test

import (
	"creek/internal/replication"
	"strings"
	"testing"
)

func TestRepCmdMarshallingString(t *testing.T) {
	testCases := []replication.RepCmd{
		{PartitionId: 1, Origin: "nodeA", Timestamp: 1234567890, Operation: "SET", Args: []string{"key1", "value1", "343"}},
		{PartitionId: 2, Origin: "nodeB", Timestamp: 987654321, Operation: "DELETE", Args: []string{"key2"}},
		{PartitionId: 3, Origin: "nodeC", Timestamp: 1111111111, Operation: "EXPIRE", Args: []string{"key3", "300"}},
	}

	for _, test := range testCases {
		s := test.String()
		output, err := replication.RepCmdFromString(s)
		if err != nil {
			t.Errorf("Failed to parse string: %v", err)
			continue
		}
		cmp := test.Equals(output)
		if !cmp {
			t.Errorf("Expected %v, got %v", test, output)
		}
	}
}

func TestRepCmdMarshallingArgs(t *testing.T) {
	testCases := []replication.RepCmd{
		{PartitionId: 1, Origin: "nodeA", Timestamp: 1234567890, Operation: "SET", Args: []string{"key1", "value1", "343"}},
		{PartitionId: 2, Origin: "nodeB", Timestamp: 987654321, Operation: "DELETE", Args: []string{"key2"}},
		{PartitionId: 3, Origin: "nodeC", Timestamp: 1111111111, Operation: "EXPIRE", Args: []string{"key3", "300"}},
	}

	for _, test := range testCases {
		s := test.String()
		args := strings.Fields(strings.TrimSpace(s))
		output, err := replication.RepCmdFromArgs(args[1:])
		if err != nil {
			t.Errorf("Failed to parse string: %v", err)
			continue
		}
		cmp := test.Equals(output)
		if !cmp {
			t.Errorf("Expected %v, got %v", test, output)
		}
	}
}
