package abi

import (
	"bytes"
	"fmt"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"

	"vault0/internal/errors"
	"vault0/internal/logger"
	"vault0/internal/types"
)

// EVMABIUtils implements the ABIUtils interface for EVM chains.
type EVMABIUtils struct {
	log       logger.Logger
	chainType types.ChainType // Store chain type for context
}

// NewEvmAbiUtils creates a new EVM ABI utility instance for a specific chain.
func NewEvmAbiUtils(chainType types.ChainType, log logger.Logger) (ABIUtils, error) {
	return &EVMABIUtils{
		log:       log,
		chainType: chainType, // Store the chain type
	}, nil
}

// Unpack parses the input data of a transaction using the provided ABI and method name.
// If methodName is provided, it will be used to find the method in the ABI.
func (u *EVMABIUtils) Unpack(contractABI string, methodName string, inputData []byte) (map[string]interface{}, error) {
	// Parse the ABI string
	parsedABI, err := abi.JSON(strings.NewReader(contractABI))
	if err != nil {
		return nil, errors.NewABIParseError(err)
	}

	// Prepare variables
	var method abi.Method
	var dataToUnpack []byte
	var methodID []byte

	// Extract method selector if available
	// Assumes inputData ALWAYS includes the selector
	if len(inputData) < 4 {
		return nil, errors.NewABIInputDataTooShortError(len(inputData), 4)
	}
	methodID = inputData[:4]
	dataToUnpack = inputData[4:]

	// Case 1: Find method by name
	if methodName != "" {
		var ok bool
		method, ok = parsedABI.Methods[methodName]
		if !ok {
			return nil, errors.NewABIMethodNotFoundError(methodName, true)
		}

		// Verify the selector matches the method name
		if !bytes.Equal(method.ID, methodID) {
			return nil, errors.NewABIMethodSelectorMismatchError(methodName, method.ID, methodID)
		}

		// Found method by name successfully
	} else {
		// Case 2: Find method by selector (no name provided)
		found := false
		for _, m := range parsedABI.Methods {
			if bytes.Equal(m.ID, methodID) {
				method = m
				found = true
				break
			}
		}

		if !found {
			return nil, errors.NewABIMethodNotFoundError(fmt.Sprintf("%x", methodID), false)
		}

		// Found method by selector successfully
	}

	// Validate input data
	if len(method.Inputs) > 0 && len(dataToUnpack) == 0 {
		return nil, errors.NewABIInputDataEmptyError(method.Name)
	}

	if len(dataToUnpack) > 0 && len(dataToUnpack)%32 != 0 {
		return nil, errors.NewABIInputDataInvalidLengthError(method.Name, len(dataToUnpack))
	}

	// Unpack the arguments
	unpackedArgs := make(map[string]any)
	err = method.Inputs.UnpackIntoMap(unpackedArgs, dataToUnpack)
	if err != nil {
		return nil, errors.NewABIUnpackFailedError(err, method.Name)
	}

	return unpackedArgs, nil
}

// ExtractMethodID extracts the 4-byte method selector from transaction data.
// Returns nil if data is shorter than 4 bytes.
func (u EVMABIUtils) ExtractMethodID(data []byte) []byte {
	if len(data) < 4 {
		return nil
	}
	return data[:4]
}

// Pack packs arguments for a method call.
func (u *EVMABIUtils) Pack(contractABI string, methodName string, args ...interface{}) ([]byte, error) {
	if contractABI == "" {
		return nil, errors.NewInvalidParameterError("contract ABI is empty", "contractABI")
	}

	// Parse the ABI string
	parsedABI, err := abi.JSON(strings.NewReader(contractABI))
	if err != nil {
		return nil, errors.NewABIParseError(err)
	}

	data, err := parsedABI.Pack(methodName, args...)
	if err != nil {
		if strings.Contains(err.Error(), "no method with id") || strings.Contains(err.Error(), "method '"+methodName+"' not found") {
			return nil, errors.NewABIMethodNotFoundError(methodName, true)
		}
		return nil, errors.NewABIPackFailedError(err, methodName)
	}
	return data, nil
}

