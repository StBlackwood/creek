package handler

import (
	"creek/internal/core"
	"creek/internal/logger"
	"errors"
	"strings"
)

// HandleMessage processes incoming messages from clients
func HandleMessage(sm *core.StateMachine, message string) (string, error) {
	// Trim and split input into arguments
	args := strings.Fields(strings.TrimSpace(message))
	if len(args) == 0 {
		return "", errors.New("no command received")
	}

	// Extract command
	command := strings.ToUpper(args[0])

	// Route to appropriate command handler
	return handleCommand(sm, command, args)
}

type handlerFunc func(sm *core.StateMachine, args []string) (string, error)

var commandHandlers = map[string]handlerFunc{
	"SET": func(sm *core.StateMachine, args []string) (string, error) { return "OK", handleSet(sm, args) },

	"DELETE": func(sm *core.StateMachine, args []string) (string, error) {
		return "OK", handleDelete(sm, args)
	},
	"EXPIRE": func(sm *core.StateMachine, args []string) (string, error) {
		return "OK", handleExpire(sm, args)
	},
	"TTL": handleTTL,
	"GET": handleGet,
	"VERSION": func(sm *core.StateMachine, args []string) (string, error) {
		return handleVersion()
	},
	"PING": func(sm *core.StateMachine, args []string) (string, error) {
		return "PONG", nil
	},
}

func handleCommand(sm *core.StateMachine, command string, args []string) (string, error) {
	log := logger.GetLogger()
	if handler, exists := commandHandlers[command]; exists {
		return handler(sm, args)
	}

	log.Warn("Unknown command received: ", command)
	return "", errors.New("unknown command")
}
