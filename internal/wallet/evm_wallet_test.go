package wallet

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/asn1"
	"math/big"
	"strings"
	"testing"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"vault0/internal/config"
	"vault0/internal/keygen"
	"vault0/internal/keystore"
	"vault0/internal/types"
)

// ecdsaSignature is used for marshalling ECDSA signatures in ASN.1 DER format
type ecdsaSignature struct {
	R, S *big.Int
}

// mockAppConfig implements the AppConfig interface for testing
type mockAppConfig struct {
	configs map[string]*config.BlockchainConfig
}

func (m *mockAppConfig) GetBlockchainConfig(chainType string) *config.BlockchainConfig {
	return m.configs[chainType]
}

// setupTest creates a test wallet with mock dependencies
func setupTest(t *testing.T) (*EVMWallet, *MockKeyStore, *mockAppConfig) {
	ks := &MockKeyStore{}
	appCfg := &mockAppConfig{
		configs: map[string]*config.BlockchainConfig{
			string(types.ChainTypeEthereum): {
				ChainID:         1,
				DefaultGasLimit: 21000,
				DefaultGasPrice: 20000000000, // 20 Gwei
			},
		},
	}
	wallet, err := NewEVMWallet(ks, types.ChainTypeEthereum, "test", appCfg)
	require.NoError(t, err)
	return wallet, ks, appCfg
}

// TestChainType tests the ChainType method
func TestChainType(t *testing.T) {
	wallet, _, _ := setupTest(t)
	assert.Equal(t, types.ChainTypeEthereum, wallet.ChainType(), "ChainType should return Ethereum")
}

// TestDeriveAddress tests the DeriveAddress method
func TestDeriveAddress(t *testing.T) {
	wallet, ks, _ := setupTest(t)
	ctx := context.Background()

	// Generate a test key pair using secp256k1 curve
	privKey, err := ecdsa.GenerateKey(keygen.Secp256k1Curve, rand.Reader)
	require.NoError(t, err)

	// Get public key bytes using secp256k1 format
	pubKeyBytes, err := keygen.MarshalPublicKey(&privKey.PublicKey)
	require.NoError(t, err)

	// Set up the mock to return our key
	ks.GetPublicKeyFunc = func(ctx context.Context, id string) (*keystore.Key, error) {
		if id == "test" {
			return &keystore.Key{
				ID:        "test",
				Name:      "test",
				Type:      keygen.KeyTypeECDSA,
				Curve:     keygen.Secp256k1Curve,
				PublicKey: pubKeyBytes,
			}, nil
		}
		return nil, keystore.ErrKeyNotFound
	}

	// Derive address
	address, err := wallet.DeriveAddress(ctx)
	require.NoError(t, err)
	expectedAddress := crypto.PubkeyToAddress(privKey.PublicKey).Hex()
	assert.Equal(t, expectedAddress, address, "Derived address should match expected address")
}

