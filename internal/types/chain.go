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

// ChainType represents the blockchain type
type ChainType string

// Supported blockchain types
const (
	ChainTypeEthereum ChainType = "ethereum"
	ChainTypePolygon  ChainType = "polygon"
	ChainTypeBase     ChainType = "base"
)

// Chain errors
var (
	ErrInvalidAddress   = errors.New("invalid address")
	ErrUnsupportedChain = errors.New("unsupported blockchain")
	ErrMissingRPCURL    = errors.New("missing RPC URL")
)

// Chain represents information about a specific blockchain network.
// It contains all the necessary configuration and metadata to interact
// with a particular blockchain, including network identifiers, connection
// details, and cryptographic parameters.
//
// This struct is used by blockchain implementations to determine how to
// connect to and interact with the specific blockchain network, and by
// application code to display information about the connected chain.
type Chain struct {
	ID              int64          // Unique identifier for the blockchain network
	Type            ChainType      // Type of blockchain (e.g., Ethereum, Polygon)
	Name            string         // Human-readable name of the blockchain network
	Symbol          string         // Native currency symbol (e.g., ETH, MATIC)
	RPCUrl          string         // JSON-RPC endpoint URL for connecting to the network
	ExplorerUrl     string         // Block explorer URL for viewing transactions
	Curve           elliptic.Curve // Elliptic curve used for cryptographic operations
	KeyType         KeyType        // Type of cryptographic keys used by the blockchain
	DefaultGasLimit uint64         // Default gas limit for transactions
	DefaultGasPrice uint64         // Default gas price for transactions
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
//   - Error if the chain type is unsupported
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
//   - Human-readable name of the blockchain
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
// This is the symbol of the blockchain's primary token used for gas fees.
//
// Parameters:
//   - chainType: The type of blockchain to get the symbol for
//
// Returns:
//   - Currency symbol (e.g., "ETH" for Ethereum)
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
// Parameters:
//   - chainType: The type of blockchain to get cryptographic parameters for
//
// Returns:
//   - Key type appropriate for the blockchain (e.g., ECDSA)
//   - Elliptic curve used by the blockchain (e.g., secp256k1 for Ethereum)
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

// IsValidAddress validates if the given address is a valid blockchain address.
//
// Parameters:
//   - address: The address to validate
//
// Returns:
//   - true if the address is valid, false otherwise
func (c *Chain) IsValidAddress(address string) bool {
	return c.ValidateAddress(address) == nil
}

// ValidateAddress validates if the given address is a valid blockchain address.
// The validation logic depends on the chain type.
//
// Parameters:
//   - address: The address to validate
//
// Returns:
//   - nil if the address is valid, otherwise returns an error with details
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
