package test

import (
	"bufio"
	"creek/internal/server"
	"net"
	"testing"
	"time"
)

func TestServer_BasicReplica(t *testing.T) {
	setupTest(&FollowerServerConfig)
	defer cleanupAfterTest(&FollowerServerConfig)

	followerSrv := server.New(&FollowerServerConfig)
	go followerSrv.Start()
	defer followerSrv.Stop()
	time.Sleep(1 * time.Second)

	setupTest(&LeaderServerConfig)
	defer cleanupAfterTest(&LeaderServerConfig)

	leaderSrv := server.New(&LeaderServerConfig)
	go leaderSrv.Start()
	defer leaderSrv.Stop()
	time.Sleep(1 * time.Second) // Allow server to start

	// connect to leader and set values
	conn, err := net.Dial("tcp", LeaderServerConfig.ServerAddress)
	if err != nil {
		t.Fatalf("Failed to connect to server: %v", err)
	}

	reader := bufio.NewReader(conn)
	_, _ = reader.ReadString('\n') // Discard welcome message

	response, err := sendRequest(conn, "set testkey testvalue")
	if err != nil || response != "OK" {
		t.Errorf("SET command failed: %v, response: %s", err, response)
	}
	time.Sleep(1 * time.Second)

	err = conn.Close()
	if err != nil {
		t.Errorf("Failed to close connection: %v", err)
	}

	// now connect to client and query the values
	conn, err = net.Dial("tcp", FollowerServerConfig.ServerAddress)
	if err != nil {
		t.Fatalf("Failed to connect to server: %v", err)
	}

	reader = bufio.NewReader(conn)
	_, _ = reader.ReadString('\n') // Discard welcome message

	response, err = sendRequest(conn, "get testkey")
	if err != nil || response != "testvalue" {
		t.Errorf("GET command failed: %v, response: %s", err, response)
	}

	err = conn.Close()
	if err != nil {
		return
	}

}
