package server

import (
	"creek/internal/replication"
)

func handleRepCommand(s *Server, args []string) error {
	repCmd, err := replication.RepCmdFromArgs(args[1:])
	if err != nil {
		return err
	}

	err = s.sm.ProcessRepCmd(repCmd)
	return err
}
