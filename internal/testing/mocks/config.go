package mocks

import (
	"github.com/stretchr/testify/mock"
)

// MockConfig implements a mock configuration for testing
type MockConfig struct {
	mock.Mock
}

func (m *MockConfig) GetArtifactPathForType(abiType string) (string, error) {
	args := m.Called(abiType)
	return args.String(0), args.Error(1)
}

// WithValidArtifactPath sets up the mock to return a valid path for a given ABI type
func (m *MockConfig) WithValidArtifactPath(abiType string, path string) *MockConfig {
	m.On("GetArtifactPathForType", abiType).Return(path, nil)
	return m
}

// WithInvalidArtifactPath sets up the mock to return an error for a given ABI type
func (m *MockConfig) WithInvalidArtifactPath(abiType string, err error) *MockConfig {
	m.On("GetArtifactPathForType", abiType).Return("", err)
	return m
}

// NewMockConfig creates a new MockConfig instance
func NewMockConfig() *MockConfig {
	return &MockConfig{}
}
