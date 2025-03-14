package wallet

import (
	"context"
	"crypto/elliptic"
	"database/sql"
	"math/big"
	"testing"

	"vault0/internal/config"
	"vault0/internal/core/blockchain"
	"vault0/internal/core/keystore"
	coreWallet "vault0/internal/core/wallet"
	"vault0/internal/logger"
	"vault0/internal/types"

	"github.com/stretchr/testify/assert"
)

// MockKeyStore mocks keystore.KeyStore
type MockKeyStore struct {
	CreateFunc func(ctx context.Context, name string, keyType types.KeyType, curve elliptic.Curve, tags map[string]string) (*keystore.Key, error)
}

func (m *MockKeyStore) Create(ctx context.Context, name string, keyType types.KeyType, curve elliptic.Curve, tags map[string]string) (*keystore.Key, error) {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, name, keyType, curve, tags)
	}
	return &keystore.Key{ID: "key123"}, nil
}

func (m *MockKeyStore) Delete(ctx context.Context, id string) error {
	return nil // No-op for testing
}

func (m *MockKeyStore) GetPublicKey(ctx context.Context, id string) (*keystore.Key, error) {
	return &keystore.Key{ID: id}, nil
}

func (m *MockKeyStore) Import(ctx context.Context, name string, keyType types.KeyType, curve elliptic.Curve, privateKey []byte, publicKey []byte, tags map[string]string) (*keystore.Key, error) {
	return &keystore.Key{ID: "imported-key"}, nil
}

func (m *MockKeyStore) List(ctx context.Context) ([]*keystore.Key, error) {
	return nil, nil
}

func (m *MockKeyStore) Sign(ctx context.Context, keyID string, data []byte, dataType keystore.DataType) ([]byte, error) {
	return []byte("mock-signature"), nil
}

func (m *MockKeyStore) Update(ctx context.Context, id string, name string, tags map[string]string) (*keystore.Key, error) {
	return &keystore.Key{ID: id, Name: name}, nil
}

// MockWalletFactory mocks coreWallet.Factory
type MockWalletFactory struct {
	NewWalletFunc func(ctx context.Context, chainType types.ChainType, keyID string) (coreWallet.Wallet, error)
}

func (m *MockWalletFactory) NewWallet(ctx context.Context, chainType types.ChainType, keyID string) (coreWallet.Wallet, error) {
	if m.NewWalletFunc != nil {
		return m.NewWalletFunc(ctx, chainType, keyID)
	}
	return &MockWallet{}, nil
}

// MockWallet mocks coreWallet.Wallet
type MockWallet struct {
	DeriveAddressFunc func(ctx context.Context) (string, error)
}

func (m *MockWallet) DeriveAddress(ctx context.Context) (string, error) {
	if m.DeriveAddressFunc != nil {
		return m.DeriveAddressFunc(ctx)
	}
	return "0x1234567890", nil
}

func (m *MockWallet) Chain() types.Chain {
	return types.Chain{}
}

func (m *MockWallet) CreateNativeTransaction(ctx context.Context, toAddress string, amount *big.Int, options types.TransactionOptions) (*types.Transaction, error) {
	return nil, nil
}

func (m *MockWallet) CreateTokenTransaction(ctx context.Context, tokenAddress, toAddress string, amount *big.Int, options types.TransactionOptions) (*types.Transaction, error) {
	return nil, nil
}

func (m *MockWallet) SignTransaction(ctx context.Context, tx *types.Transaction) ([]byte, error) {
	return nil, nil
}

// MockRepository mocks Repository
type MockRepository struct {
	CreateFunc  func(ctx context.Context, wallet *Wallet) error
	GetFunc     func(ctx context.Context, chainType types.ChainType, address string) (*Wallet, error)
	GetByIDFunc func(ctx context.Context, id string) (*Wallet, error)
	UpdateFunc  func(ctx context.Context, wallet *Wallet) error
	DeleteFunc  func(ctx context.Context, chainType types.ChainType, address string) error
	ListFunc    func(ctx context.Context, limit, offset int) ([]*Wallet, error)
	ExistsFunc  func(ctx context.Context, chainType types.ChainType, address string) (bool, error)
}

