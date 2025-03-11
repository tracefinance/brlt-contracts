package wallet

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/asn1"
	"fmt"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	coreCrypto "vault0/internal/core/crypto"
	"vault0/internal/core/keystore"
	"vault0/internal/types"
)

// ecdsaSignature is used for marshalling ECDSA signatures in ASN.1 DER format
type ecdsaSignature struct {
	R, S *big.Int
}

// setupTest creates a test wallet with mock dependencies
func setupTest(t *testing.T) (*EVMWallet, *MockKeyStore) {
	ks := &MockKeyStore{}

	// Create a chain struct
	chain := types.Chain{
		ID:              1,
		Type:            types.ChainTypeEthereum,
		Name:            "Ethereum",
		Symbol:          "ETH",
		RPCUrl:          "https://mainnet.infura.io",
		ExplorerUrl:     "https://etherscan.io",
		KeyType:         types.KeyTypeECDSA,
		Curve:           coreCrypto.Secp256k1Curve,
		DefaultGasLimit: 21000,
		DefaultGasPrice: 20000000000, // 20 Gwei
	}

	wallet, err := NewEVMWallet(ks, chain, "test")
	require.NoError(t, err)
	return wallet, ks
}

// TestChain tests the Chain method
func TestChain(t *testing.T) {
	wallet, _ := setupTest(t)
	chain := wallet.Chain()
	assert.Equal(t, types.ChainTypeEthereum, chain.Type, "Chain.Type should be Ethereum")
	assert.Equal(t, int64(1), chain.ID, "Chain.ID should be 1")
	assert.Equal(t, "Ethereum", chain.Name, "Chain.Name should be Ethereum")
}

