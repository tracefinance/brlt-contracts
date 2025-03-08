package wallet

import (
	"context"
	"crypto/ecdsa"
	"crypto/rand"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"vault0/internal/config"
	"vault0/internal/keygen"
	"vault0/internal/keystore"
	"vault0/internal/types"
)

// mockAppConfig implements the AppConfig interface for testing
type mockAppConfig struct {
	configs map[string]*config.BlockchainConfig
}

func (m *mockAppConfig) GetBlockchainConfig(chainType string) *config.BlockchainConfig {
	return m.configs[chainType]
}

// setupTest creates a test wallet with mock dependencies
func setupTest(t *testing.T) (*EVMWallet, *keystore.MockKeyStore, *mockAppConfig) {
	ks := keystore.NewMockKeyStore()
	appCfg := &mockAppConfig{
		configs: map[string]*config.BlockchainConfig{
			string(types.ChainTypeEthereum): {
				ChainID:         1,
				DefaultGasLimit: 21000,
				DefaultGasPrice: 20000000000, // 20 Gwei
			},
		},
	}
	wallet, err := NewEVMWallet(ks, types.ChainTypeEthereum, "test-key", appCfg)
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

	// Import the key into the mock keystore
	_, err = ks.Import(ctx, "test-key", "test", keygen.KeyTypeECDSA, keygen.Secp256k1Curve, nil, pubKeyBytes, nil)
	require.NoError(t, err)

	// Derive address
	address, err := wallet.DeriveAddress(ctx)
	require.NoError(t, err)
	expectedAddress := crypto.PubkeyToAddress(privKey.PublicKey).Hex()
	assert.Equal(t, expectedAddress, address, "Derived address should match expected address")

	// Test failure case: empty public key
	ks.Keys["test-key"].PublicKey = []byte{}
	_, err = wallet.DeriveAddress(ctx)
	assert.Error(t, err, "DeriveAddress should fail with empty public key")
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

	// Import the key into the mock keystore
	_, err = ks.Import(ctx, "test-key", "test", keygen.KeyTypeECDSA, keygen.Secp256k1Curve, nil, pubKeyBytes, nil)
	require.NoError(t, err)

	toAddress := "0x742d35Cc6634C0532925a3b844Bc454e4438f44e"
	amount := big.NewInt(1000000000000000000) // 1 ETH

	tx, err := wallet.CreateNativeTransaction(ctx, toAddress, amount, types.TransactionOptions{})
	require.NoError(t, err)
	assert.Equal(t, types.ChainTypeEthereum, tx.Chain, "Transaction chain should be Ethereum")
	assert.Equal(t, crypto.PubkeyToAddress(privKey.PublicKey).Hex(), tx.From, "From address should match wallet address")
	assert.Equal(t, toAddress, tx.To, "To address should match input")
	assert.Equal(t, amount, tx.Value, "Transaction value should match input")
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

	// Import the key into the mock keystore
	_, err = ks.Import(ctx, "test-key", "test", keygen.KeyTypeECDSA, keygen.Secp256k1Curve, nil, pubKeyBytes, nil)
	require.NoError(t, err)

	tokenAddress := "0x6B175474E89094C44Da98b954EedeAC495271d0F" // DAI token
	toAddress := "0x742d35Cc6634C0532925a3b844Bc454e4438f44e"
	amount := big.NewInt(1000000000000000000) // 1 token

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

	// Generate a test key pair using secp256k1 curve
	privKey, err := ecdsa.GenerateKey(keygen.Secp256k1Curve, rand.Reader)
	require.NoError(t, err)

	// Marshal private key using secp256k1 format
	privKeyBytes, err := keygen.MarshalPrivateKey(privKey)
	require.NoError(t, err)

	// Get public key bytes in Ethereum format
	pubKeyBytes := crypto.FromECDSAPub(&privKey.PublicKey)

	// Import the key into the mock keystore
	_, err = ks.Import(ctx, "test-key", "test", keygen.KeyTypeECDSA, keygen.Secp256k1Curve, privKeyBytes, pubKeyBytes, nil)
	require.NoError(t, err)

	// Create a transaction
	tx := &types.Transaction{
		Chain:    types.ChainTypeEthereum,
		From:     crypto.PubkeyToAddress(privKey.PublicKey).Hex(),
		To:       "0x742d35Cc6634C0532925a3b844Bc454e4438f44e",
		Value:    big.NewInt(1000000000000000000), // 1 ETH
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
