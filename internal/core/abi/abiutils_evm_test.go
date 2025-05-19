package abi

import (
	"bytes"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"vault0/internal/errors"
	"vault0/internal/testing/mocks"
	"vault0/internal/types"
)

func TestNewEvmAbiUtils(t *testing.T) {
	log := mocks.NewNopLogger()
	utils, err := NewEvmAbiUtils(types.ChainTypeEthereum, log)

	require.NoError(t, err)
	require.NotNil(t, utils)
	require.IsType(t, &EVMABIUtils{}, utils)
}

func TestEVMABIUtils_Pack(t *testing.T) {
	log := mocks.NewNopLogger()
	utils, err := NewEvmAbiUtils(types.ChainTypeEthereum, log)
	require.NoError(t, err)

	tests := []struct {
		name        string
		contractABI string
		methodName  string
		args        []interface{}
		want        []byte
		wantErr     bool
		errCode     string
	}{
		{
			name:        "empty ABI",
			contractABI: "",
			methodName:  "transfer",
			args:        []interface{}{},
			want:        nil,
			wantErr:     true,
			errCode:     errors.ErrCodeInvalidParameter,
		},
		{
			name:        "invalid ABI",
			contractABI: "invalid-abi",
			methodName:  "transfer",
			args:        []interface{}{},
			want:        nil,
			wantErr:     true,
			errCode:     errors.ErrCodeABIParseFailed,
		},
		{
			name:        "method not found",
			contractABI: `[{"constant":false,"inputs":[{"name":"_to","type":"address"},{"name":"_value","type":"uint256"}],"name":"transfer","outputs":[{"name":"success","type":"bool"}],"payable":false,"stateMutability":"nonpayable","type":"function"}]`,
			methodName:  "nonExistentMethod",
			args:        []interface{}{},
			want:        nil,
			wantErr:     true,
			errCode:     errors.ErrCodeABIMethodNotFound,
		},
		{
			name:        "valid method with args",
			contractABI: `[{"constant":false,"inputs":[{"name":"_to","type":"address"},{"name":"_value","type":"uint256"}],"name":"transfer","outputs":[{"name":"success","type":"bool"}],"payable":false,"stateMutability":"nonpayable","type":"function"}]`,
			methodName:  "transfer",
			args:        []interface{}{common.HexToAddress("0x1234567890123456789012345678901234567890"), big.NewInt(1000)},
			want:        nil, // We won't check the exact value, just that it's not nil and no error
			wantErr:     false,
		},
		{
			name:        "invalid args",
			contractABI: `[{"constant":false,"inputs":[{"name":"_to","type":"address"},{"name":"_value","type":"uint256"}],"name":"transfer","outputs":[{"name":"success","type":"bool"}],"payable":false,"stateMutability":"nonpayable","type":"function"}]`,
			methodName:  "transfer",
			args:        []interface{}{"not-an-address", "not-a-number"},
			want:        nil,
			wantErr:     true,
			errCode:     errors.ErrCodeABIPackFailed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := utils.Pack(tt.contractABI, tt.methodName, tt.args...)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errCode != "" {
					appErr, ok := err.(*errors.Vault0Error)
					assert.True(t, ok, "Expected a Vault0Error")
					assert.Equal(t, tt.errCode, appErr.Code)
				}
			} else {
				assert.NoError(t, err)
				if tt.want != nil {
					assert.Equal(t, tt.want, got)
				} else {
					assert.NotNil(t, got)
				}
			}
		})
	}
}

