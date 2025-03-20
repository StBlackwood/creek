package handler

import (
	"creek/internal/datastore"
	"creek/internal/logger"
	"errors"
	"strings"
)

// HandleMessage processes incoming messages from clients
func HandleMessage(store *datastore.DataStore, message string) (string, error) {
	// Trim and split input into arguments
	args := strings.Fields(strings.TrimSpace(message))
	if len(args) == 0 {
		return "", errors.New("no command received")
	}

	// Extract command
	command := strings.ToUpper(args[0])

	// Route to appropriate command handler
	return handleCommand(store, command, args)
}

type handlerFunc func(store *datastore.DataStore, args []string) (string, error)

var commandHandlers = map[string]handlerFunc{
	"SET": func(store *datastore.DataStore, args []string) (string, error) { return "OK", handleSet(store, args) },

	"DELETE": func(store *datastore.DataStore, args []string) (string, error) {
		return "OK", handleDelete(store, args)
	},
	"EXPIRE": func(store *datastore.DataStore, args []string) (string, error) {
		return "OK", handleExpire(store, args)
	},
	"TTL": handleTTL,
	"GET": handleGet,
	"VERSION": func(store *datastore.DataStore, args []string) (string, error) {
		return handleVersion()
	},
	"PING": func(store *datastore.DataStore, args []string) (string, error) {
		return "PONG", nil
	},
}

func handleCommand(store *datastore.DataStore, command string, args []string) (string, error) {
	log := logger.GetLogger()
	if handler, exists := commandHandlers[command]; exists {
		return handler(store, args)
	}

	log.Warn("Unknown command received: ", command)
	return "", errors.New("unknown command")
}
