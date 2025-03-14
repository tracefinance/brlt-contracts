package transaction

import (
	"context"
	"errors"
	"math/big"
	"testing"
	"time"

	"vault0/internal/config"
	"vault0/internal/core/blockexplorer"
	"vault0/internal/logger"
	"vault0/internal/services/wallet"
	"vault0/internal/types"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockRepository mocks the transaction Repository interface
type MockRepository struct {
	CreateFunc         func(ctx context.Context, tx *Transaction) error
	GetFunc            func(ctx context.Context, chainType types.ChainType, hash string) (*Transaction, error)
	GetByWalletFunc    func(ctx context.Context, walletID string, limit, offset int) ([]*Transaction, error)
	GetByAddressFunc   func(ctx context.Context, chainType types.ChainType, address string, limit, offset int) ([]*Transaction, error)
	CountFunc          func(ctx context.Context, walletID string) (int, error)
	CountByAddressFunc func(ctx context.Context, chainType types.ChainType, address string) (int, error)
	ExistsFunc         func(ctx context.Context, chainType types.ChainType, hash string) (bool, error)
}

func (m *MockRepository) Create(ctx context.Context, tx *Transaction) error {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, tx)
	}
	return nil
}

func (m *MockRepository) Get(ctx context.Context, chainType types.ChainType, hash string) (*Transaction, error) {
	if m.GetFunc != nil {
		return m.GetFunc(ctx, chainType, hash)
	}
	return nil, ErrTransactionNotFound
}

func (m *MockRepository) GetByWallet(ctx context.Context, walletID string, limit, offset int) ([]*Transaction, error) {
	if m.GetByWalletFunc != nil {
		return m.GetByWalletFunc(ctx, walletID, limit, offset)
	}
	return []*Transaction{}, nil
}

func (m *MockRepository) GetByAddress(ctx context.Context, chainType types.ChainType, address string, limit, offset int) ([]*Transaction, error) {
	if m.GetByAddressFunc != nil {
		return m.GetByAddressFunc(ctx, chainType, address, limit, offset)
	}
	return []*Transaction{}, nil
}

func (m *MockRepository) Count(ctx context.Context, walletID string) (int, error) {
	if m.CountFunc != nil {
		return m.CountFunc(ctx, walletID)
	}
	return 0, nil
}

func (m *MockRepository) CountByAddress(ctx context.Context, chainType types.ChainType, address string) (int, error) {
	if m.CountByAddressFunc != nil {
		return m.CountByAddressFunc(ctx, chainType, address)
	}
	return 0, nil
}

func (m *MockRepository) Exists(ctx context.Context, chainType types.ChainType, hash string) (bool, error) {
	if m.ExistsFunc != nil {
		return m.ExistsFunc(ctx, chainType, hash)
	}
	return false, nil
}

// MockWalletService mocks the wallet.Service interface
type MockWalletService struct {
	GetByIDFunc               func(ctx context.Context, id string) (*wallet.Wallet, error)
	GetFunc                   func(ctx context.Context, chainType types.ChainType, address string) (*wallet.Wallet, error)
	UpdateLastBlockNumberFunc func(ctx context.Context, chainType types.ChainType, address string, blockNumber int64) error
}

func (m *MockWalletService) Create(ctx context.Context, chainType types.ChainType, name string, tags map[string]string) (*wallet.Wallet, error) {
	return nil, nil
}

func (m *MockWalletService) Get(ctx context.Context, chainType types.ChainType, address string) (*wallet.Wallet, error) {
	if m.GetFunc != nil {
		return m.GetFunc(ctx, chainType, address)
	}
	return nil, nil
}

func (m *MockWalletService) GetByID(ctx context.Context, id string) (*wallet.Wallet, error) {
	if m.GetByIDFunc != nil {
		return m.GetByIDFunc(ctx, id)
	}
	return nil, nil
}

func (m *MockWalletService) Update(ctx context.Context, chainType types.ChainType, address string, name string, tags map[string]string) (*wallet.Wallet, error) {
	return nil, nil
}

func (m *MockWalletService) UpdateLastBlockNumber(ctx context.Context, chainType types.ChainType, address string, blockNumber int64) error {
	if m.UpdateLastBlockNumberFunc != nil {
		return m.UpdateLastBlockNumberFunc(ctx, chainType, address, blockNumber)
	}
	return nil
}