// TestCreateNativeTransaction tests the CreateNativeTransaction method
func TestCreateNativeTransaction(t *testing.T) {
	wallet, ks, _ := setupTest(t)
	ctx := context.Background()

	// Setup key using secp256k1 curve
	privKey, err := ecdsa.GenerateKey(keygen.Secp256k1Curve, rand.Reader)
	require.NoError(t, err)

	// Get public key bytes using secp256k1 format
	pubKeyBytes, err := keygen.MarshalPublicKey(&privKey.PublicKey)
	require.NoError(t, err)

	// Marshal private key using secp256k1 format
	privKeyBytes, err := keygen.MarshalPrivateKey(privKey)
	require.NoError(t, err)

	// Import the key into the mock keystore
	_, err = ks.Import(ctx, "test", keygen.KeyTypeECDSA, keygen.Secp256k1Curve, privKeyBytes, pubKeyBytes, nil)
	require.NoError(t, err)

	// Set up the mock to return our key
	ks.GetPublicKeyFunc = func(ctx context.Context, id string) (*keystore.Key, error) {
		if id == "test" {
			return &keystore.Key{
				ID:        "test",
				Name:      "test",
				Type:      keygen.KeyTypeECDSA,
				Curve:     keygen.Secp256k1Curve,
				PublicKey: pubKeyBytes,
			}, nil
		}
		return nil, keystore.ErrKeyNotFound
	}

	toAddress := "0x742d35Cc6634C0532925a3b844Bc454e4438f44e"
	amount := big.NewInt(1000000000000000000) // 1 ETH

	// Test regular native transaction
	tx, err := wallet.CreateNativeTransaction(ctx, toAddress, amount, types.TransactionOptions{})
	require.NoError(t, err)
	assert.Equal(t, types.ChainTypeEthereum, tx.Chain, "Transaction chain should be Ethereum")
	assert.Equal(t, crypto.PubkeyToAddress(privKey.PublicKey).Hex(), tx.From, "From address should match wallet address")
	assert.Equal(t, toAddress, tx.To, "To address should match input")
	assert.Equal(t, amount, tx.Value, "Transaction value should match input")
	assert.Equal(t, types.TransactionTypeNative, tx.Type, "Transaction type should be native")

	// Test contract deployment (zero address)
	deployData := []byte{0x60, 0x80, 0x60, 0x40} // Some dummy contract bytecode
	tx, err = wallet.CreateNativeTransaction(ctx, types.ZeroAddress, big.NewInt(0), types.TransactionOptions{
		Data: deployData,
	})
	require.NoError(t, err)
	assert.Equal(t, types.ZeroAddress, tx.To, "To address should be zero address for contract deployment")
	assert.Equal(t, deployData, tx.Data, "Transaction data should contain contract bytecode")
	assert.Equal(t, types.TransactionTypeNative, tx.Type, "Transaction type should be native")

	// Test failure case: invalid toAddress
	_, err = wallet.CreateNativeTransaction(ctx, "invalid-address", amount, types.TransactionOptions{})
	assert.Error(t, err, "CreateNativeTransaction should fail with invalid address")
}

// TestCreateTokenTransaction tests the CreateTokenTransaction method
func TestCreateTokenTransaction(t *testing.T) {
	wallet, ks, _ := setupTest(t)
	ctx := context.Background()

	// Setup key using secp256k1 curve
	privKey, err := ecdsa.GenerateKey(keygen.Secp256k1Curve, rand.Reader)
	require.NoError(t, err)

	// Get public key bytes using secp256k1 format
	pubKeyBytes, err := keygen.MarshalPublicKey(&privKey.PublicKey)
	require.NoError(t, err)

	// Marshal private key using secp256k1 format
	privKeyBytes, err := keygen.MarshalPrivateKey(privKey)
	require.NoError(t, err)

	// Import the key into the mock keystore
	_, err = ks.Import(ctx, "test", keygen.KeyTypeECDSA, keygen.Secp256k1Curve, privKeyBytes, pubKeyBytes, nil)
	require.NoError(t, err)

	// Set up the mock to return our key
	ks.GetPublicKeyFunc = func(ctx context.Context, id string) (*keystore.Key, error) {
		if id == "test" {
			return &keystore.Key{
				ID:        "test",
				Name:      "test",
				Type:      keygen.KeyTypeECDSA,
				Curve:     keygen.Secp256k1Curve,
				PublicKey: pubKeyBytes,
			}, nil
		}
		return nil, keystore.ErrKeyNotFound
	}

	tokenAddress := "0xdAC17F958D2ee523a2206206994597C13D831ec7" // USDT address
	toAddress := "0x742d35Cc6634C0532925a3b844Bc454e4438f44e"
	amount := big.NewInt(1000000) // 1 USDT with 6 decimals

	tx, err := wallet.CreateTokenTransaction(ctx, tokenAddress, toAddress, amount, types.TransactionOptions{})
	require.NoError(t, err)
	assert.Equal(t, types.ChainTypeEthereum, tx.Chain, "Transaction chain should be Ethereum")
	assert.Equal(t, crypto.PubkeyToAddress(privKey.PublicKey).Hex(), tx.From, "From address should match wallet address")
	assert.Equal(t, tokenAddress, tx.To, "To address should be token contract address")
	assert.Equal(t, big.NewInt(0), tx.Value, "Value should be 0 for token transactions")
	assert.NotEmpty(t, tx.Data, "Transaction data should contain ERC20 transfer ABI")
	assert.Equal(t, types.TransactionTypeERC20, tx.Type, "Transaction type should be ERC20")
	assert.Equal(t, tokenAddress, tx.TokenAddress, "Token address should match input")

	// Test failure case: invalid toAddress
	_, err = wallet.CreateTokenTransaction(ctx, tokenAddress, "invalid-address", amount, types.TransactionOptions{})
	assert.Error(t, err, "CreateTokenTransaction should fail with invalid toAddress")
}

