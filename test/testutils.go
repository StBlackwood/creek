package test

import (
	"creek/internal/commons"
	"creek/internal/config"
	"os"
	"path/filepath"
)

var SimpleServerConfig = config.Config{
	ServerAddress:        "localhost:9090",
	DataStoreDirectory:   "../data_dir/test",
	LogLevel:             "info",
	WriteConsistencyMode: commons.EventualConsistency,
}

func CleanupAfterTest(conf *config.Config) {
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
