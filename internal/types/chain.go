package types

import (
	"crypto/elliptic"

	"github.com/ethereum/go-ethereum/common"

	"vault0/internal/config"
	"vault0/internal/core/crypto"
	"vault0/internal/errors"
)

type ChainType string

// Supported blockchain types
const (
	ChainTypeEthereum ChainType = "ethereum"
	ChainTypePolygon  ChainType = "polygon"
	ChainTypeBase     ChainType = "base"
)

// ChainLayer represents the layer classification of a blockchain
type ChainLayer string

// Supported blockchain layers
const (
	ChainLayerLayer1 ChainLayer = "layer1"
	ChainLayerLayer2 ChainLayer = "layer2"
)

// Chain represents a blockchain network configuration and its operational parameters.
// It provides network identifiers, connection details, and cryptographic settings
// needed to interact with the blockchain, validate addresses, and configure transactions.
type Chain struct {
	ID              int64          // Network identifier (e.g., 1 for Ethereum mainnet)
	Type            ChainType      // Blockchain platform (Ethereum, Polygon, etc.)
	Layer           ChainLayer     // Blockchain layer (Layer1, Layer2)
	Name            string         // Human-readable network name
	Symbol          string         // Native currency symbol (ETH, MATIC)
	RPCUrl          string         // JSON-RPC endpoint URL
	ExplorerUrl     string         // Block explorer URL
	ExplorerAPIKey  string         // Block explorer API key
	Curve           elliptic.Curve // Elliptic curve for crypto operations
	KeyType         KeyType        // Cryptographic key type
	DefaultGasLimit uint64         // Default transaction gas limit
	DefaultGasPrice uint64         // Default transaction gas price
}

// Chains represents a collection of blockchain configurations.
type Chains struct {
	Chains map[ChainType]Chain // Map of chain types to their configurations
}

// NewChains creates a new Chains instance with configurations from the provided config.
func NewChains(cfg *config.Config) (*Chains, error) {
	chainsMap := make(map[ChainType]Chain)
	for _, chainType := range []ChainType{ChainTypeEthereum, ChainTypePolygon, ChainTypeBase} {
		chain, err := newChain(cfg, chainType)
		if err != nil {
			return nil, err
		}
		chainsMap[chainType] = chain
	}
	return &Chains{
		Chains: chainsMap,
	}, nil
}

// Get returns the Chain configuration for the specified chain type.
func (c *Chains) Get(chainType ChainType) (Chain, error) {
	chain, exists := c.Chains[chainType]
	if !exists {
		return Chain{}, errors.NewChainNotSupportedError(string(chainType))
	}

	if chain.RPCUrl == "" {
		return Chain{}, errors.NewInvalidBlockchainConfigError(string(chainType), "rpc_url")
	}

	return chain, nil
}

// List returns a slice of all Chain configurations.
func (c *Chains) List() []Chain {
	chains := make([]Chain, 0, len(c.Chains))
	for _, chain := range c.Chains {
		chains = append(chains, chain)
	}
	return chains
}

// newChain creates a new Chain instance for the specified blockchain type.
// It initializes the chain with configuration values from the factory's config object.
//
// Parameters:
//   - chainType: The type of blockchain to create (e.g., Ethereum, Polygon)
//
// Returns:
//   - A fully initialized Chain struct if successful
//   - Error if:
//   - The chain type is unsupported (ErrChainNotSupported)
//   - The configuration is invalid (ErrInvalidBlockchainConfig)
func newChain(cfg *config.Config, chainType ChainType) (Chain, error) {
	chainCfg, err := getChainConfig(chainType, cfg)
	if err != nil {
		return Chain{}, err
	}

	if chainCfg.RPCURL == "" {
		return Chain{}, errors.NewInvalidBlockchainConfigError(string(chainType), "rpc_url")
	}

	// Determine the key type and curve for the chain
	keyType, curve := getChainCryptoParams(chainType)

	// Determine the chain layer
	layer := getChainLayer(chainType)

	return Chain{
		ID:              chainCfg.ChainID,
		Type:            chainType,
		Layer:           layer,
		Name:            getChainName(chainType),
		Symbol:          getChainSymbol(chainType),
		RPCUrl:          chainCfg.RPCURL,
		ExplorerUrl:     chainCfg.ExplorerURL,
		ExplorerAPIKey:  chainCfg.ExplorerAPIKey,
		KeyType:         keyType,
		Curve:           curve,
		DefaultGasLimit: chainCfg.DefaultGasLimit,
		DefaultGasPrice: chainCfg.DefaultGasPrice,
	}, nil
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
	if address == "" {
		return errors.NewInvalidAddressError("")
	}

	// For EVM-compatible chains (Ethereum, Polygon, Base)
	switch c.Type {
	case ChainTypeEthereum, ChainTypePolygon, ChainTypeBase:
		// Check if the address has the correct format (0x followed by 40 hex characters)
		if !common.IsHexAddress(address) {
			return errors.NewInvalidAddressError(address)
		}

		// Convert to checksum address and verify
		checksumAddr := common.HexToAddress(address).Hex()
		if address != checksumAddr {
			return errors.NewInvalidAddressError(address)
		}

		return nil
	default:
		return errors.NewChainNotSupportedError(string(c.Type))
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

// getChainConfig returns the configuration for a given chain type.
// It extracts the appropriate blockchain configuration from the provided config object.
//
// Parameters:
//   - chainType: The type of blockchain to get configuration for
//   - config: Configuration object containing settings for all supported blockchains
//
// Returns:
//   - Pointer to the blockchain-specific configuration
//   - ErrChainNotSupported if the chain type is not supported
func getChainConfig(chainType ChainType, config *config.Config) (*config.BlockchainConfig, error) {
	switch chainType {
	case ChainTypeEthereum:
		return &config.Blockchains.Ethereum, nil
	case ChainTypePolygon:
		return &config.Blockchains.Polygon, nil
	case ChainTypeBase:
		return &config.Blockchains.Base, nil
	default:
		return nil, errors.NewChainNotSupportedError(string(chainType))
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

// getChainLayer returns the layer classification for a given blockchain type.
// This classification categorizes blockchains into Layer 1 (base chains) and
// Layer 2 (scaling solutions).
//
// Parameters:
//   - chainType: The type of blockchain to get the layer for
//
// Returns:
//   - ChainLayer indicating whether the blockchain is Layer 1 or Layer 2
func getChainLayer(chainType ChainType) ChainLayer {
	switch chainType {
	case ChainTypeEthereum:
		return ChainLayerLayer1
	case ChainTypePolygon:
		return ChainLayerLayer2
	case ChainTypeBase:
		return ChainLayerLayer2
	default:
		return ChainLayerLayer1 // Default to Layer 1 for unknown chains
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
