package handler

import (
	"creek/internal/datastore"
	"creek/internal/logger"
	"errors"
	"strings"
)

// HandleMessage processes incoming messages from clients
func HandleMessage(store *datastore.DataStore, message string) (string, error) {
	log := logger.GetLogger()

	// Trim and split input into arguments
	args := strings.Fields(strings.TrimSpace(message))
	if len(args) == 0 {
		return "", errors.New("no command received")
	}

	// Extract command
	command := strings.ToUpper(args[0])

	// Route to appropriate command handler
	switch command {
	case "SET":
		return handleSet(store, args)
	case "GET":
		return handleGet(store, args)
	case "DELETE":
		return handleDelete(store, args)
	case "EXPIRE":
		return handleExpire(store, args)
	case "TTL":
		return handleTTL(store, args)
	case "PING":
		return "PONG\n", nil
	case "VERSION":
		return handleVersion()
	default:
		log.Warn("Unknown command received: ", message)
		return "", errors.New("unknown command")
	}
}
