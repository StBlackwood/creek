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

var SimpleServerConfig = config.Config{
	ServerAddress:        "localhost:9090",
	DataStoreDirectory:   "../data_dir/test",
	LogLevel:             "info",
	WriteConsistencyMode: commons.EventualConsistency,
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
