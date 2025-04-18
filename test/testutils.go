package test

import (
	"bufio"
	"creek/internal/commons"
	"creek/internal/config"
	"net"
	"os"
	"path/filepath"
	"strings"
)

var dataDir = "../data_dir"
var testDataDir = dataDir + "/test"

const hostAddress = "localhost:7690"
const hostAddress1 = "localhost:7691"
const hostAddress2 = "localhost:7692"

var SimpleServerConfig = config.Config{
	ServerAddress:        hostAddress,
	DataStoreDirectory:   testDataDir,
	LogLevel:             "info",
	WriteConsistencyMode: commons.EventualConsistency,
}

var LeaderServerConfig = config.Config{
	ServerAddress:        hostAddress1,
	DataStoreDirectory:   testDataDir + "/leader",
	LogLevel:             "info",
	PeerNodes:            []string{hostAddress2},
	WriteConsistencyMode: commons.EventualConsistency,
	ReplicationMode:      commons.ReadAndWriteReplication,
	ServerMode:           commons.Leader,
}

var FollowerServerConfig = config.Config{
	ServerAddress:        hostAddress2,
	DataStoreDirectory:   testDataDir + "/follower",
	LogLevel:             "info",
	PeerNodes:            []string{hostAddress1},
	WriteConsistencyMode: commons.EventualConsistency,
	ReplicationMode:      commons.ReadOnlyReplication,
	ServerMode:           commons.Follower,
}

func setupTest(conf *config.Config) {
	if _, err := os.Stat(conf.DataStoreDirectory); os.IsNotExist(err) {
		err := os.MkdirAll(conf.DataStoreDirectory, os.ModePerm)
		if err != nil {
			panic(err) // Adjust error handling as needed
		}
	}
}

func cleanupAfterTest(conf *config.Config) {
	dirPath := conf.DataStoreDirectory

	files, err := os.ReadDir(dirPath)
	if err != nil {
		if !os.IsNotExist(err) {
			panic(err) // Adjust error handling as needed
		}
		return
	}

	for _, file := range files {
		err := os.RemoveAll(filepath.Join(dirPath, file.Name()))
		if err != nil {
			panic(err) // Adjust error handling as needed
		}
	}
}

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