// GetAddressFromArgs helper implementation
func (u *EVMABIUtils) GetAddressFromArgs(args map[string]any, key string) (types.Address, error) {
	val, ok := args[key]
	if !ok {
		return types.Address{}, errors.NewABIArgumentNotFoundError(key)
	}

	// Check if it's already a types.Address
	if addr, ok := val.(types.Address); ok {
		return addr, nil
	}

	// Check if it's a common.Address (expected case with go-ethereum)
	if ethAddr, ok := val.(common.Address); ok {
		// Convert common.Address to types.Address
		addr, err := types.NewAddress(u.chainType, ethAddr.Hex())
		if err != nil {
			return types.Address{}, errors.NewABIArgumentConversionError(err, key, "types.Address from common.Address", ethAddr.Hex())
		}
		return *addr, nil
	}

	// As a fallback, try to handle it as a string address
	if addrStr, ok := val.(string); ok {
		addr, err := types.NewAddress(u.chainType, addrStr)
		if err != nil {
			return types.Address{}, errors.NewABIArgumentConversionError(err, key, "types.Address from string", addrStr)
		}
		return *addr, nil
	}

	return types.Address{}, errors.NewABIArgumentInvalidTypeError(key, "types.Address, common.Address or string", fmt.Sprintf("%T", val))
}

// GetBytes32FromArgs helper implementation
func (u *EVMABIUtils) GetBytes32FromArgs(args map[string]any, key string) ([32]byte, error) {
	val, ok := args[key]
	if !ok {
		return [32]byte{}, errors.NewABIArgumentNotFoundError(key)
	}
	bytes32Val, ok := val.([32]byte)
	if !ok {
		byteSlice, sliceOk := val.([]byte)
		if sliceOk && len(byteSlice) == 32 {
			var result [32]byte
			copy(result[:], byteSlice)
			return result, nil
		} else {
			return [32]byte{}, errors.NewABIArgumentInvalidTypeError(key, "[32]byte or []byte with length 32", fmt.Sprintf("%T", val))
		}
	}
	return bytes32Val, nil
}

// GetBigIntFromArgs helper implementation
func (u *EVMABIUtils) GetBigIntFromArgs(args map[string]any, key string) (*types.BigInt, error) {
	val, ok := args[key]
	if !ok {
		return nil, errors.NewABIArgumentNotFoundError(key)
	}
	typesBigIntValue, ok := val.(types.BigInt)
	if ok {
		newVal := types.NewBigInt(typesBigIntValue.Int)
		return &newVal, nil
	}
	typesBigIntPtrValue, ok := val.(*types.BigInt)
	if ok {
		if typesBigIntPtrValue == nil || typesBigIntPtrValue.Int == nil {
			return nil, errors.NewABIArgumentNilValueError(key)
		}
		newVal := types.NewBigInt(typesBigIntPtrValue.Int)
		return &newVal, nil
	}
	goBigInt, goOk := val.(*big.Int)
	if goOk {
		if goBigInt == nil {
			return nil, errors.NewABIArgumentNilValueError(key)
		}
		newVal := types.NewBigInt(goBigInt)
		return &newVal, nil
	}
	return nil, errors.NewABIArgumentInvalidTypeError(key, "types.BigInt, *types.BigInt or *big.Int", fmt.Sprintf("%T", val))
}

// GetUint64FromArgs helper implementation
func (u *EVMABIUtils) GetUint64FromArgs(args map[string]any, key string) (uint64, error) {
	val, ok := args[key]
	if !ok {
		return 0, errors.NewABIArgumentNotFoundError(key)
	}
	uintVal, ok := val.(uint64)
	if !ok {
		bigIntVal, bigOk := val.(*big.Int)
		if bigOk && bigIntVal.IsUint64() {
			return bigIntVal.Uint64(), nil
		} else {
			return 0, errors.NewABIArgumentInvalidTypeError(key, "uint64 or *big.Int representable as uint64", fmt.Sprintf("%T", val))
		}
	}
	return uintVal, nil
}
