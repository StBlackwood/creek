package handler

import (
	"creek/internal/commons"
	"creek/internal/core"
	"errors"
	"strconv"
)

// handleSet stores a key-value pair
func handleSet(sm *core.StateMachine, args []string) error {
	if len(args) < 3 {
		return errors.New("SET requires a key and a value")
	}

	ttl := -1
	if len(args) > 3 {
		ttlParsed, err := strconv.Atoi(args[3])
		if err != nil {
			return errors.New("invalid TTL value")
		}
		ttl = ttlParsed
	}
	return sm.Set(args[1], args[2], ttl)
}

// handleGet retrieves a value by key
func handleGet(sm *core.StateMachine, args []string) (string, error) {
	if len(args) < 2 {
		return "", errors.New("GET requires a key")
	}
	value, err := sm.Get(args[1])
	if err != nil {
		return "", nil // most likely coz of key not present
	}
	return value, nil
}

// handleDelete removes a key-value pair
func handleDelete(sm *core.StateMachine, args []string) error {
	if len(args) < 2 {
		return errors.New("DELETE requires a key")
	}
	err := sm.Delete(args[1])
	if err != nil {
		return err
	}
	return nil
}

// handleVersion returns the server commons
func handleVersion() (string, error) {
	return commons.Version, nil
}

// handleExpire sets a TTL on an existing key
func handleExpire(sm *core.StateMachine, args []string) error {
	if len(args) < 3 {
		return errors.New("EXPIRE requires a key and TTL")
	}
	ttl, err := strconv.Atoi(args[2])
	if err != nil {
		return errors.New("invalid TTL value")
	}
	err = sm.Expire(args[1], ttl)
	if err != nil {
		return err
	}
	return nil
}

// handleTTL retrieves the TTL for a key
func handleTTL(sm *core.StateMachine, args []string) (string, error) {
	if len(args) < 2 {
		return "", errors.New("TTL requires a key")
	}
	ttl, err := sm.TTL(args[1])
	if err != nil {
		return "", err
	}
	return strconv.Itoa(ttl), nil
}