func (m *MockWalletService) Delete(ctx context.Context, chainType types.ChainType, address string) error {
	return nil
}

func (m *MockWalletService) List(ctx context.Context, limit, offset int) ([]*wallet.Wallet, error) {
	return []*wallet.Wallet{}, nil
}

func (m *MockWalletService) Exists(ctx context.Context, chainType types.ChainType, address string) (bool, error) {
	return false, nil
}

func (m *MockWalletService) SubscribeToBlockchainEvents(ctx context.Context) error {
	return nil
}

func (m *MockWalletService) UnsubscribeFromBlockchainEvents() {
}

func (m *MockWalletService) BlockchainEvents() <-chan *wallet.BlockchainEvent {
	return nil
}

func (m *MockWalletService) LifecycleEvents() <-chan *wallet.LifecycleEvent {
	return nil
}

// MockBlockExplorer mocks the blockexplorer interface
type MockBlockExplorer struct {
	GetTransactionsByHashFunc func(ctx context.Context, hashes []string) ([]*types.Transaction, error)
	GetTransactionHistoryFunc func(ctx context.Context, address string, options blockexplorer.TransactionHistoryOptions) ([]*types.Transaction, error)
	GetAddressBalanceFunc     func(ctx context.Context, address string) (*big.Int, error)
	GetTokenBalancesFunc      func(ctx context.Context, address string) (map[string]*big.Int, error)
	ChainFunc                 func() types.Chain
	CloseFunc                 func() error
}

func (m *MockBlockExplorer) GetTransactionsByHash(ctx context.Context, hashes []string) ([]*types.Transaction, error) {
	if m.GetTransactionsByHashFunc != nil {
		return m.GetTransactionsByHashFunc(ctx, hashes)
	}
	return []*types.Transaction{}, nil
}

func (m *MockBlockExplorer) GetTransactionHistory(ctx context.Context, address string, options blockexplorer.TransactionHistoryOptions) ([]*types.Transaction, error) {
	if m.GetTransactionHistoryFunc != nil {
		return m.GetTransactionHistoryFunc(ctx, address, options)
	}
	return []*types.Transaction{}, nil
}

func (m *MockBlockExplorer) GetAddressBalance(ctx context.Context, address string) (*big.Int, error) {
	if m.GetAddressBalanceFunc != nil {
		return m.GetAddressBalanceFunc(ctx, address)
	}
	return big.NewInt(0), nil
}

func (m *MockBlockExplorer) GetTokenBalances(ctx context.Context, address string) (map[string]*big.Int, error) {
	if m.GetTokenBalancesFunc != nil {
		return m.GetTokenBalancesFunc(ctx, address)
	}
	return make(map[string]*big.Int), nil
}

func (m *MockBlockExplorer) Chain() types.Chain {
	if m.ChainFunc != nil {
		return m.ChainFunc()
	}
	return types.Chain{Type: types.ChainTypeEthereum}
}

func (m *MockBlockExplorer) Close() error {
	if m.CloseFunc != nil {
		return m.CloseFunc()
	}
	return nil
}

// MockBlockExplorerFactory mocks the blockexplorer.Factory interface
type MockBlockExplorerFactory struct {
	GetExplorerFunc func(chainType types.ChainType) (blockexplorer.BlockExplorer, error)
}

func (m *MockBlockExplorerFactory) GetExplorer(chainType types.ChainType) (blockexplorer.BlockExplorer, error) {
	if m.GetExplorerFunc != nil {
		return m.GetExplorerFunc(chainType)
	}
	explorer := &MockBlockExplorer{}
	return explorer, nil
}

// MockTransactionService is a mock implementation of the Service interface
type MockTransactionService struct {
	mock.Mock
}

func (m *MockTransactionService) GetTransaction(ctx context.Context, chainType types.ChainType, hash string) (*Transaction, error) {
	args := m.Called(ctx, chainType, hash)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Transaction), args.Error(1)
}

func (m *MockTransactionService) GetTransactionsByWallet(ctx context.Context, walletID string, limit, offset int) ([]*Transaction, error) {
	args := m.Called(ctx, walletID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*Transaction), args.Error(1)
}

func (m *MockTransactionService) GetTransactionsByAddress(ctx context.Context, chainType types.ChainType, address string, limit, offset int) ([]*Transaction, error) {
	args := m.Called(ctx, chainType, address, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*Transaction), args.Error(1)
}

