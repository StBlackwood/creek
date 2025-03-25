package commons

const (
	DefaultPort = 7690

	CmdSysPing    = "PING"
	CmdSysPong    = "PONG"
	CmdSysVersion = "VERSION"

	// CmdSysRep prefix of msg signifying it's a replica msg
	CmdSysRep = "REP"

	CmdDataSet = "SET"
	CmdDataGet = "GET"
	CmdDataDel = "DELETE"
	CmdDataTTL = "TTL"
	CmdDataEXP = "EXPIRE"
)
