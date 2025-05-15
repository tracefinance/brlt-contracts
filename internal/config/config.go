package config

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/joho/godotenv"
	"gopkg.in/yaml.v3"
)

// LogLevel represents the logging level
type LogLevel string

const (
	// LogLevelDebug represents debug level logging
	LogLevelDebug LogLevel = "debug"
	// LogLevelInfo represents info level logging
	LogLevelInfo LogLevel = "info"
	// LogLevelWarn represents warn level logging
	LogLevelWarn LogLevel = "warn"
	// LogLevelError represents error level logging
	LogLevelError LogLevel = "error"
)

// LogFormat represents the logging output format
type LogFormat string

const (
	// LogFormatJSON represents JSON format logging
	LogFormatJSON LogFormat = "json"
	// LogFormatConsole represents human-readable console format logging
	LogFormatConsole LogFormat = "console"
)

// LogConfig holds configuration for application logging
type LogConfig struct {
	// Level is the minimum log level to output
	Level LogLevel `yaml:"level"`
	// Format is the log output format (json or console)
	Format LogFormat `yaml:"format"`
	// OutputPath is the path to the log file (empty for stdout)
	OutputPath string `yaml:"output_path"`
	// RequestLogging enables HTTP request logging when true
	RequestLogging bool `yaml:"request_logging"`
	// SQLLogging enables SQL query logging when true
	SQLLogging bool `yaml:"sql_logging"`
}

// BlockchainConfig holds configuration for a specific blockchain
type BlockchainConfig struct {
	// RPCURL is the RPC URL for the blockchain
	RPCURL string `yaml:"rpc_url"`
	// ChainID is the chain ID for the blockchain
	ChainID int64 `yaml:"chain_id"`
	// DefaultGasPrice is the default gas price for transactions in Gwei
	DefaultGasPrice uint64 `yaml:"default_gas_price"`
	// DefaultGasLimit is the default gas limit for transactions
	DefaultGasLimit uint64 `yaml:"default_gas_limit"`
	// ExplorerURL is the block explorer URL for the blockchain
	ExplorerURL string `yaml:"explorer_url"`
	// ExplorerAPIURL is the block explorer API URL for the blockchain
	ExplorerAPIURL string `yaml:"explorer_api_url"`
	// ExplorerAPIKey is the API key for the block explorer
	ExplorerAPIKey string `yaml:"explorer_api_key"`
}

// BlockchainsConfig holds configuration for all supported blockchains
type BlockchainsConfig struct {
	// Ethereum holds Ethereum blockchain configuration
	Ethereum BlockchainConfig `yaml:"ethereum"`
	// Polygon holds Polygon blockchain configuration
	Polygon BlockchainConfig `yaml:"polygon"`
	// Base holds Base blockchain configuration
	Base BlockchainConfig `yaml:"base"`
}

// SnowflakeConfig holds configuration for Twitter Snowflake ID generation
type SnowflakeConfig struct {
	// DataCenterID is the data center identifier (0-31)
	DataCenterID int64 `yaml:"data_center_id"`
	// MachineID is the machine identifier (0-31)
	MachineID int64 `yaml:"machine_id"`
}

// NOTE: Token types have been moved to internal/types/token.go

// ChainType represents the blockchain type
type ChainType string

const (
	// ChainTypeEthereum represents Ethereum blockchain
	ChainTypeEthereum ChainType = "ethereum"
	// ChainTypePolygon represents Polygon blockchain
	ChainTypePolygon ChainType = "polygon"
	// ChainTypeBase represents Base blockchain
	ChainTypeBase ChainType = "base"
)

// NOTE: Token configuration has been moved to the database and is now handled by
// the TokenStore in internal/core/tokenstore/

// PriceFeedConfig holds configuration for the external price feed provider
type PriceFeedConfig struct {
	Provider        string `yaml:"provider"` // e.g., "coincap"
	APIURL          string `yaml:"api_url"`
	APIKey          string `yaml:"api_key"`
	Limit           int    `yaml:"limit"`
	RefreshInterval int    `yaml:"refresh_interval"` // Interval in seconds
}

// TransactionConfig holds configuration for transaction processing
type TransactionConfig struct {
	// HistorySynchInterval is the time interval in seconds between transaction synching cycles
	HistorySynchInterval int `yaml:"history_synch_interval"`
	// TransactionUpdateInterval is the time interval in seconds between pending transaction polling cycles
	TransactionUpdateInterval int `yaml:"transaction_update_interval"`
}

// VaultConfig holds configuration for vault management
type VaultConfig struct {
	// DeploymentUpdateInterval is the time interval in seconds between checking pending vault deployments
	DeploymentUpdateInterval int `yaml:"deployment_update_interval"`
	// RecoveryUpdateInterval is the time interval in seconds between checking for eligible vault recoveries
	RecoveryUpdateInterval int `yaml:"recovery_update_interval"`
}