func (m *MockTransactionService) SyncTransactions(ctx context.Context, walletID string) (int, error) {
	args := m.Called(ctx, walletID)
	return args.Int(0), args.Error(1)
}

func (m *MockTransactionService) SyncTransactionsByAddress(ctx context.Context, chainType types.ChainType, address string) (int, error) {
	args := m.Called(ctx, chainType, address)
	return args.Int(0), args.Error(1)
}

func (m *MockTransactionService) CountTransactions(ctx context.Context, walletID string) (int, error) {
	args := m.Called(ctx, walletID)
	return args.Int(0), args.Error(1)
}

func (m *MockTransactionService) SubscribeToWalletEvents(ctx context.Context) {
	m.Called(ctx)
}

func (m *MockTransactionService) UnsubscribeFromWalletEvents() {
	m.Called()
}

// TestGetTransaction tests the GetTransaction method
func TestGetTransaction(t *testing.T) {
	// Setup minimal config and chains
	cfg := &config.Config{}
	chains := types.Chains{
		types.ChainTypeEthereum: types.Chain{
			Type: types.ChainTypeEthereum,
		},
	}

	// Define test cases
	tests := []struct {
		name        string
		chainType   types.ChainType
		hash        string
		setupMocks  func(*MockRepository, *MockBlockExplorerFactory)
		wantErr     bool
		errContains string
	}{
		{
			name:      "successful retrieval from database",
			chainType: types.ChainTypeEthereum,
			hash:      "0x1234567890abcdef1234567890abcdef12345678",
			setupMocks: func(repo *MockRepository, factory *MockBlockExplorerFactory) {
				repo.GetFunc = func(ctx context.Context, chainType types.ChainType, hash string) (*Transaction, error) {
					assert.Equal(t, types.ChainTypeEthereum, chainType)
					assert.Equal(t, "0x1234567890abcdef1234567890abcdef12345678", hash)
					return &Transaction{
						ID:          "tx123",
						ChainType:   chainType,
						Hash:        hash,
						FromAddress: "0xsender",
						ToAddress:   "0xrecipient",
						Value:       big.NewInt(1000000000000000000), // 1 ETH
					}, nil
				}
			},
			wantErr: false,
		},
		{
			name:      "successful retrieval from explorer",
			chainType: types.ChainTypeEthereum,
			hash:      "0x1234567890abcdef1234567890abcdef12345678",
			setupMocks: func(repo *MockRepository, factory *MockBlockExplorerFactory) {
				repo.GetFunc = func(ctx context.Context, chainType types.ChainType, hash string) (*Transaction, error) {
					return nil, ErrTransactionNotFound
				}

				mockExplorer := &MockBlockExplorer{
					GetTransactionsByHashFunc: func(ctx context.Context, hashes []string) ([]*types.Transaction, error) {
						assert.Equal(t, []string{"0x1234567890abcdef1234567890abcdef12345678"}, hashes)
						return []*types.Transaction{
							{
								Chain:     types.ChainTypeEthereum,
								Hash:      "0x1234567890abcdef1234567890abcdef12345678",
								From:      "0xsender",
								To:        "0xrecipient",
								Value:     big.NewInt(1000000000000000000), // 1 ETH
								Timestamp: time.Now().Unix(),
							},
						}, nil
					},
				}

				factory.GetExplorerFunc = func(chainType types.ChainType) (blockexplorer.BlockExplorer, error) {
					assert.Equal(t, types.ChainTypeEthereum, chainType)
					var explorer blockexplorer.BlockExplorer = mockExplorer
					return explorer, nil
				}

				repo.CreateFunc = func(ctx context.Context, tx *Transaction) error {
					assert.Equal(t, types.ChainTypeEthereum, tx.ChainType)
					assert.Equal(t, "0x1234567890abcdef1234567890abcdef12345678", tx.Hash)
					return nil
				}
			},
			wantErr: false,
		},
		{
			name:      "transaction not found in database or explorer",
			chainType: types.ChainTypeEthereum,
			hash:      "0x1234567890abcdef1234567890abcdef12345678",
			setupMocks: func(repo *MockRepository, factory *MockBlockExplorerFactory) {
				repo.GetFunc = func(ctx context.Context, chainType types.ChainType, hash string) (*Transaction, error) {
					return nil, ErrTransactionNotFound
				}

				mockExplorer := &MockBlockExplorer{
					GetTransactionsByHashFunc: func(ctx context.Context, hashes []string) ([]*types.Transaction, error) {
						return []*types.Transaction{}, nil
					},
				}

				factory.GetExplorerFunc = func(chainType types.ChainType) (blockexplorer.BlockExplorer, error) {
					var explorer blockexplorer.BlockExplorer = mockExplorer
					return explorer, nil
				}
			},
			wantErr:     true,
			errContains: "transaction not found",
		},
		{
			name:      "explorer error",
			chainType: types.ChainTypeEthereum,
			hash:      "0x1234567890abcdef1234567890abcdef12345678",
			setupMocks: func(repo *MockRepository, factory *MockBlockExplorerFactory) {
				repo.GetFunc = func(ctx context.Context, chainType types.ChainType, hash string) (*Transaction, error) {
					return nil, ErrTransactionNotFound
				}

				mockExplorer := &MockBlockExplorer{
					GetTransactionsByHashFunc: func(ctx context.Context, hashes []string) ([]*types.Transaction, error) {
						return nil, errors.New("explorer error")
					},
				}

				factory.GetExplorerFunc = func(chainType types.ChainType) (blockexplorer.BlockExplorer, error) {
					var explorer blockexplorer.BlockExplorer = mockExplorer
					return explorer, nil
				}
			},
			wantErr:     true,
			errContains: "failed to get transaction from explorer",
		},
		{
			name:      "empty chain type",
			chainType: "",
			hash:      "0x1234567890abcdef1234567890abcdef12345678",
			setupMocks: func(repo *MockRepository, factory *MockBlockExplorerFactory) {
				// No mocks needed; validation fails before repository is called
			},
			wantErr:     true,
			errContains: "chain type is required",
		},
		{
			name:      "empty hash",
			chainType: types.ChainTypeEthereum,
			hash:      "",
			setupMocks: func(repo *MockRepository, factory *MockBlockExplorerFactory) {
				// No mocks needed; validation fails before repository is called
			},
			wantErr:     true,
			errContains: "hash is required",
		},
	}

	// Run tests
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Initialize mocks
			mockRepository := &MockRepository{}
			mockWalletService := &MockWalletService{}
			mockBlockExplorerFactory := &MockBlockExplorerFactory{}

			// Setup mock behavior
			tt.setupMocks(mockRepository, mockBlockExplorerFactory)

			// Create service
			s := NewService(cfg, logger.NewNopLogger(), mockRepository, mockWalletService, mockBlockExplorerFactory, chains)

			// Call the service method
			tx, err := s.GetTransaction(context.Background(), tt.chainType, tt.hash)

			// Assertions
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
				assert.Nil(t, tx)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, tx)
				assert.Equal(t, tt.chainType, tx.ChainType)
				assert.Equal(t, tt.hash, tx.Hash)
			}
		})
	}
}

