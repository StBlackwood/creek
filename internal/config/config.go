package config

import "os"

// Config holds server configuration
type Config struct {
	ServerAddress string
	LogLevel      string
}

// LoadConfig loads configuration from environment variables or defaults
func LoadConfig() Config {
	return Config{
		ServerAddress: getEnv("SERVER_ADDRESS", "localhost:8080"),
		LogLevel:      getEnv("LOG_LEVEL", "info"),
	}
}

// getEnv fetches environment variables with a default fallback
func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}
