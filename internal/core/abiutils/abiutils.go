package abiutils

import (
	"context"

	"vault0/internal/types"
)

// SupportedABIType defines the types of ABIs that can be explicitly loaded by name.
// Currently limited to ERC20 and MultiSig.
type SupportedABIType string

const (
	ABITypeERC20    SupportedABIType = "erc20"
	ABITypeMultiSig SupportedABIType = "multisig"
)

// ABIUtils defines methods for loading, parsing, and interacting with contract ABIs.
type ABIUtils interface {
	// LoadABIByName loads a specific, known ABI type (e.g., "erc20", "multisig") using the configured mapping.
	// It loads the corresponding artifact file from the filesystem.
	// The loaded ABI is cached internally by type name.
	LoadABIByName(ctx context.Context, abiType SupportedABIType) (string, error)

	// LoadABIByAddress attempts to load the ABI for a given contract address.
	// It should utilize a block explorer or other external source.
	// The loaded ABI is cached internally by address.
	LoadABIByAddress(ctx context.Context, contractAddress types.Address) (string, error)

	// ParseContractInput parses the input data of a transaction using the provided ABI and method name.
	// If methodName is provided, it will be used to find the method in the ABI.
	// If methodName is empty and includesSelector is true, the method will be identified by the first 4 bytes of inputData.
	// If includesSelector is true, the first 4 bytes of inputData are treated as the method selector.
	ParseContractInput(contractABI string, methodName string, inputData []byte) (map[string]any, error)

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

	// Pack packs the arguments according to the ABI specification for the given method name.
	Pack(contractABI string, name string, args ...any) ([]byte, error)
}