func (m *MockRepository) Create(ctx context.Context, wallet *Wallet) error {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, wallet)
	}
	return nil
}

func (m *MockRepository) Get(ctx context.Context, chainType types.ChainType, address string) (*Wallet, error) {
	if m.GetFunc != nil {
		return m.GetFunc(ctx, chainType, address)
	}
	return &Wallet{ChainType: chainType, Address: address}, nil
}

func (m *MockRepository) GetByID(ctx context.Context, id string) (*Wallet, error) {
	if m.GetByIDFunc != nil {
		return m.GetByIDFunc(ctx, id)
	}
	return &Wallet{ID: id}, nil
}

func (m *MockRepository) Update(ctx context.Context, wallet *Wallet) error {
	if m.UpdateFunc != nil {
		return m.UpdateFunc(ctx, wallet)
	}
	return nil
}

func (m *MockRepository) Delete(ctx context.Context, chainType types.ChainType, address string) error {
	if m.DeleteFunc != nil {
		return m.DeleteFunc(ctx, chainType, address)
	}
	return nil
}

func (m *MockRepository) List(ctx context.Context, limit, offset int) ([]*Wallet, error) {
	if m.ListFunc != nil {
		return m.ListFunc(ctx, limit, offset)
	}
	return []*Wallet{}, nil
}

func (m *MockRepository) Exists(ctx context.Context, chainType types.ChainType, address string) (bool, error) {
	if m.ExistsFunc != nil {
		return m.ExistsFunc(ctx, chainType, address)
	}
	return true, nil
}

// MockBlockchainRegistry mocks blockchain.Registry
type MockBlockchainRegistry struct {
	GetBlockchainFunc func(chainType types.ChainType) (blockchain.Blockchain, error)
}

func (m *MockBlockchainRegistry) GetBlockchain(chainType types.ChainType) (blockchain.Blockchain, error) {
	if m.GetBlockchainFunc != nil {
		return m.GetBlockchainFunc(chainType)
	}
	return &MockBlockchain{}, nil
}

// MockBlockchain mocks blockchain.Blockchain
type MockBlockchain struct {
	SubscribeToEventsFunc func(ctx context.Context, addresses []string, topics [][]string) (<-chan types.Log, <-chan error, error)
}

func (m *MockBlockchain) SubscribeToEvents(ctx context.Context, addresses []string, topics [][]string) (<-chan types.Log, <-chan error, error) {
	if m.SubscribeToEventsFunc != nil {
		return m.SubscribeToEventsFunc(ctx, addresses, topics)
	}
	return make(chan types.Log), make(chan error), nil
}

func (m *MockBlockchain) CallContract(ctx context.Context, contractAddress string, methodName string, data []byte) ([]byte, error) {
	return []byte{}, nil
}

func (m *MockBlockchain) Chain() types.Chain {
	return types.Chain{}
}

func (m *MockBlockchain) BroadcastTransaction(ctx context.Context, signedTx []byte) (string, error) {
	return "", nil
}

func (m *MockBlockchain) Close() {
	// No-op for testing
}

func (m *MockBlockchain) EstimateGas(ctx context.Context, tx *types.Transaction) (uint64, error) {
	return 21000, nil
}

func (m *MockBlockchain) FilterLogs(ctx context.Context, addresses []string, topics [][]string, fromBlock, toBlock int64) ([]types.Log, error) {
	return []types.Log{}, nil
}

func (m *MockBlockchain) GetBalance(ctx context.Context, address string) (*big.Int, error) {
	return big.NewInt(1000000000000000000), nil // 1 ETH
}

func (m *MockBlockchain) GetGasPrice(ctx context.Context) (*big.Int, error) {
	return big.NewInt(20000000000), nil // 20 Gwei
}