func TestEVMABIUtils_Unpack(t *testing.T) {
	log := mocks.NewNopLogger()
	utils, err := NewEvmAbiUtils(types.ChainTypeEthereum, log)
	require.NoError(t, err)

	// Simple ABI for testing
	testABI := `[{"constant":false,"inputs":[{"name":"_to","type":"address"},{"name":"_value","type":"uint256"}],"name":"transfer","outputs":[{"name":"success","type":"bool"}],"payable":false,"stateMutability":"nonpayable","type":"function"}]`

	// Generate a method ID for testing
	parsedABI, err := abi.JSON(bytes.NewReader([]byte(testABI)))
	require.NoError(t, err)

	transferMethod := parsedABI.Methods["transfer"]
	methodID := transferMethod.ID

	// Pack some arguments for a valid call
	recipient := common.HexToAddress("0x1234567890123456789012345678901234567890")
	amount := big.NewInt(1000)
	packedArgs, err := transferMethod.Inputs.Pack(recipient, amount)
	require.NoError(t, err)

	// Create valid input data (method ID + packed args)
	validInputData := append(methodID, packedArgs...)

	tests := []struct {
		name        string
		contractABI string
		methodName  string
		inputData   []byte
		want        map[string]interface{}
		wantErr     bool
		errCode     string
	}{
		{
			name:        "invalid ABI",
			contractABI: "invalid-abi",
			methodName:  "transfer",
			inputData:   validInputData,
			want:        nil,
			wantErr:     true,
			errCode:     errors.ErrCodeABIParseFailed,
		},
		{
			name:        "input data too short",
			contractABI: testABI,
			methodName:  "transfer",
			inputData:   []byte{0x01, 0x02, 0x03}, // Less than 4 bytes
			want:        nil,
			wantErr:     true,
			errCode:     errors.ErrCodeABIInputDataTooShort,
		},
		{
			name:        "method not found by name",
			contractABI: testABI,
			methodName:  "nonExistentMethod",
			inputData:   validInputData,
			want:        nil,
			wantErr:     true,
			errCode:     errors.ErrCodeABIMethodNotFound,
		},
		{
			name:        "method selector mismatch",
			contractABI: testABI,
			methodName:  "transfer",
			inputData:   append([]byte{0xff, 0xff, 0xff, 0xff}, packedArgs...), // Wrong selector
			want:        nil,
			wantErr:     true,
			errCode:     errors.ErrCodeABIMethodSelectorMismatch,
		},
		{
			name:        "valid method name and data",
			contractABI: testABI,
			methodName:  "transfer",
			inputData:   validInputData,
			want: map[string]interface{}{
				"_to":    recipient,
				"_value": amount,
			},
			wantErr: false,
		},
		{
			name:        "find method by selector (no name)",
			contractABI: testABI,
			methodName:  "", // No method name provided
			inputData:   validInputData,
			want: map[string]interface{}{
				"_to":    recipient,
				"_value": amount,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := utils.Unpack(tt.contractABI, tt.methodName, tt.inputData)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errCode != "" {
					appErr, ok := err.(*errors.Vault0Error)
					assert.True(t, ok, "Expected a Vault0Error")
					assert.Equal(t, tt.errCode, appErr.Code)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, got)
				// Check we got the expected keys
				for k := range tt.want {
					assert.Contains(t, got, k)
				}
			}
		})
	}
}

