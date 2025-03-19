package handler

import (
	"creek/internal/datastore"
	"creek/internal/version"
	"errors"
	"fmt"
	"strconv"
)

// handleSet stores a key-value pair
func handleSet(ds *datastore.DataStore, args []string) (string, error) {
	if len(args) < 3 {
		return "", errors.New("SET requires a key and a value")
	}

	if len(args) == 3 {
		ds.Set(args[1], args[2], -1)
	} else {
		ttl, err := strconv.Atoi(args[3])
		if err != nil {
			return "", err
		}
		ds.Set(args[1], args[2], ttl)
	}
	return fmt.Sprintf("OK: %s set to %s\n", args[1], args[2]), nil
}

// handleGet retrieves a value by key
func handleGet(ds *datastore.DataStore, args []string) (string, error) {
	if len(args) < 2 {
		return "", errors.New("GET requires a key")
	}
	value, err := ds.Get(args[1])
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("VALUE: %s\n", value), nil
}

// handleDelete removes a key-value pair
func handleDelete(ds *datastore.DataStore, args []string) (string, error) {
	if len(args) < 2 {
		return "", errors.New("DELETE requires a key")
	}
	err := ds.Delete(args[1])
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("OK: %s deleted\n", args[1]), nil
}

// handleVersion returns the server version
func handleVersion() (string, error) {
	return fmt.Sprintf("Server Version: %s\n", version.Version), nil
}