// TestGetTransactionsByWallet tests the GetTransactionsByWallet method
func TestGetTransactionsByWallet(t *testing.T) {
	// Setup minimal config and chains
	cfg := &config.Config{}
	chains := types.Chains{
		types.ChainTypeEthereum: types.Chain{
			Type: types.ChainTypeEthereum,
		},
	}

	// Define test cases
	tests := []struct {
		name        string
		walletID    string
		limit       int
		offset      int
		setupMocks  func(*MockRepository)
		wantErr     bool
		errContains string
		wantCount   int
	}{
		{
			name:     "successful retrieval",
			walletID: "wallet123",
			limit:    10,
			offset:   0,
			setupMocks: func(repo *MockRepository) {
				repo.GetByWalletFunc = func(ctx context.Context, walletID string, limit, offset int) ([]*Transaction, error) {
					assert.Equal(t, "wallet123", walletID)
					assert.Equal(t, 10, limit)
					assert.Equal(t, 0, offset)
					return []*Transaction{
						{
							ID:        "tx1",
							WalletID:  walletID,
							ChainType: types.ChainTypeEthereum,
							Hash:      "0xhash1",
						},
						{
							ID:        "tx2",
							WalletID:  walletID,
							ChainType: types.ChainTypeEthereum,
							Hash:      "0xhash2",
						},
					}, nil
				}
			},
			wantErr:   false,
			wantCount: 2,
		},
		{
			name:     "empty result",
			walletID: "wallet123",
			limit:    10,
			offset:   0,
			setupMocks: func(repo *MockRepository) {
				repo.GetByWalletFunc = func(ctx context.Context, walletID string, limit, offset int) ([]*Transaction, error) {
					return []*Transaction{}, nil
				}
			},
			wantErr:   false,
			wantCount: 0,
		},
		{
			name:     "negative limit converted to default",
			walletID: "wallet123",
			limit:    -5,
			offset:   0,
			setupMocks: func(repo *MockRepository) {
				repo.GetByWalletFunc = func(ctx context.Context, walletID string, limit, offset int) ([]*Transaction, error) {
					assert.Equal(t, 10, limit) // Default value
					return []*Transaction{}, nil
				}
			},
			wantErr:   false,
			wantCount: 0,
		},
		{
			name:     "negative offset converted to default",
			walletID: "wallet123",
			limit:    10,
			offset:   -5,
			setupMocks: func(repo *MockRepository) {
				repo.GetByWalletFunc = func(ctx context.Context, walletID string, limit, offset int) ([]*Transaction, error) {
					assert.Equal(t, 0, offset) // Default value
					return []*Transaction{}, nil
				}
			},
			wantErr:   false,
			wantCount: 0,
		},
		{
			name:     "empty wallet ID",
			walletID: "",
			limit:    10,
			offset:   0,
			setupMocks: func(repo *MockRepository) {
				// No mocks needed; validation fails before repository is called
			},
			wantErr:     true,
			errContains: "wallet ID is required",
			wantCount:   0,
		},
		{
			name:     "repository error",
			walletID: "wallet123",
			limit:    10,
			offset:   0,
			setupMocks: func(repo *MockRepository) {
				repo.GetByWalletFunc = func(ctx context.Context, walletID string, limit, offset int) ([]*Transaction, error) {
					return nil, errors.New("database error")
				}
			},
			wantErr:     true,
			errContains: "database error",
			wantCount:   0,
		},
	}

	// Run tests
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Initialize mocks
			mockRepository := &MockRepository{}
			mockWalletService := &MockWalletService{}
			mockBlockExplorerFactory := &MockBlockExplorerFactory{}

			// Setup mock behavior
			tt.setupMocks(mockRepository)

			// Create service
			s := NewService(cfg, logger.NewNopLogger(), mockRepository, mockWalletService, mockBlockExplorerFactory, chains)

			// Call the service method
			txs, err := s.GetTransactionsByWallet(context.Background(), tt.walletID, tt.limit, tt.offset)

			// Assertions
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
				assert.Nil(t, txs)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, txs)
				assert.Equal(t, tt.wantCount, len(txs))
			}
		})
	}
}

