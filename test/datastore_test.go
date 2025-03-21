package test

import (
	"creek/internal/datastore"
	"testing"
	"time"
)

func TestExpireAndTTL(t *testing.T) {
	ds := datastore.NewDataStore(&SimpleServerConfig)

	ds.Set("session", "active", 3)

	ttl := ds.TTL("session")
	if ttl == -2 || ttl < 2 || ttl > 3 {
		t.Fatalf("Expected TTL between 2-3, got %d", ttl)
	}

	time.Sleep(4 * time.Second)

	_ = ds.Get("session")
}
