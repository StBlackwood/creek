package config

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
)

const DefaultConfigFile = "config/default.conf"
const EnvConfigFile = "CREEK_CONF_FILE"

// Config holds application configuration
type Config struct {
	ServerAddress      string
	LogLevel           string
	PeerNodes          []string
	DataStoreDirectory string
}

var Conf Config

// LoadConfig initializes the configuration from a file
func LoadConfig() error {
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
			return fmt.Errorf("invalid config entry in %s: %s", configFile, err)
		}

		parsedConfig[key] = value
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading config file %s: %w", configFile, err)
	}

	return populateConfig(parsedConfig)
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
func populateConfig(parsedConfig map[string]string) error {
	Conf = Config{
		ServerAddress:      parsedConfig["server_address"],
		LogLevel:           parsedConfig["log_level"],
		DataStoreDirectory: parsedConfig["data_store_directory"],
	}

	if peers, exists := parsedConfig["peer_nodes"]; exists {
		Conf.PeerNodes = strings.Split(peers, ",")
	}

	return validateConfig()
}

// validateConfig checks required configurations and ensures values are valid
func validateConfig() error {
	if Conf.ServerAddress == "" {
		return errors.New("missing required config: server_address")
	}
	if Conf.DataStoreDirectory == "" {
		return errors.New("missing required config: data_store_directory")
	}
	return nil
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
