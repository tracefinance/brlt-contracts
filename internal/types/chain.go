package types

import (
	"crypto/elliptic"
	"errors"
	"fmt"
	"strings"

	"github.com/ethereum/go-ethereum/common"

	"vault0/internal/config"
	"vault0/internal/core/crypto"
)

type ChainType string

// Supported blockchain types
const (
	ChainTypeEthereum ChainType = "ethereum"
	ChainTypePolygon  ChainType = "polygon"
	ChainTypeBase     ChainType = "base"
)

var (
	ErrInvalidAddress   = errors.New("invalid address")
	ErrUnsupportedChain = errors.New("unsupported blockchain")
	ErrMissingRPCURL    = errors.New("missing RPC URL")
)

// Chain represents a blockchain network configuration and its operational parameters.
// It provides network identifiers, connection details, and cryptographic settings
// needed to interact with the blockchain, validate addresses, and configure transactions.
type Chain struct {
	ID              int64          // Network identifier (e.g., 1 for Ethereum mainnet)
	Type            ChainType      // Blockchain platform (Ethereum, Polygon, etc.)
	Name            string         // Human-readable network name
	Symbol          string         // Native currency symbol (ETH, MATIC)
	RPCUrl          string         // JSON-RPC endpoint URL
	ExplorerUrl     string         // Block explorer URL
	Curve           elliptic.Curve // Elliptic curve for crypto operations
	KeyType         KeyType        // Cryptographic key type
	DefaultGasLimit uint64         // Default transaction gas limit
	DefaultGasPrice uint64         // Default transaction gas price
}

// ChainFactory defines an interface for creating Chain instances.
// Implementations of this interface are responsible for initializing
// Chain objects with the appropriate configuration and parameters
// for different blockchain networks.
type ChainFactory interface {
	// NewChain creates a Chain struct for the specified blockchain type.
	// It loads configuration from the provided config object and sets up
	// all necessary parameters for interacting with the blockchain.
	//
	// Parameters:
	//   - chainType: The type of blockchain to create (e.g., Ethereum, Polygon)
	//
	// Returns:
	//   - A fully initialized Chain struct if successful
	//   - Error if the chain type is unsupported or if required configuration is missing
	NewChain(chainType ChainType) (Chain, error)
}

// chainFactory implements the ChainFactory interface.
// It uses a configuration object to create properly configured Chain instances.
type chainFactory struct {
	cfg *config.Config
}

// NewChain creates a new Chain instance for the specified blockchain type.
// It initializes the chain with configuration values from the factory's config object.
//
// Parameters:
//   - chainType: The type of blockchain to create (e.g., Ethereum, Polygon)
//
// Returns:
//   - A fully initialized Chain struct if successful
//   - Error if:
//   - The chain type is unsupported (ErrUnsupportedChain)
//   - The RPC URL is not configured (ErrMissingRPCURL)
func (f *chainFactory) NewChain(chainType ChainType) (Chain, error) {
	chainCfg, err := getChainConfig(chainType, f.cfg)
	if err != nil {
		return Chain{}, err
	}

	if chainCfg.RPCURL == "" {
		return Chain{}, fmt.Errorf("missing RPC URL for %s: %w", chainType, ErrMissingRPCURL)
	}

	// Determine the key type and curve for the chain
	keyType, curve := getChainCryptoParams(chainType)

	return Chain{
		ID:              chainCfg.ChainID,
		Type:            chainType,
		Name:            getChainName(chainType),
		Symbol:          getChainSymbol(chainType),
		RPCUrl:          chainCfg.RPCURL,
		ExplorerUrl:     chainCfg.ExplorerURL,
		KeyType:         keyType,
		Curve:           curve,
		DefaultGasLimit: chainCfg.DefaultGasLimit,
		DefaultGasPrice: chainCfg.DefaultGasPrice,
	}, nil
}

// NewChainFactory creates a new instance of chainFactory.
// The factory uses the provided configuration to initialize Chain instances.
//
// Parameters:
//   - cfg: Configuration object containing settings for all supported blockchains
//
// Returns:
//   - A ChainFactory interface implementation
func NewChainFactory(cfg *config.Config) ChainFactory {
	return &chainFactory{
		cfg: cfg,
	}
}

// IsValidAddress validates if the given address is a valid blockchain address.
// This is a convenience method that returns a boolean instead of an error.
//
// Parameters:
//   - address: The address to validate
//
// Returns:
//   - true if the address is valid for this blockchain
//   - false if the address is invalid
func (c *Chain) IsValidAddress(address string) bool {
	return c.ValidateAddress(address) == nil
}