// TestDeriveAddress tests the DeriveAddress method
func TestDeriveAddress(t *testing.T) {
	wallet, ks := setupTest(t)
	ctx := context.Background()

	// Generate a test key pair using secp256k1 curve
	privKey, err := ecdsa.GenerateKey(coreCrypto.Secp256k1Curve, rand.Reader)
	require.NoError(t, err)

	// Get public key bytes using secp256k1 format
	pubKeyBytes, err := coreCrypto.MarshalPublicKey(&privKey.PublicKey)
	require.NoError(t, err)

	// Set up the mock to return our key
	ks.GetPublicKeyFunc = func(ctx context.Context, id string) (*keystore.Key, error) {
		if id == "test" {
			return &keystore.Key{
				ID:        "test",
				Name:      "test",
				Type:      types.KeyTypeECDSA,
				Curve:     coreCrypto.Secp256k1Curve,
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
	wallet, ks := setupTest(t)
	ctx := context.Background()

	// Setup key using secp256k1 curve
	privKey, err := ecdsa.GenerateKey(coreCrypto.Secp256k1Curve, rand.Reader)
	require.NoError(t, err)

	// Get public key bytes using secp256k1 format
	pubKeyBytes, err := coreCrypto.MarshalPublicKey(&privKey.PublicKey)
	require.NoError(t, err)

	// Marshal private key using secp256k1 format
	privKeyBytes, err := coreCrypto.MarshalPrivateKey(privKey)
	require.NoError(t, err)

	// Import the key into the mock keystore
	_, err = ks.Import(ctx, "test", types.KeyTypeECDSA, coreCrypto.Secp256k1Curve, privKeyBytes, pubKeyBytes, nil)
	require.NoError(t, err)

	// Set up the mock to return our key
	ks.GetPublicKeyFunc = func(ctx context.Context, id string) (*keystore.Key, error) {
		if id == "test" {
			return &keystore.Key{
				ID:        "test",
				Name:      "test",
				Type:      types.KeyTypeECDSA,
				Curve:     coreCrypto.Secp256k1Curve,
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
	wallet, ks := setupTest(t)
	ctx := context.Background()

	// Setup key using secp256k1 curve
	privKey, err := ecdsa.GenerateKey(coreCrypto.Secp256k1Curve, rand.Reader)
	require.NoError(t, err)

	// Get public key bytes using secp256k1 format
	pubKeyBytes, err := coreCrypto.MarshalPublicKey(&privKey.PublicKey)
	require.NoError(t, err)

	// Marshal private key using secp256k1 format
	privKeyBytes, err := coreCrypto.MarshalPrivateKey(privKey)
	require.NoError(t, err)

	// Import the key into the mock keystore
	_, err = ks.Import(ctx, "test", types.KeyTypeECDSA, coreCrypto.Secp256k1Curve, privKeyBytes, pubKeyBytes, nil)
	require.NoError(t, err)

	// Set up the mock to return our key
	ks.GetPublicKeyFunc = func(ctx context.Context, id string) (*keystore.Key, error) {
		if id == "test" {
			return &keystore.Key{
				ID:        "test",
				Name:      "test",
				Type:      types.KeyTypeECDSA,
				Curve:     coreCrypto.Secp256k1Curve,
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
	wallet, ks := setupTest(t)
	ctx := context.Background()

	// Generate a test key pair
	privKey, err := ecdsa.GenerateKey(coreCrypto.Secp256k1Curve, rand.Reader)
	require.NoError(t, err)

	pubKeyBytes, err := coreCrypto.MarshalPublicKey(&privKey.PublicKey)
	require.NoError(t, err)

	privKeyBytes, err := coreCrypto.MarshalPrivateKey(privKey)
	require.NoError(t, err)

	// Import the key into the mock keystore
	_, err = ks.Import(ctx, "test", types.KeyTypeECDSA, coreCrypto.Secp256k1Curve, privKeyBytes, pubKeyBytes, nil)
	require.NoError(t, err)

	// Set up the mock
	address := crypto.PubkeyToAddress(privKey.PublicKey).Hex()
	ks.GetPublicKeyFunc = func(ctx context.Context, id string) (*keystore.Key, error) {
		return &keystore.Key{
			ID:        "test",
			Type:      types.KeyTypeECDSA,
			Curve:     coreCrypto.Secp256k1Curve,
			PublicKey: pubKeyBytes,
		}, nil
	}

	// Create a test transaction
	tx := &types.Transaction{
		Chain:    types.ChainTypeEthereum,
		From:     address,
		To:       "0x742d35Cc6634C0532925a3b844Bc454e4438f44e",
		Value:    big.NewInt(1000000000000000000), // 1 ETH
		GasPrice: big.NewInt(20000000000),         // 20 Gwei
		GasLimit: 21000,
		Nonce:    0,
	}

	// Set up the mock to sign with the private key
	ks.SignFunc = func(ctx context.Context, id string, data []byte, dataType keystore.DataType) ([]byte, error) {
		if dataType != keystore.DataTypeDigest {
			return nil, fmt.Errorf("unexpected data type: %v", dataType)
		}

		// Sign the data using the private key
		r, s, err := ecdsa.Sign(rand.Reader, privKey, data)
		if err != nil {
			return nil, err
		}

		// Marshal the signature in ASN.1 DER format
		sig := ecdsaSignature{R: r, S: s}
		return asn1.Marshal(sig)
	}

	// Sign the transaction
	signedTx, err := wallet.SignTransaction(ctx, tx)
	require.NoError(t, err)
	assert.NotEmpty(t, signedTx, "Signed transaction should not be empty")
}

// TestNewEVMWalletValidation tests the validation in NewEVMWallet
func TestNewEVMWalletValidation(t *testing.T) {
	ks := &MockKeyStore{}

	// Test with nil keystore
	_, err := NewEVMWallet(nil, types.Chain{}, "test")
	assert.Error(t, err, "NewEVMWallet should fail with nil keystore")

	// Test with empty keyID
	_, err = NewEVMWallet(ks, types.Chain{}, "")
	assert.Error(t, err, "NewEVMWallet should fail with empty keyID")

	// Test with invalid key type
	_, err = NewEVMWallet(ks, types.Chain{
		KeyType: types.KeyTypeRSA,
		Curve:   coreCrypto.Secp256k1Curve,
	}, "test")
	assert.Error(t, err, "NewEVMWallet should fail with invalid key type")

	// Test with invalid curve
	_, err = NewEVMWallet(ks, types.Chain{
		KeyType: types.KeyTypeECDSA,
		Curve:   elliptic.P256(),
	}, "test")
	assert.Error(t, err, "NewEVMWallet should fail with invalid curve")
}