// TestSyncTransactions tests the SyncTransactions method
func TestSyncTransactions(t *testing.T) {
	// Setup minimal config and chains
	cfg := &config.Config{}
	chains := types.Chains{
		types.ChainTypeEthereum: types.Chain{
			Type: types.ChainTypeEthereum,
		},
	}

	// Define test cases
	tests := []struct {
		name        string
		walletID    string
		setupMocks  func(*MockRepository, *MockWalletService, *MockBlockExplorerFactory)
		wantErr     bool
		errContains string
		wantCount   int
	}{
		{
			name:     "successful sync",
			walletID: "wallet123",
			setupMocks: func(repo *MockRepository, walletSvc *MockWalletService, factory *MockBlockExplorerFactory) {
				walletSvc.GetByIDFunc = func(ctx context.Context, id string) (*wallet.Wallet, error) {
					assert.Equal(t, "wallet123", id)
					return &wallet.Wallet{
						ID:        "wallet123",
						ChainType: types.ChainTypeEthereum,
						Address:   "0xaddress",
					}, nil
				}

				walletSvc.GetFunc = func(ctx context.Context, chainType types.ChainType, address string) (*wallet.Wallet, error) {
					assert.Equal(t, types.ChainTypeEthereum, chainType)
					assert.Equal(t, "0xaddress", address)
					return &wallet.Wallet{
						ID:        "wallet123",
						ChainType: types.ChainTypeEthereum,
						Address:   "0xaddress",
					}, nil
				}

				mockExplorer := &MockBlockExplorer{
					GetTransactionHistoryFunc: func(ctx context.Context, address string, options blockexplorer.TransactionHistoryOptions) ([]*types.Transaction, error) {
						assert.Equal(t, "0xaddress", address)
						return []*types.Transaction{
							{
								Chain:     types.ChainTypeEthereum,
								Hash:      "0xhash1",
								From:      "0xsender",
								To:        "0xaddress",
								Value:     big.NewInt(1000000000000000000),
								Timestamp: time.Now().Unix(),
								Type:      types.TransactionTypeNative,
								Status:    "success",
							},
							{
								Chain:     types.ChainTypeEthereum,
								Hash:      "0xhash2",
								From:      "0xaddress",
								To:        "0xrecipient",
								Value:     big.NewInt(500000000000000000),
								Timestamp: time.Now().Unix(),
								Type:      types.TransactionTypeNative,
								Status:    "success",
							},
						}, nil
					},
				}

				factory.GetExplorerFunc = func(chainType types.ChainType) (blockexplorer.BlockExplorer, error) {
					assert.Equal(t, types.ChainTypeEthereum, chainType)
					var explorer blockexplorer.BlockExplorer = mockExplorer
					return explorer, nil
				}

				repo.ExistsFunc = func(ctx context.Context, chainType types.ChainType, hash string) (bool, error) {
					return false, nil // Transactions don't exist yet
				}

				repo.CreateFunc = func(ctx context.Context, tx *Transaction) error {
					assert.Equal(t, "wallet123", tx.WalletID)
					assert.Equal(t, types.ChainTypeEthereum, tx.ChainType)
					assert.NotEmpty(t, tx.Hash)
					assert.NotNil(t, tx.Value)
					return nil
				}
			},
			wantErr:   false,
			wantCount: 2,
		},
		{
			name:     "wallet not found",
			walletID: "nonexistent",
			setupMocks: func(repo *MockRepository, walletSvc *MockWalletService, factory *MockBlockExplorerFactory) {
				walletSvc.GetByIDFunc = func(ctx context.Context, id string) (*wallet.Wallet, error) {
					return nil, wallet.ErrWalletNotFound
				}
			},
			wantErr:     true,
			errContains: "wallet not found",
			wantCount:   0,
		},
		{
			name:     "wallet service error",
			walletID: "wallet123",
			setupMocks: func(repo *MockRepository, walletSvc *MockWalletService, factory *MockBlockExplorerFactory) {
				walletSvc.GetByIDFunc = func(ctx context.Context, id string) (*wallet.Wallet, error) {
					return nil, errors.New("wallet service error")
				}
			},
			wantErr:     true,
			errContains: "wallet service error",
			wantCount:   0,
		},
		{
			name:     "explorer error",
			walletID: "wallet123",
			setupMocks: func(repo *MockRepository, walletSvc *MockWalletService, factory *MockBlockExplorerFactory) {
				walletSvc.GetByIDFunc = func(ctx context.Context, id string) (*wallet.Wallet, error) {
					return &wallet.Wallet{
						ID:        "wallet123",
						ChainType: types.ChainTypeEthereum,
						Address:   "0xaddress",
					}, nil
				}

				walletSvc.GetFunc = func(ctx context.Context, chainType types.ChainType, address string) (*wallet.Wallet, error) {
					assert.Equal(t, types.ChainTypeEthereum, chainType)
					assert.Equal(t, "0xaddress", address)
					return &wallet.Wallet{
						ID:        "wallet123",
						ChainType: types.ChainTypeEthereum,
						Address:   "0xaddress",
					}, nil
				}

				mockExplorer := &MockBlockExplorer{
					GetTransactionHistoryFunc: func(ctx context.Context, address string, options blockexplorer.TransactionHistoryOptions) ([]*types.Transaction, error) {
						return nil, errors.New("explorer error")
					},
				}

				factory.GetExplorerFunc = func(chainType types.ChainType) (blockexplorer.BlockExplorer, error) {
					var explorer blockexplorer.BlockExplorer = mockExplorer
					return explorer, nil
				}
			},
			wantErr:     true,
			errContains: "failed to get transaction history",
			wantCount:   0,
		},
		{
			name:     "transactions already exist",
			walletID: "wallet123",
			setupMocks: func(repo *MockRepository, walletSvc *MockWalletService, factory *MockBlockExplorerFactory) {
				walletSvc.GetByIDFunc = func(ctx context.Context, id string) (*wallet.Wallet, error) {
					return &wallet.Wallet{
						ID:        "wallet123",
						ChainType: types.ChainTypeEthereum,
						Address:   "0xaddress",
					}, nil
				}

				walletSvc.GetFunc = func(ctx context.Context, chainType types.ChainType, address string) (*wallet.Wallet, error) {
					assert.Equal(t, types.ChainTypeEthereum, chainType)
					assert.Equal(t, "0xaddress", address)
					return &wallet.Wallet{
						ID:        "wallet123",
						ChainType: types.ChainTypeEthereum,
						Address:   "0xaddress",
					}, nil
				}

				mockExplorer := &MockBlockExplorer{
					GetTransactionHistoryFunc: func(ctx context.Context, address string, options blockexplorer.TransactionHistoryOptions) ([]*types.Transaction, error) {
						return []*types.Transaction{
							{
								Chain:     types.ChainTypeEthereum,
								Hash:      "0xhash1",
								From:      "0xsender",
								To:        "0xaddress",
								Value:     big.NewInt(1000000000000000000),
								Timestamp: time.Now().Unix(),
								Type:      types.TransactionTypeNative,
								Status:    "success",
							},
						}, nil
					},
				}

				factory.GetExplorerFunc = func(chainType types.ChainType) (blockexplorer.BlockExplorer, error) {
					var explorer blockexplorer.BlockExplorer = mockExplorer
					return explorer, nil
				}

				repo.ExistsFunc = func(ctx context.Context, chainType types.ChainType, hash string) (bool, error) {
					return true, nil // Transaction already exists
				}
			},
			wantErr:   false,
			wantCount: 0, // No new transactions added
		},
	}

	// Run tests
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Initialize mocks
			mockRepository := &MockRepository{}
			mockWalletService := &MockWalletService{}
			mockBlockExplorerFactory := &MockBlockExplorerFactory{}

			// Setup mock behavior
			tt.setupMocks(mockRepository, mockWalletService, mockBlockExplorerFactory)

			// Create service
			s := NewService(cfg, logger.NewNopLogger(), mockRepository, mockWalletService, mockBlockExplorerFactory, chains)

			// Call the service method
			count, err := s.SyncTransactions(context.Background(), tt.walletID)

			// Assertions
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
				assert.Equal(t, 0, count)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantCount, count)
			}
		})
	}
}