func TestEVMABIUtils_ExtractMethodID(t *testing.T) {
	log := mocks.NewNopLogger()
	utils, err := NewEvmAbiUtils(types.ChainTypeEthereum, log)
	require.NoError(t, err)

	tests := []struct {
		name     string
		data     []byte
		expected []byte
	}{
		{
			name:     "empty data",
			data:     []byte{},
			expected: nil,
		},
		{
			name:     "data too short",
			data:     []byte{0x01, 0x02, 0x03},
			expected: nil,
		},
		{
			name:     "valid data",
			data:     []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06},
			expected: []byte{0x01, 0x02, 0x03, 0x04},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := utils.ExtractMethodID(tt.data)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestEVMABIUtils_GetAddressFromArgs(t *testing.T) {
	log := mocks.NewNopLogger()
	utils, err := NewEvmAbiUtils(types.ChainTypeEthereum, log)
	require.NoError(t, err)

	ethAddr := common.HexToAddress("0x1234567890123456789012345678901234567890")
	typesAddr, _ := types.NewAddress(types.ChainTypeEthereum, ethAddr.Hex())

	tests := []struct {
		name    string
		args    map[string]interface{}
		key     string
		want    types.Address
		wantErr bool
		errCode string
	}{
		{
			name:    "key not found",
			args:    map[string]interface{}{"wrong_key": ethAddr},
			key:     "address",
			want:    types.Address{},
			wantErr: true,
			errCode: errors.ErrCodeABIArgumentNotFound,
		},
		{
			name:    "common.Address value",
			args:    map[string]interface{}{"address": ethAddr},
			key:     "address",
			want:    *typesAddr,
			wantErr: false,
		},
		{
			name:    "types.Address value",
			args:    map[string]interface{}{"address": *typesAddr},
			key:     "address",
			want:    *typesAddr,
			wantErr: false,
		},
		{
			name:    "string value",
			args:    map[string]interface{}{"address": ethAddr.Hex()},
			key:     "address",
			want:    *typesAddr,
			wantErr: false,
		},
		{
			name:    "invalid type",
			args:    map[string]interface{}{"address": 123},
			key:     "address",
			want:    types.Address{},
			wantErr: true,
			errCode: errors.ErrCodeABIArgumentInvalidType,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := utils.GetAddressFromArgs(tt.args, tt.key)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errCode != "" {
					appErr, ok := err.(*errors.Vault0Error)
					assert.True(t, ok, "Expected a Vault0Error")
					assert.Equal(t, tt.errCode, appErr.Code)
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestEVMABIUtils_GetBytes32FromArgs(t *testing.T) {
	log := mocks.NewNopLogger()
	utils, err := NewEvmAbiUtils(types.ChainTypeEthereum, log)
	require.NoError(t, err)

	// Create a sample [32]byte and byte slice
	var bytes32Value [32]byte
	for i := 0; i < 32; i++ {
		bytes32Value[i] = byte(i)
	}
	byteSlice := make([]byte, 32)
	for i := 0; i < 32; i++ {
		byteSlice[i] = byte(i)
	}

	tests := []struct {
		name    string
		args    map[string]interface{}
		key     string
		want    [32]byte
		wantErr bool
		errCode string
	}{
		{
			name:    "key not found",
			args:    map[string]interface{}{"wrong_key": bytes32Value},
			key:     "bytes32",
			want:    [32]byte{},
			wantErr: true,
			errCode: errors.ErrCodeABIArgumentNotFound,
		},
		{
			name:    "[32]byte value",
			args:    map[string]interface{}{"bytes32": bytes32Value},
			key:     "bytes32",
			want:    bytes32Value,
			wantErr: false,
		},
		{
			name:    "[]byte value with length 32",
			args:    map[string]interface{}{"bytes32": byteSlice},
			key:     "bytes32",
			want:    bytes32Value,
			wantErr: false,
		},
		{
			name:    "[]byte value with wrong length",
			args:    map[string]interface{}{"bytes32": []byte{1, 2, 3}},
			key:     "bytes32",
			want:    [32]byte{},
			wantErr: true,
			errCode: errors.ErrCodeABIArgumentInvalidType,
		},
		{
			name:    "invalid type",
			args:    map[string]interface{}{"bytes32": "not a byte array"},
			key:     "bytes32",
			want:    [32]byte{},
			wantErr: true,
			errCode: errors.ErrCodeABIArgumentInvalidType,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := utils.GetBytes32FromArgs(tt.args, tt.key)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errCode != "" {
					appErr, ok := err.(*errors.Vault0Error)
					assert.True(t, ok, "Expected a Vault0Error")
					assert.Equal(t, tt.errCode, appErr.Code)
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestEVMABIUtils_GetBigIntFromArgs(t *testing.T) {
	log := mocks.NewNopLogger()
	utils, err := NewEvmAbiUtils(types.ChainTypeEthereum, log)
	require.NoError(t, err)

	// Create sample big.Int values
	bigIntValue := big.NewInt(1000)
	typesBigInt := types.NewBigInt(bigIntValue)

	tests := []struct {
		name    string
		args    map[string]interface{}
		key     string
		want    *types.BigInt
		wantErr bool
		errCode string
	}{
		{
			name:    "key not found",
			args:    map[string]interface{}{"wrong_key": bigIntValue},
			key:     "amount",
			want:    nil,
			wantErr: true,
			errCode: errors.ErrCodeABIArgumentNotFound,
		},
		{
			name:    "types.BigInt value",
			args:    map[string]interface{}{"amount": typesBigInt},
			key:     "amount",
			want:    &typesBigInt,
			wantErr: false,
		},
		{
			name:    "*types.BigInt value",
			args:    map[string]interface{}{"amount": &typesBigInt},
			key:     "amount",
			want:    &typesBigInt,
			wantErr: false,
		},
		{
			name:    "*big.Int value",
			args:    map[string]interface{}{"amount": bigIntValue},
			key:     "amount",
			want:    &typesBigInt,
			wantErr: false,
		},
		{
			name:    "nil *types.BigInt",
			args:    map[string]interface{}{"amount": (*types.BigInt)(nil)},
			key:     "amount",
			want:    nil,
			wantErr: true,
			errCode: errors.ErrCodeABIArgumentNilValue,
		},
		{
			name:    "nil *big.Int",
			args:    map[string]interface{}{"amount": (*big.Int)(nil)},
			key:     "amount",
			want:    nil,
			wantErr: true,
			errCode: errors.ErrCodeABIArgumentNilValue,
		},
		{
			name:    "invalid type",
			args:    map[string]interface{}{"amount": "not a big int"},
			key:     "amount",
			want:    nil,
			wantErr: true,
			errCode: errors.ErrCodeABIArgumentInvalidType,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := utils.GetBigIntFromArgs(tt.args, tt.key)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errCode != "" {
					appErr, ok := err.(*errors.Vault0Error)
					assert.True(t, ok, "Expected a Vault0Error")
					assert.Equal(t, tt.errCode, appErr.Code)
				}
			} else {
				assert.NoError(t, err)
				if tt.want == nil {
					assert.Nil(t, got)
				} else {
					assert.NotNil(t, got)
					assert.Equal(t, tt.want.Int.String(), got.Int.String())
				}
			}
		})
	}
}

func TestEVMABIUtils_GetUint64FromArgs(t *testing.T) {
	log := mocks.NewNopLogger()
	utils, err := NewEvmAbiUtils(types.ChainTypeEthereum, log)
	require.NoError(t, err)

	uint64Value := uint64(1000)
	bigIntValue := big.NewInt(1000)
	tooBigValue := new(big.Int).Lsh(big.NewInt(1), 70) // Greater than max uint64

	tests := []struct {
		name    string
		args    map[string]interface{}
		key     string
		want    uint64
		wantErr bool
		errCode string
	}{
		{
			name:    "key not found",
			args:    map[string]interface{}{"wrong_key": uint64Value},
			key:     "gas",
			want:    0,
			wantErr: true,
			errCode: errors.ErrCodeABIArgumentNotFound,
		},
		{
			name:    "uint64 value",
			args:    map[string]interface{}{"gas": uint64Value},
			key:     "gas",
			want:    uint64Value,
			wantErr: false,
		},
		{
			name:    "*big.Int value representable as uint64",
			args:    map[string]interface{}{"gas": bigIntValue},
			key:     "gas",
			want:    uint64Value,
			wantErr: false,
		},
		{
			name:    "big.Int too large for uint64",
			args:    map[string]interface{}{"gas": tooBigValue},
			key:     "gas",
			want:    0,
			wantErr: true,
			errCode: errors.ErrCodeABIArgumentInvalidType,
		},
		{
			name:    "invalid type",
			args:    map[string]interface{}{"gas": "not a uint64"},
			key:     "gas",
			want:    0,
			wantErr: true,
			errCode: errors.ErrCodeABIArgumentInvalidType,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := utils.GetUint64FromArgs(tt.args, tt.key)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errCode != "" {
					appErr, ok := err.(*errors.Vault0Error)
					assert.True(t, ok, "Expected a Vault0Error")
					assert.Equal(t, tt.errCode, appErr.Code)
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}
