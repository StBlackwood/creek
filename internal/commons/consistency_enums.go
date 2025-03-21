package commons

type WriteConsistencyMode int

const (
	StrongConsistency WriteConsistencyMode = iota
	EventualConsistency
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
