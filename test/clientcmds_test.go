package test

import (
	"bufio"
	"creek/internal/server"
	"net"
	"strings"
	"testing"
	"time"
)

func sendRequest(conn net.Conn, request string) (string, error) {
	_, err := conn.Write([]byte(request + "\n"))
	if err != nil {
		return "", err
	}

	reader := bufio.NewReader(conn)
	response, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(response), nil
}

func TestServer_Commands(t *testing.T) {
	srv := server.New(&SimpleServerConfig)
	go srv.Start()
	defer srv.Stop()
	time.Sleep(1 * time.Second)

	conn, err := net.Dial("tcp", SimpleServerConfig.ServerAddress)
	if err != nil {
		t.Fatalf("Failed to connect to server: %v", err)
	}
	defer conn.Close()

	// Read and discard the initial welcome message
	reader := bufio.NewReader(conn)
	_, err = reader.ReadString('\n')
	if err != nil {
		t.Fatalf("Failed to read initial server response: %v", err)
	}

	// Test SET and GET
	response, err := sendRequest(conn, "set a b")
	if err != nil || response != "OK" {
		t.Errorf("SET command failed: %v, response: %s", err, response)
	}
	response, err = sendRequest(conn, "get a")
	if err != nil || response != "b" {
		t.Errorf("GET command failed: %v, response: %s", err, response)
	}

	// Test EXPIRE and TTL
	_, _ = sendRequest(conn, "expire a 5")
	response, err = sendRequest(conn, "ttl a")
	if err != nil || response == "" {
		t.Errorf("TTL command failed: %v, response: %s", err, response)
	}
	time.Sleep(6 * time.Second)
	response, err = sendRequest(conn, "get a")
	if err != nil || response != "" {
		t.Errorf("Key should be expired but still exists, response: %s", response)
	}

	// Test DELETE
	_, _ = sendRequest(conn, "set a b")
	response, err = sendRequest(conn, "delete a")
	if err != nil || response != "OK" {
		t.Errorf("DELETE command failed: %v, response: %s", err, response)
	}
	response, err = sendRequest(conn, "get a")
	if err != nil || response != "" {
		t.Errorf("GET should return empty for deleted key, response: %s", response)
	}

	// Test PING
	response, err = sendRequest(conn, "ping")
	if err != nil || response != "PONG" {
		t.Errorf("PING command failed: %v, response: %s", err, response)
	}

	// Test VERSION
	response, err = sendRequest(conn, "version")
	if err != nil || !strings.HasPrefix(response, "1.0.") {
		t.Errorf("VERSION response incorrect: %s", response)
	}

}