// TestSignTransaction tests the SignTransaction method with DER-encoded keys
func TestSignTransaction(t *testing.T) {
	wallet, ks, _ := setupTest(t)
	ctx := context.Background()

	// Setup key using secp256k1 curve
	privKey, err := ecdsa.GenerateKey(keygen.Secp256k1Curve, rand.Reader)
	require.NoError(t, err)

	// Get public key bytes using secp256k1 format
	pubKeyBytes, err := keygen.MarshalPublicKey(&privKey.PublicKey)
	require.NoError(t, err)

	// Marshal private key using secp256k1 format
	privKeyBytes, err := keygen.MarshalPrivateKey(privKey)
	require.NoError(t, err)

	// Import the key into the mock keystore
	_, err = ks.Import(ctx, "test", keygen.KeyTypeECDSA, keygen.Secp256k1Curve, privKeyBytes, pubKeyBytes, nil)
	require.NoError(t, err)

	// Set up the mock to return our key and sign function
	ks.GetPublicKeyFunc = func(ctx context.Context, id string) (*keystore.Key, error) {
		if id == "test" {
			return &keystore.Key{
				ID:        "test",
				Name:      "test",
				Type:      keygen.KeyTypeECDSA,
				Curve:     keygen.Secp256k1Curve,
				PublicKey: pubKeyBytes,
			}, nil
		}
		return nil, keystore.ErrKeyNotFound
	}

	// Set up the mock sign function
	ks.SignFunc = func(ctx context.Context, id string, data []byte, dataType keystore.DataType) ([]byte, error) {
		if id == "test" {
			// For testing, just sign with the private key
			var hash []byte
			if dataType == keystore.DataTypeDigest {
				hash = data
			} else {
				hasher := sha256.New()
				hasher.Write(data)
				hash = hasher.Sum(nil)
			}

			r, s, err := ecdsa.Sign(rand.Reader, privKey, hash)
			if err != nil {
				return nil, err
			}

			// Encode signature in ASN.1 DER format
			signature := ecdsaSignature{R: r, S: s}
			return asn1.Marshal(signature)
		}
		return nil, keystore.ErrKeyNotFound
	}

	toAddress := "0x742d35Cc6634C0532925a3b844Bc454e4438f44e"
	amount := big.NewInt(1000000000000000000) // 1 ETH

	// Create a transaction
	tx := &types.Transaction{
		Chain:    types.ChainTypeEthereum,
		From:     crypto.PubkeyToAddress(privKey.PublicKey).Hex(),
		To:       toAddress,
		Value:    amount,
		Nonce:    0,
		GasPrice: big.NewInt(20000000000), // 20 Gwei
		GasLimit: 21000,
		Type:     types.TransactionTypeNative,
	}

	// Sign the transaction
	signedTxBytes, err := wallet.SignTransaction(ctx, tx)
	require.NoError(t, err)
	assert.NotEmpty(t, signedTxBytes, "Signed transaction bytes should not be empty")

	// Test failure case: mismatched from address
	tx.From = "0x1234567890abcdef1234567890abcdef12345678"
	_, err = wallet.SignTransaction(ctx, tx)
	assert.Error(t, err, "SignTransaction should fail with mismatched from address")
}

