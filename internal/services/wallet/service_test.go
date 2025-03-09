package wallet

import (
	"context"
	"database/sql"
	"errors"
	"math/big"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"crypto/elliptic"
	"vault0/internal/config"
	"vault0/internal/keygen"
	"vault0/internal/keystore"
	"vault0/internal/types"
	coreWallet "vault0/internal/wallet"
)

// MockRepository is a mock implementation of the Repository interface
type MockRepository struct {
	mock.Mock
}

func (m *MockRepository) Create(ctx context.Context, wallet *Wallet) error {
	args := m.Called(ctx, wallet)
	return args.Error(0)
}

func (m *MockRepository) GetByID(ctx context.Context, id string) (*Wallet, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Wallet), args.Error(1)
}

func (m *MockRepository) Update(ctx context.Context, wallet *Wallet) error {
	args := m.Called(ctx, wallet)
	return args.Error(0)
}

func (m *MockRepository) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockRepository) List(ctx context.Context, limit, offset int) ([]*Wallet, error) {
	args := m.Called(ctx, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*Wallet), args.Error(1)
}

// MockWalletFactory is a mock implementation of the WalletFactory interface
type MockWalletFactory struct {
	mock.Mock
}

func (m *MockWalletFactory) NewWallet(ctx context.Context, chainType types.ChainType, keyID string) (coreWallet.Wallet, error) {
	args := m.Called(ctx, chainType, keyID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(coreWallet.Wallet), args.Error(1)
}

// MockWallet is a mock implementation of the coreWallet.Wallet interface
type MockWallet struct {
	mock.Mock
}

func (m *MockWallet) ChainType() types.ChainType {
	args := m.Called()
	return args.Get(0).(types.ChainType)
}

func (m *MockWallet) Create(ctx context.Context, name string, tags map[string]string) (*coreWallet.WalletInfo, error) {
	args := m.Called(ctx, name, tags)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*coreWallet.WalletInfo), args.Error(1)
}

func (m *MockWallet) DeriveAddress(ctx context.Context) (string, error) {
	args := m.Called(ctx)
	return args.String(0), args.Error(1)
}

func (m *MockWallet) CreateNativeTransaction(ctx context.Context, toAddress string, amount *big.Int, options types.TransactionOptions) (*types.Transaction, error) {
	args := m.Called(ctx, toAddress, amount, options)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*types.Transaction), args.Error(1)
}

func (m *MockWallet) CreateTokenTransaction(ctx context.Context, tokenAddress, toAddress string, amount *big.Int, options types.TransactionOptions) (*types.Transaction, error) {
	args := m.Called(ctx, tokenAddress, toAddress, amount, options)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*types.Transaction), args.Error(1)
}

func (m *MockWallet) SignTransaction(ctx context.Context, tx *types.Transaction) ([]byte, error) {
	args := m.Called(ctx, tx)
	return args.Get(0).([]byte), args.Error(1)
}

// MockKeyStore is a mock implementation of the KeyStore interface
type MockKeyStore struct {
	mock.Mock
}

func (m *MockKeyStore) Create(ctx context.Context, name string, keyType keygen.KeyType, curve elliptic.Curve, tags map[string]string) (*keystore.Key, error) {
	args := m.Called(ctx, name, keyType, curve, tags)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*keystore.Key), args.Error(1)
}

func (m *MockKeyStore) Import(ctx context.Context, name string, keyType keygen.KeyType, curve elliptic.Curve, privateKey, publicKey []byte, tags map[string]string) (*keystore.Key, error) {
	args := m.Called(ctx, name, keyType, curve, privateKey, publicKey, tags)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*keystore.Key), args.Error(1)
}

func (m *MockKeyStore) Sign(ctx context.Context, id string, data []byte, dataType keystore.DataType) ([]byte, error) {
	args := m.Called(ctx, id, data, dataType)
	return args.Get(0).([]byte), args.Error(1)
}

func (m *MockKeyStore) GetPublicKey(ctx context.Context, id string) (*keystore.Key, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*keystore.Key), args.Error(1)
}

func (m *MockKeyStore) List(ctx context.Context) ([]*keystore.Key, error) {
	args := m.Called(ctx)
	return args.Get(0).([]*keystore.Key), args.Error(1)
}

func (m *MockKeyStore) Update(ctx context.Context, id string, name string, tags map[string]string) (*keystore.Key, error) {
	args := m.Called(ctx, id, name, tags)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*keystore.Key), args.Error(1)
}

