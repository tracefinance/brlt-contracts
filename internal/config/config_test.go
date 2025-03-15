package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadConfig(t *testing.T) {
	// Create a temporary test directory
	tmpDir, err := os.MkdirTemp("", "config-test")
	require.NoError(t, err, "Failed to create temp dir")
	defer os.RemoveAll(tmpDir)

	// Create a test config file
	testConfig := `
# Server configuration
db_path: ${DB_PATH:-./test.db}
port: ${SERVER_PORT:-9090}
key_store_type: ${KEYSTORE_TYPE:-test-store}

# Logging configuration
log:
  level: ${LOG_LEVEL:-debug}
  format: ${LOG_FORMAT:-json}
  request_logging: ${LOG_REQUESTS:-true}
  sql_logging: ${LOG_SQL:-true}

# Blockchain configurations
blockchains:
  ethereum:
    rpc_url: ${ETHEREUM_RPC_URL:-https://test-eth-rpc.com}
    chain_id: ${ETHEREUM_CHAIN_ID:-1}
    default_gas_price: ${ETHEREUM_GAS_PRICE:-25}
    default_gas_limit: ${ETHEREUM_GAS_LIMIT:-21000}
    explorer_url: ${ETHEREUM_EXPLORER_URL:-https://test-etherscan.io}
`

	configPath := filepath.Join(tmpDir, ".config.yaml")
	err = os.WriteFile(configPath, []byte(testConfig), 0644)
	require.NoError(t, err, "Failed to write test config")

	// Set test environment variables
	oldConfigPath := os.Getenv("CONFIG_PATH")
	os.Setenv("CONFIG_PATH", configPath)
	defer os.Setenv("CONFIG_PATH", oldConfigPath)

	oldServerPort := os.Getenv("SERVER_PORT")
	os.Setenv("SERVER_PORT", "8888")
	defer os.Setenv("SERVER_PORT", oldServerPort)

	oldLogLevel := os.Getenv("LOG_LEVEL")
	os.Setenv("LOG_LEVEL", "warn")
	defer os.Setenv("LOG_LEVEL", oldLogLevel)

	// Load configuration
	config, err := LoadConfig()
	require.NoError(t, err, "LoadConfig failed")
	require.NotNil(t, config, "Config should not be nil")

	// Test environment variable override
	assert.Equal(t, "8888", config.Port, "Port should match environment variable")

	// Test environment variable override for nested config
	assert.Equal(t, LogLevel("warn"), config.Log.Level, "Log level should match environment variable")

	// Test default values from YAML
	assert.Equal(t, "./test.db", config.DBPath, "DBPath should match default value")
	assert.Equal(t, "test-store", config.KeyStoreType, "KeyStoreType should match default value")

	// Test nested default values from YAML
	assert.Equal(t, LogFormat("json"), config.Log.Format, "Log format should match default value")

	// Test blockchain configurations
	assert.Equal(t, "https://test-eth-rpc.com", config.Blockchains.Ethereum.RPCURL, "Ethereum RPC URL mismatch")
	assert.Equal(t, uint64(25), config.Blockchains.Ethereum.DefaultGasPrice, "Ethereum gas price mismatch")
}

func TestLoadConfigWithoutYAML(t *testing.T) {
	// Clear any existing config path
	os.Unsetenv("CONFIG_PATH")

	// Set some environment variables
	os.Setenv("SERVER_PORT", "7777")
	os.Setenv("KEYSTORE_TYPE", "env-store")
	os.Setenv("LOG_LEVEL", "error")

	// Load configuration - should fail because no config file exists
	_, err := LoadConfig()
	require.Error(t, err, "LoadConfig should fail when no config file exists")

	// Verify error message
	expectedErrMsg := "no config file found in paths: [ .config.yaml ../.config.yaml]"
	assert.Equal(t, expectedErrMsg, err.Error(), "Unexpected error message")

	// Clean up environment variables
	os.Unsetenv("SERVER_PORT")
	os.Unsetenv("KEYSTORE_TYPE")
	os.Unsetenv("LOG_LEVEL")
}

func TestInterpolateEnvVars(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		env      map[string]string
		expected string
	}{
		{
			name:     "Simple variable",
			content:  "value: ${TEST_VAR}",
			env:      map[string]string{"TEST_VAR": "test"},
			expected: "value: test",
		},
		{
			name:     "Variable with default",
			content:  "value: ${TEST_VAR:-default}",
			env:      map[string]string{},
			expected: "value: default",
		},
		{
			name:     "Variable with empty default",
			content:  "value: ${TEST_VAR:-}",
			env:      map[string]string{},
			expected: "value: ",
		},
		{
			name:     "Override default value",
			content:  "value: ${TEST_VAR:-default}",
			env:      map[string]string{"TEST_VAR": "override"},
			expected: "value: override",
		},
		{
			name:     "Multiple variables",
			content:  "first: ${FIRST_VAR:-one} second: ${SECOND_VAR:-two}",
			env:      map[string]string{"FIRST_VAR": "1", "SECOND_VAR": "2"},
			expected: "first: 1 second: 2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variables
			for k, v := range tt.env {
				os.Setenv(k, v)
			}
			// Clean up environment after test
			defer func() {
				for k := range tt.env {
					os.Unsetenv(k)
				}
			}()

			result := interpolateEnvVars(tt.content)
			assert.Equal(t, tt.expected, result, "Interpolation result mismatch")
		})
	}
}
