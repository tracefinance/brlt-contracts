package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	// Create a temporary test directory
	tmpDir, err := os.MkdirTemp("", "config-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
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

# Token configurations
tokens:
  ethereum:
    - name: ETH
      type: native
      chain_type: ethereum
      decimals: 18
    - name: USDC
      type: erc20
      chain_type: ethereum
      address: "0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48"
      decimals: 6

  polygon:
    - name: MATIC
      type: native
      chain_type: polygon
      decimals: 18

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
	if err := os.WriteFile(configPath, []byte(testConfig), 0644); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

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
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	// Test environment variable override
	if config.Port != "8888" {
		t.Errorf("Expected port 8888, got %s", config.Port)
	}

	// Test environment variable override for nested config
	if config.Log.Level != "warn" {
		t.Errorf("Expected log level warn, got %s", config.Log.Level)
	}

	// Test default values from YAML
	if config.DBPath != "./test.db" {
		t.Errorf("Expected DBPath ./test.db, got %s", config.DBPath)
	}

	if config.KeyStoreType != "test-store" {
		t.Errorf("Expected KeyStoreType test-store, got %s", config.KeyStoreType)
	}

	// Test nested default values from YAML
	if config.Log.Format != "json" {
		t.Errorf("Expected log format json, got %s", config.Log.Format)
	}

	// Test token configurations
	// Test Ethereum tokens
	ethTokens := config.Tokens.Ethereum
	if len(ethTokens) != 2 {
		t.Errorf("Expected 2 Ethereum tokens, got %d", len(ethTokens))
	}

	// Test ETH token
	eth := ethTokens[0]
	if eth.Name != "ETH" || eth.Type != TokenTypeNative || eth.ChainType != ChainTypeEthereum || eth.Decimals != 18 {
		t.Errorf("Unexpected ETH token configuration: %+v", eth)
	}

	// Test USDC token
	usdc := ethTokens[1]
	if usdc.Name != "USDC" || usdc.Type != TokenTypeERC20 || usdc.ChainType != ChainTypeEthereum ||
		usdc.Address != "0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48" || usdc.Decimals != 6 {
		t.Errorf("Unexpected USDC token configuration: %+v", usdc)
	}

	// Test Polygon tokens
	polygonTokens := config.Tokens.Polygon
	if len(polygonTokens) != 1 {
		t.Errorf("Expected 1 Polygon token, got %d", len(polygonTokens))
	}

	// Test MATIC token
	matic := polygonTokens[0]
	if matic.Name != "MATIC" || matic.Type != TokenTypeNative || matic.ChainType != ChainTypePolygon || matic.Decimals != 18 {
		t.Errorf("Unexpected MATIC token configuration: %+v", matic)
	}

	if config.Blockchains.Ethereum.RPCURL != "https://test-eth-rpc.com" {
		t.Errorf("Expected Ethereum RPC URL https://test-eth-rpc.com, got %s", config.Blockchains.Ethereum.RPCURL)
	}

	if config.Blockchains.Ethereum.DefaultGasPrice != 25 {
		t.Errorf("Expected Ethereum gas price 25, got %d", config.Blockchains.Ethereum.DefaultGasPrice)
	}
}

func TestLoadConfigWithoutYAML(t *testing.T) {
	// Clear any existing config path
	os.Unsetenv("CONFIG_PATH")

	// Set some environment variables
	os.Setenv("SERVER_PORT", "7777")
	os.Setenv("KEYSTORE_TYPE", "env-store")
	os.Setenv("LOG_LEVEL", "error")

	// Load configuration
	config, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	// Test environment variables
	if config.Port != "7777" {
		t.Errorf("Expected port 7777, got %s", config.Port)
	}

	if config.KeyStoreType != "env-store" {
		t.Errorf("Expected KeyStoreType env-store, got %s", config.KeyStoreType)
	}

	// Test default values
	if config.Log.Format != "console" {
		t.Errorf("Expected default log format console, got %s", config.Log.Format)
	}

	if config.Blockchains.Ethereum.RPCURL != "https://ethereum-rpc.publicnode.com" {
		t.Errorf("Expected default Ethereum RPC URL, got %s", config.Blockchains.Ethereum.RPCURL)
	}

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
			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}
