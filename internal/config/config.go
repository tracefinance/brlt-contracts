package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/joho/godotenv"
)

// BlockchainConfig holds configuration for a specific blockchain
type BlockchainConfig struct {
	// RPCURL is the RPC URL for the blockchain
	RPCURL string
	// ChainID is the chain ID for the blockchain
	ChainID int64
	// DefaultGasPrice is the default gas price for transactions in Gwei
	DefaultGasPrice int64
	// DefaultGasLimit is the default gas limit for transactions
	DefaultGasLimit uint64
}

// BlockchainsConfig holds configuration for all supported blockchains
type BlockchainsConfig struct {
	// Ethereum holds Ethereum blockchain configuration
	Ethereum BlockchainConfig
	// Polygon holds Polygon blockchain configuration
	Polygon BlockchainConfig
	// Base holds Base blockchain configuration
	Base BlockchainConfig
}

// Config holds the application configuration
type Config struct {
	// DBPath is the path to the SQLite database file
	DBPath string
	// Port is the server port to listen on
	Port string
	// UIPath is the path to the static React UI files
	UIPath string
	// MigrationsPath is the path to the migration files
	MigrationsPath string
	// DBEncryptionKey is the base64-encoded key used for encrypting sensitive data in the database
	DBEncryptionKey string

	// Blockchains holds configuration for all supported blockchains
	Blockchains BlockchainsConfig
}

// LoadConfig loads the application configuration from environment variables
// or falls back to default values
func LoadConfig() *Config {
	// Try to load environment variables from .env files
	loadEnvFiles()

	// Get base directory for paths
	baseDir := os.Getenv("APP_BASE_DIR")
	if baseDir == "" {
		// Default to current working directory if not specified
		currentDir, _ := os.Getwd()
		baseDir = currentDir
	}

	// Get the database path from environment or use default
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		// Default to a db file in the project root
		dbPath = filepath.Join(baseDir, "vault0.db")
	}

	// Get the port from environment or use default
	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = "8080"
	}

	// Set UI path, default to ui/dist
	uiPath := os.Getenv("UI_PATH")
	if uiPath == "" {
		uiPath = filepath.Join(baseDir, "ui", "dist")
	}

	// Set migrations path
	migrationsPath := os.Getenv("MIGRATIONS_PATH")
	if migrationsPath == "" {
		migrationsPath = filepath.Join(baseDir, "migrations")
	}

	// Get DB encryption key from environment
	dbEncryptionKey := os.Getenv("DB_ENCRYPTION_KEY")
	// No default for encryption key, it must be provided

	// Load blockchain configurations
	blockchains := BlockchainsConfig{
		Ethereum: loadEthereumConfig(),
		Polygon:  loadPolygonConfig(),
		Base:     loadBaseConfig(),
	}

	return &Config{
		DBPath:          dbPath,
		Port:            port,
		UIPath:          uiPath,
		MigrationsPath:  migrationsPath,
		DBEncryptionKey: dbEncryptionKey,
		Blockchains:     blockchains,
	}
}

// loadEthereumConfig loads Ethereum configuration from environment variables
func loadEthereumConfig() BlockchainConfig {
	rpcURL := os.Getenv("ETHEREUM_RPC_URL")
	if rpcURL == "" {
		rpcURL = "https://ethereum-rpc.publicnode.com"
	}

	return BlockchainConfig{
		RPCURL:          rpcURL,
		ChainID:         parseEnvInt("ETHEREUM_CHAIN_ID", 1),
		DefaultGasPrice: parseEnvInt("ETHEREUM_GAS_PRICE", 20), // Gwei
		DefaultGasLimit: parseEnvUint("ETHEREUM_GAS_LIMIT", 21000),
	}
}

// loadPolygonConfig loads Polygon configuration from environment variables
func loadPolygonConfig() BlockchainConfig {
	rpcURL := os.Getenv("POLYGON_RPC_URL")
	if rpcURL == "" {
		rpcURL = "https://polygon-rpc.com"
	}

	return BlockchainConfig{
		RPCURL:          rpcURL,
		ChainID:         parseEnvInt("POLYGON_CHAIN_ID", 137),
		DefaultGasPrice: parseEnvInt("POLYGON_GAS_PRICE", 30), // Gwei
		DefaultGasLimit: parseEnvUint("POLYGON_GAS_LIMIT", 21000),
	}
}

// loadBaseConfig loads Base configuration from environment variables
func loadBaseConfig() BlockchainConfig {
	rpcURL := os.Getenv("BASE_RPC_URL")
	if rpcURL == "" {
		rpcURL = "https://mainnet.base.org"
	}

	return BlockchainConfig{
		RPCURL:          rpcURL,
		ChainID:         parseEnvInt("BASE_CHAIN_ID", 8453),
		DefaultGasPrice: parseEnvInt("BASE_GAS_PRICE", 10), // Gwei
		DefaultGasLimit: parseEnvUint("BASE_GAS_LIMIT", 21000),
	}
}

// Helper function to parse integers from environment variables
func parseEnvInt(key string, defaultValue int64) int64 {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}

	value, err := strconv.ParseInt(valueStr, 10, 64)
	if err != nil {
		return defaultValue
	}

	return value
}

// Helper function to parse unsigned integers from environment variables
func parseEnvUint(key string, defaultValue uint64) uint64 {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}

	value, err := strconv.ParseUint(valueStr, 10, 64)
	if err != nil {
		return defaultValue
	}

	return value
}

// GetBlockchainConfig returns the configuration for a specific blockchain by its type
func (c *Config) GetBlockchainConfig(chainType string) *BlockchainConfig {
	switch chainType {
	case "ethereum":
		return &c.Blockchains.Ethereum
	case "polygon":
		return &c.Blockchains.Polygon
	case "base":
		return &c.Blockchains.Base
	default:
		return nil
	}
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
