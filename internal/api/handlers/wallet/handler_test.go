package wallet

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"vault0/internal/services/wallet"
	"vault0/internal/types"
)

// MockWalletService is a mock implementation of the wallet.Service interface
type MockWalletService struct {
	mock.Mock
}

func (m *MockWalletService) Create(ctx context.Context, chainType types.ChainType, name string, tags map[string]string) (*wallet.Wallet, error) {
	args := m.Called(ctx, chainType, name, tags)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*wallet.Wallet), args.Error(1)
}

func (m *MockWalletService) Update(ctx context.Context, chainType types.ChainType, address, name string, tags map[string]string) (*wallet.Wallet, error) {
	args := m.Called(ctx, chainType, address, name, tags)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*wallet.Wallet), args.Error(1)
}

func (m *MockWalletService) Delete(ctx context.Context, chainType types.ChainType, address string) error {
	args := m.Called(ctx, chainType, address)
	return args.Error(0)
}

func (m *MockWalletService) Get(ctx context.Context, chainType types.ChainType, address string) (*wallet.Wallet, error) {
	args := m.Called(ctx, chainType, address)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*wallet.Wallet), args.Error(1)
}

func (m *MockWalletService) List(ctx context.Context, limit, offset int) ([]*wallet.Wallet, error) {
	args := m.Called(ctx, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*wallet.Wallet), args.Error(1)
}

func (m *MockWalletService) SubscribeToEvents(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockWalletService) UnsubscribeFromEvents() {
	m.Called()
}

// setupTestRouter creates a new router and mock service for each test
func setupTestRouter() (*gin.Engine, *MockWalletService, *Handler) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	mockService := new(MockWalletService)
	handler := NewHandler(mockService)
	return router, mockService, handler
}

