package config

import (
	"bufio"
	"creek/internal/commons"
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
)

const DefaultConfigFile = "config/default.conf"
const EnvConfigFile = "CREEK_CONF_FILE"

// Config holds application configuration
type Config struct {
	ServerAddress        string
	LogLevel             string
	PeerNodes            []string
	DataStoreDirectory   string
	WriteConsistencyMode commons.WriteConsistencyMode
	ReplicationMode      commons.ReplicaMode
	ServerMode           commons.PartitionMode // For now, in future this config will be removed once data partition is introduced
}

// LoadConfig initializes the configuration from a file
func LoadConfig() (*Config, error) {
	configFile := getEnv(EnvConfigFile, DefaultConfigFile)

	file, err := os.Open(configFile)
	if err != nil {
		log.Fatalf("failed to open config file %s: %v", configFile, err)
	}

	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			_ = fmt.Errorf("failed to close config file %s: %w", configFile, err)
		}
	}(file)

	scanner := bufio.NewScanner(file)
	parsedConfig := make(map[string]string)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		// Remove UTF-8 BOM if present
		line = removeBOM(line)
		if len(line) == 0 || strings.HasPrefix(line, "#") {
			continue // Skip empty lines and comments
		}

		key, value, err := parseConfigLine(line)
		if err != nil {
			return nil, fmt.Errorf("invalid config entry in %s: %s", configFile, err)
		}

		parsedConfig[key] = value
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading config file %s: %w", configFile, err)
	}

	var nodes []string
	if val, exists := parsedConfig["peer_nodes"]; exists {
		nodes = strings.Split(val, ",")
	}

	conf := Config{
		ServerAddress:      parsedConfig["server_address"],
		LogLevel:           parsedConfig["log_level"],
		DataStoreDirectory: parsedConfig["data_store_directory"],
		PeerNodes:          nodes,
		WriteConsistencyMode: commons.GetConsistencyModeFromString(
			parsedConfig["write_consistency_mode"],
		),
		ReplicationMode: commons.GetReplicaModeFromString(
			parsedConfig["replication_mode"],
		),
		ServerMode: commons.GetPartitionModeFromString(
			parsedConfig["server_mode"],
		),
	}
	err = conf.populateConfig(parsedConfig)
	return &conf, err
}

// parseConfigLine parses a key=value pair from a line
func parseConfigLine(line string) (string, string, error) {
	parts := strings.SplitN(line, "=", 2)
	if len(parts) != 2 {
		return "", "", errors.New(fmt.Sprintf("invalid config entry: %s", line))
	}
	key := strings.TrimSpace(parts[0])
	value := strings.TrimSpace(parts[1])
	if key == "" || value == "" {
		return "", "", errors.New("empty key or value")
	}
	return key, value, nil
}

// populateConfig maps parsed values to the Conf struct and validates them
func (conf *Config) populateConfig(parsedConfig map[string]string) error {
	if peers, exists := parsedConfig["peer_nodes"]; exists {
		conf.PeerNodes = strings.Split(peers, ",")
	}

	conf.fillUpDefaults()
	return conf.validateConfig()
}

func (conf *Config) fillUpDefaults() {
	if conf.ServerAddress == "" {
		conf.ServerAddress = "localhost:" + strconv.Itoa(commons.DefaultPort)
	}

	if conf.DataStoreDirectory == "" {
		conf.DataStoreDirectory = "data_dir"
		if !isDirPathExists(conf.DataStoreDirectory) {
			err := os.MkdirAll(conf.DataStoreDirectory, os.ModePerm)
			if err != nil {
				panic(err) // Adjust error handling as needed
			}
		}
	}
}

// validateConfig checks required configurations and ensures values are valid
func (conf *Config) validateConfig() error {
	if conf.ServerAddress == "" {
		return errors.New("missing required config: server_address")
	}
	if conf.ReplicationMode == commons.ReadAndWriteReplication && conf.ServerMode == commons.Follower {
		return errors.New("followers cant accept writes right now")
	}

	if conf.DataStoreDirectory == "" {
		return errors.New("missing required config: data_store_directory")
	}
	isDirExists := isDirPathExists(conf.DataStoreDirectory)
	if !isDirExists {
		return fmt.Errorf("invalid data_store_directory: %s", conf.DataStoreDirectory)
	}
	return nil
}

func isDirPathExists(dir string) bool {
	info, err := os.Stat(dir)
	if err != nil || !info.IsDir() {
		return false
	}
	return true
}

// getEnv fetches environment variables with a default fallback
func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}

// removeBOM removes UTF-8 Byte Order Mark if present
func removeBOM(line string) string {
	bom := "\ufeff"
	if strings.HasPrefix(line, bom) {
		return strings.TrimPrefix(line, bom)
	}
	return line
}