// TestCountTransactions tests the CountTransactions method
func TestCountTransactions(t *testing.T) {
	// Setup minimal config and chains
	cfg := &config.Config{}
	chains := types.Chains{
		types.ChainTypeEthereum: types.Chain{
			Type: types.ChainTypeEthereum,
		},
	}

	// Define test cases
	tests := []struct {
		name        string
		walletID    string
		setupMocks  func(*MockRepository)
		wantErr     bool
		errContains string
		wantCount   int
	}{
		{
			name:     "successful count",
			walletID: "wallet123",
			setupMocks: func(repo *MockRepository) {
				repo.CountFunc = func(ctx context.Context, walletID string) (int, error) {
					assert.Equal(t, "wallet123", walletID)
					return 42, nil
				}
			},
			wantErr:   false,
			wantCount: 42,
		},
		{
			name:     "empty wallet ID",
			walletID: "",
			setupMocks: func(repo *MockRepository) {
				// No mocks needed; validation fails before repository is called
			},
			wantErr:     true,
			errContains: "wallet ID is required",
			wantCount:   0,
		},
		{
			name:     "repository error",
			walletID: "wallet123",
			setupMocks: func(repo *MockRepository) {
				repo.CountFunc = func(ctx context.Context, walletID string) (int, error) {
					return 0, errors.New("database error")
				}
			},
			wantErr:     true,
			errContains: "database error",
			wantCount:   0,
		},
	}

	// Run tests
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Initialize mocks
			mockRepository := &MockRepository{}
			mockWalletService := &MockWalletService{}
			mockBlockExplorerFactory := &MockBlockExplorerFactory{}

			// Setup mock behavior
			tt.setupMocks(mockRepository)

			// Create service
			s := NewService(cfg, logger.NewNopLogger(), mockRepository, mockWalletService, mockBlockExplorerFactory, chains)

			// Call the service method
			count, err := s.CountTransactions(context.Background(), tt.walletID)

			// Assertions
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
				assert.Equal(t, 0, count)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantCount, count)
			}
		})
	}
}