func TestGetWalletByChainTypeAndAddressParams(t *testing.T) {
	tests := []struct {
		name           string
		chainType      string
		address        string
		setupMock      func(*MockWalletService)
		expectedStatus int
		expectedBody   string
	}{
		{
			name:      "successful retrieval",
			chainType: "ethereum",
			address:   "0x1234567890abcdef1234567890abcdef12345678",
			setupMock: func(mockService *MockWalletService) {
				mockWallet := &wallet.Wallet{
					ID:        "wallet123",
					ChainType: types.ChainTypeEthereum,
					Address:   "0x1234567890abcdef1234567890abcdef12345678",
					Name:      "Test Wallet",
					Tags:      map[string]string{"tag1": "value1"},
				}
				mockService.On("Get", mock.Anything, types.ChainType("ethereum"), "0x1234567890abcdef1234567890abcdef12345678").Return(mockWallet, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:      "wallet not found",
			chainType: "ethereum",
			address:   "0x1234567890abcdef1234567890abcdef12345678",
			setupMock: func(mockService *MockWalletService) {
				mockService.On("Get", mock.Anything, types.ChainType("ethereum"), "0x1234567890abcdef1234567890abcdef12345678").Return(nil, wallet.ErrWalletNotFound)
			},
			expectedStatus: http.StatusNotFound,
			expectedBody:   `{"error":"Wallet not found"}`,
		},
		{
			name:      "invalid input",
			chainType: "ethereum",
			address:   "invalid-address",
			setupMock: func(mockService *MockWalletService) {
				mockService.On("Get", mock.Anything, types.ChainType("ethereum"), "invalid-address").Return(nil, wallet.ErrInvalidInput)
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"invalid input"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup a fresh router and mock for each test
			router, mockService, handler := setupTestRouter()
			api := router.Group("/api")
			handler.SetupRoutes(api)

			// Setup mock behavior
			tt.setupMock(mockService)

			// Create request
			req, _ := http.NewRequest("GET", fmt.Sprintf("/api/wallets/%s/%s", tt.chainType, tt.address), nil)

			// Perform request
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			// Check status code
			assert.Equal(t, tt.expectedStatus, w.Code)

			// Check response body if expected
			if tt.expectedBody != "" {
				assert.JSONEq(t, tt.expectedBody, w.Body.String())
			}

			// If successful, verify the response contains wallet data
			if tt.expectedStatus == http.StatusOK {
				var response WalletResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, "wallet123", response.ID)
				assert.Equal(t, types.ChainTypeEthereum, response.ChainType)
				assert.Equal(t, "0x1234567890abcdef1234567890abcdef12345678", response.Address)
			}

			mockService.AssertExpectations(t)
		})
	}
}

func TestUpdateWalletByChainTypeAndAddress(t *testing.T) {
	tests := []struct {
		name           string
		chainType      string
		address        string
		requestBody    UpdateWalletRequest
		setupMock      func(*MockWalletService)
		expectedStatus int
		expectedBody   string
	}{
		{
			name:      "successful update",
			chainType: "ethereum",
			address:   "0x1234567890abcdef1234567890abcdef12345678",
			requestBody: UpdateWalletRequest{
				Name: "Updated Wallet",
				Tags: map[string]string{"tag1": "updated-value"},
			},
			setupMock: func(mockService *MockWalletService) {
				mockWallet := &wallet.Wallet{
					ID:        "wallet123",
					ChainType: types.ChainTypeEthereum,
					Address:   "0x1234567890abcdef1234567890abcdef12345678",
					Name:      "Updated Wallet",
					Tags:      map[string]string{"tag1": "updated-value"},
				}
				mockService.On("Update", mock.Anything, types.ChainType("ethereum"), "0x1234567890abcdef1234567890abcdef12345678", "Updated Wallet", map[string]string{"tag1": "updated-value"}).Return(mockWallet, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:      "wallet not found",
			chainType: "ethereum",
			address:   "0x1234567890abcdef1234567890abcdef12345678",
			requestBody: UpdateWalletRequest{
				Name: "Updated Wallet",
				Tags: map[string]string{"tag1": "updated-value"},
			},
			setupMock: func(mockService *MockWalletService) {
				mockService.On("Update", mock.Anything, types.ChainType("ethereum"), "0x1234567890abcdef1234567890abcdef12345678", "Updated Wallet", map[string]string{"tag1": "updated-value"}).Return(nil, wallet.ErrWalletNotFound)
			},
			expectedStatus: http.StatusNotFound,
			expectedBody:   `{"error":"Wallet not found"}`,
		},
		{
			name:      "invalid input",
			chainType: "ethereum",
			address:   "invalid-address",
			requestBody: UpdateWalletRequest{
				Name: "Updated Wallet",
				Tags: map[string]string{"tag1": "updated-value"},
			},
			setupMock: func(mockService *MockWalletService) {
				mockService.On("Update", mock.Anything, types.ChainType("ethereum"), "invalid-address", "Updated Wallet", map[string]string{"tag1": "updated-value"}).Return(nil, wallet.ErrInvalidInput)
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"invalid input"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup a fresh router and mock for each test
			router, mockService, handler := setupTestRouter()
			api := router.Group("/api")
			handler.SetupRoutes(api)

			// Setup mock behavior
			tt.setupMock(mockService)

			// Convert request body to JSON
			jsonBody, _ := json.Marshal(tt.requestBody)
			req, _ := http.NewRequest("PUT", fmt.Sprintf("/api/wallets/%s/%s", tt.chainType, tt.address), bytes.NewBuffer(jsonBody))
			req.Header.Set("Content-Type", "application/json")

			// Perform request
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			// Check status code
			assert.Equal(t, tt.expectedStatus, w.Code)

			// Check response body if expected
			if tt.expectedBody != "" {
				assert.JSONEq(t, tt.expectedBody, w.Body.String())
			}

			// If successful, verify the response contains wallet data
			if tt.expectedStatus == http.StatusOK {
				var response WalletResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, "wallet123", response.ID)
				assert.Equal(t, types.ChainTypeEthereum, response.ChainType)
				assert.Equal(t, "0x1234567890abcdef1234567890abcdef12345678", response.Address)
				assert.Equal(t, "Updated Wallet", response.Name)
				assert.Equal(t, map[string]string{"tag1": "updated-value"}, response.Tags)
			}

			mockService.AssertExpectations(t)
		})
	}
}

func TestDeleteWalletByChainTypeAndAddress(t *testing.T) {
	tests := []struct {
		name           string
		chainType      string
		address        string
		setupMock      func(*MockWalletService)
		expectedStatus int
		expectedBody   string
	}{
		{
			name:      "successful deletion",
			chainType: "ethereum",
			address:   "0x1234567890abcdef1234567890abcdef12345678",
			setupMock: func(mockService *MockWalletService) {
				mockService.On("Delete", mock.Anything, types.ChainType("ethereum"), "0x1234567890abcdef1234567890abcdef12345678").Return(nil)
			},
			expectedStatus: http.StatusNoContent,
		},
		{
			name:      "wallet not found",
			chainType: "ethereum",
			address:   "0x1234567890abcdef1234567890abcdef12345678",
			setupMock: func(mockService *MockWalletService) {
				mockService.On("Delete", mock.Anything, types.ChainType("ethereum"), "0x1234567890abcdef1234567890abcdef12345678").Return(wallet.ErrWalletNotFound)
			},
			expectedStatus: http.StatusNotFound,
			expectedBody:   `{"error":"Wallet not found"}`,
		},
		{
			name:      "invalid input",
			chainType: "ethereum",
			address:   "invalid-address",
			setupMock: func(mockService *MockWalletService) {
				mockService.On("Delete", mock.Anything, types.ChainType("ethereum"), "invalid-address").Return(wallet.ErrInvalidInput)
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"invalid input"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup a fresh router and mock for each test
			router, mockService, handler := setupTestRouter()
			api := router.Group("/api")
			handler.SetupRoutes(api)

			// Setup mock behavior
			tt.setupMock(mockService)

			// Create request
			req, _ := http.NewRequest("DELETE", fmt.Sprintf("/api/wallets/%s/%s", tt.chainType, tt.address), nil)

			// Perform request
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			// Check status code
			assert.Equal(t, tt.expectedStatus, w.Code)

			// Check response body if expected
			if tt.expectedBody != "" {
				assert.JSONEq(t, tt.expectedBody, w.Body.String())
			}

			mockService.AssertExpectations(t)
		})
	}
}
