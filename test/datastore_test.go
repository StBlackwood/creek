package test

import (
	"creek/internal/datastore"
	"testing"
	"time"
)

func TestExpireAndTTL(t *testing.T) {
	ds := datastore.NewDataStore()

	ds.Set("session", "active", 3)

	ttl, err := ds.TTL("session")
	if err != nil || ttl < 2 || ttl > 3 {
		t.Fatalf("Expected TTL between 2-3, got %d", ttl)
	}

	time.Sleep(4 * time.Second)

	_, err = ds.Get("session")
	if err == nil {
		t.Fatalf("Expected key to be expired, but it still exists")
	}
}
