package abi

import (
	"vault0/internal/errors"
	"vault0/internal/logger"
	"vault0/internal/types"
)

// ABIUtils defines methods for parsing, and interacting with contract ABIs.
type ABIUtils interface {
	// Unpack parses the input data of a transaction using the provided ABI and method name.
	// If methodName is provided, it will be used to find the method in the ABI.
	// If methodName is empty and includesSelector is true, the method will be identified by the first 4 bytes of inputData.
	// If includesSelector is true, the first 4 bytes of inputData are treated as the method selector.
	Unpack(contractABI string, methodName string, inputData []byte) (map[string]any, error)

	// Pack packs the arguments according to the ABI specification for the given method name.
	Pack(contractABI string, methodName string, args ...any) ([]byte, error)

	// ExtractMethodID extracts the 4-byte method selector from transaction data.
	// Returns nil if data is shorter than 4 bytes.
	ExtractMethodID(data []byte) []byte

	// Helper function to get an address from parsed ABI arguments.
	GetAddressFromArgs(args map[string]any, key string) (types.Address, error)

	// Helper function to get bytes32 from parsed ABI arguments.
	GetBytes32FromArgs(args map[string]any, key string) ([32]byte, error)

	// Helper function to get *types.BigInt from parsed ABI arguments.
	GetBigIntFromArgs(args map[string]any, key string) (*types.BigInt, error)

	// Helper function to get uint64 from parsed ABI arguments.
	GetUint64FromArgs(args map[string]any, key string) (uint64, error)
}

func NewABIUtils(chainType types.ChainType, log logger.Logger) (ABIUtils, error) {
	switch chainType {
	case types.ChainTypeEthereum, types.ChainTypePolygon, types.ChainTypeBase:
		return NewEvmAbiUtils(chainType, log)
	default:
		return nil, errors.NewChainNotSupportedError(string(chainType))
	}
}
