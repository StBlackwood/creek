package commons

type WriteConsistencyMode int

const (
	StrongConsistency WriteConsistencyMode = iota
	EventualConsistency
)

type ReplicaMode int

const (
	ReadOnlyReplication ReplicaMode = iota
	ReadAndWriteReplication
)

type PartitionMode int

const (
	Leader PartitionMode = iota
	Follower
)

func GetConsistencyModeFromString(mode string) WriteConsistencyMode {
	switch mode {
	case "0":
		return StrongConsistency
	case "1":
		return EventualConsistency
	default:
		return EventualConsistency
	}
}

func GetReplicaModeFromString(mode string) ReplicaMode {
	switch mode {
	case "0":
		return ReadOnlyReplication
	case "1":
		return ReadAndWriteReplication
	default:
		return ReadOnlyReplication
	}
}

func GetPartitionModeFromString(mode string) PartitionMode {
	switch mode {
	case "0":
		return Leader
	case "1":
		return Follower
	default:
		return Leader
	}
}