func (m *MockKeyStore) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func TestCreateWallet(t *testing.T) {
	ctx := context.Background()
	repo := new(MockRepository)
	mockWallet := new(MockWallet)
	cfg := &config.Config{}

	// Create a mock factory that we control directly
	mockFactory := new(MockWalletFactory)

	// Create a mock keystore
	mockKeyStore := new(MockKeyStore)

	// Create a service with our mocked factory but manually inject it
	service := &DefaultService{
		repository:    repo,
		walletFactory: mockFactory,
		config:        cfg,
		keystore:      mockKeyStore,
	}

	// Test successful wallet creation
	t.Run("success", func(t *testing.T) {
		// Mock keystore.Create to return a key
		expectedKeyID := "key123"
		mockKeyStore.On("Create", ctx, "Test Wallet", keygen.KeyTypeECDSA, keygen.Secp256k1Curve, map[string]string{"tag1": "value1"}).
			Return(&keystore.Key{
				ID:   expectedKeyID,
				Name: "Test Wallet",
				Type: keygen.KeyTypeECDSA,
				Tags: map[string]string{"tag1": "value1"},
			}, nil).Once()

		// Setup the mock factory to return our mock wallet with the expected key ID
		mockFactory.On("NewWallet", ctx, types.ChainTypeBase, expectedKeyID).Return(mockWallet, nil).Once()

		// Mock wallet.DeriveAddress to return an address
		expectedAddress := "0x1234567890abcdef"
		mockWallet.On("DeriveAddress", ctx).Return(expectedAddress, nil).Once()

		// The wallet should be saved to the repository
		repo.On("Create", ctx, mock.MatchedBy(func(w *Wallet) bool {
			return w.KeyID == expectedKeyID && w.Address == expectedAddress && w.ChainType == types.ChainTypeBase
		})).Return(nil).Once()

		wallet, err := service.CreateWallet(ctx, types.ChainTypeBase, "Test Wallet", map[string]string{"tag1": "value1"})

		assert.NoError(t, err)
		assert.NotNil(t, wallet)
		assert.Equal(t, expectedKeyID, wallet.KeyID)
		assert.Equal(t, expectedAddress, wallet.Address)
		assert.Equal(t, types.ChainTypeBase, wallet.ChainType)

		mockKeyStore.AssertExpectations(t)
		mockFactory.AssertExpectations(t)
		mockWallet.AssertExpectations(t)
		repo.AssertExpectations(t)
	})

	// Test with empty name
	t.Run("empty_name", func(t *testing.T) {
		wallet, err := service.CreateWallet(ctx, types.ChainTypeBase, "", nil)

		assert.Error(t, err)
		assert.Nil(t, wallet)
		assert.True(t, errors.Is(err, ErrInvalidInput))
	})

	// Test with key creation error
	t.Run("key_creation_error", func(t *testing.T) {
		mockKeyStore.On("Create", ctx, "Test Wallet", keygen.KeyTypeECDSA, keygen.Secp256k1Curve, map[string]string{}).
			Return(nil, errors.New("key creation failed")).Once()

		wallet, err := service.CreateWallet(ctx, types.ChainTypeBase, "Test Wallet", map[string]string{})

		assert.Error(t, err)
		assert.Nil(t, wallet)
		assert.Contains(t, err.Error(), "failed to create key")
		mockKeyStore.AssertExpectations(t)
	})
}

func TestUpdateWallet(t *testing.T) {
	ctx := context.Background()
	repo := new(MockRepository)
	mockFactory := new(MockWalletFactory)
	cfg := &config.Config{}

	// Create service with direct injection of mocks
	service := &DefaultService{
		repository:    repo,
		walletFactory: mockFactory,
		config:        cfg,
	}

	// Test successful wallet update
	t.Run("success", func(t *testing.T) {
		existingWallet := &Wallet{
			ID:        "wallet123",
			KeyID:     "key123",
			ChainType: types.ChainTypeBase,
			Address:   "0x1234567890abcdef",
			Name:      "Old Name",
			Tags:      map[string]string{"old": "tag"},
			CreatedAt: time.Now().Add(-24 * time.Hour),
			UpdatedAt: time.Now().Add(-24 * time.Hour),
		}

		newName := "New Name"
		newTags := map[string]string{"new": "tag"}

		repo.On("GetByID", ctx, "wallet123").Return(existingWallet, nil).Once()
		repo.On("Update", ctx, mock.MatchedBy(func(w *Wallet) bool {
			return w.ID == existingWallet.ID && w.Name == newName && w.Tags["new"] == "tag"
		})).Return(nil).Once()

		wallet, err := service.UpdateWallet(ctx, "wallet123", newName, newTags)

		assert.NoError(t, err)
		assert.NotNil(t, wallet)
		assert.Equal(t, newName, wallet.Name)
		assert.Equal(t, newTags, wallet.Tags)

		repo.AssertExpectations(t)
	})

	// Test with non-existent wallet
	t.Run("wallet_not_found", func(t *testing.T) {
		repo.On("GetByID", ctx, "nonexistent").Return(nil, sql.ErrNoRows).Once()

		wallet, err := service.UpdateWallet(ctx, "nonexistent", "New Name", nil)

		assert.Error(t, err)
		assert.Nil(t, wallet)
		assert.True(t, errors.Is(err, ErrWalletNotFound))

		repo.AssertExpectations(t)
	})
}

func TestDeleteWallet(t *testing.T) {
	ctx := context.Background()
	repo := new(MockRepository)
	mockFactory := new(MockWalletFactory)
	cfg := &config.Config{}

	// Create service with direct injection of mocks
	service := &DefaultService{
		repository:    repo,
		walletFactory: mockFactory,
		config:        cfg,
	}

	// Test successful wallet deletion
	t.Run("success", func(t *testing.T) {
		repo.On("Delete", ctx, "wallet123").Return(nil).Once()

		err := service.DeleteWallet(ctx, "wallet123")

		assert.NoError(t, err)
		repo.AssertExpectations(t)
	})

	// Test with non-existent wallet
	t.Run("wallet_not_found", func(t *testing.T) {
		repo.On("Delete", ctx, "nonexistent").Return(sql.ErrNoRows).Once()

		err := service.DeleteWallet(ctx, "nonexistent")

		assert.Error(t, err)
		assert.True(t, errors.Is(err, ErrWalletNotFound))

		repo.AssertExpectations(t)
	})
}
