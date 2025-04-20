package config

import (
	"creek/internal/commons"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	// Create a temporary directory for test files
	tmpDir, err := os.MkdirTemp("", "creek-config-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a test data directory
	testDataDir := filepath.Join(tmpDir, "data")
	if err := os.MkdirAll(testDataDir, 0755); err != nil {
		t.Fatalf("Failed to create test data directory: %v", err)
	}

	testCases := []struct {
		name           string
		configContent  string
		expectedError  bool
		expectedConfig Config
	}{
		{
			name: "Valid complete config",
			configContent: `
server_address=localhost:8080
log_level=DEBUG
data_store_directory=` + testDataDir + `
peer_nodes=localhost:8081,localhost:8082
write_consistency_mode=0
replication_mode=1
server_mode=0
`,
			expectedError: false,
			expectedConfig: Config{
				ServerAddress:        "localhost:8080",
				LogLevel:             "DEBUG",
				DataStoreDirectory:   testDataDir,
				PeerNodes:            []string{"localhost:8081", "localhost:8082"},
				WriteConsistencyMode: commons.StrongConsistency,
				ReplicationMode:      commons.ReadAndWriteReplication,
				ServerMode:           commons.Leader,
			},
		},
		{
			name: "Minimal valid config",
			configContent: `
server_address=localhost:8080
data_store_directory=` + testDataDir + `
write_consistency_mode=0
replication_mode=0
server_mode=0
`,
			expectedError: false,
			expectedConfig: Config{
				ServerAddress:        "localhost:8080",
				DataStoreDirectory:   testDataDir,
				WriteConsistencyMode: commons.StrongConsistency,
				ReplicationMode:      commons.ReadOnlyReplication,
				ServerMode:           commons.Leader,
			},
		},
		{
			name: "Invalid config - missing server address",
			configContent: `
			log_level=DEBUG
data_store_directory=` + testDataDir + `
server_address=
`,
			expectedError: true,
		},
		{
			name: "Invalid config - invalid data directory",
			configContent: `
server_address=localhost:8080
data_store_directory=/nonexistent/directory
`,
			expectedError: true,
		},
		{
			name: "Invalid config - follower with write replication",
			configContent: `
server_address=localhost:8080
data_store_directory=` + testDataDir + `
replication_mode=1
server_mode=1
`,
			expectedError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create temporary config file
			configFile := filepath.Join(tmpDir, "test.conf")
			err := os.WriteFile(configFile, []byte(tc.configContent), 0644)
			if err != nil {
				t.Fatalf("Failed to write test config file: %v", err)
			}

			// Set environment variable to point to our test config
			os.Setenv(EnvConfigFile, configFile)
			defer os.Unsetenv(EnvConfigFile)

			// Load the config
			config, err := LoadConfig()

			// Check error expectation
			if tc.expectedError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tc.expectedError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			// If we expect success, verify the config values
			if !tc.expectedError {
				if config.ServerAddress != tc.expectedConfig.ServerAddress {
					t.Errorf("ServerAddress mismatch: got %v, want %v",
						config.ServerAddress, tc.expectedConfig.ServerAddress)
				}
				if config.DataStoreDirectory != tc.expectedConfig.DataStoreDirectory {
					t.Errorf("DataStoreDirectory mismatch: got %v, want %v",
						config.DataStoreDirectory, tc.expectedConfig.DataStoreDirectory)
				}
				if config.LogLevel != tc.expectedConfig.LogLevel {
					t.Errorf("LogLevel mismatch: got %v, want %v",
						config.LogLevel, tc.expectedConfig.LogLevel)
				}
				if len(config.PeerNodes) != len(tc.expectedConfig.PeerNodes) {
					t.Errorf("PeerNodes length mismatch: got %v, want %v",
						len(config.PeerNodes), len(tc.expectedConfig.PeerNodes))
				}
				if config.WriteConsistencyMode != tc.expectedConfig.WriteConsistencyMode {
					t.Errorf("WriteConsistencyMode mismatch: got %v, want %v",
						config.WriteConsistencyMode, tc.expectedConfig.WriteConsistencyMode)
				}
				if config.ReplicationMode != tc.expectedConfig.ReplicationMode {
					t.Errorf("ReplicationMode mismatch: got %v, want %v",
						config.ReplicationMode, tc.expectedConfig.ReplicationMode)
				}
				if config.ServerMode != tc.expectedConfig.ServerMode {
					t.Errorf("ServerMode mismatch: got %v, want %v",
						config.ServerMode, tc.expectedConfig.ServerMode)
				}
			}
		})
	}
}

func TestParseConfigLine(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expectedKey   string
		expectedValue string
		expectError   bool
	}{
		{
			name:          "Valid config line",
			input:         "server_address=localhost:8080",
			expectedKey:   "server_address",
			expectedValue: "localhost:8080",
			expectError:   false,
		},
		{
			name:        "Missing value",
			input:       "server_address=",
			expectError: true,
		},
		{
			name:        "Missing key",
			input:       "=localhost:8080",
			expectError: true,
		},
		{
			name:        "Invalid format",
			input:       "invalid_line",
			expectError: true,
		},
		{
			name:          "Trimmed spaces",
			input:         "  log_level  =  DEBUG  ",
			expectedKey:   "log_level",
			expectedValue: "DEBUG",
			expectError:   false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			key, value, err := parseConfigLine(tc.input)

			if tc.expectError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tc.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			if !tc.expectError {
				if key != tc.expectedKey {
					t.Errorf("Key mismatch: got %v, want %v", key, tc.expectedKey)
				}
				if value != tc.expectedValue {
					t.Errorf("Value mismatch: got %v, want %v", value, tc.expectedValue)
				}
			}
		})
	}
}

