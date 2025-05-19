package mocks

import (
	"github.com/stretchr/testify/mock"

	"vault0/internal/types"
)

// MockABIUtils implements ABIUtils interface for testing
type MockABIUtils struct {
	mock.Mock
}

func (m *MockABIUtils) Unpack(contractABI string, methodName string, inputData []byte) (map[string]any, error) {
	args := m.Called(contractABI, methodName, inputData)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[string]any), args.Error(1)
}

func (m *MockABIUtils) Pack(contractABI string, methodName string, args ...any) ([]byte, error) {
	mockArgs := m.Called(contractABI, methodName, args)
	if mockArgs.Get(0) == nil {
		return nil, mockArgs.Error(1)
	}
	return mockArgs.Get(0).([]byte), mockArgs.Error(1)
}

func (m *MockABIUtils) ExtractMethodID(data []byte) []byte {
	args := m.Called(data)
	return args.Get(0).([]byte)
}

func (m *MockABIUtils) GetAddressFromArgs(args map[string]any, key string) (types.Address, error) {
	mockArgs := m.Called(args, key)
	return mockArgs.Get(0).(types.Address), mockArgs.Error(1)
}

func (m *MockABIUtils) GetBytes32FromArgs(args map[string]any, key string) ([32]byte, error) {
	mockArgs := m.Called(args, key)
	return mockArgs.Get(0).([32]byte), mockArgs.Error(1)
}

func (m *MockABIUtils) GetBigIntFromArgs(args map[string]any, key string) (*types.BigInt, error) {
	mockArgs := m.Called(args, key)
	if mockArgs.Get(0) == nil {
		return nil, mockArgs.Error(1)
	}
	return mockArgs.Get(0).(*types.BigInt), mockArgs.Error(1)
}

func (m *MockABIUtils) GetUint64FromArgs(args map[string]any, key string) (uint64, error) {
	mockArgs := m.Called(args, key)
	return mockArgs.Get(0).(uint64), mockArgs.Error(1)
}

// WithPackSuccess sets up the mock to return successful data for Pack
func (m *MockABIUtils) WithPackSuccess(contractABI string, methodName string, result []byte) *MockABIUtils {
	m.On("Pack", contractABI, methodName, mock.Anything).Return(result, nil)
	return m
}

// WithUnpackSuccess sets up the mock to return successful data for Unpack
func (m *MockABIUtils) WithUnpackSuccess(contractABI string, methodName string, inputData []byte, result map[string]any) *MockABIUtils {
	m.On("Unpack", contractABI, methodName, inputData).Return(result, nil)
	return m
}

// NewMockABIUtils creates a new instance of MockABIUtils
func NewMockABIUtils() *MockABIUtils {
	return &MockABIUtils{}
}
