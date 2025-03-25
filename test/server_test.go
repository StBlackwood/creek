package test

import (
	"creek/internal/server"
	"net"
	"testing"
	"time"
)

func TestServer_StartStop(t *testing.T) {
	setupTest(&SimpleServerConfig)
	srv := server.New(&SimpleServerConfig)
	defer cleanupAfterTest(&SimpleServerConfig)
	go srv.Start()
	time.Sleep(1 * time.Second)

	conn, err := net.Dial("tcp", SimpleServerConfig.ServerAddress)
	if err != nil {
		t.Fatalf("Failed to connect to server: %v", err)
	}
	err = conn.Close()
	if err != nil {
		t.Fatalf("Failed to close connection: %v", err)
		return
	}

	srv.Stop()
}
