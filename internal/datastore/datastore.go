package datastore

import (
	"creek/internal/config"
	"creek/internal/logger"
	"github.com/sirupsen/logrus"
	"sync"
	"time"
)

// Entry represents a key-value pair with an optional expiration time
type Entry struct {
	Value      string
	Expiration int64 // Unix timestamp, 0 means no expiration
}

// DataStore manages key-value storage with expiration
type DataStore struct {
	data map[string]Entry
	mu   sync.Mutex
	log  *logrus.Logger
	conf *config.Config
}

// NewDataStore initializes a new datastore instance
func NewDataStore(config *config.Config) *DataStore {
	ds := &DataStore{
		data: make(map[string]Entry),
		log:  logger.CreateLogger(config.LogLevel),
		conf: config,
	}
	return ds
}

// CleanExpiredKeys removes expired keys from the datastore
func (ds *DataStore) CleanExpiredKeys() {
	ds.mu.Lock()
	defer ds.mu.Unlock()

	now := time.Now().Unix()
	for key, entry := range ds.data {
		if entry.Expiration > 0 && entry.Expiration <= now {
			delete(ds.data, key)
			ds.log.Trace("GC: Deleted expired key:", key)
		}
	}
}

// GetExpiredKeys gets expired keys as array form the datastore
func (ds *DataStore) GetExpiredKeys() []string {
	ds.mu.Lock()
	defer ds.mu.Unlock()
	var expiredKeys []string
	now := time.Now().Unix()
	for key, entry := range ds.data {
		if entry.Expiration > 0 && entry.Expiration <= now {
			expiredKeys = append(expiredKeys, key)
		}
	}
	return expiredKeys
}

// Stop gracefully shuts down the datastore and stops GC
func (ds *DataStore) Stop() {

	ds.log.Info("Datastore shutdown complete.")
}

// Set stores a key-value pair with an optional expiration time
func (ds *DataStore) Set(key, value string, ttlSeconds int) {
	ds.mu.Lock()
	defer ds.mu.Unlock()
	expiration := int64(0)
	if ttlSeconds > 0 {
		expiration = time.Now().Unix() + int64(ttlSeconds)
	}
	ds.data[key] = Entry{Value: value, Expiration: expiration}
}

// Get retrieves a value by key
func (ds *DataStore) Get(key string) string {
	ds.mu.Lock()
	defer ds.mu.Unlock()
	entry, exists := ds.data[key]
	if !exists {
		return ""
	}
	// Check if key has expired
	if entry.Expiration > 0 && entry.Expiration <= time.Now().Unix() {
		delete(ds.data, key)
		return ""
	}
	return entry.Value
}

// Delete removes a key-value pair
func (ds *DataStore) Delete(key string) {
	ds.mu.Lock()
	defer ds.mu.Unlock()
	delete(ds.data, key)
}

// Expire sets a TTL on an existing key
func (ds *DataStore) Expire(key string, ttlSeconds int) {
	ds.mu.Lock()
	defer ds.mu.Unlock()
	entry, exists := ds.data[key]
	if !exists {
		return
	}
	entry.Expiration = time.Now().Unix() + int64(ttlSeconds)
	ds.data[key] = entry
}

// TTL retrieves the remaining time before a key expires
func (ds *DataStore) TTL(key string) int {
	ds.mu.Lock()
	defer ds.mu.Unlock()
	entry, exists := ds.data[key]
	if !exists {
		return -2
	}
	if entry.Expiration <= 0 {
		return -1
	}
	remaining := entry.Expiration - time.Now().Unix()
	if remaining <= 0 {
		delete(ds.data, key)
		return -2
	}
	return int(remaining)
}