func (m *MockBlockchain) GetNonce(ctx context.Context, address string) (uint64, error) {
	return 0, nil
}

func (m *MockBlockchain) GetTransaction(ctx context.Context, txHash string) (*types.Transaction, error) {
	return &types.Transaction{}, nil
}

func (m *MockBlockchain) GetTransactionReceipt(ctx context.Context, txHash string) (*types.TransactionReceipt, error) {
	return &types.TransactionReceipt{}, nil
}

// TestCreateWallet tests the CreateWallet method
func TestCreateWallet(t *testing.T) {
	// Setup minimal config and chains
	cfg := &config.Config{}
	chains := types.Chains{
		types.ChainTypeEthereum: types.Chain{
			Type:    types.ChainTypeEthereum,
			KeyType: types.KeyTypeECDSA,
			Curve:   elliptic.P256(),
		},
	}

	// Define test cases
	tests := []struct {
		name        string
		chainType   types.ChainType
		walletName  string
		tags        map[string]string
		setupMocks  func(*MockKeyStore, *MockWalletFactory, *MockRepository, *MockBlockchainRegistry)
		wantErr     bool
		errContains string
	}{
		{
			name:       "successful creation",
			chainType:  types.ChainTypeEthereum,
			walletName: "mywallet",
			tags:       map[string]string{"tag1": "value1"},
			setupMocks: func(ks *MockKeyStore, wf *MockWalletFactory, repo *MockRepository, br *MockBlockchainRegistry) {
				ks.CreateFunc = func(ctx context.Context, name string, keyType types.KeyType, curve elliptic.Curve, tags map[string]string) (*keystore.Key, error) {
					assert.Equal(t, "mywallet", name)
					assert.Equal(t, types.KeyTypeECDSA, keyType)
					assert.Equal(t, map[string]string{"tag1": "value1"}, tags)
					return &keystore.Key{ID: "key123"}, nil
				}
				wf.NewWalletFunc = func(ctx context.Context, chainType types.ChainType, keyID string) (coreWallet.Wallet, error) {
					assert.Equal(t, types.ChainTypeEthereum, chainType)
					assert.Equal(t, "key123", keyID)
					return &MockWallet{
						DeriveAddressFunc: func(ctx context.Context) (string, error) {
							return "0x1234567890", nil
						},
					}, nil
				}
				repo.CreateFunc = func(ctx context.Context, wallet *Wallet) error {
					assert.Equal(t, "mywallet", wallet.Name)
					assert.Equal(t, "0x1234567890", wallet.Address)
					assert.Equal(t, "key123", wallet.KeyID)
					return nil
				}
				br.GetBlockchainFunc = func(chainType types.ChainType) (blockchain.Blockchain, error) {
					return &MockBlockchain{
						SubscribeToEventsFunc: func(ctx context.Context, addresses []string, topics [][]string) (<-chan types.Log, <-chan error, error) {
							assert.Equal(t, []string{"0x1234567890"}, addresses)
							return make(chan types.Log), make(chan error), nil
						},
					}, nil
				}
			},
			wantErr: false,
		},
		{
			name:       "empty name",
			chainType:  types.ChainTypeEthereum,
			walletName: "",
			tags:       map[string]string{},
			setupMocks: func(ks *MockKeyStore, wf *MockWalletFactory, repo *MockRepository, br *MockBlockchainRegistry) {
				// No mocks needed; validation fails before dependencies are called
			},
			wantErr:     true,
			errContains: "name cannot be empty",
		},
	}

	// Run tests
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Initialize mocks
			mockKeyStore := &MockKeyStore{}
			mockWalletFactory := &MockWalletFactory{}
			mockRepository := &MockRepository{}
			mockBlockchainRegistry := &MockBlockchainRegistry{}

			// Setup mock behavior
			tt.setupMocks(mockKeyStore, mockWalletFactory, mockRepository, mockBlockchainRegistry)

			// Create service
			s := NewService(cfg, logger.NewNopLogger(), mockRepository, mockKeyStore, mockWalletFactory, mockBlockchainRegistry, chains)

			// Call the service method
			wallet, err := s.Create(context.Background(), tt.chainType, tt.walletName, tt.tags)

			// Assertions
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContains)
				assert.Nil(t, wallet)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, wallet)
				assert.Equal(t, tt.walletName, wallet.Name)
				assert.Equal(t, tt.chainType, wallet.ChainType)
				assert.Equal(t, "0x1234567890", wallet.Address)
				assert.Equal(t, "key123", wallet.KeyID)
				assert.Equal(t, tt.tags, wallet.Tags)
			}
		})
	}
}