func TestCreate(t *testing.T) {
	// Setup
	wallet, ks, _ := setupTest(t)
	ctx := context.Background()

	// Generate a test key pair
	privKey, err := ecdsa.GenerateKey(keygen.Secp256k1Curve, rand.Reader)
	require.NoError(t, err)

	// Get public key bytes using secp256k1 format
	pubKeyBytes, err := keygen.MarshalPublicKey(&privKey.PublicKey)
	require.NoError(t, err)

	// Set up the CreateFunc to return our mock key
	expectedKeyID := "mock-key-id"
	ks.CreateFunc = func(ctx context.Context, name string, keyType keygen.KeyType, curve elliptic.Curve, tags map[string]string) (*keystore.Key, error) {
		return &keystore.Key{
			ID:        expectedKeyID,
			Name:      name,
			Type:      keyType,
			Curve:     curve,
			Tags:      tags,
			PublicKey: pubKeyBytes,
		}, nil
	}

	// Set up the GetPublicKeyFunc to return our key
	ks.GetPublicKeyFunc = func(ctx context.Context, id string) (*keystore.Key, error) {
		if id == expectedKeyID {
			return &keystore.Key{
				ID:        expectedKeyID,
				Type:      keygen.KeyTypeECDSA,
				Curve:     keygen.Secp256k1Curve,
				PublicKey: pubKeyBytes,
			}, nil
		}
		return nil, keystore.ErrKeyNotFound
	}

	// Set expectations
	keyName := "test-wallet"
	keyTags := map[string]string{"purpose": "testing", "env": "unit-test"}

	// Execute
	walletInfo, err := wallet.Create(ctx, keyName, keyTags)

	// Verify
	assert.NoError(t, err)
	assert.NotNil(t, walletInfo)
	assert.Equal(t, expectedKeyID, walletInfo.KeyID)
	assert.NotEmpty(t, walletInfo.Address)
	assert.True(t, strings.HasPrefix(walletInfo.Address, "0x"), "Address should start with 0x")
	assert.Equal(t, wallet.chainType, walletInfo.ChainType)
}

func TestCreateWithErrors(t *testing.T) {
	// Setup
	ks := &MockKeyStore{
		// Override the Create method to return an error
		CreateFunc: func(ctx context.Context, name string, keyType keygen.KeyType, curve elliptic.Curve, tags map[string]string) (*keystore.Key, error) {
			return nil, assert.AnError
		},
	}

	appCfg := &mockAppConfig{
		configs: map[string]*config.BlockchainConfig{
			string(types.ChainTypeEthereum): {
				ChainID:         1,
				DefaultGasLimit: 21000,
				DefaultGasPrice: 20000000000, // 20 Gwei
			},
		},
	}

	wallet, err := NewEVMWallet(ks, types.ChainTypeEthereum, "test", appCfg)
	require.NoError(t, err)

	ctx := context.Background()
	keyName := "test-wallet"
	keyTags := map[string]string{"purpose": "testing"}

	// Execute
	walletInfo, err := wallet.Create(ctx, keyName, keyTags)

	// Verify
	assert.Error(t, err)
	assert.Nil(t, walletInfo)
	assert.Contains(t, err.Error(), "failed to create key")
}

func TestCreateWithDeriveAddressError(t *testing.T) {
	// Setup - create a wallet with a mock keystore
	wallet, ks, _ := setupTest(t)
	ctx := context.Background()

	// Set up the CreateFunc to return a valid key
	mockKeyID := "test-key-id"
	ks.CreateFunc = func(ctx context.Context, name string, keyType keygen.KeyType, curve elliptic.Curve, tags map[string]string) (*keystore.Key, error) {
		return &keystore.Key{
			ID:    mockKeyID,
			Name:  name,
			Type:  keyType,
			Curve: curve,
			Tags:  tags,
		}, nil
	}

	// Override the GetPublicKeyFunc in the mock keystore to return an error
	ks.GetPublicKeyFunc = func(ctx context.Context, id string) (*keystore.Key, error) {
		return nil, assert.AnError
	}

	keyName := "test-wallet"
	keyTags := map[string]string{"purpose": "testing"}

	// Execute
	walletInfo, err := wallet.Create(ctx, keyName, keyTags)

	// Verify
	assert.Error(t, err)
	assert.Nil(t, walletInfo)
	assert.Contains(t, err.Error(), "failed to derive address")
}
