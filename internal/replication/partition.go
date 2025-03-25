package replication

import (
	"creek/internal/commons"
	"creek/internal/datastore"
	"sync"
)

type Partition struct {
	LW   *LogEntryWriter
	DS   *datastore.DataStore
	Mu   sync.Mutex
	Id   int
	Mode commons.PartitionMode

	WriteChan chan *RepCmd
}