// ValidateAddress performs a thorough validation of a blockchain address.
// For EVM-compatible chains, it checks the address format and checksum.
//
// Parameters:
//   - address: The address to validate
//
// Returns:
//   - nil if the address is valid
//   - ErrInvalidAddress with details if the address is invalid
func (c *Chain) ValidateAddress(address string) error {
	// For EVM-compatible chains (Ethereum, Polygon, Base)
	switch c.Type {
	case ChainTypeEthereum, ChainTypePolygon, ChainTypeBase:
		// Check if the address has the correct format (0x followed by 40 hex characters)
		if !common.IsHexAddress(address) {
			return fmt.Errorf("%w: invalid format", ErrInvalidAddress)
		}

		// Convert to checksum address and verify
		checksumAddr := common.HexToAddress(address)
		if address != checksumAddr.Hex() && address != strings.ToLower(checksumAddr.Hex()) {
			return fmt.Errorf("%w: checksum validation failed", ErrInvalidAddress)
		}

		return nil
	default:
		return ErrInvalidAddress
	}
}

// getChainConfig returns the configuration for a given chain type.
// It extracts the appropriate blockchain configuration from the provided config object.
//
// Parameters:
//   - chainType: The type of blockchain to get configuration for
//   - config: Configuration object containing settings for all supported blockchains
//
// Returns:
//   - Pointer to the blockchain-specific configuration
//   - ErrUnsupportedChain if the chain type is not supported
func getChainConfig(chainType ChainType, config *config.Config) (*config.BlockchainConfig, error) {
	switch chainType {
	case ChainTypeEthereum:
		return &config.Blockchains.Ethereum, nil
	case ChainTypePolygon:
		return &config.Blockchains.Polygon, nil
	case ChainTypeBase:
		return &config.Blockchains.Base, nil
	default:
		return nil, fmt.Errorf("unsupported chain type %s: %w", chainType, ErrUnsupportedChain)
	}
}

// getChainName returns the human-readable name for a given blockchain type.
// This is used for display purposes in user interfaces and logs.
//
// Parameters:
//   - chainType: The type of blockchain to get the name for
//
// Returns:
//   - Human-readable name of the blockchain (e.g., "Ethereum", "Polygon")
//   - "Unknown" if the chain type is not recognized
func getChainName(chainType ChainType) string {
	switch chainType {
	case ChainTypeEthereum:
		return "Ethereum"
	case ChainTypePolygon:
		return "Polygon"
	case ChainTypeBase:
		return "Base"
	default:
		return "Unknown"
	}
}

// getChainSymbol returns the native currency symbol for a given blockchain type.
// This is the symbol of the blockchain's primary token used for gas fees and transactions.
//
// Parameters:
//   - chainType: The type of blockchain to get the symbol for
//
// Returns:
//   - Currency symbol (e.g., "ETH" for Ethereum, "MATIC" for Polygon)
//   - "UNKNOWN" if the chain type is not recognized
func getChainSymbol(chainType ChainType) string {
	switch chainType {
	case ChainTypeEthereum:
		return "ETH"
	case ChainTypePolygon:
		return "MATIC"
	case ChainTypeBase:
		return "ETH"
	default:
		return "UNKNOWN"
	}
}

// getChainCryptoParams returns the appropriate key type and elliptic curve
// for a given blockchain type. These parameters are used for key generation
// and cryptographic operations specific to the blockchain.
//
// For EVM-compatible chains (Ethereum, Polygon, Base), this returns:
//   - KeyType: KeyTypeECDSA
//   - Curve: secp256k1
//
// For unknown chains, it defaults to:
//   - KeyType: KeyTypeECDSA
//   - Curve: P-256 (NIST P-256)
//
// Parameters:
//   - chainType: The type of blockchain to get cryptographic parameters for
//
// Returns:
//   - KeyType appropriate for the blockchain
//   - Elliptic curve implementation used by the blockchain
func getChainCryptoParams(chainType ChainType) (KeyType, elliptic.Curve) {
	switch chainType {
	case ChainTypeEthereum, ChainTypePolygon, ChainTypeBase:
		// All EVM-compatible chains use ECDSA with secp256k1
		return KeyTypeECDSA, crypto.Secp256k1Curve
	default:
		// For unknown chains, default to ECDSA with P-256
		return KeyTypeECDSA, elliptic.P256()
	}
}
