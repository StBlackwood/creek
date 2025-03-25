package test

import (
	"bufio"
	"creek/internal/server"
	"net"
	"strconv"
	"testing"
	"time"
)

func TestServer_Recovery(t *testing.T) {
	setupTest(&SimpleServerConfig)
	defer cleanupAfterTest(&SimpleServerConfig)
	srv := server.New(&SimpleServerConfig)
	go srv.Start()
	time.Sleep(1 * time.Second) // Allow server to start

	conn, err := net.Dial("tcp", SimpleServerConfig.ServerAddress)
	if err != nil {
		t.Fatalf("Failed to connect to server: %v", err)
	}

	reader := bufio.NewReader(conn)
	_, _ = reader.ReadString('\n') // Discard welcome message

	// Set a key before stopping the server
	response, err := sendRequest(conn, "set testkey testvalue")
	if err != nil || response != "OK" {
		t.Errorf("SET command failed: %v, response: %s", err, response)
	}

	err = conn.Close()
	if err != nil {
		t.Errorf("Failed to close connection: %v", err)
	}

	// Stop the server
	srv.Stop()
	time.Sleep(1 * time.Second) // Ensure clean shutdown

	// Restart the server
	srv = server.New(&SimpleServerConfig)
	go srv.Start()
	defer srv.Stop()
	time.Sleep(1 * time.Second) // Allow recovery

	// Reconnect to the server
	conn, err = net.Dial("tcp", SimpleServerConfig.ServerAddress)
	if err != nil {
		t.Fatalf("Failed to reconnect to server: %v", err)
	}
	defer func(conn net.Conn) {
		err := conn.Close()
		if err != nil {
			t.Errorf("Failed to close connection: %v", err)
		}
	}(conn)

	reader = bufio.NewReader(conn)
	_, _ = reader.ReadString('\n') // Discard welcome message

	// Verify key is still present after restart
	response, err = sendRequest(conn, "get testkey")
	if err != nil || response != "testvalue" {
		t.Errorf("Recovery failed: expected 'testvalue', got '%s'", response)
	}
}

func TestServer_RecoveryWithTTL(t *testing.T) {
	defer cleanupAfterTest(&SimpleServerConfig)
	srv := server.New(&SimpleServerConfig)
	go srv.Start()
	time.Sleep(1 * time.Second) // Allow server to start

	conn, err := net.Dial("tcp", SimpleServerConfig.ServerAddress)
	if err != nil {
		t.Fatalf("Failed to connect to server: %v", err)
	}

	reader := bufio.NewReader(conn)
	_, _ = reader.ReadString('\n') // Discard welcome message

	// Set keys with different TTLs
	_, _ = sendRequest(conn, "set key1 value1")
	_, _ = sendRequest(conn, "expire key1 3")
	_, _ = sendRequest(conn, "set key2 value2 20")

	err = conn.Close()
	if err != nil {
		t.Errorf("Failed to close connection: %v", err)
		return
	}

	// Stop the server and wait 5 seconds
	srv.Stop()
	time.Sleep(5 * time.Second)

	// Restart the server
	srv = server.New(&SimpleServerConfig)
	go srv.Start()
	defer srv.Stop()
	time.Sleep(1 * time.Second) // Allow recovery

	// Reconnect to the server
	conn, err = net.Dial("tcp", SimpleServerConfig.ServerAddress)
	if err != nil {
		t.Fatalf("Failed to reconnect to server: %v", err)
	}
	defer func(conn net.Conn) {
		err := conn.Close()
		if err != nil {
			t.Errorf("Failed to close connection: %v", err)
		}
	}(conn)

	reader = bufio.NewReader(conn)
	_, _ = reader.ReadString('\n') // Discard welcome message

	// Verify key1 is expired and no longer present
	response, err := sendRequest(conn, "get key1")
	if err != nil || response != "" {
		t.Errorf("Key1 should have expired but still exists, response: %s", response)
	}

	// Verify key2 still exists with reduced TTL
	response, err = sendRequest(conn, "ttl key2")
	expectedTTL := 15 // Original TTL 20 - 5 seconds wait time before shutdown
	if err != nil {
		t.Errorf("Failed to get TTL for key2: %v", err)
	} else {
		ttl, err := strconv.Atoi(response)
		if err != nil || ttl < expectedTTL-1 || ttl > expectedTTL+1 { // Allow slight delay variation
			t.Errorf("Incorrect TTL for key2 after recovery, expected around %d, got %d", expectedTTL, ttl)
		}
	}
}
