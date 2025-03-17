package types

import (
	"fmt"
	"strings"
)

const (
	// ZeroAddress represents the native token address (0x0)
	ZeroAddress = "0x0000000000000000000000000000000000000000"
)

// TokenType represents the type of token
type TokenType string

// Supported token types
const (
	TokenTypeNative TokenType = "native"
	TokenTypeERC20  TokenType = "erc20"
)

// Token represents a cryptocurrency token
type Token struct {
	// ID is the unique identifier for the token
	ID string

	// Address is the contract address of the token
	// For native tokens, this is typically the zero address
	Address string

	// ChainType is the blockchain network the token exists on
	ChainType ChainType

	// Symbol is the token's ticker symbol (e.g., ETH, USDC)
	Symbol string

	// Decimals is the number of decimal places the token supports
	Decimals uint8

	// Type indicates if the token is native to the chain or an ERC20 token
	Type TokenType
}

// IsNative returns true if the token is a native token of its blockchain
func (t *Token) IsNative() bool {
	return t.Type == TokenTypeNative
}

// IsERC20 returns true if the token is an ERC20 token
func (t *Token) IsERC20() bool {
	return t.Type == TokenTypeERC20
}

// NormalizeAddress ensures consistent address format for storage and comparison
func NormalizeAddress(address string) string {
	// Convert to lowercase for case-insensitive comparisons
	address = strings.ToLower(address)

	// Ensure the address has 0x prefix for EVM addresses
	if !strings.HasPrefix(address, "0x") {
		address = "0x" + address
	}

	return address
}

// IsZeroAddress checks if the address is the zero address
func IsZeroAddress(address string) bool {
	normalized := NormalizeAddress(address)
	return normalized == ZeroAddress || normalized == "0x0"
}

// Validate checks if the token configuration is valid
func (t *Token) Validate() error {
	// Validate Symbol
	if t.Symbol == "" {
		return fmt.Errorf("token symbol cannot be empty")
	}

	// Validate Type
	if t.Type != TokenTypeNative && t.Type != TokenTypeERC20 {
		return fmt.Errorf("invalid token type: %s", t.Type)
	}

	// Validate Address
	if t.Address == "" {
		return fmt.Errorf("token address cannot be empty")
	}

	// For ERC20 tokens, the address cannot be the zero address
	if t.Type == TokenTypeERC20 && IsZeroAddress(t.Address) {
		return fmt.Errorf("ERC20 token cannot have zero address")
	}

	// For native tokens, we typically expect the zero address
	if t.Type == TokenTypeNative && !IsZeroAddress(t.Address) {
		return fmt.Errorf("native token should use zero address, got: %s", t.Address)
	}

	// Validate ChainType
	switch t.ChainType {
	case ChainTypeEthereum, ChainTypePolygon, ChainTypeBase:
		// These are valid chain types
	default:
		return fmt.Errorf("unsupported chain type: %s", t.ChainType)
	}

	return nil
}

// GetID returns a unique identifier for this token
// combining the address and chain type
func (t *Token) GetID() string {
	return fmt.Sprintf("%s:%s", t.Address, t.ChainType)
}

// ParseTokenID parses a token ID into its components
func ParseTokenID(id string) (address string, chainType ChainType, err error) {
	parts := strings.Split(id, ":")
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid token ID format: %s", id)
	}

	address = parts[0]
	chainType = ChainType(parts[1])

	return address, chainType, nil
}
