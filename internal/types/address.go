package types

import (
	"strings"

	"vault0/internal/errors"

	"github.com/ethereum/go-ethereum/common"
)

const (
	// ZeroAddress represents the native token address (0x0)
	ZeroAddress = "0x0000000000000000000000000000000000000000"
)

// Address represents a blockchain address with its associated chain type
type Address struct {
	// ChainType is the blockchain network this address belongs to
	ChainType ChainType

	// Address is the string representation of the blockchain address
	Address string
}

// NewAddress creates a new Address instance and validates the address format
// based on the specified chain type
func NewAddress(chainType ChainType, address string) (*Address, error) {
	// Normalize address format
	normalizedAddress := normalizeAddress(address, chainType)

	// Create new address instance
	addr := &Address{
		ChainType: chainType,
		Address:   normalizedAddress,
	}

	// Validate the address
	if err := addr.Validate(); err != nil {
		return nil, err
	}

	return addr, nil
}

// Validate checks if the address is valid for its chain type
func (a *Address) Validate() error {
	if a.Address == "" {
		return errors.NewInvalidAddressError("")
	}

	switch a.ChainType {
	case ChainTypeEthereum, ChainTypePolygon, ChainTypeBase:
		// Check if the address has the correct format
		if !common.IsHexAddress(a.Address) {
			return errors.NewInvalidAddressError(a.Address)
		}

		// Convert to checksum address and verify
		checksumAddr := common.HexToAddress(a.Address).Hex()
		if common.HexToAddress(a.Address) != common.HexToAddress(checksumAddr) {
			return errors.NewInvalidAddressError(a.Address)
		}

		return nil
	default:
		return errors.NewChainNotSupportedError(string(a.ChainType))
	}
}

// IsValid returns true if the address is valid for its chain type
func (a *Address) IsValid() bool {
	return a.Validate() == nil
}

// String returns the string representation of the address
func (a *Address) String() string {
	return a.Address
}

// ToChecksum returns the checksum version of the address for EVM chains
func (a *Address) ToChecksum() string {
	switch a.ChainType {
	case ChainTypeEthereum, ChainTypePolygon, ChainTypeBase:
		return common.HexToAddress(a.Address).Hex()
	default:
		return a.Address
	}
}

// IsZeroAddress returns true if the address is the zero address
func (a *Address) IsZeroAddress() bool {
	return a.Address == ZeroAddress
}

// normalizeAddress normalizes an address based on chain type
func normalizeAddress(address string, chainType ChainType) string {
	switch chainType {
	case ChainTypeEthereum, ChainTypePolygon, ChainTypeBase:
		// Ensure address has 0x prefix
		if !strings.HasPrefix(address, "0x") {
			address = "0x" + address
		}
		// Return checksum address
		return common.HexToAddress(address).Hex()
	default:
		return address
	}
}

// IsZeroAddress checks if the address is the zero address
func IsZeroAddress(address string) bool {
	return address == ZeroAddress || address == "0x0"
}
