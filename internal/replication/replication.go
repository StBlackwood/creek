package replication

type RepCmd struct {
	PartitionId int
	NodeId      string
	Timestamp   int64
	Operation   string
	Args        []string
}