// Config holds the application configuration
type Config struct {
	// DBPath is the path to the SQLite database file
	DBPath string `yaml:"db_path"`
	// Port is the server port to listen on
	Port string `yaml:"port"`
	// UIPath is the path to the static React UI files
	UIPath string `yaml:"ui_path"`
	// MigrationsPath is the path to the migration files
	MigrationsPath string `yaml:"migrations_path"`
	// DBEncryptionKey is the base64-encoded key used for encrypting sensitive data in the database
	DBEncryptionKey string `yaml:"db_encryption_key"`
	// SmartContractsPath is the path to the compiled smart contract artifacts
	SmartContractsPath string `yaml:"smart_contracts_path"`
	// KeyStoreType specifies the type of key store to use (db or kms)
	KeyStoreType string `yaml:"key_store_type"`
	// Transaction holds configuration for transaction processing
	Transaction TransactionConfig `yaml:"transaction"`
	// Vault holds configuration for vault management
	Vault VaultConfig `yaml:"vault"`
	// Snowflake holds configuration for ID generation
	Snowflake SnowflakeConfig `yaml:"snowflake"`
	// Log holds the logging configuration
	Log LogConfig `yaml:"log"`
	// Blockchains holds configuration for all supported blockchains
	Blockchains BlockchainsConfig `yaml:"blockchains"`
	// PriceFeed holds configuration for the price feed service
	PriceFeed PriceFeedConfig `yaml:"price_feed"`
	// ABIMapping maps supported ABI types (e.g., "erc20") to their contract artifact names
	ABIMapping map[string]string `yaml:"abi_mapping"`
}

// LoadConfig loads the application configuration from YAML file and environment variables
func LoadConfig() (*Config, error) {
	// Try to load environment variables from .env files first
	loadEnvFiles()

	config := &Config{}
	var yamlData []byte
	var err error

	// Try to load YAML config from CONFIG_PATH or default locations
	configPaths := []string{
		os.Getenv("CONFIG_PATH"),
		".config.yaml",
		"../.config.yaml",
	}

	for _, path := range configPaths {
		if path == "" {
			continue
		}

		if yamlData, err = os.ReadFile(path); err == nil {
			fmt.Printf("Loading config from %s\n", path)
			break
		}
	}

	if err != nil {
		return nil, fmt.Errorf("no config file found in paths: %v", configPaths)
	}

	// Parse YAML with environment variable interpolation
	interpolatedYaml := interpolateEnvVars(string(yamlData))
	if err := yaml.Unmarshal([]byte(interpolatedYaml), config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return config, nil
}

// interpolateEnvVars replaces environment variables with their values, supporting default values
func interpolateEnvVars(content string) string {
	// Match ${VAR:-default} and $VAR formats
	re := regexp.MustCompile(`\$\{([^}]+)\}|\$([A-Za-z0-9_]+)`)

	return re.ReplaceAllStringFunc(content, func(match string) string {
		// Extract var name and default value
		varName := match
		defaultValue := ""

		// Remove ${ and }
		varName = strings.TrimPrefix(varName, "${")
		varName = strings.TrimPrefix(varName, "$")
		varName = strings.TrimSuffix(varName, "}")

		// Check for default value syntax: VAR:-default
		if strings.Contains(varName, ":-") {
			parts := strings.SplitN(varName, ":-", 2)
			varName = parts[0]
			defaultValue = parts[1]
		}

		// Get environment variable value
		if value, exists := os.LookupEnv(varName); exists && value != "" {
			return value
		}

		// Return default value if specified, otherwise empty string
		return defaultValue
	})
}

// loadEnvFiles tries to load environment variables from .env files in multiple locations
func loadEnvFiles() {
	// Check if a custom .env file path is provided
	customEnvPath := os.Getenv("ENV_FILE")
	if customEnvPath != "" {
		if err := godotenv.Load(customEnvPath); err != nil {
			fmt.Printf("Warning: could not load custom .env file from %s: %v\n", customEnvPath, err)
		} else {
			fmt.Printf("Loaded environment variables from custom .env file: %s\n", customEnvPath)
			return // If custom file is loaded successfully, don't try other files
		}
	}

	// Try the default .env file in the current directory
	if err := godotenv.Load(); err == nil {
		fmt.Println("Loaded environment variables from .env file")
		return
	}

	// Try .env file in the parent directory (useful for development)
	if err := godotenv.Load("../.env"); err == nil {
		fmt.Println("Loaded environment variables from ../.env file")
		return
	}

	// If no .env file is found, just continue with the default values
	fmt.Println("No .env file found, using default values")
}

// GetSmartContractsPath returns the path to the compiled smart contract artifacts
func (c *Config) GetSmartContractsPath() string {
	return c.SmartContractsPath
}

// GetArtifactPathForType returns the configured contract path
// for a given supported ABI type string.
func (c *Config) GetArtifactPathForType(abiType string) (string, error) {
	if c.ABIMapping == nil {
		return "", fmt.Errorf("ABI mapping configuration is missing")
	}
	contractName, ok := c.ABIMapping[abiType]
	if !ok {
		return "", fmt.Errorf("ABI mapping not found for type: %s", abiType)
	}
	if contractName == "" {
		return "", fmt.Errorf("ABI mapping is empty for type: %s", abiType)
	}
	return contractName, nil
}