func TestRemoveBOM(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "String with BOM",
			input:    "\ufefftest",
			expected: "test",
		},
		{
			name:     "String without BOM",
			input:    "test",
			expected: "test",
		},
		{
			name:     "Empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "Multiple BOM",
			input:    "\ufeff\ufefftest",
			expected: "\ufefftest",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := removeBOM(tc.input)
			if result != tc.expected {
				t.Errorf("Expected %q but got %q", tc.expected, result)
			}
		})
	}
}

func TestGetEnv(t *testing.T) {
	tests := []struct {
		name     string
		key      string
		envValue string
		fallback string
		expected string
	}{
		{
			name:     "Environment variable exists",
			key:      "TEST_KEY",
			envValue: "test_value",
			fallback: "fallback_value",
			expected: "test_value",
		},
		{
			name:     "Environment variable doesn't exist",
			key:      "NONEXISTENT_KEY",
			envValue: "",
			fallback: "fallback_value",
			expected: "fallback_value",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if tc.envValue != "" {
				os.Setenv(tc.key, tc.envValue)
				defer os.Unsetenv(tc.key)
			}

			result := getEnv(tc.key, tc.fallback)
			if result != tc.expected {
				t.Errorf("Expected %s but got %s", tc.expected, result)
			}
		})
	}
}

func TestLoadConfigFromEnvironmentVariable(t *testing.T) {
	// Create temporary directories
	tmpDir, err := os.MkdirTemp("", "creek-config-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create two different data directories for different configs
	defaultDataDir := filepath.Join(tmpDir, "default_data")
	envDataDir := filepath.Join(tmpDir, "env_data")

	for _, dir := range []string{defaultDataDir, envDataDir} {
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("Failed to create data directory %s: %v", dir, err)
		}
	}

	// Create default config file
	defaultConfigContent := fmt.Sprintf(`
server_address=localhost:8080
log_level=INFO
data_store_directory=%s
peer_nodes=localhost:8081
write_consistency_mode=0
replication_mode=1
server_mode=0
`, defaultDataDir)

	defaultConfigPath := filepath.Join(tmpDir, "default.conf")
	if err := os.WriteFile(defaultConfigPath, []byte(defaultConfigContent), 0644); err != nil {
		t.Fatalf("Failed to write default config file: %v", err)
	}

	// Create environment-specific config file
	envConfigContent := fmt.Sprintf(`
server_address=localhost:9090
log_level=DEBUG
data_store_directory=%s
peer_nodes=localhost:9091,localhost:9092
write_consistency_mode=0
replication_mode=1
server_mode=0
`, envDataDir)

	envConfigPath := filepath.Join(tmpDir, "env.conf")
	if err := os.WriteFile(envConfigPath, []byte(envConfigContent), 0644); err != nil {
		t.Fatalf("Failed to write env config file: %v", err)
	}

	// Store original value of environment variable
	originalEnvValue := os.Getenv(EnvConfigFile)
	defer os.Setenv(EnvConfigFile, originalEnvValue)

	// Test both configurations
	testConfigs := []struct {
		name     string
		envPath  string
		expected Config
	}{
		{
			name:    "Default configuration",
			envPath: defaultConfigPath,
			expected: Config{
				ServerAddress:      "localhost:8080",
				LogLevel:           "INFO",
				DataStoreDirectory: defaultDataDir,
				PeerNodes:          []string{"localhost:8081"},
			},
		},
		{
			name:    "Environment configuration",
			envPath: envConfigPath,
			expected: Config{
				ServerAddress:      "localhost:9090",
				LogLevel:           "DEBUG",
				DataStoreDirectory: envDataDir,
				PeerNodes:          []string{"localhost:9091", "localhost:9092"},
			},
		},
	}

	for _, tc := range testConfigs {
		t.Run(tc.name, func(t *testing.T) {
			os.Setenv(EnvConfigFile, tc.envPath)

			config, err := LoadConfig()
			if err != nil {
				t.Fatalf("Failed to load config: %v", err)
			}

			// Test basic string fields
			if config.ServerAddress != tc.expected.ServerAddress {
				t.Errorf("ServerAddress = %v, want %v",
					config.ServerAddress, tc.expected.ServerAddress)
			}
			if config.LogLevel != tc.expected.LogLevel {
				t.Errorf("LogLevel = %v, want %v",
					config.LogLevel, tc.expected.LogLevel)
			}
			if config.DataStoreDirectory != tc.expected.DataStoreDirectory {
				t.Errorf("DataStoreDirectory = %v, want %v",
					config.DataStoreDirectory, tc.expected.DataStoreDirectory)
			}
			if !slicesEqual(config.PeerNodes, tc.expected.PeerNodes) {
				t.Errorf("PeerNodes = %v, want %v",
					config.PeerNodes, tc.expected.PeerNodes)
			}

			// Note: We're not testing the enum values yet since we need to see
			// how they're defined in the commons package
		})
	}
}

// Helper function to compare string slices
func slicesEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
