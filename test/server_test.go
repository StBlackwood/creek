package test

import (
	"creek/internal/server"
	"net"
	"testing"
	"time"
)

func TestServer_StartStop(t *testing.T) {
	srv := server.New(":9090")
	go srv.Start()
	time.Sleep(1 * time.Second)

	conn, err := net.Dial("tcp", ":9090")
	if err != nil {
		t.Fatalf("Failed to connect to server: %v", err)
	}
	conn.Close()

	srv.Stop()
}
