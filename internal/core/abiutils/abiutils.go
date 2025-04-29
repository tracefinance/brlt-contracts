package abiutils

import (
	"context"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"

	"vault0/internal/config"
	"vault0/internal/core/blockexplorer"
	"vault0/internal/errors"
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
	LoadABIByName(ctx context.Context, abiType SupportedABIType) (*abi.ABI, error)

	// LoadABIByAddress attempts to load the ABI for a given contract address.
	// It should utilize a block explorer or other external source.
	// The loaded ABI is cached internally by address.
	LoadABIByAddress(ctx context.Context, contractAddress common.Address) (*abi.ABI, error)

	// GetMethodFromABI finds a specific method within a parsed ABI using its 4-byte ID.
	GetMethodFromABI(contractABI *abi.ABI, methodID []byte) (*abi.Method, error)

	// ParseContractInput parses the input data of a transaction using a specific ABI method.
	// The input data should *not* include the 4-byte method selector.
	ParseContractInput(method *abi.Method, inputData []byte) (map[string]interface{}, error)

	// ExtractMethodID extracts the 4-byte method selector from transaction data.
	// Returns nil if data is shorter than 4 bytes.
	ExtractMethodID(data []byte) []byte

	// Helper function to get an address from parsed ABI arguments.
	GetAddressFromArgs(args map[string]interface{}, key string) (common.Address, error)

	// Helper function to get bytes32 from parsed ABI arguments.
	GetBytes32FromArgs(args map[string]interface{}, key string) ([32]byte, error)

	// Helper function to get *types.BigInt from parsed ABI arguments.
	GetBigIntFromArgs(args map[string]interface{}, key string) (*types.BigInt, error)

	// Helper function to get uint64 from parsed ABI arguments.
	GetUint64FromArgs(args map[string]interface{}, key string) (uint64, error)

	// Pack packs the arguments according to the ABI specification for the given method name.
	Pack(contractABI *abi.ABI, name string, args ...interface{}) ([]byte, error)
}

// NewABIUtils creates a new ABI utility instance for the specified chain type.
// It requires configuration and a block explorer factory to function.
func NewABIUtils(
	chainType types.ChainType,
	cfg *config.Config,
	explorerFactory blockexplorer.Factory,
) (ABIUtils, error) {
	// The actual implementation is in evm_abi_utils.go (or potentially other files for different chain types)
	switch chainType {
	case types.ChainTypeEthereum, types.ChainTypePolygon, types.ChainTypeBase:
		return NewEvmAbiUtils(chainType, cfg, explorerFactory)
	default:
		return nil, errors.NewChainNotSupportedError(string(chainType))
	}
}
