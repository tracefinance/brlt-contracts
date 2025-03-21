package types

import (
	"fmt"
	"math/big"
	"strings"
	"vault0/internal/errors"
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

// ToBigFloat converts a big.Int value to its big.Float representation
// using the token's decimals (divides by 10^decimals)
func (t *Token) ToBigFloat(value *big.Int) *big.Float {
	result := new(big.Float).SetPrec(128)

	if value == nil {
		return result
	}

	// Create a divisor based on the token decimals (10^decimals)
	divisor := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(t.Decimals)), nil)

	// Convert to big.Float for decimal division
	floatValue := new(big.Float).SetPrec(128).SetInt(value)
	floatDivisor := new(big.Float).SetPrec(128).SetInt(divisor)

	// Perform the division to get the decimal value
	result.Quo(floatValue, floatDivisor)

	return result
}

// ToBigInt converts a big.Float decimal value to its big.Int representation
// using the token's decimals (multiplies by 10^decimals)
func (t *Token) ToBigInt(value *big.Float) *big.Int {
	if value == nil {
		return new(big.Int)
	}

	// Create a multiplier based on the token decimals (10^decimals)
	multiplier := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(t.Decimals)), nil)

	// Convert multiplier to big.Float for multiplication
	floatMultiplier := new(big.Float).SetPrec(128).SetInt(multiplier)

	// Multiply the value by 10^decimals
	result := new(big.Float).SetPrec(128).Mul(value, floatMultiplier)

	// Convert back to big.Int, truncating any remaining decimals
	intResult, _ := result.Int(nil)
	return intResult
}

// NewNativeToken creates a native token for the specified blockchain
// It automatically resolves the symbol and decimals based on the chain type
func NewNativeToken(chainType ChainType) (*Token, error) {
	// Get the symbol for this chain
	symbol := getChainSymbol(chainType)

	// Determine the decimals for the native token
	var decimals uint8
	switch chainType {
	case ChainTypeEthereum, ChainTypeBase, ChainTypePolygon:
		decimals = 18 // ETH and MATIC both have 18 decimals
	default:
		return nil, errors.NewChainNotSupportedError(string(chainType))
	}

	token := &Token{
		Address:   ZeroAddress,
		ChainType: chainType,
		Symbol:    symbol,
		Decimals:  decimals,
		Type:      TokenTypeNative,
	}

	// Validate the token configuration
	if err := token.Validate(); err != nil {
		return nil, err
	}

	return token, nil
}
