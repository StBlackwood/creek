package handler

import (
	"creek/internal/version"
	"errors"
	"fmt"
	"sync"
)

var (
	dataStore = make(map[string]string)
	mu        sync.Mutex
)

func handleSet(args []string) (string, error) {
	if len(args) < 3 {
		return "", errors.New("SET requires a key and a value")
	}
	mu.Lock()
	dataStore[args[1]] = args[2]
	mu.Unlock()
	return fmt.Sprintf("OK: %s set to %s\n", args[1], args[2]), nil
}

// handleGet retrieves a value by key
func handleGet(args []string) (string, error) {
	if len(args) < 2 {
		return "", errors.New("GET requires a key")
	}
	mu.Lock()
	value, exists := dataStore[args[1]]
	mu.Unlock()
	if !exists {
		return "", errors.New("key not found")
	}
	return fmt.Sprintf("VALUE: %s\n", value), nil
}

// handleDelete removes a key-value pair
func handleDelete(args []string) (string, error) {
	if len(args) < 2 {
		return "", errors.New("DELETE requires a key")
	}
	mu.Lock()
	_, exists := dataStore[args[1]]
	if exists {
		delete(dataStore, args[1])
	}
	mu.Unlock()
	if !exists {
		return "", errors.New("key not found")
	}
	return fmt.Sprintf("OK: %s deleted\n", args[1]), nil
}

// handleVersion returns the server version
func handleVersion() (string, error) {
	return fmt.Sprintf("Server Version: %s\n", version.Version), nil
}