// TestGetWallet tests the GetWallet method
func TestGetWallet(t *testing.T) {
	// Setup minimal config and chains
	cfg := &config.Config{}
	chains := types.Chains{
		types.ChainTypeEthereum: types.Chain{
			Type:    types.ChainTypeEthereum,
			KeyType: types.KeyTypeECDSA,
			Curve:   elliptic.P256(),
		},
	}

	// Define test cases
	tests := []struct {
		name        string
		chainType   types.ChainType
		address     string
		setupMocks  func(*MockRepository)
		wantErr     bool
		errContains string
	}{
		{
			name:      "successful retrieval",
			chainType: types.ChainTypeEthereum,
			address:   "0x1234567890abcdef1234567890abcdef12345678",
			setupMocks: func(repo *MockRepository) {
				repo.GetFunc = func(ctx context.Context, chainType types.ChainType, address string) (*Wallet, error) {
					assert.Equal(t, types.ChainTypeEthereum, chainType)
					assert.Equal(t, "0x1234567890abcdef1234567890abcdef12345678", address)
					return &Wallet{
						ID:        "wallet123",
						ChainType: chainType,
						Address:   address,
						Name:      "Test Wallet",
						Tags:      map[string]string{"tag1": "value1"},
					}, nil
				}
			},
			wantErr: false,
		},
		{
			name:      "wallet not found",
			chainType: types.ChainTypeEthereum,
			address:   "0x1234567890abcdef1234567890abcdef12345678",
			setupMocks: func(repo *MockRepository) {
				repo.GetFunc = func(ctx context.Context, chainType types.ChainType, address string) (*Wallet, error) {
					return nil, sql.ErrNoRows
				}
			},
			wantErr:     true,
			errContains: "wallet not found",
		},
		{
			name:      "empty chain type",
			chainType: "",
			address:   "0x1234567890abcdef1234567890abcdef12345678",
			setupMocks: func(repo *MockRepository) {
				// No mocks needed; validation fails before repository is called
			},
			wantErr:     true,
			errContains: "chain type cannot be empty",
		},
		{
			name:      "empty address",
			chainType: types.ChainTypeEthereum,
			address:   "",
			setupMocks: func(repo *MockRepository) {
				// No mocks needed; validation fails before repository is called
			},
			wantErr:     true,
			errContains: "address cannot be empty",
		},
		{
			name:      "unsupported chain type",
			chainType: "unsupported",
			address:   "0x1234567890abcdef1234567890abcdef12345678",
			setupMocks: func(repo *MockRepository) {
				// No mocks needed; validation fails before repository is called
			},
			wantErr:     true,
			errContains: "unsupported chain type",
		},
	}

	// Run tests
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Initialize mocks
			mockKeyStore := &MockKeyStore{}
			mockWalletFactory := &MockWalletFactory{}
			mockRepository := &MockRepository{}
			mockBlockchainRegistry := &MockBlockchainRegistry{}

			// Setup mock behavior
			tt.setupMocks(mockRepository)

			// Create service
			s := NewService(cfg, logger.NewNopLogger(), mockRepository, mockKeyStore, mockWalletFactory, mockBlockchainRegistry, chains)

			// Call the service method
			wallet, err := s.Get(context.Background(), tt.chainType, tt.address)

			// Assertions
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContains)
				assert.Nil(t, wallet)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, wallet)
				assert.Equal(t, tt.chainType, wallet.ChainType)
				assert.Equal(t, tt.address, wallet.Address)
			}
		})
	}
}

