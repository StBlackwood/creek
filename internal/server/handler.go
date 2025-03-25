package server

import (
	"creek/internal/commons"
	"creek/internal/core"
	"creek/internal/logger"
	"errors"
	"strings"
)

// handleMessage processes incoming messages from clients
func handleMessage(s *Server, message string) (string, error) {
	// Trim and split input into arguments
	args := strings.Fields(strings.TrimSpace(message))
	if len(args) == 0 {
		return "", errors.New("no command received")
	}

	// Extract command
	command := strings.ToUpper(args[0])

	// Route to appropriate command handler
	return handleCommand(s, command, args)
}

type handlerFunc func(sm *core.StateMachine, args []string) (string, error)
type systemCommandHandlerFunc func(s *Server, args []string) (string, error)

var commandHandlers = map[string]handlerFunc{
	commons.CmdDataSet: func(sm *core.StateMachine, args []string) (string, error) { return "OK", handleSet(sm, args) },

	commons.CmdDataDel: func(sm *core.StateMachine, args []string) (string, error) {
		return "OK", handleDelete(sm, args)
	},
	commons.CmdDataEXP: func(sm *core.StateMachine, args []string) (string, error) {
		return "OK", handleExpire(sm, args)
	},
	commons.CmdDataTTL: handleTTL,
	commons.CmdDataGet: handleGet,
}

var systemCommandHandlers = map[string]systemCommandHandlerFunc{
	"SHUTDOWN": func(s *Server, args []string) (string, error) {
		defer s.Stop()
		return "OK", nil
	},

	commons.CmdSysRep: func(s *Server, args []string) (string, error) {
		return "OK", handleRepCommand(s, args)
	},

	commons.CmdSysVersion: func(s *Server, args []string) (string, error) {
		return handleVersion()
	},
	commons.CmdSysPing: func(s *Server, args []string) (string, error) {
		return commons.CmdSysPong, nil
	},
}

func handleCommand(s *Server, command string, args []string) (string, error) {
	log := logger.GetLogger()
	if handler, exists := commandHandlers[command]; exists {
		return handler(s.sm, args)
	}

	if handler, exists := systemCommandHandlers[command]; exists {
		return handler(s, args)
	}

	log.Warn("Unknown command received: ", command)
	return "", errors.New("unknown command")
}