// TestUpdateLastBlockNumber tests the UpdateLastBlockNumber method
func TestUpdateLastBlockNumber(t *testing.T) {
	// Setup minimal config and chains
	cfg := &config.Config{}
	chains := types.Chains{
		types.ChainTypeEthereum: types.Chain{
			Type:    types.ChainTypeEthereum,
			KeyType: types.KeyTypeECDSA,
			Curve:   elliptic.P256(),
		},
	}

	// Define test cases
	tests := []struct {
		name        string
		chainType   types.ChainType
		address     string
		blockNumber int64
		setupMocks  func(*MockRepository)
		wantErr     bool
		errContains string
	}{
		{
			name:        "successful update",
			chainType:   types.ChainTypeEthereum,
			address:     "0x1234567890abcdef1234567890abcdef12345678",
			blockNumber: 12345,
			setupMocks: func(repo *MockRepository) {
				repo.GetFunc = func(ctx context.Context, chainType types.ChainType, address string) (*Wallet, error) {
					return &Wallet{
						ID:        "wallet123",
						ChainType: chainType,
						Address:   address,
						Name:      "Test Wallet",
					}, nil
				}
				repo.UpdateFunc = func(ctx context.Context, wallet *Wallet) error {
					assert.Equal(t, int64(12345), wallet.LastBlockNumber)
					return nil
				}
			},
			wantErr: false,
		},
		{
			name:        "wallet not found",
			chainType:   types.ChainTypeEthereum,
			address:     "0x1234567890abcdef1234567890abcdef12345678",
			blockNumber: 12345,
			setupMocks: func(repo *MockRepository) {
				repo.GetFunc = func(ctx context.Context, chainType types.ChainType, address string) (*Wallet, error) {
					return nil, sql.ErrNoRows
				}
			},
			wantErr:     true,
			errContains: "wallet not found",
		},
		{
			name:        "empty chain type",
			chainType:   "",
			address:     "0x1234567890abcdef1234567890abcdef12345678",
			blockNumber: 12345,
			setupMocks: func(repo *MockRepository) {
				// No mocks needed; validation fails before repository is called
			},
			wantErr:     true,
			errContains: "chain type cannot be empty",
		},
		{
			name:        "empty address",
			chainType:   types.ChainTypeEthereum,
			address:     "",
			blockNumber: 12345,
			setupMocks: func(repo *MockRepository) {
				// No mocks needed; validation fails before repository is called
			},
			wantErr:     true,
			errContains: "address cannot be empty",
		},
		{
			name:        "negative block number",
			chainType:   types.ChainTypeEthereum,
			address:     "0x1234567890abcdef1234567890abcdef12345678",
			blockNumber: -1,
			setupMocks: func(repo *MockRepository) {
				// No mocks needed; validation fails before repository is called
			},
			wantErr:     true,
			errContains: "block number cannot be negative",
		},
		{
			name:        "unsupported chain type",
			chainType:   "unsupported",
			address:     "0x1234567890abcdef1234567890abcdef12345678",
			blockNumber: 12345,
			setupMocks: func(repo *MockRepository) {
				// No mocks needed; validation fails before repository is called
			},
			wantErr:     true,
			errContains: "unsupported chain type",
		},
	}

	// Run tests
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Initialize mocks
			mockKeyStore := &MockKeyStore{}
			mockWalletFactory := &MockWalletFactory{}
			mockRepository := &MockRepository{}
			mockBlockchainRegistry := &MockBlockchainRegistry{}

			// Setup mock behavior
			tt.setupMocks(mockRepository)

			// Create service
			s := NewService(cfg, logger.NewNopLogger(), mockRepository, mockKeyStore, mockWalletFactory, mockBlockchainRegistry, chains)

			// Call the service method
			err := s.UpdateLastBlockNumber(context.Background(), tt.chainType, tt.address, tt.blockNumber)

			// Assertions
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContains)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
